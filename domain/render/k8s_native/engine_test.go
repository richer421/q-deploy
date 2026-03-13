package k8s_native

import (
	"context"
	"testing"

	"gopkg.in/yaml.v3"

	rendermodel "github.com/richer421/q-deploy/domain/render/model"
)

func TestRenderUsesDeploymentSubSpec(t *testing.T) {
	r := New()

	out, err := r.Render(context.Background(), rendermodel.Input{
		ImageRef:    "localhost:30180/q-demo/q-demo:latest",
		ProjectKey:  "q-demo-project",
		ServiceName: "q-demo",
		Env:         "dev",
		Instance: rendermodel.InstanceConfigSnapshot{
			InstanceType: "deployment",
			Spec: map[string]any{
				"deployment": map[string]any{
					"selector": map[string]any{
						"matchLabels": map[string]any{
							"app": "q-demo",
						},
					},
					"template": map[string]any{
						"metadata": map[string]any{
							"labels": map[string]any{
								"app": "q-demo",
							},
						},
						"spec": map[string]any{
							"containers": []any{
								map[string]any{
									"name":  "q-demo",
									"image": "nginx:old",
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	var manifest map[string]any
	if err := yaml.Unmarshal([]byte(out.WorkloadYAML), &manifest); err != nil {
		t.Fatalf("failed to unmarshal workload yaml: %v", err)
	}

	spec, ok := manifest["spec"].(map[string]any)
	if !ok {
		t.Fatalf("spec is not a map: %#v", manifest["spec"])
	}
	if _, exists := spec["deployment"]; exists {
		t.Fatalf("deployment wrapper should not exist in rendered spec: %#v", spec)
	}

	template, ok := spec["template"].(map[string]any)
	if !ok {
		t.Fatalf("template is not a map: %#v", spec["template"])
	}
	podSpec, ok := template["spec"].(map[string]any)
	if !ok {
		t.Fatalf("template.spec is not a map: %#v", template["spec"])
	}
	containers, ok := podSpec["containers"].([]any)
	if !ok || len(containers) != 1 {
		t.Fatalf("containers not rendered correctly: %#v", podSpec["containers"])
	}
	container := containers[0].(map[string]any)
	if got := container["image"]; got != "localhost:30180/q-demo/q-demo:latest" {
		t.Fatalf("container image = %#v, want updated image ref", got)
	}
}
