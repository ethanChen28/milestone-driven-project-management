<script setup lang="ts">
import { computed, inject, onMounted, ref, type Ref } from "vue";
import type { Locale } from "../i18n";
import {
  label, listUsers, createUser, updateUser, disableUser, enableUser, assignRole,
  can, type UserProfile, type WorkspaceRole,
} from "../api";

const locale = inject<Ref<Locale>>("locale")!;
const currentRole = inject<Ref<WorkspaceRole>>("currentRole")!;
const users = ref<UserProfile[]>([]);
const showCreateForm = ref(false);
const editingUser = ref<UserProfile | null>(null);
const error = ref("");
const isAdmin = computed(() => currentRole.value === "admin");

const createForm = ref({ username: "", displayName: "", email: "", password: "password", role: "contributor" });
const editForm = ref({ displayName: "", email: "" });

async function load() {
  try { users.value = await listUsers(); error.value = ""; }
  catch { users.value = []; }
}
onMounted(load);

async function create() {
  try {
    await createUser(createForm.value);
    showCreateForm.value = false;
    createForm.value = { username: "", displayName: "", email: "", password: "password", role: "contributor" };
    await load();
  } catch (err) { error.value = err instanceof Error ? err.message : String(err); }
}

async function saveEdit() {
  if (!editingUser.value) return;
  try {
    await updateUser(editingUser.value.id, editForm.value);
    editingUser.value = null;
    await load();
  } catch (err) { error.value = err instanceof Error ? err.message : String(err); }
}

async function toggleStatus(user: UserProfile) {
  try {
    if (user.status === "active") await disableUser(user.id);
    else await enableUser(user.id);
    await load();
  } catch (err) { error.value = err instanceof Error ? err.message : String(err); }
}

async function changeRole(userId: string, role: string) {
  try { await assignRole(userId, role); await load(); }
  catch (err) { error.value = err instanceof Error ? err.message : String(err); }
}

function startEdit(user: UserProfile) {
  editingUser.value = user;
  editForm.value = { displayName: user.displayName, email: user.email };
}

function copy(zh: string, en: string) { return locale.value === "zh-CN" ? zh : en; }
</script>

<template>
  <div class="page users-page">
    <div class="header">
      <h1>{{ label("users", locale) }}</h1>
      <button v-if="isAdmin" class="btn primary" @click="showCreateForm = !showCreateForm">
        {{ showCreateForm ? label("cancel", locale) : label("createUser", locale) }}
      </button>
    </div>
    <p v-if="!isAdmin" class="permission-hint">{{ label("noPermission", locale) }}<small>{{ label("needAdmin", locale) }}</small></p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <form v-if="showCreateForm" class="form" @submit.prevent="create">
      <div class="form-grid">
        <label>{{ label("username", locale) }}<input v-model="createForm.username" required /></label>
        <label>{{ label("name", locale) }}<input v-model="createForm.displayName" required /></label>
        <label>{{ label("email", locale) }}<input v-model="createForm.email" type="email" required /></label>
        <label>{{ copy("密码", "Password") }}<input v-model="createForm.password" type="password" required /></label>
        <label>{{ label("role", locale) }}
          <select v-model="createForm.role">
            <option v-for="role in ['admin','portfolio_manager','project_owner','contributor','viewer']" :key="role" :value="role">{{ role }}</option>
          </select>
        </label>
      </div>
      <div class="row">
        <button class="btn primary" type="submit">{{ label("save", locale) }}</button>
        <button class="btn" type="button" @click="showCreateForm = false">{{ label("cancel", locale) }}</button>
      </div>
    </form>

    <table v-if="users.length">
      <thead>
        <tr>
          <th>{{ label("name", locale) }}</th>
          <th>{{ label("username", locale) }}</th>
          <th>{{ label("email", locale) }}</th>
          <th>{{ label("role", locale) }}</th>
          <th>{{ label("status", locale) }}</th>
          <th>{{ label("edit", locale) }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="user in users" :key="user.id">
          <td v-if="editingUser?.id === user.id">
            <input v-model="editForm.displayName" class="inline-input" />
          </td>
          <td v-else><strong>{{ user.displayName }}</strong></td>
          <td>{{ user.username }}</td>
          <td v-if="editingUser?.id === user.id">
            <input v-model="editForm.email" type="email" class="inline-input" />
          </td>
          <td v-else>{{ user.email }}</td>
          <td>
            <select :value="user.roles[0] || 'contributor'" @change="changeRole(user.id, ($event.target as HTMLSelectElement).value)">
              <option v-for="role in ['admin','portfolio_manager','project_owner','contributor','viewer']" :key="role" :value="role">{{ role }}</option>
            </select>
          </td>
          <td><span class="badge" :class="user.status === 'active' ? 'active' : 'archived'">{{ user.status }}</span></td>
          <td class="actions">
            <template v-if="editingUser?.id === user.id">
              <button class="btn sm primary" @click="saveEdit">{{ label("save", locale) }}</button>
              <button class="btn sm" @click="editingUser = null">{{ label("cancel", locale) }}</button>
            </template>
            <template v-else>
              <button class="btn sm" @click="startEdit(user)">{{ label("editUser", locale) }}</button>
              <button class="btn sm" :class="user.status === 'active' ? 'danger' : ''" @click="toggleStatus(user)">
                {{ user.status === "active" ? label("disableUser", locale) : label("enableUser", locale) }}
              </button>
            </template>
          </td>
        </tr>
      </tbody>
    </table>
    <p v-else class="empty">{{ label("noData", locale) }}</p>
  </div>
</template>

<style scoped>
.users-page { max-width: 960px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
h1 { margin: 0; }
.form { display: flex; flex-direction: column; gap: 10px; max-width: 560px; margin-bottom: 20px; background: var(--color-surface); padding: 20px; border-radius: 12px; box-shadow: var(--shadow-md); }
.form-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
.form input, .form select { padding: 10px 12px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); font-family: inherit; }
.row { display: flex; gap: 10px; }
table { width: 100%; border-collapse: collapse; background: var(--color-surface); border-radius: 12px; overflow: hidden; box-shadow: var(--shadow-md); }
th { text-align: left; padding: 12px 16px; background: var(--color-surface-alt); font-size: .82rem; color: var(--color-text-muted); }
td { padding: 12px 16px; border-top: 1px solid #eaf0ed; }
td select, .inline-input { padding: 6px 8px; border: 1px solid var(--color-border); border-radius: var(--radius-sm); font-size: .85rem; }
.actions { display: flex; gap: 6px; flex-wrap: wrap; }
.btn.danger { color: var(--color-danger); border-color: #fecaca; }
</style>
