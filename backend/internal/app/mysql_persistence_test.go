package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"goal-manager/backend/internal/domain"
)

func TestMySQLPersistenceSurvivesServerRecreation(t *testing.T) {
	dsn := os.Getenv("MYSQL_INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("set MYSQL_INTEGRATION_DSN to run MySQL persistence integration test")
	}

	cfg := Config{StorageBackend: "mysql", MySQLDSN: dsn, DefaultLng: "zh-CN"}
	server, err := NewServerE(cfg)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	projBody, _ := json.Marshal(domain.Project{Name: "Persistent Project", Owner: "owner", Status: "active", HealthStatus: domain.HealthOnTrack})
	projReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(projBody))
	projReq.Header.Set("Content-Type", "application/json")
	projReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	projResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(projResp, projReq)
	if projResp.Code != http.StatusCreated {
		t.Fatalf("create project: %d %s", projResp.Code, projResp.Body.String())
	}
	var project domain.Project
	json.Unmarshal(projResp.Body.Bytes(), &project)

	msBody, _ := json.Marshal(domain.Milestone{ProjectID: project.ID, Title: "Persistent Milestone", Status: domain.MilestoneNotStarted, CompletionCriteria: "Done", PlannedDate: &now})
	msReq := httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(msBody))
	msReq.Header.Set("Content-Type", "application/json")
	msReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	msResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(msResp, msReq)
	if msResp.Code != http.StatusCreated {
		t.Fatalf("create milestone: %d %s", msResp.Code, msResp.Body.String())
	}
	var milestone domain.Milestone
	json.Unmarshal(msResp.Body.Bytes(), &milestone)

	updateBody, _ := json.Marshal(domain.WeeklyUpdate{ProjectID: project.ID, MilestoneID: milestone.ID, Author: "owner", Week: "2026-W21", Summary: "persistent update"})
	updateReq := httptest.NewRequest(http.MethodPost, "/api/v1/weekly-updates", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	updateResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusCreated {
		t.Fatalf("create weekly update: %d %s", updateResp.Code, updateResp.Body.String())
	}

	linkBody, _ := json.Marshal(domain.LinkedWorkItem{SourceType: domain.SourceGitLabIssue, SourceID: "persist-1", SourceURL: "https://gitlab.example/group/repo/-/issues/1", Title: "Persistent GitLab", ProjectID: project.ID, MilestoneID: milestone.ID})
	linkReq := httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-link", bytes.NewReader(linkBody))
	linkReq.Header.Set("Content-Type", "application/json")
	linkReq.Header.Set("X-Role", string(domain.RoleProjectOwner))
	linkResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(linkResp, linkReq)
	if linkResp.Code != http.StatusCreated {
		t.Fatalf("link gitlab: %d %s", linkResp.Code, linkResp.Body.String())
	}

	alertsReq := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", nil)
	alertsReq.Header.Set("X-Role", string(domain.RoleAdmin))
	alertsResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(alertsResp, alertsReq)

	restarted, err := NewServerE(cfg)
	if err != nil {
		t.Fatal(err)
	}
	projectReq := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/project?id="+project.ID, nil)
	projectResp := httptest.NewRecorder()
	restarted.Handler().ServeHTTP(projectResp, projectReq)
	if projectResp.Code != http.StatusOK {
		t.Fatalf("expected persisted project detail, got %d %s", projectResp.Code, projectResp.Body.String())
	}
	var detail domain.ProjectDetailView
	json.Unmarshal(projectResp.Body.Bytes(), &detail)
	if detail.Project.ID != project.ID || len(detail.Updates) == 0 || len(detail.WorkItems) == 0 {
		t.Fatalf("expected persisted project, updates, and work items: %+v", detail)
	}
}
