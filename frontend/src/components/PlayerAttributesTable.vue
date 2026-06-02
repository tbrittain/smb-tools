<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import type { main } from '../../wailsjs/go/models'
import { formatSalary } from '../composables/useStatFormatters'
import AppLink from './AppLink.vue'
import EmptyState from './EmptyState.vue'

const props = defineProps<{
  rows: main.PlayerSeasonLogDTO[]
  isPitcher: boolean
}>()

function hasFinalTeam(r: main.PlayerSeasonLogDTO): boolean {
  return r.teams.some((t) => t.sortOrder === 0)
}

// TODO: The current data model only records where a player ended a season
// (sortOrder=0 = final team; no sortOrder=0 entry = ended as FA). There is no
// way to distinguish a player who was FA all season from one who signed
// mid-season, nor to detect a player who started the season as FA before
// signing — both produce the same teams[] shape. Fixing this would require
// storing FA transitions in the save game import pipeline.
</script>

<template>
  <div class="attr-table-wrap">
    <EmptyState v-if="rows.length === 0" message="No season data" />
    <DataTable
      v-else
      :value="props.rows"
      sort-field="seasonNum"
      :sort-order="-1"
      size="small"
      removable-sort
      scrollable
    >
      <Column field="seasonNum" header="Season" sortable style="min-width: 80px" />
      <Column header="Team" sortable style="min-width: 130px">
        <template #body="{ data: r }">
          <span class="team-cell">
            <template v-for="(t, i) in r.teams" :key="t.teamHistoryId">
              <span v-if="i" class="team-separator"> · </span>
              <AppLink :to="`/teams/${t.teamId}/seasons/${t.teamHistoryId}`">{{ t.teamName }}</AppLink>
            </template>
            <template v-if="!hasFinalTeam(r)">
              <span v-if="r.teams.length > 0" class="team-separator"> · </span>
              <span class="fa-label">FA</span>
            </template>
          </span>
        </template>
      </Column>
      <Column field="age" header="Age" sortable style="min-width: 55px" />
      <Column header="Salary" sortable sort-field="salary" style="min-width: 110px">
        <template #body="{ data: r }">{{ formatSalary(r.salary) }}</template>
      </Column>
      <Column field="power" header="POW" sortable style="min-width: 58px">
        <template #body="{ data: r }">{{ r.power > 0 ? r.power : '—' }}</template>
      </Column>
      <Column field="contact" header="CON" sortable style="min-width: 58px">
        <template #body="{ data: r }">{{ r.contact > 0 ? r.contact : '—' }}</template>
      </Column>
      <Column field="speed" header="SPD" sortable style="min-width: 58px">
        <template #body="{ data: r }">{{ r.speed > 0 ? r.speed : '—' }}</template>
      </Column>
      <Column field="fielding" header="FLD" sortable style="min-width: 58px">
        <template #body="{ data: r }">{{ r.fielding > 0 ? r.fielding : '—' }}</template>
      </Column>
      <Column field="arm" header="ARM" sortable style="min-width: 58px">
        <template #body="{ data: r }">{{ r.arm > 0 ? r.arm : '—' }}</template>
      </Column>
      <Column v-if="isPitcher" field="velocity" header="VEL" sortable style="min-width: 58px">
        <template #body="{ data: r }">{{ r.velocity > 0 ? r.velocity : '—' }}</template>
      </Column>
      <Column v-if="isPitcher" field="junk" header="JNK" sortable style="min-width: 58px">
        <template #body="{ data: r }">{{ r.junk > 0 ? r.junk : '—' }}</template>
      </Column>
      <Column v-if="isPitcher" field="accuracy" header="ACC" sortable style="min-width: 58px">
        <template #body="{ data: r }">{{ r.accuracy > 0 ? r.accuracy : '—' }}</template>
      </Column>
    </DataTable>
  </div>
</template>

<style scoped>
.attr-table-wrap {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow-x: auto;
  overflow-y: hidden;
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

.fa-label {
  color: var(--color-text-secondary);
  font-style: italic;
}
</style>
