import type { Locale } from "./i18n";

const API_BASE = "/api/v1";
export const ROLE_STORAGE_KEY = "goal-manager.workspaceRole";

export const workspaceRoles = ["admin", "portfolio_manager", "project_owner", "contributor", "viewer"] as const;
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

export function can(role: WorkspaceRole, permission: string): boolean {
  return rolePermissions[role].includes(permission);
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  if (!headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  headers.set("X-Role", getCurrentRole());
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
  estimate: string;
  dueDate?: string;
  blocked: boolean;
  gitLabLabels?: string[];
  gitlabLabels?: string[];
  gitLabAssignee?: string;
  gitlabAssignee?: string;
  gitLabState?: string;
  gitlabState?: string;
  lastSyncedAt?: string;
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
  };
  return labels[key]?.[locale] ?? key;
}
