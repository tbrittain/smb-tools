<script lang="ts" setup>
import MultiSelect from 'primevue/multiselect'
import { ref, watch } from 'vue'
import type { main } from '../../wailsjs/go/models'
import {
  BAT_HANDS,
  BATTING_POSITIONS,
  CHEMISTRY_TYPES,
  PITCHING_ROLES,
  SMB4_TRAITS,
  THROW_HANDS,
} from '../constants/domain'
import FilterBar from './FilterBar.vue'

const props = defineProps<{
  mode: 'batting' | 'pitching'
  isCareer: boolean
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
    local.value = { ...v, traits: [...(v.traits ?? [])] }
  },
)

function update(patch: Partial<main.LeaderboardFiltersDTO>) {
  local.value = { ...local.value, ...patch }
  emit('update:modelValue', { ...local.value })
}

function onTraitsChange(selected: string[]) {
  update({ traits: selected.slice(0, 2) })
}
</script>

<template>
  <FilterBar>
    <div class="filter-group">
      <span class="filter-label">Game Type</span>
      <div class="toggle-group">
        <button
          class="toggle-btn"
          :class="{ active: !local.gameType || local.gameType === 'regular' }"
          @click="update({ gameType: 'regular' })"
        >
          Reg Season
        </button>
        <button
          class="toggle-btn"
          :class="{ active: local.gameType === 'playoffs' }"
          @click="update({ gameType: 'playoffs' })"
        >
          Playoffs
        </button>
        <button
          v-if="isCareer"
          class="toggle-btn"
          :class="{ active: local.gameType === 'combined' }"
          @click="update({ gameType: 'combined' })"
        >
          Combined
        </button>
      </div>
    </div>

    <label class="filter-item">
      <input
        type="checkbox"
        :checked="local.qualifiedOnly"
        @change="update({ qualifiedOnly: ($event.target as HTMLInputElement).checked })"
      />
      Qualified Only
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

    <div v-if="!isCareer" class="filter-group">
      <span class="filter-label">Traits</span>
      <MultiSelect
        :model-value="local.traits ?? []"
        :options="[...SMB4_TRAITS]"
        placeholder="Any"
        :selection-limit="2"
        :max-selected-labels="2"
        :show-toggle-all="false"
        filter
        class="trait-select"
        @update:model-value="onTraitsChange"
      />
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

.trait-select {
  font-size: 0.8125rem;
  min-width: 160px;
}

.toggle-group {
  display: flex;
}

.toggle-btn {
  padding: 0.2rem 0.5rem;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
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
</style>
