# Milestone-Driven Project Management System PRD

## Document info

- Title: Milestone-Driven Project Management System
- Version: v0.1
- Date: 2026-05-19
- Status: Draft
- Audience: founder, product, engineering, project owner

## 1. Background

The team does not want to operate in a task-first mode where members focus on completing assigned tickets without understanding the real outcome. The desired operating model is milestone-driven:

- roadmap defines strategic direction
- milestones define measurable stage outcomes
- projects organize delivery at the business or product level
- day-to-day work exists to move milestones forward
- GitLab remains the execution layer for engineering issues

This system should therefore become the project-level control plane above GitLab, not a replacement for GitLab issue management.

## 2. Product vision

Build a standalone project management system that helps the team plan and execute around milestones and outcomes, while syncing engineering work status from GitLab.

The system should answer these questions clearly:

1. What are we trying to achieve this quarter or this month?
2. Which milestones matter most right now?
3. What workstreams and issues are actually moving those milestones?
4. Which projects are blocked, at risk, or drifting away from roadmap goals?
5. How much of the team is doing milestone work versus BAU work?

## 3. Product goals

## 3.1 Business goals

- Shift team behavior from task completion to milestone ownership.
- Create a shared project layer for product, engineering, and operations.
- Keep project status visible without requiring manual status chasing.
- Preserve GitLab as the source of truth for engineering execution details.

## 3.2 User goals

- Leadership can see roadmap progress and project health.
- Project owners can define milestones, coordinate workstreams, and detect risk early.
- Engineers can keep using GitLab while still linking work to project outcomes.
- Cross-functional members can collaborate without needing to live inside GitLab.

## 3.3 Non-goals

- Replace GitLab issue boards, merge requests, or code review workflows.
- Build a generic all-in-one collaboration platform like Notion.
- Recreate every feature of Jira, Linear, or GitLab Enterprise planning modules in phase 1.

## 4. Core design principles

1. Milestone first: every important piece of work should map to a milestone or be explicitly classified as BAU.
2. Outcome over activity: progress is measured by milestone movement, not task count.
3. Project-level visibility: management view must exist above issue view.
4. Minimal double entry: engineering continues to manage technical execution in GitLab.
5. Human-readable state: status should be understandable by product, engineering, and business stakeholders.
6. Explicit ownership: each project and milestone must have a clear owner.

## 5. Target users

## 5.1 Leadership

Needs:

- see roadmap progress
- compare projects by status, value, and risk
- review milestone delays and cross-team dependencies

## 5.2 Project owner / product owner

Needs:

- create projects and milestones
- define success criteria
- track workstreams
- review linked GitLab issue progress
- run weekly reviews and milestone updates

## 5.3 Engineering manager / tech lead

Needs:

- connect implementation issues to milestones
- identify whether current issue work supports key outcomes
- review delivery risk from GitLab signal

## 5.4 Team member

Needs:

- understand why current work matters
- know which milestone a task supports
- see blockers, dependencies, and target dates

## 6. Core concepts and object model

The system should use the following primary objects.

### 6.1 Roadmap

Represents medium-term direction, usually quarterly or half-yearly.

Example:

- Q3 onboarding closed loop
- Q3 first 10 paying customers supported
- Q4 analytics capability launch

Suggested fields:

- id
- title
- description
- period_start
- period_end
- owner
- status
- priority

### 6.2 Project

Represents a bounded initiative under a roadmap theme. A project can contain multiple milestones and workstreams.

Suggested fields:

- id
- name
- summary
- objective
- roadmap_id
- owner
- participants
- project_type
- status
- health_status
- target_start_date
- target_end_date
- actual_end_date
- priority
- tags

### 6.3 Milestone

Represents a concrete stage outcome. This is the most important object in the system.

Suggested fields:

- id
- project_id
- title
- milestone_type
- description
- completion_criteria
- owner
- planned_date
- forecast_date
- completed_date
- status
- health_status
- progress_percent
- risk_level
- dependency_summary

Milestone examples:

- first customer can complete onboarding independently
- beta feature is live and stable for real users
- operations team can produce weekly report from new data pipeline

### 6.4 Workstream

Represents a delivery lane for a milestone, such as product, backend, frontend, operations, or customer enablement.

Suggested fields:

- id
- project_id
- milestone_id
- name
- owner
- status
- description

### 6.5 Linked work item

Represents an execution item that may come from either the internal system or GitLab.

Types:

- gitlab_issue
- internal_task
- external_dependency
- bau_task

Suggested fields:

- id
- source_type
- source_id
- source_url
- title
- project_id
- milestone_id
- workstream_id
- owner
- status
- estimate
- due_date
- blocked_flag

### 6.6 Weekly update

Represents a structured status update for project review.

Suggested fields:

- id
- project_id
- milestone_id optional
- author
- week
- summary
- progress
- risk
- blockers
- decisions_needed
- next_steps

## 7. High-level workflows

## 7.1 Roadmap planning workflow

1. Create roadmap period.
2. Define outcome-oriented roadmap items.
3. Create projects under roadmap items.
4. Break each project into milestones.
5. Assign owners and target dates.

## 7.2 Project setup workflow

1. Create project.
2. Define project objective and success metrics.
3. Add milestone list.
4. Define completion criteria for each milestone.
5. Create workstreams.
6. Link GitLab groups or repositories if needed.

## 7.3 Execution workflow

1. Team creates or updates engineering issues in GitLab.
2. Relevant GitLab issues are linked to milestones or workstreams.
3. System syncs issue status, assignee, labels, due dates, and MR references.
4. Project owner reviews milestone movement instead of raw task volume.
5. Weekly update highlights risk, blockers, and decisions.

## 7.4 Review workflow

1. Weekly project review shows milestone status.
2. Delayed or blocked milestones are escalated.
3. Team decides whether to re-scope, de-prioritize, or add support.
4. Closed milestones roll up into roadmap progress.

## 8. Functional requirements

## 8.1 Roadmap management

The system shall:

- create, edit, archive roadmap periods
- create roadmap items with business-oriented descriptions
- link projects to roadmap items
- show roadmap progress based on project and milestone status

## 8.2 Project management

The system shall:

- create and edit projects
- define project objective and expected outcome
- assign project owner and collaborators
- categorize projects by type, priority, and department
- support project health states such as `on_track`, `at_risk`, `off_track`, `done`
- provide project timeline view

## 8.3 Milestone management

The system shall:

- create and edit milestones under projects
- require completion criteria before a milestone can move to active
- assign milestone owner
- support milestone status such as `not_started`, `active`, `blocked`, `completed`, `cancelled`
- support manual health status separate from progress percent
- roll up linked work status into milestone progress assistance
- allow milestones without engineering work when the milestone is operational or business-oriented

## 8.4 Workstream management

The system shall:

- create workstreams under project or milestone
- assign owners
- group linked work items by workstream
- show whether each workstream is helping or blocking milestone progress

## 8.5 Work item management

The system shall:

- support linked work from GitLab issues
- support internal non-GitLab work items for product, ops, or business tasks
- classify work items as milestone work or BAU work
- require each non-BAU work item to belong to a project
- optionally require each milestone work item to belong to a milestone

## 8.6 Weekly update and review

The system shall:

- allow structured weekly updates
- surface unresolved blockers and overdue milestones
- provide review views by project owner, by roadmap, and by team
- preserve update history for audit and learning

## 8.7 Dashboard and reporting

The system shall provide:

- roadmap overview dashboard
- project portfolio dashboard
- milestone status dashboard
- overdue and blocked report
- milestone work versus BAU work ratio
- owner workload summary
- GitLab-linked execution status summary

## 8.8 Search and filtering

The system shall support filtering by:

- roadmap period
- project
- milestone
- owner
- team
- status
- health
- risk
- source type
- GitLab group or repository

## 9. GitLab integration requirements

## 9.1 Integration objective

GitLab should remain the engineering execution system. The standalone PM system should consume GitLab issue signals and map them into project and milestone context.

## 9.2 Source-of-truth boundaries

GitLab is source of truth for:

- issue title and description
- assignee
- labels
- issue state
- merge request linkage
- code and deployment activity

The PM system is source of truth for:

- roadmap
- project definition
- milestone definition
- completion criteria
- project health
- milestone health
- cross-functional work that does not live in GitLab

## 9.3 Required integration capabilities

The system shall:

- connect one or more GitLab groups or repositories
- sync issues by project, label, milestone, assignee, or query rule
- support manual link and auto-link between GitLab issue and PM objects
- pull issue metadata on scheduled sync or webhook events
- display GitLab issue state inside project and milestone views
- open the original GitLab issue directly from the PM system

## 9.4 Optional phase 2 GitLab capabilities

