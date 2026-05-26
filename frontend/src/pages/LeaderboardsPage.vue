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

type LeaderboardTab = 'batting-career' | 'batting-season' | 'pitching-career' | 'pitching-season'

const { set } = useBreadcrumbs()

const activeTab = ref<LeaderboardTab>('batting-career')

function defaultFilters(): main.LeaderboardFiltersDTO {
  return new main.LeaderboardFiltersDTO({
    isPlayoffs: false,
    onlyHallOfFamers: false,
    position: '',
    batHand: '',
    throwHand: '',
    chemistryType: '',
    seasonStart: 0,
    seasonEnd: 0,
    sortField: '',
    sortDesc: true,
    offset: 0,
    pageSize: 50,
  })
}

const filters = ref<main.LeaderboardFiltersDTO>(defaultFilters())
const seasons = ref<main.SeasonSummaryDTO[]>([])

const battingCareerRows = ref<main.BattingLeaderRowDTO[]>([])
const battingSeasonRows = ref<main.BattingLeaderRowDTO[]>([])
const battingSeasonTotal = ref(0)
const pitchingCareerRows = ref<main.PitchingLeaderRowDTO[]>([])
const pitchingSeasonRows = ref<main.PitchingLeaderRowDTO[]>([])
const pitchingSeasonTotal = ref(0)

// Separate sort/page state for each season leaderboard so switching tabs
// preserves the user's position.
const battingSeasonSort = ref({ field: '', desc: true })
const battingSeasonFirst = ref(0)
const pitchingSeasonSort = ref({ field: '', desc: true })
const pitchingSeasonFirst = ref(0)

const loading = ref(false)
const error = ref<string | null>(null)

const isCareer = computed(() => activeTab.value === 'batting-career' || activeTab.value === 'pitching-career')
const filterMode = computed<'batting' | 'pitching'>(() =>
  activeTab.value.startsWith('batting') ? 'batting' : 'pitching',
)

// Reset page to 0 when filters or tab change; keep sort.
watch(
  [activeTab, filters],
  () => {
    battingSeasonFirst.value = 0
    pitchingSeasonFirst.value = 0
    loadCurrentTab()
  },
  { deep: true },
)

onMounted(async () => {
  set([{ label: 'Leaderboards' }])
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
    const f = filters.value
    switch (activeTab.value) {
      case 'batting-career':
        battingCareerRows.value = await GetBattingCareerLeaders(f)
        break
      case 'batting-season': {
        const s = battingSeasonSort.value
        const page = await GetBattingSeasonLeaders(
          new main.LeaderboardFiltersDTO({
            ...f,
            sortField: s.field,
            sortDesc: s.desc,
            offset: battingSeasonFirst.value,
            pageSize: 50,
          }),
        )
        battingSeasonRows.value = page.rows ?? []
        battingSeasonTotal.value = page.total ?? 0
        break
      }
      case 'pitching-career':
        pitchingCareerRows.value = await GetPitchingCareerLeaders(f)
        break
      case 'pitching-season': {
        const s = pitchingSeasonSort.value
        const page = await GetPitchingSeasonLeaders(
          new main.LeaderboardFiltersDTO({
            ...f,
            sortField: s.field,
            sortDesc: s.desc,
            offset: pitchingSeasonFirst.value,
            pageSize: 50,
          }),
        )
        pitchingSeasonRows.value = page.rows ?? []
        pitchingSeasonTotal.value = page.total ?? 0
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
    filters.value = new main.LeaderboardFiltersDTO({
      ...filters.value,
      batHand: isBatting ? filters.value.batHand : '',
      throwHand: isBatting ? '' : filters.value.throwHand,
      position: '',
    })
  }
}

function onBattingSeasonSort(event: DataTableSortEvent) {
  const field = typeof event.sortField === 'string' ? event.sortField : ''
  battingSeasonSort.value = { field, desc: (event.sortOrder ?? -1) === -1 }
  battingSeasonFirst.value = 0
  loadCurrentTab()
}

function onBattingSeasonPage(event: DataTablePageEvent) {
  battingSeasonFirst.value = event.first
  loadCurrentTab()
}

function onPitchingSeasonSort(event: DataTableSortEvent) {
  const field = typeof event.sortField === 'string' ? event.sortField : ''
  pitchingSeasonSort.value = { field, desc: (event.sortOrder ?? -1) === -1 }
  pitchingSeasonFirst.value = 0
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

    <LeaderboardFilters v-model="filters" :mode="filterMode" :seasons="seasons" />

    <LoadingSpinner v-if="loading" />
    <p v-else-if="error" class="error-text">{{ error }}</p>
    <div v-else class="grid-wrap">
      <BattingLeaderboardTable
        v-if="activeTab === 'batting-career'"
        :rows="battingCareerRows"
        :is-career="true"
      />
      <BattingLeaderboardTable
        v-else-if="activeTab === 'batting-season'"
        :rows="battingSeasonRows"
        :is-career="false"
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
      />
      <PitchingLeaderboardTable
        v-else
        :rows="pitchingSeasonRows"
        :is-career="false"
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
