package vo

type ExecuteReleaseReq struct {
	DeployPlanID    int64 `json:"deploy_plan_id" binding:"required"`
	BuildArtifactID int64 `json:"build_artifact_id" binding:"required"`
}
