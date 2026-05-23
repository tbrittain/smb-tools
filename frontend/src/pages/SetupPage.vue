<script lang="ts" setup>
import { ref } from 'vue'
import { SyncSeason } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import SaveFilePicker from '../components/SaveFilePicker.vue'
import { useFranchiseStore } from '../stores/franchise'

const franchiseStore = useFranchiseStore()

// ── Save file configuration ──────────────────────────────────────────────────

const showSaveFilePicker = ref(false)
const saveFileError = ref<string | null>(null)
const showForkPicker = ref(false)
const forkSeasonOffset = ref(0)
const forkError = ref<string | null>(null)

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

// ── Sync ────────────────────────────────────────────────────────────────────

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
  } catch (e) {
    syncError.value = String(e)
  } finally {
    syncing.value = false
  }
}
</script>

<template>
  <div class="setup-page">
    <header class="page-header">
      <h2>Setup</h2>
      <span class="subtitle">{{ franchiseStore.active?.name }}</span>
    </header>

    <!-- Save file configuration -->
    <section class="card-section">
      <div class="section-header-row">
        <h3>Save File</h3>
        <div v-if="franchiseStore.active?.hasActiveSource && !showSaveFilePicker && !showForkPicker" class="source-actions">
          <AppButton variant="ghost" size="sm" @click="showSaveFilePicker = true">Replace file</AppButton>
          <AppButton variant="ghost" size="sm" @click="openForkPicker">Add fork source</AppButton>
        </div>
      </div>

      <template v-if="franchiseStore.active?.hasActiveSource && !showSaveFilePicker && !showForkPicker">
        <p class="save-path">{{ franchiseStore.active.activeSourcePath }}</p>
      </template>

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
    <section class="card-section">
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
  </div>
</template>

<style scoped>
.setup-page {
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

.subtitle {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.card-section {
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
</style>
