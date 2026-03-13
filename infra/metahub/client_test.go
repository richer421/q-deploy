package metahub

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/richer421/q-deploy/conf"
)

func TestMetahubClientGetDeployPlan(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/deploy-plans/6/full-spec", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"project":{"id":1,"name":"q-demo-project","repo_url":"https://github.com/richer421/q-demo.git"},"business_unit":{"id":2,"name":"q-demo"},"ci_config":{"id":3,"name":"q-demo-ci"},"cd_config":{"id":4,"name":"q-demo-cd","git_ops":{"enabled":true,"repo_url":"https://github.com/richer421/q-demo-gitops.git","branch":"main","app_root":"apps","manifest_root":"manifests"},"release_strategy":{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}},"instance_config":{"id":5,"name":"q-demo-dev","env":"dev","instance_type":"deployment","spec":{"deployment":{}},"attach_resources":{"services":{"q-demo":{"metadata":{"name":"q-demo"}}}}},"deploy_plan":{"id":6,"name":"q-demo-dev-plan","business_unit_id":2,"ci_config_id":3,"cd_config_id":4,"instance_config_id":5}}}`))
	}))
	defer srv.Close()

	conf.C.Metahub = conf.UpstreamServiceConfig{BaseURL: srv.URL, Timeout: 5}

	client := New()
	res, err := client.GetDeployPlan(t.Context(), 6)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int64(6), res.DeployPlan.ID)
	assert.Equal(t, "https://github.com/richer421/q-demo-gitops.git", res.CDConfig.GitOps["repo_url"])
	assert.Equal(t, "deployment", res.InstanceConfig.InstanceType)
}
