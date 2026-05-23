## ADDED Requirements

### Requirement: User service owns identity and account records
The system SHALL provide a standalone user service that owns user account records, profile fields, account status, workspace membership, and role assignment data independently from project, milestone, task, and review data.

#### Scenario: Query active workspace members
- **WHEN** the project management service requests active members for a workspace
- **THEN** the user service SHALL return stable user ids, display names, emails, account status, and assigned workspace roles

#### Scenario: Disable user account
- **WHEN** an administrator disables a user account in the user service
- **THEN** the user service SHALL prevent new login sessions and expose the disabled status to dependent services

### Requirement: Token-based authentication contract
The system SHALL authenticate production users through the user service and SHALL issue verifiable access tokens that dependent services can validate without trusting browser-controlled identity headers.

#### Scenario: Successful login issues token
- **WHEN** a valid user authenticates through a configured identity provider
- **THEN** the user service SHALL issue an access token containing at least `sub`, `workspace_id`, `roles`, `display_name`, `email`, `provider`, and `version` claims

#### Scenario: Invalid token is rejected
- **WHEN** a dependent service receives a request with an expired, malformed, revoked, or unverifiable token
- **THEN** the dependent service SHALL reject the request as unauthenticated before executing business authorization logic

#### Scenario: Service validates signing metadata
- **WHEN** the user service rotates token signing keys
- **THEN** dependent services SHALL be able to refresh signing metadata or use introspection without requiring a redeploy

### Requirement: External identity provider adapter
The user service SHALL expose an identity provider adapter boundary so built-in accounts and external user systems can be connected without changing project management business code.

#### Scenario: Resolve external identity
- **WHEN** a user authenticates through an external provider such as OIDC, OAuth2, LDAP, AD, or an enterprise user center
- **THEN** the user service SHALL map the external subject to an internal stable user id before issuing application tokens

#### Scenario: Preserve internal user id across provider changes
- **WHEN** a workspace switches from the built-in provider to an external identity provider for an existing user
- **THEN** the user service SHALL preserve or explicitly migrate the internal user id used by project domain records

### Requirement: Workspace membership and role assignment
The user service SHALL manage workspace memberships and role assignments for supported roles `admin`, `portfolio_manager`, `project_owner`, `contributor`, and `viewer`.

#### Scenario: Assign workspace role
- **WHEN** an administrator assigns a role to a workspace member
- **THEN** the user service SHALL persist the role assignment and expose it in issued tokens or membership lookup responses

#### Scenario: Remove workspace member
- **WHEN** an administrator removes a user from a workspace
- **THEN** new tokens for that workspace SHALL NOT include workspace access, and dependent services SHALL reject write attempts for that workspace

### Requirement: User directory lookup for product workflows
The user service SHALL provide user directory lookup APIs for frontend member selection and backend reference resolution without requiring hardcoded user lists in the project management frontend.

#### Scenario: Frontend loads assignable users
- **WHEN** a user creates or edits a project, milestone-linked task, or weekly update
- **THEN** the frontend SHALL load assignable users from the user service or a project service endpoint backed by the user service

#### Scenario: Render user display fields
- **WHEN** a project domain response contains owner, participant, task owner, or author user ids
- **THEN** the system SHALL provide enough user display data for the frontend to render names without exposing authentication secrets

### Requirement: Identity audit events
The user service SHALL record audit events for authentication, account status changes, role assignment changes, external identity mapping changes, and token revocation decisions.

#### Scenario: Audit role change
- **WHEN** an administrator changes a user's workspace role
- **THEN** the user service SHALL record who changed the role, the previous role, the new role, the target user, and the timestamp

#### Scenario: Audit failed login
- **WHEN** a login attempt fails because of invalid credentials, disabled account, or provider rejection
- **THEN** the user service SHALL record a security audit event without storing raw secrets

### Requirement: Development identity mode is isolated
The user service SHALL support development identity bootstrap or dev-token issuance only when explicitly configured for non-production environments.

#### Scenario: Production rejects development identity mode
- **WHEN** the application starts with a production environment and development identity mode enabled
- **THEN** startup SHALL fail with a clear configuration error

#### Scenario: Development token supports E2E testing
- **WHEN** E2E tests run in development mode
- **THEN** they SHALL be able to obtain deterministic test user tokens without using browser-controlled production identity headers
