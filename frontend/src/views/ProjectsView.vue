<script setup lang="ts">
import { inject, onMounted, ref, type Ref } from "vue";
import { useRouter } from "vue-router";
import type { Locale } from "../i18n";
import { label, apiFetch, type Project } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const router = useRouter();
const projects = ref<Project[]>([]);
const showForm = ref(false);
const form = ref({ name: "", objective: "", owner: "", status: "active", healthStatus: "on_track", priority: "P1", projectType: "product", targetStartDate: "", targetEndDate: "" });

async function load() {
  try { projects.value = await apiFetch<Project[]>("/projects"); } catch { projects.value = []; }
}

onMounted(load);

async function create() {
  await apiFetch("/projects", { method: "POST", body: JSON.stringify(form.value) });
  showForm.value = false;
  form.value = { name: "", objective: "", owner: "", status: "active", healthStatus: "on_track", priority: "P1", projectType: "product", targetStartDate: "", targetEndDate: "" };
  await load();
}

function go(id: string) { router.push({ name: "project-detail", params: { id } }); }
</script>

<template>
  <div class="page">
    <div class="header">
      <h1>{{ label("projects", locale) }}</h1>
      <button class="btn primary" @click="showForm = !showForm">{{ label("createProject", locale) }}</button>
    </div>
    <form v-if="showForm" class="form" @submit.prevent="create">
      <input v-model="form.name" :placeholder="label('name', locale)" required />
      <input v-model="form.objective" :placeholder="label('objective', locale)" />
      <input v-model="form.owner" :placeholder="label('owner', locale)" required />
      <select v-model="form.healthStatus">
        <option value="on_track">on_track</option><option value="at_risk">at_risk</option><option value="off_track">off_track</option>
      </select>
      <select v-model="form.priority">
        <option value="P0">P0</option><option value="P1">P1</option><option value="P2">P2</option><option value="P3">P3</option>
      </select>
      <div class="row">
        <button class="btn primary" type="submit">{{ label("save", locale) }}</button>
        <button class="btn" type="button" @click="showForm = false">{{ label("cancel", locale) }}</button>
      </div>
    </form>
    <table v-if="projects.length">
      <thead><tr><th>{{ label('name', locale) }}</th><th>{{ label('owner', locale) }}</th><th>{{ label('status', locale) }}</th><th>{{ label('health', locale) }}</th><th>{{ label('priority', locale) }}</th></tr></thead>
      <tbody>
        <tr v-for="p in projects" :key="p.id" class="clickable" @click="go(p.id)">
          <td>{{ p.name }}</td><td>{{ p.owner }}</td><td><span class="badge" :class="p.status">{{ p.status }}</span></td>
          <td><span class="badge health" :class="p.healthStatus">{{ p.healthStatus }}</span></td><td>{{ p.priority }}</td>
        </tr>
      </tbody>
    </table>
    <p v-else class="empty">{{ label("noData", locale) }}</p>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
h1 { margin: 0; }
.form { display: flex; flex-direction: column; gap: 10px; max-width: 480px; margin-bottom: 20px; background: #fff; padding: 20px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
.form input, .form select { padding: 10px 12px; border: 1px solid #d1d9d6; border-radius: 8px; font-size: .92rem; }
.row { display: flex; gap: 10px; }
table { width: 100%; border-collapse: collapse; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
th { text-align: left; padding: 12px 16px; background: #f0f5f3; font-size: .85rem; color: #4a7a6d; }
td { padding: 12px 16px; border-top: 1px solid #eaf0ed; }
.clickable { cursor: pointer; transition: background .1s; }
.clickable:hover { background: #f0f5f3; }
.badge { display: inline-block; padding: 3px 10px; border-radius: 999px; font-size: .78rem; font-weight: 600; }
.badge.active { background: #dcfce7; color: #15803d; }
.badge.archived { background: #f3f4f6; color: #6b7280; }
.badge.health.on_track { background: #dcfce7; color: #15803d; }
.badge.health.at_risk { background: #fef3c7; color: #b45309; }
.badge.health.off_track { background: #fee2e2; color: #dc2626; }
.btn { padding: 8px 18px; border-radius: 8px; border: 1px solid #d1d9d6; background: #fff; cursor: pointer; font-size: .9rem; }
.btn.primary { background: #10352a; color: #fff; border-color: #10352a; }
.empty { color: #6b8a80; }
</style>
