<script lang="ts" setup>
import Column from 'primevue/column'
import type { DataTablePageEvent, DataTableSortEvent } from 'primevue/datatable'
import DataTable from 'primevue/datatable'
import type { main } from '../../wailsjs/go/models'
import {
  formatAdjustedStat,
  formatERA,
  formatFIP,
  formatIP,
  formatK9,
  formatWAR,
  formatWHIP,
} from '../composables/useStatFormatters'
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

const props = defineProps<{
  rows: main.PitchingLeaderRowDTO[]
  isCareer: boolean
  highlights?: main.StatHighlightsDTO | null
  // Server-side pagination props — only used when isCareer is false.
  totalRecords?: number
  first?: number
  sortField?: string
  sortOrder?: number
}>()

const emit = defineEmits<{
  sort: [event: DataTableSortEvent]
  page: [event: DataTablePageEvent]
}>()

function leaderClass(r: main.PitchingLeaderRowDTO, statKey: string): Record<string, boolean> {
  return {
    'stat-leader': isSeasonLeader(r.playerId, r.seasonNum, statKey, props.highlights, 'pitching'),
    'stat-record': isSingleSeasonRecord(r.playerId, r.seasonNum, statKey, props.highlights, 'pitching'),
  }
}

function careerClass(r: main.PitchingLeaderRowDTO, statKey: string): Record<string, boolean> {
  return {
    'stat-record': isCareerRecordRS(r.playerId, statKey, props.highlights, 'pitching'),
  }
}

function seasonTip(r: main.PitchingLeaderRowDTO, statKey: string, label: string): string {
  return highlightTooltip(r.playerId, r.seasonNum, statKey, label, props.highlights, 'pitching', 'season')
}

function careerTip(r: main.PitchingLeaderRowDTO, statKey: string, label: string): string {
  return highlightTooltip(r.playerId, r.seasonNum, statKey, label, props.highlights, 'pitching', 'careerRS')
}
</script>

