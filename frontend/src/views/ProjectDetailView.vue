<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { Locale } from "../i18n";
import { dateInputToIso, label, apiFetch, can, gitlabAssignee, gitlabLabels, gitlabState, type LinkedWorkItem, type Milestone, type ProjectDependency, type ProjectRiskSignal, type ProjectSpaceView, type ProjectSpaceRollups, type WorkspaceRole } from "../api";
import { displayProjectWorkStatus, filterProjectWorkItems, groupProjectWorkItems, normalizeProjectSpaceTab, projectSpaceTabs, projectWorkBreadcrumb, type ProjectWorkGroup } from "../project-space";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const route = useRoute();
const router = useRouter();
const id = route.params.id as string;
const detail = ref<ProjectSpaceView | null>(null);
const showMilestoneForm = ref(false);
const error = ref("");
const msForm = ref({ title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "", riskLevel: "low" });
const canManageProject = computed(() => can(currentRole.value, "manageProject"));
const canManageMilestone = computed(() => can(currentRole.value, "manageMilestone"));
const activeTab = computed(() => normalizeProjectSpaceTab(route.query.tab));
const workGroupBy = computed(() => normalizeGroup(route.query.groupBy));
const workFilters = computed(() => ({
  milestoneId: queryValue(route.query.milestoneId),
  status: queryValue(route.query.status),
  priority: queryValue(route.query.priority),
  owner: queryValue(route.query.owner),
  sourceType: queryValue(route.query.sourceType),
  blocked: queryValue(route.query.blocked),
  overdue: queryValue(route.query.overdue),
}));
const filteredWorkItems = computed(() => detail.value ? filterProjectWorkItems(detail.value.workItems, workFilters.value) : []);
const groupedWorkItems = computed(() => detail.value ? groupProjectWorkItems(filteredWorkItems.value, workGroupBy.value, detail.value.milestones) : []);
const recentUpdates = computed(() => detail.value?.updates.slice(0, 3) ?? []);
const topRisks = computed(() => detail.value?.risks.slice(0, 4) ?? []);
const topDependencies = computed(() => detail.value?.dependencies.slice(0, 4) ?? []);

const tabLabels: Record<string, { zh: string; en: string }> = {
  overview: { zh: "概览", en: "Overview" },
  "work-items": { zh: "工作项", en: "Work Items" },
  milestones: { zh: "里程碑", en: "Milestones" },
  updates: { zh: "周报", en: "Updates" },
  risks: { zh: "风险", en: "Risks" },
  dependencies: { zh: "依赖", en: "Dependencies" },
  settings: { zh: "设置", en: "Settings" },
};

async function load() {
  try { detail.value = await apiFetch<ProjectSpaceView>(`/project-space?id=${id}`); error.value = ""; } catch (err) { error.value = err instanceof Error ? err.message : String(err); }
}

onMounted(load);

async function createMilestone() {
  if (!canManageMilestone.value) return;
  await apiFetch("/milestones", { method: "POST", body: JSON.stringify({ ...msForm.value, projectId: id, plannedDate: dateInputToIso(msForm.value.plannedDate) }) });
  showMilestoneForm.value = false;
  msForm.value = { title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "", riskLevel: "low" };
  await load();
}

async function updateHealth(h: string) {
  if (!detail.value || !canManageProject.value) return;
  await apiFetch(`/projects?id=${id}`, { method: "PUT", body: JSON.stringify({ ...detail.value.project, healthStatus: h }) });
  await load();
}

