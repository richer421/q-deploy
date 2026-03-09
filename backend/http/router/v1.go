package router

import (
	"github.com/richer421/q-deploy/http/api"

	"github.com/gin-gonic/gin"
)

func RegisterV1(apiGroup *gin.RouterGroup) {
	v1 := apiGroup.Group("/v1")

	registerHelloWorld(v1)
}

func registerHelloWorld(rg *gin.RouterGroup) {
	h := api.NewHelloWorldAPI()
	g := rg.Group("/hello-world")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
