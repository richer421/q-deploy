package router

import (
	"github.com/gin-gonic/gin"
	"github.com/richer421/q-deploy/domain/engine"
	"github.com/richer421/q-deploy/http/api"
)

func RegisterV1(apiGroup *gin.RouterGroup, engine engine.Engine) {
	v1 := apiGroup.Group("/v1")

	deployAPI := api.NewDeployAPI(engine)
	// GitOps 发布入口
	v1.POST("/deploy/gitops", deployAPI.ReleaseWithGitOps)
}
