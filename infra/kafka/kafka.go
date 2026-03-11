package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/richer421/q-deploy/conf"
	"github.com/richer421/q-deploy/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/richer421/q-deploy/infra/kafka")

var (
	Producer   *kafka.Producer
	brokers    string
	maxRetries int
)

func Init(cfg conf.KafkaConfig) error {
	brokers = strings.Join(cfg.Brokers, ",")
	maxRetries = cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var err error
	Producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return fmt.Errorf("create kafka producer: %w", err)
	}

	// 异步处理 delivery report
	go func() {
		for e := range Producer.Events() {
			if m, ok := e.(*kafka.Message); ok && m.TopicPartition.Error != nil {
				logger.Errorf("kafka delivery failed: %s, topic: %s",
					m.TopicPartition.Error, *m.TopicPartition.Topic)
			}
		}
	}()

	return nil
}

func Produce(topic string, key, value []byte) error {
	return ProduceWithContext(context.Background(), topic, key, value)
}

func ProduceWithContext(ctx context.Context, topic string, key, value []byte) error {
	_, span := tracer.Start(ctx, topic+" publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKafka,
			semconv.MessagingDestinationName(topic),
		),
	)
	defer span.End()

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          value,
	}
	if key != nil {
		msg.Key = key
	}
	if err := Producer.Produce(msg, nil); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func Close() {
	stopConsumers()
	if Producer != nil {
		Producer.Flush(5000)
		Producer.Close()
	}
}

// HandleFunc 消费处理函数签名
type HandleFunc func(msg *kafka.Message) error

// ConsumerOption 消费者选项
type ConsumerOption func(*consumerConfig)

type consumerConfig struct {
	async      bool
	maxRetries int
}

type registration struct {
	topic   string
	groupID string
	handler HandleFunc
	config  consumerConfig
}

var (
	registrations []registration
	consumers     []*kafka.Consumer
	cancelFunc    context.CancelFunc
	wg            sync.WaitGroup
)

// WithAsync 标记为异步消费模式
func WithAsync() ConsumerOption {
	return func(c *consumerConfig) {
		c.async = true
	}
}

// WithMaxRetries 设置最大重试次数，覆盖全局配置
func WithMaxRetries(n int) ConsumerOption {
	return func(c *consumerConfig) {
		c.maxRetries = n
	}
}

// Register 注册消费函数，在 StartConsumers 之前调用
func Register(topic, groupID string, handler HandleFunc, opts ...ConsumerOption) {
	cfg := consumerConfig{maxRetries: maxRetries}
	for _, opt := range opts {
		opt(&cfg)
	}
	registrations = append(registrations, registration{
		topic:   topic,
		groupID: groupID,
		handler: handler,
		config:  cfg,
	})
}

// StartConsumers 启动所有已注册的消费者
func StartConsumers(ctx context.Context) error {
	var consumeCtx context.Context
	consumeCtx, cancelFunc = context.WithCancel(ctx)

	for _, reg := range registrations {
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  brokers,
			"group.id":           reg.groupID,
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": false,
		})
		if err != nil {
			return fmt.Errorf("create consumer for topic %s: %w", reg.topic, err)
		}

		if err := c.Subscribe(reg.topic, nil); err != nil {
			err := c.Close()
			if err != nil {
				return err
			}
			return fmt.Errorf("subscribe topic %s: %w", reg.topic, err)
		}

		consumers = append(consumers, c)

		wg.Add(1)
		if reg.config.async {
			go runAsyncConsumer(consumeCtx, c, reg)
		} else {
			go runSyncConsumer(consumeCtx, c, reg)
		}
	}

	return nil
}

func runSyncConsumer(ctx context.Context, c *kafka.Consumer, reg registration) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			msg, ok := ev.(*kafka.Message)
			if !ok {
				continue
			}
			handleWithRetry(c, msg, reg)
		}
	}
}

func runAsyncConsumer(ctx context.Context, c *kafka.Consumer, reg registration) {
	defer wg.Done()
	sem := make(chan struct{}, 64) // 控制并发上限
	var asyncWg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			asyncWg.Wait()
			return
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			msg, ok := ev.(*kafka.Message)
			if !ok {
				continue
			}

			sem <- struct{}{}
			asyncWg.Add(1)
			go func(m *kafka.Message) {
				defer func() {
					<-sem
					asyncWg.Done()
				}()
				handleWithRetry(c, m, reg)
			}(msg)
		}
	}
}

func handleWithRetry(c *kafka.Consumer, msg *kafka.Message, reg registration) {
	_, span := tracer.Start(context.Background(), reg.topic+" process",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			semconv.MessagingSystemKafka,
			semconv.MessagingDestinationName(reg.topic),
			attribute.String("messaging.kafka.consumer.group", reg.groupID),
		),
	)
	defer span.End()

	var err error
	for i := 0; i <= reg.config.maxRetries; i++ {
		err = reg.handler(msg)
		if err == nil {
			if _, commitErr := c.CommitMessage(msg); commitErr != nil {
				logger.Errorf("kafka commit offset failed: %s, topic: %s", commitErr, reg.topic)
			}
			return
		}
		logger.Warnf("kafka consume retry %d/%d failed: %s, topic: %s",
			i+1, reg.config.maxRetries, err, reg.topic)
	}
	if err != nil {
		logger.Errorf("kafka consume offset failed: %s, topic: %s", err, reg.topic)
		// 重试耗尽，记录错误并发送到死信队列
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		dlqTopic := reg.topic + ".DLQ"
		dlqErr := Produce(dlqTopic, msg.Key, msg.Value)
		if dlqErr != nil {
			logger.Errorf("kafka send to DLQ failed: %s, topic: %s, original error: %s",
				dlqErr, dlqTopic, err)
		} else {
			logger.Errorf("kafka message sent to DLQ: %s, original error: %s", dlqTopic, err)
		}

		// 提交 offset，继续消费
		if _, commitErr := c.CommitMessage(msg); commitErr != nil {
			logger.Errorf("kafka commit offset after DLQ failed: %s, topic: %s", commitErr, reg.topic)
		}
	}
}

func stopConsumers() {
	if cancelFunc != nil {
		cancelFunc()
	}
	wg.Wait()
	for _, c := range consumers {
		if err := c.Close(); err != nil {
			logger.Errorf("kafka consumer close failed: %s", err)
		}
	}
	consumers = nil
}

// StopConsumers 优雅停止所有消费者
func StopConsumers() {
	stopConsumers()
}
