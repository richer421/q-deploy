package gitops

import (
	"context"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/richer421/q-deploy/domain/render"
	"github.com/richer421/q-deploy/infra/mysql/model"
)

type fakeRenderer struct {
	input  render.Input
	output *render.Output
	err    error
}

func (f *fakeRenderer) Render(_ context.Context, in render.Input) (*render.Output, error) {
	f.input = in
	if f.err != nil {
		return nil, f.err
	}
	if f.output != nil {
		return f.output, nil
	}
	return &render.Output{
		WorkloadYAML: "workload: test\n",
		ResourceYAML: "resources: test\n",
	}, nil
}

type fakeWorkingCopy struct {
	files     map[string][]byte
	commitMsg string
}

func (w *fakeWorkingCopy) WriteFile(path string, data []byte) error {
	if w.files == nil {
		w.files = make(map[string][]byte)
	}
	// copy data to avoid aliasing
	buf := make([]byte, len(data))
	copy(buf, data)
	w.files[path] = buf
	return nil
}

func (w *fakeWorkingCopy) CommitAndPush(message string) error {
	w.commitMsg = message
	return nil
}

type fakeGitClient struct {
	lastRepoURL string
	lastBranch  string
	wc          *fakeWorkingCopy
}

func (c *fakeGitClient) CloneOrPull(_ context.Context, repoURL, branch string) (WorkingCopy, error) {
	c.lastRepoURL = repoURL
	c.lastBranch = branch
	if c.wc == nil {
		c.wc = &fakeWorkingCopy{files: make(map[string][]byte)}
	}
	return c.wc, nil
}

type fakeReleaseRepo struct {
	last *model.Release
}

func (r *fakeReleaseRepo) Create(_ context.Context, rel *model.Release) (int64, error) {
	r.last = rel
	return 42, nil
}

func TestEnginePublish_InvalidGitOpsConfig(t *testing.T) {
	r := &fakeRenderer{}
	g := &fakeGitClient{}
	repo := &fakeReleaseRepo{}

	engine := NewEngine(r, g, repo,
		filepath.Join("templates", "gitops", "app-argocd.yaml.tpl"),
		"argocd",
		"default",
		"https://kubernetes.default.svc",
	)

	_, err := engine.Publish(context.Background(), PublishInput{
		GitOpsConfig: ConfigSnapshot{},
	})
	if err == nil {
		t.Fatalf("expected error for invalid GitOpsConfig, got nil")
	}
}

