## ADDED Requirements

### Requirement: Browser role context is explicit
The system SHALL allow browser users in the MVP environment to choose one of the supported workspace roles and SHALL send that role with API requests using the backend-supported role header. The role selector is an MVP debug/development tool, not a production login mechanism. Default role is `contributor` so permission boundaries are exercised by default.

#### Scenario: Select workspace role for requests
- **WHEN** a browser user selects `project_owner` as the current role and submits a milestone update
- **THEN** the request SHALL include the selected role and SHALL NOT use a hard-coded `admin` role

#### Scenario: Preserve selected role during navigation
- **WHEN** a browser user selects a workspace role and navigates between application pages
- **THEN** subsequent API requests SHALL continue using the selected role until the user changes it

### Requirement: Permission-aware browser actions
The system SHALL hide, disable, or explain mutating browser actions that the selected workspace role is not allowed to perform while retaining backend authorization as the final enforcement point.

#### Scenario: Non-admin cannot access GitLab configuration actions
- **WHEN** the selected browser role is not `admin`
- **THEN** GitLab integration management actions SHALL be unavailable or clearly denied in the UI, and backend API calls for those actions SHALL still be rejected

#### Scenario: Viewer sees read-only planning views
- **WHEN** the selected browser role is `viewer`
- **THEN** roadmap, project, milestone, weekly update, and GitLab-linked work views SHALL remain readable but creation and update actions SHALL be unavailable or rejected
