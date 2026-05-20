package service

import (
	"testing"
	"time"

	"goal-manager/backend/internal/domain"
)

func TestMemoryRepositoryContractCoversDomainFlows(t *testing.T) {
	store, err := NewStoreWithRepository(NewMemoryRepository())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()

	period, err := store.CreateRoadmapPeriod(domain.RolePortfolioManager, domain.RoadmapPeriod{Title: "H1", Description: "roadmap", Owner: "pm", Status: "active", Priority: "P1", PeriodStart: now, PeriodEnd: now.AddDate(0, 3, 0)})
	if err != nil {
		t.Fatalf("create roadmap period: %v", err)
	}
	period.Title = "H1 Updated"
	if _, err := store.UpsertRoadmapPeriod(domain.RolePortfolioManager, period.ID, period); err != nil {
		t.Fatalf("update roadmap period: %v", err)
	}
	if got, err := store.GetRoadmapPeriod(period.ID); err != nil || got.Title != "H1 Updated" || len(store.ListRoadmapPeriods()) != 1 {
		t.Fatalf("roadmap period contract failed: got=%+v err=%v", got, err)
	}

	item, err := store.CreateRoadmapItem(domain.RolePortfolioManager, domain.RoadmapItem{PeriodID: period.ID, Title: "Outcome", Description: "desc", Owner: "pm", Priority: "P1", Status: "active"})
	if err != nil {
		t.Fatalf("create roadmap item: %v", err)
	}
	item.Status = "done"
	if _, err := store.UpsertRoadmapItem(domain.RolePortfolioManager, item.ID, item); err != nil {
		t.Fatalf("update roadmap item: %v", err)
	}
	if got, err := store.GetRoadmapItem(item.ID); err != nil || got.Status != "done" || len(store.ListRoadmapItems()) != 1 {
		t.Fatalf("roadmap item contract failed: got=%+v err=%v", got, err)
	}

	project, err := store.CreateProject(domain.RoleProjectOwner, domain.Project{Name: "Project", Summary: "summary", Objective: "objective", RoadmapItemID: item.ID, Owner: "owner", ProjectType: "delivery", Status: "active", HealthStatus: domain.HealthOnTrack, Priority: "P1"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	project.HealthStatus = domain.HealthAtRisk
	if _, err := store.UpsertProject(domain.RoleProjectOwner, project.ID, project); err != nil {
		t.Fatalf("update project: %v", err)
	}
	if got, err := store.GetProject(project.ID); err != nil || got.HealthStatus != domain.HealthAtRisk || len(store.ListProjects()) != 1 {
		t.Fatalf("project contract failed: got=%+v err=%v", got, err)
	}

	milestone, err := store.CreateMilestone(domain.RoleProjectOwner, domain.Milestone{ProjectID: project.ID, Title: "Milestone", MilestoneType: "delivery", Description: "desc", CompletionCriteria: "criteria", Owner: "owner", PlannedDate: &now, Status: domain.MilestoneNotStarted, HealthStatus: domain.HealthOnTrack, RiskLevel: "low"})
	if err != nil {
		t.Fatalf("create milestone: %v", err)
	}
	milestone.Status = domain.MilestoneActive
	if _, err := store.UpsertMilestone(domain.RoleProjectOwner, milestone.ID, milestone); err != nil {
		t.Fatalf("update milestone: %v", err)
	}
	if got, err := store.GetMilestone(milestone.ID); err != nil || got.Status != domain.MilestoneActive || len(store.ListMilestones()) != 1 {
		t.Fatalf("milestone contract failed: got=%+v err=%v", got, err)
	}

	workstream, err := store.CreateWorkstream(domain.RoleProjectOwner, domain.Workstream{ProjectID: project.ID, MilestoneID: milestone.ID, Name: "Stream", Owner: "owner", Status: "active", Description: "desc"})
	if err != nil {
		t.Fatalf("create workstream: %v", err)
	}
	workstream.Status = "done"
	if _, err := store.UpsertWorkstream(domain.RoleProjectOwner, workstream.ID, workstream); err != nil {
		t.Fatalf("update workstream: %v", err)
	}
	if got, err := store.GetWorkstream(workstream.ID); err != nil || got.Status != "done" || len(store.ListWorkstreams()) != 1 {
		t.Fatalf("workstream contract failed: got=%+v err=%v", got, err)
	}

	work, err := store.CreateLinkedWorkItem(domain.RoleContributor, domain.LinkedWorkItem{SourceType: domain.SourceInternalTask, Title: "Task", ProjectID: project.ID, MilestoneID: milestone.ID, WorkstreamID: workstream.ID, Owner: "dev", Status: "open"})
	if err != nil {
		t.Fatalf("create work item: %v", err)
	}
	work.Status = "closed"
	if _, err := store.UpsertWorkItem(domain.RoleContributor, work.ID, work); err != nil {
		t.Fatalf("update work item: %v", err)
	}
	if got, err := store.GetWorkItem(work.ID); err != nil || got.Status != "closed" || len(store.ListWorkItems()) != 1 {
		t.Fatalf("work item contract failed: got=%+v err=%v", got, err)
	}

	update, err := store.CreateWeeklyUpdate(domain.RoleContributor, domain.WeeklyUpdate{ProjectID: project.ID, MilestoneID: milestone.ID, Author: "dev", Week: "2026-W21", Summary: "summary"})
	if err != nil {
		t.Fatalf("create weekly update: %v", err)
	}
	update.Summary = "updated summary"
	if _, err := store.UpsertWeeklyUpdate(domain.RoleContributor, update.ID, update); err != nil {
		t.Fatalf("update weekly update: %v", err)
	}
	if got, err := store.GetWeeklyUpdate(update.ID); err != nil || got.Summary != "updated summary" || len(store.ListWeeklyUpdates()) != 1 {
		t.Fatalf("weekly update contract failed: got=%+v err=%v", got, err)
	}

	gitlabConfig, err := store.CreateGitLabConfig(domain.RoleAdmin, domain.GitLabConfig{Name: "GitLab", BaseURL: "https://gitlab.example", Group: "group", Repository: "repo"})
	if err != nil {
		t.Fatalf("create gitlab config: %v", err)
	}
	gitlabConfig.Repository = "repo2"
	if _, err := store.UpsertGitLabConfig(domain.RoleAdmin, gitlabConfig.ID, gitlabConfig); err != nil {
		t.Fatalf("update gitlab config: %v", err)
	}
	if got, err := store.GetGitLabConfig(gitlabConfig.ID); err != nil || got.Repository != "repo2" || len(store.ListGitLabConfigs()) != 1 {
		t.Fatalf("gitlab config contract failed: got=%+v err=%v", got, err)
	}

	rule, err := store.CreateSyncRule(domain.RoleProjectOwner, domain.SyncRule{GitLabConfigID: gitlabConfig.ID, ProjectID: project.ID, MilestoneID: milestone.ID, Label: "bug", Enabled: true})
	if err != nil {
		t.Fatalf("create sync rule: %v", err)
	}
	rule.Assignee = "dev"
	if _, err := store.UpsertSyncRule(domain.RoleProjectOwner, rule.ID, rule); err != nil {
		t.Fatalf("update sync rule: %v", err)
	}
	if got, err := store.GetSyncRule(rule.ID); err != nil || got.Assignee != "dev" || len(store.ListSyncRules()) != 1 {
		t.Fatalf("sync rule contract failed: got=%+v err=%v", got, err)
	}

	job, err := store.CreateSyncJob(domain.RoleAdmin, domain.SyncJob{RuleID: rule.ID})
	if err != nil {
		t.Fatalf("create sync job: %v", err)
	}
	if err := store.CompleteSyncJob(job.ID, 2, 1, "partial"); err != nil {
		t.Fatalf("complete sync job: %v", err)
	}
	if jobs := store.ListSyncJobs(); len(jobs) != 1 || jobs[0].Status != "failed" {
		t.Fatalf("sync job contract failed: %+v", jobs)
	}

	failure := store.RecordSyncFailure(domain.SyncFailure{WorkItemID: work.ID, SourceID: "1", Error: "boom"})
	if err := store.ResolveSyncFailure(failure.ID); err != nil {
		t.Fatalf("resolve sync failure: %v", err)
	}
	if failures := store.ListSyncFailures(); len(failures) != 1 || !failures[0].Resolved {
		t.Fatalf("sync failure contract failed: %+v", failures)
	}

	alerts := store.GenerateAlerts()
	if len(alerts) == 0 {
		t.Fatal("expected generated alerts")
	}
	if err := store.DismissAlert(domain.RoleAdmin, alerts[0].ID); err != nil {
		t.Fatalf("dismiss alert: %v", err)
	}
	if listed := store.ListAlerts(); len(listed) == 0 {
		t.Fatal("expected listed alerts")
	}

	notification := store.SaveNotification(domain.NotificationEvent{EventType: "alert", Target: "owner", Channel: "email", Title: "title", Message: "message"})
	if notification.ID == "" || len(store.ListNotifications()) != 1 {
		t.Fatalf("notification contract failed: %+v", notification)
	}
}
