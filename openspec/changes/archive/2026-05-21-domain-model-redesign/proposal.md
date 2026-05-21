## Why

The current domain model still reflects traditional project management: work is nested through Project > Milestone > Workstream > Task, and permissions are driven by a global role header without user or project ownership checks. The team has moved to an Agent-driven delivery model where Team Leaders define milestones and acceptance criteria, while engineers independently create and manage their own tasks with Agent support.

This change realigns the system around milestone acceptance, project-scoped ownership, and engineer task autonomy.

## What Changes

- Freeze Workstream as a legacy database concept and remove it from active task and milestone workflows.
- Treat the active delivery model as Project > Milestone > Task.
- Add user identity through `X-User` alongside the existing `X-Role` development header.
- Enforce project-scoped permissions for project owners and contributors.
- Allow contributors to create, edit, and delete only their own tasks inside projects they participate in.
- Allow project owners to manage milestones and all tasks only within projects they own.
- Keep milestone completion as a Team Leader or administrator action because it represents acceptance, not execution.
- Strengthen milestone completion criteria from a loose text field into a checklist-oriented acceptance contract.
- ~~Add Agent-generated work as an explicit task source type.~~ (Deferred: not core functionality, revisit when Agent usage stabilizes)

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `project-delivery-management`: Freeze Workstream as an inactive delivery lane, reinforce the 3-level Project > Milestone > Task model, and make milestone completion criteria checklist-oriented with owner-only completion.
- `work-item-and-gitlab-sync`: Update task ownership rules and remove active Workstream usage from work item workflows.
- `workspace-access-and-notifications`: Add `X-User` identity, project-scoped authorization, participant checks, and owner-based task CRUD constraints.

## Impact

- Backend API handlers and store methods must read and propagate current user identity in addition to role.
- Project, milestone, and work item authorization must evaluate global role, project ownership or participation, and task ownership.
- Task create, update, and delete APIs must reject unauthorized contributor edits.
- Milestone status transitions must reject contributor completion attempts.
- Frontend task forms must hide Workstream controls and expose owner selection from project participants.
- Milestone UI must surface completion criteria as a checklist and gate completion confirmation to Team Leaders.
- Tests must cover backend authorization, frontend role/user behavior, and end-to-end task and milestone workflows.
