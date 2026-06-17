<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'
import type { ExportDatasetDef } from '../lib/exportDatasets'

const props = defineProps<{
  dataset: ExportDatasetDef
  seasonMin: number | null
  seasonMax: number | null
  selectedTeamName: string
  teams: main.TeamPickerResultDTO[]
  careerStatType: string
}>()

defineEmits<{
  'update:seasonMin': [value: number | null]
  'update:seasonMax': [value: number | null]
  'update:selectedTeamName': [value: string]
  'update:careerStatType': [value: string]
}>()

function parseSeasonInput(raw: string): number | null {
  const n = parseInt(raw, 10)
  return Number.isNaN(n) || n <= 0 ? null : n
}
</script>

<template>
  <div class="filter-panel">
    <template v-if="dataset.supportsSeasonFilter">
      <div class="filter-row">
        <span class="filter-label">From Season</span>
        <input
          type="number"
          class="season-input"
          placeholder="Any"
          :value="seasonMin ?? ''"
          min="1"
          @change="$emit('update:seasonMin', parseSeasonInput(($event.target as HTMLInputElement).value))"
        />
      </div>
      <div class="filter-row">
        <span class="filter-label">To Season</span>
        <input
          type="number"
          class="season-input"
          placeholder="Any"
          :value="seasonMax ?? ''"
          min="1"
          @change="$emit('update:seasonMax', parseSeasonInput(($event.target as HTMLInputElement).value))"
        />
      </div>
    </template>

    <div v-if="dataset.supportsTeamFilter" class="filter-row">
      <span class="filter-label">Team</span>
      <select
        class="filter-select"
        :value="selectedTeamName"
        @change="$emit('update:selectedTeamName', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Any</option>
        <option v-for="t in teams" :key="t.teamName" :value="t.teamName">
          {{ t.teamName }}
        </option>
      </select>
    </div>

    <div v-if="dataset.supportsCareerStatType" class="filter-row">
      <span class="filter-label">Stat Type</span>
      <div class="toggle-group">
        <button
          class="toggle-btn"
          :class="{ active: !careerStatType || careerStatType === 'regular_season' }"
          @click="$emit('update:careerStatType', 'regular_season')"
        >
          Reg Season
        </button>
        <button
          class="toggle-btn"
          :class="{ active: careerStatType === 'playoffs' }"
          @click="$emit('update:careerStatType', 'playoffs')"
        >
          Playoffs
        </button>
        <button
          class="toggle-btn"
          :class="{ active: careerStatType === 'total_career' }"
          @click="$emit('update:careerStatType', 'total_career')"
        >
          Total
        </button>
      </div>
    </div>

    <p v-if="!dataset.supportsSeasonFilter && !dataset.supportsTeamFilter && !dataset.supportsCareerStatType" class="no-filters">
      No filters available for this dataset.
    </p>
  </div>
</template>

<style scoped>
.filter-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
  min-width: 72px;
}

.season-input {
  width: 80px;
  padding: 0.25rem 0.375rem;
  font-size: 0.8125rem;
  font-family: inherit;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.season-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.filter-select {
  flex: 1;
  padding: 0.25rem 0.375rem;
  font-size: 0.8125rem;
  font-family: inherit;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  cursor: pointer;
}

.toggle-group {
  display: flex;
}

.toggle-btn {
  padding: 0.2rem 0.5rem;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.75rem;
  font-family: inherit;
  cursor: pointer;
  border-radius: 0;
  transition:
    background 0.1s,
    color 0.1s;
}

.toggle-btn:first-child {
  border-radius: 4px 0 0 4px;
}

.toggle-btn:last-child {
  border-radius: 0 4px 4px 0;
}

.toggle-btn + .toggle-btn {
  margin-left: -1px;
}

.toggle-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.toggle-btn.active {
  background: var(--color-surface-2);
  border-color: var(--color-accent);
  color: var(--color-accent);
  position: relative;
  z-index: 1;
}

.no-filters {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  font-style: italic;
}
</style>
