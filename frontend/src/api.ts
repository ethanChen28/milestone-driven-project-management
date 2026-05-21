import type { Locale } from "./i18n";

const API_BASE = "/api/v1";
export const ROLE_STORAGE_KEY = "goal-manager.workspaceRole";
export const USER_STORAGE_KEY = "goal-manager.currentUser";

export const workspaceRoles = ["admin", "portfolio_manager", "project_owner", "contributor", "viewer"] as const;
export const workspaceUsers = ["tester", "alice", "bob", "carol", "frontend-user"] as const;
export type WorkspaceRole = (typeof workspaceRoles)[number];

const rolePermissions: Record<WorkspaceRole, string[]> = {
  admin: ["manageIntegration", "manageRoadmap", "manageProject", "manageMilestone", "manageWorkItem", "submitUpdate"],
  portfolio_manager: ["manageRoadmap", "manageProject", "manageMilestone", "manageWorkItem", "submitUpdate"],
  project_owner: ["manageProject", "manageMilestone", "manageWorkItem", "submitUpdate"],
  contributor: ["manageWorkItem", "submitUpdate"],
  viewer: [],
};

export function isWorkspaceRole(value: string | null): value is WorkspaceRole {
  return !!value && workspaceRoles.includes(value as WorkspaceRole);
}

export function getCurrentRole(): WorkspaceRole {
  if (typeof window === "undefined") return "contributor";
  const stored = window.localStorage.getItem(ROLE_STORAGE_KEY);
  return isWorkspaceRole(stored) ? stored : "contributor";
}

export function setCurrentRole(role: WorkspaceRole) {
  if (typeof window !== "undefined") window.localStorage.setItem(ROLE_STORAGE_KEY, role);
}

export function getCurrentUser(): string {
  if (typeof window === "undefined") return "tester";
  return window.localStorage.getItem(USER_STORAGE_KEY) || "tester";
}

export function setCurrentUser(user: string) {
  if (typeof window !== "undefined") window.localStorage.setItem(USER_STORAGE_KEY, user);
}

export function dateInputToIso(value?: string): string | undefined {
  if (!value) return undefined;
  if (value.includes("T")) return value;
  return new Date(`${value}T00:00:00Z`).toISOString();
}

export function isoToDateInput(value?: string): string {
  return value?.slice(0, 10) ?? "";
}

export function can(role: WorkspaceRole, permission: string): boolean {
  return rolePermissions[role].includes(permission);
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  if (!headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  headers.set("X-Role", getCurrentRole());
  headers.set("X-User", getCurrentUser());
  const resp = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
  });
  if (!resp.ok) {
    const body = await resp.text();
    throw new Error(`${resp.status}: ${body}`);
  }
  return resp.json();
}

export interface PortfolioSummary {
  activeProjects: number;
  blockedMilestones: number;
  overdueMilestones: number;
  milestoneWorkItems: number;
  bauWorkItems: number;
  healthDistribution: Record<string, number>;
}

export interface Project {
  id: string;
  name: string;
  objective: string;
  owner: string;
  participants?: string[];
  status: string;
  healthStatus: string;
  priority: string;
  projectType: string;
  targetStartDate?: string;
  targetEndDate?: string;
}

export interface Milestone {
  id: string;
  projectId: string;
  title: string;
  status: string;
  healthStatus: string;
  owner: string;
  completionCriteria: string;
  plannedDate?: string;
  forecastDate?: string;
  completedDate?: string;
  progressPercent: number;
  riskLevel: string;
  dependencySummary: string;
}

export interface LinkedWorkItem {
  id: string;
  sourceType: string;
  sourceId: string;
  sourceUrl: string;
  title: string;
  projectId: string;
  milestoneId: string;
  workstreamId: string;
  owner: string;
  status: string;
  priority?: string;
  tags?: string[];
  estimate: string;
  plannedStartDate?: string;
  plannedEndDate?: string;
  dueDate?: string;
  blocked: boolean;
  gitLabLabels?: string[];
  gitlabLabels?: string[];
  gitLabAssignee?: string;
  gitlabAssignee?: string;
  gitLabState?: string;
  gitlabState?: string;
  lastSyncedAt?: string;
  audit?: {
    createdAt?: string;
    updatedAt?: string;
  };
}

export function gitlabLabels(item: LinkedWorkItem): string[] {
  return item.gitlabLabels ?? item.gitLabLabels ?? [];
}

export function gitlabAssignee(item: LinkedWorkItem): string {
  return item.gitlabAssignee ?? item.gitLabAssignee ?? "";
}

export function gitlabState(item: LinkedWorkItem): string {
  return item.gitlabState ?? item.gitLabState ?? "";
}

export interface WeeklyUpdate {
  id: string;
  projectId: string;
  milestoneId: string;
  summary: string;
  progress: string;
  risk: string;
  blockers: string;
  decisionsNeeded: string;
  nextSteps: string;
  author: string;
  week: string;
}

export interface RoadmapPeriod {
  id: string;
  title: string;
  description: string;
  status: string;
  owner: string;
  periodStart: string;
  periodEnd: string;
}

