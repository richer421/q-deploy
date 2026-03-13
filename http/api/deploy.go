package api

import (
	"context"

	"github.com/gin-gonic/gin"
	appdeploy "github.com/richer421/q-deploy/app/deploy"
	"github.com/richer421/q-deploy/app/deploy/vo"
	"github.com/richer421/q-deploy/domain/engine"
	"github.com/richer421/q-deploy/http/common"
)

// DeployAPI 部署相关 API

type DeployAPI struct {
	appSvc releaseService
}

func NewDeployAPI(engine engine.Engine) *DeployAPI {
	return &DeployAPI{
		appSvc: appdeploy.NewAppService(engine),
	}
}

type releaseService interface {
	ExecuteRelease(ctx context.Context, req *vo.ExecuteReleaseReq) (*vo.ReleaseDTO, error)
	ReleaseWithGitOps(ctx context.Context, cmd *vo.ReleaseWithGitOpsCmd) (*vo.ReleaseDTO, error)
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

func (a *DeployAPI) ExecuteRelease(c *gin.Context) {
	var req vo.ExecuteReleaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	res, err := a.appSvc.ExecuteRelease(c.Request.Context(), &req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}