func TestEnginePublish_HappyPath(t *testing.T) {
	fr := &fakeRenderer{
		output: &render.Output{
			WorkloadYAML: "workload: test\n",
			ResourceYAML: "resources: test\n",
		},
	}
	gc := &fakeGitClient{}
	repo := &fakeReleaseRepo{}

	appTplPath := filepath.Join("..", "..", "..", "templates", "gitops", "app-argocd.yaml.tpl")
	engine := NewEngine(fr, gc, repo,
		appTplPath,
		"argocd",
		"default",
		"https://kubernetes.default.svc",
	)

	in := PublishInput{
		BusinessUnitID:  1,
		DeployPlanID:    2,
		CDConfigID:      3,
		ProjectKey:      "proj",
		ServiceName:     "svc",
		Env:             "prod",
		BuildArtifactID: 4,
		ImageRef:        "registry.example.com/project/service:v1",
		Instance: render.InstanceConfigSnapshot{
			InstanceType: "deployment",
			Spec: map[string]any{
				"template": map[string]any{
					"spec": map[string]any{
						"containers": []any{
							map[string]any{
								"name":  "app",
								"image": "nginx:old",
							},
						},
					},
				},
			},
		},
		Strategy: render.ReleaseStrategySnapshot{
			DeploymentMode: "rolling",
			BatchRule: render.BatchRuleSnapshot{
				BatchCount:  3,
				BatchRatio:  []float64{0.1, 0.3, 0.6},
				TriggerType: "auto",
				Interval:    10,
			},
			CanaryTrafficRule: &render.CanaryTrafficRuleSnapshot{
				TrafficBatchCount: 2,
				TrafficRatioList:  []float64{0.2, 0.8},
				ManualAdjust:      true,
				AdjustTimeout:     60,
			},
		},
		GitOpsConfig: ConfigSnapshot{
			RepoURL:      "https://github.com/example/gitops.git",
			Branch:       "main",
			AppRoot:      "apps",
			ManifestRoot: "manifests",
		},
	}

	res, err := engine.Publish(context.Background(), in)
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	if res == nil {
		t.Fatalf("Publish returned nil result")
	}
	if res.ReleaseID != 42 {
		t.Errorf("unexpected ReleaseID: %d", res.ReleaseID)
	}

	// Verify Git client usage
	if gc.lastRepoURL != in.GitOpsConfig.RepoURL {
		t.Errorf("git repoURL = %q, want %q", gc.lastRepoURL, in.GitOpsConfig.RepoURL)
	}
	if gc.lastBranch != in.GitOpsConfig.Branch {
		t.Errorf("git branch = %q, want %q", gc.lastBranch, in.GitOpsConfig.Branch)
	}

	manifestPath := filepath.Join(in.GitOpsConfig.ManifestRoot, in.ProjectKey, in.Env, in.ServiceName)
	appPath := filepath.Join(in.GitOpsConfig.AppRoot, in.ProjectKey, in.Env, in.ServiceName+".yaml")

	wc := gc.wc
	if wc == nil {
		t.Fatalf("working copy is nil")
	}

	// workload.yaml
	workloadFile := filepath.Join(manifestPath, "workload.yaml")
	if string(wc.files[workloadFile]) != fr.output.WorkloadYAML {
		t.Errorf("workload.yaml content = %q, want %q", string(wc.files[workloadFile]), fr.output.WorkloadYAML)
	}

	// resources.yaml
	resourcesFile := filepath.Join(manifestPath, "resources.yaml")
	if string(wc.files[resourcesFile]) != fr.output.ResourceYAML {
		t.Errorf("resources.yaml content = %q, want %q", string(wc.files[resourcesFile]), fr.output.ResourceYAML)
	}

	// app yaml
	appFileContent, ok := wc.files[appPath]
	if !ok {
		t.Fatalf("app yaml not written at path %q", appPath)
	}

	var appManifest map[string]any
	if err := yaml.Unmarshal(appFileContent, &appManifest); err != nil {
		t.Fatalf("failed to unmarshal app yaml: %v", err)
	}

	if kind, _ := appManifest["kind"].(string); kind != "Application" {
		t.Errorf("app kind = %q, want %q", kind, "Application")
	}

	metadata, ok := appManifest["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("app.metadata is not a map: %#v", appManifest["metadata"])
	}
	if name, _ := metadata["name"].(string); name == "" {
		t.Errorf("app.metadata.name should not be empty")
	}

	spec, ok := appManifest["spec"].(map[string]any)
	if !ok {
		t.Fatalf("app.spec is not a map: %#v", appManifest["spec"])
	}
	source, ok := spec["source"].(map[string]any)
	if !ok {
		t.Fatalf("app.spec.source is not a map: %#v", spec["source"])
	}
	if repoURL, _ := source["repoURL"].(string); repoURL != in.GitOpsConfig.RepoURL {
		t.Errorf("app.spec.source.repoURL = %q, want %q", repoURL, in.GitOpsConfig.RepoURL)
	}
	if branch, _ := source["targetRevision"].(string); branch != in.GitOpsConfig.Branch {
		t.Errorf("app.spec.source.targetRevision = %q, want %q", branch, in.GitOpsConfig.Branch)
	}
	if path, _ := source["path"].(string); path != manifestPath {
		t.Errorf("app.spec.source.path = %q, want %q", path, manifestPath)
	}

	// Verify Release persisted via repo
	if repo.last == nil {
		t.Fatalf("release repo did not receive create call")
	}

	rel := repo.last
	if rel.BusinessUnitID != in.BusinessUnitID || rel.DeployPlanID != in.DeployPlanID || rel.CDConfigID != in.CDConfigID {
		t.Errorf("release ids mismatch: got BU=%d, Plan=%d, CD=%d", rel.BusinessUnitID, rel.DeployPlanID, rel.CDConfigID)
	}
	if rel.BuildArtifactID != in.BuildArtifactID || rel.ImageRef != in.ImageRef {
		t.Errorf("release artifact mismatch: got artifact=%d image=%q", rel.BuildArtifactID, rel.ImageRef)
	}
	if rel.WorkloadYAML != fr.output.WorkloadYAML || rel.ResourceYAML != fr.output.ResourceYAML {
		t.Errorf("release YAML mismatch")
	}
	if rel.RendererType != "gitops-k8s-native" {
		t.Errorf("release.RendererType = %q, want %q", rel.RendererType, "gitops-k8s-native")
	}
	if rel.EngineType != "gitops" {
		t.Errorf("release.EngineType = %q, want %q", rel.EngineType, "gitops")
	}

	// Verify ReleaseStrategy snapshot
	if rel.ReleaseStrategy.DeploymentMode != in.Strategy.DeploymentMode {
		t.Errorf("release strategy deployment mode = %q, want %q", rel.ReleaseStrategy.DeploymentMode, in.Strategy.DeploymentMode)
	}
	if rel.ReleaseStrategy.BatchRule.BatchCount != in.Strategy.BatchRule.BatchCount {
		t.Errorf("release batch count = %d, want %d", rel.ReleaseStrategy.BatchRule.BatchCount, in.Strategy.BatchRule.BatchCount)
	}
	if rel.ReleaseStrategy.BatchRule.Interval != in.Strategy.BatchRule.Interval {
		t.Errorf("release batch interval = %d, want %d", rel.ReleaseStrategy.BatchRule.Interval, in.Strategy.BatchRule.Interval)
	}
	if len(rel.ReleaseStrategy.BatchRule.BatchRatio) != len(in.Strategy.BatchRule.BatchRatio) {
		t.Errorf("release batch ratio length = %d, want %d", len(rel.ReleaseStrategy.BatchRule.BatchRatio), len(in.Strategy.BatchRule.BatchRatio))
	}

	if rel.ReleaseStrategy.CanaryTrafficRule == nil {
		t.Fatalf("release canary traffic rule is nil")
	}
	if rel.ReleaseStrategy.CanaryTrafficRule.TrafficBatchCount != in.Strategy.CanaryTrafficRule.TrafficBatchCount {
		t.Errorf("release canary batch count = %d, want %d", rel.ReleaseStrategy.CanaryTrafficRule.TrafficBatchCount, in.Strategy.CanaryTrafficRule.TrafficBatchCount)
	}

	// Verify GitOpsSnapshot
	if rel.GitOpsSnapshot.RepoURL != in.GitOpsConfig.RepoURL || rel.GitOpsSnapshot.Branch != in.GitOpsConfig.Branch {
		t.Errorf("release GitOpsSnapshot repo/branch mismatch")
	}
	if rel.GitOpsSnapshot.ManifestPath != manifestPath {
		t.Errorf("release ManifestPath = %q, want %q", rel.GitOpsSnapshot.ManifestPath, manifestPath)
	}
	if rel.GitOpsSnapshot.AppPath != appPath {
		t.Errorf("release AppPath = %q, want %q", rel.GitOpsSnapshot.AppPath, appPath)
	}
	if rel.GitOpsSnapshot.AppName == "" {
		t.Errorf("release AppName should not be empty")
	}
}
