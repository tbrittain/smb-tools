<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import {
  GetCareerLeaders,
  GetSeasonList,
  GetSeasonStatLeaders,
  GetStandings,
  SyncSeason,
} from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import CareerLeadersPanel from '../components/CareerLeadersPanel.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import SaveFilePicker from '../components/SaveFilePicker.vue'
import SeasonSelector from '../components/SeasonSelector.vue'
import StandingsTable from '../components/StandingsTable.vue'
import StatLeadersPanel from '../components/StatLeadersPanel.vue'
import { useFranchiseStore } from '../stores/franchise'

const franchiseStore = useFranchiseStore()

// ── Save file configuration ──────────────────────────────────────────────────

const showSaveFilePicker = ref(false)
const saveFileError = ref<string | null>(null)
const showForkPicker = ref(false)
const forkSeasonOffset = ref(0)
const forkError = ref<string | null>(null)

// Sets the initial source (no source exists yet) or replaces the active source (correction).
async function handleSaveFileChange(path: string, leagueGUID: string) {
  if (!franchiseStore.active) return
  saveFileError.value = null
  try {
    if (franchiseStore.active.hasActiveSource) {
      await franchiseStore.replaceActiveFranchiseSource(franchiseStore.active.id, path, leagueGUID)
    } else {
      await franchiseStore.setInitialSource(franchiseStore.active.id, path, leagueGUID)
    }
    showSaveFilePicker.value = false
  } catch (e) {
    saveFileError.value = String(e)
  }
}

// Adds a fork source for a franchise that was exported to a new league.
async function handleForkSourceChange(path: string, leagueGUID: string) {
  if (!franchiseStore.active) return
  forkError.value = null
  try {
    await franchiseStore.addFranchiseSource(franchiseStore.active.id, path, leagueGUID, forkSeasonOffset.value)
    showForkPicker.value = false
  } catch (e) {
    forkError.value = String(e)
  }
}

function openForkPicker() {
  forkSeasonOffset.value = franchiseStore.active?.lastSeason ?? 0
  showForkPicker.value = true
}

// ── Sync form ────────────────────────────────────────────────────────────────

const syncing = ref(false)
const syncError = ref<string | null>(null)
const lastResult = ref<main.SyncSeasonResult | null>(null)

async function handleSync() {
  syncing.value = true
  syncError.value = null
  lastResult.value = null
  try {
    lastResult.value = await SyncSeason()
    if (franchiseStore.active) {
      await franchiseStore.selectFranchise(franchiseStore.active.id)
    }
    await loadDashboardData()
  } catch (e) {
    syncError.value = String(e)
  } finally {
    syncing.value = false
  }
}

// ── Dashboard data ───────────────────────────────────────────────────────────

const seasons = ref<main.SeasonSummaryDTO[]>([])
const selectedSeasonID = ref<number | null>(null)
const standings = ref<main.TeamStandingDTO[]>([])
const statLeaders = ref<main.StatLeadersDTO | null>(null)
const careerLeaders = ref<main.CareerLeadersDTO | null>(null)

const loadingSeasons = ref(false)
const loadingStandings = ref(false)
const loadingLeaders = ref(false)
const loadingCareer = ref(false)
const dataError = ref<string | null>(null)

const mostRecentSeason = computed(() => (seasons.value.length > 0 ? seasons.value[seasons.value.length - 1] : null))

async function loadDashboardData() {
  loadingSeasons.value = true
  dataError.value = null
  try {
    seasons.value = await GetSeasonList()
    if (seasons.value.length > 0 && selectedSeasonID.value === null) {
      selectedSeasonID.value = seasons.value[seasons.value.length - 1].id
    }
    await Promise.all([loadSeasonData(), loadCareer()])
  } catch (e) {
    dataError.value = String(e)
  } finally {
    loadingSeasons.value = false
  }
}

async function loadSeasonData() {
  if (selectedSeasonID.value === null) return
  loadingStandings.value = true
  loadingLeaders.value = true
  try {
    const [s, l] = await Promise.all([
      GetStandings(selectedSeasonID.value),
      GetSeasonStatLeaders(selectedSeasonID.value),
    ])
    standings.value = s
    statLeaders.value = l
  } catch (e) {
    dataError.value = String(e)
  } finally {
    loadingStandings.value = false
    loadingLeaders.value = false
  }
}

async function loadCareer() {
  loadingCareer.value = true
  try {
    careerLeaders.value = await GetCareerLeaders()
  } catch (e) {
    dataError.value = String(e)
  } finally {
    loadingCareer.value = false
  }
}

watch(selectedSeasonID, (id) => {
  if (id !== null) loadSeasonData()
})

onMounted(loadDashboardData)
</script>

