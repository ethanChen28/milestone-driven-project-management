package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"goal-manager/backend/internal/domain"
	"goal-manager/backend/internal/service"
)

type Server struct {
	cfg   Config
	store *service.Store
	mux   *http.ServeMux
}

func NewServer(cfg Config) *Server {
	s := &Server{
		cfg:   cfg,
		store: service.NewStore(),
		mux:   http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/v1/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/roadmap-periods", s.handleRoadmapPeriods)
	s.mux.HandleFunc("/api/v1/roadmap-items", s.handleRoadmapItems)
	s.mux.HandleFunc("/api/v1/projects", s.handleProjects)
	s.mux.HandleFunc("/api/v1/milestones", s.handleMilestones)
	s.mux.HandleFunc("/api/v1/workstreams", s.handleWorkstreams)
	s.mux.HandleFunc("/api/v1/work-items", s.handleWorkItems)
	s.mux.HandleFunc("/api/v1/weekly-updates", s.handleWeeklyUpdates)
	s.mux.HandleFunc("/api/v1/dashboard/portfolio", s.handlePortfolioDashboard)
	s.mux.HandleFunc("/api/v1/dashboard/roadmap", s.handleRoadmapOverview)
	s.mux.HandleFunc("/api/v1/dashboard/project", s.handleProjectDetail)
	s.mux.HandleFunc("/api/v1/dashboard/milestone", s.handleMilestoneDetail)
	s.mux.HandleFunc("/api/v1/review/weekly", s.handleWeeklyReview)
	s.mux.HandleFunc("/api/v1/gitlab-configs", s.handleGitLabConfigs)
	s.mux.HandleFunc("/api/v1/sync-rules", s.handleSyncRules)
	s.mux.HandleFunc("/api/v1/gitlab-link", s.handleGitLabLink)
	s.mux.HandleFunc("/api/v1/gitlab-unlink", s.handleGitLabUnlink)
	s.mux.HandleFunc("/api/v1/sync-jobs", s.handleSyncJobs)
	s.mux.HandleFunc("/api/v1/sync-failures", s.handleSyncFailures)
	s.mux.HandleFunc("/api/v1/sync-failures/resolve", s.handleResolveSyncFailure)
	s.mux.HandleFunc("/api/v1/webhooks/gitlab", s.handleGitLabWebhook)
	s.mux.HandleFunc("/api/v1/alerts", s.handleAlerts)
	s.mux.HandleFunc("/api/v1/alerts/dismiss", s.handleDismissAlert)
	s.mux.HandleFunc("/api/v1/notifications", s.handleNotifications)
	s.mux.HandleFunc("/api/v1/ops/status", s.handleOpsStatus)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":        "ok",
		"defaultLocale": s.cfg.DefaultLng,
	})
}

