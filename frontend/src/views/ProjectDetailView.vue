<script setup lang="ts">
import { inject, onMounted, ref, type Ref } from "vue";
import { useRoute } from "vue-router";
import type { Locale } from "../i18n";
import { label, apiFetch, type ProjectDetailView, type Milestone } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const route = useRoute();
const id = route.params.id as string;
const detail = ref<ProjectDetailView | null>(null);
const showMilestoneForm = ref(false);
const msForm = ref({ title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "" });

async function load() {
  try { detail.value = await apiFetch<ProjectDetailView>(`/dashboard/project?id=${id}`); } catch { /* */ }
}

onMounted(load);

async function createMilestone() {
  await apiFetch("/milestones", { method: "POST", body: JSON.stringify({ ...msForm.value, projectId: id }) });
  showMilestoneForm.value = false;
  msForm.value = { title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "" };
  await load();
}

async function updateHealth(h: string) {
  await apiFetch(`/projects?id=${id}`, { method: "PUT", body: JSON.stringify({ ...detail.value!.project, healthStatus: h }) });
  await load();
}
</script>

<template>
  <div class="page" v-if="detail">
    <div class="header">
      <h1>{{ detail.project.name }}</h1>
      <div class="health-actions">
        <button v-for="h in ['on_track','at_risk','off_track']" :key="h" class="btn sm" :class="h" @click="updateHealth(h)">{{ h }}</button>
      </div>
    </div>
    <p class="meta">{{ label('owner', locale) }}: {{ detail.project.owner }} &middot; {{ label('status', locale) }}: {{ detail.project.status }} &middot; {{ detail.project.healthStatus }}</p>

    <div class="section">
      <div class="section-header">
        <h2>{{ label('milestones', locale) }}</h2>
        <button class="btn primary" @click="showMilestoneForm = !showMilestoneForm">{{ label('createMilestone', locale) }}</button>
      </div>
      <form v-if="showMilestoneForm" class="form" @submit.prevent="createMilestone">
        <input v-model="msForm.title" :placeholder="label('title', locale)" required />
        <input v-model="msForm.owner" :placeholder="label('owner', locale)" required />
        <input v-model="msForm.completionCriteria" :placeholder="label('criteria', locale)" required />
        <input v-model="msForm.plannedDate" type="date" :placeholder="label('plannedDate', locale)" />
        <div class="row"><button class="btn primary" type="submit">{{ label('save', locale) }}</button><button class="btn" type="button" @click="showMilestoneForm = false">{{ label('cancel', locale) }}</button></div>
      </form>
      <table v-if="detail.milestones.length">
        <thead><tr><th>{{ label('title', locale) }}</th><th>{{ label('status', locale) }}</th><th>{{ label('health', locale) }}</th><th>{{ label('owner', locale) }}</th><th>{{ label('plannedDate', locale) }}</th></tr></thead>
        <tbody><tr v-for="m in detail.milestones" :key="m.id"><td>{{ m.title }}</td><td>{{ m.status }}</td><td>{{ m.healthStatus }}</td><td>{{ m.owner }}</td><td>{{ m.plannedDate?.slice(0,10) || '-' }}</td></tr></tbody>
      </table>
      <p v-else class="empty">{{ label('noData', locale) }}</p>
    </div>

    <div class="section">
      <h2>{{ label('summary', locale) }}</h2>
      <p>{{ detail.project.objective }}</p>
    </div>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
h1 { margin: 0; }
.header { display: flex; justify-content: space-between; align-items: center; }
.meta { color: #4a7a6d; margin: 8px 0 24px; font-size: .9rem; }
.health-actions { display: flex; gap: 6px; }
.section { margin-top: 28px; }
.section-header { display: flex; justify-content: space-between; align-items: center; }
h2 { margin: 0 0 12px; font-size: 1.2rem; }
.form { display: flex; flex-direction: column; gap: 10px; max-width: 480px; margin-bottom: 16px; background: #fff; padding: 16px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
.form input { padding: 10px 12px; border: 1px solid #d1d9d6; border-radius: 8px; }
.row { display: flex; gap: 10px; }
table { width: 100%; border-collapse: collapse; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
th { text-align: left; padding: 10px 14px; background: #f0f5f3; font-size: .82rem; color: #4a7a6d; }
td { padding: 10px 14px; border-top: 1px solid #eaf0ed; }
.btn { padding: 6px 14px; border-radius: 8px; border: 1px solid #d1d9d6; background: #fff; cursor: pointer; font-size: .85rem; }
.btn.primary { background: #10352a; color: #fff; border-color: #10352a; }
.btn.sm { font-size: .78rem; padding: 4px 10px; }
.btn.on_track { background: #dcfce7; color: #15803d; border-color: #86efac; }
.btn.at_risk { background: #fef3c7; color: #b45309; border-color: #fcd34d; }
.btn.off_track { background: #fee2e2; color: #dc2626; border-color: #fca5a5; }
.empty { color: #6b8a80; }
</style>
