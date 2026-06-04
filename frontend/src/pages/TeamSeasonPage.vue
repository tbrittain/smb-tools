<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { GetLogoURLForSeason, GetTeamHistory, GetTeamSeasonDetail } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppLink from '../components/AppLink.vue'
import EmptyState from '../components/EmptyState.vue'
import HofBadge from '../components/HofBadge.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import MediaGallery from '../components/MediaGallery.vue'
import StatHighlightCell from '../components/StatHighlightCell.vue'
import StatHighlightLegend from '../components/StatHighlightLegend.vue'
import TeamLogoDisplay from '../components/TeamLogoDisplay.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'
import {
  formatAdjustedStat,
  formatBA,
  formatERA,
  formatFIP,
  formatIP,
  formatK9,
  formatWAR,
  formatWHIP,
} from '../composables/useStatFormatters'
import {
  highlightTooltip,
  isRateSeasonLeader,
  isRateSingleSeasonRecord,
  isSeasonLeader,
  isSingleSeasonRecord,
  rateHighlightTooltip,
} from '../composables/useStatHighlightHelpers'
import { useStatHighlightsStore } from '../stores/statHighlights'

const props = defineProps<{ teamId: number; historyId: number }>()

const { set } = useBreadcrumbs()
const router = useRouter()
const highlightsStore = useStatHighlightsStore()

const detail = ref<main.TeamSeasonDetailDTO | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const teamSeasons = ref<main.TeamSeasonSummaryDTO[]>([])
const logoUrl = ref('')

const sortedSeasons = computed(() => [...teamSeasons.value].sort((a, b) => a.seasonNum - b.seasonNum))

const currentIndex = computed(() => sortedSeasons.value.findIndex((s) => s.historyId === props.historyId))

const prevSeason = computed(() => (currentIndex.value > 0 ? sortedSeasons.value[currentIndex.value - 1] : null))

const nextSeason = computed(() =>
  currentIndex.value >= 0 && currentIndex.value < sortedSeasons.value.length - 1
    ? sortedSeasons.value[currentIndex.value + 1]
    : null,
)

function navigateSeason(historyId: number) {
  router.push(`/teams/${props.teamId}/seasons/${historyId}`)
}

type RosterView = 'batting' | 'pitching' | 'attributes'
const rosterView = ref<RosterView>('batting')
const playoffEligibleOnly = ref(false)

const visibleRoster = computed(() => {
  const roster = detail.value?.roster ?? []
  return playoffEligibleOnly.value ? roster.filter((r) => r.isOnFinalRoster) : roster
})

// Group playoff games by round number (1-based, derived from DENSE_RANK in backend)
const playoffByRound = computed(() => {
  const games = detail.value?.playoffs ?? []
  const map = new Map<number, main.PlayoffGameDTO[]>()
  for (const g of games) {
    if (!map.has(g.roundNumber)) map.set(g.roundNumber, [])
    ;(map.get(g.roundNumber) ?? []).push(g)
  }
  return map
})

function playoffWL(game: main.PlayoffGameDTO): 'W' | 'L' | '—' {
  if (game.homeScore == null || game.awayScore == null) return '—'
  const isHome = game.homeTeamHistoryId === props.historyId
  const myScore = isHome ? game.homeScore : game.awayScore
  const oppScore = isHome ? game.awayScore : game.homeScore
  return myScore > oppScore ? 'W' : 'L'
}

function seriesPlaceholders(games: main.PlayoffGameDTO[], seriesLength: number | undefined): number[] {
  if (!seriesLength || games.length === 0) return []
  const winsNeeded = Math.ceil(seriesLength / 2)
  // Track wins by team identity, not home/away role — home/away alternates within a series.
  const teamA = games[0].homeTeamHistoryId
  let teamAWins = 0
  let teamBWins = 0
  for (const g of games) {
    if (g.homeScore == null || g.awayScore == null || g.homeScore === g.awayScore) continue
    const homeWon = g.homeScore > g.awayScore
    const homeIsTeamA = g.homeTeamHistoryId === teamA
    if ((homeWon && homeIsTeamA) || (!homeWon && !homeIsTeamA)) teamAWins++
    else teamBWins++
  }
  if (Math.max(teamAWins, teamBWins) >= winsNeeded) return []
  const result: number[] = []
  for (let n = games.length + 1; n <= seriesLength; n++) {
    result.push(n)
  }
  return result
}

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

