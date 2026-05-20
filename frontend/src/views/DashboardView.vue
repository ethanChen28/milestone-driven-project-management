<script setup lang="ts">
import { inject, onMounted, ref, type Ref } from "vue";
import type { Locale } from "../i18n";
import { t } from "../i18n";
import { label, type PortfolioSummary } from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const summary = ref<PortfolioSummary>({
  activeProjects: 0, blockedMilestones: 0, overdueMilestones: 0,
  milestoneWorkItems: 0, bauWorkItems: 0, healthDistribution: {},
});
const loading = ref(true);

const cards = [
  { key: "activeProjects", field: "activeProjects" as const },
  { key: "blockedMilestones", field: "blockedMilestones" as const },
  { key: "overdueMilestones", field: "overdueMilestones" as const },
  { key: "workload", field: null },
];

onMounted(async () => {
  try {
    const resp = await fetch("/api/v1/dashboard/portfolio");
    if (resp.ok) summary.value = await resp.json();
  } catch { /* keep shell usable */ } finally { loading.value = false; }
});
</script>

<template>
  <div class="dashboard">
    <h1>{{ t(locale, "title") }}</h1>
    <p class="subtitle">{{ t(locale, "subtitle") }}</p>
    <div v-if="loading" class="loading">{{ t(locale, "loading") }}</div>
    <div v-else class="grid">
      <article v-for="c in cards" :key="c.key" class="stat-card">
        <span>{{ label(c.key, locale) }}</span>
        <strong v-if="c.field">{{ summary[c.field] }}</strong>
        <strong v-else>{{ summary.milestoneWorkItems }} / {{ summary.bauWorkItems }}</strong>
      </article>
    </div>
    <section v-if="Object.keys(summary.healthDistribution).length" class="health">
      <h3>{{ label("health", locale) }}</h3>
      <div class="health-bars">
        <div v-for="(v, k) in summary.healthDistribution" :key="k" class="bar-row">
          <span>{{ k }}</span>
          <div class="bar-track"><div class="bar-fill" :class="k" :style="{ width: v / summary.activeProjects * 100 + '%' }" /></div>
          <strong>{{ v }}</strong>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
h1 { margin: 0; font-size: 2rem; }
.subtitle { margin: 8px 0 24px; color: #4a7a6d; }
.loading { color: #2d7a61; }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 16px; }
.stat-card {
  padding: 20px; border-radius: 16px; background: rgba(16,53,42,.92); color: #fff;
  display: grid; gap: 8px;
}
.stat-card strong { font-size: 1.8rem; }
.health { margin-top: 28px; }
.health-bars { display: flex; flex-direction: column; gap: 8px; max-width: 480px; }
.bar-row { display: flex; align-items: center; gap: 10px; }
.bar-row span:first-child { width: 80px; text-align: right; font-size: .85rem; color: #4a7a6d; }
.bar-track { flex: 1; height: 8px; border-radius: 4px; background: #e0ebe6; }
.bar-fill { height: 100%; border-radius: 4px; }
.bar-fill.on_track { background: #22c55e; }
.bar-fill.at_risk { background: #f59e0b; }
.bar-fill.off_track { background: #ef4444; }
</style>