export interface RoadmapItem {
  id: string;
  periodId: string;
  title: string;
  status: string;
  owner: string;
  priority: string;
}

export interface WeeklyReviewView {
  updates: WeeklyUpdate[];
  delayedMilestones: Milestone[];
  blockedMilestones: Milestone[];
}

export interface ProjectDetailView {
  project: Project;
  milestones: Milestone[];
  workItems: LinkedWorkItem[];
  updates: WeeklyUpdate[];
}

export interface ProjectSpaceRollups {
  milestoneStatusCounts: Record<string, number>;
  workItemStatusCounts: Record<string, number>;
  activeMilestones: number;
  completedMilestones: number;
  blockedMilestones: number;
  overdueMilestones: number;
  blockedWorkItems: number;
  overdueWorkItems: number;
  externalDependencies: number;
  recentUpdateCount: number;
}

export interface ProjectRiskSignal {
  id: string;
  sourceType: string;
  sourceId: string;
  title: string;
  severity: string;
  message: string;
  owner: string;
  status: string;
  milestoneId?: string;
  workItemId?: string;
  updateId?: string;
}

export interface ProjectDependency {
  id: string;
  sourceType: string;
  sourceId: string;
  title: string;
  message: string;
  owner: string;
  status: string;
  milestoneId?: string;
  workItemId?: string;
}

export interface ProjectSpaceView extends ProjectDetailView {
  rollups: ProjectSpaceRollups;
  risks: ProjectRiskSignal[];
  dependencies: ProjectDependency[];
}

export interface MilestoneDetailView {
  milestone: Milestone;
  workItems: LinkedWorkItem[];
  updates: WeeklyUpdate[];
}

export interface RoadmapOverviewItem {
  period: RoadmapPeriod;
  items: RoadmapItem[];
  projectSummaries: { id: string; name: string; healthStatus: string; progressPercent: number; milestones: number }[];
}

