<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import type { Locale } from "../i18n";
import { label, apiFetch, type RoadmapOverviewItem } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const items = ref<RoadmapOverviewItem[]>([]);
const filters = ref({ owner: "", status: "" });
const filteredItems = computed(() => items.value.filter((item) => (!filters.value.status || item.period.status === filters.value.status) && (!filters.value.owner || item.period.owner === filters.value.owner)));
function clearFilters() { filters.value = { owner: "", status: "" }; }

onMounted(async () => {
  try { items.value = await apiFetch<RoadmapOverviewItem[]>("/dashboard/roadmap"); } catch { items.value = []; }
});
</script>

<template>
  <div class="page">
    <h1>{{ label("roadmap", locale) }}</h1>
    <div class="filters"><strong>{{ label('filters', locale) }}</strong><input v-model="filters.owner" :placeholder="label('owner', locale)" /><select v-model="filters.status"><option value="">{{ label('status', locale) }}</option><option value="active">active</option><option value="archived">archived</option></select><button class="btn" @click="clearFilters">{{ label('clearFilters', locale) }}</button></div>
    <p v-if="!filteredItems.length" class="empty">{{ label("noData", locale) }}</p>
    <div v-for="item in filteredItems" :key="item.period.id" class="period-card">
      <div class="period-header">
        <h2>{{ item.period.title }}</h2>
        <span class="badge" :class="item.period.status">{{ item.period.status }}</span>
      </div>
      <p class="dates">{{ item.period.periodStart?.slice(0,10) }} ~ {{ item.period.periodEnd?.slice(0,10) }}</p>
      <div v-if="item.items.length" class="items">
        <div v-for="ri in item.items" :key="ri.id" class="ri-card">
          <strong>{{ ri.title }}</strong>
          <span>{{ ri.status }} &middot; {{ ri.owner }}</span>
        </div>
      </div>
      <div v-if="item.projectSummaries?.length" class="projects">
        <h3>{{ label("projects", locale) }}</h3>
        <div v-for="ps in item.projectSummaries" :key="ps.id" class="ps-row">
          <span>{{ ps.name }}</span>
          <span class="badge health" :class="ps.healthStatus">{{ ps.healthStatus }}</span>
          <span>{{ ps.milestones }} milestones</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
h1 { margin: 0 0 20px; }
.filters { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; background: var(--color-surface); padding: 12px; border-radius: 12px; margin-bottom: 16px; box-shadow: var(--shadow-sm); }
.filters input, .filters select { padding: 8px 10px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); }
.period-card { background: var(--color-surface); border-radius: var(--radius-lg); padding: 20px; margin-bottom: 16px; box-shadow: var(--shadow-md); }
.period-header { display: flex; justify-content: space-between; align-items: center; }
h2 { margin: 0; font-size: 1.2rem; }
h3 { font-size: .95rem; margin: 16px 0 8px; color: var(--color-text-muted); }
.dates { color: var(--color-text-subtle); font-size: .85rem; margin: 4px 0 12px; }
.items { display: flex; flex-direction: column; gap: 6px; }
.ri-card { padding: 10px 14px; border-radius: var(--radius-sm); background: var(--color-surface-alt); display: flex; justify-content: space-between; }
.projects { margin-top: 12px; }
.ps-row { display: flex; gap: 12px; align-items: center; padding: 6px 0; font-size: .9rem; }
</style>
