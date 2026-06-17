<script lang="ts" setup>
import type { ExportDatasetDef } from '../lib/exportDatasets'

defineProps<{
  datasets: ExportDatasetDef[]
  modelValue: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div class="dataset-picker">
    <button
      v-for="ds in datasets"
      :key="ds.id"
      class="dataset-btn"
      :class="{ active: ds.id === modelValue }"
      @click="$emit('update:modelValue', ds.id)"
    >
      {{ ds.label }}
    </button>
  </div>
</template>

<style scoped>
.dataset-picker {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dataset-btn {
  text-align: left;
  padding: 0.4rem 0.625rem;
  border: 1px solid transparent;
  border-radius: 4px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  font-family: inherit;
  cursor: pointer;
  transition:
    background 0.1s,
    color 0.1s;
}

.dataset-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.dataset-btn.active {
  background: var(--color-surface-2);
  border-color: var(--color-accent);
  color: var(--color-accent);
}
</style>
