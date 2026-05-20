<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import { RouterLink } from 'vue-router'
import type { main } from '../../wailsjs/go/models'
import { formatBA } from '../composables/useStatFormatters'
import EmptyState from './EmptyState.vue'

defineProps<{
  rows: main.BattingLeaderRowDTO[]
  isCareer: boolean
}>()
</script>

<template>
  <div class="table-wrap">
    <EmptyState v-if="rows.length === 0" message="No results — try adjusting the filters" />
    <DataTable
      v-else
      :value="rows"
      sort-field="hits"
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
      <Column v-if="!isCareer" field="primaryPosition" header="Pos" sortable style="width: 55px" />
      <Column v-if="!isCareer" field="batHand" header="Hand" sortable style="width: 60px" />

      <!-- Stat columns (shared) -->
      <Column field="gamesPlayed" header="G" sortable style="width: 55px" />
      <Column field="atBats" header="AB" sortable style="width: 60px" />
      <Column field="hits" header="H" sortable style="width: 55px" />
      <Column field="doubles" header="2B" sortable style="width: 55px" />
      <Column field="triples" header="3B" sortable style="width: 55px" />
      <Column field="homeRuns" header="HR" sortable style="width: 55px" />
      <Column field="rbi" header="RBI" sortable style="width: 58px" />
      <Column field="stolenBases" header="SB" sortable style="width: 55px" />
      <Column field="walks" header="BB" sortable style="width: 55px" />
      <Column field="strikeouts" header="K" sortable style="width: 55px" />
      <Column header="BA" sort-field="ba" sortable style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatBA(r.ba) }}</template>
      </Column>
      <Column header="OBP" sort-field="obp" sortable style="width: 68px" class="col-rate">
        <template #body="{ data: r }">{{ formatBA(r.obp) }}</template>
      </Column>
      <Column header="SLG" sort-field="slg" sortable style="width: 68px" class="col-rate">
        <template #body="{ data: r }">{{ formatBA(r.slg) }}</template>
      </Column>
      <Column header="OPS" sort-field="ops" sortable style="width: 72px" class="col-rate">
        <template #body="{ data: r }">{{ formatBA(r.ops) }}</template>
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
