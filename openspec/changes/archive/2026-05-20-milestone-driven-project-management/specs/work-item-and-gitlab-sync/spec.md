## ADDED Requirements

### Requirement: Linked work supports multiple source types
The system SHALL support linked work items with source types `gitlab_issue`, `internal_task`, `external_dependency`, and `bau_task`.

#### Scenario: Create internal non-GitLab work item
- **WHEN** a contributor creates a linked work item for product or operations work
- **THEN** the system SHALL store it without requiring a GitLab source identifier

### Requirement: Milestone work classification
The system SHALL require each non-BAU linked work item to belong to a project, and SHALL allow milestone work items to be linked to a milestone and optional workstream.

#### Scenario: Reject non-BAU work without project
- **WHEN** a user attempts to create a non-BAU linked work item without a project association
- **THEN** the system SHALL reject the request and explain that project linkage is required

#### Scenario: Link GitLab issue to milestone
- **WHEN** a user links a GitLab issue to a milestone
- **THEN** the system SHALL record the project, milestone, and optional workstream association for that linked work item

### Requirement: GitLab issue sync preserves source-of-truth boundaries
The system SHALL sync GitLab-owned fields such as title, description, labels, assignee, state, and merge request references into linked work items without overwriting PM-owned milestone or project metadata.

#### Scenario: Scheduled sync updates execution metadata
- **WHEN** a scheduled or webhook-triggered sync receives updated GitLab issue state
- **THEN** the system SHALL update only the synced execution fields on the linked work item and preserve PM-owned fields unchanged

#### Scenario: Sync failure remains visible
- **WHEN** a GitLab sync attempt fails
- **THEN** the system SHALL record a retryable failure state that is visible to administrators

### Requirement: GitLab association supports manual and rule-based linking
The system SHALL support both manual linking of GitLab issues and rule-based association using GitLab group, repository, label, assignee, milestone, or query filters.

#### Scenario: Auto-link rule attaches matching issues
- **WHEN** a sync rule matches a GitLab issue for a project
- **THEN** the system SHALL create or update the corresponding linked work item and show its GitLab origin in project and milestone views
