package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"goal-manager/backend/internal/domain"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrInvalid   = errors.New("invalid request")
	ErrNotFound  = errors.New("not found")
)

type Store struct {
	mu             sync.RWMutex
	roadmapPeriods map[string]domain.RoadmapPeriod
	roadmapItems   map[string]domain.RoadmapItem
	projects       map[string]domain.Project
	milestones     map[string]domain.Milestone
	workstreams    map[string]domain.Workstream
	workItems      map[string]domain.LinkedWorkItem
	updates        map[string]domain.WeeklyUpdate
	gitlabConfigs  map[string]domain.GitLabConfig
	syncRules      map[string]domain.SyncRule
	syncJobs       map[string]domain.SyncJob
	syncFailures   map[string]domain.SyncFailure
	notifications  map[string]domain.NotificationEvent
	alerts         map[string]domain.Alert
	sequence       int
	repo           Repository
}

func NewStore() *Store {
	store, err := NewStoreWithRepository(NewMemoryRepository())
	if err != nil {
		panic(err)
	}
	return store
}

func NewStoreWithRepository(repo Repository) (*Store, error) {
	if repo == nil {
		repo = NewMemoryRepository()
	}
	state, err := repo.Load(context.Background())
	if err != nil {
		return nil, err
	}
	state = normalizeState(state)
	return &Store{
		roadmapPeriods: state.RoadmapPeriods,
		roadmapItems:   state.RoadmapItems,
		projects:       state.Projects,
		milestones:     state.Milestones,
		workstreams:    state.Workstreams,
		workItems:      state.WorkItems,
		updates:        state.Updates,
		gitlabConfigs:  state.GitLabConfigs,
		syncRules:      state.SyncRules,
		syncJobs:       state.SyncJobs,
		syncFailures:   state.SyncFailures,
		notifications:  state.Notifications,
		alerts:         state.Alerts,
		sequence:       state.Sequence,
		repo:           repo,
	}, nil
}

func (s *Store) StorageBackend() string {
	if s.repo == nil {
		return "memory"
	}
	return s.repo.Name()
}

func (s *Store) Durable() bool {
	return s.repo != nil && s.repo.Durable()
}

func (s *Store) stateLocked() State {
	return State{
		RoadmapPeriods: s.roadmapPeriods,
		RoadmapItems:   s.roadmapItems,
		Projects:       s.projects,
		Milestones:     s.milestones,
		Workstreams:    s.workstreams,
		WorkItems:      s.workItems,
		Updates:        s.updates,
		GitLabConfigs:  s.gitlabConfigs,
		SyncRules:      s.syncRules,
		SyncJobs:       s.syncJobs,
		SyncFailures:   s.syncFailures,
		Notifications:  s.notifications,
		Alerts:         s.alerts,
		Sequence:       s.sequence,
	}
}

func (s *Store) persistLocked() error {
	if s.repo == nil {
		return nil
	}
	return s.repo.Save(context.Background(), s.stateLocked())
}

func (s *Store) nextID(prefix string) string {
	s.sequence++
	return fmt.Sprintf("%s-%03d", prefix, s.sequence)
}

