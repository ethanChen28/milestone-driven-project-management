package domain

import "time"

type AuditFields struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type HealthStatus string

const (
	HealthOnTrack  HealthStatus = "on_track"
	HealthAtRisk   HealthStatus = "at_risk"
	HealthOffTrack HealthStatus = "off_track"
	HealthDone     HealthStatus = "done"
)

type MilestoneStatus string

const (
	MilestoneNotStarted MilestoneStatus = "not_started"
	MilestoneActive     MilestoneStatus = "active"
	MilestoneBlocked    MilestoneStatus = "blocked"
	MilestoneCompleted  MilestoneStatus = "completed"
	MilestoneCancelled  MilestoneStatus = "cancelled"
)

type WorkItemSourceType string

const (
	SourceGitLabIssue        WorkItemSourceType = "gitlab_issue"
	SourceInternalTask       WorkItemSourceType = "internal_task"
	SourceExternalDependency WorkItemSourceType = "external_dependency"
	SourceBAUTask            WorkItemSourceType = "bau_task"
)

type WorkspaceRole string

const (
	RoleAdmin            WorkspaceRole = "admin"
	RolePortfolioManager WorkspaceRole = "portfolio_manager"
	RoleProjectOwner     WorkspaceRole = "project_owner"
	RoleContributor      WorkspaceRole = "contributor"
	RoleViewer           WorkspaceRole = "viewer"
)

type RoadmapPeriod struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Owner       string      `json:"owner"`
	Status      string      `json:"status"`
	Priority    string      `json:"priority"`
	PeriodStart time.Time   `json:"periodStart"`
	PeriodEnd   time.Time   `json:"periodEnd"`
	AuditFields AuditFields `json:"audit"`
}

type RoadmapItem struct {
	ID          string      `json:"id"`
	PeriodID    string      `json:"periodId"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Owner       string      `json:"owner"`
	Priority    string      `json:"priority"`
	Status      string      `json:"status"`
	AuditFields AuditFields `json:"audit"`
}

type Project struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Summary         string       `json:"summary"`
	Objective       string       `json:"objective"`
	RoadmapItemID   string       `json:"roadmapItemId"`
	Owner           string       `json:"owner"`
	Participants    []string     `json:"participants"`
	ProjectType     string       `json:"projectType"`
	Status          string       `json:"status"`
	HealthStatus    HealthStatus `json:"healthStatus"`
	TargetStartDate *time.Time   `json:"targetStartDate,omitempty"`
	TargetEndDate   *time.Time   `json:"targetEndDate,omitempty"`
	ActualEndDate   *time.Time   `json:"actualEndDate,omitempty"`
	Priority        string       `json:"priority"`
	Tags            []string     `json:"tags"`
	AuditFields     AuditFields  `json:"audit"`
}

type Milestone struct {
	ID                 string          `json:"id"`
	ProjectID          string          `json:"projectId"`
	Title              string          `json:"title"`
	MilestoneType      string          `json:"milestoneType"`
	Description        string          `json:"description"`
	CompletionCriteria string          `json:"completionCriteria"`
	Owner              string          `json:"owner"`
	PlannedDate        *time.Time      `json:"plannedDate,omitempty"`
	ForecastDate       *time.Time      `json:"forecastDate,omitempty"`
	CompletedDate      *time.Time      `json:"completedDate,omitempty"`
	Status             MilestoneStatus `json:"status"`
	HealthStatus       HealthStatus    `json:"healthStatus"`
	ProgressPercent    int             `json:"progressPercent"`
	RiskLevel          string          `json:"riskLevel"`
	DependencySummary  string          `json:"dependencySummary"`
	AuditFields        AuditFields     `json:"audit"`
}

type Workstream struct {
	ID          string      `json:"id"`
	ProjectID   string      `json:"projectId"`
	MilestoneID string      `json:"milestoneId"`
	Name        string      `json:"name"`
	Owner       string      `json:"owner"`
	Status      string      `json:"status"`
	Description string      `json:"description"`
	AuditFields AuditFields `json:"audit"`
}

type LinkedWorkItem struct {
	ID               string             `json:"id"`
	SourceType       WorkItemSourceType `json:"sourceType"`
	SourceID         string             `json:"sourceId"`
	SourceURL        string             `json:"sourceUrl"`
	Title            string             `json:"title"`
	ProjectID        string             `json:"projectId"`
	MilestoneID      string             `json:"milestoneId"`
	WorkstreamID     string             `json:"workstreamId"`
	Owner            string             `json:"owner"`
	Status           string             `json:"status"`
	Priority         string             `json:"priority,omitempty"`
	Tags             []string           `json:"tags,omitempty"`
	Estimate         string             `json:"estimate"`
	PlannedStartDate *time.Time         `json:"plannedStartDate,omitempty"`
	PlannedEndDate   *time.Time         `json:"plannedEndDate,omitempty"`
	DueDate          *time.Time         `json:"dueDate,omitempty"`
	Blocked          bool               `json:"blocked"`
	GitLabLabels     []string           `json:"gitlabLabels,omitempty"`
	GitLabAssignee   string             `json:"gitlabAssignee,omitempty"`
	GitLabState      string             `json:"gitlabState,omitempty"`
	LastSyncedAt     *time.Time         `json:"lastSyncedAt,omitempty"`
	AuditFields      AuditFields        `json:"audit"`
}

type WeeklyUpdate struct {
	ID              string      `json:"id"`
	ProjectID       string      `json:"projectId"`
	MilestoneID     string      `json:"milestoneId"`
	Author          string      `json:"author"`
	Week            string      `json:"week"`
	Summary         string      `json:"summary"`
	Progress        string      `json:"progress"`
	Risk            string      `json:"risk"`
	Blockers        string      `json:"blockers"`
	DecisionsNeeded string      `json:"decisionsNeeded"`
	NextSteps       string      `json:"nextSteps"`
	AuditFields     AuditFields `json:"audit"`
}

type GitLabConfig struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	BaseURL     string      `json:"baseUrl"`
	AccessToken string      `json:"accessToken,omitempty"`
	Group       string      `json:"group"`
	Repository  string      `json:"repository"`
	AuditFields AuditFields `json:"audit"`
}

type SyncRule struct {
	ID              string      `json:"id"`
	GitLabConfigID  string      `json:"gitlabConfigId"`
	ProjectID       string      `json:"projectId"`
	MilestoneID     string      `json:"milestoneId,omitempty"`
	Label           string      `json:"label,omitempty"`
	Assignee        string      `json:"assignee,omitempty"`
	GitLabMilestone string      `json:"gitlabMilestone,omitempty"`
	QueryFilter     string      `json:"queryFilter,omitempty"`
	Enabled         bool        `json:"enabled"`
	AuditFields     AuditFields `json:"audit"`
}

type SyncJob struct {
	ID           string      `json:"id"`
	RuleID       string      `json:"ruleId"`
	Status       string      `json:"status"`
	StartedAt    *time.Time  `json:"startedAt,omitempty"`
	CompletedAt  *time.Time  `json:"completedAt,omitempty"`
	ItemsSynced  int         `json:"itemsSynced"`
	ItemsFailed  int         `json:"itemsFailed"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
	AuditFields  AuditFields `json:"audit"`
}

