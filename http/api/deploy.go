package api

import (
	"github.com/gin-gonic/gin"
	appdeploy "github.com/richer421/q-deploy/app/deploy"
	"github.com/richer421/q-deploy/app/deploy/vo"
	"github.com/richer421/q-deploy/domain/engine"
	"github.com/richer421/q-deploy/http/common"
)

// DeployAPI 部署相关 API

type DeployAPI struct {
	appSvc *appdeploy.AppService
}

func NewDeployAPI(engine engine.Engine) *DeployAPI {
	return &DeployAPI{
		appSvc: appdeploy.NewAppService(engine),
	}
}

// ReleaseWithGitOps 触发一次 GitOps 发布

func (a *DeployAPI) ReleaseWithGitOps(c *gin.Context) {
	var cmd vo.ReleaseWithGitOpsCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		common.Fail(c, err)
		return
	}

	res, err := a.appSvc.ReleaseWithGitOps(c.Request.Context(), &cmd)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}
