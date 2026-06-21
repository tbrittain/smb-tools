<script lang="ts" setup>
import type { ExportColumnDef } from '../lib/exportDatasets'

const props = defineProps<{
  columns: ExportColumnDef[]
  modelValue: string[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

function toggle(key: string) {
  const next = props.modelValue.includes(key) ? props.modelValue.filter((k) => k !== key) : [...props.modelValue, key]
  emit('update:modelValue', next)
}

function selectAll() {
  emit(
    'update:modelValue',
    props.columns.map((c) => c.key),
  )
}

function deselectAll() {
  emit('update:modelValue', [])
}
</script>

<template>
  <div class="column-selector">
    <div class="col-selector-header">
      <span class="col-count">{{ modelValue.length }}/{{ columns.length }}</span>
      <button class="text-btn" @click="selectAll">All</button>
      <button class="text-btn" @click="deselectAll">None</button>
    </div>
    <div class="col-list">
      <label v-for="col in columns" :key="col.key" class="col-item">
        <input
          type="checkbox"
          :checked="modelValue.includes(col.key)"
          @change="toggle(col.key)"
        />
        <span class="col-label">{{ col.label }}</span>
      </label>
    </div>
  </div>
</template>

<style scoped>
.column-selector {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.col-selector-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.col-count {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  flex: 1;
}

.text-btn {
  background: none;
  border: none;
  padding: 0;
  font-family: inherit;
  font-size: 0.75rem;
  color: var(--color-accent);
  cursor: pointer;
}

.text-btn:hover {
  text-decoration: underline;
}

.col-list {
  max-height: 240px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 0.25rem;
}

.col-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.2rem 0.25rem;
  border-radius: 3px;
  cursor: pointer;
  user-select: none;
}

.col-item:hover {
  background: var(--color-surface-2);
}

.col-item input[type='checkbox'] {
  accent-color: var(--color-accent);
  flex-shrink: 0;
}

.col-label {
  font-size: 0.8125rem;
  color: var(--color-text-primary);
}
</style>
