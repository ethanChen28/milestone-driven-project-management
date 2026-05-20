## Why

The PRD gap analysis shows that the current implementation passes backend API scenarios but leaves several product-critical workflows unusable or unsafe in the browser: every frontend user is effectively an admin, data is lost on restart, milestone-driven execution cannot be advanced from the UI, weekly updates cannot be submitted, and GitLab/search capabilities are only partially exposed. This change closes those MVP gaps so the implemented system matches the PRD's milestone-driven project management intent.

## What Changes

- Add a frontend role context that stops hard-coding admin access and sends the selected workspace role consistently with API requests.
- Replace the in-memory-only backend store with MySQL-backed persistence for core planning, reporting, access, notification, and GitLab sync data while preserving Redis for cache/ephemeral coordination if needed.
- Add UI workflows to update milestone status, health, completion criteria, progress, and completion date from project and milestone detail views.
- Add a weekly update submission workflow and ensure submitted updates appear in review views and missing-update alert logic.
- Add search and filter controls for planning/reporting lists, including risk filters and available GitLab repository/context filters.
- Expose GitLab-linked work in project and milestone views with origin, sync status, issue state, and links to the original issue.
- Document the four previously open design questions as resolved MVP decisions so future work can distinguish intentional simplifications from omissions.
- No breaking API changes are intended; existing endpoints should remain compatible unless a compatibility-preserving persistence implementation reveals invalid data assumptions.

## Capabilities

### New Capabilities
- `persistent-data-storage`: Durable MySQL-backed storage for domain data that must survive backend restarts and redeployments.

### Modified Capabilities
- `workspace-access-and-notifications`: Browser-facing role selection and permission-aware UI behavior must align with workspace RBAC and notification triggers must consume persisted update/milestone state.
- `project-delivery-management`: Milestone lifecycle management must be usable from the UI, including valid status transitions and completion metadata.
- `review-and-portfolio-reporting`: Weekly update submission and search/filter behavior must be available from reporting and planning views.
- `work-item-and-gitlab-sync`: GitLab-linked work status and source links must be visible in project and milestone workflows.

## Impact

- Backend: persistence layer, repository interfaces, MySQL migrations/schema alignment, service wiring, filters, and integration tests for restart-safe behavior.
- Frontend: role context, API headers, milestone edit/status controls, weekly update form, list filters, GitLab-linked work display, and E2E coverage.
- Infrastructure: docker-compose MySQL readiness assumptions, schema initialization, and any seed/test data needed for local development.
- Tests: unit, integration, and E2E coverage for RBAC-visible UI behavior, durable storage, milestone transitions, weekly updates, filters, and GitLab display.
