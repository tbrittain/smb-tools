<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
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
import EmptyState from './EmptyState.vue'

const props = defineProps<{
  rows: main.PlayerSeasonLogDTO[]
  mode: 'batting' | 'pitching'
  showPlayoffs: boolean
}>()

// Flatten the selected stat block into each row for easy field access
const data = computed(() =>
  props.rows.map((r) => {
    const b = props.showPlayoffs ? r.playoffBatting : r.batting
    const p = props.showPlayoffs ? r.playoffPitching : r.pitching
    return { ...r, _b: b, _p: p }
  }),
)
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
    >
      <Column field="seasonNum" header="Season" sortable style="width: 80px" />
      <Column header="Team" sortable sort-field="teams[0].teamName" style="min-width: 130px">
        <template #body="{ data: r }">
          <RouterLink v-if="r.teams.length > 0" :to="`/teams/${r.teams[0].teamId}/seasons/${r.teams[0].teamHistoryId}`" class="team-link">
            {{ r.teams[0].teamName }}
          </RouterLink>
          <span v-else class="fa-label">FA</span>
        </template>
      </Column>
      <Column field="age" header="Age" sortable style="width: 55px" />
      <Column header="Pos" style="min-width: 70px">
        <template #body="{ data: r }">
          {{ r.primaryPosition }}<span v-if="r.secondaryPosition" class="secondary-pos">/{{ r.secondaryPosition }}</span>
        </template>
      </Column>
      <Column header="Traits" style="min-width: 160px">
        <template #body="{ data: r }">
          <span v-if="r.traits.length > 0" class="traits">{{ r.traits.join(', ') }}</span>
          <span v-else class="no-traits">—</span>
        </template>
      </Column>
      <Column header="G" sortable sort-field="_b.gamesPlayed" style="width: 55px">
        <template #body="{ data: r }">{{ r._b?.gamesPlayed ?? '—' }}</template>
      </Column>
      <Column header="AB" sortable sort-field="_b.atBats" style="width: 60px">
        <template #body="{ data: r }">{{ r._b?.atBats ?? '—' }}</template>
      </Column>
      <Column header="H" sortable sort-field="_b.hits" style="width: 55px">
        <template #body="{ data: r }">{{ r._b?.hits ?? '—' }}</template>
      </Column>
      <Column header="HR" sortable sort-field="_b.homeRuns" style="width: 55px">
        <template #body="{ data: r }">{{ r._b?.homeRuns ?? '—' }}</template>
      </Column>
      <Column header="RBI" sortable sort-field="_b.rbi" style="width: 60px">
        <template #body="{ data: r }">{{ r._b?.rbi ?? '—' }}</template>
      </Column>
      <Column header="SB" sortable sort-field="_b.stolenBases" style="width: 55px">
        <template #body="{ data: r }">{{ r._b?.stolenBases ?? '—' }}</template>
      </Column>
      <Column header="BB" sortable sort-field="_b.walks" style="width: 55px">
        <template #body="{ data: r }">{{ r._b?.walks ?? '—' }}</template>
      </Column>
      <Column header="K" sortable sort-field="_b.strikeouts" style="width: 55px">
        <template #body="{ data: r }">{{ r._b?.strikeouts ?? '—' }}</template>
      </Column>
      <Column header="BA" sortable sort-field="_b.ba" style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatBA(r._b?.ba) }}</template>
      </Column>
      <Column header="OBP" sortable sort-field="_b.obp" style="width: 68px" class="col-rate">
        <template #body="{ data: r }">{{ formatBA(r._b?.obp) }}</template>
      </Column>
      <Column header="SLG" sortable sort-field="_b.slg" style="width: 68px" class="col-rate">
        <template #body="{ data: r }">{{ formatBA(r._b?.slg) }}</template>
      </Column>
      <Column header="OPS" sortable sort-field="_b.ops" style="width: 72px" class="col-rate">
        <template #body="{ data: r }">{{ formatBA(r._b?.ops) }}</template>
      </Column>
      <Column header="OPS+" sortable sort-field="_b.opsPlus" style="width: 68px" class="col-rate">
        <template #body="{ data: r }">{{ formatAdjustedStat(r._b?.opsPlus) }}</template>
      </Column>
      <Column header="smbWAR" sortable sort-field="_b.smbWar" style="width: 80px" class="col-rate">
        <template #body="{ data: r }">{{ formatWAR(r._b?.smbWar) }}</template>
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
    >
      <Column field="seasonNum" header="Season" sortable style="width: 80px" />
      <Column header="Team" sortable sort-field="teams[0].teamName" style="min-width: 130px">
        <template #body="{ data: r }">
          <RouterLink v-if="r.teams.length > 0" :to="`/teams/${r.teams[0].teamId}/seasons/${r.teams[0].teamHistoryId}`" class="team-link">
            {{ r.teams[0].teamName }}
          </RouterLink>
          <span v-else class="fa-label">FA</span>
        </template>
      </Column>
      <Column field="age" header="Age" sortable style="width: 55px" />
      <Column header="Role" style="width: 70px">
        <template #body="{ data: r }">{{ r.pitcherRole || '—' }}</template>
      </Column>
      <Column header="Traits" style="min-width: 160px">
        <template #body="{ data: r }">
          <span v-if="r.traits.length > 0" class="traits">{{ r.traits.join(', ') }}</span>
          <span v-else class="no-traits">—</span>
        </template>
      </Column>
      <Column header="G" sortable sort-field="_p.games" style="width: 55px">
        <template #body="{ data: r }">{{ r._p?.games ?? '—' }}</template>
      </Column>
      <Column header="GS" sortable sort-field="_p.gamesStarted" style="width: 55px">
        <template #body="{ data: r }">{{ r._p?.gamesStarted ?? '—' }}</template>
      </Column>
      <Column header="W" sortable sort-field="_p.wins" style="width: 50px">
        <template #body="{ data: r }">{{ r._p?.wins ?? '—' }}</template>
      </Column>
      <Column header="L" sortable sort-field="_p.losses" style="width: 50px">
        <template #body="{ data: r }">{{ r._p?.losses ?? '—' }}</template>
      </Column>
      <Column header="SV" sortable sort-field="_p.saves" style="width: 55px">
        <template #body="{ data: r }">{{ r._p?.saves ?? '—' }}</template>
      </Column>
      <Column header="IP" sortable sort-field="_p.outsPitched" style="width: 68px">
        <template #body="{ data: r }">{{ r._p != null ? formatIP(r._p.outsPitched) : '—' }}</template>
      </Column>
      <Column header="H" sortable sort-field="_p.hitsAllowed" style="width: 55px">
        <template #body="{ data: r }">{{ r._p?.hitsAllowed ?? '—' }}</template>
      </Column>
      <Column header="ER" sortable sort-field="_p.earnedRuns" style="width: 55px">
        <template #body="{ data: r }">{{ r._p?.earnedRuns ?? '—' }}</template>
      </Column>
      <Column header="BB" sortable sort-field="_p.walks" style="width: 55px">
        <template #body="{ data: r }">{{ r._p?.walks ?? '—' }}</template>
      </Column>
      <Column header="K" sortable sort-field="_p.strikeouts" style="width: 55px">
        <template #body="{ data: r }">{{ r._p?.strikeouts ?? '—' }}</template>
      </Column>
      <Column header="ERA" sortable sort-field="_p.era" style="width: 68px" class="col-rate">
        <template #body="{ data: r }">{{ formatERA(r._p?.era) }}</template>
      </Column>
      <Column header="WHIP" sortable sort-field="_p.whip" style="width: 72px" class="col-rate">
        <template #body="{ data: r }">{{ formatWHIP(r._p?.whip) }}</template>
      </Column>
      <Column header="K/9" sortable sort-field="_p.k9" style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatK9(r._p?.k9) }}</template>
      </Column>
      <Column header="ERA+" sortable sort-field="_p.eraPlus" style="width: 68px" class="col-rate">
        <template #body="{ data: r }">{{ formatAdjustedStat(r._p?.eraPlus) }}</template>
      </Column>
      <Column header="FIP" sortable sort-field="_p.fip" style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatFIP(r._p?.fip) }}</template>
      </Column>
      <Column header="FIP-" sortable sort-field="_p.fipMinus" style="width: 65px" class="col-rate">
        <template #body="{ data: r }">{{ formatAdjustedStat(r._p?.fipMinus) }}</template>
      </Column>
      <Column header="smbWAR" sortable sort-field="_p.smbWar" style="width: 80px" class="col-rate">
        <template #body="{ data: r }">{{ formatWAR(r._p?.smbWar) }}</template>
      </Column>
    </DataTable>
  </div>
</template>

<style scoped>
.stat-table-wrap {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}

.team-link {
  color: var(--color-accent);
  text-decoration: none;
}

.team-link:hover {
  text-decoration: underline;
}

.fa-label {
  color: var(--color-text-secondary);
  font-style: italic;
}

.secondary-pos {
  color: var(--color-text-secondary);
}

.traits {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.no-traits {
  color: var(--color-text-secondary);
}
</style>
