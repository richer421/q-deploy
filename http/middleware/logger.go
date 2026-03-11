package middleware

import (
	"time"

	"github.com/richer421/q-deploy/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Infof("%s %s %d %s %s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			time.Since(start),
			c.ClientIP(),
		)
	}
}
