package qci

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/richer421/q-deploy/conf"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func New() *Client {
	timeout := conf.C.QCI.Timeout
	if timeout <= 0 {
		timeout = 10
	}
	return &Client{
		baseURL: conf.C.QCI.BaseURL,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

type responseEnvelope struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *ArtifactDTO `json:"data"`
}

type ArtifactDTO struct {
	ID           int64  `json:"id"`
	DeployPlanID int64  `json:"deploy_plan_id"`
	Status       string `json:"status"`
	ImageRef     string `json:"image_ref"`
	ImageTag     string `json:"image_tag"`
}

func (c *Client) GetArtifact(ctx context.Context, artifactID int64) (*ArtifactDTO, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/artifacts/%d", c.baseURL, artifactID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var envelope responseEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || envelope.Code != 0 || envelope.Data == nil {
		return nil, fmt.Errorf("qci get artifact failed: status=%d code=%d message=%s", resp.StatusCode, envelope.Code, envelope.Message)
	}

	return envelope.Data, nil
}
