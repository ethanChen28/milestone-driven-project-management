<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { Locale } from "../i18n";
import { apiFetch, can, label, type LinkedWorkItem, type Milestone, type Project, type WorkspaceRole } from "../api";

type TaskForm = {
  sourceType: string;
  sourceId: string;
  sourceUrl: string;
  title: string;
  projectId: string;
  milestoneId: string;
  workstreamId: string;
  owner: string;
  status: string;
  priority: string;
  estimate: string;
  plannedStartDate: string;
  plannedEndDate: string;
  dueDate: string;
  blocked: boolean;
  tags: string;
};

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const route = useRoute();
const router = useRouter();

const id = route.params.id as string;
const isCreate = id === "new";
const loading = ref(false);
const saving = ref(false);
const deleting = ref(false);
const error = ref("");
const detail = ref<LinkedWorkItem | null>(null);
const projects = ref<Project[]>([]);
const milestones = ref<Milestone[]>([]);
const canManageTask = computed(() => can(currentRole.value, "manageWorkItem"));
const form = ref<TaskForm>({
  sourceType: "internal_task",
  sourceId: "",
  sourceUrl: "",
  title: "",
  projectId: "",
  milestoneId: "",
  workstreamId: "",
  owner: "",
  status: "todo",
  priority: "P2",
  estimate: "1d",
  plannedStartDate: "",
  plannedEndDate: "",
  dueDate: "",
  blocked: false,
  tags: "",
});

const filteredMilestones = computed(() => milestones.value.filter((milestone) => !form.value.projectId || milestone.projectId === form.value.projectId));

function mapItemToForm(item: LinkedWorkItem): TaskForm {
  return {
    sourceType: item.sourceType,
    sourceId: item.sourceId,
    sourceUrl: item.sourceUrl,
    title: item.title,
    projectId: item.projectId,
    milestoneId: item.milestoneId,
    workstreamId: item.workstreamId,
    owner: item.owner,
    status: item.status,
    priority: item.priority || "P2",
    estimate: item.estimate || "1d",
    plannedStartDate: item.plannedStartDate?.slice(0, 10) || "",
    plannedEndDate: item.plannedEndDate?.slice(0, 10) || "",
    dueDate: item.dueDate?.slice(0, 10) || "",
    blocked: item.blocked,
    tags: (item.tags ?? []).join(", "),
  };
}

function mapFormToPayload(): Record<string, unknown> {
  const toIso = (value: string) => (value ? new Date(`${value}T00:00:00Z`).toISOString() : undefined);
  return {
    sourceType: form.value.sourceType,
    sourceId: form.value.sourceId,
    sourceUrl: form.value.sourceUrl,
    title: form.value.title,
    projectId: form.value.projectId,
    milestoneId: form.value.milestoneId,
    workstreamId: form.value.workstreamId,
    owner: form.value.owner,
    status: form.value.status,
    priority: form.value.priority,
    estimate: form.value.estimate,
    plannedStartDate: toIso(form.value.plannedStartDate),
    plannedEndDate: toIso(form.value.plannedEndDate),
    dueDate: toIso(form.value.dueDate),
    blocked: form.value.blocked,
    tags: form.value.tags
      .split(",")
      .map((tag) => tag.trim())
      .filter(Boolean),
  };
}

