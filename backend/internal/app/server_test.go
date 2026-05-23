package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goal-manager/backend/internal/domain"
	"goal-manager/backend/internal/service"
)

func createProjectWithRole(s *Server, role domain.WorkspaceRole) domain.Project {
	body, _ := json.Marshal(domain.Project{Name: "Test", Status: "active"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(role))
	resp := httptest.NewRecorder()
	s.Handler().ServeHTTP(resp, req)
	var proj domain.Project
	json.Unmarshal(resp.Body.Bytes(), &proj)
	return proj
}

func createMilestoneWithRole(s *Server, role domain.WorkspaceRole, projID string) domain.Milestone {
	body, _ := json.Marshal(domain.Milestone{
		ProjectID:          projID,
		Title:              "M1",
		Status:             domain.MilestoneNotStarted,
		CompletionCriteria: "Done",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(role))
	resp := httptest.NewRecorder()
	s.Handler().ServeHTTP(resp, req)
	var ms domain.Milestone
	json.Unmarshal(resp.Body.Bytes(), &ms)
	return ms
}

func updateMilestoneWithRole(t *testing.T, s *Server, role domain.WorkspaceRole, ms domain.Milestone) (domain.Milestone, int) {
	t.Helper()
	body, _ := json.Marshal(ms)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/milestones?id="+ms.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(role))
	resp := httptest.NewRecorder()
	s.Handler().ServeHTTP(resp, req)
	var updated domain.Milestone
	json.Unmarshal(resp.Body.Bytes(), &updated)
	return updated, resp.Code
}

func TestHealthEndpoint(t *testing.T) {
	server := NewServer(LoadConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestMilestoneValidationRequiresCompletionCriteria(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(domain.Milestone{
		ProjectID: "prj-1",
		Title:     "Launch beta",
		Status:    domain.MilestoneActive,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleProjectOwner))
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestViewerCannotCreateRoadmapPeriod(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(domain.RoadmapPeriod{Title: "Q3"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/roadmap-periods", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleViewer))
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestNonBAUWorkItemRequiresProject(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceInternalTask,
		Title:      "Review onboarding flow",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestWorkItemCanBeDeleted(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)
	body, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceInternalTask,
		Title:      "Temporary task",
		ProjectID:  proj.ID,
		Owner:      "alice",
		Status:     "todo",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleContributor))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", resp.Code, resp.Body.String())
	}
	var created domain.LinkedWorkItem
	if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/work-items?id="+created.ID, nil)
	delReq.Header.Set("X-Role", string(domain.RoleContributor))
	delResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(delResp, delReq)
	if delResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", delResp.Code, delResp.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/work-items?id="+created.ID, nil)
	getReq.Header.Set("X-Role", string(domain.RoleContributor))
	getResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d body=%s", getResp.Code, getResp.Body.String())
	}
}

func TestRoadmapPeriodCanBeUpdatedAndArchived(t *testing.T) {
	server := NewServer(LoadConfig())
	createBody, _ := json.Marshal(domain.RoadmapPeriod{
		Title:       "Q3",
		Description: "Quarter",
		Owner:       "alice",
		Status:      "active",
		Priority:    "P1",
		PeriodStart: time.Now().UTC(),
		PeriodEnd:   time.Now().UTC().Add(24 * time.Hour),
	})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/roadmap-periods", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Role", string(domain.RolePortfolioManager))
	createResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(createResp, createReq)

	var created domain.RoadmapPeriod
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	created.Title = "Q3 Updated"
	updateBody, _ := json.Marshal(created)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/roadmap-periods?id="+created.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("X-Role", string(domain.RolePortfolioManager))
	updateResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(updateResp, updateReq)

	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", updateResp.Code)
	}

	archiveReq := httptest.NewRequest(http.MethodPut, "/api/v1/roadmap-periods?id="+created.ID+"&archive=true", nil)
	archiveReq.Header.Set("X-Role", string(domain.RolePortfolioManager))
	archiveResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(archiveResp, archiveReq)

	if archiveResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", archiveResp.Code)
	}
}

func TestProjectFilterAndDashboardSummary(t *testing.T) {
	server := NewServer(LoadConfig())

	projectBody, _ := json.Marshal(domain.Project{
		Name:          "Alpha",
		Objective:     "Improve onboarding",
		Owner:         "alice",
		Status:        "active",
		HealthStatus:  domain.HealthAtRisk,
		RoadmapItemID: "ri-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(projectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleProjectOwner))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)

	filterReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects?owner=alice&health=at_risk&q=alpha", nil)
	filterResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(filterResp, filterReq)
	if filterResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", filterResp.Code)
	}

	dashboardReq := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/portfolio", nil)
	dashboardResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(dashboardResp, dashboardReq)
	if dashboardResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", dashboardResp.Code)
	}
}

