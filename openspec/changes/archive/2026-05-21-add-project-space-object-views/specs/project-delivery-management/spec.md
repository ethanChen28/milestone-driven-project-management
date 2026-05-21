## ADDED Requirements

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
