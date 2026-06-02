<script lang="ts" setup>
import Column from 'primevue/column'
import ColumnGroup from 'primevue/columngroup'
import DataTable from 'primevue/datatable'
import Row from 'primevue/row'
import { computed } from 'vue'
import type { main } from '../../wailsjs/go/models'
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
  isCareerRecordPO,
  isCareerRecordRS,
  isRateCareerRecordPO,
  isRateCareerRecordRS,
  isRateSeasonLeader,
  isRateSingleSeasonRecord,
  isSeasonLeader,
  isSingleSeasonRecord,
  rateHighlightTooltip,
} from '../composables/useStatHighlightHelpers'
import AppLink from './AppLink.vue'
import AwardBadge from './AwardBadge.vue'
import EmptyState from './EmptyState.vue'
import StatHighlightCell from './StatHighlightCell.vue'
import StatHighlightLegend from './StatHighlightLegend.vue'
import TraitList from './TraitList.vue'

const props = defineProps<{
  rows: main.PlayerSeasonLogDTO[]
  mode: 'batting' | 'pitching'
  showPlayoffs: boolean
  awardsBySeason?: Record<string, main.AwardDTO[]>
  playerId?: number
  highlights?: main.StatHighlightsDTO | null
}>()

// Flatten the selected stat block into each row for easy field access
const data = computed(() =>
  props.rows.map((r) => {
    const b = props.showPlayoffs ? r.playoffBatting : r.batting
    const p = props.showPlayoffs ? r.playoffPitching : r.pitching
    return { ...r, _b: b, _p: p }
  }),
)

const hasAwards = computed(() => Object.keys(props.awardsBySeason ?? {}).length > 0)

function teamSortKey(r: main.PlayerSeasonLogDTO): string {
  return r.teams[0]?.teamName ?? 'FA'
}

function hasFinalTeam(teams: main.TeamRefDTO[]): boolean {
  return teams.some((t) => t.sortOrder === 0)
}

// TODO: The current data model only records where a player ended a season
// (sortOrder=0 = final team; no sortOrder=0 entry = ended as FA). There is no
// way to distinguish a player who was FA all season from one who signed
// mid-season, nor to detect a player who started the season as FA before
// signing — both produce the same teams[] shape. Fixing this would require
// storing FA transitions in the save game import pipeline.

// ── Summary row types ─────────────────────────────────────────────────────────

interface BattingSummary {
  g: number
  ab: number
  h: number
  hr: number
  rbi: number
  sb: number
  bb: number
  k: number
  ba: number | null
  obp: number | null
  slg: number | null
  ops: number | null
  war: number | null
}

interface PitchingSummary {
  g: number
  gs: number
  w: number
  l: number
  sv: number
  outsPitched: number
  h: number
  er: number
  bb: number
  k: number
  era: number | null
  whip: number | null
  k9: number | null
  war: number | null
}

function computeBattingSummary(statsArr: main.CareerBattingStatsDTO[]): BattingSummary | null {
  if (statsArr.length === 0) return null
  const g = statsArr.reduce((s, r) => s + r.gamesPlayed, 0)
  const ab = statsArr.reduce((s, r) => s + r.atBats, 0)
  const h = statsArr.reduce((s, r) => s + r.hits, 0)
  const doubles = statsArr.reduce((s, r) => s + r.doubles, 0)
  const triples = statsArr.reduce((s, r) => s + r.triples, 0)
  const hr = statsArr.reduce((s, r) => s + r.homeRuns, 0)
  const rbi = statsArr.reduce((s, r) => s + r.rbi, 0)
  const sb = statsArr.reduce((s, r) => s + r.stolenBases, 0)
  const bb = statsArr.reduce((s, r) => s + r.walks, 0)
  const k = statsArr.reduce((s, r) => s + r.strikeouts, 0)
  const hbp = statsArr.reduce((s, r) => s + r.hitByPitch, 0)
  const sf = statsArr.reduce((s, r) => s + r.sacFlies, 0)
  const warValues = statsArr.map((r) => r.smbWar).filter((w): w is number => w != null)
  const war = warValues.length > 0 ? warValues.reduce((s, w) => s + w, 0) : null

  const ba = ab > 0 ? h / ab : null
  const obpDen = ab + bb + hbp + sf
  const obp = obpDen > 0 ? (h + bb + hbp) / obpDen : null
  // TB = singles + 2×2B + 3×3B + 4×HR = hits + 2B + 2×3B + 3×HR
  const tb = h + doubles + 2 * triples + 3 * hr
  const slg = ab > 0 ? tb / ab : null
  const ops = obp !== null && slg !== null ? obp + slg : null

  return { g, ab, h, hr, rbi, sb, bb, k, ba, obp, slg, ops, war }
}

