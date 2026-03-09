package redis

import (
	"context"
	"time"

	"github.com/richer421/q-deploy/conf"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	RS     *redsync.Redsync
)

func Init(cfg conf.RedisConfig) error {
	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: poolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	// OTel instrumentation
	if err := redisotel.InstrumentTracing(Client); err != nil {
		return err
	}
	if err := redisotel.InstrumentMetrics(Client); err != nil {
		return err
	}

	pool := goredis.NewPool(Client)
	RS = redsync.New(pool)

	return nil
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
