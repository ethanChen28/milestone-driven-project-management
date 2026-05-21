# 项目空间与递进对象平台调研

调研日期：2026-05-21

## 研究问题

目标对象关系为：项目 → 里程碑 → 工作项 → 周报 / 风险 / 依赖 / 外部 Issue。

核心问题不是“能不能显示层级”，而是：

- 左侧导航、顶部 Tab、右侧内容是否重复。
- 概览页应该显示多少下级对象。
- 工作项到底以项目为中心，还是以全局任务工作台为中心。
- 里程碑、风险、周报是独立对象，还是工作项上的属性或视图。

## 总体结论

成熟平台通常采用三层设计：

1. 对象模型：明确哪些是一等对象，例如 Project、Milestone、Issue、Portfolio。
2. 入口模型：用户从哪里进入，例如 workspace sidebar、project page、global task view、timeline plan。
3. 视图模型：同一批对象可以通过 list、board、timeline、calendar、saved view 展示。

避免重复的关键原则：

- 概览页只做 rollup 和摘要，不做完整 CRUD。
- 完整列表、筛选、批量操作应该放在专门 Tab 或专门视图里。
- 右侧详情面板用于属性、关联对象、快捷操作，不应替代主列表。
- 层级导航适合复杂对象关系，但不能无限展开，否则会变成 ClickUp/Jira 式配置负担。

## 平台案例

### 1. Plane

官方模型：Workspace → Project → Work Items，并在项目内打开 Cycles、Modules、Views、Pages、Intake、Time Tracking 等功能。

关键设计：

- Project 是一个空间，里面包含 work items、cycles、modules、pages 等资源。
- Project Overview 是集中式项目仪表盘，用于项目健康度、进度、更新和活动流。
- Cycles 是时间盒，类似 sprint。
- Modules 是项目内的功能块或小项目，用来组织工作项。
- Views 是保存下来的筛选、布局、排序配置，不改变底层数据。
- 项目功能可开关，避免所有项目都显示一堆不需要的入口。

对我们的启发：

- 可以保留“项目空间”概念。
- 但不要把里程碑、任务、风险、周报全部堆在概览页。
- 最像 Plane 的做法是：项目页有 Overview；完整对象入口是 Work Items / Milestones / Updates / Risks；保存视图用于常用筛选。

适合采用：A1 Tab 导航版，但要严格限制 Overview 的职责。

参考：

- Plane Manage projects: https://docs.plane.so/core-concepts/projects/overview
- Plane Project Overview: https://docs.plane.so/core-concepts/projects/project-overview
- Plane Modules: https://docs.plane.so/core-concepts/modules
- Plane Views: https://docs.plane.so/core-concepts/views

### 2. Linear

官方模型：Initiative → Project → Issue，可选 Project Milestone。Project 有 Overview、Issues、Documents、Milestones、Graph、Status、Dependencies 等项目能力。

关键设计：

- Initiative 是手动策划的一组 Projects，用于公司目标和长期规划。
- Project 是有明确结果或计划完成日期的大块工作，由 issues 和可选文档组成。
- Project Overview 展示项目摘要、属性、文档、描述、里程碑。
- Issues 是完整执行列表。
- Milestone 可以在 overview、details pane、timeline、initiative view 中以不同密度出现。
- Milestone 在高层 timeline 中只是上下文和进度提示；双击后进入项目并应用里程碑 quick filter。
- Project 内可挂自定义 issue views，作为项目 Tab，类似 Overview / Issues 旁边的“保存视图”。

对我们的启发：

- 这是处理“Tab 与内容重复”的最佳案例。
- Linear 允许概览、详情侧栏、时间线都出现 milestone，但它们的交互密度不同：
  - Overview：展示与创建里程碑。
  - Details pane：属性和快捷过滤。
  - Timeline：上下文和进度。
  - Issues view：完整工作项列表，可按 milestone 过滤或分组。
- 重复不是绝对错误，问题是重复的职责不能一样。

适合采用：A1 和 A2 的折中。项目内保留 Overview / Issues / Updates / Risks，但里程碑既是摘要组件，也是过滤维度。

参考：

