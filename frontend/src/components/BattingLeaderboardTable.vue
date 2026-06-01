<script lang="ts" setup>
import Column from 'primevue/column'
import type { DataTablePageEvent, DataTableSortEvent } from 'primevue/datatable'
import DataTable from 'primevue/datatable'
import type { main } from '../../wailsjs/go/models'
import { formatAdjustedStat, formatBA, formatWAR } from '../composables/useStatFormatters'
import {
  highlightTooltip,
  isCareerRecordRS,
  isSeasonLeader,
  isSingleSeasonRecord,
} from '../composables/useStatHighlightHelpers'
import AppLink from './AppLink.vue'
import EmptyState from './EmptyState.vue'
import HofBadge from './HofBadge.vue'
import StatHighlightCell from './StatHighlightCell.vue'
import StatHighlightLegend from './StatHighlightLegend.vue'
import TraitList from './TraitList.vue'

const props = defineProps<{
  rows: main.BattingLeaderRowDTO[]
  isCareer: boolean
  highlights?: main.StatHighlightsDTO | null
  totalRecords?: number
  first?: number
  sortField?: string
  sortOrder?: number
}>()

const emit = defineEmits<{
  sort: [event: DataTableSortEvent]
  page: [event: DataTablePageEvent]
}>()

function leaderClass(r: main.BattingLeaderRowDTO, statKey: string): Record<string, boolean> {
  return {
    'stat-leader': isSeasonLeader(r.playerId, r.seasonNum, statKey, props.highlights, 'batting'),
    'stat-record': isSingleSeasonRecord(r.playerId, r.seasonNum, statKey, props.highlights, 'batting'),
  }
}

function careerClass(r: main.BattingLeaderRowDTO, statKey: string): Record<string, boolean> {
  return {
    'stat-record': isCareerRecordRS(r.playerId, statKey, props.highlights, 'batting'),
  }
}

function seasonTip(r: main.BattingLeaderRowDTO, statKey: string, label: string): string {
  return highlightTooltip(r.playerId, r.seasonNum, statKey, label, props.highlights, 'batting', 'season')
}

function careerTip(r: main.BattingLeaderRowDTO, statKey: string, label: string): string {
  return highlightTooltip(r.playerId, r.seasonNum, statKey, label, props.highlights, 'batting', 'careerRS')
}
</script>

