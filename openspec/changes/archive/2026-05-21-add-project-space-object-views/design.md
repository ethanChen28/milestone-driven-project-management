## Context

The product already has baseline capabilities for project delivery management, work item and GitLab sync, review/portfolio reporting, roadmap planning, and workspace access. Current UI screens expose these objects separately, but project detail currently mixes milestone management, work item cards, and summary text in one page without a clear project-space model.

Platform research in `docs/prototypes/project-space-platform-research.md` shows a consistent pattern across Plane, Linear, Jira, Asana, Notion, and ClickUp: mature tools separate object model, entry points, and views. They allow summaries to appear in multiple contexts, but avoid putting full CRUD and full filtering behavior in every context.

This design introduces a hybrid approach: keep the low-learning-cost project-space tabs from A1, but borrow A2's object context through breadcrumbs, quick filters, and a contextual detail panel instead of making an object tree the primary navigation.

## Goals / Non-Goals

**Goals:**

- Define a project-space navigation pattern for project detail pages.
- Make project overview a rollup and triage surface, not a duplicate management table.
- Provide dedicated project-scoped views for work items, milestones, weekly updates, risks, and dependencies.
- Add quick-filter transitions from summary cards to full object views.
- Preserve global entry points for projects, work items, milestones, roadmap, and weekly review.
- Keep APIs additive so existing screens and e2e flows continue to work.

**Non-Goals:**

- Do not implement arbitrary Jira-style custom hierarchy levels.
- Do not replace the global task workspace with project-only task pages.
- Do not turn weekly updates into work item children.
- Do not require a graph canvas for the first implementation.
- Do not introduce a new frontend framework or backend architecture.

## Decisions

### Decision 1: Use project-space tabs as the primary project detail navigation

Project detail SHALL use tabs for `overview`, `work-items`, `milestones`, `updates`, `risks`, `dependencies`, and `settings`.

Rationale: A tabbed project space matches Plane/Linear-style mental models and is easier for current users than an object tree. It also gives each object type a clear ownership surface.

Alternative considered: object tree as primary navigation. Rejected for the first implementation because it increases learning cost and can obscure global work item workflows.

### Decision 2: Treat overview as rollup and triage only

Overview SHALL show aggregate cards, current milestone summaries, top risks, recent updates, and blocked/overdue signals. It SHALL NOT duplicate the full milestone table, full work item table, or full review feed.

Rationale: This addresses the user's concern that tabs and content can overlap. Summary duplication is acceptable only when the interaction density is different.

Alternative considered: overview contains complete embedded lists. Rejected because it creates duplicate CRUD surfaces and inconsistent filtering behavior.

### Decision 3: Use quick filters instead of duplicated nested tables

Clicking a milestone summary in overview SHALL navigate to the project work-items tab filtered by that milestone, or to the milestone detail page when the user chooses the milestone title/action.

Rationale: This follows Linear's pattern where milestones can appear in overview, timeline, or details, but full work item management remains in the issue/work item view.

Alternative considered: render every milestone's full task table inline in overview. Rejected because it does not scale and duplicates the work item workspace.

### Decision 4: Keep work item source-of-truth boundaries intact

Project-scoped work item views SHALL reuse existing linked work item data and GitLab sync fields. GitLab-owned fields remain read-only in PM views unless updated by sync.

Rationale: The new project-space UI is a view and workflow change, not a source-of-truth rewrite.

Alternative considered: create separate project-space task records. Rejected because it would split execution state from GitLab/internal work items.

### Decision 5: Add risk and dependency as project-space first-class views, but implement incrementally

The UI SHALL expose risks and dependencies as dedicated project tabs. The first implementation MAY derive them from milestone risk fields, blocked work items, external dependency source types, and weekly update risk/blocker text before adding dedicated tables.

Rationale: The product needs risk/dependency workflows, but current data already has partial signals. A derived first version avoids schema churn while still creating the right information architecture.

Alternative considered: immediately add full `Risk` and `Dependency` tables. Deferred until the derived view proves insufficient.

## Risks / Trade-offs

- [Risk] Tabs can still become too many for small screens. → Mitigation: collapse project tabs into a segmented control or overflow menu on mobile.
- [Risk] Overview summary cards may drift from full views. → Mitigation: derive overview from the same API rollup/query helpers used by full views.
- [Risk] Derived risk/dependency views may be less precise than first-class records. → Mitigation: label derived signals clearly and keep APIs additive so dedicated records can be introduced later.
- [Risk] Global task workspace and project-scoped work item tab may diverge. → Mitigation: share grouping/filter helper logic and route query semantics.
- [Risk] e2e tests may become brittle if tabs hide content. → Mitigation: add explicit route/query-based state tests for each tab and quick-filter transition.

## Migration Plan

1. Add project-space tabs and keep existing project detail content visible under `overview` until the new tab views are available.
2. Add project overview rollup helpers using existing project dashboard data where possible.
3. Move full milestone management into the `milestones` tab and full project work item management into the `work-items` tab.
4. Add quick-filter navigation from overview summary cards to tab-specific filtered views.
5. Add derived `risks` and `dependencies` tabs from existing milestone/work item/update fields.
6. Add e2e coverage before removing any legacy inline sections from overview.

Rollback strategy: because APIs are additive and global routes remain unchanged, rollback can restore the previous `ProjectDetailView` rendering while leaving new helper functions unused.

## Open Questions

- Should project-space tabs be encoded as `/projects/:id?tab=work-items` or nested routes such as `/projects/:id/work-items`?
- Should risk/dependency become first-class backend records in this change or remain derived for the first implementation?
- Should overview support inline creation for milestones/work items, or only link to the owning tab's create action?