- Linear Projects: https://linear.app/docs/projects
- Linear Project milestones: https://linear.app/docs/project-milestones
- Linear Initiatives: https://linear.app/docs/initiatives

### 3. Jira / Advanced Roadmaps

官方模型：默认 hierarchy 是 Epic(level 1) → Story(level 0) → Subtask(level -1)。Premium/Enterprise 可以增加 Epic 以上层级，例如 Initiative，并在 Plans 中做跨团队长期规划。

关键设计：

- Jira 把执行对象都叫 work item/issue，通过 issue type 和 parent 字段组织层级。
- Advanced Roadmaps 的 Plan 是跨项目/跨空间的规划视图，不是普通项目页。
- Plan 可以保存多种预配置视图：Basic、Dependency management、Top-level planning、Sprint capacity。
- Dependency view 只关注有依赖的对象，避免普通任务列表被复杂依赖线污染。
- 层级配置需要管理员，改动会影响现有父子关系，成本较高。

对我们的启发：

- 不建议一开始做 Jira 式可配置层级，成本高且容易过度抽象。
- 但可以借鉴它的“不同问题用不同视图”：
  - 普通交付：项目/里程碑/工作项列表。
  - 依赖风险：专门的 dependency/risk view。
  - 高层规划：roadmap/portfolio timeline。
- 如果将来支持 Portfolio/Initiative，不要塞进项目页，应放到全局路线图或组合视图。

适合采用：A2 中的对象关系，但要克制，不做完全自由层级。

参考：

- Jira work type hierarchy: https://support.atlassian.com/jira-cloud-administration/docs/configure-the-issue-type-hierarchy/
- Jira custom hierarchy in Advanced Roadmaps: https://support.atlassian.com/jira-software-cloud/docs/configure-custom-hierarchy-levels-in-advanced-roadmaps/
- Jira Plans: https://support.atlassian.com/jira-software-cloud/docs/what-is-advanced-roadmaps/
- Jira preconfigured plan views: https://support.atlassian.com/jira-software-cloud/docs/preconfigured-views-in-advanced-roadmaps/

### 4. Asana

官方模型更像 Work Graph，而不是单一树。Task 可以属于多个 Project；Project 属于 Team；Portfolio 是 Projects 或 Portfolios 的集合；Subtask 是 Task 的子对象。

关键设计：

- Project 是任务集合，可用 list、board、timeline、calendar 查看。
- Portfolio 是项目集合，用于高层状态和项目级字段。
- Task 可以 multi-home 到多个 project。
- Subtask 最多 5 层，但官方不建议做太深的 sub-subtasks。
- Section 是项目内任务分组，可用于阶段、优先级、流程状态。

对我们的启发：

- 不要把所有关系都设计成严格父子树。
- 工作项可以属于项目，也可以被里程碑、风险、周报引用。
- 里程碑可以是项目内阶段，也可以是工作项过滤维度。
- “周报”更像项目/里程碑状态更新，不一定是任务层级中的子节点。

适合采用：对象图谱/关系模型，UI 上不要暴露成复杂图谱，先用列表和侧栏呈现。

参考：

- Asana Object hierarchy: https://developers.asana.com/docs/object-hierarchy

### 5. Notion Projects / Databases

官方模型：数据库页面 + 多视图 + sub-items + dependencies。Notion 不强制项目管理层级，而是让用户在数据库中配置 sub-items 和依赖。

关键设计：

- Sub-items 给任务增加深度，父子项在所有数据库视图里可见。
- 不同视图对 sub-items 的显示不同：Nested in toggle、Flattened list、Card property。
- Filter options 也可控制显示 parents only、parents and sub-items、sub-items only。
- Timeline 是数据库的一种视图，需要 date property。

对我们的启发：

- 对同一批对象，提供“嵌套 / 平铺 / 只看子项”切换，可以解决很多重复和层级困扰。
- 甘特图、时间线不一定要成为对象入口，它可以只是一个 view。
- 如果做 A2 对象树，必须提供平铺视图，否则执行团队会迷路。

适合采用：工作项列表中的层级显示模式。

参考：

