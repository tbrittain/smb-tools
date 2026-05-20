<script lang="ts" setup>
import type { ColDef, GridReadyEvent, ICellRendererParams } from 'ag-grid-community'
import { AgGridVue } from 'ag-grid-vue3'
import { computed, onMounted, ref } from 'vue'
import { ListAllTeamSeasons } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const rows = ref<main.TeamSeasonListDTO[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// Season range filter
const filterMinSeason = ref<number | null>(null)
const filterMaxSeason = ref<number | null>(null)

const allSeasons = computed(() => {
  const nums = rows.value.map((r) => r.seasonNum)
  return { min: Math.min(...nums), max: Math.max(...nums) }
})

const filtered = computed(() => {
  return rows.value.filter((r) => {
    if (filterMinSeason.value != null && r.seasonNum < filterMinSeason.value) return false
    if (filterMaxSeason.value != null && r.seasonNum > filterMaxSeason.value) return false
    return true
  })
})

function fmtPct(v: number): string {
  return v.toFixed(3).replace(/^0/, '')
}

const columnDefs: ColDef<main.TeamSeasonListDTO>[] = [
  {
    headerName: 'Season',
    field: 'seasonNum',
    width: 90,
    sort: 'desc',
    pinned: 'left',
  },
  {
    headerName: 'Team',
    field: 'teamName',
    minWidth: 160,
    flex: 1,
    pinned: 'left',
    cellRenderer: (params: ICellRendererParams<main.TeamSeasonListDTO>) => {
      const row = params.data
      if (!row) return params.value as string
      const champ = row.isChampion ? ' <span style="color:#d29922;font-size:0.7em">★</span>' : ''
      return `<a href="#/teams/${row.historyId}/seasons/${row.historyId}" style="color:inherit;text-decoration:none">${params.value}${champ}</a>`
    },
  },
  { headerName: 'Conference', field: 'conferenceName', width: 120 },
  { headerName: 'Division', field: 'divisionName', width: 110 },
  { headerName: 'W', field: 'wins', width: 70, type: 'numericColumn' },
  { headerName: 'L', field: 'losses', width: 70, type: 'numericColumn' },
  {
    headerName: 'PCT',
    field: 'winPct',
    width: 80,
    type: 'numericColumn',
    valueFormatter: (p) => fmtPct(p.value as number),
  },
  { headerName: 'R', field: 'runsFor', width: 70, type: 'numericColumn' },
  { headerName: 'RA', field: 'runsAgainst', width: 70, type: 'numericColumn' },
  {
    headerName: 'DIFF',
    valueGetter: (p) => (p.data ? p.data.runsFor - p.data.runsAgainst : 0),
    width: 75,
    type: 'numericColumn',
    cellStyle: (p) => ({
      color: (p.value as number) >= 0 ? '#3fb950' : 'var(--color-error)',
    }),
    valueFormatter: (p) => {
      const v = p.value as number
      return v > 0 ? `+${v}` : String(v)
    },
  },
  {
    headerName: 'Playoff Seed',
    field: 'playoffSeed',
    width: 110,
    type: 'numericColumn',
    valueFormatter: (p) => (p.value != null ? String(p.value) : '—'),
  },
  {
    headerName: 'Playoff W',
    field: 'playoffWins',
    width: 100,
    type: 'numericColumn',
    valueFormatter: (p) => (p.value != null ? String(p.value) : '—'),
  },
  {
    headerName: 'Playoff L',
    field: 'playoffLosses',
    width: 100,
    type: 'numericColumn',
    valueFormatter: (p) => (p.value != null ? String(p.value) : '—'),
  },
  {
    headerName: 'Champion',
    field: 'isChampion',
    width: 100,
    valueFormatter: (p) => (p.value ? '★' : ''),
    cellStyle: { color: '#d29922', textAlign: 'center' },
  },
]

const defaultColDef: ColDef = {
  sortable: true,
  resizable: true,
  suppressMovable: false,
}

function onGridReady(_params: GridReadyEvent) {
  // Grid is ready; data already bound via rowData
}

onMounted(async () => {
  loading.value = true
  try {
    rows.value = await ListAllTeamSeasons()
    if (rows.value.length > 0) {
      filterMinSeason.value = allSeasons.value.min
      filterMaxSeason.value = allSeasons.value.max
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="teams-page">
    <header class="page-header">
      <h2>Historical Teams</h2>
    </header>

    <LoadingSpinner v-if="loading" />
    <EmptyState
      v-else-if="!loading && rows.length === 0 && !error"
      message="No seasons synced yet"
      subtext="Sync your first season to see team history."
    />
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <template v-else>
      <div class="filters">
        <label class="filter-label">
          Season range
          <div class="filter-range">
            <input
              v-model.number="filterMinSeason"
              type="number"
              :min="allSeasons.min"
              :max="allSeasons.max"
              class="range-input"
            />
            <span class="range-sep">–</span>
            <input
              v-model.number="filterMaxSeason"
              type="number"
              :min="allSeasons.min"
              :max="allSeasons.max"
              class="range-input"
            />
          </div>
        </label>
        <span class="row-count">{{ filtered.length }} team seasons</span>
      </div>

      <div class="grid-wrap ag-theme-alpine-dark">
        <AgGridVue
          :row-data="filtered"
          :column-defs="columnDefs"
          :default-col-def="defaultColDef"
          :animate-rows="false"
          :suppress-cell-focus="true"
          row-height="36"
          header-height="36"
          style="width: 100%; height: 100%"
          @grid-ready="onGridReady"
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
.teams-page {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  height: 100%;
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

.range-input {
  width: 64px;
  padding: 0.25rem 0.5rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.8125rem;
  font-family: var(--font-mono);
  text-align: center;
  outline: none;
}

.range-input:focus { border-color: var(--color-accent); }

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
  min-height: 400px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}

.error-text {
  font-size: 0.875rem;
  color: var(--color-error);
}
</style>
