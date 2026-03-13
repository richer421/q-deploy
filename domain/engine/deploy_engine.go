package engine

import (
	"context"
	"fmt"

	"github.com/richer421/q-deploy/domain/engine/gitops"
	"github.com/richer421/q-deploy/domain/render"
)

// Engine 是部署发布引擎的统一抽象
// 目前只有 GitOps 实现，后续可以在 engine 包下扩展其它实现（如 rollouts 等）

type Engine interface {
	Publish(ctx context.Context, in gitops.PublishInput) (*gitops.PublishResult, error)
}

// Type 表示引擎类型（工厂模式入口）

type Type string

const (
	TypeGitOps Type = "gitops"
)

// Config 聚合创建 Engine 所需的配置
// 不关心具体实现细节，由子包各自使用相应字段

type Config struct {
	Type Type

	AppTemplatePath string
	ArgoNamespace   string
	ArgoProject     string
	ClusterServer   string
}

// New 是 Engine 总工厂
// 根据 cfg.Type 选择具体实现，保证上层只依赖 engine 包而不直接感知 gitops 等子实现

func New(cfg Config, renderer render.Renderer, gitClient gitops.GitClient, releaseRepo gitops.ReleaseRepo, appApplier gitops.ApplicationApplier) (Engine, error) {
	switch cfg.Type {
	case TypeGitOps:
		return gitops.NewEngine(
			renderer,
			gitClient,
			releaseRepo,
			appApplier,
			cfg.AppTemplatePath,
			cfg.ArgoNamespace,
			cfg.ArgoProject,
			cfg.ClusterServer,
		), nil
	default:
		return nil, fmt.Errorf("engine: unsupported type %q", cfg.Type)
	}
}
