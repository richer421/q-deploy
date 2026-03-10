package model

import (
	"encoding/json"
)

// Release 发布单 - 一次发布操作的记录
// 保存完整快照确保可重复性
type Release struct {
	BaseModel
	ServiceConfigRef string          `gorm:"column:service_config_ref;type:varchar(128);not null;index;uniqueIndex:idx_config_version" json:"service_config_ref"` // 服务配置引用（用于追溯来源）
	ConfigSnapshot   json.RawMessage `gorm:"column:config_snapshot;type:json;not null" json:"config_snapshot"`                                                    // 服务配置快照（保证一致性）
	RendererType     string          `gorm:"column:renderer_type;type:varchar(32);not null" json:"renderer_type"`                                                 // 渲染引擎类型（helm/kustomize/go_template）
	EngineType       string          `gorm:"column:engine_type;type:varchar(32);not null" json:"engine_type"`                                                     // 工作引擎类型（kubernetes/docker/ssh）
	OperationType    string          `gorm:"column:operation_type;type:varchar(16);not null" json:"operation_type"`                                               // 操作类型（deploy/update/rollback）
	ArtifactType     string          `gorm:"column:artifact_type;type:varchar(32)" json:"artifact_type,omitempty"`                                                // 包产物类型（image/binary/archive），update 时可为空
	ArtifactRef      string          `gorm:"column:artifact_ref;type:varchar(255)" json:"artifact_ref,omitempty"`                                                 // 包产物引用（镜像地址/tag/文件路径）
	RenderedYAML     string          `gorm:"column:rendered_yaml;type:text;not null" json:"rendered_yaml"`                                                        // 渲染后的 YAML
	Status           ReleaseStatus   `gorm:"column:status;type:varchar(16);not null;index" json:"status"`                                                         // 状态
	Version          int             `gorm:"column:version;not null;uniqueIndex:idx_config_version" json:"version"`                                               // 版本号（用于回滚定位，同一 ServiceConfigRef 下唯一）
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
