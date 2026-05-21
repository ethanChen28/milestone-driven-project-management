## Purpose
Defines workspace access control, alerts, notifications, and identity behavior.
## Requirements
### Requirement: Workspace role-based access control
The system SHALL support workspace roles `admin`, `portfolio_manager`, `project_owner`, `contributor`, and `viewer` with permissions aligned to roadmap, project, milestone, update, and integration actions, and SHALL combine role permissions with `X-User` identity for project-scoped and owner-scoped write authorization.

#### Scenario: Restrict integration administration
- **WHEN** a non-admin user attempts to manage GitLab integration settings
- **THEN** the system SHALL deny the action

#### Scenario: Allow project owners to update milestone state
- **WHEN** a project owner updates milestone status or health for a project they manage
- **THEN** the system SHALL allow the update if `X-User` matches the project owner or the user has administrator privileges

#### Scenario: Reject project owner outside project scope
- **WHEN** a project owner attempts to update a project, milestone, or task for a project owned by another user
- **THEN** the system SHALL reject the action as forbidden

#### Scenario: Read identity from request headers
- **WHEN** an API request includes `X-Role` and `X-User`
- **THEN** the system SHALL use `X-Role` for operation category and `X-User` for ownership and participant checks

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
The system SHALL allow contributors to write only task and weekly update data for projects where they are participants, and SHALL deny contributor writes outside those projects.

#### Scenario: Contributor writes inside participating project
- **WHEN** a contributor submits a task or weekly update for a project whose participants include `X-User`
- **THEN** the system SHALL allow the write if the operation also satisfies any task or author ownership rule

#### Scenario: Contributor denied outside participating project
- **WHEN** a contributor submits a task or weekly update for a project whose participants do not include `X-User`
- **THEN** the system SHALL reject the write as forbidden

### Requirement: Development identity selector
The system SHALL expose a development-only current user selector or equivalent configuration so frontend requests can send `X-User` consistently with `X-Role`.

#### Scenario: Send user identity with API request
- **WHEN** the frontend sends a write request during local development
- **THEN** the request SHALL include both `X-Role` and `X-User`

#### Scenario: Persist selected development identity
- **WHEN** a developer changes the selected current user in the frontend
- **THEN** the frontend SHALL persist that value locally and reuse it on subsequent API requests

