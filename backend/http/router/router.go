package router

import (
	"context"
	"net/http"
	"time"

	infrakafka "github.com/richer421/q-deploy/infra/kafka"
	inframysql "github.com/richer421/q-deploy/infra/mysql"
	infraredis "github.com/richer421/q-deploy/infra/redis"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	// 存活探针
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 就绪探针
	r.GET("/readyz", readyz)

	// pprof
	pprof.Register(r)

	// 业务路由
	api := r.Group("/api")
	RegisterV1(api)
}

func readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	checks := make(map[string]string)
	healthy := true

	// MySQL
	if inframysql.DB != nil {
		sqlDB, err := inframysql.DB.DB()
		if err != nil {
			checks["mysql"] = err.Error()
			healthy = false
		} else if err := sqlDB.PingContext(ctx); err != nil {
			checks["mysql"] = err.Error()
			healthy = false
		} else {
			checks["mysql"] = "ok"
		}
	}

	// Redis
	if infraredis.Client != nil {
		if err := infraredis.Client.Ping(ctx).Err(); err != nil {
			checks["redis"] = err.Error()
			healthy = false
		} else {
			checks["redis"] = "ok"
		}
	}

	// Kafka
	if infrakafka.Producer != nil {
		_, err := infrakafka.Producer.GetMetadata(nil, true, 3000)
		if err != nil {
			checks["kafka"] = err.Error()
			healthy = false
		} else {
			checks["kafka"] = "ok"
		}
	}

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": map[bool]string{true: "ok", false: "unavailable"}[healthy],
		"checks": checks,
	})
}
