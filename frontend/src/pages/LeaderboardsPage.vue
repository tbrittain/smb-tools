<script lang="ts" setup>
import type { DataTablePageEvent, DataTableSortEvent } from 'primevue/datatable'
import { computed, onMounted, ref, watch } from 'vue'
import {
  GetBattingCareerLeaders,
  GetBattingSeasonLeaders,
  GetPitchingCareerLeaders,
  GetPitchingSeasonLeaders,
  GetSeasonList,
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import BattingLeaderboardTable from '../components/BattingLeaderboardTable.vue'
import LeaderboardFilters from '../components/LeaderboardFilters.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import PitchingLeaderboardTable from '../components/PitchingLeaderboardTable.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'
import { useStatHighlightsStore } from '../stores/statHighlights'

type LeaderboardTab = 'batting-career' | 'batting-season' | 'pitching-career' | 'pitching-season'

interface PageCache<T> {
  rows: T[]
  total: number
}

const { set } = useBreadcrumbs()
const highlightsStore = useStatHighlightsStore()

const activeTab = ref<LeaderboardTab>('batting-career')

function defaultFilters(): main.LeaderboardFiltersDTO {
  const isSeason = activeTab.value.endsWith('season')
  return new main.LeaderboardFiltersDTO({
    gameType: activeTab.value.endsWith('career') ? 'combined' : '',
    onlyHallOfFamers: false,
    position: '',
    batHand: '',
    throwHand: '',
    chemistryType: '',
    seasonStart: 0,
    seasonEnd: 0,
    traits: [],
    qualifiedOnly: isSeason,
    sortField: '',
    sortDesc: true,
    offset: 0,
    pageSize: 50,
  })
}

const filters = ref<main.LeaderboardFiltersDTO>(defaultFilters())
const seasons = ref<main.SeasonSummaryDTO[]>([])

const battingCareerRows = ref<main.BattingLeaderRowDTO[]>([])
const battingCareerTotal = ref(0)
const battingSeasonRows = ref<main.BattingLeaderRowDTO[]>([])
const battingSeasonTotal = ref(0)
const pitchingCareerRows = ref<main.PitchingLeaderRowDTO[]>([])
const pitchingCareerTotal = ref(0)
const pitchingSeasonRows = ref<main.PitchingLeaderRowDTO[]>([])
const pitchingSeasonTotal = ref(0)

const battingCareerSort = ref({ field: '', desc: true })
const battingCareerFirst = ref(0)
const battingSeasonSort = ref({ field: '', desc: true })
const battingSeasonFirst = ref(0)
const pitchingCareerSort = ref({ field: '', desc: true })
const pitchingCareerFirst = ref(0)
const pitchingSeasonSort = ref({ field: '', desc: true })
const pitchingSeasonFirst = ref(0)

// ── Prefetch caches ───────────────────────────────────────────────────────────
// Each cache is a Map keyed by a canonical representation of filter+sort+offset.
// Cleared on filter/sort changes; generation counter prevents stale writes.

let battingCareerCache = new Map<string, PageCache<main.BattingLeaderRowDTO>>()
let battingCareerGen = 0
let battingSeasonCache = new Map<string, PageCache<main.BattingLeaderRowDTO>>()
let battingSeasonGen = 0
let pitchingCareerCache = new Map<string, PageCache<main.PitchingLeaderRowDTO>>()
let pitchingCareerGen = 0
let pitchingSeasonCache = new Map<string, PageCache<main.PitchingLeaderRowDTO>>()
let pitchingSeasonGen = 0

function clearAllCaches() {
  battingCareerCache = new Map()
  battingCareerGen++
  battingSeasonCache = new Map()
  battingSeasonGen++
  pitchingCareerCache = new Map()
  pitchingCareerGen++
  pitchingSeasonCache = new Map()
  pitchingSeasonGen++
}

function cacheKey(f: main.LeaderboardFiltersDTO, sortField: string, sortDesc: boolean, offset: number): string {
  return JSON.stringify({ ...f, sortField, sortDesc, offset })
}

async function prefetchBattingCareer(offset: number, gen: number): Promise<void> {
  if (offset < 0) return
  const key = cacheKey(filters.value, battingCareerSort.value.field, battingCareerSort.value.desc, offset)
  if (battingCareerCache.has(key)) return
  try {
    const page = await GetBattingCareerLeaders(
      new main.LeaderboardFiltersDTO({
        ...filters.value,
        sortField: battingCareerSort.value.field,
        sortDesc: battingCareerSort.value.desc,
        offset,
        pageSize: 50,
      }),
    )
    if (battingCareerGen === gen) {
      battingCareerCache.set(key, { rows: page.rows ?? [], total: page.total ?? 0 })
    }
  } catch {
    /* prefetch failures are silent */
  }
}

async function prefetchBattingSeason(offset: number, gen: number): Promise<void> {
  if (offset < 0) return
  const key = cacheKey(filters.value, battingSeasonSort.value.field, battingSeasonSort.value.desc, offset)
  if (battingSeasonCache.has(key)) return
  try {
    const page = await GetBattingSeasonLeaders(
      new main.LeaderboardFiltersDTO({
        ...filters.value,
        sortField: battingSeasonSort.value.field,
        sortDesc: battingSeasonSort.value.desc,
        offset,
        pageSize: 50,
      }),
    )
    if (battingSeasonGen === gen) {
      battingSeasonCache.set(key, { rows: page.rows ?? [], total: page.total ?? 0 })
    }
  } catch {
    /* prefetch failures are silent */
  }
}

async function prefetchPitchingCareer(offset: number, gen: number): Promise<void> {
  if (offset < 0) return
  const key = cacheKey(filters.value, pitchingCareerSort.value.field, pitchingCareerSort.value.desc, offset)
  if (pitchingCareerCache.has(key)) return
  try {
    const page = await GetPitchingCareerLeaders(
      new main.LeaderboardFiltersDTO({
        ...filters.value,
        sortField: pitchingCareerSort.value.field,
        sortDesc: pitchingCareerSort.value.desc,
        offset,
        pageSize: 50,
      }),
    )
    if (pitchingCareerGen === gen) {
      pitchingCareerCache.set(key, { rows: page.rows ?? [], total: page.total ?? 0 })
    }
  } catch {
    /* prefetch failures are silent */
  }
}

async function prefetchPitchingSeason(offset: number, gen: number): Promise<void> {
  if (offset < 0) return
  const key = cacheKey(filters.value, pitchingSeasonSort.value.field, pitchingSeasonSort.value.desc, offset)
  if (pitchingSeasonCache.has(key)) return
  try {
    const page = await GetPitchingSeasonLeaders(
      new main.LeaderboardFiltersDTO({
        ...filters.value,
        sortField: pitchingSeasonSort.value.field,
        sortDesc: pitchingSeasonSort.value.desc,
        offset,
        pageSize: 50,
      }),
    )
    if (pitchingSeasonGen === gen) {
      pitchingSeasonCache.set(key, { rows: page.rows ?? [], total: page.total ?? 0 })
    }
  } catch {
    /* prefetch failures are silent */
  }
}

// ── Loading state ─────────────────────────────────────────────────────────────

const loading = ref(false)
const error = ref<string | null>(null)

const filterMode = computed<'batting' | 'pitching'>(() =>
  activeTab.value.startsWith('batting') ? 'batting' : 'pitching',
)
const isCareer = computed(() => activeTab.value === 'batting-career' || activeTab.value === 'pitching-career')

// Reset page to 0 and clear caches when filters or tab change.
watch(
  [activeTab, filters],
  () => {
    battingCareerFirst.value = 0
    battingSeasonFirst.value = 0
    pitchingCareerFirst.value = 0
    pitchingSeasonFirst.value = 0
    clearAllCaches()
    loadCurrentTab()
  },
  { deep: true },
)

onMounted(async () => {
  set([{ label: 'Leaderboards' }])
  highlightsStore.fetch()
  try {
    seasons.value = await GetSeasonList()
  } catch {
    // seasons list failing shouldn't block the leaderboard
  }
  await loadCurrentTab()
})

async function loadCurrentTab() {
  loading.value = true
  error.value = null
  try {
    switch (activeTab.value) {
      case 'batting-career': {
        const gen = battingCareerGen
        const offset = battingCareerFirst.value
        const key = cacheKey(filters.value, battingCareerSort.value.field, battingCareerSort.value.desc, offset)
        const cached = battingCareerCache.get(key)
        if (cached) {
          battingCareerRows.value = cached.rows
          battingCareerTotal.value = cached.total
        } else {
          const page = await GetBattingCareerLeaders(
            new main.LeaderboardFiltersDTO({
              ...filters.value,
              sortField: battingCareerSort.value.field,
              sortDesc: battingCareerSort.value.desc,
              offset,
              pageSize: 50,
            }),
          )
          battingCareerRows.value = page.rows ?? []
          battingCareerTotal.value = page.total ?? 0
          if (battingCareerGen === gen) {
            battingCareerCache.set(key, { rows: page.rows ?? [], total: page.total ?? 0 })
          }
        }
        void prefetchBattingCareer(offset + 50, battingCareerGen)
        if (offset > 0) void prefetchBattingCareer(offset - 50, battingCareerGen)
        break
      }
      case 'batting-season': {
        const gen = battingSeasonGen
        const offset = battingSeasonFirst.value
        const key = cacheKey(filters.value, battingSeasonSort.value.field, battingSeasonSort.value.desc, offset)
        const cached = battingSeasonCache.get(key)
        if (cached) {
          battingSeasonRows.value = cached.rows
          battingSeasonTotal.value = cached.total
        } else {
          const page = await GetBattingSeasonLeaders(
            new main.LeaderboardFiltersDTO({
              ...filters.value,
              sortField: battingSeasonSort.value.field,
              sortDesc: battingSeasonSort.value.desc,
              offset,
              pageSize: 50,
            }),
          )
          battingSeasonRows.value = page.rows ?? []
          battingSeasonTotal.value = page.total ?? 0
          if (battingSeasonGen === gen) {
            battingSeasonCache.set(key, { rows: page.rows ?? [], total: page.total ?? 0 })
          }
        }
        void prefetchBattingSeason(offset + 50, battingSeasonGen)
        if (offset > 0) void prefetchBattingSeason(offset - 50, battingSeasonGen)
        break
      }
      case 'pitching-career': {
        const gen = pitchingCareerGen
        const offset = pitchingCareerFirst.value
        const key = cacheKey(filters.value, pitchingCareerSort.value.field, pitchingCareerSort.value.desc, offset)
        const cached = pitchingCareerCache.get(key)
        if (cached) {
          pitchingCareerRows.value = cached.rows
          pitchingCareerTotal.value = cached.total
        } else {
          const page = await GetPitchingCareerLeaders(
            new main.LeaderboardFiltersDTO({
              ...filters.value,
              sortField: pitchingCareerSort.value.field,
              sortDesc: pitchingCareerSort.value.desc,
              offset,
              pageSize: 50,
            }),
          )
          pitchingCareerRows.value = page.rows ?? []
          pitchingCareerTotal.value = page.total ?? 0
          if (pitchingCareerGen === gen) {
            pitchingCareerCache.set(key, { rows: page.rows ?? [], total: page.total ?? 0 })
          }
        }
        void prefetchPitchingCareer(offset + 50, pitchingCareerGen)
        if (offset > 0) void prefetchPitchingCareer(offset - 50, pitchingCareerGen)
        break
      }
      case 'pitching-season': {
        const gen = pitchingSeasonGen
        const offset = pitchingSeasonFirst.value
        const key = cacheKey(filters.value, pitchingSeasonSort.value.field, pitchingSeasonSort.value.desc, offset)
        const cached = pitchingSeasonCache.get(key)
        if (cached) {
          pitchingSeasonRows.value = cached.rows
          pitchingSeasonTotal.value = cached.total
        } else {
          const page = await GetPitchingSeasonLeaders(
            new main.LeaderboardFiltersDTO({
              ...filters.value,
              sortField: pitchingSeasonSort.value.field,
              sortDesc: pitchingSeasonSort.value.desc,
              offset,
              pageSize: 50,
            }),
          )
          pitchingSeasonRows.value = page.rows ?? []
          pitchingSeasonTotal.value = page.total ?? 0
          if (pitchingSeasonGen === gen) {
            pitchingSeasonCache.set(key, { rows: page.rows ?? [], total: page.total ?? 0 })
          }
        }
        void prefetchPitchingSeason(offset + 50, pitchingSeasonGen)
        if (offset > 0) void prefetchPitchingSeason(offset - 50, pitchingSeasonGen)
        break
      }
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function setTab(tab: LeaderboardTab) {
  if (tab !== activeTab.value) {
    activeTab.value = tab
    const isBatting = tab.startsWith('batting')
    const isSeason = tab.endsWith('season')
    filters.value = new main.LeaderboardFiltersDTO({
      ...filters.value,
      batHand: isBatting ? filters.value.batHand : '',
      throwHand: isBatting ? '' : filters.value.throwHand,
      position: '',
      // "combined" only applies to career tabs; reset to regular when switching to season
      gameType: isSeason && filters.value.gameType === 'combined' ? '' : filters.value.gameType,
      // qualified filter only meaningful on season tabs; default on when switching to season
      qualifiedOnly: isSeason,
    })
  }
}

function onBattingCareerSort(event: DataTableSortEvent) {
  const field = typeof event.sortField === 'string' ? event.sortField : ''
  battingCareerSort.value = { field, desc: (event.sortOrder ?? -1) === -1 }
  battingCareerFirst.value = 0
  battingCareerCache = new Map()
  battingCareerGen++
  loadCurrentTab()
}

function onBattingCareerPage(event: DataTablePageEvent) {
  battingCareerFirst.value = event.first
  loadCurrentTab()
}

function onBattingSeasonSort(event: DataTableSortEvent) {
  const field = typeof event.sortField === 'string' ? event.sortField : ''
  battingSeasonSort.value = { field, desc: (event.sortOrder ?? -1) === -1 }
  battingSeasonFirst.value = 0
  battingSeasonCache = new Map()
  battingSeasonGen++
  loadCurrentTab()
}

function onBattingSeasonPage(event: DataTablePageEvent) {
  battingSeasonFirst.value = event.first
  loadCurrentTab()
}

function onPitchingCareerSort(event: DataTableSortEvent) {
  const field = typeof event.sortField === 'string' ? event.sortField : ''
  pitchingCareerSort.value = { field, desc: (event.sortOrder ?? -1) === -1 }
  pitchingCareerFirst.value = 0
  pitchingCareerCache = new Map()
  pitchingCareerGen++
  loadCurrentTab()
}

function onPitchingCareerPage(event: DataTablePageEvent) {
  pitchingCareerFirst.value = event.first
  loadCurrentTab()
}

function onPitchingSeasonSort(event: DataTableSortEvent) {
  const field = typeof event.sortField === 'string' ? event.sortField : ''
  pitchingSeasonSort.value = { field, desc: (event.sortOrder ?? -1) === -1 }
  pitchingSeasonFirst.value = 0
  pitchingSeasonCache = new Map()
  pitchingSeasonGen++
  loadCurrentTab()
}

function onPitchingSeasonPage(event: DataTablePageEvent) {
  pitchingSeasonFirst.value = event.first
  loadCurrentTab()
}
</script>

<template>
  <div class="leaderboards-page">
    <header class="page-header">
      <h2>Leaderboards</h2>
    </header>

    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'batting-career' }" @click="setTab('batting-career')">
        Batting — Career
      </button>
      <button class="tab-btn" :class="{ active: activeTab === 'batting-season' }" @click="setTab('batting-season')">
        Batting — Season
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'pitching-career' }"
        @click="setTab('pitching-career')"
      >
        Pitching — Career
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'pitching-season' }"
        @click="setTab('pitching-season')"
      >
        Pitching — Season
      </button>
    </div>

    <LeaderboardFilters v-model="filters" :mode="filterMode" :is-career="isCareer" :seasons="seasons" />

    <LoadingSpinner v-if="loading" />
    <p v-else-if="error" class="error-text">{{ error }}</p>
    <div v-else class="grid-wrap">
      <BattingLeaderboardTable
        v-if="activeTab === 'batting-career'"
        :rows="battingCareerRows"
        :is-career="true"
        :highlights="highlightsStore.highlights"
        :total-records="battingCareerTotal"
        :first="battingCareerFirst"
        :sort-field="battingCareerSort.field || 'smbWar'"
        :sort-order="battingCareerSort.desc ? -1 : 1"
        @sort="onBattingCareerSort"
        @page="onBattingCareerPage"
      />
      <BattingLeaderboardTable
        v-else-if="activeTab === 'batting-season'"
        :rows="battingSeasonRows"
        :is-career="false"
        :highlights="highlightsStore.highlights"
        :total-records="battingSeasonTotal"
        :first="battingSeasonFirst"
        :sort-field="battingSeasonSort.field || 'smbWar'"
        :sort-order="battingSeasonSort.desc ? -1 : 1"
        @sort="onBattingSeasonSort"
        @page="onBattingSeasonPage"
      />
      <PitchingLeaderboardTable
        v-else-if="activeTab === 'pitching-career'"
        :rows="pitchingCareerRows"
        :is-career="true"
        :highlights="highlightsStore.highlights"
        :total-records="pitchingCareerTotal"
        :first="pitchingCareerFirst"
        :sort-field="pitchingCareerSort.field || 'smbWar'"
        :sort-order="pitchingCareerSort.desc ? -1 : 1"
        @sort="onPitchingCareerSort"
        @page="onPitchingCareerPage"
      />
      <PitchingLeaderboardTable
        v-else
        :rows="pitchingSeasonRows"
        :is-career="false"
        :highlights="highlightsStore.highlights"
        :total-records="pitchingSeasonTotal"
        :first="pitchingSeasonFirst"
        :sort-field="pitchingSeasonSort.field || 'smbWar'"
        :sort-order="pitchingSeasonSort.desc ? -1 : 1"
        @sort="onPitchingSeasonSort"
        @page="onPitchingSeasonPage"
      />
    </div>
  </div>
</template>

<style scoped>
.leaderboards-page {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  height: 100%;
  overflow: hidden;
}

.grid-wrap {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.page-header h2 {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
}

.tab-bar {
  display: flex;
  gap: 0.25rem;
  flex-wrap: wrap;
}

.tab-btn {
  padding: 0.3rem 0.875rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
  transition:
    background 0.1s,
    color 0.1s;
}

.tab-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.tab-btn.active {
  background: var(--color-surface-2);
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.error-text {
  font-size: 0.875rem;
  color: var(--color-error);
}
</style>