func TestMilestoneRiskAndWorkItemGitLabFilters(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)

	highRiskBody, _ := json.Marshal(domain.Milestone{ProjectID: proj.ID, Title: "High Risk", Status: domain.MilestoneNotStarted, CompletionCriteria: "Done", Owner: "owner", RiskLevel: "high"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(highRiskBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleProjectOwner))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create high risk milestone: %d %s", resp.Code, resp.Body.String())
	}

	lowRiskBody, _ := json.Marshal(domain.Milestone{ProjectID: proj.ID, Title: "Low Risk", Status: domain.MilestoneNotStarted, CompletionCriteria: "Done", Owner: "owner", RiskLevel: "low"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(lowRiskBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleProjectOwner))
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create low risk milestone: %d %s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/milestones?risk=high", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	var milestones []domain.Milestone
	json.Unmarshal(resp.Body.Bytes(), &milestones)
	if resp.Code != http.StatusOK || len(milestones) != 1 || milestones[0].RiskLevel != "high" {
		t.Fatalf("expected one high-risk milestone, code=%d body=%s", resp.Code, resp.Body.String())
	}

	linkBody, _ := json.Marshal(domain.LinkedWorkItem{SourceType: domain.SourceGitLabIssue, SourceID: "101", SourceURL: "https://gitlab.example/group/repo/-/issues/101", Title: "Issue", ProjectID: proj.ID, Owner: "dev", Status: "opened", GitLabLabels: []string{"milestone::x"}, GitLabAssignee: "dev"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-link", bytes.NewReader(linkBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleProjectOwner))
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("link gitlab issue: %d %s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/work-items?sourceType=gitlab_issue&gitlabContext=group/repo", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	var workItems []domain.LinkedWorkItem
	json.Unmarshal(resp.Body.Bytes(), &workItems)
	if resp.Code != http.StatusOK || len(workItems) != 1 || workItems[0].SourceID != "101" {
		t.Fatalf("expected one gitlab work item, code=%d body=%s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/work-items?sourceType=gitlab_issue&gitlabContext=no-match", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	workItems = nil
	json.Unmarshal(resp.Body.Bytes(), &workItems)
	if resp.Code != http.StatusOK || len(workItems) != 0 {
		t.Fatalf("expected no gitlab work items, code=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestProjectSpaceRollupsRisksAndDependencies(t *testing.T) {
	server := NewServer(LoadConfig())
	project := createScopedProject(t, server, "lead", []string{"alice"})
	past := time.Now().Add(-24 * time.Hour)

	milestoneBody, _ := json.Marshal(domain.Milestone{
		ProjectID:          project.ID,
		Title:              "Blocked Milestone",
		Status:             domain.MilestoneBlocked,
		CompletionCriteria: "Done",
		Owner:              "lead",
		PlannedDate:        &past,
		RiskLevel:          "high",
		DependencySummary:  "Waiting for security review",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(milestoneBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create milestone: %d %s", resp.Code, resp.Body.String())
	}

	workBody, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceExternalDependency,
		Title:      "Vendor dependency",
		ProjectID:  project.ID,
		Owner:      "alice",
		Status:     "blocked",
		Priority:   "P0",
		Blocked:    true,
		DueDate:    &past,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(workBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create work item: %d %s", resp.Code, resp.Body.String())
	}

	updateBody, _ := json.Marshal(domain.WeeklyUpdate{
		ProjectID:       project.ID,
		Author:          "alice",
		Week:            "2026-W21",
		Summary:         "Needs decision",
		Risk:            "Scope risk",
		Blockers:        "Vendor API unavailable",
		DecisionsNeeded: "Approve fallback",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/weekly-updates", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create weekly update: %d %s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/project-space?id="+project.ID, nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("project space: %d %s", resp.Code, resp.Body.String())
	}
	var view domain.ProjectSpaceView
	if err := json.Unmarshal(resp.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if view.Rollups.BlockedMilestones != 1 || view.Rollups.OverdueMilestones != 1 || view.Rollups.BlockedWorkItems != 1 || view.Rollups.ExternalDependencies != 1 {
		t.Fatalf("unexpected rollups: %+v", view.Rollups)
	}
	if len(view.Risks) < 3 {
		t.Fatalf("expected milestone, work item, and update risk signals, got %+v", view.Risks)
	}
	if len(view.Dependencies) < 2 {
		t.Fatalf("expected milestone and work item dependencies, got %+v", view.Dependencies)
	}
}

func TestProjectWorkItemFiltersIncludePriorityBlockedAndOverdue(t *testing.T) {
	server := NewServer(LoadConfig())
	project := createScopedProject(t, server, "lead", []string{"alice"})
	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	blockedBody, _ := json.Marshal(domain.LinkedWorkItem{SourceType: domain.SourceInternalTask, Title: "Blocked P0", ProjectID: project.ID, Owner: "alice", Status: "blocked", Priority: "P0", Blocked: true, DueDate: &past})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(blockedBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create blocked item: %d %s", resp.Code, resp.Body.String())
	}

	normalBody, _ := json.Marshal(domain.LinkedWorkItem{SourceType: domain.SourceInternalTask, Title: "Normal P2", ProjectID: project.ID, Owner: "alice", Status: "todo", Priority: "P2", DueDate: &future})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(normalBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create normal item: %d %s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/work-items?projectId="+project.ID+"&priority=P0&blocked=true&overdue=true", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("filter work items: %d %s", resp.Code, resp.Body.String())
	}
	var items []domain.LinkedWorkItem
	if err := json.Unmarshal(resp.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "Blocked P0" {
		t.Fatalf("expected only blocked P0 item, got %+v", items)
	}
}

// --- RBAC Tests ---

func TestViewerCannotCreateProject(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(domain.Project{Name: "Test"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleViewer))
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestViewerCannotCreateWorkItem(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceBAUTask,
		Title:      "BAU task",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleViewer))
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestViewerCannotSubmitWeeklyUpdate(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(domain.WeeklyUpdate{Summary: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/weekly-updates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleViewer))
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestContributorCanCreateWorkItemButNotProject(t *testing.T) {
	server := NewServer(LoadConfig())

	body, _ := json.Marshal(domain.Project{Name: "Test"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleContributor))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for project creation, got %d", resp.Code)
	}

	wiBody, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceBAUTask,
		Title:      "BAU task",
	})
	wiReq := httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(wiBody))
	wiReq.Header.Set("Content-Type", "application/json")
	wiReq.Header.Set("X-Role", string(domain.RoleContributor))
	wiResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(wiResp, wiReq)
	if wiResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for work item creation, got %d", wiResp.Code)
	}
}

func TestWeeklyUpdateRequiresProjectContextAndKeepsMilestone(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)
	ms := createMilestoneWithRole(server, domain.RoleProjectOwner, proj.ID)

	missingProjectBody, _ := json.Marshal(domain.WeeklyUpdate{Author: "owner", Week: "2026-W21", Summary: "summary"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/weekly-updates", bytes.NewReader(missingProjectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleProjectOwner))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing projectId, got %d", resp.Code)
	}

	body, _ := json.Marshal(domain.WeeklyUpdate{ProjectID: proj.ID, MilestoneID: ms.ID, Author: "owner", Week: "2026-W21", Summary: "summary"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/weekly-updates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleProjectOwner))
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for weekly update, got %d body: %s", resp.Code, resp.Body.String())
	}
	var update domain.WeeklyUpdate
	json.Unmarshal(resp.Body.Bytes(), &update)
	if update.ProjectID != proj.ID || update.MilestoneID != ms.ID {
		t.Fatalf("expected project and milestone context to be preserved: %+v", update)
	}
}

func TestOnlyAdminCanManageGitLabConfig(t *testing.T) {
	server := NewServer(LoadConfig())

	for _, role := range []domain.WorkspaceRole{domain.RolePortfolioManager, domain.RoleProjectOwner, domain.RoleContributor} {
		body, _ := json.Marshal(domain.GitLabConfig{Name: "test", BaseURL: "https://gitlab.com"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-configs", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Role", string(role))
		resp := httptest.NewRecorder()
		server.Handler().ServeHTTP(resp, req)
		if resp.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for role %s, got %d", role, resp.Code)
		}
	}

	body, _ := json.Marshal(domain.GitLabConfig{Name: "test", BaseURL: "https://gitlab.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-configs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for admin, got %d", resp.Code)
	}
}

func TestProjectOwnerCanManageMilestones(t *testing.T) {
	server := NewServer(LoadConfig())

	projBody, _ := json.Marshal(domain.Project{Name: "Test", Owner: "alice"})
	projReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(projBody))
	projReq.Header.Set("Content-Type", "application/json")
	projReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	projResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(projResp, projReq)

	var proj domain.Project
	json.Unmarshal(projResp.Body.Bytes(), &proj)

	msBody, _ := json.Marshal(domain.Milestone{
		ProjectID:          proj.ID,
		Title:              "M1",
		Status:             domain.MilestoneNotStarted,
		CompletionCriteria: "Done when shipped",
	})
	msReq := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(msBody))
	msReq.Header.Set("Content-Type", "application/json")
	msReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	msResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(msResp, msReq)

	if msResp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", msResp.Code)
	}
}

// --- GitLab Link/Unlink Tests ---

func TestGitLabLinkRequiresCorrectSourceType(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceInternalTask,
		Title:      "Internal",
		SourceID:   "123",
		SourceURL:  "https://gitlab.com/issue/123",
		ProjectID:  "prj-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestGitLabLinkAndUnlink(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)

	linkBody, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceGitLabIssue,
		SourceID:   "42",
		SourceURL:  "https://gitlab.com/group/repo/issues/42",
		Title:      "Bug fix",
		ProjectID:  proj.ID,
	})
	linkReq := httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-link", bytes.NewReader(linkBody))
	linkReq.Header.Set("Content-Type", "application/json")
	linkReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	linkResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(linkResp, linkReq)

	if linkResp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", linkResp.Code)
	}

	var linked domain.LinkedWorkItem
	json.Unmarshal(linkResp.Body.Bytes(), &linked)
	if linked.LastSyncedAt == nil {
		t.Fatal("expected lastSyncedAt to be set")
	}

	unlinkBody, _ := json.Marshal(map[string]string{"id": linked.ID})
	unlinkReq := httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-unlink", bytes.NewReader(unlinkBody))
	unlinkReq.Header.Set("Content-Type", "application/json")
	unlinkReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	unlinkResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(unlinkResp, unlinkReq)

	if unlinkResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", unlinkResp.Code)
	}
}

