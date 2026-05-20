## ADDED Requirements

### Requirement: Roadmap periods and items
The system SHALL allow portfolio managers to create, edit, archive, and view roadmap periods and outcome-oriented roadmap items within each period.

#### Scenario: Create roadmap item in an active period
- **WHEN** a portfolio manager creates a roadmap item with title, owner, period boundaries, and priority
- **THEN** the system stores the roadmap item under the selected roadmap period and makes it available for project linkage

#### Scenario: Archive roadmap period without losing history
- **WHEN** a portfolio manager archives a roadmap period
- **THEN** the system marks the period as archived and preserves all linked roadmap items and project relationships for historical reporting

### Requirement: Projects link to roadmap goals
The system SHALL allow each project to link to a roadmap item so that delivery work can be traced to medium-term outcomes.

#### Scenario: Link project to roadmap item
- **WHEN** a project owner or portfolio manager assigns a roadmap item to a project
- **THEN** the project detail and roadmap views SHALL show that relationship consistently

### Requirement: Roadmap progress rollup
The system SHALL summarize roadmap progress from the status and health of linked projects and milestones.

#### Scenario: View roadmap progress summary
- **WHEN** a user opens a roadmap overview
- **THEN** the system SHALL display linked projects, milestone summaries, health indicators, and target dates for each roadmap item
