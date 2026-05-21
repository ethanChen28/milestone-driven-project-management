import { gitlabLabels, type LinkedWorkItem } from "./api";

export const taskViews = ["list", "board", "gantt", "timeline", "project", "priority"] as const;
export type TaskView = (typeof taskViews)[number];

export const taskGroupings = ["none", "project", "status", "priority", "owner", "sourceType"] as const;
export type TaskGrouping = (typeof taskGroupings)[number];

export const taskSortFields = ["dueDate", "updatedAt", "createdAt", "title", "priority"] as const;
export type TaskSortField = (typeof taskSortFields)[number];
export type SortDirection = "asc" | "desc";
export type ScheduleScale = "week" | "month" | "quarter" | "year";

export interface TaskWorkspaceFilters {
  projectId: string;
  milestoneId: string;
  owner: string;
  status: string;
  sourceType: string;
  priority: string;
  tag: string;
  q: string;
}

export interface TaskWorkspaceState {
  view: TaskView;
  filters: TaskWorkspaceFilters;
  groupBy: TaskGrouping;
  sortBy: TaskSortField;
  sortDir: SortDirection;
  scale: ScheduleScale;
}

export interface TaskMetrics {
  total: number;
  completed: number;
  inProgress: number;
  notStarted: number;
  overdue: number;
  blocked: number;
  nearDue: number;
}

export interface GroupedTasks {
  key: string;
  label: string;
  items: LinkedWorkItem[];
}

export interface ScheduleTask extends LinkedWorkItem {
  startAt: Date;
  endAt: Date;
  leftPercent: number;
  widthPercent: number;
  overdue: boolean;
}

export interface ScheduleRange {
  start: Date;
  end: Date;
  todayPercent: number;
}

export const defaultTaskFilters = (): TaskWorkspaceFilters => ({
  projectId: "",
  milestoneId: "",
  owner: "",
  status: "",
  sourceType: "",
  priority: "",
  tag: "",
  q: "",
});

export const defaultTaskWorkspaceState = (): TaskWorkspaceState => ({
  view: "list",
  filters: defaultTaskFilters(),
  groupBy: "none",
  sortBy: "dueDate",
  sortDir: "asc",
  scale: "month",
});

const priorityOrder: Record<string, number> = { P0: 0, P1: 1, P2: 2, P3: 3 };
const statusOrder: Record<string, number> = { blocked: 0, overdue: 1, in_progress: 2, not_started: 3, done: 4, cancelled: 5 };

function toLower(value: string | undefined) {
  return (value ?? "").toLowerCase();
}

export function normalizeWorkspaceState(query: Record<string, unknown>): TaskWorkspaceState {
  const state = defaultTaskWorkspaceState();
  const get = (key: string) => String(query[key] ?? "");
  const view = get("view");
  if (taskViews.includes(view as TaskView)) state.view = view as TaskView;
  const groupBy = get("groupBy");
  if (taskGroupings.includes(groupBy as TaskGrouping)) state.groupBy = groupBy as TaskGrouping;
  const sortBy = get("sortBy");
  if (taskSortFields.includes(sortBy as TaskSortField)) state.sortBy = sortBy as TaskSortField;
  const sortDir = get("sortDir");
  if (sortDir === "asc" || sortDir === "desc") state.sortDir = sortDir;
  const scale = get("scale");
  if (["week", "month", "quarter", "year"].includes(scale)) state.scale = scale as ScheduleScale;
  Object.keys(state.filters).forEach((key) => {
    const value = get(key);
    if (value) state.filters[key as keyof TaskWorkspaceFilters] = value;
  });
  return state;
}

export function serializeWorkspaceState(state: TaskWorkspaceState): Record<string, string> {
  const params: Record<string, string> = {
    view: state.view,
    groupBy: state.groupBy,
    sortBy: state.sortBy,
    sortDir: state.sortDir,
    scale: state.scale,
  };
  Object.entries(state.filters).forEach(([key, value]) => {
    if (value) params[key] = value;
  });
  return params;
}

export function normalizeStatus(value: string): string {
  const lower = value.toLowerCase();
  if (["todo", "open", "opened", "not_started", "not-started"].includes(lower)) return "not_started";
  if (["in_progress", "doing", "progress", "active"].includes(lower)) return "in_progress";
  if (["done", "completed", "closed"].includes(lower)) return "done";
  if (["cancelled", "canceled"].includes(lower)) return "cancelled";
  if (lower === "blocked") return "blocked";
  return lower || "not_started";
}