type SyncFailure struct {
	ID          string      `json:"id"`
	WorkItemID  string      `json:"workItemId"`
	SourceID    string      `json:"sourceId"`
	Error       string      `json:"error"`
	RetryCount  int         `json:"retryCount"`
	LastAttempt time.Time   `json:"lastAttempt"`
	Resolved    bool        `json:"resolved"`
	AuditFields AuditFields `json:"audit"`
}

type NotificationEvent struct {
	ID          string      `json:"id"`
	EventType   string      `json:"eventType"`
	Target      string      `json:"target"`
	Channel     string      `json:"channel"`
	Title       string      `json:"title"`
	Message     string      `json:"message"`
	Delivered   bool        `json:"delivered"`
	AuditFields AuditFields `json:"audit"`
}

type Alert struct {
	ID          string      `json:"id"`
	AlertType   string      `json:"alertType"`
	TargetID    string      `json:"targetId"`
	TargetType  string      `json:"targetType"`
	Message     string      `json:"message"`
	Dismissed   bool        `json:"dismissed"`
	AuditFields AuditFields `json:"audit"`
}

type PortfolioSummary struct {
	ActiveProjects     int            `json:"activeProjects"`
	HealthDistribution map[string]int `json:"healthDistribution"`
	BlockedMilestones  int            `json:"blockedMilestones"`
	OverdueMilestones  int            `json:"overdueMilestones"`
	MilestoneWorkItems int            `json:"milestoneWorkItems"`
	BAUWorkItems       int            `json:"bauWorkItems"`
}

type RoadmapOverviewItem struct {
	Period           RoadmapPeriod    `json:"period"`
	Items            []RoadmapItem    `json:"items"`
	ProjectSummaries []ProjectSummary `json:"projectSummaries"`
}

type ProjectSummary struct {
	Project         Project     `json:"project"`
	Milestones      []Milestone `json:"milestones"`
	HealthStatus    string      `json:"healthStatus"`
	ProgressPercent int         `json:"progressPercent"`
}

type ProjectDetailView struct {
	Project    Project          `json:"project"`
	Milestones []Milestone      `json:"milestones"`
	WorkItems  []LinkedWorkItem `json:"workItems"`
	Updates    []WeeklyUpdate   `json:"updates"`
}

type MilestoneDetailView struct {
	Milestone Milestone        `json:"milestone"`
	WorkItems []LinkedWorkItem `json:"workItems"`
	Updates   []WeeklyUpdate   `json:"updates"`
}

type WeeklyReviewView struct {
	Updates           []WeeklyUpdate `json:"updates"`
	DelayedMilestones []Milestone    `json:"delayedMilestones"`
	BlockedMilestones []Milestone    `json:"blockedMilestones"`
}

type OperationalStatus struct {
	SyncStatus         SyncStatusSummary   `json:"syncStatus"`
	ProjectionStatus   ProjectionSummary   `json:"projectionStatus"`
	NotificationStatus NotificationSummary `json:"notificationStatus"`
	AlertSummary       AlertSummary        `json:"alertSummary"`
}

type SyncStatusSummary struct {
	TotalJobs          int        `json:"totalJobs"`
	LastRunAt          *time.Time `json:"lastRunAt,omitempty"`
	FailedJobs         int        `json:"failedJobs"`
	UnresolvedFailures int        `json:"unresolvedFailures"`
	ActiveRules        int        `json:"activeRules"`
}

type ProjectionSummary struct {
	PortfolioLastUpdated *time.Time `json:"portfolioLastUpdated,omitempty"`
	RoadmapCount         int        `json:"roadmapCount"`
	ProjectCount         int        `json:"projectCount"`
	MilestoneCount       int        `json:"milestoneCount"`
	WorkItemCount        int        `json:"workItemCount"`
}

type NotificationSummary struct {
	TotalSent      int `json:"totalSent"`
	DeliveryFailed int `json:"deliveryFailed"`
	PendingAlerts  int `json:"pendingAlerts"`
}

type AlertSummary struct {
	Total             int `json:"total"`
	Undismissed       int `json:"undismissed"`
	BlockedMilestones int `json:"blockedMilestones"`
	OverdueMilestones int `json:"overdueMilestones"`
	MissingUpdates    int `json:"missingUpdates"`
}