// --- Sync Rule Tests ---

func TestSyncRuleRequiresConfigAndProject(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(domain.SyncRule{Label: "bug"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync-rules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestSyncRuleCRUD(t *testing.T) {
	server := NewServer(LoadConfig())

	glcBody, _ := json.Marshal(domain.GitLabConfig{Name: "test", BaseURL: "https://gitlab.com", Group: "mygroup"})
	glcReq := httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-configs", bytes.NewReader(glcBody))
	glcReq.Header.Set("Content-Type", "application/json")
	glcReq.Header.Set("X-Role", string(domain.RoleAdmin))
	glcResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(glcResp, glcReq)
	var glc domain.GitLabConfig
	json.Unmarshal(glcResp.Body.Bytes(), &glc)

	proj := createProjectWithRole(server, domain.RoleProjectOwner)

	srBody, _ := json.Marshal(domain.SyncRule{
		GitLabConfigID: glc.ID,
		ProjectID:      proj.ID,
		Label:          "milestone::m1",
		Enabled:        true,
	})
	srReq := httptest.NewRequest(http.MethodPost, "/api/v1/sync-rules", bytes.NewReader(srBody))
	srReq.Header.Set("Content-Type", "application/json")
	srReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	srResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(srResp, srReq)

	if srResp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", srResp.Code)
	}

	var rule domain.SyncRule
	json.Unmarshal(srResp.Body.Bytes(), &rule)
	if rule.Label != "milestone::m1" {
		t.Fatalf("expected label milestone::m1, got %s", rule.Label)
	}
}

// --- Reporting View Tests ---

func TestRoadmapOverviewEndpoint(t *testing.T) {
	server := NewServer(LoadConfig())

	rpBody, _ := json.Marshal(domain.RoadmapPeriod{Title: "Q3", Status: "active", Owner: "alice"})
	rpReq := httptest.NewRequest(http.MethodPost, "/api/v1/roadmap-periods", bytes.NewReader(rpBody))
	rpReq.Header.Set("Content-Type", "application/json")
	rpReq.Header.Set("X-Role", string(domain.RolePortfolioManager))
	rpResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(rpResp, rpReq)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/roadmap", nil)
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestProjectDetailEndpoint(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/project?id="+proj.ID, nil)
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestMilestoneDetailEndpoint(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)
	ms := createMilestoneWithRole(server, domain.RoleProjectOwner, proj.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/milestone?id="+ms.ID, nil)
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestWeeklyReviewView(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)

	blockedMS := domain.Milestone{
		ProjectID:          proj.ID,
		Title:              "Blocked M1",
		Status:             domain.MilestoneBlocked,
		CompletionCriteria: "Done",
	}
	msBody, _ := json.Marshal(blockedMS)
	msReq := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(msBody))
	msReq.Header.Set("Content-Type", "application/json")
	msReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	server.Handler().ServeHTTP(httptest.NewRecorder(), msReq)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/review/weekly", nil)
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var view domain.WeeklyReviewView
	json.Unmarshal(resp.Body.Bytes(), &view)
	if len(view.BlockedMilestones) == 0 {
		t.Fatal("expected blocked milestones in weekly review")
	}
}

// --- Alert Tests ---

func TestAlertGeneration(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)

	past := time.Now().UTC().Add(-48 * time.Hour)
	msBody, _ := json.Marshal(domain.Milestone{
		ProjectID:          proj.ID,
		Title:              "Overdue M1",
		Status:             domain.MilestoneActive,
		CompletionCriteria: "Done",
		PlannedDate:        &past,
	})
	msReq := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(msBody))
	msReq.Header.Set("Content-Type", "application/json")
	msReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	msResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(msResp, msReq)

	alertReq := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", nil)
	alertResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(alertResp, alertReq)

	if alertResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", alertResp.Code)
	}

	var alerts []domain.Alert
	json.Unmarshal(alertResp.Body.Bytes(), &alerts)
	if len(alerts) == 0 {
		t.Fatal("expected alerts to be generated")
	}

	found := false
	for _, a := range alerts {
		if a.AlertType == "overdue_milestone" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected overdue_milestone alert")
	}
}

