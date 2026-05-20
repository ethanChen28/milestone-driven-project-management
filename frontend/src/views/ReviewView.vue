<script setup lang="ts">
import { inject, onMounted, ref, type Ref } from "vue";
import type { Locale } from "../i18n";
import { label, apiFetch, type WeeklyReviewView, type WeeklyUpdate } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const review = ref<WeeklyReviewView | null>(null);
const showForm = ref(false);
const form = ref({ projectId: "", milestoneId: "", summary: "", progress: "", risk: "", blockers: "", decisionsNeeded: "", nextSteps: "", author: "", week: "" });

async function load() {
  try { review.value = await apiFetch<WeeklyReviewView>("/review/weekly"); } catch { /* */ }
}

onMounted(load);

async function submit() {
  await apiFetch("/weekly-updates", { method: "POST", body: JSON.stringify(form.value) });
  showForm.value = false;
  form.value = { projectId: "", milestoneId: "", summary: "", progress: "", risk: "", blockers: "", decisionsNeeded: "", nextSteps: "", author: "", week: "" };
  await load();
}
</script>

<template>
  <div class="page">
    <div class="header">
      <h1>{{ label("review", locale) }}</h1>
      <button class="btn primary" @click="showForm = !showForm">{{ label("submitUpdate", locale) }}</button>
    </div>

    <form v-if="showForm" class="form" @submit.prevent="submit">
      <input v-model="form.projectId" placeholder="Project ID" required />
      <input v-model="form.milestoneId" placeholder="Milestone ID" />
      <input v-model="form.week" :placeholder="label('week', locale)" />
      <input v-model="form.author" :placeholder="label('author', locale)" required />
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
          {{ m.title }} — {{ label('owner', locale) }}: {{ m.owner }} &middot; {{ label('plannedDate', locale) }}: {{ m.plannedDate?.slice(0,10) }}
        </div>
      </section>

      <section v-if="review.blockedMilestones.length" class="section">
        <h2>{{ label('blocked', locale) }} ({{ review.blockedMilestones.length }})</h2>
        <div v-for="m in review.blockedMilestones" :key="m.id" class="alert-card warn">
          {{ m.title }} — {{ label('owner', locale) }}: {{ m.owner }}
        </div>
      </section>

      <section class="section">
        <h2>{{ label('summary', locale) }}</h2>
        <p v-if="!review.updates.length" class="empty">{{ label('noData', locale) }}</p>
        <div v-for="u in review.updates" :key="u.id" class="update-card">
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
.form { display: flex; flex-direction: column; gap: 10px; max-width: 560px; margin-bottom: 20px; background: #fff; padding: 20px; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
.form input, .form textarea { padding: 10px 12px; border: 1px solid #d1d9d6; border-radius: 8px; font-family: inherit; }
.row { display: flex; gap: 10px; }
.alert-card { padding: 12px 16px; border-radius: 10px; margin-bottom: 8px; }
.alert-card.danger { background: #fee2e2; color: #991b1b; }
.alert-card.warn { background: #fef3c7; color: #92400e; }
.update-card { background: #fff; padding: 14px; border-radius: 10px; margin-bottom: 8px; box-shadow: 0 1px 4px rgba(0,0,0,.05); }
.update-header { font-size: .9rem; }
.update-card p { margin: 6px 0 0; color: #4a7a6d; }
.btn { padding: 8px 18px; border-radius: 8px; border: 1px solid #d1d9d6; background: #fff; cursor: pointer; }
.btn.primary { background: #10352a; color: #fff; border-color: #10352a; }
.empty { color: #6b8a80; }
</style>
