## ADDED Requirements

### Requirement: Workspace role-based access control
The system SHALL support workspace roles `admin`, `portfolio_manager`, `project_owner`, `contributor`, and `viewer` with permissions aligned to roadmap, project, milestone, update, and integration actions.

#### Scenario: Restrict integration administration
- **WHEN** a non-admin user attempts to manage GitLab integration settings
- **THEN** the system SHALL deny the action

#### Scenario: Allow project owners to update milestone state
- **WHEN** a project owner updates milestone status or health for a project they manage
- **THEN** the system SHALL allow the update if the user has ownership or equivalent role permissions

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
