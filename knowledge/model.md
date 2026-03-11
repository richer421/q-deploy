# 核心数据模型

系统核心实体及其关系。新增实体时在此注册。

## BaseModel（通用基础字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | int64 | 主键，自增 |
| CreatedAt | time.Time | 创建时间，自动填充 |
| UpdatedAt | time.Time | 更新时间，自动填充 |

## 实体清单

### Release（发布单）

- **表名**: releases
- **模块**: release
- **说明**: 一次发布操作的完整记录，与 q-metahub 实体体系对齐

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| BusinessUnitID | int64 | NOT NULL, INDEX | 归属业务单元 |
| DeployPlanID | int64 | NOT NULL, INDEX, UNIQUE(idx_plan_version) | 归属部署计划 |
| CDConfigID | int64 | NOT NULL, INDEX | 归属 CD 配置 |
| BuildArtifactID | int64 | NOT NULL, INDEX | 关联构建产物（来自 q-ci） |
| ImageRef | varchar(512) | NOT NULL | 完整镜像引用 |
| Version | int64 | NOT NULL, UNIQUE(idx_plan_version) | 发布版本号（同一 DeployPlan 下唯一） |
| Status | varchar(16) | NOT NULL, DEFAULT 'pending', INDEX | 发布状态 |
| WorkloadYAML | text | NOT NULL | 核心工作负载 YAML（Deployment/StatefulSet/Job 等） |
| ResourceYAML | text | - | 配套资源 YAML（ConfigMap/Secret/Service 等） |
| RendererType | varchar(32) | NOT NULL | 渲染引擎（helm/kustomize/go_template） |
| EngineType | varchar(32) | NOT NULL | 工作引擎（kubernetes/docker/ssh） |
| ReleaseStrategy | json | NOT NULL | 发布策略快照（滚动/蓝绿/金丝雀） |

**ReleaseStatus 枚举**：

| 值 | 说明 |
|------|------|
| pending | 待发布 |
| running | 发布中 |
| success | 发布成功 |
| failed | 发布失败 |

**ReleaseStrategy 结构**（JSON 字段，从 CDConfig 快照）：

| 字段 | 类型 | 说明 |
|------|------|------|
| DeploymentMode | string | rolling/blue_green/canary |
| BatchRule | object | 分批规则（BatchCount/BatchRatio/TriggerType/Interval） |
| CanaryTrafficRule | object | 金丝雀流量规则（仅 canary 模式） |

- **设计说明**:
  - 通过 BusinessUnitID/DeployPlanID/CDConfigID 三级索引追溯到 q-metahub 元数据体系
  - BuildArtifactID + ImageRef 记录本次发布使用的构建产物
  - DeployPlanID + Version 联合唯一索引，确保同一部署计划下版本号不重复
  - WorkloadYAML 为核心工作负载渲染结果，ResourceYAML 为配套资源渲染结果
  - **更新操作**：仅变更 ResourceYAML（配套资源），WorkloadYAML 和产物不变
  - **发布操作**：新产物 + 新 WorkloadYAML + 新 ResourceYAML
  - **回滚操作**：回退到历史版本的 WorkloadYAML + ResourceYAML
  - ReleaseStrategy 是发布时的策略快照，不随 CDConfig 后续变更而变

- **关联**:
  - Release N:1 BusinessUnit（q-metahub）
  - Release N:1 DeployPlan（q-metahub）
  - Release N:1 CDConfig（q-metahub）
  - Release N:1 BuildArtifact（q-ci）

## 实体关系

```
[q-metahub]                                    [q-ci]
      │                                           │
      ├── BusinessUnit                            │
      │       │                                   │
      │       └── DeployPlan                      │
      │               │                           │
      │               ├── CIConfig ──── 触发构建 ──▶ BuildArtifact
      │               │                                   │
      │               ├── CDConfig                        │
      │               │     │                             │
      │               │     ├── RenderEngine              │
      │               │     └── ReleaseStrategy           │
      │               │                                   │
      │               └── InstanceConfig                  │
      │                     │                             │
      │                     ├── Spec (工作负载)            │
      │                     └── AttachResources (配套资源) │
      │                                                   │
      └───────────────────────┬───────────────────────────┘
                              ▼
                           Release
                              │
                              ├── WorkloadYAML  ←── 渲染引擎(InstanceConfig.Spec)
                              └── ResourceYAML  ←── 渲染引擎(InstanceConfig.AttachResources)
```
