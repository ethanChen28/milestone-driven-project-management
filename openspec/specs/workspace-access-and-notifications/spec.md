## Purpose
Defines workspace access control, alerts, notifications, and identity behavior.
## Requirements
### Requirement: Workspace role-based access control
The system SHALL support workspace roles `admin`, `portfolio_manager`, `project_owner`, `contributor`, and `viewer` with permissions aligned to roadmap, project, milestone, update, and integration actions, and SHALL combine role permissions from trusted identity claims with project owner and participant references for project-scoped and owner-scoped write authorization.

#### Scenario: Restrict integration administration
- **WHEN** a non-admin user attempts to manage GitLab integration settings
- **THEN** the system SHALL deny the action

#### Scenario: Allow project owners to update milestone state
- **WHEN** a project owner updates milestone status or health for a project they manage
- **THEN** the system SHALL allow the update if the trusted identity subject matches the project owner or the user has administrator privileges

#### Scenario: Reject project owner outside project scope
- **WHEN** a project owner attempts to update a project, milestone, or task for a project owned by another user
- **THEN** the system SHALL reject the action as forbidden

#### Scenario: Read identity from trusted token claims
- **WHEN** an API request includes a valid production access token
- **THEN** the system SHALL use token role claims for operation category and token subject claims for ownership and participant checks

#### Scenario: Reject browser-controlled production identity headers
- **WHEN** a production API request attempts to set identity using browser-controlled `X-Role` or `X-User` headers without a valid trusted token
- **THEN** the system SHALL reject the request as unauthenticated or forbidden before executing the requested mutation

### Requirement: Missing update and milestone alerts
The system SHALL generate reminders or alerts for upcoming milestone due dates, overdue milestones, blocked milestones, missing weekly updates, and stale linked GitLab work beyond a configured threshold.

#### Scenario: Missing weekly update reminder
- **WHEN** a project has not received a weekly update by the expected review window
- **THEN** the system SHALL create a reminder event for the responsible owner

#### Scenario: Overdue milestone alert
- **WHEN** a milestone passes its planned date without completion
- **THEN** the system SHALL create an overdue alert that appears in review workflows and configured notification channels

### Requirement: Notification channels are decoupled from trigger logic
The system SHALL support notification delivery through configurable channels such as email and Feishu without embedding channel-specific behavior into project or milestone state transitions.

#### Scenario: Send same alert through multiple channels
- **WHEN** an overdue milestone alert is emitted
- **THEN** the system SHALL allow channel adapters to deliver that alert according to workspace configuration

### Requirement: Project-scoped contributor access
The system SHALL allow contributors to write only task and weekly update data for projects where their trusted identity subject is a participant, and SHALL deny contributor writes outside those projects.

#### Scenario: Contributor writes inside participating project
- **WHEN** a contributor submits a task or weekly update for a project whose participants include the trusted identity subject
- **THEN** the system SHALL allow the write if the operation also satisfies any task or author ownership rule

#### Scenario: Contributor denied outside participating project
- **WHEN** a contributor submits a task or weekly update for a project whose participants do not include the trusted identity subject
- **THEN** the system SHALL reject the write as forbidden

### Requirement: Development identity selector
The system SHALL expose a development-only current user selector, development token issuer, or equivalent configuration so local frontend requests can exercise role and user authorization paths without enabling browser-controlled identity in production.

#### Scenario: Send development identity with API request
- **WHEN** the frontend sends a write request during local development with development identity mode enabled
- **THEN** the request SHALL include either a development token or development-only identity headers accepted by the backend configuration

#### Scenario: Persist selected development identity
- **WHEN** a developer changes the selected current user in the frontend during local development
- **THEN** the frontend SHALL persist that value locally and reuse it on subsequent development requests

#### Scenario: Hide development selector in production
- **WHEN** the frontend runs with production authentication enabled
- **THEN** the frontend SHALL hide manual role and user selectors and derive the current user from authenticated session state

