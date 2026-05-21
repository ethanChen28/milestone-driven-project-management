## Purpose
Defines weekly review, portfolio reporting, filtering, and project-space summary behavior.
## Requirements
### Requirement: Structured weekly updates
The system SHALL allow project owners and contributors to record weekly updates with summary, progress, risk, blockers, decisions needed, and next steps.

#### Scenario: Submit weekly update
- **WHEN** a project owner submits a weekly update for a project or milestone
- **THEN** the system SHALL store the update with author and week metadata and make it available in review history

### Requirement: Review views emphasize milestone movement
The system SHALL provide weekly review views grouped by owner, roadmap, and project that highlight milestone status, blockers, overdue dates, and unresolved decisions.

#### Scenario: Review delayed milestones
- **WHEN** a stakeholder opens the review view
- **THEN** the system SHALL show delayed or blocked milestones alongside recent updates and decisions needed

### Requirement: Portfolio and milestone dashboards
The system SHALL provide dashboards for roadmap overview, project portfolio, milestone status, blocked/overdue work, BAU versus milestone work ratio, owner workload, and GitLab-linked execution summaries.

#### Scenario: Open portfolio dashboard
- **WHEN** a leadership or portfolio user opens the portfolio dashboard
- **THEN** the system SHALL show all active projects, health distribution, delayed milestones, dependency hotspots, and milestone versus BAU work summaries

### Requirement: Search and filtering across planning views
The system SHALL support filtering planning and reporting views by roadmap period, project, milestone, owner, team, status, health, risk, source type, and GitLab repository context.

#### Scenario: Filter review by owner and health
- **WHEN** a user applies owner and health filters in a reporting view
- **THEN** the system SHALL return only matching projects, milestones, updates, and linked work summaries

### Requirement: Project overview includes recent weekly update summaries
The system SHALL expose recent project weekly updates in project overview while preserving the full weekly review view as the authoritative reporting workflow.

#### Scenario: Show recent updates in project overview
- **WHEN** a user opens the project overview
- **THEN** the system SHALL show the most recent weekly updates for that project with week, author, summary, risk, blockers, decisions needed, and next steps when available

#### Scenario: Open full update history
- **WHEN** a user wants to review all updates for a project or compare updates across projects
- **THEN** the system SHALL route the user to the project updates view or global weekly review view with relevant filters

### Requirement: Project overview highlights top risks and decision needs
The system SHALL surface top project risks, blockers, and decision needs derived from milestone state, blocked work items, dependency signals, and weekly updates.

#### Scenario: Show top project risks
- **WHEN** a project has blocked milestones, blocked work items, high risk milestones, external dependencies, or recent weekly update risks
- **THEN** the project overview SHALL show a prioritized summary of those risk signals

#### Scenario: Open risk source from overview
- **WHEN** a user selects a risk or decision signal from project overview
- **THEN** the system SHALL open the project risks view or source object with enough context to resolve the issue

### Requirement: Reporting filters remain consistent with project-space filters
The system SHALL keep project-space filters compatible with portfolio and weekly review filters for project, milestone, owner, status, health, risk, and source type.

#### Scenario: Apply project filter from reporting view
- **WHEN** a user navigates from a portfolio or weekly review view to a project-space view with a project filter
- **THEN** the project-space view SHALL preserve the selected project context and apply compatible filters where supported

#### Scenario: Compare project overview with portfolio dashboard
- **WHEN** a project appears in both the project overview and portfolio dashboard
- **THEN** shared health, blocked, overdue, update, and risk counts SHALL remain consistent after data refresh

