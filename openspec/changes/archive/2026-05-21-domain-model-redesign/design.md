## Context

The current implementation already has the main runtime pieces for a Project > Milestone > Task workflow, but the domain contract still contains a Workstream layer and the authorization model only receives a global role through `X-Role`. That leaves two gaps: Workstream remains visible as a low-value field, and contributors can manage tasks without proving they own the task or participate in the project.

The target operating model is Agent-assisted delivery. Team Leaders define milestones and acceptance criteria. Engineers create and maintain their own execution tasks. The system should govern scope and acceptance, not assign technical lanes.

## Goals / Non-Goals

**Goals:**

- Make Project > Milestone > Task the active delivery model.
- Preserve Workstream data compatibility while removing Workstream from active UI and task workflows.
- Add `X-User` as the MVP user identity input alongside `X-Role`.
- Enforce project ownership, project participation, and task ownership in backend write paths.
- Let contributors freely CRUD their own tasks inside projects where they are participants.
- Keep milestone completion restricted to project owners and administrators.
- Represent milestone completion criteria as checklist-friendly acceptance data.

**Non-Goals:**

- Replace the MVP header-based auth model with production login or SSO.
- Delete Workstream tables, structs, repository state, or historical data.
- Build a full approval workflow for task creation.
- Implement automated Agent quality scoring in this change.
- Add `agent_task` or `agent_generated` source type (deferred; not core to ownership model).

## Decisions

### Decision 1: Freeze Workstream, do not remove it

Workstream remains in the backend model and persistence state for compatibility, but active task creation and editing should no longer require or expose `workstreamId`. Existing data may still be returned where historical records contain it.

Alternative considered: remove Workstream from schema and repositories immediately. That would create unnecessary migration risk for a low-value cleanup and would complicate rollback.

### Decision 2: Add `X-User` before production auth

Backend request context should include both role and actor:

- `X-Role`: existing workspace role.
- `X-User`: current user identifier.

For local development, missing `X-User` may default to a deterministic development user only where existing tests need backward compatibility. New authorization-sensitive tests should set `X-User` explicitly.

Alternative considered: infer current user from request payload fields such as `owner` or `author`. That is not defensible for authorization because clients can change those fields.

### Decision 3: Scope global permissions with project and task ownership

Global permissions still decide operation category, but write paths must also check scope:

- `admin`: all write operations.
- `portfolio_manager`: portfolio/project administration and cross-project visibility, but no automatic milestone acceptance authority unless explicitly allowed by existing management permissions.
- `project_owner`: project, milestone, and task writes only for projects where `Project.owner == X-User`.
- `contributor`: task writes only when the project includes `X-User` in `participants` and the task owner is `X-User`.
- `viewer`: read only.

Alternative considered: create a separate membership table now. The current `Project.participants` field is already present, so using it keeps the change small and testable.

### Decision 4: Completion criteria becomes checklist-oriented while preserving API compatibility

The current `completionCriteria` string should remain accepted, but UI and service helpers should parse it into checklist items using newline-based entries. Future storage can introduce a typed checklist field after usage stabilizes.

Alternative considered: immediately change `completionCriteria` from string to an array. That would be cleaner long term but risks breaking existing payloads, tests, and persisted records.

## Risks / Trade-offs

- [Risk] Header-based `X-User` can be spoofed in production. → Mitigation: keep it explicitly documented as MVP identity plumbing and isolate actor extraction behind a helper that production auth can replace.
- [Risk] Project participant matching depends on exact user identifiers. → Mitigation: normalize empty values and add tests for owner, participant, and outsider cases.
- [Risk] Workstream endpoints may still exist after UI removal. → Mitigation: keep endpoints for compatibility but remove active navigation and task form usage.
- [Risk] String-based checklist parsing may be limited. → Mitigation: support newline items now and keep the data model backward compatible for a future typed checklist migration.

## Migration Plan

1. Add request actor extraction and pass actor through project, milestone, work item, and weekly update write paths.
2. Add authorization helpers that combine role, project ownership, project participation, and task ownership.
3. ~~Extend source type validation to support `agent_generated`.~~ (Deferred)
4. Hide Workstream controls in task UI while preserving returned `workstreamId` for existing records.
5. Change milestone UI to display and edit completion criteria as checklist lines.
6. Add backend unit tests and frontend/e2e tests for the new authorization and UI behavior.
7. Rollback by disabling the new scoped checks and continuing to accept existing `completionCriteria` strings; no destructive data migration is required.