- sync merge request status
- show issue cycle time and lead time
- infer engineering risk from stale issues or unmerged MRs
- sync comments or activity timeline summary
- create GitLab issue from PM system
- push project or milestone references back into GitLab labels or comments

## 9.5 Recommended mapping model

Recommended relationship:

- one project maps to many GitLab issues
- one milestone maps to many GitLab issues
- one GitLab issue belongs to at most one primary milestone
- one GitLab issue may optionally belong to one workstream

This prevents the same issue from pretending to make progress on multiple milestones.

## 10. Key screens

## 10.1 Roadmap overview

Shows:

- roadmap items
- linked projects
- milestone summary
- health indicators
- target dates

## 10.2 Project detail page

Shows:

- project objective
- owner and team
- milestone timeline
- workstream breakdown
- linked GitLab issue summary
- weekly updates
- risks and blockers

## 10.3 Milestone detail page

Shows:

- milestone description
- completion criteria
- owner
- planned and forecast dates
- linked work items
- linked GitLab issues
- current blockers
- recent updates

## 10.4 Portfolio dashboard

Shows:

- all active projects
- health distribution
- delayed milestones
- dependency hotspots
- BAU versus milestone work ratio

## 10.5 Review view

Shows:

- weekly updates grouped by owner or roadmap
- unresolved decisions needed
- milestones with no meaningful movement

## 11. Permissions and roles

Suggested initial roles:

- admin: manage workspace, integrations, permissions
- portfolio_manager: manage roadmap, projects, dashboards
- project_owner: manage own projects, milestones, updates
- contributor: update work items and weekly notes
- viewer: read-only access

Permission rules:

- only admins can manage GitLab integration settings
- only portfolio managers and project owners can create roadmap and project objects
- project owners can edit milestone status and health
- contributors can update linked internal tasks and weekly updates

## 12. Notifications and automation

The system should support:

- reminder for upcoming milestone due date
- alert for overdue milestone
- alert for blocked milestone
- reminder when weekly update is missing
- notification when linked GitLab issues become stale beyond threshold

Potential channels:

- email
- Feishu bot
- GitLab comment or webhook callback in later phases

## 13. Non-functional requirements

## 13.1 Reliability

- GitLab sync failures should be retryable and visible.
- Sync should not corrupt manually maintained PM data.

## 13.2 Usability

- non-engineering users must understand the system without GitLab expertise
- project and milestone creation flow should be simple enough for first-time owners

## 13.3 Performance

- portfolio dashboard should load within acceptable interactive time for dozens to low hundreds of projects
- GitLab sync should handle incremental updates efficiently

## 13.4 Auditability

- changes to project health, milestone status, and target dates should be traceable

## 14. MVP scope

Recommended phase 1 MVP:

- roadmap management
- project management
- milestone management
- workstream management
- weekly updates
- portfolio dashboard
- basic GitLab issue linking
- manual and scheduled GitLab issue sync
- simple notifications for overdue and blocked milestones

Explicitly excluded from MVP:

- full resource management
- budget tracking
- complex dependency graphs
- deep MR analytics
- advanced OKR management
- custom workflow engine

## 15. Success metrics

Product success can be measured by:

- percentage of active work linked to a project
- percentage of project work linked to a milestone
- percentage of milestones with explicit completion criteria
- weekly update completion rate
- reduction in "orphan work" not tied to roadmap priorities
- shorter time to identify blocked projects

## 16. Risks and open questions

Key risks:

- teams may still misuse milestones as renamed task buckets
- too much manual project upkeep may cause adoption failure
- unclear boundary between BAU and milestone work may distort reporting
- GitLab mapping rules may become noisy without disciplined labels or ownership

Open questions:

1. Should internal non-engineering tasks live in this system only, or also sync to another tool?
2. Is milestone progress manual-first, automatic-first, or hybrid?
3. Should one project support multiple GitLab repositories by default?
4. Do we need tenant or workspace separation from day one?
5. Should Feishu be the default notification channel?

## 17. Product positioning summary

This product is best described as:

- a milestone-driven project operating system
- a project and outcome layer above GitLab
- a coordination tool for cross-functional delivery

It is not:

- a replacement for GitLab development workflow
- a pure ticket tracker
- a generic documentation tool

## 18. Recommended next step

After this PRD, the next design document should define:

1. detailed information architecture
2. domain model and table schema
3. GitLab sync mechanism
4. MVP page list and interaction flow
5. milestone health calculation rules
