## 1. API and Data Preparation

- [x] 1.1 Audit existing project dashboard, milestone, work item, and weekly review APIs for fields needed by project-space rollups.
- [x] 1.2 Add or extend a project-space overview response that returns project metadata, milestone rollups, work item counts, recent updates, top risk signals, and dependency signals.
- [x] 1.3 Add project-scoped work item query support for milestone, status, priority, owner, blocked state, source type, and GitLab context filters.
- [x] 1.4 Add derived project risk aggregation from high-risk milestones, blocked milestones, blocked work items, external dependency work items, and weekly update risk/blocker fields.
- [x] 1.5 Add derived project dependency aggregation from milestone dependency summaries, blocked work items, and `external_dependency` linked work items.
- [x] 1.6 Preserve existing project, milestone, work item, and weekly review endpoints as backward-compatible entry points.

## 2. Frontend Project Space Structure

- [x] 2.1 Refactor project detail into project-space tabs for overview, work items, milestones, updates, risks, dependencies, and settings.
- [x] 2.2 Encode selected tab and filters in route state or query parameters so navigation is shareable and reload-safe.
- [x] 2.3 Add a project header with owner, health, status, target dates, breadcrumb, and health actions.
- [x] 2.4 Implement responsive tab behavior for desktop and mobile layouts.
- [x] 2.5 Keep global navigation entries for projects, work items, milestones, roadmap, and weekly review unchanged.

## 3. Project Overview Experience

- [x] 3.1 Build overview metric cards for milestone counts, work item status counts, blocked/overdue counts, health, and target-date signals.
- [x] 3.2 Build current milestone summary cards with status, health, owner, planned date, progress, and linked work count.
- [x] 3.3 Build recent weekly update summaries limited to the latest project updates.
- [x] 3.4 Build top risk and decision summary cards from derived risk signals.
- [x] 3.5 Ensure overview summary cards link to owning tabs or detail pages rather than embedding full management tables.

## 4. Project Work Item View

- [x] 4.1 Build project-scoped work item list using shared work item display helpers.
- [x] 4.2 Add project work item grouping by milestone, status, priority, owner, source type, and blocked state.
- [x] 4.3 Add quick filters for milestone, blocked state, overdue state, and source type.
- [x] 4.4 Show project and milestone breadcrumbs on work item cards and detail routes.
- [x] 4.5 Preserve GitLab synced fields as read-only execution context while allowing PM-owned metadata updates.

## 5. Milestones, Updates, Risks, and Dependencies

- [x] 5.1 Move full project milestone management into the project milestones tab while keeping milestone detail routes available.
- [x] 5.2 Add quick-filter navigation from overview milestone summaries to project work items filtered by milestone.
- [x] 5.3 Add a project updates tab that shows project weekly update history and links to global weekly review filters.
- [x] 5.4 Add a project risks tab that lists derived risk signals and opens the source milestone, work item, update, or dependency context.
- [x] 5.5 Add a project dependencies tab that lists external dependencies, blocked work, and milestone dependency summaries.

## 6. Tests and Validation

- [x] 6.1 Add backend unit or integration tests for project-space overview rollups, project work item filters, risk aggregation, and dependency aggregation.
- [x] 6.2 Add frontend unit tests for tab state, grouping helpers, quick-filter URL/query behavior, and breadcrumb helpers.
- [x] 6.3 Add e2e tests for opening project space, switching tabs, using milestone quick filters, and verifying overview summary boundaries.
- [x] 6.4 Add e2e tests for project risks and dependencies views using blocked work items and external dependency work items.
- [x] 6.5 Run `openspec validate add-project-space-object-views`.
- [x] 6.6 Run frontend build, frontend unit tests, backend tests, and relevant e2e tests.

## 7. Documentation and Cleanup

- [x] 7.1 Update product documentation with project-space tab responsibility boundaries.
- [x] 7.2 Link or update the HTML prototypes to reflect the selected hybrid project-space approach.
- [x] 7.3 Document any additive API query parameters or project-space response fields.
- [x] 7.4 Remove or deprecate duplicated project detail sections only after replacement views and tests are passing.
