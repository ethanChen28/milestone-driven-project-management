package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goal-manager/backend/internal/domain"
)

func seedMVPWorkspace(t *testing.T, server *Server) {
	adminRole := string(domain.RoleAdmin)
	pmRole := string(domain.RolePortfolioManager)
	poRole := string(domain.RoleProjectOwner)
	contribRole := string(domain.RoleContributor)

	// 1. Create roadmap period
	rpBody, _ := json.Marshal(domain.RoadmapPeriod{
		Title:       "H1 2026",
		Description: "First half 2026 planning period",
		Owner:       "alice",
		Status:      "active",
		Priority:    "P0",
		PeriodStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/roadmap-periods", bytes.NewReader(rpBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", pmRole)
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create roadmap period: %d %s", resp.Code, resp.Body.String())
	}
	var rp domain.RoadmapPeriod
	json.Unmarshal(resp.Body.Bytes(), &rp)

	// 2. Create roadmap items
	ri1Body, _ := json.Marshal(domain.RoadmapItem{
		PeriodID:    rp.ID,
		Title:       "Platform Reliability",
		Description: "Improve platform stability and observability",
		Owner:       "bob",
		Priority:    "P0",
		Status:      "active",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/roadmap-items", bytes.NewReader(ri1Body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", pmRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create roadmap item 1: %d %s", resp.Code, resp.Body.String())
	}
	var ri1 domain.RoadmapItem
	json.Unmarshal(resp.Body.Bytes(), &ri1)

	ri2Body, _ := json.Marshal(domain.RoadmapItem{
		PeriodID:    rp.ID,
		Title:       "User Onboarding Revamp",
		Description: "Redesign the onboarding experience",
		Owner:       "carol",
		Priority:    "P1",
		Status:      "active",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/roadmap-items", bytes.NewReader(ri2Body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", pmRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create roadmap item 2: %d %s", resp.Code, resp.Body.String())
	}
	var ri2 domain.RoadmapItem
	json.Unmarshal(resp.Body.Bytes(), &ri2)

	// 3. Create projects linked to roadmap items
	proj1Body, _ := json.Marshal(domain.Project{
		Name:          "Reliability Sprint",
		Summary:       "Improve core platform reliability",
		Objective:     "Reduce P1 incidents by 50%",
		RoadmapItemID: ri1.ID,
		Owner:         "bob",
		Participants:  []string{"bob", "dave", "eve"},
		ProjectType:   "engineering",
		Status:        "active",
		HealthStatus:  domain.HealthOnTrack,
		Priority:      "P0",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(proj1Body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", poRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create project 1: %d %s", resp.Code, resp.Body.String())
	}
	var proj1 domain.Project
	json.Unmarshal(resp.Body.Bytes(), &proj1)

	proj2Body, _ := json.Marshal(domain.Project{
		Name:          "Onboarding Redesign",
		Summary:       "Redesign new user onboarding",
		Objective:     "Increase 7-day activation rate to 60%",
		RoadmapItemID: ri2.ID,
		Owner:         "carol",
		Participants:  []string{"carol", "frank"},
		ProjectType:   "product",
		Status:        "active",
		HealthStatus:  domain.HealthAtRisk,
		Priority:      "P1",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(proj2Body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", poRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create project 2: %d %s", resp.Code, resp.Body.String())
	}
	var proj2 domain.Project
	json.Unmarshal(resp.Body.Bytes(), &proj2)

	// 4. Create milestones
	futureDate := time.Now().UTC().Add(30 * 24 * time.Hour)
	pastDate := time.Now().UTC().Add(-48 * time.Hour)
	ms1Body, _ := json.Marshal(domain.Milestone{
		ProjectID:          proj1.ID,
		Title:              "Observability Stack",
		MilestoneType:      "deliverable",
		Description:        "Deploy full observability stack",
		CompletionCriteria: "All services emitting metrics, traces, and structured logs",
		Owner:              "bob",
		PlannedDate:        &futureDate,
		Status:             domain.MilestoneActive,
		HealthStatus:       domain.HealthOnTrack,
		ProgressPercent:    40,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(ms1Body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", poRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create milestone 1: %d %s", resp.Code, resp.Body.String())
	}

	ms2Body, _ := json.Marshal(domain.Milestone{
		ProjectID:          proj2.ID,
		Title:              "Design Validation",
		MilestoneType:      "deliverable",
		Description:        "Validate new onboarding designs with users",
		CompletionCriteria: "10 user interviews completed with positive signal",
		Owner:              "carol",
		PlannedDate:        &pastDate,
		Status:             domain.MilestoneBlocked,
		HealthStatus:       domain.HealthOffTrack,
		ProgressPercent:    20,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/milestones", bytes.NewReader(ms2Body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", poRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create milestone 2: %d %s", resp.Code, resp.Body.String())
	}
	var ms2 domain.Milestone
	json.Unmarshal(resp.Body.Bytes(), &ms2)

	// 5. Create workstreams
	ws1Body, _ := json.Marshal(domain.Workstream{
		ProjectID:   proj1.ID,
		MilestoneID: ms2.ID,
		Name:        "Metrics Pipeline",
		Owner:       "dave",
		Status:      "active",
		Description: "Build metrics collection and forwarding",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workstreams", bytes.NewReader(ws1Body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", poRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create workstream: %d %s", resp.Code, resp.Body.String())
	}
	var ws1 domain.Workstream
	json.Unmarshal(resp.Body.Bytes(), &ws1)

	// 6. Create GitLab-linked work items
	linkBody, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType:     domain.SourceGitLabIssue,
		SourceID:       "101",
		SourceURL:      "https://gitlab.com/platform/infra/issues/101",
		Title:          "Set up Prometheus metrics endpoint",
		ProjectID:      proj1.ID,
		MilestoneID:    ms2.ID,
		WorkstreamID:   ws1.ID,
		Owner:          "dave",
		Status:         "in_progress",
		GitLabLabels:   []string{"milestone::observability", "team::platform"},
		GitLabState:    "opened",
		GitLabAssignee: "dave",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-link", bytes.NewReader(linkBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", poRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: link GitLab issue: %d %s", resp.Code, resp.Body.String())
	}

	// 7. Create internal and BAU work items
	bauBody, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceBAUTask,
		Title:      "Monthly infrastructure cost review",
		Owner:      "bob",
		Status:     "pending",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(bauBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", contribRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create BAU work item: %d %s", resp.Code, resp.Body.String())
	}

	internalBody, _ := json.Marshal(domain.LinkedWorkItem{
		SourceType: domain.SourceInternalTask,
		Title:      "Draft onboarding email sequence",
		ProjectID:  proj2.ID,
		MilestoneID: ms2.ID,
		Owner:      "frank",
		Status:     "in_progress",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/work-items", bytes.NewReader(internalBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", contribRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create internal work item: %d %s", resp.Code, resp.Body.String())
	}

	// 8. Create weekly updates
	wu1Body, _ := json.Marshal(domain.WeeklyUpdate{
		ProjectID:       proj1.ID,
		Author:          "bob",
		Week:            "2026-W20",
		Summary:         "Good progress on metrics pipeline",
		Progress:        "40% complete, on track",
		Risk:            "Dependency on infra team for endpoint provisioning",
		Blockers:        "",
		DecisionsNeeded: "Need approval for Prometheus vs Grafana Agent",
		NextSteps:       "Complete endpoint configuration and run load tests",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/weekly-updates", bytes.NewReader(wu1Body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", poRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create weekly update: %d %s", resp.Code, resp.Body.String())
	}

	// 9. Create GitLab config and sync rule
	glcBody, _ := json.Marshal(domain.GitLabConfig{
		Name:       "Platform GitLab",
		BaseURL:    "https://gitlab.com",
		Group:      "platform",
		Repository: "infra",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/gitlab-configs", bytes.NewReader(glcBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", adminRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create GitLab config: %d %s", resp.Code, resp.Body.String())
	}
	var glc domain.GitLabConfig
	json.Unmarshal(resp.Body.Bytes(), &glc)

	srBody, _ := json.Marshal(domain.SyncRule{
		GitLabConfigID: glc.ID,
		ProjectID:      proj1.ID,
		Label:          "milestone::observability",
		Enabled:        true,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/sync-rules", bytes.NewReader(srBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", poRole)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("seed: create sync rule: %d %s", resp.Code, resp.Body.String())
	}
}

func TestMVPEndToEnd(t *testing.T) {
	server := NewServer(LoadConfig())
	seedMVPWorkspace(t, server)

	// Verify roadmap overview
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/roadmap", nil)
	resp := httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("e2e: roadmap overview: %d", resp.Code)
	}
	var roadmapOverviews []domain.RoadmapOverviewItem
	json.Unmarshal(resp.Body.Bytes(), &roadmapOverviews)
	if len(roadmapOverviews) == 0 {
		t.Fatal("e2e: expected roadmap overview items")
	}

	// Verify portfolio dashboard
	req = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/portfolio", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("e2e: portfolio dashboard: %d", resp.Code)
	}
	var summary domain.PortfolioSummary
	json.Unmarshal(resp.Body.Bytes(), &summary)
	if summary.ActiveProjects < 2 {
		t.Fatalf("e2e: expected at least 2 active projects, got %d", summary.ActiveProjects)
	}
	if summary.BlockedMilestones < 1 {
		t.Fatalf("e2e: expected at least 1 blocked milestone, got %d", summary.BlockedMilestones)
	}
	if summary.OverdueMilestones < 1 {
		t.Fatalf("e2e: expected at least 1 overdue milestone, got %d", summary.OverdueMilestones)
	}

	// Verify weekly review
	req = httptest.NewRequest(http.MethodGet, "/api/v1/review/weekly", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("e2e: weekly review: %d", resp.Code)
	}
	var review domain.WeeklyReviewView
	json.Unmarshal(resp.Body.Bytes(), &review)
	if len(review.BlockedMilestones) == 0 {
		t.Fatal("e2e: expected blocked milestones in weekly review")
	}
	if len(review.Updates) == 0 {
		t.Fatal("e2e: expected updates in weekly review")
	}

	// Verify alert generation
	req = httptest.NewRequest(http.MethodPost, "/api/v1/alerts", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("e2e: alert generation: %d", resp.Code)
	}
	var alerts []domain.Alert
	json.Unmarshal(resp.Body.Bytes(), &alerts)
	if len(alerts) == 0 {
		t.Fatal("e2e: expected alerts to be generated")
	}

	hasOverdue := false
	hasBlocked := false
	hasMissing := false
	for _, a := range alerts {
		switch a.AlertType {
		case "overdue_milestone":
			hasOverdue = true
		case "blocked_milestone":
			hasBlocked = true
		case "missing_weekly_update":
			hasMissing = true
		}
	}
	if !hasOverdue {
		t.Fatal("e2e: expected overdue_milestone alert")
	}
	if !hasBlocked {
		t.Fatal("e2e: expected blocked_milestone alert")
	}
	if !hasMissing {
		t.Fatal("e2e: expected missing_weekly_update alert")
	}

	// Verify operational status
	req = httptest.NewRequest(http.MethodGet, "/api/v1/ops/status", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("e2e: ops status: %d", resp.Code)
	}
	var ops domain.OperationalStatus
	json.Unmarshal(resp.Body.Bytes(), &ops)
	if ops.ProjectionStatus.ProjectCount < 2 {
		t.Fatalf("e2e: expected at least 2 projects in ops status, got %d", ops.ProjectionStatus.ProjectCount)
	}
	if ops.SyncStatus.ActiveRules < 1 {
		t.Fatal("e2e: expected at least 1 active sync rule")
	}
	if ops.AlertSummary.Undismissed == 0 {
		t.Fatal("e2e: expected undismissed alerts in ops status")
	}

	// Verify project detail
	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects?owner=bob", nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("e2e: project list: %d", resp.Code)
	}
	var projects []domain.Project
	json.Unmarshal(resp.Body.Bytes(), &projects)
	if len(projects) == 0 {
		t.Fatal("e2e: expected projects for owner bob")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/project?id="+projects[0].ID, nil)
	resp = httptest.NewRecorder()
	server.Handler().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("e2e: project detail: %d", resp.Code)
	}
	var projDetail domain.ProjectDetailView
	json.Unmarshal(resp.Body.Bytes(), &projDetail)
	if projDetail.Project.Name != "Reliability Sprint" {
		t.Fatalf("e2e: expected project name 'Reliability Sprint', got %s", projDetail.Project.Name)
	}
	if len(projDetail.WorkItems) == 0 {
		t.Fatal("e2e: expected work items in project detail")
	}

	// Verify BAU vs milestone work ratio
	if summary.BAUWorkItems == 0 {
		t.Fatal("e2e: expected BAU work items")
	}
	if summary.MilestoneWorkItems == 0 {
		t.Fatal("e2e: expected milestone work items")
	}

	// Verify sync rule execution
	srReq := httptest.NewRequest(http.MethodGet, "/api/v1/sync-rules", nil)
	srResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(srResp, srReq)
	var rules []domain.SyncRule
	json.Unmarshal(srResp.Body.Bytes(), &rules)
	if len(rules) == 0 {
		t.Fatal("e2e: expected sync rules")
	}

	syncBody, _ := json.Marshal(map[string]string{"ruleId": rules[0].ID})
	syncReq := httptest.NewRequest(http.MethodPost, "/api/v1/sync-jobs", bytes.NewReader(syncBody))
	syncReq.Header.Set("Content-Type", "application/json")
	syncReq.Header.Set("X-Role", string(domain.RoleAdmin))
	syncResp := httptest.NewRecorder()
	server.Handler().ServeHTTP(syncResp, syncReq)
	if syncResp.Code != http.StatusCreated {
		t.Fatalf("e2e: sync job execution: %d %s", syncResp.Code, syncResp.Body.String())
	}
}
