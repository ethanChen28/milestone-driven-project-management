# 项目空间对象视图设计说明

## 目标

项目空间用于管理 `Project -> Milestone -> WorkItem -> WeeklyUpdate/Risk/Dependency` 的递进对象关系，同时避免概览页、Tab 页和详情页重复承载完整 CRUD。

## 职责边界

- 项目概览：只展示 rollup、当前里程碑摘要、Top 风险、最近周报、阻塞/逾期信号。
- 工作项 Tab：项目内完整工作项列表、分组、筛选和 GitLab 上下文展示。
- 里程碑 Tab：项目内完整里程碑创建、状态流转、过滤和进入里程碑详情。
- 周报 Tab：项目内周报历史；跨项目对比仍由全局周度回顾承载。
- 风险 Tab：从高风险/阻塞里程碑、阻塞工作项、外部依赖、周报风险派生风险信号。
- 依赖 Tab：从里程碑依赖说明、阻塞工作项、`external_dependency` 工作项派生依赖信号。
- 设置 Tab：项目元数据和后续配置入口。

## API

新增 additive API：

`GET /api/v1/project-space?id=<projectId>`

返回：

- `project`: 项目元数据
- `milestones`: 项目里程碑
- `workItems`: 项目工作项
- `updates`: 项目周报，按创建时间倒序
- `rollups`: 里程碑/工作项/风险/依赖聚合计数
- `risks`: 派生风险信号
- `dependencies`: 派生依赖信号

扩展 `GET /api/v1/work-items` 查询参数：

- `projectId`
- `milestoneId`
- `sourceType`
- `owner`
- `status`
- `priority`
- `blocked=true|false`
- `overdue=true|false`
- `gitlabContext` / `gitLabContext` / `gitlabRepository` / `repository`

## 原型

- `docs/prototypes/project-space-tabs.html`: A1 Tab 导航版
- `docs/prototypes/project-space-object-nav.html`: A2 对象导航版
- `docs/prototypes/project-space-hybrid.html`: 当前推荐混合版
