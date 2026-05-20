<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ListAllTeamSeasons } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const rows = ref<main.TeamSeasonListDTO[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const filterMinSeason = ref<number | null>(null)
const filterMaxSeason = ref<number | null>(null)

const seasonRange = computed(() => {
  if (rows.value.length === 0) return { min: 1, max: 1 }
  const nums = rows.value.map((r) => r.seasonNum)
  return { min: Math.min(...nums), max: Math.max(...nums) }
})

// Enrich rows with computed runDiff for sortable column
const filtered = computed(() =>
  rows.value
    .filter((r) => {
      if (filterMinSeason.value != null && r.seasonNum < filterMinSeason.value) return false
      if (filterMaxSeason.value != null && r.seasonNum > filterMaxSeason.value) return false
      return true
    })
    .map((r) => ({ ...r, runDiff: r.runsFor - r.runsAgainst })),
)

function fmtPct(v: number): string {
  return v.toFixed(3).replace(/^0/, '')
}

onMounted(async () => {
  loading.value = true
  try {
    rows.value = await ListAllTeamSeasons()
    if (rows.value.length > 0) {
      filterMinSeason.value = seasonRange.value.min
      filterMaxSeason.value = seasonRange.value.max
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
              :min="seasonRange.min"
              :max="seasonRange.max"
              class="range-input"
            />
            <span class="range-sep">–</span>
            <input
              v-model.number="filterMaxSeason"
              type="number"
              :min="seasonRange.min"
              :max="seasonRange.max"
              class="range-input"
            />
          </div>
        </label>
        <span class="row-count">{{ filtered.length }} team seasons</span>
      </div>

      <div class="grid-wrap">
        <DataTable
          :value="filtered"
          sort-field="seasonNum"
          :sort-order="-1"
          removable-sort
          size="small"
          scrollable
          scroll-height="flex"
          table-style="min-width: 800px"
        >
          <Column field="seasonNum" header="Season" sortable style="width: 80px" />
          <Column field="teamName" header="Team" sortable style="min-width: 160px">
            <template #body="{ data }">
              <RouterLink
                :to="`/teams/${data.teamId}/seasons/${data.historyId}`"
                class="team-link"
              >
                {{ data.teamName }}
              </RouterLink>
              <span v-if="data.isChampion" class="champ-star">★</span>
            </template>
          </Column>
          <Column field="conferenceName" header="Conf" sortable style="width: 90px" />
          <Column field="divisionName" header="Div" sortable style="width: 90px" />
          <Column field="wins" header="W" sortable style="width: 55px" />
          <Column field="losses" header="L" sortable style="width: 55px" />
          <Column field="winPct" header="PCT" sortable style="width: 70px">
            <template #body="{ data }">{{ fmtPct(data.winPct) }}</template>
          </Column>
          <Column field="runsFor" header="R" sortable style="width: 60px" />
          <Column field="runsAgainst" header="RA" sortable style="width: 60px" />
          <Column field="runDiff" header="DIFF" sortable style="width: 70px">
            <template #body="{ data }">
              <span :class="data.runDiff >= 0 ? 'pos' : 'neg'">
                {{ data.runDiff > 0 ? '+' : '' }}{{ data.runDiff }}
              </span>
            </template>
          </Column>
          <Column field="playoffSeed" header="Seed" sortable style="width: 65px">
            <template #body="{ data }">{{ data.playoffSeed ?? '—' }}</template>
          </Column>
          <Column field="playoffWins" header="PW" sortable style="width: 55px">
            <template #body="{ data }">{{ data.playoffWins ?? '—' }}</template>
          </Column>
          <Column field="playoffLosses" header="PL" sortable style="width: 55px">
            <template #body="{ data }">{{ data.playoffLosses ?? '—' }}</template>
          </Column>
          <Column field="isChampion" header="Champ" sortable style="width: 70px">
            <template #body="{ data }">
              <span v-if="data.isChampion" class="champ-star">★</span>
            </template>
          </Column>
        </DataTable>
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

.range-sep { color: var(--color-text-secondary); font-size: 0.8125rem; }

.row-count { font-size: 0.8125rem; color: var(--color-text-secondary); }

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
.team-link:hover { color: var(--color-accent); }

.champ-star {
  color: #d29922;
  margin-left: 0.375rem;
}

.pos { color: #3fb950; font-family: var(--font-mono); }
.neg { color: var(--color-error); font-family: var(--font-mono); }

.error-text { font-size: 0.875rem; color: var(--color-error); }
</style>
