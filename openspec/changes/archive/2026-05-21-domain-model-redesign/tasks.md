## 1. Backend Identity and Authorization

- [x] 1.1 Add request actor extraction for `X-User` alongside existing `X-Role`.
- [x] 1.2 Introduce an authorization context carrying role and user through store write methods.
- [x] 1.3 Add backend helpers for project ownership, project participation, task ownership, and administrator bypass.
- [x] 1.4 Enforce project owner scope on project update, milestone create/update, and task create/update/delete paths.
- [x] 1.5 Enforce contributor scope so contributors can write only their own tasks inside participating projects.
- [x] 1.6 Enforce contributor scope on weekly update create/update paths.
- [x] 1.7 Add backend tests for admin, project owner, contributor, outsider, and viewer authorization cases.

## 2. Domain Model Adjustments

- [x] ~~2.1 Add `agent_generated` as a supported work item source type.~~ (Deferred)
- [x] 2.2 Keep Workstream persistence and read behavior intact while removing Workstream as a required task workflow field.
- [x] 2.3 Update validation so new milestone-linked tasks do not require `workstreamId`.
- [x] 2.4 Prevent contributor requests from transitioning milestones to `completed`.
- [x] 2.5 Preserve automatic `completedDate` behavior for authorized milestone completion.

## 3. Frontend Workflow

- [x] 3.1 Add a development current-user selector or equivalent state next to the existing role selector.
- [x] 3.2 Send `X-User` from `apiFetch` on every API request.
- [x] 3.3 Remove the Workstream input from task create and edit forms.
- [x] 3.4 Change task owner entry to a project participant selector when participant data is available.
- [x] ~~3.5 Add `agent_generated` to task source type choices and labels.~~ (Deferred)
- [x] 3.6 Display milestone completion criteria as checklist items in milestone detail views.
- [x] 3.7 Gate milestone completion controls so contributors cannot complete milestones from the UI.

## 4. Verification

- [x] 4.1 Add or update unit tests for task workspace source type and owner selection behavior.
- [x] 4.2 Add backend tests for forbidden contributor edit/delete of another user's task.
- [x] 4.3 Add backend tests for project owner forbidden writes outside owned projects.
- [x] 4.4 Add e2e coverage for `X-User` propagation, contributor-owned task CRUD, and Workstream removal from task forms.
- [x] 4.5 Run `go test ./...`.
- [x] 4.6 Run `npm run build`.
- [x] 4.7 Run `npm run test:unit`.
- [x] 4.8 Run `npm run test:e2e`.
- [x] 4.9 Run `openspec validate "domain-model-redesign"`.
