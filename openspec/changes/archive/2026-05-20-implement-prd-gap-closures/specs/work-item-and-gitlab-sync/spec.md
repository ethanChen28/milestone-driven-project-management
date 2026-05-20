## ADDED Requirements

### Requirement: GitLab-linked work is visible in delivery views
The system SHALL display GitLab-linked work items in project and milestone views with source type, issue title, issue state, assignee, labels, sync freshness, and a link to the original GitLab issue when a source URL is available.

#### Scenario: View GitLab-linked issue on milestone detail
- **WHEN** a milestone has linked work items whose source type is `gitlab_issue`
- **THEN** the milestone detail view SHALL show their GitLab state, assignee, labels, last sync time, and original issue link

#### Scenario: Open original GitLab issue
- **WHEN** a user selects the original issue link for a GitLab-linked work item
- **THEN** the system SHALL open the source URL without changing the PM-owned milestone or project metadata

#### Scenario: Show non-GitLab work alongside GitLab work
- **WHEN** a project or milestone contains internal tasks, external dependencies, BAU tasks, and GitLab issues
- **THEN** the system SHALL show all linked work items and SHALL only render GitLab-specific metadata for GitLab issue items
