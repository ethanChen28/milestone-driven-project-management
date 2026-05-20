<script setup lang="ts">
import { inject, onMounted, ref, type Ref } from "vue";
import type { Locale } from "../i18n";
import { label, apiFetch, type Milestone } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const milestones = ref<Milestone[]>([]);
const showForm = ref(false);
const form = ref({ projectId: "", title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "" });

async function load() {
  try { milestones.value = await apiFetch<Milestone[]>("/milestones"); } catch { milestones.value = []; }
}

onMounted(load);

async function create() {
  await apiFetch("/milestones", { method: "POST", body: JSON.stringify(form.value) });
  showForm.value = false;
  form.value = { projectId: "", title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "" };
  await load();
}
</script>

<template>
  <div class="page">
    <div class="header">
      <h1>{{ label("milestones", locale) }}</h1>
      <button class="btn primary" @click="showForm = !showForm">{{ label("createMilestone", locale) }}</button>
    </div>
    <form v-if="showForm" class="form" @submit.prevent="create">
      <input v-model="form.projectId" placeholder="Project ID" required />
      <input v-model="form.title" :placeholder="label('title', locale)" required />
      <input v-model="form.owner" :placeholder="label('owner', locale)" required />
      <input v-model="form.completionCriteria" :placeholder="label('criteria', locale)" required />
      <input v-model="form.plannedDate" type="date" />
      <div class="row"><button class="btn primary" type="submit">{{ label('save', locale) }}</button><button class="btn" type="button" @click="showForm = false">{{ label('cancel', locale) }}</button></div>
    </form>
    <table v-if="milestones.length">
      <thead><tr><th>{{ label('title', locale) }}</th><th>{{ label('status', locale) }}</th><th>{{ label('health', locale) }}</th><th>{{ label('owner', locale) }}</th><th>{{ label('plannedDate', locale) }}</th></tr></thead>
      <tbody><tr v-for="m in milestones" :key="m.id"><td>{{ m.title }}</td><td>{{ m.status }}</td><td>{{ m.healthStatus }}</td><td>{{ m.owner }}</td><td>{{ m.plannedDate?.slice(0,10) || '-' }}</td></tr></tbody>
    </table>
    <p v-else class="empty">{{ label('noData', locale) }}</p>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
h1 { margin: 0; }
.form { display: flex; flex-direction: column; gap: 10px; max-width: 480px; margin-bottom: 20px; background: #fff; padding: 20px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
.form input { padding: 10px 12px; border: 1px solid #d1d9d6; border-radius: 8px; }
.row { display: flex; gap: 10px; }
table { width: 100%; border-collapse: collapse; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
th { text-align: left; padding: 12px 16px; background: #f0f5f3; font-size: .85rem; color: #4a7a6d; }
td { padding: 12px 16px; border-top: 1px solid #eaf0ed; }
.btn { padding: 8px 18px; border-radius: 8px; border: 1px solid #d1d9d6; background: #fff; cursor: pointer; }
.btn.primary { background: #10352a; color: #fff; border-color: #10352a; }
.empty { color: #6b8a80; }
</style>
