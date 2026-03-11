package deploy

import (
	"context"

	"github.com/richer421/q-deploy/app/deploy/vo"
	"github.com/richer421/q-deploy/domain/engine"
	"github.com/richer421/q-deploy/domain/engine/gitops"
)

// AppService GitOps 发布应用服务

type AppService struct {
	gitOpsEngine engine.Engine
}

func NewAppService(engine engine.Engine) *AppService {
	return &AppService{gitOpsEngine: engine}
}

// ReleaseWithGitOps 执行一次 GitOps 发布
// 当前版本直接使用调用方提供的 Snapshot，后续可替换为从 q-metahub/q-ci 查询

func (s *AppService) ReleaseWithGitOps(ctx context.Context, cmd *vo.ReleaseWithGitOpsCmd) (*vo.ReleaseDTO, error) {
	in := gitops.PublishInput{
		BusinessUnitID:  cmd.BusinessUnitID,
		DeployPlanID:    cmd.DeployPlanID,
		CDConfigID:      cmd.CDConfigID,
		ProjectKey:      cmd.ProjectKey,
		ServiceName:     cmd.ServiceName,
		Env:             cmd.Env,
		BuildArtifactID: cmd.BuildArtifactID,
		ImageRef:        cmd.ImageRef,
		Instance:        cmd.InstanceSnapshot,
		Strategy:        cmd.StrategySnapshot,
		GitOpsConfig:    cmd.GitOpsSnapshot,
	}

	res, err := s.gitOpsEngine.Publish(ctx, in)
	if err != nil {
		return nil, err
	}

	return &vo.ReleaseDTO{
		ReleaseID: res.ReleaseID,
		GitOps: vo.GitOpsSnapshotDTO{
			RepoURL:      res.GitOpsSnapshot.RepoURL,
			Branch:       res.GitOpsSnapshot.Branch,
			ManifestPath: res.GitOpsSnapshot.ManifestPath,
			AppPath:      res.GitOpsSnapshot.AppPath,
			AppName:      res.GitOpsSnapshot.AppName,
		},
	}, nil
}
