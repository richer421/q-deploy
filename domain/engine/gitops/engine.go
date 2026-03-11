package gitops

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/richer421/q-deploy/domain/render"
	"github.com/richer421/q-deploy/infra/mysql/dao"
	"github.com/richer421/q-deploy/infra/mysql/model"
)

// ConfigSnapshot 来自 q-metahub CDConfig.GitOps 的快照

type ConfigSnapshot struct {
	RepoURL      string
	Branch       string
	AppRoot      string
	ManifestRoot string
}

// PublishInput 聚合一次 GitOps 发布所需的全部上下文

type PublishInput struct {
	BusinessUnitID int64
	DeployPlanID   int64
	CDConfigID     int64

	ProjectKey  string
	ServiceName string
	Env         string

	BuildArtifactID int64
	ImageRef        string

	Instance     render.InstanceConfigSnapshot
	Strategy     render.ReleaseStrategySnapshot
	GitOpsConfig ConfigSnapshot
}

// PublishResult 返回 ReleaseID 以及 GitOps 快照（方便调用方使用）

type PublishResult struct {
	ReleaseID      int64
	GitOpsSnapshot model.GitOpsSnapshot
}

// GitClient 抽象 git 操作，方便单测/替换实现

type GitClient interface {
	CloneOrPull(ctx context.Context, repoURL, branch string) (WorkingCopy, error)
}

// WorkingCopy 表示某个分支的工作副本

type WorkingCopy interface {
	WriteFile(path string, data []byte) error
	CommitAndPush(message string) error
}

// ReleaseRepo 抽象 Release 持久化
// 默认实现可以基于 infra/mysql/dao

type ReleaseRepo interface {
	Create(ctx context.Context, r *model.Release) (int64, error)
}

// Engine 将一次发布转换为 GitOps + Release 记录的具体实现
// 满足 domain/engine.Engine 接口

type Engine struct {
	renderer        render.Renderer
	gitClient       GitClient
	releaseRepo     ReleaseRepo
	appTemplatePath string
	argoNamespace   string
	argoProject     string
	clusterServer   string
}

func NewEngine(renderer render.Renderer, gitClient GitClient, releaseRepo ReleaseRepo,
	appTemplatePath, argoNamespace, argoProject, clusterServer string,
) *Engine {
	if releaseRepo == nil {
		releaseRepo = &defaultReleaseRepo{}
	}
	return &Engine{
		renderer:        renderer,
		gitClient:       gitClient,
		releaseRepo:     releaseRepo,
		appTemplatePath: appTemplatePath,
		argoNamespace:   argoNamespace,
		argoProject:     argoProject,
		clusterServer:   clusterServer,
	}
}

// defaultReleaseRepo 使用 gorm-gen 生成的 dao.Release 进行持久化

type defaultReleaseRepo struct{}

func (r *defaultReleaseRepo) Create(ctx context.Context, rel *model.Release) (int64, error) {
	if err := dao.Q.WithContext(ctx).Release.Create(rel); err != nil {
		return 0, err
	}
	return rel.ID, nil
}