export function label(key: string, locale: Locale): string {
  const labels: Record<string, Record<Locale, string>> = {
    dashboard: { "zh-CN": "仪表盘", "en-US": "Dashboard" },
    projects: { "zh-CN": "项目", "en-US": "Projects" },
    milestones: { "zh-CN": "里程碑", "en-US": "Milestones" },
    roadmap: { "zh-CN": "路线图", "en-US": "Roadmap" },
    review: { "zh-CN": "周度回顾", "en-US": "Weekly Review" },
    createProject: { "zh-CN": "创建项目", "en-US": "Create Project" },
    createMilestone: { "zh-CN": "创建里程碑", "en-US": "Create Milestone" },
    submitUpdate: { "zh-CN": "提交周报", "en-US": "Submit Update" },
    name: { "zh-CN": "名称", "en-US": "Name" },
    objective: { "zh-CN": "目标", "en-US": "Objective" },
    owner: { "zh-CN": "负责人", "en-US": "Owner" },
    status: { "zh-CN": "状态", "en-US": "Status" },
    health: { "zh-CN": "健康度", "en-US": "Health" },
    priority: { "zh-CN": "优先级", "en-US": "Priority" },
    title: { "zh-CN": "标题", "en-US": "Title" },
    criteria: { "zh-CN": "完成标准", "en-US": "Completion Criteria" },
    plannedDate: { "zh-CN": "计划日期", "en-US": "Planned Date" },
    forecastDate: { "zh-CN": "预测日期", "en-US": "Forecast Date" },
    dependencySummary: { "zh-CN": "依赖说明", "en-US": "Dependency Summary" },
    summary: { "zh-CN": "摘要", "en-US": "Summary" },
    progress: { "zh-CN": "进展", "en-US": "Progress" },
    progressPercent: { "zh-CN": "进度百分比", "en-US": "Progress Percent" },
    risk: { "zh-CN": "风险", "en-US": "Risk" },
    blockers: { "zh-CN": "阻塞项", "en-US": "Blockers" },
    decisionsNeeded: { "zh-CN": "需要决策", "en-US": "Decisions Needed" },
    nextSteps: { "zh-CN": "下一步", "en-US": "Next Steps" },
    week: { "zh-CN": "周", "en-US": "Week" },
    author: { "zh-CN": "作者", "en-US": "Author" },
    save: { "zh-CN": "保存", "en-US": "Save" },
    cancel: { "zh-CN": "取消", "en-US": "Cancel" },
    delayed: { "zh-CN": "延期里程碑", "en-US": "Delayed Milestones" },
    blocked: { "zh-CN": "阻塞里程碑", "en-US": "Blocked Milestones" },
    noData: { "zh-CN": "暂无数据", "en-US": "No data" },
    projectType: { "zh-CN": "项目类型", "en-US": "Project Type" },
    startDate: { "zh-CN": "开始日期", "en-US": "Start Date" },
    endDate: { "zh-CN": "结束日期", "en-US": "End Date" },
    edit: { "zh-CN": "编辑", "en-US": "Edit" },
    period: { "zh-CN": "周期", "en-US": "Period" },
    roleTool: { "zh-CN": "MVP 角色调试", "en-US": "MVP Role Debug" },
    userTool: { "zh-CN": "当前用户", "en-US": "Current User" },
    roleWarning: { "zh-CN": "非生产登录", "en-US": "Not production auth" },
    noPermission: { "zh-CN": "当前角色无权限", "en-US": "Current role cannot edit" },
    filters: { "zh-CN": "筛选", "en-US": "Filters" },
    clearFilters: { "zh-CN": "清除筛选", "en-US": "Clear Filters" },
    workItems: { "zh-CN": "工作项", "en-US": "Work Items" },
    source: { "zh-CN": "来源", "en-US": "Source" },
    assignee: { "zh-CN": "处理人", "en-US": "Assignee" },
    labels: { "zh-CN": "标签", "en-US": "Labels" },
    lastSynced: { "zh-CN": "最近同步", "en-US": "Last Synced" },
    openIssue: { "zh-CN": "打开 Issue", "en-US": "Open Issue" },
    gitlabContext: { "zh-CN": "GitLab 上下文", "en-US": "GitLab Context" },
    tasks: { "zh-CN": "任务", "en-US": "Tasks" },
    taskWorkspace: { "zh-CN": "任务工作台", "en-US": "Task Workspace" },
    taskWorkspaceSubtitle: { "zh-CN": "共享筛选、看板、甘特图、时间线和分组视图", "en-US": "Shared filters, board, Gantt, timeline and grouped views" },
    taskList: { "zh-CN": "任务列表", "en-US": "Task List" },
    taskBoard: { "zh-CN": "状态看板", "en-US": "Status Board" },
    taskGantt: { "zh-CN": "进展甘特图", "en-US": "Gantt View" },
    taskTimeline: { "zh-CN": "时间线", "en-US": "Timeline" },
    taskByProject: { "zh-CN": "按项目查看", "en-US": "By Project" },
    taskByPriority: { "zh-CN": "按优先级查看", "en-US": "By Priority" },
    taskMetrics: { "zh-CN": "关键指标", "en-US": "Key Metrics" },
    taskFilters: { "zh-CN": "任务筛选", "en-US": "Task Filters" },
    newTask: { "zh-CN": "新建任务", "en-US": "New Task" },
    editTask: { "zh-CN": "编辑任务", "en-US": "Edit Task" },
    taskDetail: { "zh-CN": "任务详情", "en-US": "Task Detail" },
    deleteTask: { "zh-CN": "删除任务", "en-US": "Delete Task" },
    saveTask: { "zh-CN": "保存任务", "en-US": "Save Task" },
    taskCreateSuccess: { "zh-CN": "任务已创建", "en-US": "Task created" },
    taskUpdateSuccess: { "zh-CN": "任务已更新", "en-US": "Task updated" },
    taskDeleteConfirm: { "zh-CN": "确认删除该任务？", "en-US": "Delete this task?" },
    taskDeleteSuccess: { "zh-CN": "任务已删除", "en-US": "Task deleted" },
    project: { "zh-CN": "项目", "en-US": "Project" },
    milestone: { "zh-CN": "里程碑", "en-US": "Milestone" },
    sourceType: { "zh-CN": "来源类型", "en-US": "Source Type" },
    sourceId: { "zh-CN": "来源 ID", "en-US": "Source ID" },
    sourceUrl: { "zh-CN": "来源链接", "en-US": "Source URL" },
    workstream: { "zh-CN": "工作流", "en-US": "Workstream" },
    groupBy: { "zh-CN": "分组", "en-US": "Group By" },
    sortBy: { "zh-CN": "排序", "en-US": "Sort By" },
    scale: { "zh-CN": "时间尺度", "en-US": "Scale" },
    keyword: { "zh-CN": "关键词", "en-US": "Keyword" },
    tags: { "zh-CN": "标签", "en-US": "Tags" },
    estimate: { "zh-CN": "预估工作量", "en-US": "Estimate" },
    blockedFlag: { "zh-CN": "阻塞", "en-US": "Blocked" },
    dueDate: { "zh-CN": "截止日期", "en-US": "Due Date" },
    priorityBucket: { "zh-CN": "优先级", "en-US": "Priority" },
    ownerTeam: { "zh-CN": "负责人", "en-US": "Owner" },
    blockedTask: { "zh-CN": "阻塞任务", "en-US": "Blocked Task" },
    overdueTask: { "zh-CN": "逾期任务", "en-US": "Overdue Task" },
    nearDueTask: { "zh-CN": "临近截止", "en-US": "Near Due" },
    completedTask: { "zh-CN": "已完成", "en-US": "Completed" },
    inProgressTask: { "zh-CN": "进行中", "en-US": "In Progress" },
    notStartedTask: { "zh-CN": "未开始", "en-US": "Not Started" },
    taskCount: { "zh-CN": "任务总数", "en-US": "Total Tasks" },
    viewAll: { "zh-CN": "查看全部", "en-US": "View All" },
    currentDay: { "zh-CN": "今天", "en-US": "Today" },
    projectDeadline: { "zh-CN": "项目截止日期", "en-US": "Project Deadline" },
    milestoneTimeline: { "zh-CN": "里程碑日期", "en-US": "Milestone Dates" },
  };
  return labels[key]?.[locale] ?? key;
}