function computePitchingSummary(statsArr: main.CareerPitchingStatsDTO[]): PitchingSummary | null {
  if (statsArr.length === 0) return null
  const g = statsArr.reduce((s, r) => s + r.games, 0)
  const gs = statsArr.reduce((s, r) => s + r.gamesStarted, 0)
  const w = statsArr.reduce((s, r) => s + r.wins, 0)
  const l = statsArr.reduce((s, r) => s + r.losses, 0)
  const sv = statsArr.reduce((s, r) => s + r.saves, 0)
  const outsPitched = statsArr.reduce((s, r) => s + r.outsPitched, 0)
  const h = statsArr.reduce((s, r) => s + r.hitsAllowed, 0)
  const er = statsArr.reduce((s, r) => s + r.earnedRuns, 0)
  const bb = statsArr.reduce((s, r) => s + r.walks, 0)
  const k = statsArr.reduce((s, r) => s + r.strikeouts, 0)
  const warValues = statsArr.map((r) => r.smbWar).filter((w): w is number => w != null)
  const war = warValues.length > 0 ? warValues.reduce((s, w) => s + w, 0) : null

  const era = outsPitched > 0 ? (27 * er) / outsPitched : null
  const whip = outsPitched > 0 ? (3 * (h + bb)) / outsPitched : null
  const k9 = outsPitched > 0 ? (27 * k) / outsPitched : null

  return { g, gs, w, l, sv, outsPitched, h, er, bb, k, era, whip, k9, war }
}

const rsBattingSummary = computed(() => {
  const statsArr = props.rows.map((r) => r.batting).filter((b): b is main.CareerBattingStatsDTO => b != null)
  return computeBattingSummary(statsArr)
})

const poBattingSummary = computed(() => {
  const statsArr = props.rows.map((r) => r.playoffBatting).filter((b): b is main.CareerBattingStatsDTO => b != null)
  return computeBattingSummary(statsArr)
})

const rsPitchingSummary = computed(() => {
  const statsArr = props.rows.map((r) => r.pitching).filter((p): p is main.CareerPitchingStatsDTO => p != null)
  return computePitchingSummary(statsArr)
})

const poPitchingSummary = computed(() => {
  const statsArr = props.rows.map((r) => r.playoffPitching).filter((p): p is main.CareerPitchingStatsDTO => p != null)
  return computePitchingSummary(statsArr)
})

const careerBattingSummary = computed(() => {
  const statsArr = [
    ...props.rows.map((r) => r.batting).filter((b): b is main.CareerBattingStatsDTO => b != null),
    ...props.rows.map((r) => r.playoffBatting).filter((b): b is main.CareerBattingStatsDTO => b != null),
  ]
  return computeBattingSummary(statsArr)
})

const careerPitchingSummary = computed(() => {
  const statsArr = [
    ...props.rows.map((r) => r.pitching).filter((p): p is main.CareerPitchingStatsDTO => p != null),
    ...props.rows.map((r) => r.playoffPitching).filter((p): p is main.CareerPitchingStatsDTO => p != null),
  ]
  return computePitchingSummary(statsArr)
})

// Non-stat prefix columns: Season, Team, Age, Pos/Role, Traits, [Awards]
const batchingPrefixCols = computed(() => (hasAwards.value ? 6 : 5))

// ── Highlight helpers ─────────────────────────────────────────────────────────

function bSeasonClass(r: { seasonNum: number }, statKey: string): Record<string, boolean> {
  const pid = props.playerId
  if (!pid || props.showPlayoffs) return {}
  return {
    'stat-leader': isSeasonLeader(pid, r.seasonNum, statKey, props.highlights, 'batting'),
    'stat-record': isSingleSeasonRecord(pid, r.seasonNum, statKey, props.highlights, 'batting'),
  }
}

function pSeasonClass(r: { seasonNum: number }, statKey: string): Record<string, boolean> {
  const pid = props.playerId
  if (!pid || props.showPlayoffs) return {}
  return {
    'stat-leader': isSeasonLeader(pid, r.seasonNum, statKey, props.highlights, 'pitching'),
    'stat-record': isSingleSeasonRecord(pid, r.seasonNum, statKey, props.highlights, 'pitching'),
  }
}

function bSeasonTip(r: { seasonNum: number }, statKey: string, label: string): string {
  const pid = props.playerId
  if (!pid || props.showPlayoffs) return ''
  return highlightTooltip(pid, r.seasonNum, statKey, label, props.highlights, 'batting', 'season')
}

function pSeasonTip(r: { seasonNum: number }, statKey: string, label: string): string {
  const pid = props.playerId
  if (!pid || props.showPlayoffs) return ''
  return highlightTooltip(pid, r.seasonNum, statKey, label, props.highlights, 'pitching', 'season')
}

