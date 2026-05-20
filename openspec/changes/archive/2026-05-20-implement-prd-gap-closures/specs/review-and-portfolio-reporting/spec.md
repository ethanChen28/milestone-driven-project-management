## MODIFIED Requirements

### Requirement: Structured weekly updates
The system SHALL allow project owners and contributors to record weekly updates with summary, progress, risk, blockers, decisions needed, and next steps, using selectable project and optional milestone context where browser UI is available.

#### Scenario: Submit weekly update
- **WHEN** a project owner submits a weekly update for a project or milestone
- **THEN** the system SHALL store the update with author and week metadata and make it available in review history

#### Scenario: Submit weekly update from browser with project context
- **WHEN** an authorized browser user selects a project, optionally selects one of that project's milestones, enters weekly update details, and submits the update
- **THEN** the system SHALL create the weekly update, refresh the review view, and show the update under the selected project or milestone context

#### Scenario: Prevent weekly update without project context
- **WHEN** a user submits a weekly update without a project association
- **THEN** the system SHALL reject the update and explain that project context is required

### Requirement: Search and filtering across planning views
The system SHALL support filtering planning and reporting views by roadmap period, project, milestone, owner, status, health, risk, source type, and GitLab repository context through API query parameters and browser controls where those views are available. Team filtering is out of scope for this change since the system serves a single team.

#### Scenario: Filter review by owner and health
- **WHEN** a user applies owner and health filters in a reporting view
- **THEN** the system SHALL return only matching projects, milestones, updates, and linked work summaries

#### Scenario: Filter milestones by risk
- **WHEN** a user applies a risk filter to milestone or reporting views
- **THEN** the system SHALL return only milestones and related summaries whose risk level matches the requested risk

#### Scenario: Filter linked work by GitLab context
- **WHEN** a user filters project or milestone work by GitLab repository or source type
- **THEN** the system SHALL return only matching linked work items while keeping non-matching work out of the displayed summary

#### Scenario: Clear filters restores full view
- **WHEN** a user clears all filters in a planning or reporting view
- **THEN** the system SHALL reload the unfiltered data for that view
