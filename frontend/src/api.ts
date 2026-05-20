import type { Locale } from "./i18n";

const API_BASE = "/api/v1";

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const resp = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json", "X-Role": "admin" },
    ...init,
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
  completedDate?: string;
  progressPercent: number;
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
  workItems: unknown[];
  updates: WeeklyUpdate[];
}

export interface MilestoneDetailView {
  milestone: Milestone;
  workItems: unknown[];
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
    summary: { "zh-CN": "摘要", "en-US": "Summary" },
    progress: { "zh-CN": "进展", "en-US": "Progress" },
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
  };
  return labels[key]?.[locale] ?? key;
}
