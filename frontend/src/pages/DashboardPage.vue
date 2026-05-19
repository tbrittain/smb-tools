<script lang="ts" setup>
import { ref } from 'vue'
import { SyncSeason } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import { useFranchiseStore } from '../stores/franchise'

const franchiseStore = useFranchiseStore()

const seasonID = ref<number>(0)
const seasonNum = ref<number>(1)
const syncing = ref(false)
const syncError = ref<string | null>(null)
const lastResult = ref<main.SyncSeasonResult | null>(null)

async function handleSync() {
  if (!seasonID.value || !seasonNum.value) {
    syncError.value = 'Season ID and season number are required'
    return
  }
  syncing.value = true
  syncError.value = null
  lastResult.value = null
  try {
    lastResult.value = await SyncSeason(seasonID.value, seasonNum.value)
    // Refresh the active franchise to show updated last-synced
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

    <section class="sync-section">
      <h3>Sync Season</h3>
      <p class="sync-help">
        Import a season directly from the save game file. Run at the end of each
        season before simulating the offseason.
      </p>

      <div class="sync-inputs">
        <label>
          Save game season ID
          <input v-model.number="seasonID" type="number" min="1" placeholder="e.g. 100" />
        </label>
        <label>
          Season number (display)
          <input v-model.number="seasonNum" type="number" min="1" placeholder="e.g. 1" />
        </label>
      </div>

      <p v-if="syncError" class="error-text">{{ syncError }}</p>

      <div v-if="lastResult" class="sync-result">
        <span>✓ Season {{ lastResult.seasonNum }} imported —</span>
        <span>{{ lastResult.players }} players,</span>
        <span>{{ lastResult.teams }} teams,</span>
        <span>{{ lastResult.games }} games</span>
        <span v-if="lastResult.playoffGames">, {{ lastResult.playoffGames }} playoff games</span>
      </div>

      <AppButton
        variant="primary"
        :disabled="syncing || !franchiseStore.active?.saveFilePath"
        @click="handleSync"
      >
        {{ syncing ? 'Syncing…' : 'Sync Season' }}
      </AppButton>

      <p v-if="!franchiseStore.active?.saveFilePath" class="hint-text">
        No save file path configured for this franchise.
      </p>
    </section>

    <section class="placeholder-section">
      <p class="placeholder">
        Franchise stats and leaderboards — coming in Phase 5.
      </p>
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

.last-synced {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.sync-section {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 520px;
}

h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.sync-help {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.sync-inputs {
  display: flex;
  gap: 1rem;
}

.sync-inputs label {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  flex: 1;
}

.sync-inputs input {
  padding: 0.4rem 0.625rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
  font-family: var(--font-mono);
  outline: none;
}

.sync-inputs input:focus {
  border-color: var(--color-accent);
}

.sync-result {
  display: flex;
  gap: 0.375rem;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  flex-wrap: wrap;
}

.error-text {
  font-size: 0.875rem;
  color: var(--color-error);
}

.hint-text {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.placeholder-section {
  padding: 1rem 0;
}

.placeholder {
  color: var(--color-text-secondary);
  font-size: 0.9375rem;
}
</style>