async function fetchDetail(historyId: number) {
  loading.value = true
  error.value = null
  detail.value = null
  rosterView.value = 'batting'
  playoffEligibleOnly.value = false
  try {
    detail.value = await GetTeamSeasonDetail(historyId)
    set([{ label: `${detail.value.team.teamName} Season ${detail.value.team.seasonNum}` }])
    try {
      logoUrl.value = await GetLogoURLForSeason(props.teamId, detail.value.team.seasonNum)
    } catch {
      logoUrl.value = ''
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  highlightsStore.fetch()
  const [, historyResult] = await Promise.allSettled([fetchDetail(props.historyId), GetTeamHistory(props.teamId)])
  if (historyResult.status === 'fulfilled') {
    teamSeasons.value = historyResult.value.seasons
  }
})

watch(
  () => props.historyId,
  (id) => fetchDetail(id),
)

watch(
  () => props.teamId,
  async (id) => {
    const result = await GetTeamHistory(id)
    teamSeasons.value = result.seasons
  },
)

function rosterBClass(playerId: number, statKey: string): Record<string, boolean> {
  const seasonNum = detail.value?.team.seasonNum
  if (!seasonNum) return {}
  return {
    'stat-leader': isSeasonLeader(playerId, seasonNum, statKey, highlightsStore.highlights, 'batting'),
    'stat-record': isSingleSeasonRecord(playerId, seasonNum, statKey, highlightsStore.highlights, 'batting'),
  }
}

function rosterBTip(playerId: number, statKey: string, label: string): string {
  const seasonNum = detail.value?.team.seasonNum
  if (!seasonNum) return ''
  return highlightTooltip(playerId, seasonNum, statKey, label, highlightsStore.highlights, 'batting', 'season')
}

function rosterPClass(playerId: number, statKey: string): Record<string, boolean> {
  const seasonNum = detail.value?.team.seasonNum
  if (!seasonNum) return {}
  return {
    'stat-leader': isSeasonLeader(playerId, seasonNum, statKey, highlightsStore.highlights, 'pitching'),
    'stat-record': isSingleSeasonRecord(playerId, seasonNum, statKey, highlightsStore.highlights, 'pitching'),
  }
}

function rosterPTip(playerId: number, statKey: string, label: string): string {
  const seasonNum = detail.value?.team.seasonNum
  if (!seasonNum) return ''
  return highlightTooltip(playerId, seasonNum, statKey, label, highlightsStore.highlights, 'pitching', 'season')
}

function rosterBRateClass(playerId: number, statKey: string): Record<string, boolean> {
  const seasonNum = detail.value?.team.seasonNum
  if (!seasonNum) return {}
  return {
    'stat-leader': isRateSeasonLeader(playerId, seasonNum, statKey, highlightsStore.highlights, 'batting'),
    'stat-record': isRateSingleSeasonRecord(playerId, seasonNum, statKey, highlightsStore.highlights, 'batting'),
  }
}

function rosterBRateTip(playerId: number, statKey: string, label: string): string {
  const seasonNum = detail.value?.team.seasonNum
  if (!seasonNum) return ''
  return rateHighlightTooltip(playerId, seasonNum, statKey, label, highlightsStore.highlights, 'batting', 'season')
}

function rosterPRateClass(playerId: number, statKey: string): Record<string, boolean> {
  const seasonNum = detail.value?.team.seasonNum
  if (!seasonNum) return {}
  return {
    'stat-leader': isRateSeasonLeader(playerId, seasonNum, statKey, highlightsStore.highlights, 'pitching'),
    'stat-record': isRateSingleSeasonRecord(playerId, seasonNum, statKey, highlightsStore.highlights, 'pitching'),
  }
}

function rosterPRateTip(playerId: number, statKey: string, label: string): string {
  const seasonNum = detail.value?.team.seasonNum
  if (!seasonNum) return ''
  return rateHighlightTooltip(playerId, seasonNum, statKey, label, highlightsStore.highlights, 'pitching', 'season')
}
</script>

<template>
  <div class="season-page">
    <LoadingSpinner v-if="loading" />
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <template v-else-if="detail">
      <!-- Season nav toolbar (top-right) -->
      <div v-if="sortedSeasons.length > 1" class="season-toolbar">
        <button
          class="nav-btn"
          :disabled="!prevSeason"
          @click="prevSeason && navigateSeason(prevSeason.historyId)"
        >‹ Prev Season</button>
        <button
          class="nav-btn"
          :disabled="!nextSeason"
          @click="nextSeason && navigateSeason(nextSeason.historyId)"
        >Next Season ›</button>
      </div>

      <!-- Header (constrained width) -->
      <div class="season-content">
        <header class="page-header">
          <div class="header-identity">
            <TeamLogoDisplay v-if="logoUrl" :logoUrl="logoUrl" size="lg" />
            <h2>
              <AppLink :to="`/teams/${props.teamId}`" class="team-name-link">{{ detail.team.teamName }}</AppLink>
              <span class="season-num-label">Season {{ detail.team.seasonNum }}</span>
              <span v-if="detail.team.isChampion" class="champ-badge">🏆 Champion</span>
            </h2>
          </div>
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
          <div class="roster-controls">
            <label class="playoff-filter">
              <input v-model="playoffEligibleOnly" type="checkbox" />
              Playoff eligible only
            </label>
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
        </div>

        <EmptyState v-if="visibleRoster.length === 0" message="No roster data" />

        <!-- Batting view -->
        <DataTable
          v-else-if="rosterView === 'batting'"
          :value="visibleRoster"
          sort-field="batting.smbWar"
          :sort-order="-1"
          size="small"
        >
          <Column header="Player" sortable sort-field="lastName" style="min-width: 150px">
            <template #body="{ data }">
              <AppLink :to="`/players/${data.playerId}`">
                {{ data.firstName }} {{ data.lastName }}
              </AppLink>
              <HofBadge v-if="data.isHallOfFamer" />
              <span v-if="!data.isOnFinalRoster" class="released-badge">Released</span>
            </template>
          </Column>
          <Column field="primaryPosition" header="Pos" sortable style="min-width: 55px" />
          <Column field="age" header="Age" sortable style="min-width: 50px" />
          <Column field="batting.gamesPlayed" header="G" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.batting?.gamesPlayed" />
            </template>
          </Column>
          <Column field="batting.atBats" header="AB" sortable style="min-width: 55px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.batting?.atBats" :class-map="rosterBClass(data.playerId, 'atBats')" :tooltip="rosterBTip(data.playerId, 'atBats', 'AB')" />
            </template>
          </Column>
          <Column field="batting.hits" header="H" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.batting?.hits" :class-map="rosterBClass(data.playerId, 'hits')" :tooltip="rosterBTip(data.playerId, 'hits', 'H')" />
            </template>
          </Column>
          <Column field="batting.homeRuns" header="HR" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.batting?.homeRuns" :class-map="rosterBClass(data.playerId, 'homeRuns')" :tooltip="rosterBTip(data.playerId, 'homeRuns', 'HR')" />
            </template>
          </Column>
          <Column field="batting.rbi" header="RBI" sortable style="min-width: 55px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.batting?.rbi" :class-map="rosterBClass(data.playerId, 'rbi')" :tooltip="rosterBTip(data.playerId, 'rbi', 'RBI')" />
            </template>
          </Column>
          <Column field="batting.stolenBases" header="SB" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.batting?.stolenBases" :class-map="rosterBClass(data.playerId, 'stolenBases')" :tooltip="rosterBTip(data.playerId, 'stolenBases', 'SB')" />
            </template>
          </Column>
          <Column field="batting.walks" header="BB" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.batting?.walks" :class-map="rosterBClass(data.playerId, 'walks')" :tooltip="rosterBTip(data.playerId, 'walks', 'BB')" />
            </template>
          </Column>
          <Column field="batting.ba" header="BA" sortable style="min-width: 65px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatBA(data.batting?.ba)" :class-map="rosterBRateClass(data.playerId, 'ba')" :tooltip="rosterBRateTip(data.playerId, 'ba', 'BA')" />
            </template>
          </Column>
          <Column field="batting.obp" header="OBP" sortable style="min-width: 68px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatBA(data.batting?.obp)" :class-map="rosterBRateClass(data.playerId, 'obp')" :tooltip="rosterBRateTip(data.playerId, 'obp', 'OBP')" />
            </template>
          </Column>
          <Column field="batting.slg" header="SLG" sortable style="min-width: 68px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatBA(data.batting?.slg)" :class-map="rosterBRateClass(data.playerId, 'slg')" :tooltip="rosterBRateTip(data.playerId, 'slg', 'SLG')" />
            </template>
          </Column>
          <Column field="batting.ops" header="OPS" sortable style="min-width: 72px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatBA(data.batting?.ops)" :class-map="rosterBRateClass(data.playerId, 'ops')" :tooltip="rosterBRateTip(data.playerId, 'ops', 'OPS')" />
            </template>
          </Column>
          <Column field="batting.opsPlus" header="OPS+" sortable style="min-width: 68px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatAdjustedStat(data.batting?.opsPlus)" :class-map="rosterBRateClass(data.playerId, 'opsPlus')" :tooltip="rosterBRateTip(data.playerId, 'opsPlus', 'OPS+')" />
            </template>
          </Column>
          <Column field="batting.smbWar" header="smbWAR" sortable style="min-width: 75px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatWAR(data.batting?.smbWar)" :class-map="rosterBRateClass(data.playerId, 'smbWar')" :tooltip="rosterBRateTip(data.playerId, 'smbWar', 'smbWAR')" />
            </template>
          </Column>
        </DataTable>

        <!-- Pitching view -->
        <DataTable
          v-else-if="rosterView === 'pitching'"
          :value="visibleRoster.filter((r) => r.pitching != null)"
          sort-field="pitching.smbWar"
          :sort-order="-1"
          size="small"
        >
          <Column header="Player" sortable sort-field="lastName" style="min-width: 150px">
            <template #body="{ data }">
              <AppLink :to="`/players/${data.playerId}`">
                {{ data.firstName }} {{ data.lastName }}
              </AppLink>
              <HofBadge v-if="data.isHallOfFamer" />
              <span v-if="!data.isOnFinalRoster" class="released-badge">Released</span>
            </template>
          </Column>
          <Column field="pitcherRole" header="Role" sortable style="min-width: 55px" />
          <Column field="pitching.games" header="G" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.pitching?.games" />
            </template>
          </Column>
          <Column field="pitching.gamesStarted" header="GS" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.pitching?.gamesStarted" :class-map="rosterPClass(data.playerId, 'gamesStarted')" :tooltip="rosterPTip(data.playerId, 'gamesStarted', 'GS')" />
            </template>
          </Column>
          <Column field="pitching.wins" header="W" sortable style="min-width: 48px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.pitching?.wins" :class-map="rosterPClass(data.playerId, 'wins')" :tooltip="rosterPTip(data.playerId, 'wins', 'W')" />
            </template>
          </Column>
          <Column field="pitching.losses" header="L" sortable style="min-width: 48px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.pitching?.losses" :class-map="rosterPClass(data.playerId, 'losses')" :tooltip="rosterPTip(data.playerId, 'losses', 'L')" />
            </template>
          </Column>
          <Column field="pitching.saves" header="SV" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.pitching?.saves" :class-map="rosterPClass(data.playerId, 'saves')" :tooltip="rosterPTip(data.playerId, 'saves', 'SV')" />
            </template>
          </Column>
          <Column field="pitching.outsPitched" header="IP" sortable style="min-width: 68px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.pitching != null ? formatIP(data.pitching.outsPitched) : null" :class-map="rosterPClass(data.playerId, 'outsPitched')" :tooltip="rosterPTip(data.playerId, 'outsPitched', 'IP')" />
            </template>
          </Column>
          <Column field="pitching.strikeouts" header="K" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.pitching?.strikeouts" :class-map="rosterPClass(data.playerId, 'strikeouts')" :tooltip="rosterPTip(data.playerId, 'strikeouts', 'K')" />
            </template>
          </Column>
          <Column field="pitching.walks" header="BB" sortable style="min-width: 52px">
            <template #body="{ data }">
              <StatHighlightCell :value="data.pitching?.walks" :class-map="rosterPClass(data.playerId, 'walks')" :tooltip="rosterPTip(data.playerId, 'walks', 'BB')" />
            </template>
          </Column>
          <Column field="pitching.era" header="ERA" sortable style="min-width: 68px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatERA(data.pitching?.era)" :class-map="rosterPRateClass(data.playerId, 'era')" :tooltip="rosterPRateTip(data.playerId, 'era', 'ERA')" />
            </template>
          </Column>
          <Column field="pitching.whip" header="WHIP" sortable style="min-width: 72px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatWHIP(data.pitching?.whip)" :class-map="rosterPRateClass(data.playerId, 'whip')" :tooltip="rosterPRateTip(data.playerId, 'whip', 'WHIP')" />
            </template>
          </Column>
          <Column field="pitching.k9" header="K/9" sortable style="min-width: 65px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatK9(data.pitching?.k9)" :class-map="rosterPRateClass(data.playerId, 'k9')" :tooltip="rosterPRateTip(data.playerId, 'k9', 'K/9')" />
            </template>
          </Column>
          <Column field="pitching.eraPlus" header="ERA+" sortable style="min-width: 68px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatAdjustedStat(data.pitching?.eraPlus)" :class-map="rosterPRateClass(data.playerId, 'eraPlus')" :tooltip="rosterPRateTip(data.playerId, 'eraPlus', 'ERA+')" />
            </template>
          </Column>
          <Column field="pitching.fip" header="FIP" sortable style="min-width: 65px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatFIP(data.pitching?.fip)" :class-map="rosterPRateClass(data.playerId, 'fip')" :tooltip="rosterPRateTip(data.playerId, 'fip', 'FIP')" />
            </template>
          </Column>
          <Column field="pitching.fipMinus" header="FIP-" sortable style="min-width: 65px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatAdjustedStat(data.pitching?.fipMinus)" :class-map="rosterPRateClass(data.playerId, 'fipMinus')" :tooltip="rosterPRateTip(data.playerId, 'fipMinus', 'FIP-')" />
            </template>
          </Column>
          <Column field="pitching.smbWar" header="smbWAR" sortable style="min-width: 75px">
            <template #body="{ data }">
              <StatHighlightCell :value="formatWAR(data.pitching?.smbWar)" :class-map="rosterPRateClass(data.playerId, 'smbWar')" :tooltip="rosterPRateTip(data.playerId, 'smbWar', 'smbWAR')" />
            </template>
          </Column>
        </DataTable>

        <!-- Attributes view -->
        <DataTable
          v-else
          :value="visibleRoster"
          sort-field="salary"
          :sort-order="-1"
          size="small"
        >
          <Column header="Player" sortable sort-field="lastName" style="min-width: 150px">
            <template #body="{ data }">
              <AppLink :to="`/players/${data.playerId}`">
                {{ data.firstName }} {{ data.lastName }}
              </AppLink>
              <span v-if="!data.isOnFinalRoster" class="released-badge">Released</span>
            </template>
          </Column>
          <Column field="primaryPosition" header="Pos" sortable style="min-width: 55px" />
          <Column field="power" header="POW" sortable style="min-width: 58px" />
          <Column field="contact" header="CON" sortable style="min-width: 58px" />
          <Column field="speed" header="SPD" sortable style="min-width: 58px" />
          <Column field="fielding" header="FLD" sortable style="min-width: 58px" />
          <Column field="arm" header="ARM" sortable style="min-width: 58px" />
          <Column field="velocity" header="VEL" sortable style="min-width: 58px">
            <template #body="{ data }">{{ data.velocity > 0 ? data.velocity : '—' }}</template>
          </Column>
          <Column field="junk" header="JNK" sortable style="min-width: 58px">
            <template #body="{ data }">{{ data.junk > 0 ? data.junk : '—' }}</template>
          </Column>
          <Column field="accuracy" header="ACC" sortable style="min-width: 58px">
            <template #body="{ data }">{{ data.accuracy > 0 ? data.accuracy : '—' }}</template>
          </Column>
          <Column field="salary" header="Salary" sortable style="min-width: 90px">
            <template #body="{ data }">${{ data.salary.toLocaleString() }}</template>
          </Column>
        </DataTable>

        <StatHighlightLegend v-if="rosterView !== 'attributes'" />
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
          <Column field="teamGameNum" header="#" sortable style="min-width: 48px" />
          <Column field="gameNumber" sortable style="min-width: 58px">
            <template #header>
              <span title="Season-wide game number — position among all games played by all teams this season">Glbl#</span>
            </template>
          </Column>
          <Column header="W/L" style="min-width: 48px">
            <template #body="{ data }">
              <span :class="winLoss(data) === 'W' ? 'win' : winLoss(data) === 'L' ? 'loss' : ''">
                {{ winLoss(data) }}
              </span>
            </template>
          </Column>
          <Column header="Score" style="min-width: 80px">
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
              <template v-if="data.homeTeamHistoryId === historyId">
                <AppLink :to="`/teams/${data.awayTeamId}/seasons/${data.awayTeamHistoryId}`">
                  @ {{ data.awayTeamName }}
                </AppLink>
              </template>
              <template v-else>
                <AppLink :to="`/teams/${data.homeTeamId}/seasons/${data.homeTeamHistoryId}`">
                  vs {{ data.homeTeamName }}
                </AppLink>
              </template>
            </template>
          </Column>
          <Column header="SP" style="min-width: 130px">
            <template #body="{ data }">
              <template v-if="data.homeTeamHistoryId === historyId">
                <AppLink
                  v-if="data.homePitcherPlayerId"
                  :to="`/players/${data.homePitcherPlayerId}`"
                >{{ data.homePitcherName }}</AppLink>
                <span v-else>{{ data.homePitcherName || '—' }}</span>
              </template>
              <template v-else>
                <AppLink
                  v-if="data.awayPitcherPlayerId"
                  :to="`/players/${data.awayPitcherPlayerId}`"
                >{{ data.awayPitcherName }}</AppLink>
                <span v-else>{{ data.awayPitcherName || '—' }}</span>
              </template>
            </template>
          </Column>
          <Column header="Opp SP" style="min-width: 130px">
            <template #body="{ data }">
              <template v-if="data.homeTeamHistoryId === historyId">
                <AppLink
                  v-if="data.awayPitcherPlayerId"
                  :to="`/players/${data.awayPitcherPlayerId}`"
                >{{ data.awayPitcherName }}</AppLink>
                <span v-else>{{ data.awayPitcherName || '—' }}</span>
              </template>
              <template v-else>
                <AppLink
                  v-if="data.homePitcherPlayerId"
                  :to="`/players/${data.homePitcherPlayerId}`"
                >{{ data.homePitcherName }}</AppLink>
                <span v-else>{{ data.homePitcherName || '—' }}</span>
              </template>
            </template>
          </Column>
        </DataTable>
      </section>

      <!-- Playoffs -->
      <section v-if="detail.playoffs.length > 0" class="grid-section">
        <h3>Playoffs</h3>
        <div class="playoff-panel">
          <div
            v-for="[roundNum, games] in playoffByRound"
            :key="roundNum"
            class="series-block"
          >
            <h4 class="series-label">
              {{ games[0].roundLabel }}
              <span class="series-teams">
                {{ games[0].homeTeamName }} vs {{ games[0].awayTeamName }}
              </span>
            </h4>
            <table class="series-table">
              <thead>
                <tr>
                  <th>Game</th>
                  <th>W/L</th>
                  <th>Home</th>
                  <th class="score-col">Score</th>
                  <th>Away</th>
                  <th>SP</th>
                  <th>Opp SP</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(g, idx) in games" :key="g.gameNumber">
                  <td>
                    <template v-if="detail.playoffSeriesLength != null">
                      Game {{ idx + 1 }} of {{ detail.playoffSeriesLength }}
                    </template>
                    <template v-else>{{ idx + 1 }}</template>
                  </td>
                  <td>
                    <span
                      v-if="g.homeScore != null"
                      :class="playoffWL(g) === 'W' ? 'win' : 'loss'"
                    >{{ playoffWL(g) }}</span>
                    <span v-else>—</span>
                  </td>
                  <td>
                    <template v-if="g.homeTeamHistoryId === historyId">{{ g.homeTeamName }}</template>
                    <AppLink v-else :to="`/teams/${g.homeTeamId}/seasons/${g.homeTeamHistoryId}`">
                      {{ g.homeTeamName }}
                    </AppLink>
                  </td>
                  <td class="score-col mono">
                    <template v-if="g.homeScore != null">
                      {{ g.homeScore }}–{{ g.awayScore }}
                    </template>
                    <template v-else>—</template>
                  </td>
                  <td>
                    <template v-if="g.awayTeamHistoryId === historyId">{{ g.awayTeamName }}</template>
                    <AppLink v-else :to="`/teams/${g.awayTeamId}/seasons/${g.awayTeamHistoryId}`">
                      {{ g.awayTeamName }}
                    </AppLink>
                  </td>
                  <td>
                    <template v-if="g.homeTeamHistoryId === historyId">
                      <AppLink v-if="g.homePitcherPlayerId" :to="`/players/${g.homePitcherPlayerId}`">
                        {{ g.homePitcherName }}
                      </AppLink>
                      <span v-else>{{ g.homePitcherName || '—' }}</span>
                    </template>
                    <template v-else>
                      <AppLink v-if="g.awayPitcherPlayerId" :to="`/players/${g.awayPitcherPlayerId}`">
                        {{ g.awayPitcherName }}
                      </AppLink>
                      <span v-else>{{ g.awayPitcherName || '—' }}</span>
                    </template>
                  </td>
                  <td>
                    <template v-if="g.homeTeamHistoryId === historyId">
                      <AppLink v-if="g.awayPitcherPlayerId" :to="`/players/${g.awayPitcherPlayerId}`">
                        {{ g.awayPitcherName }}
                      </AppLink>
                      <span v-else>{{ g.awayPitcherName || '—' }}</span>
                    </template>
                    <template v-else>
                      <AppLink v-if="g.homePitcherPlayerId" :to="`/players/${g.homePitcherPlayerId}`">
                        {{ g.homePitcherName }}
                      </AppLink>
                      <span v-else>{{ g.homePitcherName || '—' }}</span>
                    </template>
                  </td>
                </tr>
                <tr
                  v-for="gameNum in seriesPlaceholders(games, detail.playoffSeriesLength)"
                  :key="`upcoming-${gameNum}`"
                  class="upcoming-game"
                >
                  <td>Game {{ gameNum }} of {{ detail.playoffSeriesLength }}</td>
                  <td>—</td>
                  <td>—</td>
                  <td class="score-col">—</td>
                  <td>—</td>
                  <td>—</td>
                  <td>—</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <!-- Media gallery -->
      <section class="grid-section">
        <MediaGallery
          entity-type="team_season"
          :entity-id="historyId"
          :entity-label="`${detail.team.teamName} S${detail.team.seasonNum}`"
        />
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

.page-header {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.header-identity {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.grid-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0 2rem;
}

.season-toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.75rem 2rem 0;
}

.nav-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.875rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
}

.nav-btn:hover:not(:disabled) {
  background: var(--color-surface-2);
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.nav-btn:disabled {
  opacity: 0.35;
  cursor: default;
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

.team-name-link {
  font-size: inherit;
  font-weight: inherit;
}

.season-num-label {
  font-size: 0.875rem;
  font-weight: 400;
  color: var(--color-text-secondary);
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

.roster-controls { display: flex; align-items: center; gap: 1rem; }

.playoff-filter {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  cursor: pointer;
  user-select: none;
}
.playoff-filter input { cursor: pointer; }

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

.released-badge {
  margin-left: 0.375rem;
  font-size: 0.6rem;
  font-weight: 600;
  color: #8b949e;
  background: color-mix(in srgb, #8b949e 15%, transparent);
  border: 1px solid color-mix(in srgb, #8b949e 40%, transparent);
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

.series-table a { text-decoration: none; }
.series-table a:hover { text-decoration: underline; }

.upcoming-game td {
  color: var(--color-text-secondary);
  font-style: italic;
  border-bottom-color: color-mix(in srgb, var(--color-border) 25%, transparent);
}

.error-text { font-size: 0.875rem; color: var(--color-error); }
</style>