<template>
  <div class="table-wrap">
    <EmptyState v-if="rows.length === 0 && (totalRecords ?? 0) === 0" message="No results — try adjusting the filters" />
    <template v-else>
      <DataTable
        :value="rows"
        lazy
        :total-records="totalRecords ?? 0"
        :first="first ?? 0"
        :sort-field="sortField ?? 'smbWar'"
        :sort-order="sortOrder ?? -1"
        size="small"
        scrollable
        scroll-height="flex"
        paginator
        :rows="50"
        @sort="emit('sort', $event)"
        @page="emit('page', $event)"
      >
        <Column header="Player" sort-field="lastName" sortable style="min-width: 160px">
          <template #body="{ data: r }">
            <AppLink :to="'/players/' + r.playerId">
              {{ r.firstName }} {{ r.lastName }}
            </AppLink>
            <HofBadge v-if="r.isHallOfFamer" />
          </template>
        </Column>

        <!-- Career identity columns -->
        <Column v-if="isCareer" field="seasonsPlayed" header="Seasons" sortable style="min-width: 75px" />

        <!-- Season identity columns -->
        <Column v-if="!isCareer" field="seasonNum" header="Season" sortable style="min-width: 72px" />
        <Column v-if="!isCareer" header="Team" sort-field="teamName" sortable style="min-width: 120px">
          <template #body="{ data: r }">
            <span v-if="r.teams && r.teams.length > 0" class="team-cell">
              <template v-for="(t, i) in r.teams" :key="t.teamHistoryId">
                <span v-if="i" class="team-separator"> · </span>
                <AppLink :to="`/teams/${t.teamId}/seasons/${t.teamHistoryId}`">{{ t.teamName }}</AppLink>
              </template>
            </span>
            <span v-else class="fa-label">FA</span>
          </template>
        </Column>
        <Column v-if="!isCareer" field="age" header="Age" sortable style="min-width: 55px" />
        <Column v-if="!isCareer" field="primaryPosition" header="Pos" sortable style="min-width: 55px" />
        <Column v-if="!isCareer" field="batHand" header="Hand" sortable style="min-width: 60px" />
        <Column v-if="!isCareer" header="Traits" style="min-width: 180px">
          <template #body="{ data: r }"><TraitList :traits="r.traits ?? []" /></template>
        </Column>

        <!-- Stat columns -->
        <Column field="gamesPlayed" header="G" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.gamesPlayed" :class-map="isCareer ? careerClass(r, 'gamesPlayed') : {}" :tooltip="isCareer ? careerTip(r, 'gamesPlayed', 'G') : ''" />
          </template>
        </Column>
        <Column field="atBats" header="AB" sortable style="min-width: 60px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.atBats" :class-map="isCareer ? careerClass(r, 'atBats') : leaderClass(r, 'atBats')" :tooltip="isCareer ? careerTip(r, 'atBats', 'AB') : seasonTip(r, 'atBats', 'AB')" />
          </template>
        </Column>
        <Column field="hits" header="H" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.hits" :class-map="isCareer ? careerClass(r, 'hits') : leaderClass(r, 'hits')" :tooltip="isCareer ? careerTip(r, 'hits', 'H') : seasonTip(r, 'hits', 'H')" />
          </template>
        </Column>
        <Column field="doubles" header="2B" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.doubles" :class-map="isCareer ? careerClass(r, 'doubles') : leaderClass(r, 'doubles')" :tooltip="isCareer ? careerTip(r, 'doubles', '2B') : seasonTip(r, 'doubles', '2B')" />
          </template>
        </Column>
        <Column field="triples" header="3B" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.triples" :class-map="isCareer ? careerClass(r, 'triples') : leaderClass(r, 'triples')" :tooltip="isCareer ? careerTip(r, 'triples', '3B') : seasonTip(r, 'triples', '3B')" />
          </template>
        </Column>
        <Column field="homeRuns" header="HR" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.homeRuns" :class-map="isCareer ? careerClass(r, 'homeRuns') : leaderClass(r, 'homeRuns')" :tooltip="isCareer ? careerTip(r, 'homeRuns', 'HR') : seasonTip(r, 'homeRuns', 'HR')" />
          </template>
        </Column>
        <Column field="rbi" header="RBI" sortable style="min-width: 58px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.rbi" :class-map="isCareer ? careerClass(r, 'rbi') : leaderClass(r, 'rbi')" :tooltip="isCareer ? careerTip(r, 'rbi', 'RBI') : seasonTip(r, 'rbi', 'RBI')" />
          </template>
        </Column>
        <Column field="stolenBases" header="SB" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.stolenBases" :class-map="isCareer ? careerClass(r, 'stolenBases') : leaderClass(r, 'stolenBases')" :tooltip="isCareer ? careerTip(r, 'stolenBases', 'SB') : seasonTip(r, 'stolenBases', 'SB')" />
          </template>
        </Column>
        <Column field="walks" header="BB" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.walks" :class-map="isCareer ? careerClass(r, 'walks') : leaderClass(r, 'walks')" :tooltip="isCareer ? careerTip(r, 'walks', 'BB') : seasonTip(r, 'walks', 'BB')" />
          </template>
        </Column>
        <Column field="strikeouts" header="K" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.strikeouts" :class-map="isCareer ? careerClass(r, 'strikeouts') : leaderClass(r, 'strikeouts')" :tooltip="isCareer ? careerTip(r, 'strikeouts', 'K') : seasonTip(r, 'strikeouts', 'K')" />
          </template>
        </Column>
        <Column header="BA" sort-field="ba" sortable style="min-width: 65px" class="col-rate">
          <template #body="{ data: r }">{{ formatBA(r.ba) }}</template>
        </Column>
        <Column header="OBP" sort-field="obp" sortable style="min-width: 68px" class="col-rate">
          <template #body="{ data: r }">{{ formatBA(r.obp) }}</template>
        </Column>
        <Column header="SLG" sort-field="slg" sortable style="min-width: 68px" class="col-rate">
          <template #body="{ data: r }">{{ formatBA(r.slg) }}</template>
        </Column>
        <Column header="OPS" sort-field="ops" sortable style="min-width: 72px" class="col-rate">
          <template #body="{ data: r }">{{ formatBA(r.ops) }}</template>
        </Column>
        <Column header="OPS+" sort-field="opsPlus" sortable style="min-width: 68px" class="col-rate">
          <template #body="{ data: r }">{{ formatAdjustedStat(r.opsPlus) }}</template>
        </Column>
        <Column header="smbWAR" sort-field="smbWar" sortable style="min-width: 80px" class="col-rate">
          <template #body="{ data: r }">{{ formatWAR(r.smbWar) }}</template>
        </Column>
      </DataTable>
      <StatHighlightLegend :show-leader="!isCareer" />
    </template>
  </div>
</template>

<style scoped>
.table-wrap {
  height: 100%;
  display: flex;
  flex-direction: column;
}

:deep(.col-rate) {
  font-variant-numeric: tabular-nums;
}

.team-cell {
  display: flex;
  align-items: center;
}

.team-separator {
  color: var(--color-text-secondary);
  padding: 0 3px;
}

.fa-label {
  color: var(--color-text-secondary);
  font-style: italic;
}
</style>
