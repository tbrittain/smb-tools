<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import { RouterLink } from 'vue-router'
import type { main } from '../../wailsjs/go/models'
import { formatERA, formatIP, formatK9, formatWHIP } from '../composables/useStatFormatters'
import EmptyState from './EmptyState.vue'

defineProps<{
  rows: main.PitchingLeaderRowDTO[]
  isCareer: boolean
}>()
</script>

<template>
  <div class="table-wrap">
    <EmptyState v-if="rows.length === 0" message="No results — try adjusting the filters" />
    <DataTable
      v-else
      :value="rows"
      sort-field="strikeouts"
      :sort-order="-1"
      size="small"
      removable-sort
      scrollable
      scroll-height="flex"
      :paginator="rows.length > 50"
      :rows="50"
    >
      <Column header="Player" sort-field="lastName" sortable style="min-width: 160px">
        <template #body="{ data: r }">
          <RouterLink :to="'/players/' + r.playerId" class="player-link">
            {{ r.firstName }} {{ r.lastName }}
          </RouterLink>
          <span v-if="r.isHallOfFamer" class="hof-badge">HoF</span>
        </template>
      </Column>

      <!-- Career identity columns -->
      <Column v-if="isCareer" field="seasonsPlayed" header="Seasons" sortable style="width: 75px" />

      <!-- Season identity columns -->
      <Column v-if="!isCareer" field="seasonNum" header="Season" sortable style="width: 72px" />
      <Column v-if="!isCareer" field="teamName" header="Team" sortable style="min-width: 120px" />
      <Column v-if="!isCareer" field="age" header="Age" sortable style="width: 55px" />
      <Column v-if="!isCareer" field="pitcherRole" header="Role" sortable style="width: 58px" />
      <Column v-if="!isCareer" field="throwHand" header="Hand" sortable style="width: 60px" />

      <!-- Stat columns (shared) -->
      <Column field="games" header="G" sortable style="width: 55px" />
      <Column field="gamesStarted" header="GS" sortable style="width: 55px" />
      <Column field="wins" header="W" sortable style="width: 50px" />
      <Column field="losses" header="L" sortable style="width: 50px" />
      <Column field="saves" header="SV" sortable style="width: 55px" />
      <Column header="IP" sort-field="outsPitched" sortable style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatIP(r.outsPitched) }}</template>
      </Column>
      <Column field="hitsAllowed" header="H" sortable style="width: 55px" />
      <Column field="earnedRuns" header="ER" sortable style="width: 55px" />
      <Column field="walks" header="BB" sortable style="width: 55px" />
      <Column field="strikeouts" header="K" sortable style="width: 55px" />
      <Column header="ERA" sort-field="era" sortable style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatERA(r.era) }}</template>
      </Column>
      <Column header="WHIP" sort-field="whip" sortable style="width: 70px" class="col-rate">
        <template #body="{ data: r }">{{ formatWHIP(r.whip) }}</template>
      </Column>
      <Column header="K/9" sort-field="k9" sortable style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatK9(r.k9) }}</template>
      </Column>
      <Column header="BB/9" sort-field="bb9" sortable style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatK9(r.bb9) }}</template>
      </Column>
      <Column header="K/BB" sort-field="kPerBb" sortable style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatK9(r.kPerBb) }}</template>
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

.player-link {
  color: var(--color-accent);
  text-decoration: none;
}
.player-link:hover {
  text-decoration: underline;
}
.hof-badge {
  margin-left: 0.375rem;
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--color-gold, #c9a227);
  border: 1px solid var(--color-gold, #c9a227);
  border-radius: 3px;
  padding: 0 3px;
  vertical-align: middle;
}
:deep(.col-rate) {
  font-variant-numeric: tabular-nums;
}
</style>
