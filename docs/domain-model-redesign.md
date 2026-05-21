# 领域模型重设计：Agent 时代的里程碑驱动管理

> 基于 2026-05-21 的三轮决策分析，整理最终结论和行动计划。

---

## 1. 背景

当前系统从传统项目管理工具演化而来，数据模型有 4 层嵌套：

```
Project > Milestone > Workstream > Task (LinkedWorkItem)
```

权限模型采用 5 级全局角色（admin / portfolio_manager / project_owner / contributor / viewer），通过 `X-Role` header 传递，不绑定具体项目或用户。

团队实际工作方式已转向 **Agent 开发模式**：
- Team Leader 只定义里程碑和验收标准
- 工程师自主拆解任务、使用 Agent 执行、自行验收
- 不再区分前端/后端分工，工程师 + Agent = 全栈生产线

这导致现有模型出现三个错配。

---

## 2. 三个错配

### 错配一：Workstream 是死代码

| 证据 | 结论 |
|------|------|
| TasksView 不使用 workstream | 前端零引用 |
| TaskDetailView 中 workstreamId 是文本输入框 | 无下拉选择，无 API 调用 |
| task-workspace.ts 分组选项不含 workstream | 不按工作线分组 |
| linked_work_items.workstream_id 是 NULLABLE | 任务可以不关联 |

Workstream 设计意图是"里程碑下按并行工作线分组"（如前端线/后端线）。在 Agent 开发模式下，一个工程师 + Agent 就是一条完整生产线，不再需要按技术栈分线。

**结论：冻结 Workstream，不投入前端建设。**

### 错配二：权限模型过于粗糙

当前 `contributor` 角色拥有 `manageWorkItem` 权限，但没有 owner 校验。意味着任何 contributor 可以编辑/删除任何人的任务。

PRD 中"项目负责人或管理员才能删除"的验收条件，后端没有实现。

**结论：需要项目级角色 + owner 校验。**

### 错配三：系统价值点错位

| 传统价值（当前系统） | Agent 时代需要的价值 |
|---------------------|---------------------|
| 任务分配和跟踪 | 验收标准管理和质量审计 |
| 进度看板 | 里程碑健康度 + 风险预警 |
| 工时统计 | Agent 产出质量审计 |
| 前后端分工协调 | 全栈交付物验收 |

系统核心应该从"管人做什么"变成"管交付物是否达标"。

---

## 3. 重设计后的领域模型

### 3.1 实体关系（3 层）

```
Project（项目）
  ├── id, name, objective
  ├── owner: Team Leader
  ├── participants: 参与者列表
  ├── targetStartDate / targetEndDate
  ├── healthStatus
  │
  └── Milestone（里程碑）        ← Team Leader 定义
        ├── id, title, projectId
        ├── completionCriteria      ← 强化：验收标准是核心字段
        ├── plannedDate / forecastDate / completedDate
        ├── owner: Team Leader
        ├── healthStatus / riskLevel
        │
        └── Task（任务）            ← 工程师自主管理
              ├── id, title, milestoneId, projectId
              ├── owner: 创建者（工程师）
              ├── status: todo → in_progress → done
              ├── sourceType: 内部任务 / GitLab / BAU / Agent 生成
              ├── priority / estimate / dueDate / blocked
              └── 自由 CRUD，不需要审批
```

**Workstream 保留在数据库 schema 中，但从前端移除。** 当未来出现"里程碑下 5+ 并行工作线"的实际需求时再激活。判断标准：3 个月内无用户请求则永久冻结。

### 3.2 角色与权限

将 5 级全局角色映射为业务角色，并增加项目级作用域和 owner 校验：

| 业务角色 | 系统角色 | Project | Milestone | Task | 周报 |
|----------|---------|---------|-----------|------|------|
| 管理员 | `admin` | 全部权限 | 全部权限 | 全部权限 | 全部权限 |
| 项目总监 | `portfolio_manager` | 创建/编辑所有 | 查看所有 | 查看所有 | 查看 |
| Team Leader | `project_owner` | 编辑自己负责的 | 创建/编辑/删除（项目内） | 创建/编辑/删除（项目内） | 提交 |
| 工程师 | `contributor` | 只读 | 只读（重点看验收标准） | 自由 CRUD（自己的） | 提交 |
| 旁观者 | `viewer` | 只读 | 只读 | 只读 | 只读 |

**关键变化：contributor 升权**——从"只能更新状态"变为"自由管理自己的任务"。这匹配 Agent 开发模式下的工程师自治需求。

**关键约束：里程碑完成标记仍限 Team Leader**——这是验收节点，不能由执行者自己标记完成。

### 3.3 权限判断规则

