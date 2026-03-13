package metahub

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
	timeout := conf.C.Metahub.Timeout
	if timeout <= 0 {
		timeout = 10
	}
	return &Client{
		baseURL: conf.C.Metahub.BaseURL,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

type responseEnvelope struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *DeployPlanSpecDTO `json:"data"`
}

type ProjectDTO struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
}

type BusinessUnitDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type DeployPlanDTO struct {
	ID         int64 `json:"id"`
	CDConfigID int64 `json:"cd_config_id"`
}

type CDConfigDTO struct {
	ID              int64          `json:"id"`
	GitOps          map[string]any `json:"git_ops"`
	ReleaseStrategy map[string]any `json:"release_strategy"`
}

type InstanceConfigDTO struct {
	ID              int64          `json:"id"`
	Env             string         `json:"env"`
	InstanceType    string         `json:"instance_type"`
	Spec            map[string]any `json:"spec"`
	AttachResources map[string]any `json:"attach_resources"`
}

type DeployPlanSpecDTO struct {
	Project        ProjectDTO        `json:"project"`
	BusinessUnit   BusinessUnitDTO   `json:"business_unit"`
	CDConfig       CDConfigDTO       `json:"cd_config"`
	InstanceConfig InstanceConfigDTO `json:"instance_config"`
	DeployPlan     DeployPlanDTO     `json:"deploy_plan"`
}

func (c *Client) GetDeployPlan(ctx context.Context, deployPlanID int64) (*DeployPlanSpecDTO, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/deploy-plans/%d/full-spec", c.baseURL, deployPlanID), nil)
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
		return nil, fmt.Errorf("metahub get deploy plan failed: status=%d code=%d message=%s", resp.StatusCode, envelope.Code, envelope.Message)
	}

	return envelope.Data, nil
}
