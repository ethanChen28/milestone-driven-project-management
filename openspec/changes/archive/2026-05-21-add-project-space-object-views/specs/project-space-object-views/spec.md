## ADDED Requirements

### Requirement: Project space provides dedicated object views
The system SHALL provide a project-space navigation model for each project with dedicated views for overview, work items, milestones, weekly updates, risks, dependencies, and settings.

#### Scenario: Open project space default overview
- **WHEN** a user opens a project detail page without selecting a project-space view
- **THEN** the system SHALL show the project overview view and expose navigation to the other project-space views

#### Scenario: Switch project-space view
- **WHEN** a user selects the work items, milestones, weekly updates, risks, dependencies, or settings view
- **THEN** the system SHALL update the visible content to that view without losing project context

### Requirement: Overview summarizes without duplicating full management views
The system SHALL limit the project overview to rollups, summary cards, current milestone summaries, top risk signals, recent updates, and key blocked/overdue indicators.

#### Scenario: Overview shows summary cards
- **WHEN** a user opens the project overview
- **THEN** the system SHALL show aggregate project health, timeline, milestone progress, work item counts, recent updates, and top risks

#### Scenario: Overview avoids duplicate full object tables
- **WHEN** a user needs to manage all milestones, all work items, all updates, all risks, or all dependencies for a project
- **THEN** the system SHALL direct the user to the owning project-space view instead of duplicating that full table in overview

### Requirement: Summary objects support quick-filter transitions
The system SHALL allow summary objects in overview to navigate into the relevant full project-space view with context-preserving filters.

#### Scenario: Open work items filtered by milestone
- **WHEN** a user selects a milestone summary action to inspect its execution work
- **THEN** the system SHALL open the project work-items view filtered to that milestone

#### Scenario: Open source object from summary
- **WHEN** a user selects a milestone, risk, dependency, or update title from overview
- **THEN** the system SHALL open the source object's detail view or owning project-space view with the selected object in context

### Requirement: Project space preserves global entry points
The system SHALL preserve global project, work item, milestone, roadmap, and weekly review entry points while adding project-scoped views.

#### Scenario: Navigate from global work item view to project context
- **WHEN** a user opens a work item that belongs to a project from the global work item workspace
- **THEN** the system SHALL show project and milestone breadcrumbs that allow navigation back to the relevant project space

#### Scenario: Navigate from project space to global workspace
- **WHEN** a user wants to inspect work across all projects from a project-scoped view
- **THEN** the system SHALL provide a route to the global workspace without changing the underlying work item data

### Requirement: Project space includes relationship-focused risk and dependency views
The system SHALL provide project-scoped risk and dependency views that focus on blockers, decision needs, external dependencies, and cross-object relationships.

#### Scenario: View project risks
- **WHEN** a user opens the project risks view
- **THEN** the system SHALL show risk signals associated with the project, its milestones, its work items, and recent weekly updates

#### Scenario: View project dependencies
- **WHEN** a user opens the project dependencies view
- **THEN** the system SHALL show blocked work, external dependency work items, milestone dependency summaries, and related source objects
