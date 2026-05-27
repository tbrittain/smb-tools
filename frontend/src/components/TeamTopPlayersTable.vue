<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import type { main } from '../../wailsjs/go/models'
import { getAwardIcon } from '../composables/useAwardIcons'
import { formatAdjustedStat, formatSeasonRanges, formatWAR } from '../composables/useStatFormatters'
import AppLink from './AppLink.vue'

defineProps<{ players: main.TeamTopPlayerDTO[] }>()
</script>

<template>
  <DataTable
    :value="players"
    sort-field="smbWarWithTeam"
    :sort-order="-1"
    size="small"
    removable-sort
  >
    <Column header="Player" style="min-width: 160px">
      <template #body="{ data }: { data: main.TeamTopPlayerDTO }">
        <AppLink :to="`/players/${data.playerId}`">
          {{ data.firstName }} {{ data.lastName }}
        </AppLink>
        <span v-if="data.isHallOfFamer" class="hof-badge" title="Hall of Famer"> HOF</span>
      </template>
    </Column>
    <Column field="numSeasons" header="Seasons" sortable style="min-width: 80px" />
    <Column header="Season(s)" style="min-width: 100px">
      <template #body="{ data }: { data: main.TeamTopPlayerDTO }">
        {{ formatSeasonRanges(data.seasonNums) }}
      </template>
    </Column>
    <Column field="position" header="Pos" sortable style="min-width: 60px" />
    <Column field="smbWarWithTeam" header="smbWAR w/ Team" sortable style="min-width: 120px">
      <template #body="{ data }: { data: main.TeamTopPlayerDTO }">
        {{ formatWAR(data.smbWarWithTeam) }}
      </template>
    </Column>
    <Column header="OPS+ / ERA+" style="min-width: 95px">
      <template #body="{ data }: { data: main.TeamTopPlayerDTO }">
        {{ formatAdjustedStat(data.isPitcher ? data.avgEraPlus : data.avgOpsPlus) }}
      </template>
    </Column>
    <Column header="Awards" style="min-width: 180px">
      <template #body="{ data }: { data: main.TeamTopPlayerDTO }">
        <span v-if="!data.awards || data.awards.length === 0" class="no-awards">—</span>
        <span v-else class="awards-list">
          <span
            v-for="(award, idx) in data.awards"
            :key="idx"
            class="award-badge"
            :title="award"
          >{{ getAwardIcon(award) || award }}</span>
        </span>
      </template>
    </Column>
  </DataTable>
</template>

<style scoped>
.hof-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin-left: 0.25rem;
}

.awards-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.award-badge {
  font-size: 0.75rem;
  background: var(--p-surface-100, #f3f4f6);
  border-radius: 3px;
  padding: 1px 4px;
  white-space: nowrap;
}

.no-awards {
  color: var(--color-text-secondary);
}
</style>
