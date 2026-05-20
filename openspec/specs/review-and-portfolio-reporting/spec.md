## ADDED Requirements

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