- Notion Sub-items & dependencies: https://www.notion.com/en-gb/help/tasks-and-dependencies
- Notion Timeline view: https://www.notion.com/help/timelines

### 6. ClickUp

官方模型：Workspace → Space → Folder → List → Task → Subtask。ClickUp 是最典型的显式层级产品。

关键设计：

- Folders 可选，适合复杂 workflow。
- Lists 直接包含 Tasks。
- Tasks 必须至少属于一个 List，也可加入多个 Lists。
- Task 顶部显示 Space / Folder / List breadcrumb。
- 可以从 sidebar 展开 Space、Folder、List。
- Nested subtasks 可用于复杂项目，但需要开启。

对我们的启发：

- 层级非常清晰，但产品负担也最大。
- 我们不应照搬 Workspace/Space/Folder/List，因为当前对象已经是 Project/Milestone/WorkItem。
- 可以借鉴 breadcrumb、对象位置感、在详情页快速上下跳转父子对象。

适合采用：A2 对象导航版中的 breadcrumb 和对象树，但不建议照搬多层容器。

参考：

- ClickUp Intro to the Hierarchy: https://help.clickup.com/hc/en-us/articles/13856392825367-Intro-to-the-Hierarchy
- ClickUp Folders overview: https://help.clickup.com/hc/en-us/articles/6311450560407-Folders-overview
- ClickUp Task view overview: https://help.clickup.com/hc/en-us/articles/10552031987735-Task-View-3-0-overview

## 对当前产品的推荐方案

推荐不是纯 A1，也不是纯 A2，而是“项目空间 + 视图化对象”的混合：

### 信息架构

- 全局一级导航：仪表盘、项目、工作项、里程碑、路线图、周度回顾。
- 项目详情页二级导航：概览、工作项、里程碑、周报、风险、依赖、设置。
- 项目概览页：只显示摘要、rollup、最近更新、关键风险、当前里程碑。
- 工作项页：承载完整列表、看板、甘特、时间线、分组视图。
- 里程碑页：承载里程碑完整管理，可进入单个里程碑详情。
- 风险/依赖页：不是普通列表的重复，而是专门聚焦阻塞关系和决策项。

### 对象关系

建议一等对象：

- Project
- Milestone
- WorkItem
- WeeklyUpdate
- Risk
- Dependency
- ExternalIssueLink

建议关系：

- Project has many Milestones
- Project has many WorkItems
- Milestone has many WorkItems
- WeeklyUpdate belongs to Project, optionally Milestone
- Risk belongs to Project, optionally Milestone or WorkItem
- Dependency links WorkItem/Milestone to WorkItem/Milestone
- ExternalIssueLink maps to WorkItem

### 如何避免重复

- 概览页显示“当前里程碑摘要”，但点击后跳转里程碑详情或应用 quick filter。
- 工作项列表显示 milestone 字段，但不复制里程碑完整管理功能。
- 里程碑详情显示关联工作项，但不复制全局工作项工作台全部视图。
- 周报页按时间流展示更新；项目概览只展示最近 1-3 条。
- 风险页展示需要决策/阻塞的对象；项目概览只展示 Top risks。

## 原型选择建议

当前两版 HTML：

- `docs/prototypes/project-space-tabs.html`
- `docs/prototypes/project-space-object-nav.html`

建议下一版基于 A1 改造，而不是直接使用 A2：

- A1 的学习成本低，更适合当前阶段。
- 引入 A2 的对象树思想，但只用于项目详情左侧的“当前项目结构”小组件或详情面板，不作为主导航。
- 先做 Linear/Plane 式的项目空间，再逐步增加对象关系视图。

## 下一步设计动作

1. 把 A1 原型调整为推荐混合版：项目内 Tab + 右侧项目详情面板 + 当前里程碑 quick filter。
2. 定义每个 Tab 的职责边界，避免重复 CRUD。
3. 增加对象关系字段到 PRD：Risk、Dependency、WeeklyUpdate 的归属关系。
4. 在工作项页支持 `按里程碑分组`、`只看当前里程碑`、`显示/隐藏子项`。
5. 在项目概览页只展示摘要卡，不放完整表格。
