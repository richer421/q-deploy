package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/richer421/q-deploy/app/deploy/vo"
	"github.com/richer421/q-deploy/http/common"
)

type fakeReleaseService struct{}

func (f *fakeReleaseService) ExecuteRelease(_ context.Context, req *vo.ExecuteReleaseReq) (*vo.ReleaseDTO, error) {
	return &vo.ReleaseDTO{
		ReleaseID: req.DeployPlanID + req.BuildArtifactID + 61,
		GitOps: vo.GitOpsSnapshotDTO{
			RepoURL:      "https://github.com/richer421/q-demo-gitops.git",
			Branch:       "main",
			ManifestPath: "manifests/q-demo/dev/q-demo",
			AppPath:      "apps/q-demo/dev/q-demo.yaml",
			AppName:      "q-demo-q-demo-dev",
		},
	}, nil
}

func (f *fakeReleaseService) ReleaseWithGitOps(_ context.Context, _ *vo.ReleaseWithGitOpsCmd) (*vo.ReleaseDTO, error) {
	return &vo.ReleaseDTO{}, nil
}

func TestDeployAPIExecuteRelease(t *testing.T) {
	gin.SetMode(gin.TestMode)

	api := &DeployAPI{
		appSvc: &fakeReleaseService{},
	}

	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.POST("/releases/execute", api.ExecuteRelease)

	body, err := json.Marshal(vo.ExecuteReleaseReq{
		DeployPlanID:    6,
		BuildArtifactID: 21,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/releases/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp common.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Code)
	data := resp.Data.(map[string]any)
	assert.Equal(t, float64(88), data["release_id"])
}
