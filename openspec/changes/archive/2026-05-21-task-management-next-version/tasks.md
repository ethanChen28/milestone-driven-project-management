## 1. Workspace Shell

- [x] 1.1 Add a unified `/tasks` workspace route with top-level view tabs for list, status board, Gantt, timeline, by-project, and by-priority views
- [x] 1.2 Define the shared task workspace state model for selected view, filters, grouping, and sort order
- [x] 1.3 Sync the workspace state to URL query parameters so refresh and sharing preserve the current context

## 2. Shared Data Layer

- [x] 2.1 Normalize task query parameters into a reusable client-side query builder
- [x] 2.2 Add support for combined filters across project, status, owner, priority, source type, milestone, tag, and keyword
- [x] 2.3 Add grouping and sorting adapters that can drive list, board, and timeline views from the same result set

## 3. Task Views

- [x] 3.1 Implement the default task list view using the shared workspace state
- [x] 3.2 Implement the status board view with grouped status columns and empty-state handling
- [x] 3.3 Implement the Gantt view with week, month, quarter, and year scale switching
- [x] 3.4 Implement the timeline view with milestone markers and overdue highlighting
- [x] 3.5 Implement the by-project and by-priority summary views using the same filtered dataset

## 4. Insights and Risk Signals

- [x] 4.1 Add summary metric cards for total, completed, in progress, not started, overdue, blocked, and near-due tasks
- [x] 4.2 Make metric cards drill down into the matching filtered task set
- [x] 4.3 Add visible source, blocked, and risk badges for GitLab issue, internal task, and external dependency records
- [x] 4.4 Add current-day marker and due-date emphasis to schedule-based views

## 5. Verification

- [x] 5.1 Add unit tests for workspace state persistence, query composition, and card drill-down behavior
- [x] 5.2 Add component tests for view switching, grouping, and empty/loading states
- [x] 5.3 Add end-to-end coverage for the main task navigation flow, cross-view filter preservation, and summary card filtering
