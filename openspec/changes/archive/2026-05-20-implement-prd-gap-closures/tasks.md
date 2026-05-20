## 1. Storage Foundation

- [x] 1.1 Add repository contract tests covering create/list/get/update flows for roadmap periods, roadmap items, projects, milestones, workstreams, linked work items, weekly updates, GitLab configs, sync rules, sync jobs, sync failures, alerts, and notifications.
- [x] 1.2 Extract a repository boundary from `backend/internal/service.Store` so business rules and RBAC remain in the service layer while storage operations can be backed by memory or MySQL. Repository methods default to self-contained transactions; the service layer manages `*sql.Tx` for multi-write operations (GenerateAlerts, RunSyncForRule, LinkGitLabIssue).
- [x] 1.3 Keep an in-memory repository implementation for fast unit tests and existing API tests.
- [x] 1.4 Add a MySQL repository implementation using `MYSQL_DSN` with explicit startup failure when durable storage is required but unavailable.
- [x] 1.5 Align `infra/mysql/init/*.sql` with domain models, including optional non-GitLab source fields, JSON/text array encoding, GitLab sync fields, alerts, notifications, and a `id_sequences(prefix, last_val)` table for durable `prefix-NNN` ID generation.
- [x] 1.6 Wire server startup to use MySQL storage in Docker/runtime and memory storage only in tests or explicit local override.
- [x] 1.7 Expose active storage backend and durable persistence availability in health or operational status output.
- [x] 1.8 Add integration tests proving projects, weekly updates, GitLab sync state, alerts, and dismissed/resolved states survive backend restart against the same MySQL database.

## 2. Frontend Role Context

- [x] 2.1 Add a typed workspace role model and permission helper in the frontend matching `admin`, `portfolio_manager`, `project_owner`, `contributor`, and `viewer`.
- [x] 2.2 Add a role selector to the application shell labeled as MVP debug tool, with Chinese default copy and English translations. Default role is `contributor`.
- [x] 2.3 Persist the selected role across navigation and reloads using browser-local state.
- [x] 2.4 Update `apiFetch` to send the selected role in `X-Role` and remove the hard-coded admin header.
- [x] 2.5 Hide, disable, or explain unauthorized mutating actions for viewer and non-admin roles while preserving backend enforcement.
- [x] 2.6 Add frontend tests or E2E coverage proving selected roles affect request headers and non-admin users cannot manage GitLab integration settings.

## 3. Milestone Lifecycle UI And Rules

- [x] 3.1 Add backend tests for valid milestone transitions: `not_started→active` (requires criteria), `active→blocked`, `active→completed`, `active→cancelled`, `blocked→active`. Verify rejected transitions: `not_started→completed`, `completed→active`, `cancelled→active`, and any terminal-state reentry.
- [x] 3.2 Enforce backend terminal-state protection for `completed` and `cancelled` milestones unless a future reopening workflow is added.
- [x] 3.3 Ensure completing a milestone records `completedDate` when no completed date is provided.
- [x] 3.4 Add milestone edit/status controls to milestone detail view for status, health, progress percent, forecast date, completion criteria, and dependency/risk fields.
- [x] 3.5 Add milestone status actions or links from project detail so users can advance milestones without copying IDs.
- [x] 3.6 Refresh project, milestone, roadmap, dashboard, and review views after milestone lifecycle updates.
- [x] 3.7 Add E2E coverage for not_started -> active -> completed and blocked milestone review visibility.

## 4. Weekly Update Workflow

- [x] 4.1 Add backend validation tests that weekly updates require project context and preserve optional milestone context.
- [x] 4.2 Load project options in review view and replace raw project ID entry with a project selector.
- [x] 4.3 Load milestone options scoped to the selected project and replace raw milestone ID entry with an optional milestone selector.
- [x] 4.4 Default week metadata for the submission form while allowing the user to override it.
- [x] 4.5 Show submission errors inline and refresh review history after successful weekly update creation.
- [x] 4.6 Verify missing weekly update alerts use persisted MySQL-backed update history after restart.
- [x] 4.7 Add E2E coverage for submitting a weekly update and seeing it in review and milestone/project context.

## 5. Search And Filtering

- [x] 5.1 Add missing backend filters for milestone `riskLevel` and GitLab repository/context where applicable. Team filtering is out of scope (single-team system).
- [x] 5.2 Normalize query parameter names across list, review, dashboard, project, and milestone endpoints for owner, status, health, risk, sourceType, roadmap period, project, milestone, and GitLab context.
- [x] 5.3 Add browser filter controls to projects, milestones, roadmap, review, and relevant detail work-item sections.
- [x] 5.4 Implement clear-filter behavior that reloads the unfiltered view.
- [x] 5.5 Add backend API tests for combined filters and no-match cases.
- [x] 5.6 Add E2E coverage for applying and clearing filters in at least milestone and review views.

## 6. GitLab Delivery Visibility

- [x] 6.1 Extend frontend work-item types to include source type, source URL, GitLab labels, assignee, state, and last sync time.
- [x] 6.2 Show GitLab-linked work metadata and original issue links in milestone detail view.
- [x] 6.3 Show GitLab-linked work metadata and original issue links in project detail view.
- [x] 6.4 Render non-GitLab work items in the same sections without GitLab-only fields.
- [x] 6.5 Add frontend tests or E2E coverage for opening a GitLab issue link and verifying PM-owned milestone/project fields remain unchanged.

## 7. Documentation And Validation

- [x] 7.1 Record the four resolved MVP design questions in the active change design or implementation docs after code changes are complete.
- [x] 7.2 Update local development documentation with MySQL-backed startup, storage mode override, and required environment variables.
- [x] 7.3 Run backend unit tests.
- [x] 7.4 Run backend MySQL integration tests.
- [x] 7.5 Run frontend unit tests and typecheck.
- [x] 7.6 Run frontend end-to-end tests.
- [x] 7.7 Run `openspec validate "implement-prd-gap-closures"`.
