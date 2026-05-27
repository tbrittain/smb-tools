<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import { computed, onMounted, ref } from 'vue'
import { GetTeamHistory, GetTeamTopPlayers } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppLink from '../components/AppLink.vue'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import TeamTopPlayersTable from '../components/TeamTopPlayersTable.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'

const props = defineProps<{ teamId: number }>()

const { set } = useBreadcrumbs()

const history = ref<main.TeamHistoryDTO | null>(null)
const topPlayers = ref<main.TeamTopPlayerDTO[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// Aggregate stats across all seasons
const summary = computed(() => {
  const seasons = history.value?.seasons ?? []
  return {
    totalWins: seasons.reduce((s, r) => s + r.wins, 0),
    totalLosses: seasons.reduce((s, r) => s + r.losses, 0),
    championships: seasons.filter((r) => r.isChampion).length,
    playoffSeasons: seasons.filter((r) => r.playoffSeed != null).length,
    seasonCount: seasons.length,
    currentName: seasons.length > 0 ? seasons[seasons.length - 1].teamName : '',
  }
})

function fmtPct(v: number): string {
  return v.toFixed(3).replace(/^0/, '')
}

onMounted(async () => {
  loading.value = true
  error.value = null
  try {
    const [hist, players] = await Promise.all([GetTeamHistory(props.teamId), GetTeamTopPlayers(props.teamId)])
    history.value = hist
    topPlayers.value = players ?? []
    set([{ label: summary.value.currentName }])
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="team-page">
    <LoadingSpinner v-if="loading" />
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <template v-else-if="history">
      <!-- Header -->
      <header class="page-header">
        <h2>{{ summary.currentName }}</h2>
        <div class="header-stats">
          <div class="hstat">
            <span class="hstat-label">Seasons</span>
            <span class="hstat-val">{{ summary.seasonCount }}</span>
          </div>
          <div class="hstat">
            <span class="hstat-label">Record</span>
            <span class="hstat-val mono">{{ summary.totalWins }}–{{ summary.totalLosses }}</span>
          </div>
          <div class="hstat">
            <span class="hstat-label">Championships</span>
            <span class="hstat-val">
              {{ summary.championships > 0 ? `${summary.championships} ★` : '0' }}
            </span>
          </div>
          <div class="hstat">
            <span class="hstat-label">Playoff Seasons</span>
            <span class="hstat-val">{{ summary.playoffSeasons }}</span>
          </div>
        </div>
      </header>

      <!-- Season history table -->
      <section class="section">
        <h3>Season History</h3>
        <EmptyState v-if="history.seasons.length === 0" message="No seasons recorded" />
        <DataTable
          v-else
          :value="history.seasons"
          sort-field="seasonNum"
          :sort-order="-1"
          size="small"
          removable-sort
          paginator
          :rows="20"
          :rows-per-page-options="[10, 20, 50]"
        >
          <Column field="seasonNum" header="Season" sortable style="width: 80px">
            <template #body="{ data }">
              <AppLink :to="`/teams/${teamId}/seasons/${data.historyId}`">
                {{ data.seasonNum }}
              </AppLink>
            </template>
          </Column>
          <Column field="teamName" header="Team" sortable style="min-width: 140px" />
          <Column field="conferenceName" header="Conf" sortable style="width: 80px" />
          <Column field="divisionName" header="Div" sortable style="width: 80px" />
          <Column field="wins" header="W" sortable style="width: 55px" />
          <Column field="losses" header="L" sortable style="width: 55px" />
          <Column field="winPct" header="PCT" sortable style="width: 68px">
            <template #body="{ data }">{{ fmtPct(data.winPct) }}</template>
          </Column>
          <Column field="runsFor" header="R" sortable style="width: 55px" />
          <Column field="runsAgainst" header="RA" sortable style="width: 55px" />
          <Column field="playoffSeed" header="Seed" sortable style="width: 62px">
            <template #body="{ data }">{{ data.playoffSeed ?? '—' }}</template>
          </Column>
          <Column field="playoffWins" header="PW" sortable style="width: 52px">
            <template #body="{ data }">{{ data.playoffWins ?? '—' }}</template>
          </Column>
          <Column field="playoffLosses" header="PL" sortable style="width: 52px">
            <template #body="{ data }">{{ data.playoffLosses ?? '—' }}</template>
          </Column>
          <Column field="isChampion" header="" sortable style="width: 50px">
            <template #body="{ data }">
              <span v-if="data.isChampion" class="champ-star" title="Champion">★</span>
            </template>
          </Column>
        </DataTable>
      </section>
      <!-- Top players section -->
      <section class="section">
        <h3>Top Players</h3>
        <EmptyState v-if="topPlayers.length === 0" message="No player data yet — import a season to see top players" />
        <TeamTopPlayersTable v-else :players="topPlayers" />
      </section>
    </template>

    <EmptyState v-else message="Team not found" />
  </div>
</template>

<style scoped>
.team-page {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

.page-header {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
}

.header-stats {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}

.hstat {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.hstat-label {
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.hstat-val {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.mono { font-family: var(--font-mono); }

.section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.champ-star { color: #d29922; }

.error-text { font-size: 0.875rem; color: var(--color-error); }
</style>
