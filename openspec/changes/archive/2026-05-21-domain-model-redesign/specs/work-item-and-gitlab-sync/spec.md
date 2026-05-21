## MODIFIED Requirements

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

## ADDED Requirements

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
