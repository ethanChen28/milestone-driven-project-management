## 1. Domain Model And Persistence

- [x] 1.1 Define persistence schema and core models for roadmap, roadmap item, project, milestone, workstream, linked work item, weekly update, and sync metadata
- [x] 1.2 Implement lifecycle/status enums, ownership fields, timeline fields, and audit fields required by the new planning objects
- [x] 1.3 Add validation rules for milestone activation, non-BAU work item project linkage, and role-scoped write permissions

## 2. Planning And Delivery Management APIs

- [x] 2.1 Implement CRUD APIs/services for roadmap periods and roadmap items, including archive behavior and project linkage
- [x] 2.2 Implement CRUD APIs/services for projects, milestones, and workstreams with health, progress, and completion-criteria rules
- [x] 2.3 Implement CRUD APIs/services for internal linked work items and weekly updates

## 3. GitLab Linking And Sync

- [x] 3.1 Implement GitLab connection settings and sync rule configuration for group, repository, label, assignee, milestone, and query filters
- [x] 3.2 Implement manual GitLab issue linking and unlinking flows that create or update linked work item records
- [x] 3.3 Implement scheduled and webhook-triggered sync jobs that update GitLab-owned fields only and record retryable sync failures

## 4. Portfolio Views And Reporting

- [x] 4.1 Build read models or summary queries for roadmap progress, project health distribution, delayed milestones, dependency hotspots, and GitLab-linked execution summaries
- [x] 4.2 Implement roadmap overview, project detail, milestone detail, portfolio dashboard, and weekly review APIs/views
- [x] 4.3 Implement filtering and search by roadmap period, project, milestone, owner, team, status, health, risk, source type, and GitLab context

## 5. Permissions And Notifications

- [x] 5.1 Implement workspace RBAC for admin, portfolio_manager, project_owner, contributor, and viewer actions
- [x] 5.2 Implement reminder and alert triggers for upcoming milestone dates, overdue milestones, blocked milestones, missing weekly updates, and stale linked GitLab work
- [x] 5.3 Add notification delivery adapters and configuration for email and Feishu-compatible event handling

## 6. Validation And Rollout

- [x] 6.1 Add automated tests covering core validations, source-of-truth boundaries, sync behavior, and reporting filters
- [x] 6.2 Add operational visibility for sync status, projection freshness, and notification failures
- [x] 6.3 Seed an MVP workspace and verify the milestone-driven review workflow end to end with representative roadmap, project, and GitLab-linked data
