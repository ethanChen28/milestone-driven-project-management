## ADDED Requirements

### Requirement: Domain data persists across backend restarts
The system SHALL store durable roadmap, project, milestone, workstream, linked work item, weekly update, GitLab sync, alert, and notification data in MySQL so successful API writes remain available after backend restart or redeployment.

#### Scenario: Created project survives restart
- **WHEN** an authorized user creates a project and the backend process restarts while using the same MySQL database
- **THEN** the project SHALL remain available through project list, project detail, dashboard, and roadmap-linked views

#### Scenario: Weekly update survives restart
- **WHEN** an authorized user submits a weekly update and the backend process restarts while using the same MySQL database
- **THEN** the update SHALL remain available in weekly review history and milestone or project detail views

#### Scenario: GitLab sync state survives restart
- **WHEN** GitLab-linked work, sync rules, sync jobs, sync failures, or alerts are created before a backend restart
- **THEN** those records SHALL remain available after restart with their source metadata, retry state, and dismissed/resolved state preserved

### Requirement: Durable writes are committed before success responses
The system SHALL only return success for durable create, update, archive, dismiss, link, unlink, and resolve operations after the corresponding MySQL transaction has been committed.

#### Scenario: Failed persistence returns an error
- **WHEN** a durable write cannot be committed to MySQL
- **THEN** the API SHALL return an error response and SHALL NOT report the operation as successful

#### Scenario: Related records remain consistent
- **WHEN** an operation changes multiple related durable records as one logical action
- **THEN** the system SHALL commit all related changes together or roll them back together

### Requirement: Runtime storage mode is explicit
The system SHALL make the active storage backend visible in operational health or status output so operators can distinguish MySQL-backed runtime from test-only in-memory storage.

#### Scenario: Check storage backend
- **WHEN** an operator opens the health or operational status endpoint
- **THEN** the response SHALL identify the active storage backend and whether durable persistence is available
