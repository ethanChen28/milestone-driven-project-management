## ADDED Requirements

### Requirement: Unified task workspace
The system SHALL provide a unified task management workspace that can switch between task list, status board, Gantt view, timeline view, by-project view, and by-priority view while operating on the same task dataset.

#### Scenario: Switch task views without losing context
- **WHEN** a user switches from the task list view to the Gantt view
- **THEN** the system SHALL preserve the current task filters and selected grouping
- **AND** the system SHALL continue to show the same underlying task set in the new view

#### Scenario: Highlight active view
- **WHEN** a user opens the task workspace
- **THEN** the system SHALL highlight the currently selected view so the user can tell which perspective is active

### Requirement: Advanced task querying
The system SHALL allow users to filter, search, group, and sort tasks by project, status, owner, priority, source type, tag, milestone, and text keyword.

#### Scenario: Combine multiple filters
- **WHEN** a user applies project, status, and owner filters together
- **THEN** the system SHALL return only tasks matching all active filters

#### Scenario: Sort task results
- **WHEN** a user selects a sort order by due date, created time, or updated time
- **THEN** the system SHALL order the visible tasks according to the selected field and direction

#### Scenario: Group tasks by delivery dimension
- **WHEN** a user groups tasks by project, status, priority, or a custom field
- **THEN** the system SHALL aggregate the visible tasks under the chosen grouping and allow groups to be collapsed or expanded

### Requirement: Task status board
The system SHALL provide a task status board that organizes tasks by execution state, including not started, in progress, done, and blocked or cancelled states.

#### Scenario: Show tasks in status columns
- **WHEN** a user opens the status board view
- **THEN** the system SHALL display tasks in columns based on their current status

#### Scenario: Move a task between states
- **WHEN** a user updates a task status from not started to in progress
- **THEN** the system SHALL update the task's placement in the board and keep the new state visible after refresh

### Requirement: Task schedule visualization
The system SHALL provide a schedule visualization that shows task duration over time in Gantt or timeline form, including a current-day marker and support for week, month, quarter, and year scale changes.

#### Scenario: Display task duration bars
- **WHEN** a user opens the Gantt view
- **THEN** the system SHALL render each scheduled task as a bar whose length corresponds to its planned duration

#### Scenario: Show current time position
- **WHEN** the user views the schedule visualization on a given day
- **THEN** the system SHALL display a current-day marker aligned to the visible time axis

#### Scenario: Change time scale
- **WHEN** a user switches the time scale from month to quarter
- **THEN** the system SHALL re-render the schedule view using the selected scale without changing the task selection state

### Requirement: Summary metrics and risk visibility
The system SHALL display summary cards for task totals and health indicators, including total, completed, in progress, not started, overdue, blocked, and near-due counts, and SHALL surface task source and risk cues.

#### Scenario: Open the task overview
- **WHEN** a user opens the task overview section
- **THEN** the system SHALL show summary metrics for the current filtered task set

#### Scenario: Drill down from a metric card
- **WHEN** a user clicks the overdue task card
- **THEN** the system SHALL apply an overdue filter and show the matching tasks

#### Scenario: Display source and risk cues
- **WHEN** a task originates from GitLab, an internal task, or an external dependency
- **THEN** the system SHALL expose the source type and any blocked or risk indicator in the task presentation
