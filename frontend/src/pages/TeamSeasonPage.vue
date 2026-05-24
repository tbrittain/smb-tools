<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { GetTeamSeasonDetail } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'
import { formatBA, formatERA, formatIP, formatK9, formatWHIP } from '../composables/useStatFormatters'

const props = defineProps<{ teamId: number; historyId: number }>()

const { set } = useBreadcrumbs()

const detail = ref<main.TeamSeasonDetailDTO | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

type RosterView = 'batting' | 'pitching' | 'attributes'
const rosterView = ref<RosterView>('batting')

// Group playoff games by series number
const playoffBySeries = computed(() => {
  const games = detail.value?.playoffs ?? []
  const map = new Map<number, main.PlayoffGameDTO[]>()
  for (const g of games) {
    if (!map.has(g.seriesNumber)) map.set(g.seriesNumber, [])
    ;(map.get(g.seriesNumber) ?? []).push(g)
    map.set(g.seriesNumber, map.get(g.seriesNumber) ?? [])
  }
  return map
})

function fmtPct(v: number): string {
  return v.toFixed(3).replace(/^0/, '')
}

function winLoss(game: main.ScheduleGameDTO): string {
  if (game.homeScore == null || game.awayScore == null) return '—'
  const isHome = game.homeTeamHistoryId === props.historyId
  const myScore = isHome ? game.homeScore : game.awayScore
  const oppScore = isHome ? game.awayScore : game.homeScore
  return myScore > oppScore ? 'W' : 'L'
}