export function deriveTaskPriority(item: LinkedWorkItem): string {
  const explicit = item.priority?.trim();
  if (explicit) return explicit;
  if (item.blocked) return "P0";
  if (isOverdue(item.dueDate)) return "P0";
  const days = estimateDays(item.estimate);
  const due = parseDate(item.dueDate);
  if (due) {
    const leadDays = Math.max(0, Math.ceil((due.getTime() - Date.now()) / 86400000));
    if (leadDays <= 3) return "P1";
    if (leadDays <= 7) return "P2";
  }
  if (days <= 1) return "P1";
  if (days <= 3) return "P2";
  return "P3";
}

export function taskTags(item: LinkedWorkItem): string[] {
  const tags = new Set<string>();
  tags.add(item.sourceType);
  if (item.blocked) tags.add("blocked");
  gitlabLabels(item).forEach((tag) => tags.add(tag));
  item.tags?.forEach((tag) => tags.add(tag));
  if (item.dueDate && isOverdue(item.dueDate)) tags.add("overdue");
  return [...tags];
}

export function taskDisplayStatus(item: LinkedWorkItem): string {
  if (item.blocked) return "blocked";
  const normalized = normalizeStatus(item.status);
  if (item.dueDate && isOverdue(item.dueDate) && normalized !== "done") return "overdue";
  if (normalized === "done" && item.dueDate && isOverdue(item.dueDate)) return "done";
  if (normalized === "not_started" && item.dueDate && isNearDue(item.dueDate)) return "near_due";
  return normalized;
}

export function filterTasks(tasks: LinkedWorkItem[], state: TaskWorkspaceState): LinkedWorkItem[] {
  const filters = state.filters;
  const needle = filters.q.trim().toLowerCase();
  return tasks.filter((item) => {
    if (filters.projectId && item.projectId !== filters.projectId) return false;
    if (filters.milestoneId && item.milestoneId !== filters.milestoneId) return false;
    if (filters.owner && item.owner !== filters.owner) return false;
    if (filters.status && taskDisplayStatus(item) !== filters.status) return false;
    if (filters.sourceType && item.sourceType !== filters.sourceType) return false;
    if (filters.priority && deriveTaskPriority(item) !== filters.priority) return false;
    if (filters.tag && !taskTags(item).includes(filters.tag)) return false;
    if (!needle) return true;
    const haystack = [
      item.title,
      item.sourceId,
      item.sourceUrl,
      item.owner,
      item.status,
      item.projectId,
      item.milestoneId,
      item.workstreamId,
      deriveTaskPriority(item),
      ...taskTags(item),
    ]
      .map((value) => toLower(value))
      .join(" ");
    return haystack.includes(needle);
  });
}

export function sortTasks(tasks: LinkedWorkItem[], state: TaskWorkspaceState): LinkedWorkItem[] {
  const sorted = [...tasks];
  const direction = state.sortDir === "asc" ? 1 : -1;
  sorted.sort((a, b) => {
    const left = sortValue(a, state.sortBy);
    const right = sortValue(b, state.sortBy);
    if (left < right) return -1 * direction;
    if (left > right) return 1 * direction;
    return a.title.localeCompare(b.title);
  });
  return sorted;
}

export function groupTasks(tasks: LinkedWorkItem[], groupBy: TaskGrouping): GroupedTasks[] {
  if (groupBy === "none") return [{ key: "all", label: "All", items: tasks }];
  const groups = new Map<string, LinkedWorkItem[]>();
  tasks.forEach((item) => {
    const key = groupValue(item, groupBy);
    const bucket = groups.get(key) ?? [];
    bucket.push(item);
    groups.set(key, bucket);
  });
  return [...groups.entries()]
    .map(([key, items]) => ({ key, label: groupLabel(groupBy, key), items }))
    .sort((a, b) => a.label.localeCompare(b.label, "zh-Hans-CN"));
}

export function computeMetrics(tasks: LinkedWorkItem[], now = new Date()): TaskMetrics {
  const metrics: TaskMetrics = { total: tasks.length, completed: 0, inProgress: 0, notStarted: 0, overdue: 0, blocked: 0, nearDue: 0 };
  tasks.forEach((item) => {
    const status = taskDisplayStatus(item);
    if (status === "done") metrics.completed += 1;
    else if (status === "in_progress") metrics.inProgress += 1;
    else if (status === "not_started") metrics.notStarted += 1;
    if (status === "overdue") metrics.overdue += 1;
    if (item.blocked) metrics.blocked += 1;
    if (item.dueDate && isNearDue(item.dueDate, now)) metrics.nearDue += 1;
  });
  return metrics;
}

export function buildScheduleRange(tasks: LinkedWorkItem[], scale: ScheduleScale, now = new Date()): ScheduleRange {
  const dates: Date[] = [];
  tasks.forEach((item) => {
    const start = scheduleStart(item);
    const end = scheduleEnd(item, start);
    dates.push(start, end);
  });
  const min = dates.length ? new Date(Math.min(...dates.map((date) => date.getTime()))) : new Date(now);
  const max = dates.length ? new Date(Math.max(...dates.map((date) => date.getTime()))) : new Date(now);
  const padding = scale === "year" ? 30 : scale === "quarter" ? 14 : scale === "month" ? 7 : 3;
  const start = shiftDays(min, -padding);
  const end = shiftDays(max, padding);
  const todayPercent = percentBetween(now, start, end);
  return { start, end, todayPercent };
}