function rsFooterClass(statKey: string, type: 'batting' | 'pitching'): Record<string, boolean> {
  const pid = props.playerId
  if (!pid) return {}
  return {
    'stat-record':
      isCareerRecordRS(pid, statKey, props.highlights, type) ||
      isRateCareerRecordRS(pid, statKey, props.highlights, type),
  }
}

function poFooterClass(statKey: string, type: 'batting' | 'pitching'): Record<string, boolean> {
  const pid = props.playerId
  if (!pid) return {}
  return {
    'stat-record':
      isCareerRecordPO(pid, statKey, props.highlights, type) ||
      isRateCareerRecordPO(pid, statKey, props.highlights, type),
  }
}

function rsFooterTip(statKey: string, label: string, type: 'batting' | 'pitching'): string {
  const pid = props.playerId
  if (!pid) return ''
  if (isCareerRecordRS(pid, statKey, props.highlights, type)) {
    return highlightTooltip(pid, 0, statKey, label, props.highlights, type, 'careerRS')
  }
  return rateHighlightTooltip(pid, 0, statKey, label, props.highlights, type, 'careerRS')
}

function poFooterTip(statKey: string, label: string, type: 'batting' | 'pitching'): string {
  const pid = props.playerId
  if (!pid) return ''
  if (isCareerRecordPO(pid, statKey, props.highlights, type)) {
    return highlightTooltip(pid, 0, statKey, label, props.highlights, type, 'careerPO')
  }
  return rateHighlightTooltip(pid, 0, statKey, label, props.highlights, type, 'careerPO')
}

function bRateSeasonClass(r: { seasonNum: number }, statKey: string): Record<string, boolean> {
  const pid = props.playerId
  if (!pid || props.showPlayoffs) return {}
  return {
    'stat-leader': isRateSeasonLeader(pid, r.seasonNum, statKey, props.highlights, 'batting'),
    'stat-record': isRateSingleSeasonRecord(pid, r.seasonNum, statKey, props.highlights, 'batting'),
  }
}

function pRateSeasonClass(r: { seasonNum: number }, statKey: string): Record<string, boolean> {
  const pid = props.playerId
  if (!pid || props.showPlayoffs) return {}
  return {
    'stat-leader': isRateSeasonLeader(pid, r.seasonNum, statKey, props.highlights, 'pitching'),
    'stat-record': isRateSingleSeasonRecord(pid, r.seasonNum, statKey, props.highlights, 'pitching'),
  }
}

function bRateSeasonTip(r: { seasonNum: number }, statKey: string, label: string): string {
  const pid = props.playerId
  if (!pid || props.showPlayoffs) return ''
  return rateHighlightTooltip(pid, r.seasonNum, statKey, label, props.highlights, 'batting', 'season')
}

function pRateSeasonTip(r: { seasonNum: number }, statKey: string, label: string): string {
  const pid = props.playerId
  if (!pid || props.showPlayoffs) return ''
  return rateHighlightTooltip(pid, r.seasonNum, statKey, label, props.highlights, 'pitching', 'season')
}
</script>