<template>
  <div class="table-wrap">
    <EmptyState v-if="rows.length === 0 && !isCareer && totalRecords === 0" message="No results — try adjusting the filters" />
    <EmptyState v-else-if="rows.length === 0 && isCareer" message="No results — try adjusting the filters" />
    <template v-else>
      <DataTable
        :value="rows"
        :lazy="!isCareer"
        :total-records="isCareer ? undefined : totalRecords"
        :first="isCareer ? undefined : first"
        :sort-field="isCareer ? 'smbWar' : sortField"
        :sort-order="isCareer ? -1 : sortOrder"
        size="small"
        :removable-sort="isCareer"
        scrollable
        scroll-height="flex"
        :paginator="isCareer ? rows.length > 50 : true"
        :rows="50"
        @sort="!isCareer && emit('sort', $event)"
        @page="!isCareer && emit('page', $event)"
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
        <Column v-if="!isCareer" field="teamName" header="Team" sortable style="min-width: 120px" />
        <Column v-if="!isCareer" field="age" header="Age" sortable style="min-width: 55px" />
        <Column v-if="!isCareer" field="pitcherRole" header="Role" sortable style="min-width: 58px" />
        <Column v-if="!isCareer" field="throwHand" header="Hand" sortable style="min-width: 60px" />

        <!-- Stat columns -->
        <Column field="games" header="G" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.games" :class-map="isCareer ? careerClass(r, 'games') : {}" :tooltip="isCareer ? careerTip(r, 'games', 'G') : ''" />
          </template>
        </Column>
        <Column field="gamesStarted" header="GS" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.gamesStarted" :class-map="isCareer ? careerClass(r, 'gamesStarted') : leaderClass(r, 'gamesStarted')" :tooltip="isCareer ? careerTip(r, 'gamesStarted', 'GS') : seasonTip(r, 'gamesStarted', 'GS')" />
          </template>
        </Column>
        <Column field="wins" header="W" sortable style="min-width: 50px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.wins" :class-map="isCareer ? careerClass(r, 'wins') : leaderClass(r, 'wins')" :tooltip="isCareer ? careerTip(r, 'wins', 'W') : seasonTip(r, 'wins', 'W')" />
          </template>
        </Column>
        <Column field="losses" header="L" sortable style="min-width: 50px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.losses" :class-map="isCareer ? careerClass(r, 'losses') : leaderClass(r, 'losses')" :tooltip="isCareer ? careerTip(r, 'losses', 'L') : seasonTip(r, 'losses', 'L')" />
          </template>
        </Column>
        <Column field="saves" header="SV" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.saves" :class-map="isCareer ? careerClass(r, 'saves') : leaderClass(r, 'saves')" :tooltip="isCareer ? careerTip(r, 'saves', 'SV') : seasonTip(r, 'saves', 'SV')" />
          </template>
        </Column>
        <Column header="IP" sort-field="outsPitched" sortable style="min-width: 65px" class="col-rate">
          <template #body="{ data: r }">
            <StatHighlightCell :value="formatIP(r.outsPitched)" :class-map="isCareer ? careerClass(r, 'outsPitched') : leaderClass(r, 'outsPitched')" :tooltip="isCareer ? careerTip(r, 'outsPitched', 'IP') : seasonTip(r, 'outsPitched', 'IP')" />
          </template>
        </Column>
        <Column field="hitsAllowed" header="H" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.hitsAllowed" :class-map="isCareer ? careerClass(r, 'hitsAllowed') : leaderClass(r, 'hitsAllowed')" :tooltip="isCareer ? careerTip(r, 'hitsAllowed', 'H') : seasonTip(r, 'hitsAllowed', 'H')" />
          </template>
        </Column>
        <Column field="earnedRuns" header="ER" sortable style="min-width: 55px">
          <template #body="{ data: r }">
            <StatHighlightCell :value="r.earnedRuns" :class-map="isCareer ? careerClass(r, 'earnedRuns') : leaderClass(r, 'earnedRuns')" :tooltip="isCareer ? careerTip(r, 'earnedRuns', 'ER') : seasonTip(r, 'earnedRuns', 'ER')" />
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
        <Column header="ERA" sort-field="era" sortable style="min-width: 65px" class="col-rate">
          <template #body="{ data: r }">{{ formatERA(r.era) }}</template>
        </Column>
        <Column header="WHIP" sort-field="whip" sortable style="min-width: 70px" class="col-rate">
          <template #body="{ data: r }">{{ formatWHIP(r.whip) }}</template>
        </Column>
        <Column header="K/9" sort-field="k9" sortable style="min-width: 65px" class="col-rate">
          <template #body="{ data: r }">{{ formatK9(r.k9) }}</template>
        </Column>
        <Column header="BB/9" sort-field="bb9" sortable style="min-width: 65px" class="col-rate">
          <template #body="{ data: r }">{{ formatK9(r.bb9) }}</template>
        </Column>
        <Column header="K/BB" sort-field="kPerBb" sortable style="min-width: 65px" class="col-rate">
          <template #body="{ data: r }">{{ formatK9(r.kPerBb) }}</template>
        </Column>
        <Column header="ERA+" sort-field="eraPlus" sortable style="min-width: 68px" class="col-rate">
          <template #body="{ data: r }">{{ formatAdjustedStat(r.eraPlus) }}</template>
        </Column>
        <Column v-if="!isCareer" header="FIP" sort-field="fip" sortable style="min-width: 65px" class="col-rate">
          <template #body="{ data: r }">{{ formatFIP(r.fip) }}</template>
        </Column>
        <Column v-if="!isCareer" header="FIP-" sort-field="fipMinus" sortable style="min-width: 65px" class="col-rate">
          <template #body="{ data: r }">{{ formatAdjustedStat(r.fipMinus) }}</template>
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

</style>
