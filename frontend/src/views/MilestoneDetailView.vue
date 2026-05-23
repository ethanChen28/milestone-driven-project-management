<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { Locale } from "../i18n";
import { dateInputToIso, isoToDateInput, label, apiFetch, can, gitlabAssignee, gitlabLabels, gitlabState, type Milestone, type MilestoneDetailView, type WorkspaceRole } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const route = useRoute();
const router = useRouter();
const id = route.params.id as string;
const detail = ref<MilestoneDetailView | null>(null);
const error = ref("");
const editForm = ref<Partial<Milestone>>({});
const canEdit = computed(() => can(currentRole.value, "manageMilestone"));
const canComplete = computed(() => currentRole.value !== "contributor" && canEdit.value);
const criteriaItems = computed(() =>
  (detail.value?.milestone.completionCriteria ?? "")
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean),
);

async function load() {
  try {
    detail.value = await apiFetch<MilestoneDetailView>(`/dashboard/milestone?id=${id}`);
    editForm.value = {
      ...detail.value.milestone,
      plannedDate: isoToDateInput(detail.value.milestone.plannedDate),
      forecastDate: isoToDateInput(detail.value.milestone.forecastDate),
    };
    error.value = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}

onMounted(load);

function createTask() {
  const milestone = detail.value?.milestone;
  if (!milestone) return;
  router.push({ name: "task-create", query: { projectId: milestone.projectId, milestoneId: milestone.id } });
}

async function saveMilestone() {
  if (!detail.value) return;
  try {
    await apiFetch(`/milestones?id=${detail.value.milestone.id}`, {
      method: "PUT",
      body: JSON.stringify({
        ...editForm.value,
        plannedDate: dateInputToIso(editForm.value.plannedDate),
        forecastDate: dateInputToIso(editForm.value.forecastDate),
      }),
    });
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}
</script>

<template>
  <div class="page" v-if="detail">
    <div class="page-header">
      <h1>{{ detail.milestone.title }}</h1>
      <button class="btn primary" @click="createTask">{{ label('newTask', locale) }}</button>
    </div>
    <p class="meta">{{ label('status', locale) }}: {{ detail.milestone.status }} &middot; {{ label('health', locale) }}: {{ detail.milestone.healthStatus }} &middot; {{ label('owner', locale) }}: {{ detail.milestone.owner }}</p>
    <div class="progress-bar-container">
      <div class="progress-bar-track">
        <div class="progress-bar-fill" :class="detail.milestone.healthStatus" :style="{ width: (detail.milestone.progressPercent || 0) + '%' }"></div>
      </div>
      <span class="progress-label">{{ detail.milestone.progressPercent || 0 }}%</span>
    </div>
    <div v-if="criteriaItems.length" class="criteria-card">
      <strong>{{ label('criteria', locale) }}</strong>
      <ul class="criteria-list">
        <li v-for="item in criteriaItems" :key="item"><input type="checkbox" disabled /> <span>{{ item }}</span></li>
      </ul>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <form class="form lifecycle" @submit.prevent="saveMilestone">
      <div class="form-grid">
        <label>{{ label('status', locale) }}<select v-model="editForm.status" :disabled="!canEdit"><option value="not_started">not_started</option><option value="active">active</option><option value="blocked">blocked</option><option v-if="canComplete" value="completed">completed</option><option value="cancelled">cancelled</option></select></label>
        <label>{{ label('health', locale) }}<select v-model="editForm.healthStatus" :disabled="!canEdit"><option value="on_track">on_track</option><option value="at_risk">at_risk</option><option value="off_track">off_track</option><option value="done">done</option></select></label>
        <label>{{ label('progressPercent', locale) }}<input v-model.number="editForm.progressPercent" type="number" min="0" max="100" :disabled="!canEdit" /></label>
        <label>{{ label('risk', locale) }}<select v-model="editForm.riskLevel" :disabled="!canEdit"><option value="low">low</option><option value="medium">medium</option><option value="high">high</option></select></label>
        <label>{{ label('plannedDate', locale) }}<input v-model="editForm.plannedDate" type="date" :disabled="!canEdit" /></label>
        <label>{{ label('forecastDate', locale) }}<input v-model="editForm.forecastDate" type="date" :disabled="!canEdit" /></label>
      </div>
      <label>{{ label('criteria', locale) }}<textarea v-model="editForm.completionCriteria" rows="2" :disabled="!canEdit" /></label>
      <label>{{ label('dependencySummary', locale) }}<textarea v-model="editForm.dependencySummary" rows="2" :disabled="!canEdit" /></label>
      <button v-if="canEdit" class="btn primary" type="submit">{{ label('save', locale) }}</button>
      <p v-else class="empty">{{ label('noPermission', locale) }}</p>
    </form>

    <div class="section">
      <h2>{{ label('workItems', locale) }}</h2>
      <p v-if="!detail.workItems.length" class="empty">{{ label('noData', locale) }}</p>
      <div v-for="w in detail.workItems" :key="w.id" class="work-card">
        <div><strong>{{ w.title || w.id }}</strong><span class="badge">{{ w.sourceType }}</span></div>
        <p>{{ label('status', locale) }}: {{ w.status || '-' }} &middot; {{ label('owner', locale) }}: {{ w.owner || '-' }}</p>
        <p v-if="w.sourceType === 'gitlab_issue'" class="gitlab-meta">
          GitLab: {{ gitlabState(w) || '-' }} &middot; {{ label('assignee', locale) }}: {{ gitlabAssignee(w) || '-' }} &middot; {{ label('labels', locale) }}: {{ gitlabLabels(w).join(', ') || '-' }} &middot; {{ label('lastSynced', locale) }}: {{ w.lastSyncedAt?.slice(0, 10) || '-' }}
        </p>
        <a v-if="w.sourceType === 'gitlab_issue' && w.sourceUrl" :href="w.sourceUrl" target="_blank" rel="noreferrer">{{ label('openIssue', locale) }}</a>
      </div>
    </div>

    <div class="section">
      <h2>{{ label('summary', locale) }}</h2>
      <p v-if="!detail.updates.length" class="empty">{{ label('noData', locale) }}</p>
      <div v-for="u in detail.updates" :key="u.id" class="update-card">
        <strong>{{ u.week }}</strong> — {{ u.author }}
        <p>{{ u.summary }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
h1 { margin: 0; }
.meta { color: var(--color-text-muted); margin: 8px 0 12px; }
.progress-bar-container { display: flex; align-items: center; gap: 12px; margin-bottom: 24px; }
.progress-bar-track { flex: 1; height: 10px; border-radius: 5px; background: #e0ebe6; overflow: hidden; }
.progress-bar-fill { height: 100%; border-radius: 5px; transition: width .3s ease; }
.progress-bar-fill.on_track, .progress-bar-fill.done { background: #22c55e; }
.progress-bar-fill.at_risk { background: #f59e0b; }
.progress-bar-fill.off_track { background: #ef4444; }
.progress-label { font-size: .85rem; font-weight: 700; color: var(--color-text-muted); min-width: 40px; }
.section { margin-top: 24px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.section-header h2 { margin: 0; }
h2 { font-size: 1.1rem; margin: 0 0 10px; }
.criteria-card { background: var(--color-surface); padding: 14px; border-radius: var(--radius-md); margin: 0 0 16px; box-shadow: var(--shadow-sm); }
.criteria-card ul { margin: 10px 0 0; padding: 0; list-style: none; display: grid; gap: 8px; }
.criteria-card li { display: flex; align-items: center; gap: 8px; color: #315b50; }
.form { display: grid; gap: 12px; background: var(--color-surface); padding: 16px; border-radius: 12px; box-shadow: var(--shadow-sm); }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
label { display: grid; gap: 5px; font-size: .82rem; color: var(--color-text-muted); }
select, input, textarea { padding: 9px 10px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); font-family: inherit; color: var(--color-text); }
.update-card, .work-card { background: var(--color-surface); padding: 14px; border-radius: var(--radius-md); margin-bottom: 8px; box-shadow: var(--shadow-sm); }
.update-card p, .work-card p { margin: 6px 0 0; color: var(--color-text-muted); }
.gitlab-meta { font-size: .84rem; }
.badge { margin-left: 8px; display: inline-block; padding: 2px 8px; border-radius: var(--radius-full); background: #e0f2fe; color: #0369a1; font-size: .74rem; }
.btn { justify-self: start; padding: 8px 18px; border-radius: var(--radius-sm); border: 1px solid var(--color-border); background: var(--color-surface); cursor: pointer; }
.btn.primary { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
a { color: var(--color-primary-light); font-weight: 700; }
@media (max-width: 720px) { .form-grid { grid-template-columns: 1fr; } }
</style>