func (s *Server) handleRoadmapPeriods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetRoadmapPeriod(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, filterRoadmapPeriods(s.store.ListRoadmapPeriods(), r))
	case http.MethodPost:
		var item domain.RoadmapPeriod
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateRoadmapPeriod(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		if r.URL.Query().Get("archive") == "true" {
			item, err := s.store.ArchiveRoadmapPeriod(roleFromHeader(r), id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		var item domain.RoadmapPeriod
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertRoadmapPeriod(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleRoadmapItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetRoadmapItem(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, filterRoadmapItems(s.store.ListRoadmapItems(), r))
	case http.MethodPost:
		var item domain.RoadmapItem
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateRoadmapItem(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		var item domain.RoadmapItem
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertRoadmapItem(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetProject(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, filterProjects(s.store.ListProjects(), r))
	case http.MethodPost:
		var item domain.Project
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateProject(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		var item domain.Project
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertProject(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleMilestones(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetMilestone(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, filterMilestones(s.store.ListMilestones(), r))
	case http.MethodPost:
		var item domain.Milestone
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateMilestone(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		var item domain.Milestone
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertMilestone(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleWorkstreams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetWorkstream(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, filterWorkstreams(s.store.ListWorkstreams(), r))
	case http.MethodPost:
		var item domain.Workstream
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateWorkstream(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		var item domain.Workstream
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertWorkstream(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleWorkItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetWorkItem(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, filterWorkItems(s.store.ListWorkItems(), r))
	case http.MethodPost:
		var item domain.LinkedWorkItem
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateLinkedWorkItem(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		var item domain.LinkedWorkItem
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertWorkItem(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleWeeklyUpdates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetWeeklyUpdate(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, filterWeeklyUpdates(s.store.ListWeeklyUpdates(), r))
	case http.MethodPost:
		var item domain.WeeklyUpdate
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateWeeklyUpdate(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		var item domain.WeeklyUpdate
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertWeeklyUpdate(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handlePortfolioDashboard(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.PortfolioSummary())
}

func (s *Server) handleRoadmapOverview(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.RoadmapOverview())
}

func (s *Server) handleProjectDetail(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	detail, err := s.store.ProjectDetail(id)
	writeStoreResultWithStatus(w, http.StatusOK, detail, err)
}

func (s *Server) handleMilestoneDetail(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	detail, err := s.store.MilestoneDetail(id)
	writeStoreResultWithStatus(w, http.StatusOK, detail, err)
}

func (s *Server) handleWeeklyReview(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.WeeklyReviewView())
}

func (s *Server) handleGitLabConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetGitLabConfig(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, s.store.ListGitLabConfigs())
	case http.MethodPost:
		var item domain.GitLabConfig
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateGitLabConfig(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		var item domain.GitLabConfig
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertGitLabConfig(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		if err := s.store.DeleteGitLabConfig(roleFromHeader(r), id); err != nil {
			writeStoreResultWithStatus(w, http.StatusOK, nil, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleSyncRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if id := r.URL.Query().Get("id"); id != "" {
			item, err := s.store.GetSyncRule(id)
			writeStoreResultWithStatus(w, http.StatusOK, item, err)
			return
		}
		writeJSON(w, http.StatusOK, s.store.ListSyncRules())
	case http.MethodPost:
		var item domain.SyncRule
		if !decodeJSON(w, r, &item) {
			return
		}
		created, err := s.store.CreateSyncRule(roleFromHeader(r), item)
		writeStoreResult(w, created, err)
	case http.MethodPut:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		var item domain.SyncRule
		if !decodeJSON(w, r, &item) {
			return
		}
		updated, err := s.store.UpsertSyncRule(roleFromHeader(r), id, item)
		writeStoreResultWithStatus(w, http.StatusOK, updated, err)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
			return
		}
		if err := s.store.DeleteSyncRule(roleFromHeader(r), id); err != nil {
			writeStoreResultWithStatus(w, http.StatusOK, nil, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleGitLabLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var item domain.LinkedWorkItem
	if !decodeJSON(w, r, &item) {
		return
	}
	created, err := s.store.LinkGitLabIssue(roleFromHeader(r), item)
	writeStoreResult(w, created, err)
}

func (s *Server) handleGitLabUnlink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var body struct {
		ID string `json:"id"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	if err := s.store.UnlinkGitLabIssue(roleFromHeader(r), body.ID); err != nil {
		writeStoreResultWithStatus(w, http.StatusOK, nil, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "unlinked"})
}

func (s *Server) handleSyncJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.store.ListSyncJobs())
	case http.MethodPost:
		var body struct {
			RuleID string `json:"ruleId"`
		}
		if !decodeJSON(w, r, &body) {
			return
		}
		if body.RuleID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ruleId is required"})
			return
		}
		job, err := s.store.RunSyncForRule(roleFromHeader(r), body.RuleID)
		if err != nil {
			writeStoreResultWithStatus(w, http.StatusOK, nil, err)
			return
		}
		writeStoreResult(w, job, nil)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleSyncFailures(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, s.store.ListSyncFailures())
}

func (s *Server) handleResolveSyncFailure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var body struct {
		ID string `json:"id"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	role := roleFromHeader(r)
	if !service.HasPermission(role, service.PermManageSyncRule) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	if err := s.store.ResolveSyncFailure(body.ID); err != nil {
		writeStoreResultWithStatus(w, http.StatusOK, nil, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (s *Server) handleGitLabWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var payload struct {
		ObjectKind string `json:"object_kind"`
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if payload.ObjectKind != "issue" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	rules := s.store.ListSyncRules()
	for _, rule := range rules {
		if rule.Enabled {
			s.store.RunSyncForRule(domain.RoleAdmin, rule.ID)
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, s.store.ListAlerts())
		return
	}
	if r.Method == http.MethodPost {
		alerts := s.store.GenerateAlerts()
		writeJSON(w, http.StatusOK, alerts)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleDismissAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var body struct {
		ID string `json:"id"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	if err := s.store.DismissAlert(roleFromHeader(r), body.ID); err != nil {
		writeStoreResultWithStatus(w, http.StatusOK, nil, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "dismissed"})
}

func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, s.store.ListNotifications())
		return
	}
	if r.Method == http.MethodPost {
		var body struct {
			EventType string `json:"eventType"`
			Target    string `json:"target"`
			Title     string `json:"title"`
			Message   string `json:"message"`
		}
		if !decodeJSON(w, r, &body) {
			return
		}
		notifier := service.NewNotifier(s.store)
		events := notifier.Notify(body.EventType, body.Target, body.Title, body.Message)
		writeJSON(w, http.StatusOK, events)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleOpsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, s.store.OperationalStatus())
}

func roleFromHeader(r *http.Request) domain.WorkspaceRole {
	switch domain.WorkspaceRole(r.Header.Get("X-Role")) {
	case domain.RoleAdmin, domain.RolePortfolioManager, domain.RoleProjectOwner, domain.RoleContributor, domain.RoleViewer:
		return domain.WorkspaceRole(r.Header.Get("X-Role"))
	default:
		return domain.RoleContributor
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

func writeStoreResult(w http.ResponseWriter, payload interface{}, err error) {
	writeStoreResultWithStatus(w, http.StatusCreated, payload, err)
}

func writeStoreResultWithStatus(w http.ResponseWriter, successStatus int, payload interface{}, err error) {
	switch err {
	case nil:
		writeJSON(w, successStatus, payload)
	case service.ErrForbidden:
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	case service.ErrNotFound:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func containsIgnoreCase(value, search string) bool {
	if search == "" {
		return true
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(search))
}

func filterRoadmapPeriods(items []domain.RoadmapPeriod, r *http.Request) []domain.RoadmapPeriod {
	status := r.URL.Query().Get("status")
	owner := r.URL.Query().Get("owner")
	filtered := make([]domain.RoadmapPeriod, 0, len(items))
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if owner != "" && item.Owner != owner {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func filterRoadmapItems(items []domain.RoadmapItem, r *http.Request) []domain.RoadmapItem {
	periodID := r.URL.Query().Get("periodId")
	owner := r.URL.Query().Get("owner")
	filtered := make([]domain.RoadmapItem, 0, len(items))
	for _, item := range items {
		if periodID != "" && item.PeriodID != periodID {
			continue
		}
		if owner != "" && item.Owner != owner {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func filterProjects(items []domain.Project, r *http.Request) []domain.Project {
	roadmapItemID := r.URL.Query().Get("roadmapItemId")
	owner := r.URL.Query().Get("owner")
	status := r.URL.Query().Get("status")
	health := r.URL.Query().Get("health")
	search := r.URL.Query().Get("q")
	filtered := make([]domain.Project, 0, len(items))
	for _, item := range items {
		if roadmapItemID != "" && item.RoadmapItemID != roadmapItemID {
			continue
		}
		if owner != "" && item.Owner != owner {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		if health != "" && string(item.HealthStatus) != health {
			continue
		}
		if search != "" && !containsIgnoreCase(item.Name+" "+item.Objective, search) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func filterMilestones(items []domain.Milestone, r *http.Request) []domain.Milestone {
	projectID := r.URL.Query().Get("projectId")
	owner := r.URL.Query().Get("owner")
	status := r.URL.Query().Get("status")
	health := r.URL.Query().Get("health")
	filtered := make([]domain.Milestone, 0, len(items))
	for _, item := range items {
		if projectID != "" && item.ProjectID != projectID {
			continue
		}
		if owner != "" && item.Owner != owner {
			continue
		}
		if status != "" && string(item.Status) != status {
			continue
		}
		if health != "" && string(item.HealthStatus) != health {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func filterWorkstreams(items []domain.Workstream, r *http.Request) []domain.Workstream {
	projectID := r.URL.Query().Get("projectId")
	milestoneID := r.URL.Query().Get("milestoneId")
	owner := r.URL.Query().Get("owner")
	filtered := make([]domain.Workstream, 0, len(items))
	for _, item := range items {
		if projectID != "" && item.ProjectID != projectID {
			continue
		}
		if milestoneID != "" && item.MilestoneID != milestoneID {
			continue
		}
		if owner != "" && item.Owner != owner {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func filterWorkItems(items []domain.LinkedWorkItem, r *http.Request) []domain.LinkedWorkItem {
	projectID := r.URL.Query().Get("projectId")
	milestoneID := r.URL.Query().Get("milestoneId")
	sourceType := r.URL.Query().Get("sourceType")
	owner := r.URL.Query().Get("owner")
	status := r.URL.Query().Get("status")
	filtered := make([]domain.LinkedWorkItem, 0, len(items))
	for _, item := range items {
		if projectID != "" && item.ProjectID != projectID {
			continue
		}
		if milestoneID != "" && item.MilestoneID != milestoneID {
			continue
		}
		if sourceType != "" && string(item.SourceType) != sourceType {
			continue
		}
		if owner != "" && item.Owner != owner {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func filterWeeklyUpdates(items []domain.WeeklyUpdate, r *http.Request) []domain.WeeklyUpdate {
	projectID := r.URL.Query().Get("projectId")
	milestoneID := r.URL.Query().Get("milestoneId")
	author := r.URL.Query().Get("author")
	week := r.URL.Query().Get("week")
	filtered := make([]domain.WeeklyUpdate, 0, len(items))
	for _, item := range items {
		if projectID != "" && item.ProjectID != projectID {
			continue
		}
		if milestoneID != "" && item.MilestoneID != milestoneID {
			continue
		}
		if author != "" && item.Author != author {
			continue
		}
		if week != "" && item.Week != week {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}
