<script lang="ts" setup>
import Column from 'primevue/column'
import type { DataTablePageEvent, DataTableSortEvent } from 'primevue/datatable'
import DataTable from 'primevue/datatable'
import type { main } from '../../wailsjs/go/models'
import { formatAdjustedStat, formatBA, formatWAR } from '../composables/useStatFormatters'
import AppLink from './AppLink.vue'
import EmptyState from './EmptyState.vue'
import HofBadge from './HofBadge.vue'

const props = defineProps<{
  rows: main.BattingLeaderRowDTO[]
  isCareer: boolean
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
</script>

<template>
  <div class="table-wrap">
    <EmptyState v-if="rows.length === 0 && !isCareer && totalRecords === 0" message="No results — try adjusting the filters" />
    <EmptyState v-else-if="rows.length === 0 && isCareer" message="No results — try adjusting the filters" />
    <DataTable
      v-else
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
      <Column v-if="!isCareer" field="primaryPosition" header="Pos" sortable style="min-width: 55px" />
      <Column v-if="!isCareer" field="batHand" header="Hand" sortable style="min-width: 60px" />

      <!-- Stat columns (shared) -->
      <Column field="gamesPlayed" header="G" sortable style="min-width: 55px" />
      <Column field="atBats" header="AB" sortable style="min-width: 60px" />
      <Column field="hits" header="H" sortable style="min-width: 55px" />
      <Column field="doubles" header="2B" sortable style="min-width: 55px" />
      <Column field="triples" header="3B" sortable style="min-width: 55px" />
      <Column field="homeRuns" header="HR" sortable style="min-width: 55px" />
      <Column field="rbi" header="RBI" sortable style="min-width: 58px" />
      <Column field="stolenBases" header="SB" sortable style="min-width: 55px" />
      <Column field="walks" header="BB" sortable style="min-width: 55px" />
      <Column field="strikeouts" header="K" sortable style="min-width: 55px" />
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