func TestAlertDismissal(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)

	past := time.Now().UTC().Add(-48 * time.Hour)
	msBody, _ := json.Marshal(domain.Milestone{
		ProjectID:          proj.ID,
		Title:              "M1",
		Status:             domain.MilestoneActive,
		CompletionCriteria: "Done",
		PlannedDate:        &past,
	})
	msReq := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(msBody))
	msReq.Header.Set("Content-Type", "application/json")
	msReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	server.Handler().ServeHTTP(httptest.NewRecorder(), msReq)

	alertReq := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", nil)
	alertResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(alertResp, alertReq)

	var alerts []domain.Alert
	json.Unmarshal(alertResp.Body.Bytes(), &alerts)
	if len(alerts) == 0 {
		t.Fatal("expected alerts")
	}

	dismissBody, _ := json.Marshal(map[string]string{"id": alerts[0].ID})
	dismissReq := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/dismiss", bytes.NewReader(dismissBody))
	dismissReq.Header.Set("Content-Type", "application/json")
	dismissReq.Header.Set("X-Role", string(domain.RoleAdmin))
	dismissResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(dismissResp, dismissReq)

	if dismissResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", dismissResp.Code)
	}
}

// --- Notification Tests ---

func TestNotificationCreation(t *testing.T) {
	server := NewServer(LoadConfig())

	body, _ := json.Marshal(map[string]string{
		"eventType": "test",
		"target":    "user1",
		"title":     "Test Alert",
		"message":   "This is a test",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var events []domain.NotificationEvent
	json.Unmarshal(resp.Body.Bytes(), &events)
	if len(events) != 2 {
		t.Fatalf("expected 2 notification events (email+feishu), got %d", len(events))
	}
}

func TestNotificationList(t *testing.T) {
	server := NewServer(LoadConfig())

	body, _ := json.Marshal(map[string]string{
		"eventType": "test",
		"target":    "user1",
		"title":     "T",
		"message":   "M",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(httptest.NewRecorder(), req)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	listResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(listResp, listReq)

	if listResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listResp.Code)
	}
}

// --- Webhook Test ---

func TestGitLabWebhookIgnoresNonIssue(t *testing.T) {
	server := NewServer(LoadConfig())
	body, _ := json.Marshal(map[string]string{"object_kind": "push"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/gitlab", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var result map[string]string
	json.Unmarshal(resp.Body.Bytes(), &result)
	if result["status"] != "ignored" {
		t.Fatalf("expected ignored, got %s", result["status"])
	}
}

// --- Source-of-Truth Boundary Test ---

func TestSyncUpdatesGitLabFieldsOnly(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)

	linkBody, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType:     domain.SourceGitLabIssue,
		SourceID:       "99",
		SourceURL:      "https://gitlab.com/issues/99",
		Title:          "Original Title",
		ProjectID:      proj.ID,
		GitLabLabels:   []string{"bug"},
		GitLabState:    "opened",
		GitLabAssignee: "alice",
	})
	linkReq := httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-link", bytes.NewReader(linkBody))
	linkReq.Header.Set("Content-Type", "application/json")
	linkReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	linkResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(linkResp, linkReq)

	if linkResp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body: %s", linkResp.Code, linkResp.Body.String())
	}

	var linked domain.LinkedWorkItem
	json.Unmarshal(linkResp.Body.Bytes(), &linked)

	if linked.Title != "Original Title" {
		t.Fatalf("PM title should be preserved, got %s", linked.Title)
	}
	if linked.LastSyncedAt == nil {
		t.Fatal("lastSyncedAt should be set on link")
	}
}

