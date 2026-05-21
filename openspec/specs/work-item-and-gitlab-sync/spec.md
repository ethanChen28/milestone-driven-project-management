## Purpose
Defines linked work item behavior and GitLab synchronization boundaries.
## Requirements
### Requirement: Linked work supports multiple source types
The system SHALL support linked work items with source types `gitlab_issue`, `internal_task`, `external_dependency`, and `bau_task`.

#### Scenario: Create internal non-GitLab work item
- **WHEN** a contributor creates a linked work item for product or operations work
- **THEN** the system SHALL store it without requiring a GitLab source identifier

### Requirement: Milestone work classification
The system SHALL require each non-BAU linked work item to belong to a project, SHALL allow milestone work items to be linked to a milestone, and SHALL treat any Workstream association as optional legacy metadata.

#### Scenario: Reject non-BAU work without project
- **WHEN** a user attempts to create a non-BAU linked work item without a project association
- **THEN** the system SHALL reject the request and explain that project linkage is required

#### Scenario: Link GitLab issue to milestone
- **WHEN** a user links a GitLab issue to a milestone
- **THEN** the system SHALL record the project and milestone association for that linked work item

#### Scenario: Preserve legacy workstream association
- **WHEN** an existing linked work item already has `workstreamId`
- **THEN** the system SHALL preserve that value as legacy metadata without requiring users to maintain it

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

### Requirement: Contributor-owned task autonomy
The system SHALL allow contributors to create, edit, and delete their own task work items inside projects where they are participants, and SHALL prevent contributors from changing or deleting tasks owned by another user.

#### Scenario: Contributor creates own task in participating project
- **WHEN** a contributor sends a task create request with `X-User` matching the task owner and the project includes that user in participants
- **THEN** the system SHALL create the task

#### Scenario: Contributor cannot edit another user's task
- **WHEN** a contributor sends a task update request for a task whose owner differs from `X-User`
- **THEN** the system SHALL reject the request as forbidden

#### Scenario: Contributor cannot delete another user's task
- **WHEN** a contributor sends a task delete request for a task whose owner differs from `X-User`
- **THEN** the system SHALL reject the request as forbidden and preserve the task

### Requirement: Project owner task oversight
The system SHALL allow project owners to create, edit, and delete task work items within projects they own, regardless of the task owner.

#### Scenario: Project owner edits project task
- **WHEN** a project owner sends a task update request for a task in a project where `Project.owner` equals `X-User`
- **THEN** the system SHALL allow the update

#### Scenario: Project owner cannot edit task in another project
- **WHEN** a project owner sends a task update request for a task in a project owned by another user
- **THEN** the system SHALL reject the request as forbidden

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