export function buildScheduleTasks(tasks: LinkedWorkItem[], range: ScheduleRange): ScheduleTask[] {
  return tasks.map((item) => {
    const startAt = scheduleStart(item);
    const endAt = scheduleEnd(item, startAt);
    const leftPercent = percentBetween(startAt, range.start, range.end);
    const rightPercent = percentBetween(endAt, range.start, range.end);
    const widthPercent = Math.max(3, rightPercent - leftPercent);
    return { ...item, startAt, endAt, leftPercent: clamp(leftPercent), widthPercent: clamp(widthPercent), overdue: isOverdue(item.dueDate ?? item.plannedEndDate ?? item.plannedStartDate) };
  });
}

export function estimateDays(estimate: string | undefined): number {
  const value = (estimate ?? "").trim().toLowerCase();
  if (!value) return 1;
  const match = value.match(/^(\d+)([dhw])$/);
  if (!match) return 1;
  const amount = Number(match[1]);
  if (match[2] === "d") return amount;
  if (match[2] === "w") return amount * 5;
  return Math.max(1, Math.ceil(amount / 8));
}

export function formatDate(value?: string): string {
  const date = parseDate(value);
  return date ? date.toISOString().slice(0, 10) : "-";
}

export function parseDate(value?: string): Date | null {
  if (!value) return null;
  const parsed = new Date(`${value.slice(0, 10)}T00:00:00Z`);
  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

export function isOverdue(value?: string, now = new Date()): boolean {
  const date = parseDate(value);
  if (!date) return false;
  return date.getTime() < startOfDay(now).getTime();
}

export function isNearDue(value?: string, now = new Date()): boolean {
  const date = parseDate(value);
  if (!date) return false;
  const diff = date.getTime() - startOfDay(now).getTime();
  return diff >= 0 && diff <= 3 * 86400000;
}

export function statusBadgeClass(item: LinkedWorkItem): string {
  const status = taskDisplayStatus(item);
  if (status === "done") return "done";
  if (status === "in_progress") return "in-progress";
  if (status === "blocked") return "blocked";
  if (status === "overdue") return "overdue";
  if (status === "near_due") return "near-due";
  return "not-started";
}

function sortValue(item: LinkedWorkItem, field: TaskSortField): string {
  if (field === "title") return item.title;
  if (field === "priority") return String(priorityOrder[deriveTaskPriority(item)] ?? 99).padStart(2, "0");
  if (field === "createdAt") return item.audit?.createdAt ?? item.lastSyncedAt ?? "";
  if (field === "updatedAt") return item.audit?.updatedAt ?? item.lastSyncedAt ?? "";
  if (field === "dueDate") return item.dueDate ?? "9999-12-31";
  return item.title;
}

function groupValue(item: LinkedWorkItem, groupBy: TaskGrouping): string {
  if (groupBy === "project") return item.projectId || "unknown";
  if (groupBy === "status") return taskDisplayStatus(item);
  if (groupBy === "priority") return deriveTaskPriority(item);
  if (groupBy === "owner") return item.owner || "unknown";
  if (groupBy === "sourceType") return item.sourceType || "unknown";
  return "all";
}

function groupLabel(groupBy: TaskGrouping, key: string): string {
  if (groupBy === "priority") return key;
  if (groupBy === "status") return key;
  if (groupBy === "owner") return key;
  if (groupBy === "sourceType") return key;
  return key;
}

function scheduleStart(item: LinkedWorkItem): Date {
  return parseDate(item.plannedStartDate) ?? parseDate(item.dueDate) ?? parseDate(item.audit?.createdAt) ?? parseDate(item.lastSyncedAt) ?? new Date();
}

function scheduleEnd(item: LinkedWorkItem, startAt: Date): Date {
  return parseDate(item.plannedEndDate) ?? parseDate(item.dueDate) ?? shiftDays(startAt, Math.max(1, estimateDays(item.estimate)));
}

function shiftDays(date: Date, delta: number): Date {
  return new Date(date.getTime() + delta * 86400000);
}

function percentBetween(value: Date, start: Date, end: Date): number {
  const range = Math.max(1, end.getTime() - start.getTime());
  return ((value.getTime() - start.getTime()) / range) * 100;
}

function clamp(value: number): number {
  return Math.min(100, Math.max(0, value));
}

function startOfDay(date: Date): Date {
  return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate()));
}
