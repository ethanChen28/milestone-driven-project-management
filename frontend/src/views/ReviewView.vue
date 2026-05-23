<script setup lang="ts">
import { computed, inject, onMounted, ref, watch, type Ref } from "vue";
import type { Locale } from "../i18n";
import PersonPicker from "../components/PersonPicker.vue";
import { label, apiFetch, can, listUsers, type UserProfile, type Milestone, type Project, type WeeklyReviewView, type WorkspaceRole } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const review = ref<WeeklyReviewView | null>(null);
const projects = ref<Project[]>([]);
const directoryUsers = ref<UserProfile[]>([]);
const milestones = ref<Milestone[]>([]);
const showForm = ref(false);
const error = ref("");
const canSubmit = computed(() => can(currentRole.value, "submitUpdate"));
const filters = ref({ owner: "", health: "", risk: "" });
const form = ref({ projectId: "", milestoneId: "", summary: "", progress: "", risk: "", blockers: "", decisionsNeeded: "", nextSteps: "", author: "", week: defaultWeek() });
const projectMilestones = computed(() => milestones.value.filter((m) => !form.value.projectId || m.projectId === form.value.projectId));
const sortedUpdates = computed(() =>
  [...(review.value?.updates ?? [])].sort((left, right) => String(right.id).localeCompare(String(left.id))),
);

function defaultWeek() {
  const now = new Date();
  const start = new Date(Date.UTC(now.getFullYear(), 0, 1));
  const days = Math.floor((Date.UTC(now.getFullYear(), now.getMonth(), now.getDate()) - start.getTime()) / 86400000);
  return `${now.getFullYear()}-W${String(Math.ceil((days + start.getUTCDay() + 1) / 7)).padStart(2, "0")}`;
}

function query() {
  const params = new URLSearchParams();
  Object.entries(filters.value).forEach(([key, value]) => { if (value) params.set(key, value); });
  return params.toString() ? `?${params.toString()}` : "";
}

