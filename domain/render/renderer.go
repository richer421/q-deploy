package render

import (
	"context"
	"fmt"

	rendermodel "github.com/richer421/q-deploy/domain/render/model"
	"github.com/richer421/q-deploy/domain/render/k8s_native"
)

// 对外暴露的快照与输入输出类型，基于 model 包做类型别名，保持现有引用路径不变

type InstanceConfigSnapshot = rendermodel.InstanceConfigSnapshot

type ReleaseStrategySnapshot = rendermodel.ReleaseStrategySnapshot

type BatchRuleSnapshot = rendermodel.BatchRuleSnapshot

type CanaryTrafficRuleSnapshot = rendermodel.CanaryTrafficRuleSnapshot

type Input = rendermodel.Input

type Output = rendermodel.Output

// Renderer 定义不同发布策略渲染器的统一接口

type Renderer interface {
	Render(ctx context.Context, in Input) (*Output, error)
}

// Type 表示渲染器类型（工厂入口）

type Type string

const (
	TypeK8sNative Type = "k8s_native"
)

// Config 聚合创建渲染器所需的配置
// 目前只有类型信息，后续可根据不同实现扩展字段

type Config struct {
	Type Type
}

// New 是渲染器总工厂
// 目前仅支持 TypeK8sNative

func New(cfg Config) (Renderer, error) {
	switch cfg.Type {
	case TypeK8sNative:
		return k8s_native.New(), nil
	default:
		return nil, fmt.Errorf("render: unsupported type %q", cfg.Type)
	}
}
