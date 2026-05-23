<script setup lang="ts">
import { computed, onMounted, ref, provide, watch } from "vue";
import { RouterLink, RouterView } from "vue-router";
import type { Locale } from "./i18n";
import { label, getCurrentRole, getCurrentUser, setCurrentRole, setCurrentUser, workspaceRoles, workspaceUsers, type WorkspaceRole, type UserProfile, isTokenAuthMode, login, clearAccessToken, listUsers } from "./api";

const locale = ref<Locale>("zh-CN");
const currentRole = ref<WorkspaceRole>(getCurrentRole());
const currentUser = ref(getCurrentUser());
const authMode = ref(isTokenAuthMode() ? "token" : "dev-header");
const directoryUsers = ref<UserProfile[]>(workspaceUsers.map((user) => ({ id: user, username: user, displayName: user, email: `${user}@example.local`, status: "active", roles: [] })));
const loginUser = ref(getCurrentUser());
const loginPassword = ref("password");
const loginError = ref("");

function toggleLocale() {
  locale.value = locale.value === "zh-CN" ? "en-US" : "zh-CN";
  document.documentElement.lang = locale.value;
}

watch(currentRole, (role) => setCurrentRole(role), { immediate: true });
watch(currentUser, (user) => setCurrentUser(user.trim() || "tester"), { immediate: true });

async function refreshDirectory() {
  directoryUsers.value = await listUsers();
}

async function submitLogin() {
  loginError.value = "";
  try {
    const result = await login(loginUser.value.trim(), loginPassword.value);
    currentUser.value = result.user.id;
    const firstRole = result.user.roles[0];
    if (workspaceRoles.includes(firstRole as WorkspaceRole)) currentRole.value = firstRole as WorkspaceRole;
    await refreshDirectory();
  } catch (error) {
    loginError.value = error instanceof Error ? error.message : String(error);
  }
}

function logout() {
  clearAccessToken();
}

onMounted(refreshDirectory);

provide("locale", locale);
provide("toggleLocale", toggleLocale);
provide("currentRole", currentRole);
provide("currentUser", currentUser);

const baseNavItems = ["dashboard", "projects", "tasks", "milestones", "roadmap", "review"] as const;
const navItems = computed(() => {
  const items = [...baseNavItems];
  if (currentRole.value === "admin") items.push("users" as never);
  return items;
});
</script>

<template>
  <div class="app">
    <a href="#main-content" class="skip-link">Skip to content</a>
    <main id="main-content" class="content" role="main">
      <RouterView />
    </main>
    <nav class="sidebar" role="navigation" aria-label="Main navigation">
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
      <div class="role-panel">
        <template v-if="authMode === 'token'">
          <span>登录用户</span>
          <input v-model="loginUser" class="role-select" aria-label="login user" />
          <input v-model="loginPassword" class="role-select" aria-label="login password" type="password" />
          <button class="auth-btn" @click="submitLogin">登录</button>
          <button class="auth-btn secondary" @click="logout">退出</button>
          <small v-if="loginError">{{ loginError }}</small>
          <small v-else>{{ currentUser }} · {{ currentRole }}</small>
        </template>
        <template v-else>
          <span>{{ label("roleTool", locale) }}</span>
          <select v-model="currentRole" class="role-select" aria-label="workspace role">
            <option v-for="role in workspaceRoles" :key="role" :value="role">{{ role }}</option>
          </select>
          <span>{{ label("userTool", locale) }}</span>
          <select v-model="currentUser" class="user-select" aria-label="current user">
            <option v-for="user in directoryUsers" :key="user.id" :value="user.id">{{ user.displayName }}</option>
          </select>
          <small>{{ label("roleWarning", locale) }}</small>
        </template>
      </div>
      <button class="locale-btn" @click="toggleLocale">
        {{ locale === "zh-CN" ? "EN" : "中" }}
      </button>
    </nav>
  </div>
</template>

<style>
:root { --nav-w: 180px; --gap: 12px; }
body { margin: 0; font-family: "Noto Sans SC","PingFang SC","Microsoft YaHei",sans-serif; background: #f5f7f6; color: #10352a; }
.skip-link { position: absolute; top: -100%; left: 0; padding: 8px 16px; background: #10352a; color: #fff; z-index: 200; font-size: .9rem; border-radius: 0 0 8px 0; }
.skip-link:focus { top: 0; }
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
.role-panel { margin-top: auto; padding: 12px; border: 1px solid rgba(79,209,165,.45); border-radius: 14px; display: grid; gap: 7px; color: #a8d5c8; font-size: .76rem; }
.role-select, .user-select { width: 100%; border: 0; border-radius: 8px; padding: 7px 8px; background: #e9fff6; color: #10352a; font-size: .78rem; }
.role-panel small { color: #4fd1a5; }
.auth-btn { border: 0; border-radius: 8px; padding: 7px 8px; background: #4fd1a5; color: #062e24; cursor: pointer; font-weight: 700; }
.auth-btn.secondary { background: transparent; color: #a8d5c8; border: 1px solid rgba(168,213,200,.45); }
.locale-btn {
  border: 1px solid #4fd1a5; border-radius: 999px; background: none;
  color: #4fd1a5; padding: 8px 0; cursor: pointer; font-size: .85rem;
}
.content { margin-left: var(--nav-w); flex: 1; padding: 32px 28px; width: calc(100% - var(--nav-w)); }
</style>