<template>
  <div class="stat-table-wrap">
    <EmptyState v-if="rows.length === 0" message="No season data" />

    <!-- Batting mode -->
    <DataTable
      v-else-if="mode === 'batting'"
      :value="data"
      sort-field="seasonNum"
      :sort-order="-1"
      size="small"
      removable-sort
      scrollable
    >
      <ColumnGroup type="footer">
        <Row v-if="rsBattingSummary">
          <Column :colspan="batchingPrefixCols" footer="Regular Season" footer-class="summary-label" />
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsBattingSummary.g" :class-map="rsFooterClass('gamesPlayed', 'batting')" :tooltip="rsFooterTip('gamesPlayed', 'G', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsBattingSummary.ab" :class-map="rsFooterClass('atBats', 'batting')" :tooltip="rsFooterTip('atBats', 'AB', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsBattingSummary.h" :class-map="rsFooterClass('hits', 'batting')" :tooltip="rsFooterTip('hits', 'H', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsBattingSummary.hr" :class-map="rsFooterClass('homeRuns', 'batting')" :tooltip="rsFooterTip('homeRuns', 'HR', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsBattingSummary.rbi" :class-map="rsFooterClass('rbi', 'batting')" :tooltip="rsFooterTip('rbi', 'RBI', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsBattingSummary.sb" :class-map="rsFooterClass('stolenBases', 'batting')" :tooltip="rsFooterTip('stolenBases', 'SB', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsBattingSummary.bb" :class-map="rsFooterClass('walks', 'batting')" :tooltip="rsFooterTip('walks', 'BB', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsBattingSummary.k" :class-map="rsFooterClass('strikeouts', 'batting')" :tooltip="rsFooterTip('strikeouts', 'K', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatBA(rsBattingSummary.ba)" :class-map="rsFooterClass('ba', 'batting')" :tooltip="rsFooterTip('ba', 'BA', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatBA(rsBattingSummary.obp)" :class-map="rsFooterClass('obp', 'batting')" :tooltip="rsFooterTip('obp', 'OBP', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatBA(rsBattingSummary.slg)" :class-map="rsFooterClass('slg', 'batting')" :tooltip="rsFooterTip('slg', 'SLG', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatBA(rsBattingSummary.ops)" :class-map="rsFooterClass('ops', 'batting')" :tooltip="rsFooterTip('ops', 'OPS', 'batting')" /></template>
          </Column>
          <Column footer="—" footer-class="summary-cell" />
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatWAR(rsBattingSummary.war)" :class-map="rsFooterClass('smbWar', 'batting')" :tooltip="rsFooterTip('smbWar', 'smbWAR', 'batting')" /></template>
          </Column>
        </Row>
        <Row v-if="poBattingSummary">
          <Column :colspan="batchingPrefixCols" footer="Playoffs" footer-class="summary-label summary-po" />
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poBattingSummary.g" :class-map="poFooterClass('gamesPlayed', 'batting')" :tooltip="poFooterTip('gamesPlayed', 'G', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poBattingSummary.ab" :class-map="poFooterClass('atBats', 'batting')" :tooltip="poFooterTip('atBats', 'AB', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poBattingSummary.h" :class-map="poFooterClass('hits', 'batting')" :tooltip="poFooterTip('hits', 'H', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poBattingSummary.hr" :class-map="poFooterClass('homeRuns', 'batting')" :tooltip="poFooterTip('homeRuns', 'HR', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poBattingSummary.rbi" :class-map="poFooterClass('rbi', 'batting')" :tooltip="poFooterTip('rbi', 'RBI', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poBattingSummary.sb" :class-map="poFooterClass('stolenBases', 'batting')" :tooltip="poFooterTip('stolenBases', 'SB', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poBattingSummary.bb" :class-map="poFooterClass('walks', 'batting')" :tooltip="poFooterTip('walks', 'BB', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poBattingSummary.k" :class-map="poFooterClass('strikeouts', 'batting')" :tooltip="poFooterTip('strikeouts', 'K', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatBA(poBattingSummary.ba)" :class-map="poFooterClass('ba', 'batting')" :tooltip="poFooterTip('ba', 'BA', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatBA(poBattingSummary.obp)" :class-map="poFooterClass('obp', 'batting')" :tooltip="poFooterTip('obp', 'OBP', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatBA(poBattingSummary.slg)" :class-map="poFooterClass('slg', 'batting')" :tooltip="poFooterTip('slg', 'SLG', 'batting')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatBA(poBattingSummary.ops)" :class-map="poFooterClass('ops', 'batting')" :tooltip="poFooterTip('ops', 'OPS', 'batting')" /></template>
          </Column>
          <Column footer="—" footer-class="summary-cell summary-po" />
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatWAR(poBattingSummary.war)" :class-map="poFooterClass('smbWar', 'batting')" :tooltip="poFooterTip('smbWar', 'smbWAR', 'batting')" /></template>
          </Column>
        </Row>
        <Row v-if="careerBattingSummary">
          <Column :colspan="batchingPrefixCols" footer="Career" footer-class="summary-label summary-career" />
          <Column :footer="String(careerBattingSummary.g)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerBattingSummary.ab)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerBattingSummary.h)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerBattingSummary.hr)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerBattingSummary.rbi)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerBattingSummary.sb)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerBattingSummary.bb)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerBattingSummary.k)" footer-class="summary-cell summary-career" />
          <Column :footer="formatBA(careerBattingSummary.ba)" footer-class="summary-cell summary-career" />
          <Column :footer="formatBA(careerBattingSummary.obp)" footer-class="summary-cell summary-career" />
          <Column :footer="formatBA(careerBattingSummary.slg)" footer-class="summary-cell summary-career" />
          <Column :footer="formatBA(careerBattingSummary.ops)" footer-class="summary-cell summary-career" />
          <Column footer="—" footer-class="summary-cell summary-career" />
          <Column :footer="formatWAR(careerBattingSummary.war)" footer-class="summary-cell summary-career" />
        </Row>
      </ColumnGroup>

      <Column field="seasonNum" header="Season" sortable style="min-width: 80px" />
      <Column header="Team" sortable :sort-field="teamSortKey" style="min-width: 130px">
        <template #body="{ data: r }">
          <span class="team-cell">
            <template v-for="(t, i) in r.teams" :key="t.teamHistoryId">
              <span v-if="i" class="team-separator"> · </span>
              <AppLink :to="`/teams/${t.teamId}/seasons/${t.teamHistoryId}`">{{ t.teamName }}</AppLink>
            </template>
            <template v-if="!hasFinalTeam(r.teams)">
              <span v-if="r.teams.length > 0" class="team-separator"> · </span>
              <span class="fa-label">FA</span>
            </template>
          </span>
        </template>
      </Column>
      <Column field="age" header="Age" sortable style="min-width: 55px" />
      <Column header="Pos" style="min-width: 70px">
        <template #body="{ data: r }">
          {{ r.primaryPosition }}<span v-if="r.secondaryPosition" class="secondary-pos">/{{ r.secondaryPosition }}</span>
        </template>
      </Column>
      <Column header="Traits" style="min-width: 160px">
        <template #body="{ data: r }">
          <TraitList :traits="r.traits" />
        </template>
      </Column>
      <Column v-if="hasAwards" header="Awards" style="min-width: 200px">
        <template #body="{ data: r }">
          <AppLink
            v-if="awardsBySeason?.[String(r.seasonNum)]?.length"
            :to="`/awards?seasonId=${r.seasonId}&view=1`"
            class="award-cell"
          >
            <AwardBadge
              v-for="award in awardsBySeason![String(r.seasonNum)]"
              :key="award.id"
              :award="award"
              size="sm"
            />
          </AppLink>
          <span v-else class="no-traits">—</span>
        </template>
      </Column>
      <Column header="G" sortable sort-field="_b.gamesPlayed" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._b?.gamesPlayed" />
        </template>
      </Column>
      <Column header="AB" sortable sort-field="_b.atBats" style="min-width: 60px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._b?.atBats" :class-map="bSeasonClass(r, 'atBats')" :tooltip="bSeasonTip(r, 'atBats', 'AB')" />
        </template>
      </Column>
      <Column header="H" sortable sort-field="_b.hits" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._b?.hits" :class-map="bSeasonClass(r, 'hits')" :tooltip="bSeasonTip(r, 'hits', 'H')" />
        </template>
      </Column>
      <Column header="HR" sortable sort-field="_b.homeRuns" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._b?.homeRuns" :class-map="bSeasonClass(r, 'homeRuns')" :tooltip="bSeasonTip(r, 'homeRuns', 'HR')" />
        </template>
      </Column>
      <Column header="RBI" sortable sort-field="_b.rbi" style="min-width: 60px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._b?.rbi" :class-map="bSeasonClass(r, 'rbi')" :tooltip="bSeasonTip(r, 'rbi', 'RBI')" />
        </template>
      </Column>
      <Column header="SB" sortable sort-field="_b.stolenBases" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._b?.stolenBases" :class-map="bSeasonClass(r, 'stolenBases')" :tooltip="bSeasonTip(r, 'stolenBases', 'SB')" />
        </template>
      </Column>
      <Column header="BB" sortable sort-field="_b.walks" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._b?.walks" :class-map="bSeasonClass(r, 'walks')" :tooltip="bSeasonTip(r, 'walks', 'BB')" />
        </template>
      </Column>
      <Column header="K" sortable sort-field="_b.strikeouts" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._b?.strikeouts" :class-map="bSeasonClass(r, 'strikeouts')" :tooltip="bSeasonTip(r, 'strikeouts', 'K')" />
        </template>
      </Column>
      <Column header="BA" sortable sort-field="_b.ba" style="min-width: 65px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatBA(r._b?.ba)" :class-map="bRateSeasonClass(r, 'ba')" :tooltip="bRateSeasonTip(r, 'ba', 'BA')" />
        </template>
      </Column>
      <Column header="OBP" sortable sort-field="_b.obp" style="min-width: 68px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatBA(r._b?.obp)" :class-map="bRateSeasonClass(r, 'obp')" :tooltip="bRateSeasonTip(r, 'obp', 'OBP')" />
        </template>
      </Column>
      <Column header="SLG" sortable sort-field="_b.slg" style="min-width: 68px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatBA(r._b?.slg)" :class-map="bRateSeasonClass(r, 'slg')" :tooltip="bRateSeasonTip(r, 'slg', 'SLG')" />
        </template>
      </Column>
      <Column header="OPS" sortable sort-field="_b.ops" style="min-width: 72px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatBA(r._b?.ops)" :class-map="bRateSeasonClass(r, 'ops')" :tooltip="bRateSeasonTip(r, 'ops', 'OPS')" />
        </template>
      </Column>
      <Column header="OPS+" sortable sort-field="_b.opsPlus" style="min-width: 68px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatAdjustedStat(r._b?.opsPlus)" :class-map="bRateSeasonClass(r, 'opsPlus')" :tooltip="bRateSeasonTip(r, 'opsPlus', 'OPS+')" />
        </template>
      </Column>
      <Column header="smbWAR" sortable sort-field="_b.smbWar" style="min-width: 80px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatWAR(r._b?.smbWar)" :class-map="bRateSeasonClass(r, 'smbWar')" :tooltip="bRateSeasonTip(r, 'smbWar', 'smbWAR')" />
        </template>
      </Column>
    </DataTable>

    <!-- Pitching mode -->
    <DataTable
      v-else
      :value="data"
      sort-field="seasonNum"
      :sort-order="-1"
      size="small"
      removable-sort
      scrollable
    >
      <ColumnGroup type="footer">
        <Row v-if="rsPitchingSummary">
          <Column :colspan="batchingPrefixCols" footer="Regular Season" footer-class="summary-label" />
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.g" :class-map="rsFooterClass('games', 'pitching')" :tooltip="rsFooterTip('games', 'G', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.gs" :class-map="rsFooterClass('gamesStarted', 'pitching')" :tooltip="rsFooterTip('gamesStarted', 'GS', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.w" :class-map="rsFooterClass('wins', 'pitching')" :tooltip="rsFooterTip('wins', 'W', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.l" :class-map="rsFooterClass('losses', 'pitching')" :tooltip="rsFooterTip('losses', 'L', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.sv" :class-map="rsFooterClass('saves', 'pitching')" :tooltip="rsFooterTip('saves', 'SV', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatIP(rsPitchingSummary.outsPitched)" :class-map="rsFooterClass('outsPitched', 'pitching')" :tooltip="rsFooterTip('outsPitched', 'IP', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.h" :class-map="rsFooterClass('hitsAllowed', 'pitching')" :tooltip="rsFooterTip('hitsAllowed', 'H', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.er" :class-map="rsFooterClass('earnedRuns', 'pitching')" :tooltip="rsFooterTip('earnedRuns', 'ER', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.bb" :class-map="rsFooterClass('walks', 'pitching')" :tooltip="rsFooterTip('walks', 'BB', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="rsPitchingSummary.k" :class-map="rsFooterClass('strikeouts', 'pitching')" :tooltip="rsFooterTip('strikeouts', 'K', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatERA(rsPitchingSummary.era)" :class-map="rsFooterClass('era', 'pitching')" :tooltip="rsFooterTip('era', 'ERA', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatWHIP(rsPitchingSummary.whip)" :class-map="rsFooterClass('whip', 'pitching')" :tooltip="rsFooterTip('whip', 'WHIP', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatK9(rsPitchingSummary.k9)" :class-map="rsFooterClass('k9', 'pitching')" :tooltip="rsFooterTip('k9', 'K/9', 'pitching')" /></template>
          </Column>
          <Column footer="—" footer-class="summary-cell" />
          <Column footer="—" footer-class="summary-cell" />
          <Column footer="—" footer-class="summary-cell" />
          <Column footer-class="summary-cell">
            <template #footer><StatHighlightCell :value="formatWAR(rsPitchingSummary.war)" :class-map="rsFooterClass('smbWar', 'pitching')" :tooltip="rsFooterTip('smbWar', 'smbWAR', 'pitching')" /></template>
          </Column>
        </Row>
        <Row v-if="poPitchingSummary">
          <Column :colspan="batchingPrefixCols" footer="Playoffs" footer-class="summary-label summary-po" />
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.g" :class-map="poFooterClass('games', 'pitching')" :tooltip="poFooterTip('games', 'G', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.gs" :class-map="poFooterClass('gamesStarted', 'pitching')" :tooltip="poFooterTip('gamesStarted', 'GS', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.w" :class-map="poFooterClass('wins', 'pitching')" :tooltip="poFooterTip('wins', 'W', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.l" :class-map="poFooterClass('losses', 'pitching')" :tooltip="poFooterTip('losses', 'L', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.sv" :class-map="poFooterClass('saves', 'pitching')" :tooltip="poFooterTip('saves', 'SV', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatIP(poPitchingSummary.outsPitched)" :class-map="poFooterClass('outsPitched', 'pitching')" :tooltip="poFooterTip('outsPitched', 'IP', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.h" :class-map="poFooterClass('hitsAllowed', 'pitching')" :tooltip="poFooterTip('hitsAllowed', 'H', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.er" :class-map="poFooterClass('earnedRuns', 'pitching')" :tooltip="poFooterTip('earnedRuns', 'ER', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.bb" :class-map="poFooterClass('walks', 'pitching')" :tooltip="poFooterTip('walks', 'BB', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="poPitchingSummary.k" :class-map="poFooterClass('strikeouts', 'pitching')" :tooltip="poFooterTip('strikeouts', 'K', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatERA(poPitchingSummary.era)" :class-map="poFooterClass('era', 'pitching')" :tooltip="poFooterTip('era', 'ERA', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatWHIP(poPitchingSummary.whip)" :class-map="poFooterClass('whip', 'pitching')" :tooltip="poFooterTip('whip', 'WHIP', 'pitching')" /></template>
          </Column>
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatK9(poPitchingSummary.k9)" :class-map="poFooterClass('k9', 'pitching')" :tooltip="poFooterTip('k9', 'K/9', 'pitching')" /></template>
          </Column>
          <Column footer="—" footer-class="summary-cell summary-po" />
          <Column footer="—" footer-class="summary-cell summary-po" />
          <Column footer="—" footer-class="summary-cell summary-po" />
          <Column footer-class="summary-cell summary-po">
            <template #footer><StatHighlightCell :value="formatWAR(poPitchingSummary.war)" :class-map="poFooterClass('smbWar', 'pitching')" :tooltip="poFooterTip('smbWar', 'smbWAR', 'pitching')" /></template>
          </Column>
        </Row>
        <Row v-if="careerPitchingSummary">
          <Column :colspan="batchingPrefixCols" footer="Career" footer-class="summary-label summary-career" />
          <Column :footer="String(careerPitchingSummary.g)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerPitchingSummary.gs)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerPitchingSummary.w)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerPitchingSummary.l)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerPitchingSummary.sv)" footer-class="summary-cell summary-career" />
          <Column :footer="formatIP(careerPitchingSummary.outsPitched)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerPitchingSummary.h)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerPitchingSummary.er)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerPitchingSummary.bb)" footer-class="summary-cell summary-career" />
          <Column :footer="String(careerPitchingSummary.k)" footer-class="summary-cell summary-career" />
          <Column :footer="formatERA(careerPitchingSummary.era)" footer-class="summary-cell summary-career" />
          <Column :footer="formatWHIP(careerPitchingSummary.whip)" footer-class="summary-cell summary-career" />
          <Column :footer="formatK9(careerPitchingSummary.k9)" footer-class="summary-cell summary-career" />
          <Column footer="—" footer-class="summary-cell summary-career" />
          <Column footer="—" footer-class="summary-cell summary-career" />
          <Column footer="—" footer-class="summary-cell summary-career" />
          <Column :footer="formatWAR(careerPitchingSummary.war)" footer-class="summary-cell summary-career" />
        </Row>
      </ColumnGroup>

      <Column field="seasonNum" header="Season" sortable style="min-width: 80px" />
      <Column header="Team" sortable :sort-field="teamSortKey" style="min-width: 130px">
        <template #body="{ data: r }">
          <span class="team-cell">
            <template v-for="(t, i) in r.teams" :key="t.teamHistoryId">
              <span v-if="i" class="team-separator"> · </span>
              <AppLink :to="`/teams/${t.teamId}/seasons/${t.teamHistoryId}`">{{ t.teamName }}</AppLink>
            </template>
            <template v-if="!hasFinalTeam(r.teams)">
              <span v-if="r.teams.length > 0" class="team-separator"> · </span>
              <span class="fa-label">FA</span>
            </template>
          </span>
        </template>
      </Column>
      <Column field="age" header="Age" sortable style="min-width: 55px" />
      <Column header="Role" style="min-width: 70px">
        <template #body="{ data: r }">{{ r.pitcherRole || '—' }}</template>
      </Column>
      <Column header="Traits" style="min-width: 160px">
        <template #body="{ data: r }">
          <TraitList :traits="r.traits" />
        </template>
      </Column>
      <Column v-if="hasAwards" header="Awards" style="min-width: 200px">
        <template #body="{ data: r }">
          <AppLink
            v-if="awardsBySeason?.[String(r.seasonNum)]?.length"
            :to="`/awards?seasonId=${r.seasonId}&view=1`"
            class="award-cell"
          >
            <AwardBadge
              v-for="award in awardsBySeason![String(r.seasonNum)]"
              :key="award.id"
              :award="award"
              size="sm"
            />
          </AppLink>
          <span v-else class="no-traits">—</span>
        </template>
      </Column>
      <Column header="G" sortable sort-field="_p.games" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.games" />
        </template>
      </Column>
      <Column header="GS" sortable sort-field="_p.gamesStarted" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.gamesStarted" :class-map="pSeasonClass(r, 'gamesStarted')" :tooltip="pSeasonTip(r, 'gamesStarted', 'GS')" />
        </template>
      </Column>
      <Column header="W" sortable sort-field="_p.wins" style="min-width: 50px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.wins" :class-map="pSeasonClass(r, 'wins')" :tooltip="pSeasonTip(r, 'wins', 'W')" />
        </template>
      </Column>
      <Column header="L" sortable sort-field="_p.losses" style="min-width: 50px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.losses" :class-map="pSeasonClass(r, 'losses')" :tooltip="pSeasonTip(r, 'losses', 'L')" />
        </template>
      </Column>
      <Column header="SV" sortable sort-field="_p.saves" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.saves" :class-map="pSeasonClass(r, 'saves')" :tooltip="pSeasonTip(r, 'saves', 'SV')" />
        </template>
      </Column>
      <Column header="IP" sortable sort-field="_p.outsPitched" style="min-width: 68px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p != null ? formatIP(r._p.outsPitched) : null" :class-map="pSeasonClass(r, 'outsPitched')" :tooltip="pSeasonTip(r, 'outsPitched', 'IP')" />
        </template>
      </Column>
      <Column header="H" sortable sort-field="_p.hitsAllowed" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.hitsAllowed" :class-map="pSeasonClass(r, 'hitsAllowed')" :tooltip="pSeasonTip(r, 'hitsAllowed', 'H')" />
        </template>
      </Column>
      <Column header="ER" sortable sort-field="_p.earnedRuns" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.earnedRuns" :class-map="pSeasonClass(r, 'earnedRuns')" :tooltip="pSeasonTip(r, 'earnedRuns', 'ER')" />
        </template>
      </Column>
      <Column header="BB" sortable sort-field="_p.walks" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.walks" :class-map="pSeasonClass(r, 'walks')" :tooltip="pSeasonTip(r, 'walks', 'BB')" />
        </template>
      </Column>
      <Column header="K" sortable sort-field="_p.strikeouts" style="min-width: 55px">
        <template #body="{ data: r }">
          <StatHighlightCell :value="r._p?.strikeouts" :class-map="pSeasonClass(r, 'strikeouts')" :tooltip="pSeasonTip(r, 'strikeouts', 'K')" />
        </template>
      </Column>
      <Column header="ERA" sortable sort-field="_p.era" style="min-width: 68px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatERA(r._p?.era)" :class-map="pRateSeasonClass(r, 'era')" :tooltip="pRateSeasonTip(r, 'era', 'ERA')" />
        </template>
      </Column>
      <Column header="WHIP" sortable sort-field="_p.whip" style="min-width: 72px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatWHIP(r._p?.whip)" :class-map="pRateSeasonClass(r, 'whip')" :tooltip="pRateSeasonTip(r, 'whip', 'WHIP')" />
        </template>
      </Column>
      <Column header="K/9" sortable sort-field="_p.k9" style="min-width: 65px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatK9(r._p?.k9)" :class-map="pRateSeasonClass(r, 'k9')" :tooltip="pRateSeasonTip(r, 'k9', 'K/9')" />
        </template>
      </Column>
      <Column header="ERA+" sortable sort-field="_p.eraPlus" style="min-width: 68px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatAdjustedStat(r._p?.eraPlus)" :class-map="pRateSeasonClass(r, 'eraPlus')" :tooltip="pRateSeasonTip(r, 'eraPlus', 'ERA+')" />
        </template>
      </Column>
      <Column header="FIP" sortable sort-field="_p.fip" style="min-width: 65px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatFIP(r._p?.fip)" :class-map="pRateSeasonClass(r, 'fip')" :tooltip="pRateSeasonTip(r, 'fip', 'FIP')" />
        </template>
      </Column>
      <Column header="FIP-" sortable sort-field="_p.fipMinus" style="min-width: 65px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatAdjustedStat(r._p?.fipMinus)" :class-map="pRateSeasonClass(r, 'fipMinus')" :tooltip="pRateSeasonTip(r, 'fipMinus', 'FIP-')" />
        </template>
      </Column>
      <Column header="smbWAR" sortable sort-field="_p.smbWar" style="min-width: 80px" class="col-rate">
        <template #body="{ data: r }">
          <StatHighlightCell :value="formatWAR(r._p?.smbWar)" :class-map="pRateSeasonClass(r, 'smbWar')" :tooltip="pRateSeasonTip(r, 'smbWar', 'smbWAR')" />
        </template>
      </Column>
    </DataTable>

    <StatHighlightLegend v-if="rows.length > 0" :show-leader="!showPlayoffs" />
  </div>
</template>

<style scoped>
.stat-table-wrap {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow-x: auto;
  overflow-y: hidden;
}

.fa-label {
  color: var(--color-text-secondary);
  font-style: italic;
}

.team-cell {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.team-separator {
  color: var(--color-text-secondary);
  padding: 0 3px;
}

.secondary-pos {
  color: var(--color-text-secondary);
}


.award-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

:deep(.summary-label) {
  font-weight: 600;
  color: var(--color-text-primary);
  background: var(--color-surface-2, var(--p-datatable-footer-background));
  border-top: 2px solid var(--color-border);
}

:deep(.summary-cell) {
  font-weight: 600;
  background: var(--color-surface-2, var(--p-datatable-footer-background));
  border-top: 2px solid var(--color-border);
}

:deep(.summary-po) {
  font-weight: 400;
  color: var(--color-text-secondary);
  background: var(--color-surface-1, var(--p-datatable-footer-background));
  border-top: none;
  font-style: italic;
}

:deep(.summary-career) {
  font-weight: 700;
  color: var(--color-text-primary);
  background: var(--color-surface-2, var(--p-datatable-footer-background));
  border-top: 2px solid var(--color-border);
}

a.award-cell {
  color: inherit;
  text-decoration: none;
}

</style>