async function load() {
  loading.value = true;
  try {
    const [projectData, milestoneData] = await Promise.all([
      apiFetch<Project[]>("/projects"),
      apiFetch<Milestone[]>("/milestones"),
    ]);
    projects.value = projectData;
    milestones.value = milestoneData;
    if (!isCreate) {
      detail.value = await apiFetch<LinkedWorkItem>(`/work-items?id=${id}`);
      form.value = mapItemToForm(detail.value);
    }
    error.value = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

onMounted(load);

async function save() {
  if (!canManageTask.value) return;
  saving.value = true;
  try {
    const payload = JSON.stringify(mapFormToPayload());
    if (isCreate) {
      const created = await apiFetch<LinkedWorkItem>("/work-items", { method: "POST", body: payload });
      router.push({ name: "tasks", query: { createdId: created.id } });
    } else {
      await apiFetch<LinkedWorkItem>(`/work-items?id=${id}`, { method: "PUT", body: payload });
      router.push({ name: "tasks", query: { updatedId: id } });
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    saving.value = false;
  }
}

async function remove() {
  if (!detail.value || !canManageTask.value) return;
  if (!window.confirm(label("taskDeleteConfirm", locale))) return;
  deleting.value = true;
  try {
    await apiFetch(`/work-items?id=${detail.value.id}`, { method: "DELETE" });
    router.push({ name: "tasks" });
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    deleting.value = false;
  }
}
</script>

<template>
  <div class="page">
    <div class="header">
      <div>
        <p class="eyebrow">{{ label("taskWorkspace", locale) }}</p>
        <h1>{{ isCreate ? label("newTask", locale) : label("taskDetail", locale) }}</h1>
      </div>
      <button class="btn" @click="router.push({ name: 'tasks' })">{{ label("viewAll", locale) }}</button>
    </div>

    <p v-if="loading" class="empty">Loading...</p>
    <p v-if="error" class="error">{{ error }}</p>

    <form v-if="!loading" class="form" @submit.prevent="save">
      <div class="form-grid">
        <label>{{ label("title", locale) }}<input v-model="form.title" required /></label>
        <label>{{ label("project", locale) }}<select v-model="form.projectId" required><option value="">{{ label("project", locale) }}</option><option v-for="project in projects" :key="project.id" :value="project.id">{{ project.name }}</option></select></label>
        <label>{{ label("milestone", locale) }}<select v-model="form.milestoneId"><option value="">{{ label("milestone", locale) }}</option><option v-for="milestone in filteredMilestones" :key="milestone.id" :value="milestone.id">{{ milestone.title }}</option></select></label>
        <label>{{ label("workstream", locale) }}<input v-model="form.workstreamId" /></label>
        <label>{{ label("ownerTeam", locale) }}<input v-model="form.owner" required /></label>
        <label>{{ label("status", locale) }}<select v-model="form.status"><option value="todo">todo</option><option value="in_progress">in_progress</option><option value="done">done</option><option value="cancelled">cancelled</option></select></label>
        <label>{{ label("priorityBucket", locale) }}<select v-model="form.priority"><option value="P0">P0</option><option value="P1">P1</option><option value="P2">P2</option><option value="P3">P3</option></select></label>
        <label>{{ label("estimate", locale) }}<input v-model="form.estimate" placeholder="1d" /></label>
        <label>{{ label("sourceType", locale) }}<select v-model="form.sourceType"><option value="internal_task">internal_task</option><option value="gitlab_issue">gitlab_issue</option><option value="external_dependency">external_dependency</option><option value="bau_task">bau_task</option></select></label>
        <label>{{ label("sourceId", locale) }}<input v-model="form.sourceId" /></label>
        <label>{{ label("sourceUrl", locale) }}<input v-model="form.sourceUrl" /></label>
        <label>{{ label("plannedDate", locale) }}<input v-model="form.plannedStartDate" type="date" /></label>
        <label>{{ label("endDate", locale) }}<input v-model="form.plannedEndDate" type="date" /></label>
        <label>{{ label("dueDate", locale) }}<input v-model="form.dueDate" type="date" /></label>
        <label>{{ label("tags", locale) }}<input v-model="form.tags" placeholder="tag1, tag2" /></label>
      </div>
      <label class="check"><input v-model="form.blocked" type="checkbox" /> {{ label("blockedFlag", locale) }}</label>
      <div class="row">
        <button class="btn primary" type="submit" :disabled="saving">{{ label("saveTask", locale) }}</button>
        <button v-if="!isCreate" class="btn danger" type="button" :disabled="deleting" @click="remove">{{ label("deleteTask", locale) }}</button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.page { max-width: 1120px; }
.header { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; margin-bottom: 18px; }
.eyebrow { margin: 0 0 4px; color: #6b8a80; font-size: .82rem; }
h1 { margin: 0; }
.form { display: grid; gap: 14px; background: #fff; padding: 18px; border-radius: 16px; box-shadow: 0 1px 6px rgba(0,0,0,.06); }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
label { display: grid; gap: 6px; font-size: .84rem; color: #4a7a6d; }
input, select { padding: 10px 12px; border: 1px solid #d1d9d6; border-radius: 10px; font-family: inherit; }
.check { display: flex; align-items: center; gap: 8px; }
.row { display: flex; gap: 10px; flex-wrap: wrap; }
.btn { padding: 8px 18px; border-radius: 10px; border: 1px solid #d1d9d6; background: #fff; cursor: pointer; }
.btn.primary { background: #10352a; color: #fff; border-color: #10352a; }
.btn.danger { background: #fff5f5; color: #b91c1c; border-color: #fecaca; }
.empty { color: #6b8a80; }
.error { color: #b91c1c; background: #fee2e2; padding: 10px 12px; border-radius: 10px; }
@media (max-width: 720px) {
  .form-grid { grid-template-columns: 1fr; }
  .header { flex-direction: column; }
}
</style>
