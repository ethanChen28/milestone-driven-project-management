<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import { useRouter } from "vue-router";
import type { Locale } from "../i18n";
import PersonPicker from "../components/PersonPicker.vue";
import { label, apiFetch, can, dateInputToIso, listUsers, type UserProfile, type Project, type WorkspaceRole } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const currentUser = inject<Ref<string>>("currentUser")!;
const router = useRouter();
const projects = ref<Project[]>([]);
const showForm = ref(false);
const loading = ref(false);
const error = ref("");
const canCreate = computed(() => can(currentRole.value, "manageProject"));
const filters = ref({ owner: "", status: "", health: "", q: "" });
const participantsText = ref("");
const participantsArray = ref<string[]>([]);
const directoryUsers = ref<UserProfile[]>([]);

function initialForm() {
  return {
    name: "",
    objective: "",
    owner: currentUser.value,
    status: "active",
    healthStatus: "on_track",
    priority: "P1",
    projectType: "product",
    targetStartDate: "",
    targetEndDate: "",
  };
}

const form = ref(initialForm());

const projectStats = computed(() => ({
  total: projects.value.length,
  active: projects.value.filter((project) => project.status === "active").length,
  risky: projects.value.filter((project) => ["at_risk", "off_track"].includes(project.healthStatus)).length,
}));

const ownerOptions = computed(() => {
  const owners = new Set<string>();
  directoryUsers.value.forEach((user) => owners.add(user.id));
  projects.value.forEach((project) => { if (project.owner) owners.add(project.owner); });
  if (filters.value.owner) owners.add(filters.value.owner);
  return [...owners].sort();
});

function query() {
  const params = new URLSearchParams();
  Object.entries(filters.value).forEach(([key, value]) => { if (value) params.set(key, value); });
  return params.toString() ? `?${params.toString()}` : "";
}

