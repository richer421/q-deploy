package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Release 发布单 - 一次发布操作的完整记录
// 与 q-metahub 实体体系对齐：BusinessUnit → DeployPlan → CDConfig
type Release struct {
	BaseModel

	// ========== 归属关联（索引，追溯到元数据体系） ==========
	BusinessUnitID int64 `gorm:"column:business_unit_id;not null;index" json:"business_unit_id"` // 归属业务单元
	DeployPlanID   int64 `gorm:"column:deploy_plan_id;not null;index" json:"deploy_plan_id"`     // 归属部署计划
	CDConfigID     int64 `gorm:"column:cd_config_id;not null;index" json:"cd_config_id"`         // 归属 CD 配置

	// ========== 产物信息（记录本次发布使用的构建产物） ==========
	BuildArtifactID int64  `gorm:"column:build_artifact_id;not null;index" json:"build_artifact_id"`          // 关联构建产物（来自 q-ci）
	ImageRef        string `gorm:"column:image_ref;type:varchar(512);not null" json:"image_ref"`              // 完整镜像引用（registry/repo:tag 或 registry/repo@sha256:xxx）

	// ========== 发布核心 ==========
	Version      int64         `gorm:"column:version;not null;uniqueIndex:idx_plan_version" json:"version"`                   // 发布版本号（同一 DeployPlan 下唯一）
	Status       ReleaseStatus `gorm:"column:status;type:varchar(16);not null;default:'pending';index" json:"status"`         // 发布状态
	RenderedYAML string        `gorm:"column:rendered_yaml;type:text;not null" json:"rendered_yaml"`                          // 渲染后的最终 YAML

	// ========== 引擎与策略（发布时快照，不随 CDConfig 变更而变） ==========
	RendererType    string          `gorm:"column:renderer_type;type:varchar(32);not null" json:"renderer_type"`            // 渲染引擎（helm/kustomize/go_template）
	EngineType      string          `gorm:"column:engine_type;type:varchar(32);not null" json:"engine_type"`                // 工作引擎（kubernetes/docker/ssh）
	ReleaseStrategy ReleaseStrategy `gorm:"column:release_strategy;type:json;not null" json:"release_strategy"`             // 发布策略快照（滚动/蓝绿/金丝雀）
}

func (Release) TableName() string {
	return "releases"
}

// ReleaseStatus 发布状态枚举
type ReleaseStatus string

const (
	ReleaseStatusPending ReleaseStatus = "pending" // 待发布
	ReleaseStatusRunning ReleaseStatus = "running" // 发布中
	ReleaseStatusSuccess ReleaseStatus = "success" // 发布成功
	ReleaseStatusFailed  ReleaseStatus = "failed"  // 发布失败
)

// ReleaseStrategy 发布策略快照（从 CDConfig 快照而来）
type ReleaseStrategy struct {
	DeploymentMode    string             `json:"deployment_mode"`                // rolling/blue_green/canary
	BatchRule         BatchRule          `json:"batch_rule"`                     // 分批规则
	CanaryTrafficRule *CanaryTrafficRule `json:"canary_traffic_rule,omitempty"`  // 金丝雀流量规则
}

// BatchRule 分批规则
type BatchRule struct {
	BatchCount  int       `json:"batch_count"`  // 总批次
	BatchRatio  []float64 `json:"batch_ratio"`  // 每批实例比例
	TriggerType string    `json:"trigger_type"` // auto/manual
	Interval    int       `json:"interval"`     // 批次间隔（秒）
}

// CanaryTrafficRule 金丝雀流量规则
type CanaryTrafficRule struct {
	TrafficBatchCount int       `json:"traffic_batch_count"` // 流量分批数
	TrafficRatioList  []float64 `json:"traffic_ratio_list"`  // 每批流量比例
	ManualAdjust      bool      `json:"manual_adjust"`       // 是否允许手动调整
	AdjustTimeout     int       `json:"adjust_timeout"`      // 超时（秒）
}

// ========== ReleaseStrategy 数据库序列化 ==========

func (s ReleaseStrategy) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *ReleaseStrategy) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ReleaseStrategy: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, s)
}
