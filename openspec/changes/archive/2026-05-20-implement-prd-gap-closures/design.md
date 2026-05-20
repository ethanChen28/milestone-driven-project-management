## Context

`docs/PRD_GAPS.md` identifies that the backend API scenarios are mostly implemented, but several PRD workflows remain incomplete at the product level. The most important gaps are frontend RBAC bypass through a hard-coded admin role, in-memory-only data storage despite MySQL/Redis infrastructure, incomplete milestone lifecycle UI, incomplete weekly review entry/selection experience, missing filter controls, limited GitLab visibility in planning views, and unresolved design questions.

Current constraints remain unchanged: Vue 3 + Vite + TypeScript frontend, Go backend, MySQL as durable storage, Redis available for cache/ephemeral coordination, Dockerfile-based deployment, and Simplified Chinese as the default locale with English support.

## Goals / Non-Goals

**Goals:**

- Make browser requests permission-aware by replacing the hard-coded admin role with a user-selectable workspace role for MVP.
- Persist domain data in MySQL so roadmap, project, milestone, work item, weekly update, GitLab sync, alert, and notification records survive backend restarts.
- Allow users to advance milestone status and health from the UI while preserving backend validation as the source of truth.
- Complete weekly update submission so users can choose project/milestone context instead of typing opaque IDs.
- Add filters to core planning and reporting views for the dimensions already required by the specs, including risk and GitLab context.
- Display GitLab-linked work details and original issue links in project and milestone workflows.
- Record the PRD/design open questions as resolved MVP decisions.

**Non-Goals:**

- Introduce JWT, OAuth2, SSO, password login, or multi-tenant identity management in this change.
- Build full GitLab scheduled synchronization beyond the existing webhook/rule-based skeleton unless required to display persisted sync status.
- Replace all dashboard aggregation with optimized materialized projections in this change.
- Add a full organization/team management module; team filtering will use a minimal persisted team field or existing project metadata until a dedicated team capability is proposed.

## Decisions

### 1. Keep header-based MVP role context and make it explicit in the UI

Use the existing `X-Role` contract for MVP. Add a frontend role context with role selector, local persistence, and helper permission predicates. `apiFetch` reads the current role instead of hard-coding `admin`. Mutating controls are shown or disabled according to the selected role, but the backend remains authoritative and still rejects unauthorized requests.

The role selector is labeled as an MVP debug/development tool (not a production login mechanism). Default role is `contributor` (not admin) so permission boundaries are exercised by default. The UI clearly marks the selector as non-production.

Alternatives considered:

- JWT/OAuth2 now: stronger production posture, but it adds identity provider, session, token refresh, and user mapping scope that is not required to close the current MVP gap.
- Keep hard-coded admin: simplest, but it invalidates PRD RBAC and hides permission bugs from browser testing.
- Default to admin role: hides RBAC issues during development. Defaulting to contributor forces RBAC testing.

### 2. Introduce a repository boundary before wiring MySQL

Separate storage from domain rules by defining repository operations for the entities already managed by `service.Store`. Keep an in-memory repository for unit tests and implement a MySQL repository for runtime when `MYSQL_DSN` is configured. The service layer should own validation, RBAC, and rollup behavior; repositories should own CRUD, query filters, and transactional persistence.

The MySQL implementation must align `infra/mysql/init/*.sql` with the Go domain model. Known schema mismatches to address include optional source fields for non-GitLab/internal work, JSON/text encoding for string arrays, GitLab fields, alerts, notification events, and durable ID generation after restart.

Repository methods default to self-contained transactions (begin/commit/rollback within each method call). When the service layer needs to coordinate multiple repository writes atomically (e.g., `GenerateAlerts` writes multiple alert records, `RunSyncForRule` updates a sync job plus multiple work items), the service manages a `*sql.Tx` and passes it to repository methods that accept an optional transaction argument. In-memory repository implementations ignore the transaction parameter.

A database sequence table `id_sequences(prefix VARCHAR(32) PK, last_val BIGINT)` will be used to generate durable `prefix-NNN` IDs (e.g., `rp-001`, `prj-002`). This preserves the current ID format and avoids changes to API responses or E2E tests. On MySQL startup the table is seeded from `max(id)` of each entity table if rows exist.

MVP data migration is out of scope: the system starts from an empty MySQL database. No production data exists to migrate.

Alternatives considered:

- Rewrite `Store` directly around SQL calls: faster initially, but it mixes business rules and persistence even more tightly.
- Add an ORM: not necessary for current table count and would increase dependency and migration complexity.
- UUID-based IDs: would change all API response formats and break existing E2E test assertions. Sequence table preserves format.
- AUTO_INCREMENT: does not produce the `prefix-NNN` format used throughout the codebase.

### 3. Use MySQL as source of truth; Redis remains optional

All domain data listed in the specs must be committed to MySQL before an API write returns success. Redis may be used later for cache or background coordination, but it must not be the only copy of any roadmap/project/milestone/update/GitLab/alert/notification record.

Alternatives considered:

- Continue in-memory storage for MVP: unacceptable for restart/redeploy reliability.
- Cache-first writes: higher complexity and no current requirement for it.

### 4. Backend owns milestone transition invariants

The UI will expose milestone edit/status controls, but backend validation remains authoritative. The valid milestone state machine is:

```
not_started ──(requires completion_criteria)──► active
active ──► blocked
active ──► completed (auto-sets completedDate)
active ──► cancelled
blocked ──► active
completed ──► ✗ rejected (terminal state)
cancelled ──► ✗ rejected (terminal state)
```

Skipping `active` is not allowed: `not_started → completed` is rejected. A milestone must pass through `active` (which requires completion criteria) before it can be completed. Reactivating a `completed` or `cancelled` milestone is rejected unless a future spec defines reopening.

Alternatives considered:

- Allow `not_started → completed` directly: contradicts PRD principle "milestone first" and bypasses completion criteria validation.
- Let frontend enforce all transition rules: easier UI work, but API clients and tests could bypass rules.
- Add a separate transition endpoint now: cleaner command semantics, but existing `PUT /api/v1/milestones?id=...` can remain compatible for MVP.

### 5. Complete weekly updates with contextual selectors

Keep the existing weekly update endpoint and review view, but make the UI use loaded project/milestone options, default week metadata, validation, and refresh behavior. Missing weekly update alerts should be based on persisted update history after MySQL is introduced.

Alternatives considered:

- Keep free-form ID fields: technically functional but poor fit for PRD review workflow and error-prone for non-technical users.

### 6. Implement filters at API and UI layers

Backend list endpoints and review/dashboard endpoints should accept the existing query dimensions where applicable. Add missing risk and GitLab context filtering for milestones/work items, and expose filter controls in frontend list/reporting pages.

Team filtering is out of scope for this change: the system serves a single team, so filtering by team adds no value. PRD §8.8 team filtering will be revisited when multi-team usage is needed.

Alternatives considered:

- Frontend-only filtering: insufficient once datasets grow and inconsistent with API testability.
- Add a minimal team field now: no current use case since the system serves one team.

### 7. Show GitLab-linked work where delivery decisions are made

Project and milestone detail views should show linked work items with source type, GitLab state, assignee, labels, sync freshness, and source URL. The original GitLab issue link opens in a new tab. Non-GitLab work items remain visible without GitLab-specific fields.

Alternatives considered:

- Add a standalone GitLab page only: useful for administrators but does not satisfy the PRD need to see issue state in project/milestone context.

### 8. Resolve prior open questions for MVP

- Internal tasks remain lightweight linked work items; no richer workflow is added in this change.
- GitLab issue cross-milestone movement is not implemented; PM-owned project/milestone metadata remains controlled by this system.
- `progressPercent` remains manually controlled; linked work can inform display but does not override manual health/status.
- Single workspace remains the MVP scope; multi-workspace support requires a future spec.

## Risks / Trade-offs

- [Schema/domain mismatch] Existing SQL schema does not fully match current Go models and source-type semantics. -> Update migrations and add integration tests that create each supported entity type through the API and reload it from MySQL.
- [Role selector is not authentication] Users can still choose roles in the browser. -> Treat it as MVP RBAC exercise only; document that production authentication is out of scope and backend role checks remain centralized.
- [Persistence refactor regression] Moving from maps to SQL can change ordering, default values, and empty field behavior. -> Add repository contract tests and API integration tests before replacing runtime storage.
- [Filter semantics drift] Different endpoints may interpret the same filter differently. -> Define shared query names and add tests for owner/status/health/risk/sourceType/gitLab context combinations.
- [Dashboard query performance] SQL-backed rollups may be less efficient than future projections. -> Use direct SQL aggregation for MVP and keep projection/materialized read models as future optimization.
