<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { Locale } from "../i18n";
import PersonPicker from "../components/PersonPicker.vue";
import { dateInputToIso, isoToDateInput, label, apiFetch, can, listUsers, type UserProfile, type Project, type WorkspaceRole } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const route = useRoute();
const router = useRouter();
const id = route.params.id as string;
const loading = ref(true);
const error = ref("");
const directoryUsers = ref<UserProfile[]>([]);
const canEdit = computed(() => can(currentRole.value, "manageProject"));

const form = ref<Partial<Project & { targetStartDate: string; targetEndDate: string }>>({});

async function load() {
  try {
    const project = await apiFetch<Project>(`/projects?id=${id}`);
    form.value = {
      ...project,
      targetStartDate: isoToDateInput(project.targetStartDate),
      targetEndDate: isoToDateInput(project.targetEndDate),
    };
    directoryUsers.value = await listUsers();
    error.value = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

onMounted(load);

async function save() {
  if (!canEdit.value) return;
  try {
    await apiFetch(`/projects?id=${id}`, {
      method: "PUT",
      body: JSON.stringify({
        ...form.value,
        targetStartDate: dateInputToIso(form.value.targetStartDate),
        targetEndDate: dateInputToIso(form.value.targetEndDate),
      }),
    });
    router.push({ name: "project-detail", params: { id } });
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}

function cancel() {
  router.push({ name: "project-detail", params: { id } });
}

function copy(key: string, zh: string, en: string) {
  const translated = label(key, locale.value);
  if (translated !== key) return translated;
  return locale.value === "zh-CN" ? zh : en;
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <p class="crumb">
          <RouterLink :to="{ name: 'projects' }">{{ label('projects', locale) }}</RouterLink>
          /
          <RouterLink :to="{ name: 'project-detail', params: { id } }">{{ form.name || '...' }}</RouterLink>
          /
          {{ copy('editProject', '编辑项目', 'Edit Project') }}
        </p>
        <h1>{{ copy('editProject', '编辑项目', 'Edit Project') }}</h1>
      </div>
    </div>

    <p v-if="loading" class="empty">{{ copy('loading', '加载中...', 'Loading...') }}</p>
    <p v-else-if="!canEdit" class="empty">{{ label('noPermission', locale) }}</p>

    <form v-else class="form" @submit.prevent="save">
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <div class="form-grid">
        <label class="field span-2">
          <span>{{ label('name', locale) }}</span>
          <input v-model="form.name" :placeholder="label('name', locale)" required />
        </label>
        <label class="field span-2">
          <span>{{ label('objective', locale) }}</span>
          <textarea v-model="form.objective" :placeholder="label('objective', locale)" rows="3" />
        </label>
        <label class="field">
          <span>{{ label('owner', locale) }}</span>
          <PersonPicker v-model="form.owner!" :users="directoryUsers" :placeholder="label('owner', locale)" />
        </label>
        <label class="field">
          <span>{{ copy('participants', '参与者', 'Participants') }}</span>
          <input :value="(form.participants || []).join(', ')" disabled />
        </label>
        <label class="field">
          <span>{{ label('projectType', locale) }}</span>
          <select v-model="form.projectType">
            <option value="product">product</option>
            <option value="platform">platform</option>
            <option value="operations">operations</option>
            <option value="research">research</option>
          </select>
        </label>
        <label class="field">
          <span>{{ label('status', locale) }}</span>
          <select v-model="form.status">
            <option value="active">active</option>
            <option value="done">done</option>
            <option value="archived">archived</option>
          </select>
        </label>
        <label class="field">
          <span>{{ label('health', locale) }}</span>
          <select v-model="form.healthStatus">
            <option value="on_track">on_track</option>
            <option value="at_risk">at_risk</option>
            <option value="off_track">off_track</option>
          </select>
        </label>
        <label class="field">
          <span>{{ label('priority', locale) }}</span>
          <select v-model="form.priority">
            <option value="P0">P0</option>
            <option value="P1">P1</option>
            <option value="P2">P2</option>
            <option value="P3">P3</option>
          </select>
        </label>
        <label class="field">
          <span>{{ label('startDate', locale) }}</span>
          <input v-model="form.targetStartDate" type="date" />
        </label>
        <label class="field">
          <span>{{ label('endDate', locale) }}</span>
          <input v-model="form.targetEndDate" type="date" />
        </label>
      </div>
      <div class="row form-actions">
        <button class="btn primary" type="submit">{{ label('save', locale) }}</button>
        <button class="btn" type="button" @click="cancel">{{ label('cancel', locale) }}</button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
.page-header { margin-bottom: 24px; }
.crumb { margin: 0 0 6px; color: var(--color-text-muted); font-size: .9rem; }
.crumb a { color: var(--color-primary-light); font-weight: 700; text-decoration: none; }
h1 { margin: 0; font-size: 1.8rem; letter-spacing: -.04em; }
.form { display: grid; gap: 16px; background: linear-gradient(135deg, #ffffff 0%, #f7fbf8 100%); padding: 20px; border: 1px solid #e1ebe6; border-radius: 20px; box-shadow: var(--shadow-md); }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.span-2 { grid-column: span 2; }
.field { display: grid; gap: 5px; }
.field span { font-size: .82rem; color: var(--color-text-muted); }
.form input, .form select, .form textarea { width: 100%; box-sizing: border-box; padding: 10px 12px; border: 1px solid var(--color-border); border-radius: var(--radius-md); color: var(--color-text); font-family: inherit; font-size: .9rem; background: var(--color-surface); }
.form textarea { resize: vertical; }
.row { display: flex; gap: 10px; }
.form-actions { justify-content: flex-end; border-top: 1px solid #e7efeb; padding-top: 14px; }
.btn { padding: 8px 18px; border-radius: var(--radius-sm); border: 1px solid var(--color-border); background: var(--color-surface); cursor: pointer; }
.btn.primary { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
@media (max-width: 640px) { .form-grid { grid-template-columns: 1fr; } .span-2 { grid-column: auto; } }
</style>
