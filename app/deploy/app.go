package deploy

import (
	"context"
	"fmt"

	"github.com/richer421/q-deploy/app/deploy/vo"
	"github.com/richer421/q-deploy/domain/engine"
	"github.com/richer421/q-deploy/domain/engine/gitops"
	render "github.com/richer421/q-deploy/domain/render"
	metahubclient "github.com/richer421/q-deploy/infra/metahub"
	qciclient "github.com/richer421/q-deploy/infra/qci"
)

// AppService GitOps 发布应用服务

type AppService struct {
	gitOpsEngine  engine.Engine
	metahubClient deployPlanGetter
	qciClient     artifactGetter
}

func NewAppService(engine engine.Engine) *AppService {
	return &AppService{
		gitOpsEngine:  engine,
		metahubClient: metahubclient.New(),
		qciClient:     qciclient.New(),
	}
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

type deployPlanGetter interface {
	GetDeployPlan(ctx context.Context, deployPlanID int64) (*metahubclient.DeployPlanSpecDTO, error)
}

type artifactGetter interface {
	GetArtifact(ctx context.Context, artifactID int64) (*qciclient.ArtifactDTO, error)
}

type fakeMetahubClient struct {
	res *metahubclient.DeployPlanSpecDTO
	err error
}

func (f *fakeMetahubClient) GetDeployPlan(_ context.Context, _ int64) (*metahubclient.DeployPlanSpecDTO, error) {
	return f.res, f.err
}

type fakeQCIClient struct {
	res *qciclient.ArtifactDTO
	err error
}

func (f *fakeQCIClient) GetArtifact(_ context.Context, _ int64) (*qciclient.ArtifactDTO, error) {
	return f.res, f.err
}

func (s *AppService) ExecuteRelease(ctx context.Context, req *vo.ExecuteReleaseReq) (*vo.ReleaseDTO, error) {
	plan, err := s.metahubClient.GetDeployPlan(ctx, req.DeployPlanID)
	if err != nil {
		return nil, err
	}

	artifact, err := s.qciClient.GetArtifact(ctx, req.BuildArtifactID)
	if err != nil {
		return nil, err
	}
	if artifact.DeployPlanID != 0 && artifact.DeployPlanID != req.DeployPlanID {
		return nil, fmt.Errorf("artifact %d does not belong to deploy plan %d", artifact.ID, req.DeployPlanID)
	}

	in, err := buildPublishInput(plan, artifact)
	if err != nil {
		return nil, err
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

func buildPublishInput(plan *metahubclient.DeployPlanSpecDTO, artifact *qciclient.ArtifactDTO) (gitops.PublishInput, error) {
	gitOpsCfg := gitops.ConfigSnapshot{
		RepoURL:      stringMap(plan.CDConfig.GitOps, "repo_url"),
		Branch:       stringMap(plan.CDConfig.GitOps, "branch"),
		AppRoot:      stringMap(plan.CDConfig.GitOps, "app_root"),
		ManifestRoot: stringMap(plan.CDConfig.GitOps, "manifest_root"),
	}

	strategy := render.ReleaseStrategySnapshot{
		DeploymentMode: stringMap(plan.CDConfig.ReleaseStrategy, "deployment_mode"),
	}
	if batchRule, ok := plan.CDConfig.ReleaseStrategy["batch_rule"].(map[string]any); ok {
		strategy.BatchRule = render.BatchRuleSnapshot{
			BatchCount:  intMap(batchRule, "batch_count"),
			BatchRatio:  floatSliceMap(batchRule, "batch_ratio"),
			TriggerType: stringMap(batchRule, "trigger_type"),
			Interval:    intMap(batchRule, "interval"),
		}
	}

	return gitops.PublishInput{
		BusinessUnitID:  plan.BusinessUnit.ID,
		DeployPlanID:    plan.DeployPlan.ID,
		CDConfigID:      plan.CDConfig.ID,
		ProjectKey:      plan.Project.Name,
		ServiceName:     plan.BusinessUnit.Name,
		Env:             plan.InstanceConfig.Env,
		BuildArtifactID: artifact.ID,
		ImageRef:        artifact.ImageRef,
		Instance: render.InstanceConfigSnapshot{
			Env:             plan.InstanceConfig.Env,
			InstanceType:    plan.InstanceConfig.InstanceType,
			Spec:            plan.InstanceConfig.Spec,
			AttachResources: plan.InstanceConfig.AttachResources,
		},
		Strategy:     strategy,
		GitOpsConfig: gitOpsCfg,
	}, nil
}

func stringMap(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func intMap(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func floatSliceMap(m map[string]any, key string) []float64 {
	items, ok := m[key].([]any)
	if !ok {
		if typed, ok := m[key].([]float64); ok {
			return typed
		}
		return nil
	}
	out := make([]float64, 0, len(items))
	for _, item := range items {
		switch v := item.(type) {
		case float64:
			out = append(out, v)
		case int:
			out = append(out, float64(v))
		case int64:
			out = append(out, float64(v))
		}
	}
	return out
}
