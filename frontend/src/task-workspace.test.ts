import { describe, expect, it } from "vitest";

import { type LinkedWorkItem } from "./api";
import {
  buildScheduleRange,
  buildScheduleTasks,
  computeMetrics,
  defaultTaskWorkspaceState,
  deriveTaskPriority,
  filterTasks,
  normalizeWorkspaceState,
  serializeWorkspaceState,
  sortTasks,
  taskDisplayStatus,
} from "./task-workspace";

const tasks: LinkedWorkItem[] = [
  {
    id: "1",
    sourceType: "internal_task",
    sourceId: "1",
    sourceUrl: "",
    title: "Ship board",
    projectId: "p1",
    milestoneId: "m1",
    workstreamId: "w1",
    owner: "alice",
    status: "in_progress",
    estimate: "2d",
    dueDate: "2026-05-22",
    blocked: false,
  },
  {
    id: "2",
    sourceType: "gitlab_issue",
    sourceId: "2",
    sourceUrl: "https://gitlab.example/group/repo/-/issues/2",
    title: "Fix outage",
    projectId: "p2",
    milestoneId: "m2",
    workstreamId: "w2",
    owner: "bob",
    status: "todo",
    estimate: "1d",
    dueDate: "2026-05-19",
    blocked: true,
    gitlabLabels: ["ops", "urgent"],
    audit: { createdAt: "2026-05-18T00:00:00Z", updatedAt: "2026-05-19T00:00:00Z" },
  },
];

describe("task workspace helpers", () => {
  it("serializes and normalizes workspace state", () => {
    const state = defaultTaskWorkspaceState();
    state.view = "gantt";
    state.filters.owner = "alice";
    state.sortBy = "priority";
    state.groupBy = "project";
    const params = serializeWorkspaceState(state);
    expect(params.view).toBe("gantt");
    expect(params.owner).toBe("alice");
    expect(normalizeWorkspaceState(params).groupBy).toBe("project");
  });

  it("filters tasks by derived status and tags", () => {
    const state = defaultTaskWorkspaceState();
    state.filters.status = "blocked";
    expect(filterTasks(tasks, state)).toHaveLength(1);
    state.filters = { ...state.filters, status: "", tag: "urgent" };
    expect(filterTasks(tasks, state)).toHaveLength(1);
  });

  it("derives priority and task status", () => {
    expect(deriveTaskPriority(tasks[0])).toBe("P1");
    expect(deriveTaskPriority(tasks[1])).toBe("P0");
    expect(taskDisplayStatus(tasks[1])).toBe("blocked");
  });

  it("sorts tasks and computes metrics", () => {
    const state = defaultTaskWorkspaceState();
    state.sortBy = "dueDate";
    const sorted = sortTasks(tasks, state);
    expect(sorted[0].id).toBe("2");
    const metrics = computeMetrics(tasks, new Date("2026-05-20T00:00:00Z"));
    expect(metrics.total).toBe(2);
    expect(metrics.blocked).toBe(1);
    expect(metrics.overdue).toBe(0);
    expect(metrics.nearDue).toBe(1);
  });

  it("builds schedule metadata for gantt view", () => {
    const range = buildScheduleRange(tasks, "month", new Date("2026-05-20T00:00:00Z"));
    const scheduled = buildScheduleTasks(tasks, range);
    expect(scheduled).toHaveLength(2);
    expect(scheduled[0].widthPercent).toBeGreaterThan(0);
  });
});
