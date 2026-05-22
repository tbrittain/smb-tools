<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { GetHoFCandidates, GetHoFInducted, SetHallOfFamer } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const candidates = ref<main.HoFCandidateDTO[]>([])
const inducted = ref<main.HoFCandidateDTO[]>([])
const selected = ref<Set<number>>(new Set())
const loading = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)

async function load() {
  loading.value = true
  error.value = null
  try {
    const [cands, ind] = await Promise.all([GetHoFCandidates(), GetHoFInducted()])
    candidates.value = cands ?? []
    inducted.value = ind ?? []
    selected.value = new Set()
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function toggleSelect(id: number) {
  if (selected.value.has(id)) {
    selected.value.delete(id)
  } else {
    selected.value.add(id)
  }
}

async function inductSelected() {
  if (selected.value.size === 0) return
  saving.value = true
  error.value = null
  try {
    await Promise.all([...selected.value].map((id) => SetHallOfFamer(id, true)))
    await load()
  } catch (e) {
    error.value = String(e)
  } finally {
    saving.value = false
  }
}

async function removeInductee(id: number) {
  saving.value = true
  error.value = null
  try {
    await SetHallOfFamer(id, false)
    await load()
  } catch (e) {
    error.value = String(e)
  } finally {
    saving.value = false
  }
}

function ba(p: main.HoFCandidateDTO): string {
  if (!p.atBats) return '—'
  return (p.hits / p.atBats).toFixed(3).replace(/^0/, '')
}

function era(p: main.HoFCandidateDTO): string {
  if (!p.outsPitched) return '—'
  return ((p.earnedRuns * 27) / p.outsPitched).toFixed(2)
}

onMounted(load)
</script>

<template>
  <div class="hof-page">
    <div class="page-header">
      <h1 class="page-title">Hall of Fame</h1>
    </div>

    <div v-if="error" class="error-msg">{{ error }}</div>

    <LoadingSpinner v-if="loading" />

    <div v-else class="panels">
      <!-- Left: Candidates -->
      <section class="panel">
        <div class="panel-header">
          <h2 class="panel-title">Candidates</h2>
          <button
            class="btn btn-primary"
            :disabled="saving || selected.size === 0"
            @click="inductSelected"
          >
            {{ saving ? 'Saving…' : `Induct Selected (${selected.size})` }}
          </button>
        </div>

        <EmptyState
          v-if="candidates.length === 0"
          message="No retired players eligible yet."
        />

        <div v-else class="player-list">
          <div
            v-for="p in candidates"
            :key="p.playerId"
            class="player-row"
            :class="{ selected: selected.has(p.playerId) }"
            @click="toggleSelect(p.playerId)"
          >
            <input
              type="checkbox"
              :checked="selected.has(p.playerId)"
              class="hof-checkbox"
              @click.stop="toggleSelect(p.playerId)"
            />
            <div class="player-info">
              <RouterLink
                :to="`/players/${p.playerId}`"
                class="player-name"
                @click.stop
              >
                {{ p.lastName }}, {{ p.firstName }}
              </RouterLink>
              <span class="player-meta">
                Seasons {{ p.firstSeason }}–{{ p.lastSeason }}
                ({{ p.seasons }} seasons)
              </span>
            </div>
            <div class="career-stats">
              <template v-if="p.atBats > 0">
                <span class="stat-item"><span class="stat-label">H</span>{{ p.hits }}</span>
                <span class="stat-item"><span class="stat-label">HR</span>{{ p.homeRuns }}</span>
                <span class="stat-item"><span class="stat-label">RBI</span>{{ p.rbi }}</span>
                <span class="stat-item"><span class="stat-label">BA</span>{{ ba(p) }}</span>
              </template>
              <template v-if="p.outsPitched > 0">
                <span class="stat-item"><span class="stat-label">W</span>{{ p.wins }}</span>
                <span class="stat-item"><span class="stat-label">K</span>{{ p.strikeouts }}</span>
                <span class="stat-item"><span class="stat-label">ERA</span>{{ era(p) }}</span>
              </template>
            </div>
          </div>
        </div>
      </section>

      <!-- Right: Inducted -->
      <section class="panel">
        <div class="panel-header">
          <h2 class="panel-title">Inducted ({{ inducted.length }})</h2>
        </div>

        <EmptyState
          v-if="inducted.length === 0"
          message="No Hall of Famers yet."
        />

        <div v-else class="player-list">
          <div v-for="p in inducted" :key="p.playerId" class="player-row inducted">
            <div class="player-info">
              <RouterLink :to="`/players/${p.playerId}`" class="player-name">
                {{ p.lastName }}, {{ p.firstName }}
              </RouterLink>
              <span class="player-meta">
                Seasons {{ p.firstSeason }}–{{ p.lastSeason }}
              </span>
            </div>
            <div class="career-stats">
              <template v-if="p.atBats > 0">
                <span class="stat-item"><span class="stat-label">H</span>{{ p.hits }}</span>
                <span class="stat-item"><span class="stat-label">HR</span>{{ p.homeRuns }}</span>
                <span class="stat-item"><span class="stat-label">BA</span>{{ ba(p) }}</span>
              </template>
              <template v-if="p.outsPitched > 0">
                <span class="stat-item"><span class="stat-label">W</span>{{ p.wins }}</span>
                <span class="stat-item"><span class="stat-label">ERA</span>{{ era(p) }}</span>
              </template>
            </div>
            <button
              class="btn btn-ghost remove-btn"
              :disabled="saving"
              @click="removeInductee(p.playerId)"
            >
              Remove
            </button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.hof-page {
  padding: 1.5rem 2rem;
  max-width: 1200px;
}

.page-header {
  margin-bottom: 1.25rem;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
}

.error-msg {
  color: var(--color-error, #f87171);
  font-size: 0.875rem;
  margin-bottom: 0.75rem;
}

.panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

@media (max-width: 900px) {
  .panels {
    grid-template-columns: 1fr;
  }
}

.panel {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
}

.btn {
  padding: 0.4rem 0.9rem;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 0.8rem;
  font-weight: 500;
}

.btn-primary {
  background: var(--color-accent, #4c9aff);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn-ghost {
  background: none;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}

.btn-ghost:hover:not(:disabled) {
  color: var(--color-error, #f87171);
  border-color: var(--color-error, #f87171);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.player-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.player-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.1s;
}

.player-row.selected {
  background: var(--color-surface-2);
  border-color: var(--color-accent, #4c9aff);
}

.player-row.inducted {
  cursor: default;
  background: #1a2a1a;
  border-color: #2d4d2d;
}

.player-row.inducted .player-name {
  color: #6ccc6c;
}

.hof-checkbox {
  flex-shrink: 0;
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.player-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  flex: 1;
  min-width: 0;
}

.player-name {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--color-text-primary);
  text-decoration: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.player-name:hover {
  color: var(--color-accent, #4c9aff);
}

.player-meta {
  font-size: 0.72rem;
  color: var(--color-text-secondary);
}

.career-stats {
  display: flex;
  gap: 0.625rem;
  flex-wrap: wrap;
}

.stat-item {
  font-size: 0.78rem;
  color: var(--color-text-primary);
  white-space: nowrap;
}

.stat-label {
  font-size: 0.65rem;
  color: var(--color-text-secondary);
  margin-right: 0.2rem;
}

.remove-btn {
  flex-shrink: 0;
}
</style>