async function transitionMilestone(m: Milestone, status: string) {
  if (!canManageMilestone.value) return;
  try {
    await apiFetch(`/milestones?id=${m.id}`, { method: "PUT", body: JSON.stringify({ ...m, status }) });
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}

function setTab(tab: string, extra: Record<string, string | undefined> = {}) {
  router.replace({ query: cleanQuery({ ...route.query, tab, ...extra }) });
}

function updateWorkQuery(next: Record<string, string>) {
  router.replace({ query: cleanQuery({ ...route.query, tab: "work-items", ...next }) });
}

function openCreateMilestone() {
  showMilestoneForm.value = true;
  setTab("milestones");
}

function goMilestone(id: string) { router.push({ name: "milestone-detail", params: { id } }); }
function openMilestoneWork(milestoneId: string) { setTab("work-items", { milestoneId, groupBy: "status" }); }
function formatDate(value?: string) { return value?.slice(0, 10) || "-"; }
function tabText(tab: string) { return locale.value === "zh-CN" ? tabLabels[tab].zh : tabLabels[tab].en; }
function copy(zh: string, en: string) { return locale.value === "zh-CN" ? zh : en; }
function milestoneWorkCount(milestoneId: string) { return detail.value?.workItems.filter((item) => item.milestoneId === milestoneId).length ?? 0; }
function riskSourceLabel(risk: ProjectRiskSignal) { return risk.sourceType === "weekly_update" ? copy("周报", "Update") : risk.sourceType === "work_item" ? copy("工作项", "Work Item") : copy("里程碑", "Milestone"); }
function dependencySourceLabel(dep: ProjectDependency) { return dep.sourceType === "milestone" ? copy("里程碑", "Milestone") : dep.sourceType; }
function rollupValue(rollups: ProjectSpaceRollups | undefined, key: keyof ProjectSpaceRollups) { return rollups?.[key] ?? 0; }
function openRisk(risk: ProjectRiskSignal) {
  if (risk.milestoneId && risk.sourceType === "milestone") goMilestone(risk.milestoneId);
  else if (risk.workItemId) setTab("work-items", { status: risk.status, milestoneId: risk.milestoneId });
  else setTab("risks");
}

function queryValue(value: unknown): string | undefined {
  const raw = Array.isArray(value) ? value[0] : value;
  return typeof raw === "string" && raw ? raw : undefined;
}

function normalizeGroup(value: unknown): ProjectWorkGroup {
  const raw = queryValue(value);
  return ["milestone", "status", "priority", "owner", "sourceType", "blocked"].includes(raw ?? "") ? raw as ProjectWorkGroup : "milestone";
}

function cleanQuery(query: Record<string, unknown>) {
  return Object.fromEntries(Object.entries(query).filter(([, value]) => typeof value === "string" && value !== ""));
}
</script>

<template>
  <div class="page project-space" v-if="detail">
    <header class="project-header">
      <div>
        <p class="crumb"><RouterLink :to="{ name: 'projects' }">{{ label('projects', locale) }}</RouterLink> / {{ detail.project.name }}</p>
        <h1>{{ detail.project.name }}</h1>
        <p class="meta">{{ label('owner', locale) }}: {{ detail.project.owner }} &middot; {{ label('status', locale) }}: {{ detail.project.status }} &middot; {{ label('health', locale) }}: {{ detail.project.healthStatus }} &middot; {{ label('endDate', locale) }}: {{ formatDate(detail.project.targetEndDate) }}</p>
      </div>
      <div class="header-actions">
        <div class="health-actions">
          <button v-for="h in ['on_track','at_risk','off_track']" :key="h" class="btn sm" :class="h" :disabled="!canManageProject" @click="updateHealth(h)">{{ h }}</button>
        </div>
        <button v-if="canManageMilestone" class="btn primary" @click="openCreateMilestone">{{ label('createMilestone', locale) }}</button>
      </div>
    </header>

    <nav class="space-tabs" aria-label="project space views">
      <button v-for="tab in projectSpaceTabs" :key="tab" class="tab" :class="{ active: activeTab === tab }" @click="setTab(tab)">{{ tabText(tab) }}</button>
    </nav>

    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <section v-if="activeTab === 'overview'" class="space-grid">
      <div class="main-column">
        <div class="metrics">
          <button class="metric" @click="setTab('milestones')"><strong>{{ detail.milestones.length }}</strong><span>{{ label('milestones', locale) }}</span></button>
          <button class="metric" @click="setTab('work-items')"><strong>{{ detail.workItems.length }}</strong><span>{{ label('workItems', locale) }}</span></button>
          <button class="metric warning" @click="setTab('risks')"><strong>{{ detail.risks.length }}</strong><span>{{ copy('风险信号', 'Risk signals') }}</span></button>
          <button class="metric blue" @click="setTab('dependencies')"><strong>{{ detail.dependencies.length }}</strong><span>{{ copy('依赖信号', 'Dependencies') }}</span></button>
        </div>

        <section class="panel">
          <div class="panel-head"><h2>{{ copy('当前里程碑摘要', 'Current milestone summary') }}</h2><button class="btn sm" @click="setTab('milestones')">{{ copy('管理全部', 'Manage all') }}</button></div>
          <p v-if="!detail.milestones.length" class="empty">{{ label('noData', locale) }}</p>
          <div v-for="m in detail.milestones" :key="m.id" class="milestone-card">
            <div>
              <button class="link-title" @click="goMilestone(m.id)">{{ m.title }}</button>
              <p>{{ label('owner', locale) }}: {{ m.owner }} &middot; {{ label('plannedDate', locale) }}: {{ formatDate(m.plannedDate) }} &middot; {{ milestoneWorkCount(m.id) }} {{ label('workItems', locale) }}</p>
            </div>
            <div class="milestone-actions"><span class="badge health" :class="m.healthStatus">{{ m.healthStatus }}</span><button class="btn sm" @click="openMilestoneWork(m.id)">{{ copy('查看工作项', 'View work') }}</button></div>
          </div>
        </section>
      </div>

      <aside class="side-column">
        <section class="panel pad"><h2>{{ copy('交付 Rollup', 'Delivery rollup') }}</h2><div class="rollups"><span>{{ copy('阻塞里程碑', 'Blocked milestones') }} <b>{{ rollupValue(detail.rollups, 'blockedMilestones') }}</b></span><span>{{ copy('逾期里程碑', 'Overdue milestones') }} <b>{{ rollupValue(detail.rollups, 'overdueMilestones') }}</b></span><span>{{ copy('阻塞工作项', 'Blocked work') }} <b>{{ rollupValue(detail.rollups, 'blockedWorkItems') }}</b></span><span>{{ copy('外部依赖', 'External deps') }} <b>{{ rollupValue(detail.rollups, 'externalDependencies') }}</b></span></div></section>
        <section class="panel pad"><div class="panel-head compact"><h2>{{ copy('Top 风险', 'Top risks') }}</h2><button class="btn sm" @click="setTab('risks')">{{ copy('全部', 'All') }}</button></div><p v-if="!topRisks.length" class="empty">{{ label('noData', locale) }}</p><button v-for="risk in topRisks" :key="risk.id" class="signal" @click="openRisk(risk)"><strong>{{ risk.title }}</strong><small>{{ riskSourceLabel(risk) }} · {{ risk.severity }} · {{ risk.owner || '-' }}</small></button></section>
        <section class="panel pad"><div class="panel-head compact"><h2>{{ copy('最近周报', 'Recent updates') }}</h2><button class="btn sm" @click="setTab('updates')">{{ copy('全部', 'All') }}</button></div><p v-if="!recentUpdates.length" class="empty">{{ label('noData', locale) }}</p><article v-for="u in recentUpdates" :key="u.id" class="update-card"><strong>{{ u.week }} · {{ u.author }}</strong><p>{{ u.summary }}</p></article></section>
      </aside>
    </section>

    <section v-else-if="activeTab === 'work-items'" class="section">
      <div class="section-header"><h2>{{ label('workItems', locale) }}</h2><RouterLink class="btn" :to="{ name: 'tasks', query: { projectId: id } }">{{ copy('打开全局工作台', 'Open global workspace') }}</RouterLink></div>
      <div class="filters compact-filters">
        <select :value="workGroupBy" @change="updateWorkQuery({ groupBy: ($event.target as HTMLSelectElement).value })"><option value="milestone">{{ label('milestone', locale) }}</option><option value="status">{{ label('status', locale) }}</option><option value="priority">{{ label('priority', locale) }}</option><option value="owner">{{ label('owner', locale) }}</option><option value="sourceType">{{ label('sourceType', locale) }}</option><option value="blocked">{{ label('blockedFlag', locale) }}</option></select>
        <select :value="workFilters.milestoneId || ''" @change="updateWorkQuery({ milestoneId: ($event.target as HTMLSelectElement).value })"><option value="">{{ label('milestones', locale) }}</option><option v-for="m in detail.milestones" :key="m.id" :value="m.id">{{ m.title }}</option></select>
        <select :value="workFilters.status || ''" @change="updateWorkQuery({ status: ($event.target as HTMLSelectElement).value })"><option value="">{{ label('status', locale) }}</option><option value="todo">todo</option><option value="in_progress">in_progress</option><option value="done">done</option><option value="blocked">blocked</option></select>
        <select :value="workFilters.sourceType || ''" @change="updateWorkQuery({ sourceType: ($event.target as HTMLSelectElement).value })"><option value="">{{ label('sourceType', locale) }}</option><option value="internal_task">internal_task</option><option value="gitlab_issue">gitlab_issue</option><option value="external_dependency">external_dependency</option><option value="bau_task">bau_task</option></select>
        <select :value="workFilters.blocked || ''" @change="updateWorkQuery({ blocked: ($event.target as HTMLSelectElement).value })"><option value="">{{ label('blockedFlag', locale) }}</option><option value="true">true</option><option value="false">false</option></select>
        <select :value="workFilters.overdue || ''" @change="updateWorkQuery({ overdue: ($event.target as HTMLSelectElement).value })"><option value="">{{ label('overdueTask', locale) }}</option><option value="true">true</option><option value="false">false</option></select>
        <button class="btn" @click="setTab('work-items', { milestoneId: undefined, status: undefined, priority: undefined, owner: undefined, sourceType: undefined, blocked: undefined, overdue: undefined })">{{ label('clearFilters', locale) }}</button>
      </div>
      <div v-for="group in groupedWorkItems" :key="group.key" class="work-group">
        <h3>{{ group.label }} <span>{{ group.items.length }}</span></h3>
        <div v-for="w in group.items" :key="w.id" class="work-card">
          <strong>{{ w.title || w.id }}</strong> <span class="badge">{{ w.sourceType }}</span> <span class="badge" :class="displayProjectWorkStatus(w)">{{ displayProjectWorkStatus(w) }}</span>
          <p>{{ projectWorkBreadcrumb(w, detail.project.name, detail.milestones) }} &middot; {{ label('owner', locale) }}: {{ w.owner || '-' }} &middot; {{ label('priority', locale) }}: {{ w.priority || 'P1' }}</p>
          <p v-if="w.sourceType === 'gitlab_issue'" class="gitlab-meta">GitLab: {{ gitlabState(w) || '-' }} &middot; {{ label('assignee', locale) }}: {{ gitlabAssignee(w) || '-' }} &middot; {{ label('labels', locale) }}: {{ gitlabLabels(w).join(', ') || '-' }} &middot; {{ label('lastSynced', locale) }}: {{ w.lastSyncedAt?.slice(0,10) || '-' }}</p>
          <div class="card-actions"><RouterLink class="btn sm" :to="{ name: 'task-detail', params: { id: w.id } }">{{ label('edit', locale) }}</RouterLink><a v-if="w.sourceType === 'gitlab_issue' && w.sourceUrl" :href="w.sourceUrl" target="_blank" rel="noreferrer">{{ label('openIssue', locale) }}</a></div>
        </div>
      </div>
      <p v-if="!groupedWorkItems.length" class="empty">{{ label('noData', locale) }}</p>
    </section>

    <section v-else-if="activeTab === 'milestones'" class="section">
      <div class="section-header"><h2>{{ label('milestones', locale) }}</h2><button v-if="canManageMilestone" class="btn primary" @click="showMilestoneForm = !showMilestoneForm">{{ label('createMilestone', locale) }}</button><span v-else class="empty">{{ label('noPermission', locale) }}</span></div>
      <form v-if="showMilestoneForm" class="form" @submit.prevent="createMilestone"><input v-model="msForm.title" :placeholder="label('title', locale)" required /><input v-model="msForm.owner" :placeholder="label('owner', locale)" required /><input v-model="msForm.completionCriteria" :placeholder="label('criteria', locale)" required /><select v-model="msForm.riskLevel"><option value="low">low</option><option value="medium">medium</option><option value="high">high</option></select><input v-model="msForm.plannedDate" type="date" :placeholder="label('plannedDate', locale)" /><div class="row"><button class="btn primary" type="submit">{{ label('save', locale) }}</button><button class="btn" type="button" @click="showMilestoneForm = false">{{ label('cancel', locale) }}</button></div></form>
      <table v-if="detail.milestones.length"><thead><tr><th>{{ label('title', locale) }}</th><th>{{ label('status', locale) }}</th><th>{{ label('health', locale) }}</th><th>{{ label('owner', locale) }}</th><th>{{ label('plannedDate', locale) }}</th><th>{{ label('edit', locale) }}</th></tr></thead><tbody><tr v-for="m in detail.milestones" :key="m.id"><td><button class="link-title" @click="goMilestone(m.id)">{{ m.title }}</button></td><td>{{ m.status }}</td><td>{{ m.healthStatus }}</td><td>{{ m.owner }}</td><td>{{ formatDate(m.plannedDate) }}</td><td class="actions"><button class="btn sm" :disabled="!canManageMilestone" @click="transitionMilestone(m, 'active')">active</button><button class="btn sm" :disabled="!canManageMilestone" @click="transitionMilestone(m, 'blocked')">blocked</button><button class="btn sm" :disabled="!canManageMilestone" @click="transitionMilestone(m, 'completed')">completed</button><button class="btn sm" @click="openMilestoneWork(m.id)">{{ label('workItems', locale) }}</button></td></tr></tbody></table>
      <p v-else class="empty">{{ label('noData', locale) }}</p>
    </section>

    <section v-else-if="activeTab === 'updates'" class="section"><div class="section-header"><h2>{{ label('review', locale) }}</h2><RouterLink class="btn" :to="{ name: 'review', query: { projectId: id } }">{{ copy('打开周度回顾', 'Open review') }}</RouterLink></div><article v-for="u in detail.updates" :key="u.id" class="update-card"><strong>{{ u.week }} · {{ u.author }}</strong><p>{{ u.summary }}</p><p class="muted">{{ u.risk || u.blockers || u.decisionsNeeded || '-' }}</p></article><p v-if="!detail.updates.length" class="empty">{{ label('noData', locale) }}</p></section>

    <section v-else-if="activeTab === 'risks'" class="section"><h2>{{ copy('风险', 'Risks') }}</h2><button v-for="risk in detail.risks" :key="risk.id" class="risk-card" @click="openRisk(risk)"><strong>{{ risk.title }}</strong><span>{{ riskSourceLabel(risk) }} · {{ risk.severity }} · {{ risk.status }} · {{ risk.owner || '-' }}</span><p>{{ risk.message }}</p></button><p v-if="!detail.risks.length" class="empty">{{ label('noData', locale) }}</p></section>

    <section v-else-if="activeTab === 'dependencies'" class="section"><h2>{{ copy('依赖', 'Dependencies') }}</h2><article v-for="dep in detail.dependencies" :key="dep.id" class="risk-card"><strong>{{ dep.title }}</strong><span>{{ dependencySourceLabel(dep) }} · {{ dep.status }} · {{ dep.owner || '-' }}</span><p>{{ dep.message }}</p></article><p v-if="!detail.dependencies.length" class="empty">{{ label('noData', locale) }}</p></section>

    <section v-else class="section"><h2>{{ copy('设置', 'Settings') }}</h2><div class="panel pad"><p>{{ label('summary', locale) }}: {{ detail.project.objective || '-' }}</p><p>{{ label('projectType', locale) }}: {{ detail.project.projectType || '-' }}</p><p>{{ label('priority', locale) }}: {{ detail.project.priority || '-' }}</p></div></section>
  </div>
</template>

<style scoped>
.project-space { max-width: 1180px; }
.project-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 18px; margin-bottom: 18px; }
h1 { margin: 0; font-size: clamp(1.9rem, 4vw, 2.8rem); letter-spacing: -.05em; }
h2 { margin: 0 0 12px; font-size: 1.15rem; }
h3 { margin: 0 0 10px; font-size: .98rem; }
.crumb, .meta, .muted { color: var(--color-text-muted); font-size: .9rem; }
.crumb { margin: 0 0 6px; }
.crumb a { color: var(--color-primary-light); font-weight: 700; text-decoration: none; }
.meta { margin: 8px 0 0; }
.header-actions, .health-actions, .row, .actions, .milestone-actions, .card-actions { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; }
.space-tabs { display: flex; gap: 6px; padding: 6px; border: 1px solid #e0e9e4; border-radius: 18px; background: rgba(255,255,255,.78); margin-bottom: 18px; overflow-x: auto; box-shadow: var(--shadow-sm); }
.tab { border: 0; border-radius: var(--radius-full); padding: 9px 13px; background: transparent; color: var(--color-text-muted); cursor: pointer; white-space: nowrap; }
.tab.active { background: var(--color-primary); color: #fff; font-weight: 800; }
.space-grid { display: grid; grid-template-columns: minmax(0, 1fr) 340px; gap: 16px; align-items: start; }
.main-column, .side-column { display: grid; gap: 16px; }
.metrics { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }
.metric { text-align: left; border: 1px solid #e2ebe6; border-radius: 18px; padding: 15px; background: var(--color-surface); color: var(--color-text); box-shadow: var(--shadow-sm); cursor: pointer; }
.metric strong { display: block; font-size: 2rem; letter-spacing: -.06em; }
.metric span { color: var(--color-text-muted); font-size: .82rem; }
.metric.warning strong { color: var(--color-warning); }
.metric.blue strong { color: var(--color-info); }
.panel, .work-group { background: var(--color-surface); border: 1px solid #e4ece8; border-radius: 18px; box-shadow: var(--shadow-sm); }
.panel { padding: 16px; }
.panel.pad { padding: 16px; }
.panel-head, .section-header { display: flex; justify-content: space-between; align-items: center; gap: 10px; margin-bottom: 12px; }
.panel-head.compact { margin-bottom: 8px; }
.milestone-card { display: flex; justify-content: space-between; gap: 12px; padding: 13px 0; border-top: 1px solid #edf2ef; }
.milestone-card:first-of-type { border-top: 0; }
.milestone-card p, .update-card p, .risk-card p, .work-card p { margin: 5px 0 0; color: var(--color-text-muted); }
.link-title { border: 0; padding: 0; background: none; color: var(--color-primary-light); font-weight: 800; cursor: pointer; font-size: .98rem; }
.rollups { display: grid; gap: 8px; }
.rollups span { display: flex; justify-content: space-between; padding: 9px 10px; border-radius: 12px; background: var(--color-surface-hover); color: var(--color-text-muted); }
.rollups b { color: var(--color-text); }
.signal, .risk-card { width: 100%; display: block; text-align: left; border: 1px solid #e6efea; border-radius: 14px; background: var(--color-surface-hover); padding: 12px; margin-bottom: 8px; color: var(--color-text); cursor: pointer; }
.signal small, .risk-card span { display: block; margin-top: 4px; color: var(--color-text-subtle); font-size: .78rem; }
.section { margin-top: 18px; }
.filters.compact-filters { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; background: var(--color-surface); padding: 12px; border-radius: 14px; margin-bottom: 14px; box-shadow: var(--shadow-sm); }
.filters select { padding: 8px 10px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); }
.work-group { padding: 14px; margin-bottom: 12px; }
.work-group h3 { display: flex; justify-content: space-between; color: var(--color-text); }
.work-group h3 span { color: var(--color-text-muted); }
.work-card, .update-card { background: var(--color-surface-hover); padding: 13px; border-radius: var(--radius-md); margin-bottom: 8px; border: 1px solid #edf2ef; }
.gitlab-meta { font-size: .84rem; }
.form { display: flex; flex-direction: column; gap: 10px; max-width: 520px; margin-bottom: 16px; background: var(--color-surface); padding: 16px; border-radius: 14px; box-shadow: var(--shadow-md); }
.form input, .form select { padding: 10px 12px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); }
table { width: 100%; border-collapse: collapse; background: var(--color-surface); border-radius: 14px; overflow: hidden; box-shadow: var(--shadow-sm); }
th { text-align: left; padding: 10px 14px; background: var(--color-surface-alt); font-size: .82rem; color: var(--color-text-muted); }
td { padding: 10px 14px; border-top: 1px solid #eaf0ed; }
.badge { display: inline-block; padding: 2px 8px; border-radius: var(--radius-full); background: #e0f2fe; color: #0369a1; font-size: .74rem; font-weight: 700; }
.badge.blocked { background: var(--color-danger-bg); color: var(--color-danger); }
.btn { padding: 8px 14px; border-radius: var(--radius-sm); border: 1px solid var(--color-border); background: var(--color-surface); cursor: pointer; font-size: .85rem; text-decoration: none; color: var(--color-text); }
.btn:disabled { opacity: .45; cursor: not-allowed; }
.btn.primary { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.btn.sm { font-size: .78rem; padding: 5px 10px; }
.btn.on_track { background: var(--color-success-bg); color: var(--color-success); border-color: #86efac; }
.btn.at_risk { background: var(--color-warning-bg); color: var(--color-warning); border-color: #fcd34d; }
.btn.off_track { background: var(--color-danger-bg); color: #dc2626; border-color: #fca5a5; }
a { color: var(--color-primary-light); font-weight: 700; }
@media (max-width: 980px) { .project-header, .space-grid { display: grid; grid-template-columns: 1fr; } .metrics { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 620px) { .metrics { grid-template-columns: 1fr; } .milestone-card, .section-header { display: grid; } }
</style>
