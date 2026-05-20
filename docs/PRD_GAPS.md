# PRD/Design Gap Analysis - Milestone-Driven Project Management

**Date:** 2026-05-20
**Method:** Decision Clarity (Clarify → Deconstruct → Simplify)
**Scope:** PRD (`doc_*-prd.md`) vs Design (`openspec/.../design.md`) vs 当前实现

---

## 0. 核心发现

PRD 定义了 5 个用户角色和 5 个关键页面，design.md 做了 6 个架构决策并记录了 4 个 Open Questions。当前实现中后端 API 层 22/22 spec 场景通过，但 PRD 到实现之间存在 **6 个结构性断裂**——其中 3 个是未做的决策，3 个是实现与设计的不一致。

---

## 1. CRITICAL: 前端鉴权缺失，RBAC 在 UI 层完全失效

**PRD §11 定义了 5 种角色：** admin / portfolio_manager / project_owner / contributor / viewer
**Design 决策：** X-Role header 传递角色
**实际状态：** `frontend/src/api.ts:7` 硬编码 `X-Role: "admin"`

```typescript
// api.ts 第 7 行
headers: { "Content-Type": "application/json", "X-Role": "admin" },
```

**影响：**
- 每个浏览器用户都是 admin 权限
- PRD §11 要求 "only admins can manage GitLab integration settings"、"contributors can update linked internal tasks"，但 UI 上无任何角色区分
- 后端 60/60 RBAC 组合全部正确，但前端绕过了它们

**需要决策：** MVP 阶段是继续 header-based role（需要前端角色选择 UI）还是引入 JWT/OAuth2？

---

## 2. HIGH: 数据持久化空转

**事实链：**
- `docker-compose.yml` 配置了 MySQL 8.4 + Redis 7.4
- `infra/mysql/init/001_schema.sql` 创建了 11 张表（roadmap_periods, projects, milestones...）
- `infra/mysql/init/002_gitlab_sync_alerts.sql` 创建了 5 张表（gitlab_configs, sync_rules...）
- `backend/internal/service/store.go` 全部使用 `map[string]T` + `sync.RWMutex`
- `Config.MySQLDSN` 和 `Config.RedisAddr` 被加载但从未使用

**影响：** 每次后端重启（docker restart、重新部署），所有数据丢失。

**Design.md 未明确选择持久化方案。** MySQL schema 存在但未接入，形成了"看起来有但实际没用"的状态。

**需要决策：** 何时迁移到 MySQL？是 MVP 阶段就做还是等 UI 稳定后？

---

## 3. HIGH: 里程碑状态流转 UI 缺失

**PRD §8.3 要求：** "support milestone status such as `not_started`, `active`, `blocked`, `completed`, `cancelled`"
**PRD 核心原则 §4.1：** "Milestone first: every important piece of work should map to a milestone"

**实际状态：**
- `ProjectDetailView.vue` 只能创建 `not_started` 状态的里程碑
- `MilestoneDetailView.vue` 只展示里程碑信息，无状态切换按钮
- 没有编辑已有里程碑的 UI
- 后端已支持 `PUT /api/v1/milestones?id=X` 修改状态

**这是 PRD 的核心功能（"milestone-driven"），但用户无法在 UI 上推动里程碑从 not_started → active → completed。**

---

## 4. HIGH: 周报提交入口缺失

**PRD §8.6 要求：** "allow structured weekly updates"
**PRD §7.4 Review Workflow：** "Weekly project review shows milestone status"
**PRD §10.5 Review View：** "weekly updates grouped by owner or roadmap, unresolved decisions needed"

**实际状态：**
- `ReviewView.vue` 只展示 review 数据（延迟里程碑、阻塞里程碑、周报列表）
- 没有周报提交表单
- 后端 `POST /api/v1/weekly-updates` 完全可用，有 summary/progress/risk/blockers/decisionsNeeded/nextSteps 7 个字段

**影响：** PRD §12 的 "reminder when weekly update is missing" 依赖周报数据存在，没有提交入口就无法触发缺失提醒。

---

## 5. MEDIUM: Design.md 的 4 个 Open Questions 未正式回答

design.md 记录了 4 个开放问题，实现中隐式选择了最简单路径但未记录决策：

| # | Open Question | 隐式选择 | 问题 |
|---|--------------|---------|------|
| 1 | 内部任务需要更丰富的工作流吗？ | 选了轻量 work item | 未记录 |
| 2 | GitLab issue 能跨里程碑移动吗？ | 未实现移动功能 | 需要回答 |
| 3 | progress_percent 自动化程度？ | 完全手动 | 未记录 |
| 4 | 单 workspace 足够吗？ | 单 workspace | 未记录 |

**影响：** 后续开发者无法区分"有意的简化"和"遗漏"。

**建议：** 在 design.md 中新增 "Resolved Questions" 部分，记录每个问题的决策和理由。

---

## 6. MEDIUM: PRD §8.8 搜索/筛选不完整

**PRD 要求按以下维度筛选：** roadmap period, project, milestone, owner, team, status, health, risk, source type, GitLab context

**后端已实现：** owner, status, health, q (text search), periodId, projectId, milestoneId, sourceType
**缺失：**
- **team 筛选** — 没有团队数据模型，PRD §5.1-5.4 提到 team 但 domain model 没有 team 字段
- **risk 筛选** — Milestone 有 `riskLevel` 字段但后端 filter 函数没有处理
- **前端无筛选 UI** — 所有列表页没有搜索框或筛选下拉

---

## 7. MEDIUM: GitLab 集成只有后端骨架

**PRD §9 要求：**
- 连接 GitLab groups/repositories — 后端有数据模型
- 手动/自动关联 issues — 后端有 API
- 定时同步或 webhook — webhook handler 存在但不验证签名；无定时任务
- 在项目/里程碑视图中展示 GitLab issue 状态 — **前端无展示**
- 打开原始 GitLab issue — **前端无链接**

---

## 8. LOW: 设计决策的隐含假设未验证

### 8.1 "Read models from projections" 假设

Design.md 决定 "Build review and dashboard views from read models"，但当前用内存 map 实时遍历。对 MVP 没有问题，但如果迁移到 MySQL，所有聚合查询（PortfolioSummary、WeeklyReviewView、RoadmapOverview）需要重写为 SQL。

### 8.2 "X-Role header" 假设

Design.md 决定用 header 传递角色。这在 API 测试中工作良好，但在浏览器环境中无法保证——前端硬编码 admin 角色使得整个 RBAC 体系在前端无效。

---

## 9. 决策优先级

按阻塞其他工作的程度排序：

| 优先级 | 决策 | 阻塞项 |
|--------|------|--------|
| P0 | 鉴权策略（header-based role selector vs JWT） | 前端 RBAC、所有角色相关功能 |
| P1 | 持久化方案（何时迁移 MySQL） | 数据可靠性、生产部署 |
| P1 | 里程碑状态流转 UI | PRD 核心功能 |
| P2 | 周报提交 UI | Review workflow、告警触发 |
| P2 | 回答 4 个 Open Questions | design.md 完整性 |
| P3 | 搜索/筛选 UI | 用户体验 |
| P3 | GitLab 前端展示 | 工程师日常使用 |
