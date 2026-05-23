<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import { useRouter } from "vue-router";
import type { Locale } from "../i18n";
import PersonPicker from "../components/PersonPicker.vue";
import { dateInputToIso, label, apiFetch, can, listUsers, type UserProfile, type Milestone, type WorkspaceRole } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const router = useRouter();
const milestones = ref<Milestone[]>([]);
const directoryUsers = ref<UserProfile[]>([]);
const showForm = ref(false);
const canCreate = computed(() => can(currentRole.value, "manageMilestone"));
const filters = ref({ projectId: "", owner: "", status: "", health: "", risk: "" });
const form = ref({ projectId: "", title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "", riskLevel: "low" });

function query() {
  const params = new URLSearchParams();
  Object.entries(filters.value).forEach(([key, value]) => { if (value) params.set(key, value); });
  return params.toString() ? `?${params.toString()}` : "";
}

async function load() {
  try {
    milestones.value = await apiFetch<Milestone[]>(`/milestones${query()}`);
    directoryUsers.value = await listUsers();
  } catch { milestones.value = []; }
}

onMounted(load);

async function create() {
  if (!canCreate.value) return;
  await apiFetch("/milestones", { method: "POST", body: JSON.stringify({ ...form.value, plannedDate: dateInputToIso(form.value.plannedDate) }) });
  showForm.value = false;
  form.value = { projectId: "", title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "", riskLevel: "low" };
  await load();
}

function clearFilters() { filters.value = { projectId: "", owner: "", status: "", health: "", risk: "" }; load(); }
function go(id: string) { router.push({ name: "milestone-detail", params: { id } }); }
</script>

<template>
  <div class="page">
    <div class="header">
      <h1>{{ label("milestones", locale) }}</h1>
      <button v-if="canCreate" class="btn primary" @click="showForm = !showForm">{{ label("createMilestone", locale) }}</button>
      <span v-else class="permission-hint">{{ label('noPermission', locale) }}<small>{{ label("needProjectOwner", locale) }}</small></span>
    </div>
    <div class="filters">
      <strong>{{ label('filters', locale) }}</strong>
      <input v-model="filters.projectId" placeholder="Project ID" />
      <input v-model="filters.owner" :placeholder="label('owner', locale)" />
      <select v-model="filters.status"><option value="">{{ label('status', locale) }}</option><option value="not_started">not_started</option><option value="active">active</option><option value="blocked">blocked</option><option value="completed">completed</option><option value="cancelled">cancelled</option></select>
      <select v-model="filters.health"><option value="">{{ label('health', locale) }}</option><option value="on_track">on_track</option><option value="at_risk">at_risk</option><option value="off_track">off_track</option></select>
      <select v-model="filters.risk"><option value="">{{ label('risk', locale) }}</option><option value="low">low</option><option value="medium">medium</option><option value="high">high</option></select>
      <button class="btn" @click="load">{{ label('filters', locale) }}</button><button class="btn" @click="clearFilters">{{ label('clearFilters', locale) }}</button>
    </div>
    <form v-if="showForm" class="form" @submit.prevent="create">
      <input v-model="form.projectId" placeholder="Project ID" required />
      <input v-model="form.title" :placeholder="label('title', locale)" required />
      <PersonPicker v-model="form.owner" :users="directoryUsers" :placeholder="label('owner', locale)" />
      <input v-model="form.completionCriteria" :placeholder="label('criteria', locale)" required />
      <select v-model="form.riskLevel"><option value="low">low</option><option value="medium">medium</option><option value="high">high</option></select>
      <input v-model="form.plannedDate" type="date" />
      <div class="row"><button class="btn primary" type="submit">{{ label('save', locale) }}</button><button class="btn" type="button" @click="showForm = false">{{ label('cancel', locale) }}</button></div>
    </form>
    <table v-if="milestones.length">
      <thead><tr><th>{{ label('title', locale) }}</th><th>{{ label('status', locale) }}</th><th>{{ label('health', locale) }}</th><th>{{ label('risk', locale) }}</th><th>{{ label('owner', locale) }}</th><th>{{ label('plannedDate', locale) }}</th><th>{{ label('actions', locale) }}</th></tr></thead>
      <tbody><tr v-for="m in milestones" :key="m.id" class="clickable" tabindex="0" role="link" @click="go(m.id)" @keyup.enter="go(m.id)"><td>{{ m.title }}</td><td>{{ m.status }}</td><td>{{ m.healthStatus }}</td><td>{{ m.riskLevel || '-' }}</td><td>{{ m.owner }}</td><td>{{ m.plannedDate?.slice(0,10) || '-' }}</td><td><button class="btn sm" @click.stop="go(m.id)">{{ label('edit', locale) }}</button></td></tr></tbody>
    </table>
    <p v-else class="empty">{{ label('noData', locale) }}</p>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
h1 { margin: 0; }
.filters { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; background: var(--color-surface); padding: 12px; border-radius: 12px; margin-bottom: 16px; box-shadow: var(--shadow-sm); }
.filters input, .filters select { padding: 8px 10px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); }
.form { display: flex; flex-direction: column; gap: 10px; max-width: 480px; margin-bottom: 20px; background: var(--color-surface); padding: 20px; border-radius: 12px; box-shadow: var(--shadow-md); }
.form input, .form select { padding: 10px 12px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); }
.row { display: flex; gap: 10px; }
table { width: 100%; border-collapse: collapse; background: var(--color-surface); border-radius: 12px; overflow: hidden; box-shadow: var(--shadow-md); }
th { text-align: left; padding: 12px 16px; background: var(--color-surface-alt); font-size: .85rem; color: var(--color-text-muted); }
td { padding: 12px 16px; border-top: 1px solid #eaf0ed; }
.clickable { cursor: pointer; transition: background .1s; }
.clickable:hover { background: var(--color-surface-alt); }
</style>
