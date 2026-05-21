import type { LinkedWorkItem, Milestone } from "./api";

export const projectSpaceTabs = ["overview", "work-items", "milestones", "updates", "risks", "dependencies", "settings"] as const;
export type ProjectSpaceTab = (typeof projectSpaceTabs)[number];
export type ProjectWorkGroup = "milestone" | "status" | "priority" | "owner" | "sourceType" | "blocked";

export interface ProjectWorkFilters {
  milestoneId?: string;
  status?: string;
  priority?: string;
  owner?: string;
  sourceType?: string;
  blocked?: string;
  overdue?: string;
}

export interface ProjectWorkGroupResult {
  key: string;
  label: string;
  items: LinkedWorkItem[];
}

export function normalizeProjectSpaceTab(value: unknown): ProjectSpaceTab {
  const raw = Array.isArray(value) ? value[0] : value;
  return projectSpaceTabs.includes(raw as ProjectSpaceTab) ? (raw as ProjectSpaceTab) : "overview";
}

export function filterProjectWorkItems(items: LinkedWorkItem[], filters: ProjectWorkFilters): LinkedWorkItem[] {
  return items.filter((item) => {
    if (filters.milestoneId && item.milestoneId !== filters.milestoneId) return false;
    if (filters.status && displayProjectWorkStatus(item) !== filters.status) return false;
    if (filters.priority && (item.priority || "P1") !== filters.priority) return false;
    if (filters.owner && item.owner !== filters.owner) return false;
    if (filters.sourceType && item.sourceType !== filters.sourceType) return false;
    if (filters.blocked && String(item.blocked) !== filters.blocked) return false;
    if (filters.overdue && String(isProjectWorkOverdue(item)) !== filters.overdue) return false;
    return true;
  });
}

export function groupProjectWorkItems(items: LinkedWorkItem[], groupBy: ProjectWorkGroup, milestones: Milestone[] = []): ProjectWorkGroupResult[] {
  const labels = new Map<string, string>();
  milestones.forEach((milestone) => labels.set(milestone.id, milestone.title));
  const groups = new Map<string, LinkedWorkItem[]>();
  for (const item of items) {
    const key = groupKey(item, groupBy);
    const current = groups.get(key) ?? [];
    current.push(item);
    groups.set(key, current);
  }
  return Array.from(groups.entries())
    .map(([key, groupItems]) => ({ key, label: groupLabel(key, groupBy, labels), items: groupItems }))
    .sort((a, b) => a.label.localeCompare(b.label));
}

export function displayProjectWorkStatus(item: LinkedWorkItem): string {
  if (item.blocked) return "blocked";
  if (isProjectWorkOverdue(item)) return "overdue";
  return item.status || "not_started";
}

export function isProjectWorkOverdue(item: LinkedWorkItem, now = new Date()): boolean {
  return !!item.dueDate && new Date(item.dueDate) < now && item.status !== "done";
}

export function projectWorkBreadcrumb(item: LinkedWorkItem, projectName: string, milestones: Milestone[]): string {
  const milestone = milestones.find((candidate) => candidate.id === item.milestoneId);
  if (projectName && milestone) return `${projectName} / ${milestone.title}`;
  if (projectName) return projectName;
  return item.sourceType === "bau_task" ? "BAU" : item.sourceType;
}

function groupKey(item: LinkedWorkItem, groupBy: ProjectWorkGroup): string {
  if (groupBy === "milestone") return item.milestoneId || "unassigned";
  if (groupBy === "status") return displayProjectWorkStatus(item);
  if (groupBy === "priority") return item.priority || "P1";
  if (groupBy === "owner") return item.owner || "unassigned";
  if (groupBy === "sourceType") return item.sourceType || "internal_task";
  return item.blocked ? "blocked" : "not_blocked";
}

function groupLabel(key: string, groupBy: ProjectWorkGroup, milestoneLabels: Map<string, string>): string {
  if (groupBy === "milestone") return milestoneLabels.get(key) ?? (key === "unassigned" ? "未分配里程碑" : key);
  if (groupBy === "blocked") return key === "blocked" ? "阻塞" : "未阻塞";
  if (key === "unassigned") return "未分配";
  return key;
}
