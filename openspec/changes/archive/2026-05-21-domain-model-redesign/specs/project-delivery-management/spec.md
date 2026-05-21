## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Three-level delivery domain
The system SHALL treat Project > Milestone > Task as the active delivery hierarchy for Agent-assisted project management.

#### Scenario: Show task under project and milestone
- **WHEN** a task is linked to both a project and milestone
- **THEN** the system SHALL show it in project, milestone, and task workspace views without requiring a Workstream grouping

#### Scenario: Keep milestone as acceptance boundary
- **WHEN** task execution status changes to `done`
- **THEN** the system SHALL NOT automatically complete the milestone without an authorized milestone completion action
