<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { Locale } from "../i18n";
import { label, apiFetch, can, gitlabAssignee, gitlabLabels, gitlabState, type Milestone, type ProjectDetailView, type WorkspaceRole } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const route = useRoute();
const router = useRouter();
const id = route.params.id as string;
const detail = ref<ProjectDetailView | null>(null);
const showMilestoneForm = ref(false);
const error = ref("");
const msForm = ref({ title: "", owner: "", completionCriteria: "", status: "not_started", healthStatus: "on_track", plannedDate: "", riskLevel: "low" });
const canManageProject = computed(() => can(currentRole.value, "manageProject"));
const canManageMilestone = computed(() => can(currentRole.value, "manageMilestone"));

async function load() {
  try { detail.value = await apiFetch<ProjectDetailView>(`/dashboard/project?id=${id}`); error.value = ""; } catch (err) { error.value = err instanceof Error ? err.message : String(err); }
}

onMounted(load);

async function createMilestone() {
  if (!canManageMilestone.value) return;
  await apiFetch("/milestones", { method: "POST", body: JSON.stringify({ ...msForm.value, projectId: id }) });
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

function goMilestone(id: string) { router.push({ name: "milestone-detail", params: { id } }); }
</script>

<template>
  <div class="page" v-if="detail">
    <div class="header">
      <h1>{{ detail.project.name }}</h1>
      <div class="health-actions">
        <button v-for="h in ['on_track','at_risk','off_track']" :key="h" class="btn sm" :class="h" :disabled="!canManageProject" @click="updateHealth(h)">{{ h }}</button>
      </div>
    </div>
    <p class="meta">{{ label('owner', locale) }}: {{ detail.project.owner }} &middot; {{ label('status', locale) }}: {{ detail.project.status }} &middot; {{ detail.project.healthStatus }}</p>
    <p v-if="error" class="error">{{ error }}</p>

    <div class="section">
      <div class="section-header">
        <h2>{{ label('milestones', locale) }}</h2>
        <button v-if="canManageMilestone" class="btn primary" @click="showMilestoneForm = !showMilestoneForm">{{ label('createMilestone', locale) }}</button>
        <span v-else class="empty">{{ label('noPermission', locale) }}</span>
      </div>
      <form v-if="showMilestoneForm" class="form" @submit.prevent="createMilestone">
        <input v-model="msForm.title" :placeholder="label('title', locale)" required />
        <input v-model="msForm.owner" :placeholder="label('owner', locale)" required />
        <input v-model="msForm.completionCriteria" :placeholder="label('criteria', locale)" required />
        <select v-model="msForm.riskLevel"><option value="low">low</option><option value="medium">medium</option><option value="high">high</option></select>
        <input v-model="msForm.plannedDate" type="date" :placeholder="label('plannedDate', locale)" />
        <div class="row"><button class="btn primary" type="submit">{{ label('save', locale) }}</button><button class="btn" type="button" @click="showMilestoneForm = false">{{ label('cancel', locale) }}</button></div>
      </form>
      <table v-if="detail.milestones.length">
        <thead><tr><th>{{ label('title', locale) }}</th><th>{{ label('status', locale) }}</th><th>{{ label('health', locale) }}</th><th>{{ label('owner', locale) }}</th><th>{{ label('plannedDate', locale) }}</th><th>{{ label('edit', locale) }}</th></tr></thead>
        <tbody>
          <tr v-for="m in detail.milestones" :key="m.id">
            <td class="clickable" @click="goMilestone(m.id)">{{ m.title }}</td><td>{{ m.status }}</td><td>{{ m.healthStatus }}</td><td>{{ m.owner }}</td><td>{{ m.plannedDate?.slice(0,10) || '-' }}</td>
            <td class="actions">
              <button class="btn sm" :disabled="!canManageMilestone" @click="transitionMilestone(m, 'active')">active</button>
              <button class="btn sm" :disabled="!canManageMilestone" @click="transitionMilestone(m, 'blocked')">blocked</button>
              <button class="btn sm" :disabled="!canManageMilestone" @click="transitionMilestone(m, 'completed')">completed</button>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">{{ label('noData', locale) }}</p>
    </div>

    <div class="section">
      <h2>{{ label('workItems', locale) }}</h2>
      <p v-if="!detail.workItems.length" class="empty">{{ label('noData', locale) }}</p>
      <div v-for="w in detail.workItems" :key="w.id" class="work-card">
        <strong>{{ w.title || w.id }}</strong> <span class="badge">{{ w.sourceType }}</span>
        <p>{{ label('status', locale) }}: {{ w.status || '-' }} &middot; {{ label('owner', locale) }}: {{ w.owner || '-' }}</p>
        <p v-if="w.sourceType === 'gitlab_issue'" class="gitlab-meta">GitLab: {{ gitlabState(w) || '-' }} &middot; {{ label('assignee', locale) }}: {{ gitlabAssignee(w) || '-' }} &middot; {{ label('labels', locale) }}: {{ gitlabLabels(w).join(', ') || '-' }} &middot; {{ label('lastSynced', locale) }}: {{ w.lastSyncedAt?.slice(0,10) || '-' }}</p>
        <a v-if="w.sourceType === 'gitlab_issue' && w.sourceUrl" :href="w.sourceUrl" target="_blank" rel="noreferrer">{{ label('openIssue', locale) }}</a>
      </div>
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
.header { display: flex; justify-content: space-between; align-items: center; gap: 14px; }
.meta { color: #4a7a6d; margin: 8px 0 24px; font-size: .9rem; }
.health-actions { display: flex; gap: 6px; }
.section { margin-top: 28px; }
.section-header { display: flex; justify-content: space-between; align-items: center; }
h2 { margin: 0 0 12px; font-size: 1.2rem; }
.form { display: flex; flex-direction: column; gap: 10px; max-width: 480px; margin-bottom: 16px; background: #fff; padding: 16px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
.form input, .form select { padding: 10px 12px; border: 1px solid #d1d9d6; border-radius: 8px; }
.row, .actions { display: flex; gap: 8px; flex-wrap: wrap; }
table { width: 100%; border-collapse: collapse; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
th { text-align: left; padding: 10px 14px; background: #f0f5f3; font-size: .82rem; color: #4a7a6d; }
td { padding: 10px 14px; border-top: 1px solid #eaf0ed; }
.clickable { cursor: pointer; font-weight: 700; color: #047857; }
.work-card { background: #fff; padding: 14px; border-radius: 10px; margin-bottom: 8px; box-shadow: 0 1px 4px rgba(0,0,0,.05); }
.work-card p { margin: 6px 0 0; color: #4a7a6d; }
.gitlab-meta { font-size: .84rem; }
.badge { display: inline-block; padding: 2px 8px; border-radius: 999px; background: #e0f2fe; color: #0369a1; font-size: .74rem; }
.btn { padding: 6px 14px; border-radius: 8px; border: 1px solid #d1d9d6; background: #fff; cursor: pointer; font-size: .85rem; }
.btn:disabled { opacity: .45; cursor: not-allowed; }
.btn.primary { background: #10352a; color: #fff; border-color: #10352a; }
.btn.sm { font-size: .78rem; padding: 4px 10px; }
.btn.on_track { background: #dcfce7; color: #15803d; border-color: #86efac; }
.btn.at_risk { background: #fef3c7; color: #b45309; border-color: #fcd34d; }
.btn.off_track { background: #fee2e2; color: #dc2626; border-color: #fca5a5; }
.empty { color: #6b8a80; }
.error { color: #b91c1c; background: #fee2e2; padding: 10px; border-radius: 8px; }
a { color: #047857; font-weight: 700; }
</style>
