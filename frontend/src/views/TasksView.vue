<script setup lang="ts">
import { computed, inject, onMounted, ref, watch, type Ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { Locale } from "../i18n";
import { apiFetch, can, label, type LinkedWorkItem, type Milestone, type Project, type WorkspaceRole } from "../api";
import {
  buildScheduleRange,
  buildScheduleTasks,
  computeMetrics,
  defaultTaskWorkspaceState,
  deriveTaskPriority,
  filterTasks,
  formatDate,
  groupTasks,
  normalizeWorkspaceState,
  serializeWorkspaceState,
  sortTasks,
  statusBadgeClass,
  taskDisplayStatus,
  taskTags,
  taskViews,
  type TaskGrouping,
  type TaskView,
} from "../task-workspace";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const currentUser = inject<Ref<string>>("currentUser")!;
const route = useRoute();
const router = useRouter();

const workspace = ref(defaultTaskWorkspaceState());
const tasks = ref<LinkedWorkItem[]>([]);
const projects = ref<Project[]>([]);
const milestones = ref<Milestone[]>([]);
const collapsedGroups = ref<string[]>([]);
const loading = ref(false);
const error = ref("");
const syncingRoute = ref(false);
const isMyTasks = ref(false);
const canManageTask = computed(() => can(currentRole.value, "manageWorkItem"));

const projectNames = computed(() => Object.fromEntries(projects.value.map((project) => [project.id, project.name])));
const milestoneNames = computed(() => Object.fromEntries(milestones.value.map((milestone) => [milestone.id, milestone.title])));

const owners = computed(() => unique(tasks.value.map((task) => task.owner).filter(Boolean)));
const statuses = computed(() => unique(tasks.value.map((task) => taskDisplayStatus(task))));
const sourceTypes = computed(() => unique(tasks.value.map((task) => task.sourceType)));
const priorities = computed(() => unique(tasks.value.map((task) => deriveTaskPriority(task))));
const tags = computed(() => unique(tasks.value.flatMap((task) => taskTags(task))));

const filteredTasks = computed(() => {
  let result = sortTasks(filterTasks(tasks.value, workspace.value), workspace.value);
  if (isMyTasks.value) result = result.filter((task) => task.owner === currentUser.value);
  return result;
});
const boardColumns = computed(() => {
  const statuses = ['not_started', 'in_progress', 'done', 'blocked', 'overdue'] as const;
  const map = new Map<string, LinkedWorkItem[]>();
  for (const status of statuses) map.set(status, []);
  for (const task of filteredTasks.value) {
    const status = taskDisplayStatus(task);
    const list = map.get(status);
    if (list) list.push(task);
  }
  return map;
});
const metrics = computed(() => computeMetrics(filteredTasks.value));
const groupedTasks = computed(() => groupTasks(filteredTasks.value, workspace.value.groupBy));
const scheduleRange = computed(() => buildScheduleRange(filteredTasks.value, workspace.value.scale));
const scheduleTasks = computed(() => buildScheduleTasks(filteredTasks.value, scheduleRange.value));
const ganttGroups = computed(() => groupTasks(filteredTasks.value, workspace.value.groupBy === "none" ? "project" : workspace.value.groupBy));
const timelineMilestones = computed(() => {
  const milestoneIds = new Set(filteredTasks.value.map((task) => task.milestoneId).filter(Boolean));
  return milestones.value
    .filter((milestone) => milestoneIds.has(milestone.id))
    .sort((left, right) => {
      const leftDate = left.plannedDate || left.forecastDate || left.completedDate || "";
      const rightDate = right.plannedDate || right.forecastDate || right.completedDate || "";
      return leftDate.localeCompare(rightDate);
    });
});
const projectDeadline = computed(() => {
  const projectIds = [...new Set(filteredTasks.value.map((task) => task.projectId).filter(Boolean))];
  const projectsWithDates = projects.value
    .filter((project) => projectIds.includes(project.id) && project.targetEndDate)
    .sort((left, right) => (left.targetEndDate ?? "").localeCompare(right.targetEndDate ?? ""));
  return projectsWithDates[0] ?? null;
});
const currentView = computed(() => workspace.value.view);

function unique(values: string[]) {
  return [...new Set(values.filter(Boolean))].sort((left, right) => left.localeCompare(right, "zh-Hans-CN"));
}

function applyRouteQuery(query: Record<string, unknown>) {
  workspace.value = normalizeWorkspaceState(query);
}

async function load() {
  loading.value = true;
  try {
    const [taskData, projectData, milestoneData] = await Promise.all([
      apiFetch<LinkedWorkItem[]>("/work-items"),
      apiFetch<Project[]>("/projects"),
      apiFetch<Milestone[]>("/milestones"),
    ]);
    tasks.value = taskData;
    projects.value = projectData;
    milestones.value = milestoneData;
    error.value = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
    tasks.value = [];
  } finally {
    loading.value = false;
  }
}

function syncRoute() {
  syncingRoute.value = true;
  const preservedQuery: Record<string, unknown> = {};
  Object.entries(route.query).forEach(([key, value]) => {
    if (key === "view" || key === "groupBy" || key === "sortBy" || key === "sortDir" || key === "scale" || key in workspace.value.filters) {
      return;
    }
    preservedQuery[key] = value;
  });
  router.replace({ query: { ...preservedQuery, ...serializeWorkspaceState(workspace.value) } }).finally(() => {
    syncingRoute.value = false;
  });
}

function setView(view: TaskView) {
  workspace.value.view = view;
}

function setGroupBy(groupBy: TaskGrouping) {
  workspace.value.groupBy = groupBy;
}

function changeGroupBy(event: Event) {
  setGroupBy((event.target as HTMLSelectElement).value as TaskGrouping);
}

function setFilter(key: keyof typeof workspace.value.filters, value: string) {
  workspace.value.filters[key] = value;
}

function clearFilters() {
  workspace.value.filters = defaultTaskWorkspaceState().filters;
  workspace.value.groupBy = "none";
  workspace.value.sortBy = "dueDate";
  workspace.value.sortDir = "asc";
  workspace.value.scale = "month";
  isMyTasks.value = false;
}

function toggleMyTasks() {
  isMyTasks.value = !isMyTasks.value;
}

function openTask(id: string) {
  router.push({ name: "task-detail", params: { id } });
}

function createTask() {
  router.push({ name: "task-create" });
}

function focusProjectDeadline(projectId: string) {
  setFilter("projectId", projectId);
  setView("list");
}

function metricFilter(metric: "completed" | "inProgress" | "notStarted" | "overdue" | "blocked" | "nearDue") {
  if (metric === "completed") workspace.value.filters.status = "done";
  else if (metric === "inProgress") workspace.value.filters.status = "in_progress";
  else if (metric === "notStarted") workspace.value.filters.status = "not_started";
  else if (metric === "overdue") workspace.value.filters.status = "overdue";
  else if (metric === "blocked") workspace.value.filters.status = "blocked";
  else if (metric === "nearDue") workspace.value.filters.status = "near_due";
  workspace.value.view = "list";
}

function onSortChange(value: string) {
  if (value === "dueDate" || value === "updatedAt" || value === "createdAt" || value === "title" || value === "priority") {
    workspace.value.sortBy = value;
  }
}

function changeSortBy(event: Event) {
  onSortChange((event.target as HTMLSelectElement).value);
}

function navigateMetric(metric: string) {
  metricFilter(metric as never);
}

function labelByProject(projectId: string) {
  return projectNames.value[projectId] ?? projectId;
}

function labelByMilestone(milestoneId: string) {
  return milestoneNames.value[milestoneId] ?? milestoneId;
}

watch(
  () => route.query,
  (query) => {
    if (syncingRoute.value) return;
    applyRouteQuery(query as Record<string, unknown>);
  },
  { immediate: true, deep: true },
);

watch(
  workspace,
  () => {
    if (!syncingRoute.value) syncRoute();
  },
  { deep: true },
);

onMounted(async () => {
  await load();
});

function taskSubtitle(task: LinkedWorkItem) {
  return [
    labelByProject(task.projectId),
    labelByMilestone(task.milestoneId),
    `${label("ownerTeam", locale)}: ${task.owner || "-"}`,
    `${label("priorityBucket", locale)}: ${deriveTaskPriority(task)}`,
    `${label("sourceType", locale)}: ${task.sourceType}`,
  ].join(" · ");
}

function taskRisk(task: LinkedWorkItem) {
  if (task.blocked) return label("blockedTask", locale);
  if (taskDisplayStatus(task) === "overdue") return label("overdueTask", locale);
  if (taskDisplayStatus(task) === "near_due") return label("nearDueTask", locale);
  return "";
}

function taskSource(task: LinkedWorkItem) {
  const gitlabMeta = task.sourceType === "gitlab_issue" ? `GitLab ${task.gitlabState ?? task.gitLabState ?? "-"}` : "";
  return gitlabMeta || task.sourceType;
}

function toggleCollapsed(key: string) {
  if (collapsedGroups.value.includes(key)) {
    collapsedGroups.value = collapsedGroups.value.filter((item) => item !== key);
    return;
  }
  collapsedGroups.value = [...collapsedGroups.value, key];
}

function isCollapsed(key: string) {
  return collapsedGroups.value.includes(key);
}
</script>

<template>
  <div class="page">
    <div class="header">
      <div>
        <h1>{{ label("taskWorkspace", locale) }}</h1>
        <p class="subtle">{{ label("taskWorkspaceSubtitle", locale) }}</p>
      </div>
      <div class="header-actions">
        <button class="btn" :class="{ active: isMyTasks }" @click="toggleMyTasks">{{ label("myTasks", locale) }}</button>
        <button v-if="canManageTask" class="btn primary" @click="createTask">{{ label("newTask", locale) }}</button>
        <span class="role-chip">{{ currentRole }}</span>
      </div>
    </div>

    <div class="tab-strip" role="tablist" aria-label="Task views">
      <button
        v-for="view in taskViews"
        :key="view"
        class="tab"
        :class="{ active: currentView === view }"
        role="tab"
        :aria-selected="currentView === view"
        :tabindex="currentView === view ? 0 : -1"
        @click="setView(view)"
      >
        {{ label(view === "list" ? "taskList" : view === "board" ? "taskBoard" : view === "gantt" ? "taskGantt" : view === "timeline" ? "taskTimeline" : view === "project" ? "taskByProject" : "taskByPriority", locale) }}
      </button>
    </div>

    <section class="summary-grid">
      <button class="summary-card total" @click="setView('list')">
        <span>{{ label("taskCount", locale) }}</span>
        <strong>{{ metrics.total }}</strong>
        <small>{{ label("viewAll", locale) }}</small>
      </button>
      <button class="summary-card" @click="navigateMetric('completed')">
        <span>{{ label("completedTask", locale) }}</span>
        <strong class="green">{{ metrics.completed }}</strong>
      </button>
      <button class="summary-card" @click="navigateMetric('inProgress')">
        <span>{{ label("inProgressTask", locale) }}</span>
        <strong class="blue">{{ metrics.inProgress }}</strong>
      </button>
      <button class="summary-card" @click="navigateMetric('notStarted')">
        <span>{{ label("notStartedTask", locale) }}</span>
        <strong class="gray">{{ metrics.notStarted }}</strong>
      </button>
      <button class="summary-card" @click="navigateMetric('overdue')">
        <span>{{ label("overdueTask", locale) }}</span>
        <strong class="red">{{ metrics.overdue }}</strong>
      </button>
      <button class="summary-card" @click="navigateMetric('blocked')">
        <span>{{ label("blockedTask", locale) }}</span>
        <strong class="amber">{{ metrics.blocked }}</strong>
      </button>
      <button class="summary-card" @click="navigateMetric('nearDue')">
        <span>{{ label("nearDueTask", locale) }}</span>
        <strong class="violet">{{ metrics.nearDue }}</strong>
      </button>
      <button v-if="projectDeadline" class="summary-card deadline" @click="focusProjectDeadline(projectDeadline.id)">
        <span>{{ label("projectDeadline", locale) }}</span>
        <strong>{{ formatDate(projectDeadline.targetEndDate) }}</strong>
        <small>{{ projectDeadline.name }}</small>
      </button>
    </section>

    <section class="filters">
      <div class="filters-head">
        <strong>{{ label("taskFilters", locale) }}</strong>
        <button class="clear-btn" @click="clearFilters">{{ label("clearFilters", locale) }}</button>
      </div>
      <div class="filters-grid">
        <select v-model="workspace.filters.projectId">
          <option value="">{{ label("project", locale) }}</option>
          <option v-for="project in projects" :key="project.id" :value="project.id">{{ project.name }}</option>
        </select>
        <select v-model="workspace.filters.milestoneId">
          <option value="">{{ label("milestone", locale) }}</option>
          <option v-for="milestone in milestones" :key="milestone.id" :value="milestone.id">{{ milestone.title }}</option>
        </select>
        <select v-model="workspace.filters.owner">
          <option value="">{{ label("ownerTeam", locale) }}</option>
          <option v-for="owner in owners" :key="owner" :value="owner">{{ owner }}</option>
        </select>
        <select v-model="workspace.filters.status">
          <option value="">{{ label("status", locale) }}</option>
          <option v-for="status in statuses" :key="status" :value="status">{{ status }}</option>
        </select>
        <select v-model="workspace.filters.sourceType">
          <option value="">{{ label("sourceType", locale) }}</option>
          <option v-for="sourceType in sourceTypes" :key="sourceType" :value="sourceType">{{ sourceType }}</option>
        </select>
        <select v-model="workspace.filters.priority">
          <option value="">{{ label("priorityBucket", locale) }}</option>
          <option v-for="priority in priorities" :key="priority" :value="priority">{{ priority }}</option>
        </select>
        <select v-model="workspace.filters.tag">
          <option value="">{{ label("labels", locale) }}</option>
          <option v-for="tag in tags" :key="tag" :value="tag">{{ tag }}</option>
        </select>
        <input v-model="workspace.filters.q" :placeholder="label('keyword', locale)" />
        <select data-testid="group-by-select" :value="workspace.groupBy" @change="changeGroupBy">
          <option v-for="option in ['none','project','status','priority','owner','sourceType']" :key="option" :value="option">{{ option }}</option>
        </select>
        <select data-testid="sort-by-select" :value="workspace.sortBy" @change="changeSortBy">
          <option value="dueDate">{{ label("plannedDate", locale) }}</option>
          <option value="updatedAt">{{ label("lastSynced", locale) }}</option>
          <option value="createdAt">{{ label("startDate", locale) }}</option>
          <option value="title">{{ label("title", locale) }}</option>
          <option value="priority">{{ label("priorityBucket", locale) }}</option>
        </select>
        <select v-model="workspace.sortDir">
          <option value="asc">asc</option>
          <option value="desc">desc</option>
        </select>
        <select v-model="workspace.scale">
          <option value="week">{{ label("week", locale) }}</option>
          <option value="month">{{ label("period", locale) }}</option>
          <option value="quarter">quarter</option>
          <option value="year">year</option>
        </select>
      </div>
    </section>

    <p v-if="loading" class="empty" aria-live="polite">Loading...</p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <section v-if="currentView === 'list'" class="panel">
      <div class="panel-head">
        <h2>{{ label("taskList", locale) }}</h2>
        <span class="count">{{ filteredTasks.length }}</span>
      </div>
      <table class="task-table">
        <thead>
          <tr>
            <th>{{ label("title", locale) }}</th>
            <th>{{ label("project", locale) }}</th>
            <th>{{ label("milestone", locale) }}</th>
            <th>{{ label("ownerTeam", locale) }}</th>
            <th>{{ label("status", locale) }}</th>
            <th>{{ label("priorityBucket", locale) }}</th>
            <th>{{ label("source", locale) }}</th>
            <th>{{ label("plannedDate", locale) }}</th>
            <th>{{ label("edit", locale) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="task in filteredTasks" :key="task.id" :data-task-id="task.id" class="task-row" tabindex="0" role="link" @click="openTask(task.id)" @keyup.enter="openTask(task.id)">
            <td>
              <strong class="clickable" @click="openTask(task.id)">{{ task.title }}</strong>
              <div class="hint">{{ taskSubtitle(task) }}</div>
              <div v-if="taskRisk(task)" class="risk">{{ taskRisk(task) }}</div>
            </td>
            <td>{{ labelByProject(task.projectId) }}</td>
            <td>{{ labelByMilestone(task.milestoneId) }}</td>
            <td>{{ task.owner || "-" }}</td>
            <td><span class="status-pill" :class="statusBadgeClass(task)">{{ taskDisplayStatus(task) }}</span></td>
            <td>{{ deriveTaskPriority(task) }}</td>
            <td>{{ taskSource(task) }}</td>
            <td>{{ formatDate(task.dueDate ?? task.plannedEndDate ?? task.plannedStartDate) }}</td>
            <td><button class="btn sm" @click.stop="openTask(task.id)">{{ label("edit", locale) }}</button></td>
          </tr>
        </tbody>
      </table>
      <p v-if="!filteredTasks.length" class="empty">{{ label("noData", locale) }}</p>
    </section>

    <section v-else-if="currentView === 'board'" class="board">
      <div v-for="status in ['not_started', 'in_progress', 'done', 'blocked', 'overdue']" :key="status" class="board-column">
        <div class="panel-head">
          <h2>{{ status }}</h2>
          <span class="count">{{ boardColumns.get(status)?.length ?? 0 }}</span>
        </div>
        <div class="stack">
          <article
            v-for="task in boardColumns.get(status) ?? []"
            :key="task.id"
            class="task-card"
            :class="statusBadgeClass(task)"
            :data-task-id="task.id"
          >
            <strong>{{ task.title }}</strong>
            <p>{{ taskSubtitle(task) }}</p>
            <small>{{ taskRisk(task) }}</small>
          </article>
          <p v-if="!boardColumns.get(status)?.length" class="empty">{{ label("noData", locale) }}</p>
        </div>
      </div>
    </section>

    <section v-else-if="currentView === 'gantt'" class="panel">
      <div class="panel-head">
        <h2>{{ label("taskGantt", locale) }}</h2>
        <span class="count">{{ scheduleTasks.length }}</span>
      </div>
      <div class="gantt-legend">
        <span><i class="today-line"></i>{{ label("currentDay", locale) }}</span>
        <span>■ overdue</span>
        <span>■ blocked</span>
      </div>
      <div class="gantt-chart">
        <div class="gantt-axis">
          <div>{{ formatDate(scheduleRange.start.toISOString()) }}</div>
          <div>{{ label("currentDay", locale) }}</div>
          <div>{{ formatDate(scheduleRange.end.toISOString()) }}</div>
        </div>
        <div class="gantt-grid">
          <div class="gantt-today" :style="{ left: `${scheduleRange.todayPercent}%` }"></div>
          <div v-for="group in ganttGroups" :key="group.key" class="gantt-group">
            <button class="gantt-group-head" @click="toggleCollapsed(group.key)">
              <strong>{{ workspace.groupBy === "project" ? labelByProject(group.key) : group.key === "all" ? label("taskList", locale) : group.key }}</strong>
              <span>{{ group.items.length }}</span>
              <span>{{ isCollapsed(group.key) ? "+" : "−" }}</span>
            </button>
            <div v-if="!isCollapsed(group.key)" class="gantt-rows">
              <div v-for="task in scheduleTasks.filter((row) => group.items.some((item) => item.id === row.id))" :key="task.id" class="gantt-row" :data-task-id="task.id">
                <div class="gantt-label">
                  <strong>{{ task.title }}</strong>
                  <small>{{ taskSubtitle(task) }}</small>
                </div>
                <div class="gantt-track">
                  <div
                    class="gantt-bar"
                    :class="{ overdue: task.overdue, blocked: task.blocked }"
                    :style="{ left: `${task.leftPercent}%`, width: `${task.widthPercent}%` }"
                  >
                    {{ formatDate(task.startAt.toISOString()) }} → {{ formatDate(task.endAt.toISOString()) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
          <p v-if="!scheduleTasks.length" class="empty">{{ label("noData", locale) }}</p>
        </div>
      </div>
    </section>

    <section v-else-if="currentView === 'timeline'" class="panel">
      <div class="panel-head">
        <h2>{{ label("taskTimeline", locale) }}</h2>
        <span class="count">{{ filteredTasks.length }}</span>
      </div>
      <div v-if="timelineMilestones.length" class="milestone-strip">
        <div v-for="milestone in timelineMilestones" :key="milestone.id" class="milestone-chip">
          <strong>{{ milestone.title }}</strong>
          <small>{{ formatDate(milestone.plannedDate ?? milestone.forecastDate ?? milestone.completedDate) }}</small>
        </div>
      </div>
      <div v-for="task in filteredTasks" :key="task.id" class="timeline-row" :data-task-id="task.id">
        <div class="timeline-date">
          <strong>{{ formatDate(task.dueDate ?? task.plannedEndDate ?? task.plannedStartDate) }}</strong>
          <small>{{ taskDisplayStatus(task) === "overdue" ? label("overdueTask", locale) : taskRisk(task) || taskDisplayStatus(task) }}</small>
        </div>
        <div class="timeline-card" :class="statusBadgeClass(task)">
          <strong>{{ task.title }}</strong>
          <p>{{ taskSubtitle(task) }}</p>
        </div>
      </div>
      <p v-if="!filteredTasks.length" class="empty">{{ label("noData", locale) }}</p>
    </section>

    <section v-else class="grouped-grid">
      <div v-for="group in groupedTasks" :key="group.key" class="panel grouped-panel">
        <div class="panel-head">
          <h2>{{ workspace.groupBy === "project" ? labelByProject(group.key) : group.label }}</h2>
          <span class="count">{{ group.items.length }}</span>
        </div>
        <div class="stack">
          <article v-for="task in group.items" :key="task.id" class="task-card" :class="statusBadgeClass(task)" :data-task-id="task.id">
            <strong>{{ task.title }}</strong>
            <p>{{ taskSubtitle(task) }}</p>
          </article>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.page { max-width: 1400px; }
.header { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
.header-actions { display: flex; align-items: center; gap: 10px; }
h1 { margin: 0; font-size: 2rem; }
.subtle { margin: 4px 0 0; color: var(--color-text-subtle); }
.role-chip { padding: 6px 12px; border-radius: var(--radius-full); background: #d9f5ea; color: #0f5132; font-size: .82rem; font-weight: 700; }
.tab-strip { display: flex; flex-wrap: wrap; gap: 10px; margin: 18px 0; }
.tab { border: 1px solid #d7dfdc; background: var(--color-surface); color: #335247; padding: 10px 14px; border-radius: var(--radius-full); cursor: pointer; }
.tab.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.btn.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 12px; margin-bottom: 18px; }
.summary-card { text-align: left; border: 0; border-radius: var(--radius-xl); padding: 16px; background: linear-gradient(180deg, #ffffff, #f4faf7); box-shadow: var(--shadow-sm); cursor: pointer; }
.summary-card.total { background: linear-gradient(180deg, #f1f8ff, #e0efff); }
.summary-card.deadline { background: linear-gradient(180deg, #fff7ed, #ffedd5); }
.summary-card span { display: block; color: #5e7a71; font-size: .85rem; }
.summary-card strong { display: block; font-size: 2rem; margin-top: 6px; }
.summary-card small { color: #5e7a71; }
.green { color: #16a34a; }
.blue { color: var(--color-info); }
.gray { color: #64748b; }
.red { color: #dc2626; }
.amber { color: #d97706; }
.violet { color: #7c3aed; }
.filters, .panel, .board-column, .grouped-panel { background: var(--color-surface); border-radius: 18px; box-shadow: var(--shadow-sm); }
.filters { padding: 16px; margin-bottom: 18px; }
.filters-head, .panel-head { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.filters-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 10px; margin-top: 12px; }
.filters-grid input, .filters-grid select { min-width: 0; border: 1px solid #d7dfdc; border-radius: var(--radius-md); padding: 10px 12px; }
.clear-btn { border: 0; background: none; color: var(--color-primary-light); cursor: pointer; }
.panel { padding: 16px; margin-bottom: 18px; }
.count { display: inline-flex; align-items: center; justify-content: center; min-width: 2rem; height: 2rem; border-radius: var(--radius-full); background: #edf7f2; color: var(--color-primary-light); font-weight: 700; }
.btn.sm { padding: 6px 12px; font-size: .82rem; }
.task-table { width: 100%; border-collapse: collapse; margin-top: 12px; }
.task-table th, .task-table td { text-align: left; padding: 12px 10px; vertical-align: top; border-bottom: 1px solid #edf1ef; }
.task-row { cursor: pointer; }
.hint { margin-top: 4px; color: var(--color-text-subtle); font-size: .84rem; }
.risk { margin-top: 4px; color: var(--color-warning); font-size: .8rem; font-weight: 700; }
.status-pill { display: inline-flex; padding: 4px 10px; border-radius: var(--radius-full); font-size: .76rem; font-weight: 700; }
.status-pill.done, .status-pill.in-progress { background: var(--color-success-bg); color: #166534; }
.status-pill.blocked, .status-pill.overdue { background: var(--color-danger-bg); color: var(--color-danger); }
.status-pill.near-due { background: var(--color-warning-bg); color: #92400e; }
.status-pill.not-started { background: #e5e7eb; color: #374151; }
.board { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 12px; align-items: start; }
.board-column { padding: 14px; }
.stack { display: grid; gap: 10px; margin-top: 12px; }
.task-card { padding: 12px; border-radius: var(--radius-lg); background: var(--color-surface-hover); border: 1px solid #e6f0ea; }
.task-card p { margin: 6px 0 0; color: #5e7a71; font-size: .84rem; }
.task-card small { display: block; margin-top: 6px; color: var(--color-warning); font-weight: 700; }
.task-card.done { border-color: #86efac; }
.task-card.in-progress { border-color: #93c5fd; }
.task-card.blocked, .task-card.overdue { border-color: #fca5a5; background: #fff5f5; }
.task-card.near-due { border-color: #fcd34d; }
.gantt-legend { display: flex; gap: 16px; align-items: center; color: #5e7a71; margin: 8px 0 16px; font-size: .85rem; }
.today-line { width: 16px; height: 2px; display: inline-block; background: var(--color-info); vertical-align: middle; margin-right: 6px; }
.gantt-chart { position: relative; }
.gantt-axis { display: grid; grid-template-columns: 1fr auto 1fr; color: #5e7a71; margin-bottom: 10px; }
.gantt-grid { position: relative; display: grid; gap: 10px; }
.gantt-today { position: absolute; top: 0; bottom: 0; width: 2px; background: var(--color-info); opacity: .75; }
.gantt-group { display: grid; gap: 10px; }
.gantt-group-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; border: 1px solid #e6f0ea; background: var(--color-surface-hover); border-radius: 12px; padding: 8px 12px; cursor: pointer; }
.gantt-rows { display: grid; gap: 10px; }
.gantt-row { display: grid; grid-template-columns: 280px 1fr; gap: 12px; align-items: center; }
.gantt-label small { display: block; color: var(--color-text-subtle); margin-top: 4px; }
.gantt-track { position: relative; min-height: 42px; border-radius: 12px; background: linear-gradient(90deg, #f2f7f4 0%, var(--color-surface-hover) 100%); overflow: hidden; }
.gantt-bar { position: absolute; top: 6px; height: 30px; border-radius: var(--radius-md); background: var(--color-info); color: #fff; font-size: .75rem; padding: 7px 10px; box-sizing: border-box; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.gantt-bar.overdue { background: #dc2626; }
.gantt-bar.blocked { background: #d97706; }
.milestone-strip { display: flex; gap: 10px; flex-wrap: wrap; margin: 12px 0 16px; }
.milestone-chip { padding: 10px 12px; border-radius: 12px; border: 1px solid #dfe8e3; background: var(--color-surface-hover); }
.milestone-chip strong { display: block; }
.milestone-chip small { display: block; margin-top: 4px; color: var(--color-text-subtle); }
.timeline-row { display: grid; grid-template-columns: 140px 1fr; gap: 14px; margin-top: 12px; }
.timeline-date { padding-top: 10px; color: #334155; }
.timeline-date strong { display: block; }
.timeline-date small { color: var(--color-text-subtle); }
.timeline-card { padding: 14px; border-radius: var(--radius-lg); border: 1px solid #e6f0ea; background: var(--color-surface-hover); }
.timeline-card p { margin: 6px 0 0; color: #5e7a71; }
.grouped-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.grouped-panel { padding: 14px; }
@media (max-width: 1200px) {
  .filters-grid, .board, .grouped-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .gantt-row, .timeline-row { grid-template-columns: 1fr; }
}
@media (max-width: 720px) {
  .filters-grid, .board, .grouped-grid { grid-template-columns: 1fr; }
  .header { flex-direction: column; align-items: flex-start; }
}
</style>
