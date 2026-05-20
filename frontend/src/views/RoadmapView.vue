<script setup lang="ts">
import { inject, onMounted, ref, type Ref } from "vue";
import type { Locale } from "../i18n";
import { label, apiFetch, type RoadmapOverviewItem } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const items = ref<RoadmapOverviewItem[]>([]);

onMounted(async () => {
  try { items.value = await apiFetch<RoadmapOverviewItem[]>("/dashboard/roadmap"); } catch { items.value = []; }
});
</script>

<template>
  <div class="page">
    <h1>{{ label("roadmap", locale) }}</h1>
    <p v-if="!items.length" class="empty">{{ label("noData", locale) }}</p>
    <div v-for="item in items" :key="item.period.id" class="period-card">
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
.period-card { background: #fff; border-radius: 14px; padding: 20px; margin-bottom: 16px; box-shadow: 0 2px 8px rgba(0,0,0,.06); }
.period-header { display: flex; justify-content: space-between; align-items: center; }
h2 { margin: 0; font-size: 1.2rem; }
h3 { font-size: .95rem; margin: 16px 0 8px; color: #4a7a6d; }
.dates { color: #6b8a80; font-size: .85rem; margin: 4px 0 12px; }
.badge { display: inline-block; padding: 3px 10px; border-radius: 999px; font-size: .78rem; font-weight: 600; }
.badge.active { background: #dcfce7; color: #15803d; }
.badge.archived { background: #f3f4f6; color: #6b7280; }
.badge.health.on_track { background: #dcfce7; color: #15803d; }
.badge.health.at_risk { background: #fef3c7; color: #b45309; }
.items { display: flex; flex-direction: column; gap: 6px; }
.ri-card { padding: 10px 14px; border-radius: 8px; background: #f0f5f3; display: flex; justify-content: space-between; }
.projects { margin-top: 12px; }
.ps-row { display: flex; gap: 12px; align-items: center; padding: 6px 0; font-size: .9rem; }
.empty { color: #6b8a80; }
</style>
