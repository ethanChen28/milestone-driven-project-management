<script setup lang="ts">
import { ref, computed } from "vue";
import type { UserProfile } from "../api";
import { fuzzyMatchUsers } from "../api";

interface Props {
  modelValue: string | string[];
  mode?: "single" | "multi";
  placeholder?: string;
  disabled?: boolean;
  users: UserProfile[];
}

const props = withDefaults(defineProps<Props>(), {
  mode: "single",
  placeholder: "Search users...",
  disabled: false,
});

const emit = defineEmits<{
  "update:modelValue": [value: string | string[]];
}>();

const query = ref("");
const isOpen = ref(false);
const highlightIndex = ref(-1);

const selectedUsers = computed<UserProfile[]>(() => {
  const ids = Array.isArray(props.modelValue) ? props.modelValue : props.modelValue ? [props.modelValue] : [];
  return ids
    .map((id) => props.users.find((u) => u.id === id))
    .filter((u): u is UserProfile => u !== undefined);
});

const filteredUsers = computed<UserProfile[]>(() => {
  const matched = fuzzyMatchUsers(props.users, query.value);
  const selectedIds = new Set(Array.isArray(props.modelValue) ? props.modelValue : props.modelValue ? [props.modelValue] : []);
  return matched.filter((u) => !selectedIds.has(u.id));
});

function selectUser(user: UserProfile) {
  if (props.disabled) return;
  if (props.mode === "multi") {
    const current = Array.isArray(props.modelValue) ? [...props.modelValue] : [];
    emit("update:modelValue", [...current, user.id]);
  } else {
    emit("update:modelValue", user.id);
    isOpen.value = false;
  }
  query.value = "";
  highlightIndex.value = -1;
}

function removeUser(id: string) {
  if (props.disabled) return;
  if (props.mode === "multi") {
    const current = Array.isArray(props.modelValue) ? props.modelValue : [];
    emit(
      "update:modelValue",
      current.filter((uid) => uid !== id),
    );
  }
}

function clear() {
  if (props.disabled) return;
  emit("update:modelValue", props.mode === "multi" ? [] : "");
  query.value = "";
}

function open() {
  if (props.disabled) return;
  isOpen.value = true;
}

let blurTimer: ReturnType<typeof setTimeout> | null = null;

function delayedClose() {
  blurTimer = setTimeout(() => {
    isOpen.value = false;
    highlightIndex.value = -1;
  }, 200);
}

function onKeydown(e: KeyboardEvent) {
  if (!isOpen.value) return;

  if (e.key === "ArrowDown") {
    e.preventDefault();
    highlightIndex.value = Math.min(highlightIndex.value + 1, filteredUsers.value.length - 1);
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    highlightIndex.value = Math.max(highlightIndex.value - 1, 0);
  } else if (e.key === "Enter") {
    e.preventDefault();
    if (highlightIndex.value >= 0 && highlightIndex.value < filteredUsers.value.length) {
      selectUser(filteredUsers.value[highlightIndex.value]);
    }
  } else if (e.key === "Escape") {
    isOpen.value = false;
    highlightIndex.value = -1;
  }
}
</script>

<template>
  <div class="person-picker">
    <!-- Multi-select chips -->
    <div v-if="mode === 'multi' && selectedUsers.length" class="chips">
      <span v-for="user in selectedUsers" :key="user.id" class="chip">
        {{ user.displayName }}
        <button type="button" class="chip-remove" @click="removeUser(user.id)" :disabled="disabled">&times;</button>
      </span>
    </div>
    <!-- Single-select display -->
    <div v-if="mode === 'single' && modelValue" class="selected-display">
      <span>{{ selectedUsers[0]?.displayName || modelValue }}</span>
      <button v-if="!disabled" type="button" class="clear-btn" @click="clear">&times;</button>
    </div>
    <!-- Search input -->
    <input
      v-model="query"
      @focus="open"
      @blur="delayedClose"
      @keydown="onKeydown"
      :placeholder="mode === 'single' && modelValue ? '' : placeholder"
      :disabled="disabled"
      role="combobox"
      :aria-expanded="isOpen"
      aria-autocomplete="list"
    />
    <!-- Dropdown -->
    <ul v-if="isOpen && filteredUsers.length" class="dropdown" role="listbox">
      <li
        v-for="(user, index) in filteredUsers"
        :key="user.id"
        @mousedown.prevent="selectUser(user)"
        :class="{ highlighted: index === highlightIndex }"
        role="option"
      >
        <strong>{{ user.displayName }}</strong>
        <small>{{ user.username }} &middot; {{ user.email }}</small>
      </li>
    </ul>
    <div v-if="isOpen && query && !filteredUsers.length" class="dropdown empty">No matching users</div>
  </div>
</template>

<style scoped>
.person-picker {
  position: relative;
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 6px;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--color-surface-alt);
  color: var(--color-text);
  font-size: 0.82rem;
}
.chip-remove {
  border: 0;
  background: none;
  cursor: pointer;
  color: var(--color-text-subtle);
  font-size: 0.9rem;
  padding: 0;
}
.selected-display {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 0.88rem;
}
.clear-btn {
  border: 0;
  background: none;
  cursor: pointer;
  color: var(--color-text-subtle);
  font-size: 0.9rem;
  padding: 0;
}
input {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-family: inherit;
  color: var(--color-text);
}
input:focus {
  outline: 2px solid var(--color-info);
  outline-offset: -1px;
}
.dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  z-index: 50;
  margin-top: 4px;
  max-height: 240px;
  overflow-y: auto;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  list-style: none;
  padding: 4px;
}
.dropdown li {
  padding: 10px 12px;
  cursor: pointer;
  border-radius: var(--radius-sm);
}
.dropdown li:hover,
.dropdown li.highlighted {
  background: var(--color-surface-alt);
}
.dropdown li strong {
  display: block;
  font-size: 0.88rem;
}
.dropdown li small {
  color: var(--color-text-subtle);
  font-size: 0.76rem;
}
.dropdown.empty {
  padding: 12px;
  color: var(--color-text-subtle);
  font-size: 0.85rem;
}
</style>
