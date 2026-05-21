## Why

The current project UI exposes project, milestone, work item, weekly update, and risk-related information as separate screens, but it does not define clear responsibility boundaries between overview pages, object lists, detail panels, and cross-object views. As the product adds progressive objects such as Project -> Milestone -> WorkItem -> WeeklyUpdate/Risk/Dependency, the UI needs a deliberate project-space model to avoid duplicated navigation and duplicated CRUD surfaces.

## What Changes

- Introduce a project-space information architecture modeled after mature project tools: project overview for rollups, dedicated tabs/views for full object management, and contextual detail panels for selected objects.
- Add project detail navigation for `overview`, `work-items`, `milestones`, `updates`, `risks`, `dependencies`, and `settings` without turning every summary card into a full duplicate list.
- Add overview rollups for current milestones, key risks, recent weekly updates, work item counts, health, timeline, and blocked/overdue state.
- Add quick-filter behavior so milestone summaries can open the work item view filtered to the selected milestone instead of duplicating the complete work item table inside overview.
- Add project-scoped work item grouping modes including by milestone, priority, status, and owner, with support for nested and flattened display where applicable.
- Add first-class project-space risk and dependency views focused on blockers, decision needs, and cross-object relationships.
- Preserve existing global entry points for projects, work items, milestones, roadmap, and weekly review.
- No breaking API change is intended; new APIs or query parameters should be additive.

## Capabilities

### New Capabilities
- `project-space-object-views`: Defines project-space navigation, overview rollups, tab responsibility boundaries, quick filters, detail-panel behavior, and relationship-focused risk/dependency views.

### Modified Capabilities
- `project-delivery-management`: Project and milestone views gain project-space navigation, overview rollups, and explicit rules preventing overview cards from replacing full milestone/work-item management views.
- `work-item-and-gitlab-sync`: Work item views gain project-scoped grouping/filtering by milestone and contextual breadcrumbs while preserving GitLab source-of-truth boundaries.
- `review-and-portfolio-reporting`: Weekly updates, risks, blockers, and decisions become project-space summary inputs with clear ownership between overview summaries and full review/reporting views.

## Impact

- Frontend: project detail route, project overview layout, project-scoped tabs, work item filters/grouping, milestone quick filters, risk/dependency views, responsive navigation behavior.
- Backend/API: additive dashboard or query endpoints may be needed for project-space rollups, filtered work item lists, risks, and dependencies.
- Data model: may add first-class `Risk` and `Dependency` records or additive fields/relations if existing work item data is insufficient.
- Tests: unit tests for view state and grouping helpers; integration tests for project-space rollup APIs; e2e tests for tab navigation, quick filters, overview summaries, and risk/dependency workflows.
- Documentation: update PRD/spec docs to define responsibility boundaries between overview, full object lists, and detail panels.
