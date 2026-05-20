## Context

The product is a standalone milestone-driven project management system that sits above GitLab. It must support cross-functional planning and review workflows without forcing engineering teams to abandon GitLab issue management. The design therefore needs to define a clean domain model, explicit source-of-truth boundaries, and read/write paths that keep milestone health understandable for non-engineering users.

Current state is a PRD only. There is no existing OpenSpec baseline for roadmap, project, milestone, review, or GitLab-sync behavior, so this design establishes the initial MVP architecture contract for implementation planning.

## Goals / Non-Goals

**Goals:**
- Model roadmap, project, milestone, workstream, linked work item, and weekly update as first-class domain objects with explicit ownership and lifecycle.
- Keep milestone status and health readable to product, engineering, and leadership while allowing linked execution signals to inform progress.
- Integrate GitLab as an external execution source without allowing GitLab data to overwrite PM-owned business fields.
- Provide review and dashboard surfaces that summarize milestone movement, blockers, and BAU versus milestone work.
- Support an MVP permission model and reminder/alert automation with low operational complexity.

**Non-Goals:**
- Rebuild GitLab issue boards, merge request review, or code deployment workflows inside this system.
- Implement advanced dependency graphing, budget/resource management, or OKR features in MVP.
- Infer milestone completion automatically from issue counts alone.
- Design a multi-tenant architecture beyond a single-workspace MVP unless later requirements demand it.

## Decisions

### Decision: Use PM-owned aggregate objects with GitLab-linked execution records

The core aggregate roots are `Roadmap`, `Project`, `Milestone`, `Workstream`, and `WeeklyUpdate`. `LinkedWorkItem` acts as a bridge object that can reference either internal tasks or GitLab issues. This keeps business planning separate from engineering execution and allows non-GitLab work to be represented without special cases.

Alternatives considered:
- Store GitLab issues directly under milestones without a linking entity.
  Rejected because it cannot represent BAU work, external dependencies, or internal business tasks consistently.
- Make milestone the only aggregate and hang projects/roadmaps off metadata.
  Rejected because it weakens portfolio-level ownership and reporting.

### Decision: Separate milestone status, health, and progress assistance

Milestones keep three related but distinct fields:
- `status`: lifecycle state such as `not_started`, `active`, `blocked`, `completed`, `cancelled`
- `health_status`: human judgment such as `on_track`, `at_risk`, `off_track`, `done`
- `progress_percent`: assisted value informed by linked work, but still editable through PM workflows

This prevents raw issue state from being mistaken for business outcome completion while still giving project owners useful automation.

Alternatives considered:
- Derive milestone completion entirely from linked issue closure.
  Rejected because many milestones include operational or business work and require completion criteria review.
- Use health as a computed field only.
  Rejected because teams need a manual override for risk signaling before execution data fully reflects the problem.

### Decision: Enforce source-of-truth boundaries at the field level

GitLab owns issue title, description, labels, assignee, state, and merge request references. The PM system owns roadmap structure, project/milestone definitions, ownership, completion criteria, manual health, and cross-functional work items. Sync jobs update only GitLab-owned fields on `LinkedWorkItem` records and write sync metadata such as last-seen timestamps and error states.

Alternatives considered:
- Allow PM users to edit synced issue fields locally.
  Rejected because it creates reconciliation conflicts and hides the canonical execution state.
- Mirror all GitLab data into PM-native issue tables with unrestricted edits.
  Rejected because it increases drift and makes webhook/scheduled sync behavior harder to reason about.

### Decision: Support both manual link and rule-based GitLab association

The integration model supports:
- manual issue linking by URL or issue identifier
- sync rules filtered by GitLab group, repository, label, assignee, milestone, or query
- scheduled sync plus webhook-triggered incremental refresh

This balances precise owner control with scalable onboarding for projects that already follow labeling discipline.

Alternatives considered:
- Manual linking only.
  Rejected because it does not scale and creates stale coverage quickly.
- Auto-linking only.
  Rejected because teams need a way to attach exceptions and fix noisy mapping rules.

### Decision: Build review and dashboard views from read models

Portfolio, roadmap, milestone, and weekly review pages should read from query-optimized projections rather than reconstructing aggregates on every request. Core write models remain normalized, while background jobs refresh summary tables or materialized views for:
- project health distribution
- delayed and blocked milestones
- BAU versus milestone work ratio
- owner workload summaries
- GitLab-linked execution summaries

Alternatives considered:
- Compute every dashboard metric from transactional tables on demand.
  Rejected because portfolio queries will become slow and hard to evolve as more metrics are added.

### Decision: Start with workspace-scoped RBAC and event-driven notifications