<template>
  <div class="dashboard">
    <header class="page-header">
      <h2>{{ franchiseStore.active?.name }}</h2>
      <span class="last-synced">
        {{
          franchiseStore.active?.lastSynced
            ? `Last synced: Season ${franchiseStore.active.lastSeason} · ${new Date(franchiseStore.active.lastSynced).toLocaleDateString()}`
            : 'Never synced'
        }}
      </span>
    </header>

    <p v-if="dataError" class="error-text">{{ dataError }}</p>

    <!-- Save file configuration -->
    <section class="save-file-section">
      <div class="section-header-row">
        <h3>Save File</h3>
        <div v-if="franchiseStore.active?.hasActiveSource && !showSaveFilePicker && !showForkPicker" class="source-actions">
          <AppButton variant="ghost" size="sm" @click="showSaveFilePicker = true">Replace file</AppButton>
          <AppButton variant="ghost" size="sm" @click="openForkPicker">Add fork source</AppButton>
        </div>
      </div>

      <!-- Configured and not editing -->
      <template v-if="franchiseStore.active?.hasActiveSource && !showSaveFilePicker && !showForkPicker">
        <p class="save-path">{{ franchiseStore.active.activeSourcePath }}</p>
      </template>

      <!-- Not configured, or replacing active source -->
      <template v-else-if="!showForkPicker">
        <p v-if="!franchiseStore.active?.hasActiveSource" class="hint-text">
          Connect a save file to enable syncing.
        </p>
        <SaveFilePicker
          :selected-path="franchiseStore.active?.activeSourcePath"
          @change="handleSaveFileChange"
        />
        <p v-if="saveFileError" class="error-text">{{ saveFileError }}</p>
        <AppButton
          v-if="showSaveFilePicker"
          variant="ghost"
          size="sm"
          style="margin-top: 0.25rem"
          @click="showSaveFilePicker = false"
        >
          Cancel
        </AppButton>
      </template>

      <!-- Adding a fork source -->
      <template v-else>
        <p class="hint-text">
          Select the save game file for the forked league. Seasons from this source will
          be numbered starting after Season {{ forkSeasonOffset }}.
        </p>
        <div class="fork-offset-row">
          <label class="fork-label">Season offset</label>
          <input v-model.number="forkSeasonOffset" type="number" min="0" class="fork-offset-input" />
        </div>
        <SaveFilePicker @change="handleForkSourceChange" />
        <p v-if="forkError" class="error-text">{{ forkError }}</p>
        <AppButton variant="ghost" size="sm" style="margin-top: 0.25rem" @click="showForkPicker = false">
          Cancel
        </AppButton>
      </template>
    </section>

    <!-- Sync -->
    <section class="sync-section">
      <h3>Sync Season</h3>
      <p class="sync-help">
        Reads the current season from your save file. Sync once after the regular
        season ends, then again after the playoffs conclude —
        <strong>before</strong> progressing to the offseason. Advancing to the
        offseason triggers in-game data compaction that can cause stat loss.
      </p>
      <p v-if="syncError" class="error-text">{{ syncError }}</p>
      <div v-if="lastResult" class="sync-result">
        <span>✓ Season {{ lastResult.seasonNum }} synced —</span>
        <span>{{ lastResult.players }} players,</span>
        <span>{{ lastResult.teams }} teams,</span>
        <span>{{ lastResult.games }} games</span>
        <span v-if="lastResult.playoffGames">, {{ lastResult.playoffGames }} playoff games</span>
      </div>
      <AppButton
        variant="primary"
        :disabled="syncing || !franchiseStore.active?.hasActiveSource"
        @click="handleSync"
      >
        {{ syncing ? 'Syncing…' : 'Sync Season' }}
      </AppButton>
    </section>

    <!-- Stats only shown once at least one season is synced -->
    <template v-if="seasons.length > 0">

      <!-- Season summary bar -->
      <div v-if="mostRecentSeason" class="summary-bar">
        <div class="summary-item">
          <span class="summary-label">Seasons</span>
          <span class="summary-val">{{ seasons.length }}</span>
        </div>
        <div v-if="mostRecentSeason.championTeamName" class="summary-item">
          <span class="summary-label">Last Champion</span>
          <span class="summary-val">{{ mostRecentSeason.championTeamName }}</span>
        </div>
      </div>

      <!-- Season picker + stat leaders -->
      <section class="section">
        <div class="section-header">
          <h3>Season Leaders</h3>
          <SeasonSelector
            v-model="selectedSeasonID"
            :seasons="seasons"
          />
        </div>
        <StatLeadersPanel
          :leaders="statLeaders"
          :loading="loadingLeaders"
        />
      </section>

      <!-- Career leaders -->
      <section class="section">
        <h3>All-Time Leaders</h3>
        <LoadingSpinner v-if="loadingCareer" />
        <CareerLeadersPanel v-else :leaders="careerLeaders" />
      </section>

      <!-- Standings -->
      <section class="section">
        <div class="section-header">
          <h3>Standings</h3>
        </div>
        <LoadingSpinner v-if="loadingStandings" />
        <StandingsTable v-else :standings="standings" />
      </section>

    </template>

    <LoadingSpinner v-else-if="loadingSeasons" />

    <section v-else class="placeholder-section">
      <p class="placeholder">Sync your first season to see franchise stats.</p>
    </section>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.page-header {
  display: flex;
  align-items: baseline;
  gap: 1rem;
}

h2 {
  font-size: 1.4rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.last-synced {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.save-file-section,
.sync-section {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  max-width: 560px;
}

.section-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.source-actions {
  display: flex;
  gap: 0.375rem;
}

.fork-offset-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.fork-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.fork-offset-input {
  width: 80px;
  padding: 0.25rem 0.5rem;
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  color: var(--color-text-primary);
  font-size: 0.875rem;
}

.save-path {
  font-size: 0.8125rem;
  font-family: var(--font-mono);
  color: var(--color-text-secondary);
  word-break: break-all;
}

.sync-help {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}


.sync-result {
  display: flex;
  gap: 0.375rem;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  flex-wrap: wrap;
}

.error-text { font-size: 0.875rem; color: var(--color-error); }
.hint-text  { font-size: 0.8125rem; color: var(--color-text-secondary); }

.summary-bar {
  display: flex;
  gap: 2rem;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.summary-label {
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.summary-val {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.section {
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.placeholder-section { padding: 1rem 0; }
.placeholder { color: var(--color-text-secondary); font-size: 0.9375rem; }

</style>
