# 核心数据模型

系统核心实体及其关系。新增实体时在此注册。

## BaseModel（通用基础字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键，自增 |
| CreatedAt | time.Time | 创建时间，自动填充 |
| UpdatedAt | time.Time | 更新时间，自动填充 |

## 实体清单

<!-- 新增实体时按以下格式注册 -->
<!-- 示例：
### User
- **表名**: users
- **模块**: user
- **说明**: 系统用户

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Username | varchar(64) | UNIQUE, NOT NULL | 用户名 |
| Email | varchar(255) | UNIQUE, NOT NULL | 邮箱 |

- **关联**: User 1:N Order
-->

## 实体关系

<!-- ER 关系概览 -->
