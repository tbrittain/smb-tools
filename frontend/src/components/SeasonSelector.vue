<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'

const props = defineProps<{
  seasons: main.SeasonSummaryDTO[]
  modelValue: number | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

function handleChange(e: Event) {
  const v = (e.target as HTMLSelectElement).value
  emit('update:modelValue', v ? Number(v) : null)
}
</script>

<template>
  <select
    :value="modelValue ?? ''"
    class="season-select"
    aria-label="Select season"
    @change="handleChange"
  >
    <option value="" disabled>Select a season…</option>
    <option v-for="s in seasons" :key="s.id" :value="s.id">
      Season {{ s.seasonNum }}
      <template v-if="s.championTeamName"> · {{ s.championTeamName }}</template>
    </option>
  </select>
</template>

<style scoped>
.season-select {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
  font-family: var(--font-sans);
  padding: 0.375rem 0.625rem;
  outline: none;
  cursor: pointer;
}

.season-select:focus {
  border-color: var(--color-accent);
}
</style>
