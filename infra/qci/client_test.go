package qci

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/richer421/q-deploy/conf"
)

func TestQCIClientGetArtifact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/artifacts/21", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"id":21,"deploy_plan_id":6,"status":"success","image_ref":"harbor.local/q-demo/q-demo:sha-123","image_tag":"sha-123"}}`))
	}))
	defer srv.Close()

	conf.C.QCI = conf.UpstreamServiceConfig{BaseURL: srv.URL, Timeout: 5}

	client := New()
	res, err := client.GetArtifact(t.Context(), 21)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int64(21), res.ID)
	assert.Equal(t, "harbor.local/q-demo/q-demo:sha-123", res.ImageRef)
}
