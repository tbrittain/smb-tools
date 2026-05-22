<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Select from 'primevue/select'
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { GetHistoricalTeams, GetSeasonList } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const rows = ref<main.HistoricalTeamDTO[]>([])
const seasons = ref<main.SeasonSummaryDTO[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const seasonStart = ref<number | null>(null)
const seasonEnd = ref<number | null>(null)

const seasonOptions = computed(() => seasons.value.map((s) => ({ label: `Season ${s.seasonNum}`, value: s.seasonNum })))

async function fetchTeams() {
  if (seasonStart.value == null || seasonEnd.value == null) return
  loading.value = true
  error.value = null
  try {
    rows.value = await GetHistoricalTeams(seasonStart.value, seasonEnd.value)
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

watch([seasonStart, seasonEnd], () => {
  fetchTeams()
})

onMounted(async () => {
  loading.value = true
  try {
    seasons.value = await GetSeasonList()
    if (seasons.value.length > 0) {
      const nums = seasons.value.map((s) => s.seasonNum)
      seasonStart.value = Math.min(...nums)
      seasonEnd.value = Math.max(...nums)
    }
  } catch (e) {
    error.value = String(e)
    loading.value = false
  }
})

function fmtPct(v: number): string {
  return v.toFixed(3).replace(/^0/, '')
}

function fmtBA(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(3).replace(/^0/, '')
}

function fmtERA(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(2)
}

const rowCount = computed(() => {
  const n = rows.value.length
  return `${n} team${n !== 1 ? 's' : ''}`
})
</script>

<template>
  <div class="teams-page">
    <header class="page-header">
      <h2>Historical Teams</h2>
    </header>

    <div v-if="seasons.length > 0" class="filters">
      <label class="filter-label">
        Season range
        <div class="filter-range">
          <Select
            v-model="seasonStart"
            :options="seasonOptions"
            option-label="label"
            option-value="value"
            size="small"
            class="season-select"
          />
          <span class="range-sep">–</span>
          <Select
            v-model="seasonEnd"
            :options="seasonOptions"
            option-label="label"
            option-value="value"
            size="small"
            class="season-select"
          />
        </div>
      </label>
      <span class="row-count">{{ rowCount }}</span>
    </div>

    <LoadingSpinner v-if="loading" />
    <EmptyState
      v-else-if="!loading && rows.length === 0 && !error && seasons.length === 0"
      message="No seasons synced yet"
      subtext="Sync your first season to see team history."
    />
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <div v-else-if="rows.length > 0" class="grid-wrap">
      <DataTable
        :value="rows"
        sort-field="wins"
        :sort-order="-1"
        removable-sort
        size="small"
        scrollable
        scroll-height="flex"
        table-style="min-width: 1200px"
      >
        <Column field="teamName" header="Team" sortable style="min-width: 160px" frozen>
          <template #body="{ data }">
            <RouterLink :to="`/teams/${data.teamId}`" class="team-link">
              {{ data.teamName }}
            </RouterLink>
          </template>
        </Column>
        <Column field="numSeasons" header="Seasons" sortable style="width: 80px" />
        <Column field="wins" header="W" sortable style="width: 60px" />
        <Column field="losses" header="L" sortable style="width: 60px" />
        <Column field="winPct" header="PCT" sortable style="width: 70px">
          <template #body="{ data }">{{ fmtPct(data.winPct) }}</template>
        </Column>
        <Column field="gamesOver500" header="G>500" sortable style="width: 70px">
          <template #body="{ data }">
            <span :class="data.gamesOver500 > 0 ? 'pos' : data.gamesOver500 < 0 ? 'neg' : ''">
              {{ data.gamesOver500 > 0 ? '+' : '' }}{{ data.gamesOver500 }}
            </span>
          </template>
        </Column>
        <Column field="playoffAppearances" header="Playoff App" sortable style="width: 95px" />
        <Column field="playoffWins" header="PW" sortable style="width: 55px" />
        <Column field="playoffLosses" header="PL" sortable style="width: 55px" />
        <Column field="divisionTitles" header="Div" sortable style="width: 55px" />
        <Column field="conferenceTitles" header="Conf" sortable style="width: 55px" />
        <Column field="championships" header="Champ" sortable style="width: 65px">
          <template #body="{ data }">
            <span v-if="data.championships > 0" class="champ-value">{{ data.championships }}</span>
            <span v-else>0</span>
          </template>
        </Column>
        <Column field="championshipDrought" header="Drought" sortable style="width: 75px" />
        <Column field="numPlayers" header="Players" sortable style="width: 70px" />
        <Column field="numHoF" header="HoF" sortable style="width: 55px">
          <template #body="{ data }">
            <span v-if="data.numHoF > 0" class="hof-value">{{ data.numHoF }}</span>
            <span v-else>0</span>
          </template>
        </Column>
        <Column field="runsFor" header="R" sortable style="width: 65px" />
        <Column field="runsAgainst" header="RA" sortable style="width: 65px" />
        <Column field="totalAB" header="AB" sortable style="width: 70px" />
        <Column field="totalHits" header="H" sortable style="width: 65px" />
        <Column field="totalHR" header="HR" sortable style="width: 60px" />
        <Column field="ba" header="BA" sortable style="width: 65px">
          <template #body="{ data }">{{ fmtBA(data.ba) }}</template>
        </Column>
        <Column field="era" header="ERA" sortable style="width: 65px">
          <template #body="{ data }">{{ fmtERA(data.era) }}</template>
        </Column>
      </DataTable>
    </div>
  </div>
</template>

<style scoped>
.teams-page {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  height: 100%;
  overflow: hidden;
}

.page-header h2 {
  font-size: 1.4rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.filters {
  display: flex;
  align-items: center;
  gap: 2rem;
  flex-shrink: 0;
}

.filter-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.filter-range {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.season-select {
  width: 130px;
}

.range-sep {
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
}

.row-count {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.grid-wrap {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.team-link {
  color: var(--color-text-primary);
  text-decoration: none;
}
.team-link:hover {
  color: var(--color-accent);
}

.champ-value {
  color: #d29922;
  font-weight: 600;
}

.hof-value {
  color: var(--color-accent);
  font-weight: 600;
}

.pos {
  color: #3fb950;
  font-family: var(--font-mono);
}
.neg {
  color: var(--color-error);
  font-family: var(--font-mono);
}

.error-text {
  font-size: 0.875rem;
  color: var(--color-error);
}
</style>
