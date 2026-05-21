import { describe, expect, it } from "vitest";
import type { LinkedWorkItem, Milestone } from "./api";
import { filterProjectWorkItems, groupProjectWorkItems, normalizeProjectSpaceTab, projectWorkBreadcrumb } from "./project-space";

const milestones: Milestone[] = [
  { id: "m1", projectId: "p1", title: "Alpha", status: "active", healthStatus: "on_track", owner: "alice", completionCriteria: "Done", progressPercent: 40, riskLevel: "low", dependencySummary: "" },
];

const items: LinkedWorkItem[] = [
  { id: "w1", sourceType: "internal_task", sourceId: "", sourceUrl: "", title: "Build", projectId: "p1", milestoneId: "m1", workstreamId: "", owner: "alice", status: "todo", priority: "P0", estimate: "1d", blocked: true, dueDate: "2020-01-01T00:00:00Z" },
  { id: "w2", sourceType: "gitlab_issue", sourceId: "2", sourceUrl: "", title: "Fix", projectId: "p1", milestoneId: "", workstreamId: "", owner: "bob", status: "done", priority: "P2", estimate: "1d", blocked: false },
];

describe("project space helpers", () => {
  it("normalizes tab route state", () => {
    expect(normalizeProjectSpaceTab("work-items")).toBe("work-items");
    expect(normalizeProjectSpaceTab("invalid")).toBe("overview");
    expect(normalizeProjectSpaceTab(["risks"])).toBe("risks");
  });

  it("filters and groups project work items", () => {
    expect(filterProjectWorkItems(items, { milestoneId: "m1", blocked: "true", overdue: "true" }).map((item) => item.id)).toEqual(["w1"]);
    const groups = groupProjectWorkItems(items, "milestone", milestones);
    expect(groups.map((group) => group.label)).toEqual(["Alpha", "未分配里程碑"]);
  });

  it("builds project work breadcrumbs", () => {
    expect(projectWorkBreadcrumb(items[0], "Project A", milestones)).toBe("Project A / Alpha");
    expect(projectWorkBreadcrumb(items[1], "Project A", milestones)).toBe("Project A");
  });
});
