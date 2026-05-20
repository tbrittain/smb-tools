<script lang="ts" setup>
import { ref, watch } from 'vue'
import type { main } from '../../wailsjs/go/models'
import { BAT_HANDS, BATTING_POSITIONS, CHEMISTRY_TYPES, PITCHING_ROLES, THROW_HANDS } from '../constants/domain'
import FilterBar from './FilterBar.vue'

const props = defineProps<{
  mode: 'batting' | 'pitching'
  seasons: main.SeasonSummaryDTO[]
  modelValue: main.LeaderboardFiltersDTO
}>()

const emit = defineEmits<{
  'update:modelValue': [v: main.LeaderboardFiltersDTO]
}>()

const local = ref<main.LeaderboardFiltersDTO>({ ...props.modelValue })

watch(
  () => props.modelValue,
  (v) => {
    local.value = { ...v }
  },
  { deep: true },
)

function update(patch: Partial<main.LeaderboardFiltersDTO>) {
  local.value = { ...local.value, ...patch }
  emit('update:modelValue', { ...local.value })
}
</script>

<template>
  <FilterBar>
    <label class="filter-item">
      <input
        type="checkbox"
        :checked="local.isPlayoffs"
        @change="update({ isPlayoffs: ($event.target as HTMLInputElement).checked })"
      />
      Playoffs
    </label>

    <label class="filter-item">
      <input
        type="checkbox"
        :checked="local.onlyHallOfFamers"
        @change="update({ onlyHallOfFamers: ($event.target as HTMLInputElement).checked })"
      />
      HoF Only
    </label>

    <div class="filter-group">
      <span class="filter-label">{{ mode === 'batting' ? 'Position' : 'Role' }}</span>
      <select
        :value="local.position"
        @change="update({ position: ($event.target as HTMLSelectElement).value })"
      >
        <option value="">Any</option>
        <option
          v-for="opt in mode === 'batting' ? BATTING_POSITIONS : PITCHING_ROLES"
          :key="opt"
          :value="opt"
        >
          {{ opt }}
        </option>
      </select>
    </div>

    <div class="filter-group">
      <span class="filter-label">{{ mode === 'batting' ? 'Bat Hand' : 'Throw Hand' }}</span>
      <select
        v-if="mode === 'batting'"
        :value="local.batHand"
        @change="update({ batHand: ($event.target as HTMLSelectElement).value })"
      >
        <option value="">Any</option>
        <option v-for="h in BAT_HANDS" :key="h" :value="h">{{ h }}</option>
      </select>
      <select
        v-else
        :value="local.throwHand"
        @change="update({ throwHand: ($event.target as HTMLSelectElement).value })"
      >
        <option value="">Any</option>
        <option v-for="h in THROW_HANDS" :key="h" :value="h">{{ h }}</option>
      </select>
    </div>

    <div class="filter-group">
      <span class="filter-label">Chemistry</span>
      <select
        :value="local.chemistryType"
        @change="update({ chemistryType: ($event.target as HTMLSelectElement).value })"
      >
        <option value="">Any</option>
        <option v-for="c in CHEMISTRY_TYPES" :key="c" :value="c">{{ c }}</option>
      </select>
    </div>

    <div class="filter-group">
      <span class="filter-label">From Season</span>
      <select
        :value="local.seasonStart"
        @change="update({ seasonStart: Number(($event.target as HTMLSelectElement).value) })"
      >
        <option :value="0">All</option>
        <option v-for="s in seasons" :key="s.seasonNum" :value="s.seasonNum">
          S{{ s.seasonNum }}
        </option>
      </select>
    </div>

    <div class="filter-group">
      <span class="filter-label">To Season</span>
      <select
        :value="local.seasonEnd"
        @change="update({ seasonEnd: Number(($event.target as HTMLSelectElement).value) })"
      >
        <option :value="0">All</option>
        <option v-for="s in seasons" :key="s.seasonNum" :value="s.seasonNum">
          S{{ s.seasonNum }}
        </option>
      </select>
    </div>
  </FilterBar>
</template>

<style scoped>
.filter-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  cursor: pointer;
  white-space: nowrap;
}

.filter-item input[type='checkbox'] {
  accent-color: var(--color-accent);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.filter-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.filter-group select {
  font-size: 0.8125rem;
  padding: 0.2rem 0.4rem;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  cursor: pointer;
}
</style>
