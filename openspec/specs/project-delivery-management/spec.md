## Purpose
Defines project, milestone, delivery lane, and project-space delivery management behavior.
## Requirements
### Requirement: Project lifecycle and ownership
The system SHALL allow authorized users to create and maintain projects with objective, owner, participants, type, priority, timeline, and health fields.

#### Scenario: Create project with outcome context
- **WHEN** a project owner creates a project with name, objective, owner, target dates, and project type
- **THEN** the system SHALL create the project and expose it in portfolio and roadmap-linked views

#### Scenario: Update project health
- **WHEN** a project owner changes a project's health to `at_risk` or `off_track`
- **THEN** the system SHALL save the new health state and surface it in dashboard and review views

### Requirement: Milestones require outcome definition
The system SHALL require milestones to include owner, planned date, status, and completion criteria before they can move to an active state, and SHALL treat completion criteria as checklist-oriented acceptance items that guide final milestone completion.

#### Scenario: Prevent activation without completion criteria
- **WHEN** a user attempts to change a milestone status from `not_started` to `active` without completion criteria
- **THEN** the system SHALL reject the transition and explain that completion criteria are required

#### Scenario: Display completion criteria as acceptance checklist
- **WHEN** a user opens a milestone with multiline completion criteria
- **THEN** the system SHALL present each non-empty line as an acceptance checklist item

#### Scenario: Complete milestone with recorded date
- **WHEN** a project owner marks a milestone as completed after confirming the acceptance criteria
- **THEN** the system SHALL store the completed date and reflect the change in project and roadmap rollups

#### Scenario: Prevent contributor milestone completion
- **WHEN** a contributor attempts to mark a milestone as `completed`
- **THEN** the system SHALL reject the transition because milestone completion is an acceptance decision

### Requirement: Workstreams organize delivery lanes
The system SHALL preserve existing Workstream records for compatibility, but SHALL freeze Workstream as an inactive delivery lane and SHALL NOT require users to create, select, or manage Workstreams in active project, milestone, or task workflows.

#### Scenario: Existing workstream data remains readable
- **WHEN** existing data contains a Workstream or a task with `workstreamId`
- **THEN** the system SHALL preserve and return that data without requiring users to edit it

#### Scenario: Create task without workstream
- **WHEN** a user creates a milestone-linked task
- **THEN** the system SHALL allow the task to be saved without a Workstream association

#### Scenario: Hide workstream from active task editing
- **WHEN** a user opens the task creation or editing form
- **THEN** the system SHALL NOT show Workstream as a required or primary editable field

### Requirement: Milestone progress assistance does not replace human judgment
The system SHALL maintain milestone status, health, and progress percent as separate fields, and SHALL allow linked work state to inform progress assistance without automatically overriding manual health.

#### Scenario: Linked work changes while milestone health remains manual
- **WHEN** linked work items move forward in execution status
- **THEN** the system MAY update milestone progress assistance but SHALL NOT overwrite a manually set milestone health status

### Requirement: Three-level delivery domain
The system SHALL treat Project > Milestone > Task as the active delivery hierarchy for Agent-assisted project management.

#### Scenario: Show task under project and milestone
- **WHEN** a task is linked to both a project and milestone
- **THEN** the system SHALL show it in project, milestone, and task workspace views without requiring a Workstream grouping

#### Scenario: Keep milestone as acceptance boundary
- **WHEN** task execution status changes to `done`
- **THEN** the system SHALL NOT automatically complete the milestone without an authorized milestone completion action

### Requirement: Project detail separates overview from full milestone management
The system SHALL separate project overview rollups from full milestone management within the project detail experience.

#### Scenario: View milestone summaries in overview
- **WHEN** a user opens a project overview with milestones
- **THEN** the system SHALL show milestone summaries with status, health, owner, planned date, progress, and linked work count

#### Scenario: Manage milestones from project milestone view
- **WHEN** a user creates, edits, transitions, filters, or reviews all milestones for a project
- **THEN** the system SHALL provide those actions in the project milestones view or milestone detail view, not as duplicated full management controls in overview

### Requirement: Project overview exposes delivery rollups
The system SHALL calculate and display project delivery rollups from project, milestone, work item, and update state.

#### Scenario: Show delivery rollups
- **WHEN** a user opens the project overview
- **THEN** the system SHALL show active milestone count, completed milestone count, blocked milestone count, overdue milestone count, work item status counts, health, and target date signals

#### Scenario: Rollup reflects underlying object changes
- **WHEN** a milestone, work item, or weekly update changes for a project
- **THEN** the project overview rollups SHALL reflect the updated state after data reload or refresh

### Requirement: Project space maintains object breadcrumbs
The system SHALL expose project and milestone context when users inspect project-scoped objects.

#### Scenario: Open milestone from project space
- **WHEN** a user opens a milestone from a project-space view
- **THEN** the system SHALL show a breadcrumb that includes the parent project and allows returning to the project space

#### Scenario: Open work item from project space
- **WHEN** a user opens a work item from a project-space view
- **THEN** the system SHALL show project and milestone context when those relationships exist