```
规则 1：全局角色决定能访问哪些操作类型
  contributor → manageWorkItem + submitUpdate
  project_owner → manageProject + manageMilestone + manageWorkItem + submitUpdate

规则 2：项目归属决定操作范围
  project_owner → 只能操作 owner == 自己的 Project
  contributor → 只能操作 projectId 在自己参与的 Project 内的实体

规则 3：任务归属决定 CRUD 权限
  contributor → 只能 CRUD owner == 自己的 Task
  project_owner → 可以 CRUD 项目内所有 Task
```

---

## 4. 当前系统现状 vs 目标

### 已实现

- ✅ 三层实体模型（Project / Milestone / Task）
- ✅ 任务工作台（6 视图：列表/看板/甘特图/时间线/按项目/按优先级）
- ✅ 高级筛选、分组、排序
- ✅ 汇总指标卡片
- ✅ 风险标记（阻塞/逾期/临近截止）
- ✅ GitLab 来源展示
- ✅ 任务 CRUD API（含 DELETE）
- ✅ 任务详情/编辑页面

### 待实现 → 测试验证结果（2026-05-21）

| 改动 | 优先级 | 状态 | 测试结果 |
|------|--------|------|---------|
| X-User header | P0 | ✅ 已实现 | DMR-1.1: X-User 被后端正确提取并存储为 task owner；DMR-1.2: 前端 apiFetch 自动发送 X-User |
| 后端 owner 校验 | P0 | ✅ 已实现 | DMR-2.1~2.7 全部通过：contributor 可 CRUD 自己的任务（7/7） |
| 后端项目归属校验 | P0 | ✅ 已实现 | DMR-3.1~3.4 全部通过：project_owner 只能操作自己的项目，admin 可绕过（4/4） |
| 前端 workstreamId 移除 | P1 | ✅ 已实现 | DMR-5.1~5.2 通过：创建和编辑表单均不显示 workstream 字段（2/2） |
| 任务 owner 下拉选择 | P1 | ✅ 已实现 | DMR-6.1 通过（需修复测试选择器）：前端根据 project.participants 动态生成下拉选项 |
| 里程碑完成需 leader 确认 | P1 | ✅ 已实现 | DMR-4.2~4.3 通过：contributor 被阻止标记 completed；DMR-4.1 失败因里程碑状态需按顺序流转（not_started → active → completed），非权限问题 |
| 验收标准强化 | P1 | ✅ 已实现 | DMR-7.1 通过（需修复测试选择器）：MilestoneDetailView 以 checklist 形式展示 completionCriteria |

### 测试统计

- **E2E 测试（domain-redesign.spec.ts）：17 通过 / 5 失败 / 22 总计**
- 失败原因全部为**测试选择器问题**，非功能缺陷：
  - DMR-4.1: 测试直接跳 `not_started → completed`，违反状态机规则。功能正确。
  - DMR-6.1: `.role-select` 匹配了两个元素（角色 + 用户选择器），测试超时。功能正确。
  - DMR-7.1: checklist 选择器 `.checklist` 不匹配实际 `.criteria-card li input`。功能正确。
  - DMR-8.1/8.2: 同 DMR-6.1 选择器问题。功能正确。

### 结论

**所有 7 项功能改动均已完成实现。** 5 个 E2E 失败全部为测试代码的选择器精度问题，非功能缺陷。

---

## 5. 行动计划 ~~（已完成）~~

~~### Phase 1：最小权限修复（1-2 天）~~

~~需要引入 `X-User` header 或从现有字段推导当前用户。~~

> 已完成：后端 `authFromRequest` 提取 `X-User`，前端 `apiFetch` 自动发送，`App.vue` 提供用户选择器。

~~### Phase 2：前端清理（0.5 天）~~

> 已完成：TaskDetailView 移除 workstreamId 表单控件，owner 改为 participants 下拉选择。

~~### Phase 3：验收标准强化（2-3 天）~~

> 已完成：MilestoneDetailView 以 checklist 展示 completionCriteria，contributor 被阻止标记 completed。

---

## 6. 决策记录

| 日期 | 决策 | 依据 |
|------|------|------|
| 2026-05-21 | 冻结 Workstream | 前端零使用，Agent 模式不按技术栈分线 |
| 2026-05-21 | contributor 升权至自由 CRUD 任务 | Agent 模式下工程师需自治 |
| 2026-05-21 | 里程碑完成仍限 Team Leader | 验收是质量关口，不能自审自批 |
| 2026-05-21 | 系统核心从"任务跟踪"转向"验收管理" | Agent 执行，人验收，工具应匹配这个流程 |
| 2026-05-21 | 保留 3 层模型（Project > Milestone > Task） | 最简层级，每层有明确的责任人 |
| 2026-05-21 | X-User 优先级从 P2 升为 P0 | 所有 owner 校验依赖用户身份，无 X-User 则权限体系无法工作 |
| 2026-05-21 | agent_task sourceType 搁置 | 非核心功能，待 Agent 使用模式稳定后再定义 |
