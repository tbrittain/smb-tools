<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import type { main } from '../../wailsjs/go/models'
import { getAwardIcon, getAwardImportance } from '../composables/useAwardIcons'
import { formatAdjustedStat, formatSeasonRanges, formatWAR } from '../composables/useStatFormatters'
import AppLink from './AppLink.vue'
import AwardBadge from './AwardBadge.vue'
import HofBadge from './HofBadge.vue'

defineProps<{ players: main.TeamTopPlayerDTO[] }>()

interface GroupedAward {
  originalName: string
  name: string
  importance: number
  count: number
}

function groupAwards(awards: string[]): GroupedAward[] {
  const counts = new Map<string, number>()
  for (const name of awards) {
    counts.set(name, (counts.get(name) ?? 0) + 1)
  }
  const groups: GroupedAward[] = []
  for (const [originalName, count] of counts) {
    groups.push({ originalName, name: originalName, importance: getAwardImportance(originalName), count })
  }
  return groups.sort((a, b) => a.importance - b.importance)
}
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
        <HofBadge v-if="data.isHallOfFamer" />
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
    <Column header="Awards With Team" style="min-width: 200px">
      <template #body="{ data }: { data: main.TeamTopPlayerDTO }">
        <span v-if="!data.awards || data.awards.length === 0" class="no-awards">—</span>
        <span v-else class="awards-list">
          <AwardBadge
            v-for="group in groupAwards(data.awards)"
            :key="group.originalName"
            :award="group"
            :count="group.count"
            size="sm"
          />
        </span>
      </template>
    </Column>
  </DataTable>
</template>

<style scoped>
.awards-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.no-awards {
  color: var(--color-text-secondary);
}
</style>
