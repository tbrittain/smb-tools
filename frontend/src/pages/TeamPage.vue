<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import { computed, onMounted, ref } from 'vue'
import { GetLogoURLForSeason, GetTeamHistory, GetTeamTopPlayers } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import AppLink from '../components/AppLink.vue'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import TeamLogoDisplay from '../components/TeamLogoDisplay.vue'
import TeamLogoManager from '../components/TeamLogoManager.vue'
import TeamMediaGallery from '../components/TeamMediaGallery.vue'
import TeamTopPlayersTable from '../components/TeamTopPlayersTable.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'

const props = defineProps<{ teamId: number }>()

const { set } = useBreadcrumbs()

const history = ref<main.TeamHistoryDTO | null>(null)
const topPlayers = ref<main.TeamTopPlayerDTO[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const logoUrl = ref('')
const showLogoManager = ref(false)

const maxSeason = computed(() => {
  const seasons = history.value?.seasons ?? []
  return seasons.length > 0 ? Math.max(...seasons.map((s) => s.seasonNum)) : 0
})

const availableSeasons = computed(() => (history.value?.seasons ?? []).map((s) => s.seasonNum).sort((a, b) => a - b))

async function refreshLogo() {
  if (maxSeason.value > 0) {
    try {
      logoUrl.value = await GetLogoURLForSeason(props.teamId, maxSeason.value)
    } catch {
      logoUrl.value = ''
    }
  }
}

// Aggregate stats across all seasons
const summary = computed(() => {
  const seasons = history.value?.seasons ?? []
  const latest = seasons.length > 0 ? seasons[seasons.length - 1] : null
  return {
    totalWins: seasons.reduce((s, r) => s + r.wins, 0),
    totalLosses: seasons.reduce((s, r) => s + r.losses, 0),
    championships: seasons.filter((r) => r.isChampion).length,
    playoffAppearances: seasons.filter((r) => r.playoffSeed != null).length,
    seasonCount: seasons.length,
    currentName: latest?.teamName ?? '',
    conference: latest?.conferenceName ?? '',
    division: latest?.divisionName ?? '',
  }
})

function hasChange<K extends keyof main.TeamSeasonSummaryDTO>(field: K): boolean {
  const seasons = history.value?.seasons ?? []
  if (seasons.length === 0) return false
  const first = seasons[0][field]
  return seasons.some((s) => s[field] !== first)
}

const hasNameChange = computed(() => hasChange('teamName'))
const hasConferenceChange = computed(() => hasChange('conferenceName'))
const hasDivisionChange = computed(() => hasChange('divisionName'))

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
    await refreshLogo()
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
        <div class="header-top">
          <div class="header-identity">
            <TeamLogoDisplay v-if="logoUrl" :logoUrl="logoUrl" size="lg" />
            <h2>{{ summary.currentName }}</h2>
          </div>
          <div class="header-actions">
            <AppButton variant="secondary" size="sm" icon="pi pi-image" @click="showLogoManager = true">
              Manage Logos
            </AppButton>
          </div>
        </div>
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
              {{ summary.championships > 0 ? `${summary.championships} 🏆` : '0' }}
            </span>
          </div>
          <div class="hstat">
            <span class="hstat-label">Playoff Appearances</span>
            <span class="hstat-val">{{ summary.playoffAppearances }}</span>
          </div>
          <div class="hstat">
            <span class="hstat-label">Conference</span>
            <span class="hstat-val">{{ summary.conference }}</span>
          </div>
          <div class="hstat">
            <span class="hstat-label">Division</span>
            <span class="hstat-val">{{ summary.division }}</span>
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
          <Column v-if="hasNameChange" field="teamName" header="Team" sortable style="min-width: 140px" />
          <Column v-if="hasConferenceChange" field="conferenceName" sortable style="width: 80px">
            <template #header><span title="Conference">Conf</span></template>
          </Column>
          <Column v-if="hasDivisionChange" field="divisionName" sortable style="width: 80px">
            <template #header><span title="Division">Div</span></template>
          </Column>
          <Column field="wins" sortable style="width: 55px">
            <template #header><span title="Wins">W</span></template>
          </Column>
          <Column field="losses" sortable style="width: 55px">
            <template #header><span title="Losses">L</span></template>
          </Column>
          <Column field="winPct" sortable style="width: 68px">
            <template #header><span title="Win Percentage">PCT</span></template>
            <template #body="{ data }">{{ fmtPct(data.winPct) }}</template>
          </Column>
          <Column field="runsFor" sortable style="width: 55px">
            <template #header><span title="Runs Scored">R</span></template>
          </Column>
          <Column field="runsAgainst" sortable style="width: 55px">
            <template #header><span title="Runs Allowed">RA</span></template>
          </Column>
          <Column field="playoffSeed" sortable style="width: 62px">
            <template #header><span title="Playoff Seed">Seed</span></template>
            <template #body="{ data }">{{ data.playoffSeed ?? '—' }}</template>
          </Column>
          <Column field="playoffWins" sortable style="width: 52px">
            <template #header><span title="Playoff Wins">PW</span></template>
            <template #body="{ data }">{{ data.playoffWins ?? '—' }}</template>
          </Column>
          <Column field="playoffLosses" sortable style="width: 52px">
            <template #header><span title="Playoff Losses">PL</span></template>
            <template #body="{ data }">{{ data.playoffLosses ?? '—' }}</template>
          </Column>
          <Column field="isChampion" sortable style="width: 55px">
            <template #header><span title="League Champion">Champ</span></template>
            <template #body="{ data }">
              <span v-if="data.isChampion">🏆</span>
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

      <!-- Media gallery — all seasons, grouped by season -->
      <section class="section">
        <TeamMediaGallery :team-id="teamId" />
      </section>
    </template>

    <EmptyState v-else message="Team not found" />
  </div>

  <TeamLogoManager
    v-if="history"
    v-model:visible="showLogoManager"
    :teamId="teamId"
    :latestSeason="maxSeason"
    :availableSeasons="availableSeasons"
    @hide="refreshLogo"
  />
</template>

<style scoped>
.team-page {
  padding-bottom: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

.page-header {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 2rem 2rem 0;
}

.header-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.header-identity {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-actions {
  flex-shrink: 0;
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
  padding: 0 2rem;
}

h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}


.error-text { font-size: 0.875rem; color: var(--color-error); }
</style>