// --- Milestone Completion Test ---

func TestMilestoneCompletionRecordsDate(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)
	ms := createMilestoneWithRole(server, domain.RoleProjectOwner, proj.ID)

	ms.Status = domain.MilestoneActive
	active, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, ms)
	if code != http.StatusOK {
		t.Fatalf("expected active transition 200, got %d", code)
	}

	active.Status = domain.MilestoneCompleted
	updated, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, active)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}

	if updated.CompletedDate == nil {
		t.Fatal("expected completedDate to be set")
	}
}

func TestMilestoneTransitionStateMachine(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)
	ms := createMilestoneWithRole(server, domain.RoleProjectOwner, proj.ID)

	ms.Status = domain.MilestoneCompleted
	if _, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, ms); code != http.StatusBadRequest {
		t.Fatalf("expected not_started->completed to be rejected, got %d", code)
	}

	ms.Status = domain.MilestoneActive
	active, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, ms)
	if code != http.StatusOK {
		t.Fatalf("expected not_started->active 200, got %d", code)
	}

	active.Status = domain.MilestoneBlocked
	blocked, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, active)
	if code != http.StatusOK {
		t.Fatalf("expected active->blocked 200, got %d", code)
	}

	blocked.Status = domain.MilestoneActive
	active, code = updateMilestoneWithRole(t, server, domain.RoleProjectOwner, blocked)
	if code != http.StatusOK {
		t.Fatalf("expected blocked->active 200, got %d", code)
	}

	active.Status = domain.MilestoneCancelled
	cancelled, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, active)
	if code != http.StatusOK {
		t.Fatalf("expected active->cancelled 200, got %d", code)
	}

	cancelled.Status = domain.MilestoneActive
	if _, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, cancelled); code != http.StatusBadRequest {
		t.Fatalf("expected cancelled->active to be rejected, got %d", code)
	}
}

