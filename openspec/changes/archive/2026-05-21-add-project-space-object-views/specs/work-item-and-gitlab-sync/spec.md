## ADDED Requirements

### Requirement: Project-scoped work item views support grouping and filters
The system SHALL provide project-scoped work item views that can filter and group linked work items by milestone, status, priority, owner, blocked state, source type, and GitLab context.

#### Scenario: Filter project work items by milestone
- **WHEN** a user opens a project's work item view with a milestone filter
- **THEN** the system SHALL show only linked work items for that project and milestone

#### Scenario: Group project work items by milestone
- **WHEN** a user groups project work items by milestone
- **THEN** the system SHALL organize work items into milestone groups and include an unassigned group for project work without a milestone

#### Scenario: Group project work items by execution fields
- **WHEN** a user groups project work items by status, priority, owner, source type, or blocked state
- **THEN** the system SHALL group the same project-scoped work items without changing their stored source data

### Requirement: Project work item views preserve source-of-truth boundaries
The system SHALL preserve GitLab-owned and PM-owned field boundaries when linked work items are displayed or edited from project-space views.

#### Scenario: Display GitLab work in project space
- **WHEN** a GitLab-linked work item appears in a project-space work item view
- **THEN** the system SHALL show GitLab state, assignee, labels, source URL, and last synced time as synced execution context

#### Scenario: Edit PM-owned metadata from project space
- **WHEN** a project owner updates project, milestone, priority, owner, blocked flag, or PM-owned metadata for a linked work item
- **THEN** the system SHALL preserve GitLab-owned fields until a GitLab sync updates them

### Requirement: Work item breadcrumbs connect global and project-scoped context
The system SHALL show stable breadcrumbs for linked work items that identify project, milestone, source type, and source identifier where available.

#### Scenario: Show breadcrumb for milestone work item
- **WHEN** a user views a work item linked to both a project and milestone
- **THEN** the system SHALL show a breadcrumb containing the project name and milestone title

#### Scenario: Show breadcrumb for BAU work item
- **WHEN** a user views a BAU work item without project or milestone context
- **THEN** the system SHALL show source and owner context without implying project membership