func (e *Engine) Publish(ctx context.Context, in PublishInput) (*PublishResult, error) {
	if in.GitOpsConfig.RepoURL == "" || in.GitOpsConfig.Branch == "" ||
		in.GitOpsConfig.AppRoot == "" || in.GitOpsConfig.ManifestRoot == "" {
		return nil, fmt.Errorf("gitops: invalid GitOpsConfig")
	}

	// 1. 策略渲染出最终工作负载 YAML
	renderOut, err := e.renderer.Render(ctx, render.Input{
		Instance:    in.Instance,
		Strategy:    in.Strategy,
		ImageRef:    in.ImageRef,
		ProjectKey:  in.ProjectKey,
		ServiceName: in.ServiceName,
		Env:         in.Env,
	})
	if err != nil {
		return nil, fmt.Errorf("gitops: strategy render failed: %w", err)
	}

	// 2. 计算路径和 Application 名称
	manifestPath := filepath.Join(in.GitOpsConfig.ManifestRoot, in.ProjectKey, in.Env, in.ServiceName)
	appPath := filepath.Join(in.GitOpsConfig.AppRoot, in.ProjectKey, in.Env, in.ServiceName+".yaml")
	appName := fmt.Sprintf("%s-%s-%s", in.ProjectKey, in.ServiceName, in.Env)

	// 3. 准备 Application 模板数据
	appData := map[string]any{
		"ProjectKey":     in.ProjectKey,
		"ServiceName":    in.ServiceName,
		"Env":            in.Env,
		"ArgoNamespace":  e.argoNamespace,
		"ArgoProject":    e.argoProject,
		"GitOpsRepoURL":  in.GitOpsConfig.RepoURL,
		"GitOpsBranch":   in.GitOpsConfig.Branch,
		"ManifestPath":   manifestPath,
		"ClusterServer":  e.clusterServer,
		"Namespace":      fmt.Sprintf("%s-%s", in.ServiceName, in.Env),
		"PlanID":         fmt.Sprintf("%d", in.DeployPlanID),
		"DeploymentMode": in.Strategy.DeploymentMode,
	}

	appYAML, err := e.renderAppTemplate(appData)
	if err != nil {
		return nil, fmt.Errorf("gitops: render app template failed: %w", err)
	}

	// 4. 写入 Git 仓库
	wc, err := e.gitClient.CloneOrPull(ctx, in.GitOpsConfig.RepoURL, in.GitOpsConfig.Branch)
	if err != nil {
		return nil, fmt.Errorf("gitops: clone/pull failed: %w", err)
	}

	if err := wc.WriteFile(filepath.Join(manifestPath, "workload.yaml"), []byte(renderOut.WorkloadYAML)); err != nil {
		return nil, fmt.Errorf("gitops: write workload.yaml failed: %w", err)
	}
	if renderOut.ResourceYAML != "" {
		if err := wc.WriteFile(filepath.Join(manifestPath, "resources.yaml"), []byte(renderOut.ResourceYAML)); err != nil {
			return nil, fmt.Errorf("gitops: write resources.yaml failed: %w", err)
		}
	}
	if err := wc.WriteFile(appPath, []byte(appYAML)); err != nil {
		return nil, fmt.Errorf("gitops: write app yaml failed: %w", err)
	}

	commitMsg := fmt.Sprintf("deploy plan %d via gitops", in.DeployPlanID)
	if err := wc.CommitAndPush(commitMsg); err != nil {
		return nil, fmt.Errorf("gitops: commit/push failed: %w", err)
	}

	// 5. 创建 Release 记录
	release := &model.Release{
		BusinessUnitID:  in.BusinessUnitID,
		DeployPlanID:    in.DeployPlanID,
		CDConfigID:      in.CDConfigID,
		BuildArtifactID: in.BuildArtifactID,
		ImageRef:        in.ImageRef,
		WorkloadYAML:    renderOut.WorkloadYAML,
		ResourceYAML:    renderOut.ResourceYAML,
		RendererType:    "gitops-k8s-native",
		EngineType:      "gitops",
		ReleaseStrategy: model.ReleaseStrategy{
			DeploymentMode: in.Strategy.DeploymentMode,
			BatchRule: model.BatchRule{
				BatchCount:  in.Strategy.BatchRule.BatchCount,
				BatchRatio:  in.Strategy.BatchRule.BatchRatio,
				TriggerType: in.Strategy.BatchRule.TriggerType,
				Interval:    in.Strategy.BatchRule.Interval,
			},
			CanaryTrafficRule: convertCanarySnapshot(in.Strategy.CanaryTrafficRule),
		},
		GitOpsSnapshot: model.GitOpsSnapshot{
			RepoURL:      in.GitOpsConfig.RepoURL,
			Branch:       in.GitOpsConfig.Branch,
			ManifestPath: manifestPath,
			AppPath:      appPath,
			AppName:      appName,
		},
	}

	id, err := e.releaseRepo.Create(ctx, release)
	if err != nil {
		return nil, fmt.Errorf("gitops: create release failed: %w", err)
	}

	return &PublishResult{
		ReleaseID:      id,
		GitOpsSnapshot: release.GitOpsSnapshot,
	}, nil
}

func (e *Engine) renderAppTemplate(data map[string]any) (string, error) {
	tpl, err := template.ParseFiles(e.appTemplatePath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func convertCanarySnapshot(in *render.CanaryTrafficRuleSnapshot) *model.CanaryTrafficRule {
	if in == nil {
		return nil
	}
	return &model.CanaryTrafficRule{
		TrafficBatchCount: in.TrafficBatchCount,
		TrafficRatioList:  in.TrafficRatioList,
		ManualAdjust:      in.ManualAdjust,
		AdjustTimeout:     in.AdjustTimeout,
	}
}