onMounted(async () => {
  loading.value = true
  error.value = null
  try {
    detail.value = await GetTeamSeasonDetail(props.historyId)
    set([{ label: `${detail.value.team.teamName} Season ${detail.value.team.seasonNum}` }])
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="season-page">
    <LoadingSpinner v-if="loading" />
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <template v-else-if="detail">
      <!-- Header (constrained width) -->
      <div class="season-content">
        <header class="page-header">
          <h2>
            {{ detail.team.teamName }}
            <span v-if="detail.team.isChampion" class="champ-badge">★ Champion</span>
          </h2>
          <div class="header-stats">
            <div class="hstat">
              <span class="hstat-label">Record</span>
              <span class="hstat-val mono">{{ detail.team.wins }}–{{ detail.team.losses }}</span>
            </div>
            <div class="hstat">
              <span class="hstat-label">PCT</span>
              <span class="hstat-val mono">{{ fmtPct(detail.team.winPct) }}</span>
            </div>
            <div v-if="detail.team.playoffSeed" class="hstat">
              <span class="hstat-label">Playoff Seed</span>
              <span class="hstat-val">#{{ detail.team.playoffSeed }}</span>
            </div>
            <div v-if="detail.team.playoffWins != null" class="hstat">
              <span class="hstat-label">Playoff</span>
              <span class="hstat-val mono">{{ detail.team.playoffWins }}–{{ detail.team.playoffLosses }}</span>
            </div>
            <div class="hstat">
              <span class="hstat-label">R / RA</span>
              <span class="hstat-val mono">{{ detail.team.runsFor }} / {{ detail.team.runsAgainst }}</span>
            </div>
          </div>
        </header>
      </div>

      <!-- Roster (full-width grid) -->
      <section class="grid-section">
        <div class="section-header">
          <h3>Roster</h3>
          <div class="tab-bar">
            <button
              v-for="v in (['batting', 'pitching', 'attributes'] as RosterView[])"
              :key="v"
              class="tab-btn"
              :class="{ active: rosterView === v }"
              @click="rosterView = v"
            >
              {{ v.charAt(0).toUpperCase() + v.slice(1) }}
            </button>
          </div>
        </div>

        <EmptyState v-if="detail.roster.length === 0" message="No roster data" />

        <!-- Batting view -->
        <DataTable
          v-else-if="rosterView === 'batting'"
          :value="detail.roster"
          sort-field="lastName"
          :sort-order="1"
          size="small"
        >
          <Column header="Player" sortable sort-field="lastName" style="min-width: 150px">
            <template #body="{ data }">
              <RouterLink :to="`/players/${data.playerId}`" class="player-link">
                {{ data.firstName }} {{ data.lastName }}
              </RouterLink>
              <span v-if="data.isHallOfFamer" class="hof-badge">HoF</span>
            </template>
          </Column>
          <Column field="primaryPosition" header="Pos" sortable style="width: 55px" />
          <Column field="age" header="Age" sortable style="width: 50px" />
          <Column header="G" style="width: 52px">
            <template #body="{ data }">{{ data.batting?.gamesPlayed ?? '—' }}</template>
          </Column>
          <Column header="AB" style="width: 55px">
            <template #body="{ data }">{{ data.batting?.atBats ?? '—' }}</template>
          </Column>
          <Column header="H" style="width: 52px">
            <template #body="{ data }">{{ data.batting?.hits ?? '—' }}</template>
          </Column>
          <Column header="HR" style="width: 52px">
            <template #body="{ data }">{{ data.batting?.homeRuns ?? '—' }}</template>
          </Column>
          <Column header="RBI" style="width: 55px">
            <template #body="{ data }">{{ data.batting?.rbi ?? '—' }}</template>
          </Column>
          <Column header="SB" style="width: 52px">
            <template #body="{ data }">{{ data.batting?.stolenBases ?? '—' }}</template>
          </Column>
          <Column header="BB" style="width: 52px">
            <template #body="{ data }">{{ data.batting?.walks ?? '—' }}</template>
          </Column>
          <Column header="BA" style="width: 65px">
            <template #body="{ data }">{{ formatBA(data.batting?.ba) }}</template>
          </Column>
          <Column header="OBP" style="width: 68px">
            <template #body="{ data }">{{ formatBA(data.batting?.obp) }}</template>
          </Column>
          <Column header="SLG" style="width: 68px">
            <template #body="{ data }">{{ formatBA(data.batting?.slg) }}</template>
          </Column>
          <Column header="OPS" style="width: 72px">
            <template #body="{ data }">{{ formatBA(data.batting?.ops) }}</template>
          </Column>
        </DataTable>

        <!-- Pitching view -->
        <DataTable
          v-else-if="rosterView === 'pitching'"
          :value="detail.roster.filter(r => r.pitching != null)"
          sort-field="lastName"
          :sort-order="1"
          size="small"
        >
          <Column header="Player" sortable sort-field="lastName" style="min-width: 150px">
            <template #body="{ data }">
              <RouterLink :to="`/players/${data.playerId}`" class="player-link">
                {{ data.firstName }} {{ data.lastName }}
              </RouterLink>
              <span v-if="data.isHallOfFamer" class="hof-badge">HoF</span>
            </template>
          </Column>
          <Column field="pitcherRole" header="Role" sortable style="width: 55px" />
          <Column header="G" style="width: 52px">
            <template #body="{ data }">{{ data.pitching?.games ?? '—' }}</template>
          </Column>
          <Column header="GS" style="width: 52px">
            <template #body="{ data }">{{ data.pitching?.gamesStarted ?? '—' }}</template>
          </Column>
          <Column header="W" style="width: 48px">
            <template #body="{ data }">{{ data.pitching?.wins ?? '—' }}</template>
          </Column>
          <Column header="L" style="width: 48px">
            <template #body="{ data }">{{ data.pitching?.losses ?? '—' }}</template>
          </Column>
          <Column header="SV" style="width: 52px">
            <template #body="{ data }">{{ data.pitching?.saves ?? '—' }}</template>
          </Column>
          <Column header="IP" style="width: 68px">
            <template #body="{ data }">{{ data.pitching != null ? formatIP(data.pitching.outsPitched) : '—' }}</template>
          </Column>
          <Column header="K" style="width: 52px">
            <template #body="{ data }">{{ data.pitching?.strikeouts ?? '—' }}</template>
          </Column>
          <Column header="BB" style="width: 52px">
            <template #body="{ data }">{{ data.pitching?.walks ?? '—' }}</template>
          </Column>
          <Column header="ERA" style="width: 68px">
            <template #body="{ data }">{{ formatERA(data.pitching?.era) }}</template>
          </Column>
          <Column header="WHIP" style="width: 72px">
            <template #body="{ data }">{{ formatWHIP(data.pitching?.whip) }}</template>
          </Column>
          <Column header="K/9" style="width: 65px">
            <template #body="{ data }">{{ formatK9(data.pitching?.k9) }}</template>
          </Column>
        </DataTable>

        <!-- Attributes view -->
        <DataTable
          v-else
          :value="detail.roster"
          sort-field="lastName"
          :sort-order="1"
          size="small"
        >
          <Column header="Player" sortable sort-field="lastName" style="min-width: 150px">
            <template #body="{ data }">
              <RouterLink :to="`/players/${data.playerId}`" class="player-link">
                {{ data.firstName }} {{ data.lastName }}
              </RouterLink>
            </template>
          </Column>
          <Column field="primaryPosition" header="Pos" sortable style="width: 55px" />
          <Column field="power" header="POW" sortable style="width: 58px" />
          <Column field="contact" header="CON" sortable style="width: 58px" />
          <Column field="speed" header="SPD" sortable style="width: 58px" />
          <Column field="fielding" header="FLD" sortable style="width: 58px" />
          <Column field="arm" header="ARM" sortable style="width: 58px" />
          <Column field="velocity" header="VEL" sortable style="width: 58px">
            <template #body="{ data }">{{ data.velocity > 0 ? data.velocity : '—' }}</template>
          </Column>
          <Column field="junk" header="JNK" sortable style="width: 58px">
            <template #body="{ data }">{{ data.junk > 0 ? data.junk : '—' }}</template>
          </Column>
          <Column field="accuracy" header="ACC" sortable style="width: 58px">
            <template #body="{ data }">{{ data.accuracy > 0 ? data.accuracy : '—' }}</template>
          </Column>
          <Column field="salary" header="Salary" sortable style="width: 90px">
            <template #body="{ data }">${{ data.salary.toLocaleString() }}</template>
          </Column>
        </DataTable>
      </section>

      <!-- Schedule -->
      <section class="grid-section">
        <h3>Schedule <span class="record-note">({{ detail.schedule.length }} games)</span></h3>
        <EmptyState v-if="detail.schedule.length === 0" message="No schedule data" />
        <DataTable
          v-else
          :value="detail.schedule"
          sort-field="gameNumber"
          :sort-order="1"
          size="small"
          paginator
          :rows="20"
        >
          <Column field="gameNumber" header="#" sortable style="width: 50px" />
          <Column header="W/L" style="width: 48px">
            <template #body="{ data }">
              <span :class="winLoss(data) === 'W' ? 'win' : winLoss(data) === 'L' ? 'loss' : ''">
                {{ winLoss(data) }}
              </span>
            </template>
          </Column>
          <Column header="Score" style="width: 80px">
            <template #body="{ data }">
              <span v-if="data.homeScore != null" class="mono">
                {{ data.homeTeamHistoryId === historyId ? data.homeScore : data.awayScore }}–{{
                  data.homeTeamHistoryId === historyId ? data.awayScore : data.homeScore
                }}
              </span>
              <span v-else>—</span>
            </template>
          </Column>
          <Column header="Opponent" style="min-width: 130px">
            <template #body="{ data }">
              {{ data.homeTeamHistoryId === historyId ? '@ ' + data.awayTeamName : 'vs ' + data.homeTeamName }}
            </template>
          </Column>
          <Column header="SP" style="min-width: 120px">
            <template #body="{ data }">
              {{
                data.homeTeamHistoryId === historyId
                  ? data.homePitcherName || '—'
                  : data.awayPitcherName || '—'
              }}
            </template>
          </Column>
        </DataTable>
      </section>

      <!-- Playoffs -->
      <section v-if="detail.playoffs.length > 0" class="grid-section">
        <h3>Playoffs</h3>
        <div class="playoff-panel">
          <div
            v-for="[seriesNum, games] in playoffBySeries"
            :key="seriesNum"
            class="series-block"
          >
            <h4 class="series-label">
              Round {{ seriesNum }}
              <span class="series-teams">
                {{ games[0].homeTeamName }} vs {{ games[0].awayTeamName }}
              </span>
            </h4>
            <table class="series-table">
              <thead>
                <tr>
                  <th>Game</th>
                  <th>Home</th>
                  <th class="score-col">Score</th>
                  <th>Away</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="g in games" :key="g.gameNumber">
                  <td>{{ g.gameNumber }}</td>
                  <td :class="g.homeTeamHistoryId === historyId ? 'our-team' : ''">
                    {{ g.homeTeamName }}
                  </td>
                  <td class="score-col mono">
                    <template v-if="g.homeScore != null">
                      {{ g.homeScore }}–{{ g.awayScore }}
                    </template>
                    <template v-else>—</template>
                  </td>
                  <td :class="g.awayTeamHistoryId === historyId ? 'our-team' : ''">
                    {{ g.awayTeamName }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </template>

    <EmptyState v-else message="Season not found" />
  </div>
</template>

<style scoped>
.season-page {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  padding-bottom: 2rem;
}

.season-content {
  padding: 2rem 2rem 0;
  display: flex;
  flex-direction: column;
  gap: 2rem;
  max-width: 1000px;
}

.grid-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0 2rem;
}

h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.champ-badge {
  font-size: 0.75rem;
  font-weight: 600;
  color: #d29922;
  background: color-mix(in srgb, #d29922 15%, transparent);
  border: 1px solid color-mix(in srgb, #d29922 40%, transparent);
  border-radius: 4px;
  padding: 2px 8px;
}

.header-stats {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}

.hstat { display: flex; flex-direction: column; gap: 0.125rem; }
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

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.record-note { font-size: 0.75rem; font-weight: 400; color: var(--color-text-secondary); }

.tab-bar { display: flex; gap: 0.25rem; }
.tab-btn {
  padding: 0.25rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
}
.tab-btn.active { border-color: var(--color-accent); color: var(--color-accent); background: var(--color-surface-2); }
.tab-btn:hover:not(.active) { background: var(--color-surface-2); color: var(--color-text-primary); }

.player-link { color: var(--color-text-primary); text-decoration: none; }
.player-link:hover { color: var(--color-accent); }

.hof-badge {
  margin-left: 0.375rem;
  font-size: 0.6rem;
  font-weight: 600;
  color: #d29922;
  background: color-mix(in srgb, #d29922 15%, transparent);
  border: 1px solid color-mix(in srgb, #d29922 40%, transparent);
  border-radius: 3px;
  padding: 0 4px;
  vertical-align: middle;
}

.win { color: #3fb950; font-weight: 600; }
.loss { color: var(--color-error); font-weight: 600; }

/* Playoff panel */
.playoff-panel { display: flex; flex-direction: column; gap: 1.5rem; }

.series-block {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.series-label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.series-teams {
  font-weight: 400;
  color: var(--color-text-secondary);
}

.series-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.series-table th {
  text-align: left;
  padding: 0.25rem 0.5rem;
  color: var(--color-text-secondary);
  font-size: 0.6875rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-bottom: 1px solid var(--color-border);
}

.series-table td {
  padding: 0.375rem 0.5rem;
  color: var(--color-text-primary);
  border-bottom: 1px solid color-mix(in srgb, var(--color-border) 40%, transparent);
}

.score-col { text-align: center; width: 70px; }
.our-team { font-weight: 600; }

.error-text { font-size: 0.875rem; color: var(--color-error); }
</style>
