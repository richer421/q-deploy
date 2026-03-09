# 核心抽象

业务层面的关键抽象与领域概念。

## 核心工作流程

系统采用两阶段流水线架构：

```
阶段一：配置渲染
服务配置（来自元数据管理中心） → 选择渲染器 → YAML 渲染结果

阶段二：执行发布
YAML 渲染结果 → 工作引擎 → 发布/回滚/更新
```

### 阶段一：配置渲染

将抽象的服务配置（外部输入）转换为具体的部署 YAML。

| 输入 | 处理 | 输出 |
|------|------|------|
| 服务配置（外部） | 渲染器（Helm/Kustomize/自定义） | 渲染后的 YAML |

### 阶段二：执行发布

将渲染结果应用到目标环境。

| 操作类型 | 包产物 | 配置 | 场景 |
|----------|--------|------|------|
| 发布 | 变更 | 变更 | 新版本上线 |
| 更新 | 不变 | 变更 | 配置调整（扩缩容、环境变量等） |
| 回滚 | 回退 | 回退 | 版本回退 |

## 发布操作语义

三种操作的本质区别：

- **发布（Deploy）**: 产物流 + 配置流均变更
- **更新（Update）**: 产物流不变，仅配置流变更
- **回滚（Rollback）**: 产物流 + 配置流均回退到历史版本

## 渲染器抽象

渲染器是可插拔的模板引擎：

```
type Renderer interface {
    Render(config ServiceConfig) ([]byte, error)
}
```

支持多种实现：
- Helm Renderer
- Kustomize Renderer
- Go Template Renderer
- 自定义 Renderer

## 工作引擎抽象

工作引擎执行实际的发布操作：

```
type Engine interface {
    Deploy(yaml []byte) error
    Update(yaml []byte) error
    Rollback(version string) error
}
```

支持多种目标：
- Kubernetes Engine
- Docker Engine
- SSH Engine（物理机/虚拟机）
