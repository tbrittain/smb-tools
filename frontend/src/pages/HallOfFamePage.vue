<script lang="ts" setup>
import InputNumber from 'primevue/inputnumber'
import Paginator, { type PageState } from 'primevue/paginator'
import { onMounted, ref, watch } from 'vue'
import { GetHoFCandidates, GetHoFInducted, SetHallOfFamer } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppLink from '../components/AppLink.vue'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'

const PAGE_SIZE = 25

const candidates = ref<main.HoFCandidateDTO[]>([])
const candidatesTotal = ref(0)
const candidatesFirst = ref(0)

const inducted = ref<main.HoFCandidateDTO[]>([])
const inductedTotal = ref(0)
const inductedFirst = ref(0)

const selected = ref<Set<number>>(new Set())
const lastSeasons = ref(1)
const loadingCandidates = ref(false)
const loadingInducted = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)

async function loadCandidates() {
  loadingCandidates.value = true
  error.value = null
  try {
    const page = Math.floor(candidatesFirst.value / PAGE_SIZE) + 1
    const result = await GetHoFCandidates(page, PAGE_SIZE, lastSeasons.value)
    candidates.value = result.items ?? []
    candidatesTotal.value = result.total ?? 0
    selected.value = new Set()
  } catch (e) {
    error.value = String(e)
  } finally {
    loadingCandidates.value = false
  }
}

async function loadInducted() {
  loadingInducted.value = true
  error.value = null
  try {
    const page = Math.floor(inductedFirst.value / PAGE_SIZE) + 1
    const result = await GetHoFInducted(page, PAGE_SIZE, lastSeasons.value)
    inducted.value = result.items ?? []
    inductedTotal.value = result.total ?? 0
  } catch (e) {
    error.value = String(e)
  } finally {
    loadingInducted.value = false
  }
}

async function load() {
  await Promise.all([loadCandidates(), loadInducted()])
}

watch(lastSeasons, () => {
  candidatesFirst.value = 0
  inductedFirst.value = 0
  load()
})

function onCandidatesPage(e: PageState) {
  candidatesFirst.value = e.first
  loadCandidates()
}

function onInductedPage(e: PageState) {
  inductedFirst.value = e.first
  loadInducted()
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
    candidatesFirst.value = 0
    inductedFirst.value = 0
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
    inductedFirst.value = 0
    candidatesFirst.value = 0
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

const { set } = useBreadcrumbs()
onMounted(() => set([{ label: 'Hall of Fame' }]))
onMounted(load)
</script>

<template>
  <div class="hof-page">
    <div class="page-header">
      <h1 class="page-title">Hall of Fame</h1>
      <div class="filter-row">
        <label class="filter-label" for="last-seasons-input">Past seasons</label>
        <InputNumber
          id="last-seasons-input"
          v-model="lastSeasons"
          :min="1"
          :allow-empty="false"
          input-class="last-seasons-input"
        />
      </div>
    </div>

    <div v-if="error" class="error-msg">{{ error }}</div>

    <div class="panels">
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

        <LoadingSpinner v-if="loadingCandidates" />

        <template v-else>
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
                <AppLink
                  :to="`/players/${p.playerId}`"
                  class="player-name"
                  @click.stop
                >
                  {{ p.lastName }}, {{ p.firstName }}
                </AppLink>
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

          <Paginator
            v-if="candidatesTotal > PAGE_SIZE"
            :first="candidatesFirst"
            :rows="PAGE_SIZE"
            :total-records="candidatesTotal"
            @page="onCandidatesPage"
          />
        </template>
      </section>

      <!-- Right: Inducted -->
      <section class="panel">
        <div class="panel-header">
          <h2 class="panel-title">Inducted ({{ inductedTotal }})</h2>
        </div>

        <LoadingSpinner v-if="loadingInducted" />

        <template v-else>
          <EmptyState
            v-if="inducted.length === 0"
            message="No Hall of Famers yet."
          />

          <div v-else class="player-list">
            <div v-for="p in inducted" :key="p.playerId" class="player-row inducted">
              <div class="player-info">
                <AppLink :to="`/players/${p.playerId}`" class="player-name">
                  {{ p.lastName }}, {{ p.firstName }}
                </AppLink>
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

          <Paginator
            v-if="inductedTotal > PAGE_SIZE"
            :first="inductedFirst"
            :rows="PAGE_SIZE"
            :total-records="inductedTotal"
            @page="onInductedPage"
          />
        </template>
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
  display: flex;
  align-items: center;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-label {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

:deep(.last-seasons-input) {
  width: 5rem;
  text-align: center;
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
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
