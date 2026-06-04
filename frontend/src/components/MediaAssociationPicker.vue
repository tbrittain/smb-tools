<script lang="ts" setup>
import Button from 'primevue/button'
import { ref } from 'vue'
import { GetTeamSeasonsForMediaPicker, SearchPlayers, SearchTeamsForMediaPicker } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import { useSearchDebounce } from '../composables/useSearchDebounce'

const props = defineProps<{
  mode: 'team_season' | 'player'
  alreadySelectedTeamHistoryIds?: number[]
  alreadySelectedPlayerIds?: number[]
}>()

const emit = defineEmits<{
  picked: [type: 'team_season' | 'player', id: number, label: string]
}>()

// Team-season mode state
const teamResults = ref<main.TeamPickerResultDTO[]>([])
const selectedTeam = ref<main.TeamPickerResultDTO | null>(null)
const teamSeasons = ref<main.TeamSeasonPickerResultDTO[]>([])
const selectedSeasonHistoryId = ref<number | null>(null)
const loadingSeasons = ref(false)

// Player mode state
const playerResults = ref<main.PlayerSearchResultDTO[]>([])

async function runTeamSearch(q: string) {
  const results = await SearchTeamsForMediaPicker(q)
  teamResults.value = results ?? []
  selectedTeam.value = null
  teamSeasons.value = []
  selectedSeasonHistoryId.value = null
}

async function runPlayerSearch(q: string) {
  const results = await SearchPlayers(q)
  playerResults.value = (results ?? []).slice(0, 10)
}

const { query: teamQuery, loading: teamSearching } = useSearchDebounce(runTeamSearch, 300)
const { query: playerQuery, loading: playerSearching } = useSearchDebounce(runPlayerSearch, 300)

async function pickTeam(team: main.TeamPickerResultDTO) {
  selectedTeam.value = team
  teamSeasons.value = []
  selectedSeasonHistoryId.value = null
  loadingSeasons.value = true
  try {
    const seasons = await GetTeamSeasonsForMediaPicker(team.teamId)
    teamSeasons.value = seasons ?? []
  } finally {
    loadingSeasons.value = false
  }
}

function addTeamSeason() {
  if (selectedTeam.value === null || selectedSeasonHistoryId.value === null) return
  const season = teamSeasons.value.find((s) => s.teamHistoryId === selectedSeasonHistoryId.value)
  if (!season) return
  const label = `${selectedTeam.value.teamName} S${season.seasonNum}`
  emit('picked', 'team_season', selectedSeasonHistoryId.value, label)
  teamQuery.value = ''
  selectedTeam.value = null
  teamSeasons.value = []
  selectedSeasonHistoryId.value = null
}

function addPlayer(player: main.PlayerSearchResultDTO) {
  const label = `${player.firstName} ${player.lastName}`
  emit('picked', 'player', player.playerId, label)
  playerQuery.value = ''
  playerResults.value = []
}

function isTeamSeasonAlreadySelected(historyId: number): boolean {
  return props.alreadySelectedTeamHistoryIds?.includes(historyId) ?? false
}

function isPlayerAlreadySelected(playerId: number): boolean {
  return props.alreadySelectedPlayerIds?.includes(playerId) ?? false
}
</script>

<template>
  <div class="association-picker">
    <!-- Team-season mode -->
    <template v-if="mode === 'team_season'">
      <div v-if="!selectedTeam" class="search-step">
        <div class="search-wrap">
          <input
            v-model="teamQuery"
            type="text"
            class="picker-input"
            placeholder="Search teams…"
            autocomplete="off"
          />
          <span v-if="teamSearching" class="search-spinner" />
        </div>
        <div v-if="teamResults.length > 0" class="picker-results">
          <button
            v-for="team in teamResults"
            :key="team.teamId"
            class="picker-result-item"
            @click="pickTeam(team)"
          >
            {{ team.teamName }}
          </button>
        </div>
      </div>

      <div v-else class="season-step">
        <div class="team-name-row">
          <span class="team-name-label">{{ selectedTeam.teamName }}</span>
          <button class="back-btn" @click="selectedTeam = null">Change team</button>
        </div>
        <div v-if="loadingSeasons" class="loading-hint">Loading seasons…</div>
        <div v-else class="season-select-row">
          <select v-model="selectedSeasonHistoryId" class="picker-select">
            <option :value="null" disabled>Select season…</option>
            <option
              v-for="s in teamSeasons"
              :key="s.teamHistoryId"
              :value="s.teamHistoryId"
              :disabled="isTeamSeasonAlreadySelected(s.teamHistoryId)"
            >
              Season {{ s.seasonNum }}{{ isTeamSeasonAlreadySelected(s.teamHistoryId) ? ' (already added)' : '' }}
            </option>
          </select>
          <Button
            label="Add"
            size="small"
            :disabled="selectedSeasonHistoryId === null"
            @click="addTeamSeason"
          />
        </div>
      </div>
    </template>

    <!-- Player mode -->
    <template v-else>
      <div class="search-step">
        <div class="search-wrap">
          <input
            v-model="playerQuery"
            type="text"
            class="picker-input"
            placeholder="Search players…"
            autocomplete="off"
          />
          <span v-if="playerSearching" class="search-spinner" />
        </div>
        <div v-if="playerResults.length > 0" class="picker-results">
          <button
            v-for="p in playerResults"
            :key="p.playerId"
            class="picker-result-item"
            :class="{ 'picker-result-item--disabled': isPlayerAlreadySelected(p.playerId) }"
            :disabled="isPlayerAlreadySelected(p.playerId)"
            @click="!isPlayerAlreadySelected(p.playerId) && addPlayer(p)"
          >
            {{ p.firstName }} {{ p.lastName }}
            <span v-if="isPlayerAlreadySelected(p.playerId)" class="already-added">(already added)</span>
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.association-picker {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.search-step,
.season-step {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.picker-input {
  width: 100%;
  padding: 0.3125rem 0.625rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.8125rem;
  font-family: var(--font-sans);
  outline: none;
}

.picker-input:focus {
  border-color: var(--color-accent);
}

.search-spinner {
  position: absolute;
  right: 0.5rem;
  width: 12px;
  height: 12px;
  border: 2px solid var(--color-surface-3);
  border-top-color: var(--color-accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.picker-results {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  max-height: 160px;
  overflow-y: auto;
}

.picker-result-item {
  display: block;
  width: 100%;
  padding: 0.375rem 0.75rem;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  font-family: var(--font-sans);
  font-size: 0.8125rem;
  color: var(--color-text-primary);
}

.picker-result-item:hover:not(:disabled) {
  background: var(--color-surface-2);
}

.picker-result-item--disabled,
.picker-result-item:disabled {
  opacity: 0.5;
  cursor: default;
}

.already-added {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  margin-left: 0.25rem;
}

.team-name-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.team-name-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.back-btn {
  background: none;
  border: none;
  font-size: 0.75rem;
  color: var(--color-accent);
  cursor: pointer;
  padding: 0;
  font-family: var(--font-sans);
  text-decoration: underline;
}

.loading-hint {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.season-select-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.picker-select {
  flex: 1;
  padding: 0.3125rem 0.5rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.8125rem;
  font-family: var(--font-sans);
  outline: none;
}

.picker-select:focus {
  border-color: var(--color-accent);
}
</style>
