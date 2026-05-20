<script setup lang="ts">
import { ref, provide } from "vue";
import { RouterLink, RouterView } from "vue-router";
import type { Locale } from "./i18n";
import { t } from "./i18n";
import { label } from "./api";

const locale = ref<Locale>("zh-CN");

function toggleLocale() {
  locale.value = locale.value === "zh-CN" ? "en-US" : "zh-CN";
  document.documentElement.lang = locale.value;
}

provide("locale", locale);
provide("toggleLocale", toggleLocale);

const navItems = ["dashboard", "projects", "milestones", "roadmap", "review"] as const;
</script>

<template>
  <div class="app">
    <nav class="sidebar">
      <p class="brand">Goal Manager</p>
      <RouterLink
        v-for="item in navItems"
        :key="item"
        :to="{ name: item === 'dashboard' ? 'dashboard' : item }"
        class="nav-link"
        active-class="active"
      >
        {{ label(item, locale) }}
      </RouterLink>
      <button class="locale-btn" @click="toggleLocale">
        {{ locale === "zh-CN" ? "EN" : "中" }}
      </button>
    </nav>
    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<style>
:root { --nav-w: 180px; --gap: 12px; }
body { margin: 0; font-family: "Noto Sans SC","PingFang SC","Microsoft YaHei",sans-serif; background: #f5f7f6; color: #10352a; }
</style>

<style scoped>
.app { display: flex; min-height: 100vh; }
.sidebar {
  width: var(--nav-w); padding: 24px 16px; background: #10352a; color: #fff;
  display: flex; flex-direction: column; gap: 6px; position: fixed; top: 0; bottom: 0;
  z-index: 100; overflow-y: auto;
}
.brand { margin: 0 0 20px; font-size: .85rem; letter-spacing: .15em; text-transform: uppercase; color: #4fd1a5; }
.nav-link {
  display: block; padding: 10px 14px; border-radius: 10px; color: #a8d5c8;
  text-decoration: none; font-size: .92rem; transition: background .15s;
}
.nav-link:hover { background: rgba(255,255,255,.08); }
.nav-link.active { background: rgba(255,255,255,.14); color: #fff; font-weight: 600; }
.locale-btn {
  margin-top: auto; border: 1px solid #4fd1a5; border-radius: 999px; background: none;
  color: #4fd1a5; padding: 8px 0; cursor: pointer; font-size: .85rem;
}
.content { margin-left: var(--nav-w); flex: 1; padding: 32px 28px; width: calc(100% - var(--nav-w)); }
</style>
