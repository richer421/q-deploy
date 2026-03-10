package router

import (
	"github.com/gin-gonic/gin"
)

func RegisterV1(apiGroup *gin.RouterGroup) {
	v1 := apiGroup.Group("/v1")

	// TODO: 注册业务 API
	_ = v1
}
