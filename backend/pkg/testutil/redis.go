package testutil

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// NewMockRedis 启动一个内存 Redis 并返回 go-redis Client。
// 测试结束后调用 miniredis.Close() 释放资源。
func NewMockRedis() (*redis.Client, *miniredis.Miniredis, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr, nil
}