MVP permissions use workspace roles such as `admin`, `portfolio_manager`, `project_owner`, `contributor`, and `viewer`. Reminders and alerts are triggered by milestone dates, missing weekly updates, blocked status, or stale linked GitLab work. Delivery channels can start with email and a Feishu adapter behind a shared notification event contract.

Alternatives considered:
- Per-field ACLs from day one.
  Rejected because they add complexity without clear MVP value.
- Hardcode channel logic into milestone workflows.
  Rejected because a notification event layer keeps future channel additions simpler.

## Risks / Trade-offs

- [Milestones become renamed task buckets] -> Require completion criteria before activation and keep milestone review centered on outcome language rather than issue counts.
- [GitLab auto-link rules create noisy associations] -> Start with explicit preview/confirmation and let owners override with manual link/unlink controls.
- [Manual health updates are neglected] -> Send reminders for stale weekly updates and surface health fields prominently in review workflows.
- [Dashboard summaries drift from source records] -> Rebuild projections after write events and provide last-updated timestamps plus sync status visibility.
- [Notification overload reduces trust] -> Start with a narrow set of milestone-based alerts and allow role-aware channel preferences later.

## Migration Plan

1. Implement base domain schema and CRUD flows for roadmap, project, milestone, workstream, internal work items, and weekly updates.
2. Add read models and dashboard/review queries once transactional writes are stable.
3. Introduce GitLab connection settings, manual issue linking, and scheduled sync.
4. Add webhook-triggered incremental sync, stale-work detection, and operational visibility for sync failures.
5. Enable notifications for overdue milestones, blocked milestones, and missing weekly updates.
6. Roll out by seeding one workspace with a limited set of roadmap/project owners, validate review workflows, then expand usage.

Rollback strategy: disable sync jobs and notification workers first, preserve PM-owned records, and hide GitLab-linked summaries if integration behavior proves unstable.

## Open Questions

- Should internal non-engineering tasks remain lightweight linked work items, or do they need a richer task workflow in MVP?
- Should one GitLab issue be allowed to move between milestones with a preserved audit trail of prior linkage?
- How much of `progress_percent` should be auto-assisted in MVP versus fully manual?
- Is single-workspace operation sufficient for the initial release, or should workspace partitioning be prepared in the schema now?

## Resolved Questions

_2026-05-20: 以下问题在实现中被隐式回答，正式记录如下。_

### RQ-1: 内部任务的工作流丰富度

**决策：** 保持轻量 linked work item，不需要独立的任务工作流。
**理由：** MVP 阶段内部任务主要用于记录非工程工作（运营、产品、商务），不需要指派、评论、子任务等项目管理功能。GitLab 已经承担了工程任务的详细管理。
**何时重新评估：** 当非工程用户反馈需要协作功能时。

### RQ-2: progress_percent 自动化程度

**决策：** MVP 阶段完全手动，不自动计算。
**理由：** 自动计算需要定义权重规则（不同工作项对里程碑的贡献度不同），这在没有用户反馈的情况下容易猜错。先让项目 owner 手动填写，观察使用模式后再引入辅助。
**何时重新评估：** 收集 2-3 个迭代周期的手动 progress 数据后。

### RQ-3: 单 workspace 操作

**决策：** 单 workspace，schema 不预留多租户字段。
**理由：** MVP 目标是验证 milestone-driven 工作流本身，不是验证多租户架构。提前引入 workspace partitioning 增加复杂度但没有验证价值。
**何时重新评估：** 当第二个团队想独立使用系统时。

### RQ-4: GitLab issue 跨里程碑移动

**决策：** 暂不支持。一个 GitLab issue 只能关联一个里程碑。需要更换时先 unlink 再 link 新里程碑。
**理由：** 跨里程碑移动需要审计追踪（哪个里程碑何时关联/取消），增加了状态机复杂度。unlink + re-link 已经满足需求，只是操作步骤多一步。
**何时重新评估：** 当频繁出现 issue 需要在里程碑间迁移的场景时。

## Known Gaps

_2026-05-20: PRD/Design 与实现的差距分析，详见 `docs/PRD_GAPS.md`。_

| # | 问题 | 严重程度 | 状态 |
|---|------|---------|------|
| GAP-1 | 前端 api.ts 硬编码 X-Role: admin，RBAC 在 UI 层失效 | CRITICAL | 待决策 |
| GAP-2 | MySQL schema 存在但未接入，数据重启即丢失 | HIGH | 待决策 |
| GAP-3 | 里程碑无法在 UI 上流转状态（not_started→active→completed） | HIGH | 待实现 |
| GAP-4 | 无周报提交 UI，Review 页面只有展示 | HIGH | 待实现 |
| GAP-5 | 搜索/筛选 UI 缺失 | MEDIUM | 待实现 |
| GAP-6 | GitLab issue 状态未在前端展示 | MEDIUM | 待实现 |
