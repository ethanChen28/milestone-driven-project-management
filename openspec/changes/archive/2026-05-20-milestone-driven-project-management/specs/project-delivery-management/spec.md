## ADDED Requirements

### Requirement: Project lifecycle and ownership
The system SHALL allow authorized users to create and maintain projects with objective, owner, participants, type, priority, timeline, and health fields.

#### Scenario: Create project with outcome context
- **WHEN** a project owner creates a project with name, objective, owner, target dates, and project type
- **THEN** the system SHALL create the project and expose it in portfolio and roadmap-linked views

#### Scenario: Update project health
- **WHEN** a project owner changes a project's health to `at_risk` or `off_track`
- **THEN** the system SHALL save the new health state and surface it in dashboard and review views

### Requirement: Milestones require outcome definition
The system SHALL require milestones to include owner, planned date, status, and completion criteria before they can move to an active state.

#### Scenario: Prevent activation without completion criteria
- **WHEN** a user attempts to change a milestone status from `not_started` to `active` without completion criteria
- **THEN** the system SHALL reject the transition and explain that completion criteria are required

#### Scenario: Complete milestone with recorded date
- **WHEN** a project owner marks a milestone as completed
- **THEN** the system SHALL store the completed date and reflect the change in project and roadmap rollups

### Requirement: Workstreams organize delivery lanes
The system SHALL allow projects and milestones to contain workstreams with owner, description, and status fields.

#### Scenario: Group delivery work by workstream
- **WHEN** a user creates a workstream for a milestone
- **THEN** the system SHALL allow linked work items to reference that workstream for grouped display in milestone and project views

### Requirement: Milestone progress assistance does not replace human judgment
The system SHALL maintain milestone status, health, and progress percent as separate fields, and SHALL allow linked work state to inform progress assistance without automatically overriding manual health.

#### Scenario: Linked work changes while milestone health remains manual
- **WHEN** linked work items move forward in execution status
- **THEN** the system MAY update milestone progress assistance but SHALL NOT overwrite a manually set milestone health status
