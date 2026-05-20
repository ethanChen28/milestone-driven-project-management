export type Locale = "zh-CN" | "en-US";

const messages = {
  "zh-CN": {
    title: "里程碑驱动项目管理",
    subtitle: "默认语言为中文，支持中英文切换。",
    backend: "后端：Golang",
    frontend: "前端：Vue 3",
    infra: "基础设施：MySQL + Redis",
    loading: "正在加载项目概览...",
    activeProjects: "活跃项目",
    blockedMilestones: "阻塞里程碑",
    overdueMilestones: "逾期里程碑",
    workload: "里程碑工作 / BAU 工作",
  },
  "en-US": {
    title: "Milestone-Driven Project Management",
    subtitle: "Chinese is the default locale with English support.",
    backend: "Backend: Golang",
    frontend: "Frontend: Vue 3",
    infra: "Infrastructure: MySQL + Redis",
    loading: "Loading portfolio summary...",
    activeProjects: "Active Projects",
    blockedMilestones: "Blocked Milestones",
    overdueMilestones: "Overdue Milestones",
    workload: "Milestone Work / BAU Work",
  },
} as const;

export function t(locale: Locale, key: keyof typeof messages["zh-CN"]): string {
  return messages[locale][key];
}