async function load() {
  loading.value = true;
  try {
    projects.value = await apiFetch<Project[]>(`/projects${query()}`);
    directoryUsers.value = await listUsers();
    error.value = "";
  } catch (err) {
    projects.value = [];
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function parseParticipants() {
  const participants = [...participantsArray.value];
  if (form.value.owner.trim()) participants.unshift(form.value.owner.trim());
  return [...new Set(participants)];
}

function resetForm() {
  form.value = initialForm();
  participantsText.value = "";
}

function toggleForm() {
  if (!showForm.value) resetForm();
  showForm.value = !showForm.value;
}

async function create() {
  if (!canCreate.value) return;
  try {
    await apiFetch("/projects", {
      method: "POST",
      body: JSON.stringify({
        ...form.value,
        participants: parseParticipants(),
        targetStartDate: dateInputToIso(form.value.targetStartDate),
        targetEndDate: dateInputToIso(form.value.targetEndDate),
      }),
    });
    showForm.value = false;
    resetForm();
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}

function clearFilters() { filters.value = { owner: "", status: "", health: "", q: "" }; load(); }
function go(id: string) { router.push({ name: "project-detail", params: { id } }); }
function formatDate(value?: string) { return value?.slice(0, 10) || "-"; }
function copy(key: string, zh: string, en: string) {
  const translated = label(key, locale.value);
  if (translated !== key) return translated;
  return locale.value === "zh-CN" ? zh : en;
}
</script>

<template>
  <div class="page projects-page">
    <div class="workspace-header">
      <div>
        <p class="eyebrow">Goal Manager</p>
        <h1>{{ label("projects", locale) }}</h1>
        <p class="subtitle">
          {{ copy("projectWorkspaceSubtitle", "按负责人、健康度和交付日期管理项目组合", "Manage project portfolios by owner, health and delivery date") }}
        </p>
      </div>
      <div class="header-actions">
        <div class="metric-strip" aria-label="project summary">
          <span><strong>{{ projectStats.total }}</strong>{{ copy("totalProjects", "全部", "Total") }}</span>
          <span><strong>{{ projectStats.active }}</strong>{{ copy("activeProjects", "活跃", "Active") }}</span>
          <span><strong>{{ projectStats.risky }}</strong>{{ copy("riskProjects", "风险", "Risk") }}</span>
        </div>
        <button v-if="canCreate" class="btn primary" @click="toggleForm">{{ showForm ? label("cancel", locale) : label("createProject", locale) }}</button>
        <span v-else class="empty">{{ label('noPermission', locale) }}</span>
      </div>
    </div>

    <template v-if="showForm">
      <p v-if="error" class="error">{{ error }}</p>

      <form class="form project-form" @submit.prevent="create">
        <div class="form-heading">
          <div>
            <p class="eyebrow">{{ copy("newProject", "新项目", "New project") }}</p>
            <h2>{{ label("createProject", locale) }}</h2>
          </div>
          <span class="panel-tag">{{ copy("workspaceObject", "工作区对象", "Workspace object") }}</span>
        </div>
        <div class="form-grid">
          <label class="field span-2">
            <span>{{ label("name", locale) }}</span>
            <input v-model="form.name" :placeholder="label('name', locale)" required />
          </label>
          <label class="field span-2">
            <span>{{ label("objective", locale) }}</span>
            <textarea v-model="form.objective" :placeholder="label('objective', locale)" rows="3" />
          </label>
          <label class="field">
            <span>{{ label("owner", locale) }}</span>
            <PersonPicker v-model="form.owner" :users="directoryUsers" :placeholder="label('owner', locale)" />
          </label>
          <label class="field">
            <span>{{ copy("participants", "参与者", "Participants") }}</span>
            <PersonPicker v-model="form.participantsArray" mode="multi" :users="directoryUsers" :placeholder="copy('participantsPlaceholder', '选择参与者', 'Select participants')" />
          </label>
          <label class="field">
            <span>{{ label("projectType", locale) }}</span>
            <select v-model="form.projectType">
              <option value="product">product</option>
              <option value="platform">platform</option>
              <option value="operations">operations</option>
              <option value="research">research</option>
            </select>
          </label>
          <label class="field">
            <span>{{ label("status", locale) }}</span>
            <select v-model="form.status">
              <option value="active">active</option>
              <option value="done">done</option>
              <option value="archived">archived</option>
            </select>
          </label>
          <label class="field">
            <span>{{ label("health", locale) }}</span>
            <select v-model="form.healthStatus">
              <option value="on_track">on_track</option>
              <option value="at_risk">at_risk</option>
              <option value="off_track">off_track</option>
            </select>
          </label>
          <label class="field">
            <span>{{ label("priority", locale) }}</span>
            <select v-model="form.priority">
              <option value="P0">P0</option>
              <option value="P1">P1</option>
              <option value="P2">P2</option>
              <option value="P3">P3</option>
            </select>
          </label>
          <label class="field">
            <span>{{ label("startDate", locale) }}</span>
            <input v-model="form.targetStartDate" type="date" />
          </label>
          <label class="field">
            <span>{{ label("endDate", locale) }}</span>
            <input v-model="form.targetEndDate" type="date" />
          </label>
        </div>
        <div class="row form-actions">
          <button class="btn primary" type="submit">{{ label("save", locale) }}</button>
          <button class="btn" type="button" @click="showForm = false">{{ label("cancel", locale) }}</button>
        </div>
      </form>
    </template>

    <template v-else>
      <section class="filters" aria-label="project filters">
        <div class="filter-title">
          <strong>{{ label('filters', locale) }}</strong>
          <span>{{ copy("filterHint", "回车或点击筛选刷新列表", "Press Enter or filter to refresh") }}</span>
        </div>
        <label>
          <span>{{ label("keyword", locale) }}</span>
          <input v-model="filters.q" placeholder="q" @keyup.enter="load" />
        </label>
        <label>
          <span>{{ label("owner", locale) }}</span>
          <select v-model="filters.owner" @change="load">
            <option value="">{{ label('owner', locale) }}</option>
            <option v-for="owner in ownerOptions" :key="owner" :value="owner">{{ owner }}</option>
          </select>
        </label>
        <label>
          <span>{{ label("status", locale) }}</span>
          <select v-model="filters.status">
            <option value="">{{ label('status', locale) }}</option>
            <option value="active">active</option>
            <option value="done">done</option>
            <option value="archived">archived</option>
          </select>
        </label>
        <label>
          <span>{{ label("health", locale) }}</span>
          <select v-model="filters.health">
            <option value="">{{ label('health', locale) }}</option>
            <option value="on_track">on_track</option>
            <option value="at_risk">at_risk</option>
            <option value="off_track">off_track</option>
          </select>
        </label>
        <div class="filter-actions">
          <button class="btn" @click="load">{{ label('filters', locale) }}</button>
          <button class="btn" @click="clearFilters">{{ label('clearFilters', locale) }}</button>
        </div>
      </section>

      <p v-if="error" class="error">{{ error }}</p>

      <div class="table-shell" v-if="projects.length">
      <table>
        <thead>
          <tr>
            <th>{{ label('name', locale) }}</th>
            <th>{{ label('owner', locale) }}</th>
            <th>{{ label('status', locale) }}</th>
            <th>{{ label('health', locale) }}</th>
            <th>{{ label('priority', locale) }}</th>
            <th>{{ label('endDate', locale) }}</th>
            <th>{{ label('actions', locale) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in projects" :key="p.id" class="clickable" tabindex="0" role="link" @click="go(p.id)" @keyup.enter="go(p.id)">
            <td>
              <strong class="project-name">{{ p.name }}</strong>
              <small>{{ p.objective || copy("noObjective", "未填写项目目标", "No objective") }}</small>
            </td>
            <td>
              <strong>{{ p.owner }}</strong>
              <small>{{ (p.participants?.length || 0) }} {{ copy("members", "名成员", "members") }}</small>
            </td>
            <td><span class="badge" :class="p.status">{{ p.status }}</span></td>
            <td><span class="badge health" :class="p.healthStatus">{{ p.healthStatus }}</span></td>
            <td><span class="priority-pill">{{ p.priority }}</span></td>
            <td>{{ formatDate(p.targetEndDate) }}</td>
            <td><button class="btn sm" @click.stop="router.push({ name: 'project-edit', params: { id: p.id } })">{{ label('edit', locale) }}</button></td>
          </tr>
        </tbody>
      </table>
    </div>
    <section v-else class="empty-state">
      <p class="eyebrow">{{ loading ? "Loading" : label("noData", locale) }}</p>
      <h2>{{ copy("emptyProjectsTitle", "还没有项目", "No projects yet") }}</h2>
      <p>{{ copy("emptyProjectsBody", "先创建一个项目，再拆解里程碑和工作项。", "Create a project first, then break it down into milestones and work items.") }}</p>
      <button v-if="canCreate" class="btn" @click="toggleForm">{{ label("createProject", locale) }}</button>
    </section>
    </template>
  </div>
</template>

<style scoped>
.projects-page { max-width: 1120px; }
.workspace-header { display: flex; justify-content: space-between; gap: 24px; align-items: flex-start; margin-bottom: 18px; }
.eyebrow { margin: 0 0 6px; color: var(--color-primary-light); font-size: .72rem; font-weight: 800; letter-spacing: .16em; text-transform: uppercase; }
h1, h2 { margin: 0; color: var(--color-text); }
h1 { font-size: clamp(2rem, 4vw, 3rem); letter-spacing: -.05em; }
h2 { font-size: 1.05rem; }
.subtitle { margin: 8px 0 0; color: var(--color-text-muted); }
.header-actions { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; justify-content: flex-end; }
.metric-strip { display: flex; gap: 8px; padding: 6px; border: 1px solid #dce7e1; border-radius: var(--radius-full); background: rgba(255,255,255,.72); box-shadow: var(--shadow-sm); }
.metric-strip span { display: inline-flex; gap: 6px; align-items: center; padding: 6px 10px; border-radius: var(--radius-full); color: var(--color-text-muted); font-size: .78rem; }
.metric-strip strong { color: var(--color-text); font-size: .95rem; }
.filters { display: grid; grid-template-columns: 1.2fr repeat(4, minmax(130px, 1fr)) auto; gap: 10px; align-items: end; background: var(--color-surface); padding: 14px; border: 1px solid #edf2ef; border-radius: 18px; margin-bottom: 16px; box-shadow: var(--shadow-sm); }
.filter-title { align-self: center; }
.filter-title strong { display: block; }
.filter-title span, .field span, .filters label span, table small { color: var(--color-text-subtle); font-size: .78rem; }
.filters label, .field { display: grid; gap: 6px; }
.filters input, .filters select, .form input, .form select, .form textarea { width: 100%; box-sizing: border-box; padding: 10px 12px; border: 1px solid var(--color-border); border-radius: var(--radius-md); color: var(--color-text); font-family: inherit; font-size: .9rem; background: var(--color-surface); }
.form textarea { resize: vertical; }
.filter-actions { display: flex; gap: 8px; }
.project-form { display: grid; gap: 16px; margin-bottom: 18px; background: linear-gradient(135deg, #ffffff 0%, #f7fbf8 100%); padding: 18px; border: 1px solid #e1ebe6; border-radius: 20px; box-shadow: var(--shadow-md); }
.form-heading { display: flex; justify-content: space-between; gap: 16px; align-items: flex-start; }
.panel-tag { padding: 4px 10px; border-radius: var(--radius-full); background: #e9fff6; color: var(--color-primary-light); font-size: .78rem; font-weight: 700; }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.span-2 { grid-column: span 2; }
.row { display: flex; gap: 10px; }
.form-actions { justify-content: flex-end; border-top: 1px solid #e7efeb; padding-top: 14px; }
.table-shell { overflow: hidden; border: 1px solid #e4ece8; border-radius: 18px; background: var(--color-surface); box-shadow: var(--shadow-md); }
table { width: 100%; border-collapse: collapse; }
th { text-align: left; padding: 12px 16px; background: #f8fbf9; font-size: .78rem; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: .06em; }
td { padding: 14px 16px; border-top: 1px solid #eaf0ed; vertical-align: middle; }
td:first-child { min-width: 260px; }
td strong, td small { display: block; }
.project-name { color: var(--color-text); }
.clickable { cursor: pointer; transition: background .12s ease, transform .12s ease; color: inherit; }
.clickable:hover { background: var(--color-surface-hover); }
.badge.done { background: #dbeafe; color: #1d4ed8; }
.priority-pill { display: inline-flex; min-width: 34px; justify-content: center; padding: 4px 8px; border-radius: var(--radius-full); background: #eef2ff; color: #4338ca; font-weight: 800; font-size: .78rem; }
.empty-state { display: grid; gap: 10px; max-width: 520px; padding: 42px; border: 1px dashed #b9cbc4; border-radius: 24px; background: rgba(255,255,255,.76); }
.empty-state p { margin: 0; color: var(--color-text-muted); }

@media (max-width: 920px) {
  .workspace-header { display: grid; }
  .header-actions { justify-content: flex-start; }
  .filters { grid-template-columns: 1fr 1fr; }
  .filter-title, .filter-actions { grid-column: span 2; }
  .table-shell { overflow-x: auto; }
}

@media (max-width: 640px) {
  .filters, .form-grid { grid-template-columns: 1fr; }
  .filter-title, .filter-actions, .span-2 { grid-column: auto; }
  .metric-strip { width: 100%; justify-content: space-between; border-radius: var(--radius-lg); }
}
</style>