func TestCompletedMilestoneCannotReenterActiveState(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)
	ms := createMilestoneWithRole(server, domain.RoleProjectOwner, proj.ID)

	ms.Status = domain.MilestoneActive
	active, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, ms)
	if code != http.StatusOK {
		t.Fatalf("expected active transition 200, got %d", code)
	}

	active.Status = domain.MilestoneCompleted
	completed, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, active)
	if code != http.StatusOK {
		t.Fatalf("expected completed transition 200, got %d", code)
	}

	completed.Status = domain.MilestoneActive
	if _, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, completed); code != http.StatusBadRequest {
		t.Fatalf("expected completed->active to be rejected, got %d", code)
	}
}

func createScopedProject(t *testing.T, s *Server, owner string, participants []string) domain.Project {
	t.Helper()
	body, _ := json.Marshal(domain.Project{Name: "Scoped " + owner, Owner: owner, Participants: participants, Status: "active", HealthStatus: domain.HealthOnTrack})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	req.Header.Set("X-User", "admin")
	resp := httptest.NewRecorder()
	s.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create scoped project: expected 201, got %d body=%s", resp.Code, resp.Body.String())
	}
	var project domain.Project
	if err := json.Unmarshal(resp.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}
	return project
}

func createScopedTask(t *testing.T, s *Server, projectID, owner string) domain.LinkedWorkItem {
	t.Helper()
	body, _ := json.Marshal(domain.LinkedWorkItem{SourceType: domain.SourceInternalTask, Title: "Scoped Task", ProjectID: projectID, Owner: owner, Status: "todo"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	req.Header.Set("X-User", "admin")
	resp := httptest.NewRecorder()
	s.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create scoped task: expected 201, got %d body=%s", resp.Code, resp.Body.String())
	}
	var task domain.LinkedWorkItem
	if err := json.Unmarshal(resp.Body.Bytes(), &task); err != nil {
		t.Fatal(err)
	}
	return task
}

func TestContributorTaskOwnershipWithXUser(t *testing.T) {
	server := NewServer(LoadConfig())
	project := createScopedProject(t, server, "lead", []string{"alice", "bob"})

	body, _ := json.Marshal(domain.LinkedWorkItem{SourceType: domain.SourceInternalTask, Title: "Alice Task", ProjectID: project.ID, Owner: "alice", Status: "todo"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleContributor))
	req.Header.Set("X-User", "alice")
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected alice to create own task, got %d body=%s", resp.Code, resp.Body.String())
	}
	var task domain.LinkedWorkItem
	json.Unmarshal(resp.Body.Bytes(), &task)

	task.Title = "Bob Edit Attempt"
	task.Owner = "bob"
	updateBody, _ := json.Marshal(task)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/work-items?id="+task.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("X-Role", string(domain.RoleContributor))
	updateReq.Header.Set("X-User", "bob")
	updateResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusForbidden {
		t.Fatalf("expected bob update forbidden, got %d body=%s", updateResp.Code, updateResp.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/work-items?id="+task.ID, nil)
	deleteReq.Header.Set("X-Role", string(domain.RoleContributor))
	deleteReq.Header.Set("X-User", "bob")
	deleteResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusForbidden {
		t.Fatalf("expected bob delete forbidden, got %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
}

func TestProjectOwnerScopeWithXUser(t *testing.T) {
	server := NewServer(LoadConfig())
	aliceProject := createScopedProject(t, server, "alice", []string{"dev"})
	bobProject := createScopedProject(t, server, "bob", []string{"dev"})
	bobTask := createScopedTask(t, server, bobProject.ID, "dev")

	bobTask.Title = "Alice should not edit"
	body, _ := json.Marshal(bobTask)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/work-items?id="+bobTask.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleProjectOwner))
	req.Header.Set("X-User", "alice")
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected cross-project task update forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}

	projectBody, _ := json.Marshal(bobProject)
	projectReq := httptest.NewRequest(http.MethodPut, "/api/v1/projects?id="+bobProject.ID, bytes.NewReader(projectBody))
	projectReq.Header.Set("Content-Type", "application/json")
	projectReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	projectReq.Header.Set("X-User", "alice")
	projectResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(projectResp, projectReq)
	if projectResp.Code != http.StatusForbidden {
		t.Fatalf("expected cross-project project update forbidden, got %d body=%s", projectResp.Code, projectResp.Body.String())
	}

	milestoneBody, _ := json.Marshal(domain.Milestone{ProjectID: aliceProject.ID, Title: "Alice Milestone", Status: domain.MilestoneNotStarted, CompletionCriteria: "Ship"})
	milestoneReq := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(milestoneBody))
	milestoneReq.Header.Set("Content-Type", "application/json")
	milestoneReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	milestoneReq.Header.Set("X-User", "alice")
	milestoneResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(milestoneResp, milestoneReq)
	if milestoneResp.Code != http.StatusCreated {
		t.Fatalf("expected owned milestone create allowed, got %d body=%s", milestoneResp.Code, milestoneResp.Body.String())
	}
}

func TestContributorWeeklyUpdateScopeWithXUser(t *testing.T) {
	server := NewServer(LoadConfig())
	project := createScopedProject(t, server, "lead", []string{"alice"})

	body, _ := json.Marshal(domain.WeeklyUpdate{ProjectID: project.ID, Author: "alice", Week: "2026-W21", Summary: "progress"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/weekly-updates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleContributor))
	req.Header.Set("X-User", "alice")
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected participant update allowed, got %d body=%s", resp.Code, resp.Body.String())
	}

	bobBody, _ := json.Marshal(domain.WeeklyUpdate{ProjectID: project.ID, Author: "bob", Week: "2026-W21", Summary: "progress"})
	bobReq := httptest.NewRequest(http.MethodPost, "/api/v1/weekly-updates", bytes.NewReader(bobBody))
	bobReq.Header.Set("Content-Type", "application/json")
	bobReq.Header.Set("X-Role", string(domain.RoleContributor))
	bobReq.Header.Set("X-User", "bob")
	bobResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(bobResp, bobReq)
	if bobResp.Code != http.StatusForbidden {
		t.Fatalf("expected non-participant update forbidden, got %d body=%s", bobResp.Code, bobResp.Body.String())
	}
}

func TestContributorCannotCompleteMilestoneWithXUser(t *testing.T) {
	server := NewServer(LoadConfig())
	project := createScopedProject(t, server, "lead", []string{"alice"})
	milestoneBody, _ := json.Marshal(domain.Milestone{ProjectID: project.ID, Title: "Acceptance", Status: domain.MilestoneActive, CompletionCriteria: "Review\nApprove"})
	milestoneReq := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(milestoneBody))
	milestoneReq.Header.Set("Content-Type", "application/json")
	milestoneReq.Header.Set("X-Role", string(domain.RoleAdmin))
	milestoneReq.Header.Set("X-User", "admin")
	milestoneResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(milestoneResp, milestoneReq)
	if milestoneResp.Code != http.StatusCreated {
		t.Fatalf("expected milestone create, got %d body=%s", milestoneResp.Code, milestoneResp.Body.String())
	}
	var milestone domain.Milestone
	json.Unmarshal(milestoneResp.Body.Bytes(), &milestone)

	milestone.Status = domain.MilestoneCompleted
	body, _ := json.Marshal(milestone)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/milestones?id="+milestone.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleContributor))
	req.Header.Set("X-User", "alice")
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected contributor completion forbidden, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestMilestoneHealthAndProgressUpdateDoesNotChangeStatus(t *testing.T) {
	server := NewServer(LoadConfig())
	proj := createProjectWithRole(server, domain.RoleProjectOwner)
	ms := createMilestoneWithRole(server, domain.RoleProjectOwner, proj.ID)

	ms.Status = domain.MilestoneCompleted
	ms.HealthStatus = domain.HealthAtRisk
	ms.ProgressPercent = 45
	if _, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, ms); code != http.StatusBadRequest {
		t.Fatalf("expected not_started->completed to be rejected before health/progress assertion, got %d", code)
	}

	ms.Status = domain.MilestoneNotStarted
	ms.HealthStatus = domain.HealthAtRisk
	ms.ProgressPercent = 45
	updated, code := updateMilestoneWithRole(t, server, domain.RoleProjectOwner, ms)
	if code != http.StatusOK {
		t.Fatalf("expected same-status update 200, got %d", code)
	}
	if updated.Status != domain.MilestoneNotStarted || updated.HealthStatus != domain.HealthAtRisk || updated.ProgressPercent != 45 {
		t.Fatalf("health/progress update changed unexpected fields: %+v", updated)
	}
}

