package vo

// ReleaseWithGitOpsCmd GitOps 发布命令入参
// 假定调用方已经选定 DeployPlan 和 BuildArtifact

import (
	"github.com/richer421/q-deploy/domain/engine/gitops"
	render "github.com/richer421/q-deploy/domain/render"
)

type ReleaseWithGitOpsCmd struct {
	BusinessUnitID int64  `json:"business_unit_id" binding:"required"`
	DeployPlanID   int64  `json:"deploy_plan_id" binding:"required"`
	CDConfigID     int64  `json:"cd_config_id" binding:"required"`
	ProjectKey     string `json:"project_key" binding:"required"`
	ServiceName    string `json:"service_name" binding:"required"`
	Env            string `json:"env" binding:"required"`

	BuildArtifactID int64  `json:"build_artifact_id" binding:"required"`
	ImageRef        string `json:"image_ref" binding:"required"`

	// 下面三块 Snapshot 正常是由 app 层从 q-metahub 聚合，这里先直接从入参拿快照，后续可以改成 ID 查询
	InstanceSnapshot render.InstanceConfigSnapshot  `json:"instance_snapshot"`
	StrategySnapshot render.ReleaseStrategySnapshot `json:"strategy_snapshot"`
	GitOpsSnapshot   gitops.ConfigSnapshot          `json:"gitops_snapshot"`
}

type ReleaseDTO struct {
	ReleaseID int64             `json:"release_id"`
	GitOps    GitOpsSnapshotDTO `json:"gitops"`
}

type GitOpsSnapshotDTO struct {
	RepoURL      string `json:"repo_url"`
	Branch       string `json:"branch"`
	ManifestPath string `json:"manifest_path"`
	AppPath      string `json:"app_path"`
	AppName      string `json:"app_name"`
}
