## 1. User Service Foundation

- [x] 1.1 Create `user-service` backend module with Dockerfile, health endpoint, config loading, and MySQL connection setup.
- [x] 1.2 Add user service database migrations for users, workspaces, memberships, role assignments, external identities, sessions, signing keys, and audit events.
- [x] 1.3 Seed deterministic development users matching existing demo identities (`tester`, `alice`, `bob`, `carol`, `frontend-user`) and default workspace roles.
- [x] 1.4 Add unit tests for user model validation, account status transitions, membership lookup, and role assignment rules.

## 2. Authentication And Token Contract

- [x] 2.1 Implement built-in identity provider login flow for development and initial production use.
- [x] 2.2 Implement access token issuance with required claims: `sub`, `workspace_id`, `roles`, `display_name`, `email`, `provider`, and `version`.
- [x] 2.3 Expose signing metadata through JWKS or implement token introspection, including key rotation support.
- [x] 2.4 Implement token revocation or token-version invalidation for disabled accounts, removed memberships, and role changes.
- [x] 2.5 Add integration tests for successful login, failed login, disabled account, expired token, and signing metadata refresh.

## 3. External Identity Adapter Boundary

- [x] 3.1 Define the identity provider adapter interface for authenticate, profile sync, external identity resolution, and session refresh.
- [x] 3.2 Implement built-in provider through the adapter interface instead of hardcoding login behavior in handlers.
- [x] 3.3 Add configuration placeholders and contract tests for future OIDC/OAuth2 and LDAP/AD providers.
- [x] 3.4 Add audit events for external identity mapping creation, update, and conflict handling.

## 4. Project Service Authentication Integration

- [x] 4.1 Add project service auth middleware with `AUTH_MODE=dev-header` and `AUTH_MODE=token` modes.
- [x] 4.2 In token mode, validate access tokens and build the existing `AuthContext` from trusted token subject and role claims.
- [x] 4.3 Reject production mutation requests that rely only on browser-controlled `X-Role` or `X-User` headers.
- [x] 4.4 Add a user service client for member directory lookup and user id display resolution.
- [x] 4.5 Update project, milestone, task, and weekly update authorization checks to compare stable user ids while retaining temporary compatibility for legacy string owners.
- [x] 4.6 Add backend integration tests for project owner, contributor, viewer, invalid token, and removed member authorization paths.

## 5. Frontend Login And User Directory

- [x] 5.1 Add frontend auth session state, token storage strategy, login page or login panel, logout action, and current user menu.
- [x] 5.2 Hide manual role/user selectors in production authentication mode while keeping development identity tooling available in development mode.
- [x] 5.3 Replace hardcoded `workspaceUsers` member lists with user directory data for project participants, task owner, and weekly update author selection.
- [x] 5.4 Update API client to send `Authorization: Bearer <token>` in token mode and keep dev identity headers only in development mode.
- [x] 5.5 Add frontend unit tests for auth state, API headers, production selector hiding, and directory-backed member options.

## 6. Data Migration And Compatibility

- [x] 6.1 Add migration script to map existing owner, participant, task owner, and author strings to seeded or imported user ids.
- [x] 6.2 Add unresolved identity report output for records that cannot be automatically mapped.
- [x] 6.3 Keep read compatibility for legacy string identity fields during migration and mark unresolved identities in API responses.
- [x] 6.4 Document rollback path from token mode to existing development header mode for non-production only.

## 7. End-To-End Verification

- [x] 7.1 Add E2E setup that obtains development tokens from user service rather than directly spoofing production identity headers.
- [x] 7.2 Add E2E tests for login, project creation as project owner, contributor task edit limits, viewer read-only behavior, and disabled user rejection.
- [x] 7.3 Run user service unit tests and project service unit tests.
- [x] 7.4 Run backend integration tests covering service-to-service identity lookup and token verification.
- [x] 7.5 Run frontend unit tests and full E2E test suite.
- [x] 7.6 Run `openspec validate user-service-design` and resolve all validation errors before implementation review.