func testToken(t *testing.T, secret, sub string, role domain.WorkspaceRole) string {
	t.Helper()
	now := time.Now().UTC()
	token, err := service.IssueIdentityToken(secret, service.IdentityClaims{Sub: sub, WorkspaceID: "default", Roles: []string{string(role)}, DisplayName: sub, Email: sub + "@example.local", Provider: "builtin", Version: 1, Iat: now.Unix(), Exp: now.Add(time.Hour).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestTokenAuthAllowsTrustedProjectOwner(t *testing.T) {
	secret := "test-secret"
	server := NewServer(Config{StorageBackend: "memory", AuthMode: "token", TokenSecret: secret, AppEnv: "development"})
	adminToken := testToken(t, secret, "tester", domain.RoleAdmin)
	ownerToken := testToken(t, secret, "alice", domain.RoleProjectOwner)

	projectBody, _ := json.Marshal(domain.Project{Name: "Token Project", Owner: "alice", Participants: []string{"alice"}, Status: "active", HealthStatus: domain.HealthOnTrack})
	projectReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(projectBody))
	projectReq.Header.Set("Content-Type", "application/json")
	projectReq.Header.Set("Authorization", "Bearer "+adminToken)
	projectResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(projectResp, projectReq)
	if projectResp.Code != http.StatusCreated {
		t.Fatalf("expected admin token to create project, got %d body=%s", projectResp.Code, projectResp.Body.String())
	}
	var project domain.Project
	json.Unmarshal(projectResp.Body.Bytes(), &project)

	project.Objective = "updated by trusted owner"
	updateBody, _ := json.Marshal(project)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/projects?id="+project.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+ownerToken)
	updateReq.Header.Set("X-User", "mallory")
	updateReq.Header.Set("X-Role", string(domain.RoleAdmin))
	updateResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected trusted owner token to win over spoofed headers, got %d body=%s", updateResp.Code, updateResp.Body.String())
	}
}

func TestTokenAuthRejectsMissingOrInvalidTokenMutation(t *testing.T) {
	server := NewServer(Config{StorageBackend: "memory", AuthMode: "token", TokenSecret: "test-secret", AppEnv: "development"})
	body, _ := json.Marshal(domain.Project{Name: "No Token", Owner: "alice", Status: "active"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(domain.RoleAdmin))
	req.Header.Set("X-User", "alice")
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected spoofed headers rejected in token mode, got %d body=%s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid")
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected invalid token rejected, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestUserDirectoryAndIdentityMigrationReport(t *testing.T) {
	server := NewServer(Config{StorageBackend: "memory", AuthMode: "dev-header", AppEnv: "development"})
	project := createScopedProject(t, server, "alice", []string{"alice", "unknown-user"})
	_ = project

	usersReq := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	usersResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(usersResp, usersReq)
	if usersResp.Code != http.StatusOK {
		t.Fatalf("expected users endpoint 200, got %d", usersResp.Code)
	}
	var users []service.UserProfile
	json.Unmarshal(usersResp.Body.Bytes(), &users)
	if len(users) == 0 {
		t.Fatal("expected seeded user directory")
	}

	reportReq := httptest.NewRequest(http.MethodGet, "/api/v1/identity-migration/report", nil)
	reportResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(reportResp, reportReq)
	if reportResp.Code != http.StatusOK {
		t.Fatalf("expected migration report 200, got %d", reportResp.Code)
	}
	var report service.IdentityMigrationReport
	json.Unmarshal(reportResp.Body.Bytes(), &report)
	if report.UnresolvedReferences == 0 {
		t.Fatalf("expected unresolved identity reference for unknown-user, got %+v", report)
	}
}

func TestProductionRejectsDevHeaderAuthMode(t *testing.T) {
	_, err := NewServerE(Config{StorageBackend: "memory", AuthMode: "dev-header", AppEnv: "production"})
	if err == nil {
		t.Fatal("expected production dev-header configuration to fail")
	}
}
