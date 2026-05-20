<script setup lang="ts">
import { inject, onMounted, ref, type Ref } from "vue";
import { useRoute } from "vue-router";
import type { Locale } from "../i18n";
import { label, apiFetch, type MilestoneDetailView } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const route = useRoute();
const id = route.params.id as string;
const detail = ref<MilestoneDetailView | null>(null);

onMounted(async () => {
  try { detail.value = await apiFetch<MilestoneDetailView>(`/dashboard/milestone?id=${id}`); } catch { /* */ }
});
</script>

<template>
  <div class="page" v-if="detail">
    <h1>{{ detail.milestone.title }}</h1>
    <p class="meta">{{ label('status', locale) }}: {{ detail.milestone.status }} &middot; {{ label('health', locale) }}: {{ detail.milestone.healthStatus }} &middot; {{ label('owner', locale) }}: {{ detail.milestone.owner }}</p>
    <p v-if="detail.milestone.completionCriteria"><strong>{{ label('criteria', locale) }}:</strong> {{ detail.milestone.completionCriteria }}</p>

    <div class="section">
      <h2>Work Items</h2>
      <p v-if="!detail.workItems.length" class="empty">{{ label('noData', locale) }}</p>
      <ul v-else>
        <li v-for="w in detail.workItems" :key="(w as any).id">{{ (w as any).title || (w as any).id }}</li>
      </ul>
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
h1 { margin: 0; }
.meta { color: #4a7a6d; margin: 8px 0 24px; }
.section { margin-top: 24px; }
h2 { font-size: 1.1rem; margin: 0 0 10px; }
.update-card { background: #fff; padding: 14px; border-radius: 10px; margin-bottom: 8px; box-shadow: 0 1px 4px rgba(0,0,0,.05); }
.update-card p { margin: 6px 0 0; color: #4a7a6d; }
.empty { color: #6b8a80; }
ul { list-style: none; padding: 0; }
li { padding: 8px 14px; background: #fff; border-radius: 8px; margin-bottom: 6px; box-shadow: 0 1px 4px rgba(0,0,0,.05); }
</style>
