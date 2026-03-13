package mcp

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deployvo "github.com/richer421/q-deploy/app/deploy/vo"
)

func TestMCPServerRegistersReleaseTools(t *testing.T) {
	server := NewServer()

	tools := server.toolSpecs()
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}

	assert.Contains(t, names, "read_logs")
	assert.Contains(t, names, "execute_release")
}

func TestExecuteReleaseToolReturnsStructuredResult(t *testing.T) {
	server := NewServer()
	payload := &deployvo.ReleaseDTO{
		ReleaseID: 88,
		GitOps: deployvo.GitOpsSnapshotDTO{
			RepoURL:      "https://github.com/richer421/q-demo-gitops.git",
			Branch:       "main",
			ManifestPath: "manifests/q-demo/dev/q-demo",
			AppPath:      "apps/q-demo/dev/q-demo.yaml",
			AppName:      "q-demo-q-demo-dev",
		},
	}

	res, err := server.jsonResult(payload)
	require.NoError(t, err)
	require.Len(t, res.Content, 1)

	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(text.Text), &decoded))
	assert.Equal(t, float64(88), decoded["release_id"])
	assert.Equal(t, "q-demo-q-demo-dev", decoded["gitops"].(map[string]any)["app_name"])
}
