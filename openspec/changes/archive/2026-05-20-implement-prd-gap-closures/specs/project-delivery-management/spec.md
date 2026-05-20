## MODIFIED Requirements

### Requirement: Milestones require outcome definition
The system SHALL require milestones to include owner, planned date, status, and completion criteria before they can move to an active state, and SHALL allow authorized users to manage milestone lifecycle state from project and milestone workflows.

Valid milestone state transitions: `not_started → active` (requires completion criteria), `active → blocked`, `active → completed` (auto-sets completedDate), `active → cancelled`, `blocked → active`. Transitions to or from `completed` and `cancelled` from any non-adjacent state are rejected. `not_started → completed` (skipping active) is rejected.

#### Scenario: Prevent activation without completion criteria
- **WHEN** a user attempts to change a milestone status from `not_started` to `active` without completion criteria
- **THEN** the system SHALL reject the transition and explain that completion criteria are required

#### Scenario: Complete milestone with recorded date
- **WHEN** a project owner marks a milestone as completed
- **THEN** the system SHALL store the completed date and reflect the change in project and roadmap rollups

#### Scenario: Update milestone status from detail view
- **WHEN** an authorized project owner changes a milestone status to `active`, `blocked`, `completed`, or `cancelled` from the milestone detail workflow
- **THEN** the system SHALL persist the new status and refresh project, milestone, roadmap, dashboard, and review views with the updated state

#### Scenario: Reject invalid terminal-state transition
- **WHEN** a user attempts to move a `completed` or `cancelled` milestone back to an active delivery state without an approved reopening workflow
- **THEN** the system SHALL reject the transition and keep the existing terminal state unchanged

#### Scenario: Update milestone health and progress separately
- **WHEN** an authorized user updates milestone health or progress percent without changing lifecycle status
- **THEN** the system SHALL persist the requested health or progress values without overwriting milestone status
