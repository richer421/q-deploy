package deploy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/richer421/q-deploy/app/deploy/vo"
	"github.com/richer421/q-deploy/domain/engine/gitops"
	render "github.com/richer421/q-deploy/domain/render"
	metahubclient "github.com/richer421/q-deploy/infra/metahub"
	"github.com/richer421/q-deploy/infra/mysql/model"
	qciclient "github.com/richer421/q-deploy/infra/qci"
)

type fakeEngine struct {
	lastInput gitops.PublishInput
}

func (f *fakeEngine) Publish(_ context.Context, in gitops.PublishInput) (*gitops.PublishResult, error) {
	f.lastInput = in
	return &gitops.PublishResult{
		ReleaseID: 99,
		GitOpsSnapshot: model.GitOpsSnapshot{
			RepoURL:      in.GitOpsConfig.RepoURL,
			Branch:       in.GitOpsConfig.Branch,
			ManifestPath: "manifests/q-demo/dev/q-demo",
			AppPath:      "apps/q-demo/dev/q-demo.yaml",
			AppName:      "q-demo-q-demo-dev",
		},
	}, nil
}

func TestAppServiceExecuteReleaseFromIDs(t *testing.T) {
	engine := &fakeEngine{}
	svc := NewAppService(engine)
	svc.metahubClient = &fakeMetahubClient{
		res: &metahubclient.DeployPlanSpecDTO{
			Project:      metahubclient.ProjectDTO{Name: "q-demo"},
			BusinessUnit: metahubclient.BusinessUnitDTO{ID: 2, Name: "q-demo"},
			CDConfig: metahubclient.CDConfigDTO{
				ID: 4,
				GitOps: map[string]any{
					"repo_url":      "https://github.com/richer421/q-demo-gitops.git",
					"branch":        "main",
					"app_root":      "apps",
					"manifest_root": "manifests",
				},
				ReleaseStrategy: map[string]any{
					"deployment_mode": "rolling",
					"batch_rule": map[string]any{
						"batch_count":  1,
						"batch_ratio":  []any{float64(1)},
						"trigger_type": "auto",
						"interval":     float64(0),
					},
				},
			},
			InstanceConfig: metahubclient.InstanceConfigDTO{
				Env:             "dev",
				InstanceType:    "deployment",
				Spec:            map[string]any{"deployment": map[string]any{}},
				AttachResources: map[string]any{"services": map[string]any{}},
			},
			DeployPlan: metahubclient.DeployPlanDTO{ID: 6, CDConfigID: 4},
		},
	}
	svc.qciClient = &fakeQCIClient{
		res: &qciclient.ArtifactDTO{
			ID:           21,
			DeployPlanID: 6,
			ImageRef:     "harbor.local/q-demo/q-demo:sha-123",
		},
	}

	resp, err := svc.ExecuteRelease(t.Context(), &vo.ExecuteReleaseReq{
		DeployPlanID:    6,
		BuildArtifactID: 21,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(99), resp.ReleaseID)
}

func TestAppServiceExecuteReleaseLoadsPlanAndArtifact(t *testing.T) {
	engine := &fakeEngine{}
	svc := NewAppService(engine)

	svc.metahubClient = &fakeMetahubClient{
		res: &metahubclient.DeployPlanSpecDTO{
			Project:      metahubclient.ProjectDTO{Name: "q-demo"},
			BusinessUnit: metahubclient.BusinessUnitDTO{ID: 2, Name: "q-demo"},
			CDConfig: metahubclient.CDConfigDTO{
				ID: 4,
				GitOps: map[string]any{
					"repo_url":      "https://github.com/richer421/q-demo-gitops.git",
					"branch":        "main",
					"app_root":      "apps",
					"manifest_root": "manifests",
				},
				ReleaseStrategy: map[string]any{
					"deployment_mode": "rolling",
					"batch_rule": map[string]any{
						"batch_count":  1,
						"batch_ratio":  []any{float64(1)},
						"trigger_type": "auto",
						"interval":     float64(0),
					},
				},
			},
			InstanceConfig: metahubclient.InstanceConfigDTO{
				Env:          "dev",
				InstanceType: "deployment",
				Spec: map[string]any{
					"deployment": map[string]any{},
				},
				AttachResources: map[string]any{
					"services": map[string]any{
						"q-demo": map[string]any{
							"metadata": map[string]any{"name": "q-demo"},
						},
					},
				},
			},
			DeployPlan: metahubclient.DeployPlanDTO{ID: 6, CDConfigID: 4},
		},
	}
	svc.qciClient = &fakeQCIClient{
		res: &qciclient.ArtifactDTO{
			ID:           21,
			DeployPlanID: 6,
			ImageRef:     "harbor.local/q-demo/q-demo:sha-123",
		},
	}

	resp, err := svc.ExecuteRelease(t.Context(), &vo.ExecuteReleaseReq{
		DeployPlanID:    6,
		BuildArtifactID: 21,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, int64(6), engine.lastInput.DeployPlanID)
	assert.Equal(t, int64(21), engine.lastInput.BuildArtifactID)
	assert.Equal(t, "harbor.local/q-demo/q-demo:sha-123", engine.lastInput.ImageRef)
	assert.Equal(t, "q-demo", engine.lastInput.ProjectKey)
	assert.Equal(t, "q-demo", engine.lastInput.ServiceName)
	assert.Equal(t, "dev", engine.lastInput.Env)
	assert.Equal(t, render.InstanceConfigSnapshot{
		Env:          "dev",
		InstanceType: "deployment",
		Spec:         map[string]any{"deployment": map[string]any{}},
		AttachResources: map[string]any{
			"services": map[string]any{
				"q-demo": map[string]any{
					"metadata": map[string]any{"name": "q-demo"},
				},
			},
		},
	}, engine.lastInput.Instance)
	assert.Equal(t, "main", engine.lastInput.GitOpsConfig.Branch)
	assert.Equal(t, int64(99), resp.ReleaseID)
}