async function load() {
  try {
    const [reviewData, projectData, milestoneData] = await Promise.all([
      apiFetch<WeeklyReviewView>(`/review/weekly${query()}`),
      apiFetch<Project[]>("/projects"),
      apiFetch<Milestone[]>("/milestones"),
    ]);
    review.value = reviewData;
    projects.value = projectData;
    milestones.value = milestoneData;
    directoryUsers.value = await listUsers();
    error.value = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}

onMounted(load);
watch(() => form.value.projectId, () => { form.value.milestoneId = ""; });

async function submit() {
  if (!canSubmit.value) return;
  try {
    await apiFetch("/weekly-updates", { method: "POST", body: JSON.stringify(form.value) });
    showForm.value = false;
    form.value = { projectId: "", milestoneId: "", summary: "", progress: "", risk: "", blockers: "", decisionsNeeded: "", nextSteps: "", author: "", week: defaultWeek() };
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  }
}

function clearFilters() {
  filters.value = { owner: "", health: "", risk: "" };
  load();
}
</script>

<template>
  <div class="page">
    <div class="header">
      <h1>{{ label("review", locale) }}</h1>
      <button v-if="canSubmit" class="btn primary" @click="showForm = !showForm">{{ label("submitUpdate", locale) }}</button>
      <span v-else class="permission-hint">{{ label('noPermission', locale) }}<small>{{ label("needContributor", locale) }}</small></span>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <div class="filters">
      <strong>{{ label('filters', locale) }}</strong>
      <input v-model="filters.owner" :placeholder="label('owner', locale)" />
      <select v-model="filters.health"><option value="">{{ label('health', locale) }}</option><option value="on_track">on_track</option><option value="at_risk">at_risk</option><option value="off_track">off_track</option></select>
      <select v-model="filters.risk"><option value="">{{ label('risk', locale) }}</option><option value="low">low</option><option value="medium">medium</option><option value="high">high</option></select>
      <button class="btn" @click="load">{{ label('filters', locale) }}</button>
      <button class="btn" @click="clearFilters">{{ label('clearFilters', locale) }}</button>
    </div>

    <form v-if="showForm" class="form" @submit.prevent="submit">
      <select v-model="form.projectId" required>
        <option value="">{{ label('projects', locale) }}</option>
        <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select>
      <select v-model="form.milestoneId">
        <option value="">{{ label('milestones', locale) }}</option>
        <option v-for="m in projectMilestones" :key="m.id" :value="m.id">{{ m.title }}</option>
      </select>
      <input v-model="form.week" :placeholder="label('week', locale)" />
      <PersonPicker v-model="form.author" :users="directoryUsers" :placeholder="label('author', locale)" />
      <textarea v-model="form.summary" :placeholder="label('summary', locale)" rows="2" required />
      <textarea v-model="form.progress" :placeholder="label('progress', locale)" rows="2" />
      <textarea v-model="form.risk" :placeholder="label('risk', locale)" rows="2" />
      <textarea v-model="form.blockers" :placeholder="label('blockers', locale)" rows="2" />
      <textarea v-model="form.decisionsNeeded" :placeholder="label('decisionsNeeded', locale)" rows="2" />
      <textarea v-model="form.nextSteps" :placeholder="label('nextSteps', locale)" rows="2" />
      <div class="row"><button class="btn primary" type="submit">{{ label('save', locale) }}</button><button class="btn" type="button" @click="showForm = false">{{ label('cancel', locale) }}</button></div>
    </form>

    <template v-if="review">
      <section v-if="review.delayedMilestones.length" class="section">
        <h2>{{ label('delayed', locale) }} ({{ review.delayedMilestones.length }})</h2>
        <div v-for="m in review.delayedMilestones" :key="m.id" class="alert-card danger">
          {{ m.title }} — {{ label('owner', locale) }}: {{ m.owner }} &middot; {{ label('plannedDate', locale) }}: {{ m.plannedDate?.slice(0,10) }} &middot; {{ label('risk', locale) }}: {{ m.riskLevel || '-' }}
        </div>
      </section>

      <section v-if="review.blockedMilestones.length" class="section">
        <h2>{{ label('blocked', locale) }} ({{ review.blockedMilestones.length }})</h2>
        <div v-for="m in review.blockedMilestones" :key="m.id" class="alert-card warn">
          {{ m.title }} — {{ label('owner', locale) }}: {{ m.owner }} &middot; {{ label('risk', locale) }}: {{ m.riskLevel || '-' }}
        </div>
      </section>

      <section class="section">
        <h2>{{ label('summary', locale) }}</h2>
        <p v-if="!sortedUpdates.length" class="empty">{{ label('noData', locale) }}</p>
        <div v-for="u in sortedUpdates" :key="u.id" class="update-card">
          <div class="update-header"><strong>{{ u.week }}</strong> — {{ u.author }}</div>
          <p>{{ u.summary }}</p>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
h1 { margin: 0; }
h2 { font-size: 1.1rem; margin: 0 0 10px; }
.section { margin-top: 24px; }
.filters { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; background: var(--color-surface); padding: 12px; border-radius: 12px; margin-bottom: 16px; box-shadow: var(--shadow-sm); }
.filters input, .filters select { padding: 8px 10px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); }
.form { display: flex; flex-direction: column; gap: 10px; max-width: 560px; margin-bottom: 20px; background: var(--color-surface); padding: 20px; border-radius: 12px; box-shadow: var(--shadow-md); }
.form input, .form textarea, .form select { padding: 10px 12px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); font-family: inherit; }
.row { display: flex; gap: 10px; }
.alert-card { padding: 12px 16px; border-radius: var(--radius-md); margin-bottom: 8px; }
.alert-card.danger { background: var(--color-danger-bg); color: #991b1b; }
.alert-card.warn { background: var(--color-warning-bg); color: #92400e; }
.update-card { background: var(--color-surface); padding: 14px; border-radius: var(--radius-md); margin-bottom: 8px; box-shadow: var(--shadow-sm); }
.update-header { font-size: .9rem; }
.update-card p { margin: 6px 0 0; color: var(--color-text-muted); }
</style>
