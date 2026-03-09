package middleware

import (
	"github.com/richer421/q-deploy/conf"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func OTel() gin.HandlerFunc {
	return otelgin.Middleware(conf.C.OTel.ServiceName)
}