func sameUser(left, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func hasParticipant(project domain.Project, user string) bool {
	for _, participant := range project.Participants {
		if sameUser(participant, user) {
			return true
		}
	}
	return false
}

func (s *Store) canManageProjectLocked(auth AuthContext, project domain.Project) bool {
	if !HasPermission(auth.Role, PermManageProject) {
		return false
	}
	if !auth.HasUser() || auth.Role == domain.RoleAdmin || auth.Role == domain.RolePortfolioManager {
		return true
	}
	return auth.Role == domain.RoleProjectOwner && sameUser(project.Owner, auth.User)
}

func (s *Store) canManageMilestoneProjectLocked(auth AuthContext, projectID string) bool {
	if !HasPermission(auth.Role, PermManageMilestone) {
		return false
	}
	if !auth.HasUser() {
		return true
	}
	if auth.Role == domain.RoleAdmin {
		return true
	}
	project, ok := s.projects[projectID]
	if !ok {
		return false
	}
	return auth.Role == domain.RoleProjectOwner && sameUser(project.Owner, auth.User)
}

func (s *Store) canManageWorkItemLocked(auth AuthContext, item domain.LinkedWorkItem) bool {
	if !HasPermission(auth.Role, PermManageWorkItem) {
		return false
	}
	if !auth.HasUser() || auth.Role == domain.RoleAdmin {
		return true
	}
	project, ok := s.projects[item.ProjectID]
	if !ok {
		return item.ProjectID == "" && item.SourceType == domain.SourceBAUTask && sameUser(item.Owner, auth.User)
	}
	switch auth.Role {
	case domain.RoleProjectOwner:
		return sameUser(project.Owner, auth.User)
	case domain.RoleContributor:
		return sameUser(item.Owner, auth.User) && hasParticipant(project, auth.User)
	default:
		return false
	}
}

func (s *Store) canSubmitUpdateLocked(auth AuthContext, item domain.WeeklyUpdate) bool {
	if !HasPermission(auth.Role, PermSubmitUpdate) {
		return false
	}
	if !auth.HasUser() || auth.Role == domain.RoleAdmin {
		return true
	}
	project, ok := s.projects[item.ProjectID]
	if !ok {
		return false
	}
	switch auth.Role {
	case domain.RoleProjectOwner:
		return sameUser(project.Owner, auth.User)
	case domain.RoleContributor:
		return sameUser(item.Author, auth.User) && hasParticipant(project, auth.User)
	default:
		return false
	}
}

func (s *Store) CreateRoadmapPeriod(role domain.WorkspaceRole, period domain.RoadmapPeriod) (domain.RoadmapPeriod, error) {
	if !HasPermission(role, PermManageRoadmap) {
		return domain.RoadmapPeriod{}, ErrForbidden
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	period.ID = s.nextID("rp")
	period.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.roadmapPeriods[period.ID] = period
	if err := s.persistLocked(); err != nil {
		return domain.RoadmapPeriod{}, err
	}
	return period, nil
}

func (s *Store) UpsertRoadmapPeriod(role domain.WorkspaceRole, id string, period domain.RoadmapPeriod) (domain.RoadmapPeriod, error) {
	if !HasPermission(role, PermManageRoadmap) {
		return domain.RoadmapPeriod{}, ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.roadmapPeriods[id]
	if !ok {
		return domain.RoadmapPeriod{}, ErrNotFound
	}
	period.ID = id
	period.AuditFields = current.AuditFields
	period.AuditFields.UpdatedAt = time.Now().UTC()
	s.roadmapPeriods[id] = period
	if err := s.persistLocked(); err != nil {
		return domain.RoadmapPeriod{}, err
	}
	return period, nil
}

func (s *Store) ArchiveRoadmapPeriod(role domain.WorkspaceRole, id string) (domain.RoadmapPeriod, error) {
	if !HasPermission(role, PermManageRoadmap) {
		return domain.RoadmapPeriod{}, ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	period, ok := s.roadmapPeriods[id]
	if !ok {
		return domain.RoadmapPeriod{}, ErrNotFound
	}
	period.Status = "archived"
	period.AuditFields.UpdatedAt = time.Now().UTC()
	s.roadmapPeriods[id] = period
	if err := s.persistLocked(); err != nil {
		return domain.RoadmapPeriod{}, err
	}
	return period, nil
}

func (s *Store) ListRoadmapPeriods() []domain.RoadmapPeriod {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.RoadmapPeriod, 0, len(s.roadmapPeriods))
	for _, item := range s.roadmapPeriods {
		items = append(items, item)
	}
	return items
}

func (s *Store) GetRoadmapPeriod(id string) (domain.RoadmapPeriod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.roadmapPeriods[id]
	if !ok {
		return domain.RoadmapPeriod{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) ListRoadmapItems() []domain.RoadmapItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.RoadmapItem, 0, len(s.roadmapItems))
	for _, item := range s.roadmapItems {
		items = append(items, item)
	}
	return items
}

func (s *Store) GetRoadmapItem(id string) (domain.RoadmapItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.roadmapItems[id]
	if !ok {
		return domain.RoadmapItem{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) CreateRoadmapItem(role domain.WorkspaceRole, item domain.RoadmapItem) (domain.RoadmapItem, error) {
	if !HasPermission(role, PermManageRoadmap) {
		return domain.RoadmapItem{}, ErrForbidden
	}
	if item.PeriodID == "" {
		return domain.RoadmapItem{}, ErrInvalid
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = s.nextID("ri")
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.roadmapItems[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.RoadmapItem{}, err
	}
	return item, nil
}

func (s *Store) UpsertRoadmapItem(role domain.WorkspaceRole, id string, item domain.RoadmapItem) (domain.RoadmapItem, error) {
	if !HasPermission(role, PermManageRoadmap) {
		return domain.RoadmapItem{}, ErrForbidden
	}
	if item.PeriodID == "" {
		return domain.RoadmapItem{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.roadmapItems[id]
	if !ok {
		return domain.RoadmapItem{}, ErrNotFound
	}
	item.ID = id
	item.AuditFields = current.AuditFields
	item.AuditFields.UpdatedAt = time.Now().UTC()
	s.roadmapItems[id] = item
	if err := s.persistLocked(); err != nil {
		return domain.RoadmapItem{}, err
	}
	return item, nil
}

func (s *Store) CreateProject(role domain.WorkspaceRole, item domain.Project) (domain.Project, error) {
	return s.CreateProjectForActor(RoleAuth(role), item)
}

func (s *Store) CreateProjectForActor(auth AuthContext, item domain.Project) (domain.Project, error) {
	if !HasPermission(auth.Role, PermManageProject) {
		return domain.Project{}, ErrForbidden
	}
	if auth.HasUser() && auth.Role == domain.RoleProjectOwner && !sameUser(item.Owner, auth.User) {
		return domain.Project{}, ErrForbidden
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = s.nextID("prj")
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.projects[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.Project{}, err
	}
	return item, nil
}

func (s *Store) ListProjects() []domain.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Project, 0, len(s.projects))
	for _, item := range s.projects {
		items = append(items, item)
	}
	return items
}

func (s *Store) GetProject(id string) (domain.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.projects[id]
	if !ok {
		return domain.Project{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) UpsertProject(role domain.WorkspaceRole, id string, item domain.Project) (domain.Project, error) {
	return s.UpsertProjectForActor(RoleAuth(role), id, item)
}

func (s *Store) UpsertProjectForActor(auth AuthContext, id string, item domain.Project) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.projects[id]
	if !ok {
		return domain.Project{}, ErrNotFound
	}
	if !s.canManageProjectLocked(auth, current) {
		return domain.Project{}, ErrForbidden
	}
	if auth.HasUser() && auth.Role == domain.RoleProjectOwner && !sameUser(item.Owner, auth.User) {
		return domain.Project{}, ErrForbidden
	}
	item.ID = id
	item.AuditFields = current.AuditFields
	item.AuditFields.UpdatedAt = time.Now().UTC()
	s.projects[id] = item
	if err := s.persistLocked(); err != nil {
		return domain.Project{}, err
	}
	return item, nil
}

func (s *Store) CreateMilestone(role domain.WorkspaceRole, item domain.Milestone) (domain.Milestone, error) {
	return s.CreateMilestoneForActor(RoleAuth(role), item)
}

func (s *Store) CreateMilestoneForActor(auth AuthContext, item domain.Milestone) (domain.Milestone, error) {
	if err := validateMilestone(item); err != nil {
		return domain.Milestone{}, err
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.canManageMilestoneProjectLocked(auth, item.ProjectID) {
		return domain.Milestone{}, ErrForbidden
	}
	if auth.HasUser() && auth.Role == domain.RoleContributor && item.Status == domain.MilestoneCompleted {
		return domain.Milestone{}, ErrForbidden
	}
	item.ID = s.nextID("ms")
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.milestones[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.Milestone{}, err
	}
	return item, nil
}

func (s *Store) ListMilestones() []domain.Milestone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Milestone, 0, len(s.milestones))
	for _, item := range s.milestones {
		items = append(items, item)
	}
	return items
}

func (s *Store) GetMilestone(id string) (domain.Milestone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.milestones[id]
	if !ok {
		return domain.Milestone{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) UpsertMilestone(role domain.WorkspaceRole, id string, item domain.Milestone) (domain.Milestone, error) {
	return s.UpsertMilestoneForActor(RoleAuth(role), id, item)
}

func (s *Store) UpsertMilestoneForActor(auth AuthContext, id string, item domain.Milestone) (domain.Milestone, error) {
	if err := validateMilestone(item); err != nil {
		return domain.Milestone{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.milestones[id]
	if !ok {
		return domain.Milestone{}, ErrNotFound
	}
	if !s.canManageMilestoneProjectLocked(auth, current.ProjectID) || !s.canManageMilestoneProjectLocked(auth, item.ProjectID) {
		return domain.Milestone{}, ErrForbidden
	}
	if auth.HasUser() && auth.Role == domain.RoleContributor && item.Status == domain.MilestoneCompleted {
		return domain.Milestone{}, ErrForbidden
	}
	if err := validateMilestoneTransition(current, item); err != nil {
		return domain.Milestone{}, err
	}
	item.ID = id
	item.AuditFields = current.AuditFields
	item.AuditFields.UpdatedAt = time.Now().UTC()
	if item.Status == domain.MilestoneCompleted && item.CompletedDate == nil {
		now := time.Now().UTC()
		item.CompletedDate = &now
	}
	s.milestones[id] = item
	if err := s.persistLocked(); err != nil {
		return domain.Milestone{}, err
	}
	return item, nil
}

func (s *Store) CreateWorkstream(role domain.WorkspaceRole, item domain.Workstream) (domain.Workstream, error) {
	if !HasPermission(role, PermManageWorkstream) {
		return domain.Workstream{}, ErrForbidden
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = s.nextID("ws")
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.workstreams[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.Workstream{}, err
	}
	return item, nil
}

func (s *Store) ListWorkstreams() []domain.Workstream {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Workstream, 0, len(s.workstreams))
	for _, item := range s.workstreams {
		items = append(items, item)
	}
	return items
}

func (s *Store) GetWorkstream(id string) (domain.Workstream, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.workstreams[id]
	if !ok {
		return domain.Workstream{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) UpsertWorkstream(role domain.WorkspaceRole, id string, item domain.Workstream) (domain.Workstream, error) {
	if !HasPermission(role, PermManageWorkstream) {
		return domain.Workstream{}, ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.workstreams[id]
	if !ok {
		return domain.Workstream{}, ErrNotFound
	}
	item.ID = id
	item.AuditFields = current.AuditFields
	item.AuditFields.UpdatedAt = time.Now().UTC()
	s.workstreams[id] = item
	if err := s.persistLocked(); err != nil {
		return domain.Workstream{}, err
	}
	return item, nil
}

func (s *Store) CreateLinkedWorkItem(role domain.WorkspaceRole, item domain.LinkedWorkItem) (domain.LinkedWorkItem, error) {
	return s.CreateLinkedWorkItemForActor(RoleAuth(role), item)
}

func (s *Store) CreateLinkedWorkItemForActor(auth AuthContext, item domain.LinkedWorkItem) (domain.LinkedWorkItem, error) {
	if err := validateWorkItem(item); err != nil {
		return domain.LinkedWorkItem{}, err
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.canManageWorkItemLocked(auth, item) {
		return domain.LinkedWorkItem{}, ErrForbidden
	}
	item.ID = s.nextID("work")
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.workItems[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.LinkedWorkItem{}, err
	}
	return item, nil
}

func (s *Store) ListWorkItems() []domain.LinkedWorkItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.LinkedWorkItem, 0, len(s.workItems))
	for _, item := range s.workItems {
		items = append(items, item)
	}
	return items
}

func (s *Store) GetWorkItem(id string) (domain.LinkedWorkItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.workItems[id]
	if !ok {
		return domain.LinkedWorkItem{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) UpsertWorkItem(role domain.WorkspaceRole, id string, item domain.LinkedWorkItem) (domain.LinkedWorkItem, error) {
	return s.UpsertWorkItemForActor(RoleAuth(role), id, item)
}

func (s *Store) UpsertWorkItemForActor(auth AuthContext, id string, item domain.LinkedWorkItem) (domain.LinkedWorkItem, error) {
	if err := validateWorkItem(item); err != nil {
		return domain.LinkedWorkItem{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.workItems[id]
	if !ok {
		return domain.LinkedWorkItem{}, ErrNotFound
	}
	if !s.canManageWorkItemLocked(auth, current) || !s.canManageWorkItemLocked(auth, item) {
		return domain.LinkedWorkItem{}, ErrForbidden
	}
	item.ID = id
	item.AuditFields = current.AuditFields
	item.AuditFields.UpdatedAt = time.Now().UTC()
	s.workItems[id] = item
	if err := s.persistLocked(); err != nil {
		return domain.LinkedWorkItem{}, err
	}
	return item, nil
}

func (s *Store) DeleteWorkItem(role domain.WorkspaceRole, id string) error {
	return s.DeleteWorkItemForActor(RoleAuth(role), id)
}

func (s *Store) DeleteWorkItemForActor(auth AuthContext, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.workItems[id]
	if !ok {
		return ErrNotFound
	}
	if !s.canManageWorkItemLocked(auth, item) {
		return ErrForbidden
	}
	delete(s.workItems, id)
	return s.persistLocked()
}

func (s *Store) LinkGitLabIssue(role domain.WorkspaceRole, item domain.LinkedWorkItem) (domain.LinkedWorkItem, error) {
	return s.LinkGitLabIssueForActor(RoleAuth(role), item)
}

func (s *Store) LinkGitLabIssueForActor(auth AuthContext, item domain.LinkedWorkItem) (domain.LinkedWorkItem, error) {
	if item.SourceType != domain.SourceGitLabIssue {
		return domain.LinkedWorkItem{}, fmt.Errorf("%w: source type must be gitlab_issue", ErrInvalid)
	}
	if item.SourceID == "" || item.SourceURL == "" {
		return domain.LinkedWorkItem{}, fmt.Errorf("%w: sourceId and sourceUrl are required for GitLab issues", ErrInvalid)
	}
	if err := validateWorkItem(item); err != nil {
		return domain.LinkedWorkItem{}, err
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.canManageWorkItemLocked(auth, item) {
		return domain.LinkedWorkItem{}, ErrForbidden
	}
	item.ID = s.nextID("work")
	item.LastSyncedAt = &now
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.workItems[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.LinkedWorkItem{}, err
	}
	return item, nil
}

func (s *Store) UnlinkGitLabIssue(role domain.WorkspaceRole, id string) error {
	return s.UnlinkGitLabIssueForActor(RoleAuth(role), id)
}

func (s *Store) UnlinkGitLabIssueForActor(auth AuthContext, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.workItems[id]
	if !ok {
		return ErrNotFound
	}
	if !s.canManageWorkItemLocked(auth, item) {
		return ErrForbidden
	}
	if item.SourceType != domain.SourceGitLabIssue {
		return fmt.Errorf("%w: item is not a GitLab-linked work item", ErrInvalid)
	}
	delete(s.workItems, id)
	return s.persistLocked()
}

func (s *Store) CreateWeeklyUpdate(role domain.WorkspaceRole, item domain.WeeklyUpdate) (domain.WeeklyUpdate, error) {
	return s.CreateWeeklyUpdateForActor(RoleAuth(role), item)
}

func (s *Store) CreateWeeklyUpdateForActor(auth AuthContext, item domain.WeeklyUpdate) (domain.WeeklyUpdate, error) {
	if !HasPermission(auth.Role, PermSubmitUpdate) {
		return domain.WeeklyUpdate{}, ErrForbidden
	}
	if item.ProjectID == "" {
		return domain.WeeklyUpdate{}, fmt.Errorf("%w: projectId is required", ErrInvalid)
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.canSubmitUpdateLocked(auth, item) {
		return domain.WeeklyUpdate{}, ErrForbidden
	}
	item.ID = s.nextID("wu")
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.updates[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.WeeklyUpdate{}, err
	}
	return item, nil
}

func (s *Store) ListWeeklyUpdates() []domain.WeeklyUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.WeeklyUpdate, 0, len(s.updates))
	for _, item := range s.updates {
		items = append(items, item)
	}
	return items
}

func (s *Store) GetWeeklyUpdate(id string) (domain.WeeklyUpdate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.updates[id]
	if !ok {
		return domain.WeeklyUpdate{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) UpsertWeeklyUpdate(role domain.WorkspaceRole, id string, item domain.WeeklyUpdate) (domain.WeeklyUpdate, error) {
	return s.UpsertWeeklyUpdateForActor(RoleAuth(role), id, item)
}

func (s *Store) UpsertWeeklyUpdateForActor(auth AuthContext, id string, item domain.WeeklyUpdate) (domain.WeeklyUpdate, error) {
	if !HasPermission(auth.Role, PermSubmitUpdate) {
		return domain.WeeklyUpdate{}, ErrForbidden
	}
	if item.ProjectID == "" {
		return domain.WeeklyUpdate{}, fmt.Errorf("%w: projectId is required", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.updates[id]
	if !ok {
		return domain.WeeklyUpdate{}, ErrNotFound
	}
	if !s.canSubmitUpdateLocked(auth, current) || !s.canSubmitUpdateLocked(auth, item) {
		return domain.WeeklyUpdate{}, ErrForbidden
	}
	item.ID = id
	item.AuditFields = current.AuditFields
	item.AuditFields.UpdatedAt = time.Now().UTC()
	s.updates[id] = item
	if err := s.persistLocked(); err != nil {
		return domain.WeeklyUpdate{}, err
	}
	return item, nil
}

func (s *Store) PortfolioSummary() domain.PortfolioSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary := domain.PortfolioSummary{
		HealthDistribution: map[string]int{},
	}
	now := time.Now().UTC()
	for _, project := range s.projects {
		if project.Status != "done" {
			summary.ActiveProjects++
		}
		summary.HealthDistribution[string(project.HealthStatus)]++
	}
	for _, milestone := range s.milestones {
		if milestone.Status == domain.MilestoneBlocked {
			summary.BlockedMilestones++
		}
		if milestone.PlannedDate != nil && milestone.CompletedDate == nil && milestone.PlannedDate.Before(now) {
			summary.OverdueMilestones++
		}
	}
	for _, item := range s.workItems {
		if item.SourceType == domain.SourceBAUTask {
			summary.BAUWorkItems++
		} else {
			summary.MilestoneWorkItems++
		}
	}
	return summary
}

func (s *Store) RoadmapOverview() []domain.RoadmapOverviewItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	periodItems := map[string][]domain.RoadmapItem{}
	for _, item := range s.roadmapItems {
		periodItems[item.PeriodID] = append(periodItems[item.PeriodID], item)
	}

	roadmapProjects := map[string][]domain.Project{}
	for _, project := range s.projects {
		if project.RoadmapItemID != "" {
			roadmapProjects[project.RoadmapItemID] = append(roadmapProjects[project.RoadmapItemID], project)
		}
	}

	overviews := make([]domain.RoadmapOverviewItem, 0, len(s.roadmapPeriods))
	for _, period := range s.roadmapPeriods {
		items := periodItems[period.ID]
		if items == nil {
			items = []domain.RoadmapItem{}
		}
		projectSummaries := make([]domain.ProjectSummary, 0)
		for _, ri := range items {
			for _, proj := range roadmapProjects[ri.ID] {
				ms := s.milestonesForProject(proj.ID)
				projectSummaries = append(projectSummaries, domain.ProjectSummary{
					Project:      proj,
					Milestones:   ms,
					HealthStatus: string(proj.HealthStatus),
				})
			}
		}
		overviews = append(overviews, domain.RoadmapOverviewItem{
			Period:           period,
			Items:            items,
			ProjectSummaries: projectSummaries,
		})
	}
	return overviews
}

func (s *Store) ProjectDetail(id string) (domain.ProjectDetailView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, ok := s.projects[id]
	if !ok {
		return domain.ProjectDetailView{}, ErrNotFound
	}
	return domain.ProjectDetailView{
		Project:    project,
		Milestones: s.milestonesForProject(id),
		WorkItems:  s.workItemsForProject(id),
		Updates:    s.updatesForProject(id),
	}, nil
}

func (s *Store) ProjectSpace(id string) (domain.ProjectSpaceView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, ok := s.projects[id]
	if !ok {
		return domain.ProjectSpaceView{}, ErrNotFound
	}
	milestones := s.milestonesForProject(id)
	workItems := s.workItemsForProject(id)
	updates := s.updatesForProject(id)
	sort.Slice(updates, func(i, j int) bool {
		return updates[i].AuditFields.CreatedAt.After(updates[j].AuditFields.CreatedAt)
	})
	return domain.ProjectSpaceView{
		Project:      project,
		Milestones:   milestones,
		WorkItems:    workItems,
		Updates:      updates,
		Rollups:      projectSpaceRollups(milestones, workItems, updates),
		Risks:        projectRiskSignals(milestones, workItems, updates),
		Dependencies: projectDependencies(milestones, workItems),
	}, nil
}

func (s *Store) MilestoneDetail(id string) (domain.MilestoneDetailView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ms, ok := s.milestones[id]
	if !ok {
		return domain.MilestoneDetailView{}, ErrNotFound
	}
	return domain.MilestoneDetailView{
		Milestone: ms,
		WorkItems: s.workItemsForMilestone(id),
		Updates:   s.updatesForMilestone(id),
	}, nil
}

func projectSpaceRollups(milestones []domain.Milestone, workItems []domain.LinkedWorkItem, updates []domain.WeeklyUpdate) domain.ProjectSpaceRollups {
	now := time.Now().UTC()
	rollups := domain.ProjectSpaceRollups{
		MilestoneStatusCounts: map[string]int{},
		WorkItemStatusCounts:  map[string]int{},
		RecentUpdateCount:     len(updates),
	}
	for _, milestone := range milestones {
		rollups.MilestoneStatusCounts[string(milestone.Status)]++
		switch milestone.Status {
		case domain.MilestoneActive:
			rollups.ActiveMilestones++
		case domain.MilestoneCompleted:
			rollups.CompletedMilestones++
		case domain.MilestoneBlocked:
			rollups.BlockedMilestones++
		}
		if milestone.PlannedDate != nil && milestone.CompletedDate == nil && milestone.PlannedDate.Before(now) && milestone.Status != domain.MilestoneCompleted && milestone.Status != domain.MilestoneCancelled {
			rollups.OverdueMilestones++
		}
	}
	for _, item := range workItems {
		rollups.WorkItemStatusCounts[item.Status]++
		if item.Blocked || item.Status == "blocked" {
			rollups.BlockedWorkItems++
		}
		if item.DueDate != nil && item.DueDate.Before(now) && item.Status != "done" {
			rollups.OverdueWorkItems++
		}
		if item.SourceType == domain.SourceExternalDependency {
			rollups.ExternalDependencies++
		}
	}
	return rollups
}

func projectRiskSignals(milestones []domain.Milestone, workItems []domain.LinkedWorkItem, updates []domain.WeeklyUpdate) []domain.ProjectRiskSignal {
	signals := make([]domain.ProjectRiskSignal, 0)
	for _, milestone := range milestones {
		if milestone.Status == domain.MilestoneBlocked || milestone.RiskLevel == "high" || milestone.HealthStatus == domain.HealthAtRisk || milestone.HealthStatus == domain.HealthOffTrack {
			severity := milestone.RiskLevel
			if severity == "" {
				severity = string(milestone.HealthStatus)
			}
			signals = append(signals, domain.ProjectRiskSignal{
				ID:          "milestone:" + milestone.ID,
				SourceType:  "milestone",
				SourceID:    milestone.ID,
				Title:       milestone.Title,
				Severity:    severity,
				Message:     firstText(milestone.DependencySummary, milestone.RiskLevel, string(milestone.HealthStatus)),
				Owner:       milestone.Owner,
				Status:      string(milestone.Status),
				MilestoneID: milestone.ID,
			})
		}
	}
	for _, item := range workItems {
		if item.Blocked || item.Status == "blocked" || item.SourceType == domain.SourceExternalDependency {
			severity := "medium"
			if item.SourceType == domain.SourceExternalDependency {
				severity = "dependency"
			}
			signals = append(signals, domain.ProjectRiskSignal{
				ID:          "work-item:" + item.ID,
				SourceType:  "work_item",
				SourceID:    item.ID,
				Title:       firstText(item.Title, item.ID),
				Severity:    severity,
				Message:     "Blocked or dependency work requires attention",
				Owner:       item.Owner,
				Status:      item.Status,
				MilestoneID: item.MilestoneID,
				WorkItemID:  item.ID,
			})
		}
	}
	for _, update := range updates {
		if strings.TrimSpace(update.Risk) != "" || strings.TrimSpace(update.Blockers) != "" || strings.TrimSpace(update.DecisionsNeeded) != "" {
			signals = append(signals, domain.ProjectRiskSignal{
				ID:          "update:" + update.ID,
				SourceType:  "weekly_update",
				SourceID:    update.ID,
				Title:       firstText(update.Summary, update.Week),
				Severity:    "update",
				Message:     firstText(update.Risk, update.Blockers, update.DecisionsNeeded),
				Owner:       update.Author,
				Status:      update.Week,
				MilestoneID: update.MilestoneID,
				UpdateID:    update.ID,
			})
		}
	}
	return signals
}

func projectDependencies(milestones []domain.Milestone, workItems []domain.LinkedWorkItem) []domain.ProjectDependency {
	dependencies := make([]domain.ProjectDependency, 0)
	for _, milestone := range milestones {
		if strings.TrimSpace(milestone.DependencySummary) != "" {
			dependencies = append(dependencies, domain.ProjectDependency{
				ID:          "milestone:" + milestone.ID,
				SourceType:  "milestone",
				SourceID:    milestone.ID,
				Title:       milestone.Title,
				Message:     milestone.DependencySummary,
				Owner:       milestone.Owner,
				Status:      string(milestone.Status),
				MilestoneID: milestone.ID,
			})
		}
	}
	for _, item := range workItems {
		if item.SourceType == domain.SourceExternalDependency || item.Blocked || item.Status == "blocked" {
			dependencies = append(dependencies, domain.ProjectDependency{
				ID:          "work-item:" + item.ID,
				SourceType:  string(item.SourceType),
				SourceID:    item.ID,
				Title:       firstText(item.Title, item.ID),
				Message:     firstText(item.SourceURL, "Blocked or external dependency work item"),
				Owner:       item.Owner,
				Status:      item.Status,
				MilestoneID: item.MilestoneID,
				WorkItemID:  item.ID,
			})
		}
	}
	return dependencies
}

func firstText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (s *Store) WeeklyReviewView() domain.WeeklyReviewView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()

	var delayed, blocked []domain.Milestone
	delayed = make([]domain.Milestone, 0)
	blocked = make([]domain.Milestone, 0)
	for _, ms := range s.milestones {
		if ms.Status == domain.MilestoneBlocked {
			blocked = append(blocked, ms)
		}
		if ms.PlannedDate != nil && ms.CompletedDate == nil && ms.PlannedDate.Before(now) && ms.Status != domain.MilestoneCompleted && ms.Status != domain.MilestoneCancelled {
			delayed = append(delayed, ms)
		}
	}

	allUpdates := make([]domain.WeeklyUpdate, 0, len(s.updates))
	for _, u := range s.updates {
		allUpdates = append(allUpdates, u)
	}

	return domain.WeeklyReviewView{
		Updates:           allUpdates,
		DelayedMilestones: delayed,
		BlockedMilestones: blocked,
	}
}

func (s *Store) milestonesForProject(projectID string) []domain.Milestone {
	items := make([]domain.Milestone, 0)
	for _, ms := range s.milestones {
		if ms.ProjectID == projectID {
			items = append(items, ms)
		}
	}
	return items
}

func (s *Store) workItemsForProject(projectID string) []domain.LinkedWorkItem {
	items := make([]domain.LinkedWorkItem, 0)
	for _, wi := range s.workItems {
		if wi.ProjectID == projectID {
			items = append(items, wi)
		}
	}
	return items
}

func (s *Store) workItemsForMilestone(milestoneID string) []domain.LinkedWorkItem {
	items := make([]domain.LinkedWorkItem, 0)
	for _, wi := range s.workItems {
		if wi.MilestoneID == milestoneID {
			items = append(items, wi)
		}
	}
	return items
}

func (s *Store) updatesForProject(projectID string) []domain.WeeklyUpdate {
	items := make([]domain.WeeklyUpdate, 0)
	for _, u := range s.updates {
		if u.ProjectID == projectID {
			items = append(items, u)
		}
	}
	return items
}

func (s *Store) updatesForMilestone(milestoneID string) []domain.WeeklyUpdate {
	items := make([]domain.WeeklyUpdate, 0)
	for _, u := range s.updates {
		if u.MilestoneID == milestoneID {
			items = append(items, u)
		}
	}
	return items
}

func (s *Store) CreateGitLabConfig(role domain.WorkspaceRole, item domain.GitLabConfig) (domain.GitLabConfig, error) {
	if !HasPermission(role, PermManageIntegration) {
		return domain.GitLabConfig{}, ErrForbidden
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = s.nextID("glc")
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.gitlabConfigs[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.GitLabConfig{}, err
	}
	return item, nil
}

func (s *Store) GetGitLabConfig(id string) (domain.GitLabConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.gitlabConfigs[id]
	if !ok {
		return domain.GitLabConfig{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) ListGitLabConfigs() []domain.GitLabConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.GitLabConfig, 0, len(s.gitlabConfigs))
	for _, item := range s.gitlabConfigs {
		items = append(items, item)
	}
	return items
}

func (s *Store) UpsertGitLabConfig(role domain.WorkspaceRole, id string, item domain.GitLabConfig) (domain.GitLabConfig, error) {
	if !HasPermission(role, PermManageIntegration) {
		return domain.GitLabConfig{}, ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.gitlabConfigs[id]
	if !ok {
		return domain.GitLabConfig{}, ErrNotFound
	}
	item.ID = id
	item.AuditFields = current.AuditFields
	item.AuditFields.UpdatedAt = time.Now().UTC()
	s.gitlabConfigs[id] = item
	if err := s.persistLocked(); err != nil {
		return domain.GitLabConfig{}, err
	}
	return item, nil
}

func (s *Store) DeleteGitLabConfig(role domain.WorkspaceRole, id string) error {
	if !HasPermission(role, PermManageIntegration) {
		return ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.gitlabConfigs[id]; !ok {
		return ErrNotFound
	}
	delete(s.gitlabConfigs, id)
	return s.persistLocked()
}

func (s *Store) CreateSyncRule(role domain.WorkspaceRole, item domain.SyncRule) (domain.SyncRule, error) {
	if !HasPermission(role, PermManageSyncRule) {
		return domain.SyncRule{}, ErrForbidden
	}
	if item.GitLabConfigID == "" || item.ProjectID == "" {
		return domain.SyncRule{}, fmt.Errorf("%w: gitlabConfigId and projectId are required", ErrInvalid)
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = s.nextID("sr")
	item.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.syncRules[item.ID] = item
	if err := s.persistLocked(); err != nil {
		return domain.SyncRule{}, err
	}
	return item, nil
}

func (s *Store) GetSyncRule(id string) (domain.SyncRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.syncRules[id]
	if !ok {
		return domain.SyncRule{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) ListSyncRules() []domain.SyncRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.SyncRule, 0, len(s.syncRules))
	for _, item := range s.syncRules {
		items = append(items, item)
	}
	return items
}

func (s *Store) UpsertSyncRule(role domain.WorkspaceRole, id string, item domain.SyncRule) (domain.SyncRule, error) {
	if !HasPermission(role, PermManageSyncRule) {
		return domain.SyncRule{}, ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.syncRules[id]
	if !ok {
		return domain.SyncRule{}, ErrNotFound
	}
	item.ID = id
	item.AuditFields = current.AuditFields
	item.AuditFields.UpdatedAt = time.Now().UTC()
	s.syncRules[id] = item
	if err := s.persistLocked(); err != nil {
		return domain.SyncRule{}, err
	}
	return item, nil
}

func (s *Store) DeleteSyncRule(role domain.WorkspaceRole, id string) error {
	if !HasPermission(role, PermManageSyncRule) {
		return ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.syncRules[id]; !ok {
		return ErrNotFound
	}
	delete(s.syncRules, id)
	return s.persistLocked()
}

func (s *Store) CreateSyncJob(role domain.WorkspaceRole, job domain.SyncJob) (domain.SyncJob, error) {
	if !HasPermission(role, PermRunSync) {
		return domain.SyncJob{}, ErrForbidden
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	job.ID = s.nextID("sj")
	job.Status = "running"
	job.StartedAt = &now
	job.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.syncJobs[job.ID] = job
	if err := s.persistLocked(); err != nil {
		return domain.SyncJob{}, err
	}
	return job, nil
}

func (s *Store) CompleteSyncJob(id string, itemsSynced, itemsFailed int, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.syncJobs[id]
	if !ok {
		return ErrNotFound
	}
	now := time.Now().UTC()
	job.Status = "completed"
	if errMsg != "" {
		job.Status = "failed"
	}
	job.CompletedAt = &now
	job.ItemsSynced = itemsSynced
	job.ItemsFailed = itemsFailed
	job.ErrorMessage = errMsg
	job.AuditFields.UpdatedAt = now
	s.syncJobs[id] = job
	return s.persistLocked()
}

func (s *Store) ListSyncJobs() []domain.SyncJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.SyncJob, 0, len(s.syncJobs))
	for _, item := range s.syncJobs {
		items = append(items, item)
	}
	return items
}

func (s *Store) RecordSyncFailure(failure domain.SyncFailure) domain.SyncFailure {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	failure.ID = s.nextID("sf")
	failure.LastAttempt = now
	failure.AuditFields = domain.AuditFields{CreatedAt: now, UpdatedAt: now}
	s.syncFailures[failure.ID] = failure
	_ = s.persistLocked()
	return failure
}

func (s *Store) ListSyncFailures() []domain.SyncFailure {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.SyncFailure, 0, len(s.syncFailures))
	for _, item := range s.syncFailures {
		items = append(items, item)
	}
	return items
}

func (s *Store) ResolveSyncFailure(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	failure, ok := s.syncFailures[id]
	if !ok {
		return ErrNotFound
	}
	failure.Resolved = true
	failure.AuditFields.UpdatedAt = time.Now().UTC()
	s.syncFailures[id] = failure
	return s.persistLocked()
}

func (s *Store) GenerateAlerts() []domain.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	alerts := make([]domain.Alert, 0)

	for _, ms := range s.milestones {
		if ms.Status == domain.MilestoneBlocked && ms.CompletedDate == nil {
			alerts = append(alerts, domain.Alert{
				ID:          s.nextID("alert"),
				AlertType:   "blocked_milestone",
				TargetID:    ms.ID,
				TargetType:  "milestone",
				Message:     fmt.Sprintf("Milestone %q is blocked", ms.Title),
				AuditFields: domain.AuditFields{CreatedAt: now, UpdatedAt: now},
			})
		}
		if ms.PlannedDate != nil && ms.CompletedDate == nil && ms.PlannedDate.Before(now) && ms.Status != domain.MilestoneCompleted && ms.Status != domain.MilestoneCancelled {
			alerts = append(alerts, domain.Alert{
				ID:          s.nextID("alert"),
				AlertType:   "overdue_milestone",
				TargetID:    ms.ID,
				TargetType:  "milestone",
				Message:     fmt.Sprintf("Milestone %q is overdue (planned: %s)", ms.Title, ms.PlannedDate.Format("2006-01-02")),
				AuditFields: domain.AuditFields{CreatedAt: now, UpdatedAt: now},
			})
		}
		if ms.PlannedDate != nil && ms.CompletedDate == nil {
			daysUntil := ms.PlannedDate.Sub(now).Hours() / 24
			if daysUntil > 0 && daysUntil <= 7 {
				alerts = append(alerts, domain.Alert{
					ID:          s.nextID("alert"),
					AlertType:   "upcoming_milestone",
					TargetID:    ms.ID,
					TargetType:  "milestone",
					Message:     fmt.Sprintf("Milestone %q is due in %.0f days", ms.Title, daysUntil),
					AuditFields: domain.AuditFields{CreatedAt: now, UpdatedAt: now},
				})
			}
		}
	}

	for _, proj := range s.projects {
		hasRecentUpdate := false
		weekAgo := now.Add(-7 * 24 * time.Hour)
		for _, u := range s.updates {
			if u.ProjectID == proj.ID && u.AuditFields.CreatedAt.After(weekAgo) {
				hasRecentUpdate = true
				break
			}
		}
		if !hasRecentUpdate && proj.Status == "active" {
			alerts = append(alerts, domain.Alert{
				ID:          s.nextID("alert"),
				AlertType:   "missing_weekly_update",
				TargetID:    proj.ID,
				TargetType:  "project",
				Message:     fmt.Sprintf("Project %q has no weekly update in the last 7 days", proj.Name),
				AuditFields: domain.AuditFields{CreatedAt: now, UpdatedAt: now},
			})
		}
	}

	staleThreshold := now.Add(-3 * 24 * time.Hour)
	for _, wi := range s.workItems {
		if wi.SourceType == domain.SourceGitLabIssue && wi.LastSyncedAt != nil && wi.LastSyncedAt.Before(staleThreshold) {
			alerts = append(alerts, domain.Alert{
				ID:          s.nextID("alert"),
				AlertType:   "stale_gitlab_work",
				TargetID:    wi.ID,
				TargetType:  "work_item",
				Message:     fmt.Sprintf("GitLab-linked work item %q has not been synced in 3+ days", wi.Title),
				AuditFields: domain.AuditFields{CreatedAt: now, UpdatedAt: now},
			})
		}
	}

	for _, alert := range alerts {
		s.alerts[alert.ID] = alert
	}
	_ = s.persistLocked()
	return alerts
}

func (s *Store) ListAlerts() []domain.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Alert, 0, len(s.alerts))
	for _, item := range s.alerts {
		items = append(items, item)
	}
	return items
}

func (s *Store) DismissAlert(role domain.WorkspaceRole, id string) error {
	if !HasPermission(role, PermManageAlert) {
		return ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	alert, ok := s.alerts[id]
	if !ok {
		return ErrNotFound
	}
	alert.Dismissed = true
	alert.AuditFields.UpdatedAt = time.Now().UTC()
	s.alerts[id] = alert
	return s.persistLocked()
}

func (s *Store) SaveNotification(event domain.NotificationEvent) domain.NotificationEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	event.ID = s.nextID("notif")
	s.notifications[event.ID] = event
	_ = s.persistLocked()
	return event
}

func (s *Store) ListNotifications() []domain.NotificationEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.NotificationEvent, 0, len(s.notifications))
	for _, item := range s.notifications {
		items = append(items, item)
	}
	return items
}

func (s *Store) OperationalStatus() domain.OperationalStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sync := domain.SyncStatusSummary{}
	var lastRun *time.Time
	for _, job := range s.syncJobs {
		sync.TotalJobs++
		if job.Status == "failed" || job.Status == "completed_with_errors" {
			sync.FailedJobs++
		}
		if job.CompletedAt != nil && (lastRun == nil || job.CompletedAt.After(*lastRun)) {
			lastRun = job.CompletedAt
		}
	}
	sync.LastRunAt = lastRun
	for _, rule := range s.syncRules {
		if rule.Enabled {
			sync.ActiveRules++
		}
	}
	for _, f := range s.syncFailures {
		if !f.Resolved {
			sync.UnresolvedFailures++
		}
	}

	notif := domain.NotificationSummary{}
	for _, n := range s.notifications {
		if n.Delivered {
			notif.TotalSent++
		} else {
			notif.DeliveryFailed++
		}
	}
	for _, a := range s.alerts {
		if !a.Dismissed {
			notif.PendingAlerts++
		}
	}

	alerts := domain.AlertSummary{}
	for _, a := range s.alerts {
		alerts.Total++
		if !a.Dismissed {
			alerts.Undismissed++
		}
		switch a.AlertType {
		case "blocked_milestone":
			alerts.BlockedMilestones++
		case "overdue_milestone":
			alerts.OverdueMilestones++
		case "missing_weekly_update":
			alerts.MissingUpdates++
		}
	}

	proj := domain.ProjectionSummary{}
	var portfolioUpdated *time.Time
	for _, p := range s.projects {
		proj.ProjectCount++
		if portfolioUpdated == nil || p.AuditFields.UpdatedAt.After(*portfolioUpdated) {
			portfolioUpdated = &p.AuditFields.UpdatedAt
		}
	}
	proj.PortfolioLastUpdated = portfolioUpdated
	proj.RoadmapCount = len(s.roadmapPeriods)
	proj.MilestoneCount = len(s.milestones)
	proj.WorkItemCount = len(s.workItems)

	return domain.OperationalStatus{
		SyncStatus:         sync,
		ProjectionStatus:   proj,
		NotificationStatus: notif,
		AlertSummary:       alerts,
	}
}

func (s *Store) RunSyncForRule(role domain.WorkspaceRole, ruleID string) (domain.SyncJob, error) {
	if !HasPermission(role, PermRunSync) {
		return domain.SyncJob{}, ErrForbidden
	}
	s.mu.RLock()
	rule, ok := s.syncRules[ruleID]
	if !ok {
		s.mu.RUnlock()
		return domain.SyncJob{}, ErrNotFound
	}
	if !rule.Enabled {
		s.mu.RUnlock()
		return domain.SyncJob{}, fmt.Errorf("%w: sync rule is disabled", ErrInvalid)
	}
	s.mu.RUnlock()

	now := time.Now().UTC()
	job := domain.SyncJob{
		RuleID:      ruleID,
		Status:      "running",
		StartedAt:   &now,
		AuditFields: domain.AuditFields{CreatedAt: now, UpdatedAt: now},
	}
	s.mu.Lock()
	job.ID = s.nextID("sj")
	s.syncJobs[job.ID] = job
	if err := s.persistLocked(); err != nil {
		s.mu.Unlock()
		return domain.SyncJob{}, err
	}
	s.mu.Unlock()

	synced, failed := s.executeSync(rule)

	now = time.Now().UTC()
	s.mu.Lock()
	job.CompletedAt = &now
	job.ItemsSynced = synced
	job.ItemsFailed = failed
	if failed > 0 {
		job.Status = "completed_with_errors"
	} else {
		job.Status = "completed"
	}
	job.AuditFields.UpdatedAt = now
	s.syncJobs[job.ID] = job
	if err := s.persistLocked(); err != nil {
		s.mu.Unlock()
		return domain.SyncJob{}, err
	}
	s.mu.Unlock()

	return job, nil
}

func (s *Store) executeSync(rule domain.SyncRule) (synced, failed int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.workItems {
		if item.SourceType != domain.SourceGitLabIssue || item.ProjectID != rule.ProjectID {
			continue
		}
		if rule.Label != "" {
			matched := false
			for _, label := range item.GitLabLabels {
				if label == rule.Label {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if rule.Assignee != "" && item.GitLabAssignee != rule.Assignee {
			continue
		}
		now := time.Now().UTC()
		item.LastSyncedAt = &now
		item.AuditFields.UpdatedAt = now
		if item.MilestoneID == "" && rule.MilestoneID != "" {
			item.MilestoneID = rule.MilestoneID
		}
		s.workItems[item.ID] = item
		synced++
	}
	_ = s.persistLocked()
	return synced, failed
}

func validateMilestone(item domain.Milestone) error {
	if item.Status == domain.MilestoneActive && item.CompletionCriteria == "" {
		return fmt.Errorf("%w: completion criteria are required before activation", ErrInvalid)
	}
	return nil
}

func validateMilestoneTransition(current, next domain.Milestone) error {
	if current.Status == next.Status {
		return nil
	}
	switch current.Status {
	case domain.MilestoneNotStarted:
		if next.Status == domain.MilestoneActive {
			return nil
		}
	case domain.MilestoneActive:
		switch next.Status {
		case domain.MilestoneBlocked, domain.MilestoneCompleted, domain.MilestoneCancelled:
			return nil
		}
	case domain.MilestoneBlocked:
		if next.Status == domain.MilestoneActive {
			return nil
		}
	case domain.MilestoneCompleted, domain.MilestoneCancelled:
		return fmt.Errorf("%w: milestone status %q is terminal", ErrInvalid, current.Status)
	}
	return fmt.Errorf("%w: invalid milestone status transition from %q to %q", ErrInvalid, current.Status, next.Status)
}

func validateWorkItem(item domain.LinkedWorkItem) error {
	if strings.TrimSpace(item.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalid)
	}
	if item.SourceType == "" {
		return fmt.Errorf("%w: sourceType is required", ErrInvalid)
	}
	if item.SourceType != domain.SourceBAUTask && item.ProjectID == "" {
		return fmt.Errorf("%w: non-BAU work items must belong to a project", ErrInvalid)
	}
	if item.SourceType == domain.SourceGitLabIssue {
		if item.SourceID == "" || item.SourceURL == "" {
			return fmt.Errorf("%w: sourceId and sourceUrl are required for GitLab issues", ErrInvalid)
		}
	}
	return nil
}
