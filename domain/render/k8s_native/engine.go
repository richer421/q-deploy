package k8s_native

import (
	"context"

	"gopkg.in/yaml.v3"

	rendermodel "github.com/richer421/q-deploy/domain/render/model"
)

// Renderer 当前版本唯一落地的渲染实现（K8s 原生 Deployment）
// - 仅支持 rolling 模式（其他模式先降级为 rolling）

type Renderer struct{}

func New() *Renderer {
	return &Renderer{}
}

// Deployment 是一个极简的 Deployment 结构，只覆盖当前需要的字段
// 为了避免直接依赖 k8s 官方库，这里手写必要字段

type Deployment struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   map[string]any `yaml:"metadata"`
	Spec       map[string]any `yaml:"spec"`
}

func (e *Renderer) Render(_ context.Context, in rendermodel.Input) (*rendermodel.Output, error) {
	// 当前版本只实现 deployment + rolling
	// 其他实例类型如果传进来，也按 deployment 方式渲染，调用方自己约束
	// 拷贝一份 spec，避免直接修改输入
	spec := deploymentSpec(in.Instance.Spec)

	// 注入镜像
	injectImage(spec, in.ImageRef)

	dep := Deployment{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata: map[string]any{
			"name": in.ServiceName,
			"labels": map[string]any{
				"app":     in.ServiceName,
				"project": in.ProjectKey,
				"env":     in.Env,
			},
		},
		Spec: spec,
	}

	workloadYAML, err := yaml.Marshal(&dep)
	if err != nil {
		return nil, err
	}

	return &rendermodel.Output{
		WorkloadYAML: string(workloadYAML),
		ResourceYAML: "",
	}, nil
}

func deploymentSpec(spec map[string]any) map[string]any {
	root := copyMap(spec)
	if root == nil {
		return nil
	}
	if deployment, ok := root["deployment"].(map[string]any); ok {
		return copyMap(deployment)
	}
	return root
}

// copyMap 做一个浅拷贝，避免直接修改输入数据结构

func copyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// injectImage 在 deployment spec 中递归查找 template.spec.containers[*].image 并替换为 ImageRef

func injectImage(spec map[string]any, image string) {
	template, ok := spec["template"].(map[string]any)
	if !ok {
		return
	}
	podSpec, ok := template["spec"].(map[string]any)
	if !ok {
		return
	}
	containers, ok := podSpec["containers"].([]any)
	if !ok {
		return
	}
	for i, c := range containers {
		cont, ok := c.(map[string]any)
		if !ok {
			continue
		}
		cont["image"] = image
		containers[i] = cont
	}
	podSpec["containers"] = containers
	template["spec"] = podSpec
	spec["template"] = template
}
