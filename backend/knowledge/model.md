# 核心数据模型

系统核心实体及其关系。新增实体时在此注册。

## BaseModel（通用基础字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键，自增 |
| CreatedAt | time.Time | 创建时间，自动填充 |
| UpdatedAt | time.Time | 更新时间，自动填充 |

## 实体清单

### Release（发布单）

- **表名**: releases
- **模块**: release
- **说明**: 一次发布操作的记录，保存完整快照确保可重复性

| 字段 | 类型 | 约束 | 说明                                |
|------|------|------|-----------------------------------|
| (BaseModel) | - | - | 嵌入通用基础字段                          |
| ServiceConfigRef | varchar(128) | NOT NULL, INDEX | 服务配置引用（用于追溯来源）                    |
| ConfigSnapshot | json | NOT NULL | 服务配置快照（保证一致性）                     |
| RendererType | varchar(32) | NOT NULL | 渲染引擎类型（helm/kustomize/go_template） |
| EngineType | varchar(32) | NOT NULL | 工作引擎类型（kubernetes/docker/ssh）     |
| Type | varchar(16) | NOT NULL | 操作类型（deploy/update/rollback）      |
| ArtifactType | varchar(32) | NOT NULL | 包产物类型（image/binary/archive）       |
| ArtifactRef | varchar(255) | - | 包产物引用（镜像地址/tag/文件路径）              |
| RenderedYAML | text | NOT NULL | 渲染后的 YAML                         |
| Status | varchar(16) | NOT NULL | 状态（pending/running/success/failed） |
| Version | int | NOT NULL | 版本号（用于回滚定位）                       |

- **设计说明**:
  - `ConfigSnapshot` 保存发布时的配置快照，确保回滚时配置一致
  - `RendererType` + `EngineType` 记录使用的引擎，支持异构环境
  - `ArtifactType` 区分不同产物类型，未来扩展虚拟机场景

- **关联**: Release 引用外部服务配置（元数据管理中心）

## 实体关系

```
[元数据管理中心]
      │
      │ 服务配置（外部）
      ▼
   Release ──────▶ 工作引擎 ──▶ 目标环境
      │
      ▼
   渲染器
```
