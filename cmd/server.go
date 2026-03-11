package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/richer421/q-deploy/conf"
	apphttp "github.com/richer421/q-deploy/http"
	infrakafka "github.com/richer421/q-deploy/infra/kafka"
	inframysql "github.com/richer421/q-deploy/infra/mysql"
	infraredis "github.com/richer421/q-deploy/infra/redis"
	"github.com/richer421/q-deploy/pkg/logger"
	appotel "github.com/richer421/q-deploy/pkg/otel"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 HTTP 服务",
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Sync()

		initInfra()
		defer closeInfra()

		serve()
	},
}

// initInfra 初始化所有基础设施组件
func initInfra() {
	// OTel
	otelShutdown, err := appotel.Init(conf.C.OTel)
	if err != nil {
		logger.Fatalf("otel init: %s", err)
	}
	closers = append(closers, namedCloser{"otel", func() error { return otelShutdown(context.Background()) }})

	// MySQL
	if err := inframysql.Init(conf.C.MySQL); err != nil {
		logger.Fatalf("mysql init: %s", err)
	}
	closers = append(closers, namedCloser{"mysql", inframysql.Close})

	// Redis
	if err := infraredis.Init(conf.C.Redis); err != nil {
		logger.Fatalf("redis init: %s", err)
	}
	closers = append(closers, namedCloser{"redis", infraredis.Close})

	// Kafka
	if err := infrakafka.Init(conf.C.Kafka); err != nil {
		logger.Fatalf("kafka init: %s", err)
	}
	closers = append(closers, namedCloser{"kafka", func() error { infrakafka.Close(); return nil }})

	// Kafka 消费者
	consumeCtx, consumeCancel := context.WithCancel(context.Background())
	closers = append(closers, namedCloser{"kafka-consumers", func() error {
		infrakafka.StopConsumers()
		consumeCancel()
		return nil
	}})
	if err := infrakafka.StartConsumers(consumeCtx); err != nil {
		logger.Fatalf("kafka start consumers: %s", err)
	}
}

type namedCloser struct {
	name string
	fn   func() error
}

var closers []namedCloser

// closeInfra 按注册逆序关闭所有组件
func closeInfra() {
	for i := len(closers) - 1; i >= 0; i-- {
		c := closers[i]
		if err := c.fn(); err != nil {
			logger.Errorf("%s close: %s", c.name, err)
		}
	}
}

// serve 启动 HTTP 服务并等待信号优雅关停
func serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.C.Server.Port),
		Handler: apphttp.NewServer(),
	}

	go func() {
		logger.Infof("server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("server forced to shutdown: %s", err)
	}
	logger.Infof("server exited")
}
