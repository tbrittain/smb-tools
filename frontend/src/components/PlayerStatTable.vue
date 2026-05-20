<script lang="ts" setup>
import type { ColDef } from 'ag-grid-community'
import { AgGridVue } from 'ag-grid-vue3'
import { computed } from 'vue'
import type { main } from '../../wailsjs/go/models'
import { formatBA, formatERA, formatIP, formatK9, formatWHIP } from '../composables/useStatFormatters'
import EmptyState from './EmptyState.vue'

const props = defineProps<{
  rows: main.PlayerSeasonLogDTO[]
  mode: 'batting' | 'pitching'
  showPlayoffs: boolean
}>()

const data = computed(() =>
  props.rows.map((r) => ({
    ...r,
    stats: props.showPlayoffs ? (r.playoffBatting ?? r.playoffPitching) : (r.batting ?? r.pitching),
    _batting: props.showPlayoffs ? r.playoffBatting : r.batting,
    _pitching: props.showPlayoffs ? r.playoffPitching : r.pitching,
  })),
)

const battingCols: ColDef[] = [
  { headerName: 'Season', field: 'seasonNum', width: 82, pinned: 'left', sort: 'asc' },
  { headerName: 'Team', field: 'teamName', minWidth: 130, pinned: 'left' },
  { headerName: 'Age', field: 'age', width: 62, type: 'numericColumn' },
  {
    headerName: 'G',
    valueGetter: (p) => p.data?._batting?.gamesPlayed ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'AB',
    valueGetter: (p) => p.data?._batting?.atBats ?? null,
    width: 70,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'H',
    valueGetter: (p) => p.data?._batting?.hits ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'HR',
    valueGetter: (p) => p.data?._batting?.homeRuns ?? null,
    width: 65,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'RBI',
    valueGetter: (p) => p.data?._batting?.rbi ?? null,
    width: 65,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'SB',
    valueGetter: (p) => p.data?._batting?.stolenBases ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'BB',
    valueGetter: (p) => p.data?._batting?.walks ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'K',
    valueGetter: (p) => p.data?._batting?.strikeouts ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'BA',
    valueGetter: (p) => p.data?._batting?.ba ?? null,
    width: 72,
    type: 'numericColumn',
    valueFormatter: (p) => formatBA(p.value as number | null),
  },
  {
    headerName: 'OBP',
    valueGetter: (p) => p.data?._batting?.obp ?? null,
    width: 72,
    type: 'numericColumn',
    valueFormatter: (p) => formatBA(p.value as number | null),
  },
  {
    headerName: 'SLG',
    valueGetter: (p) => p.data?._batting?.slg ?? null,
    width: 72,
    type: 'numericColumn',
    valueFormatter: (p) => formatBA(p.value as number | null),
  },
  {
    headerName: 'OPS',
    valueGetter: (p) => p.data?._batting?.ops ?? null,
    width: 80,
    type: 'numericColumn',
    valueFormatter: (p) => formatBA(p.value as number | null),
  },
]

const pitchingCols: ColDef[] = [
  { headerName: 'Season', field: 'seasonNum', width: 82, pinned: 'left', sort: 'asc' },
  { headerName: 'Team', field: 'teamName', minWidth: 130, pinned: 'left' },
  { headerName: 'Age', field: 'age', width: 62, type: 'numericColumn' },
  {
    headerName: 'G',
    valueGetter: (p) => p.data?._pitching?.games ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'GS',
    valueGetter: (p) => p.data?._pitching?.gamesStarted ?? null,
    width: 65,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'W',
    valueGetter: (p) => p.data?._pitching?.wins ?? null,
    width: 55,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'L',
    valueGetter: (p) => p.data?._pitching?.losses ?? null,
    width: 55,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'SV',
    valueGetter: (p) => p.data?._pitching?.saves ?? null,
    width: 60,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'IP',
    valueGetter: (p) => p.data?._pitching?.outsPitched ?? null,
    width: 75,
    type: 'numericColumn',
    valueFormatter: (p) => (p.value != null ? formatIP(p.value as number) : '—'),
  },
  {
    headerName: 'H',
    valueGetter: (p) => p.data?._pitching?.hitsAllowed ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'ER',
    valueGetter: (p) => p.data?._pitching?.earnedRuns ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'BB',
    valueGetter: (p) => p.data?._pitching?.walks ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'K',
    valueGetter: (p) => p.data?._pitching?.strikeouts ?? null,
    width: 62,
    type: 'numericColumn',
    valueFormatter: (p) => p.value ?? '—',
  },
  {
    headerName: 'ERA',
    valueGetter: (p) => p.data?._pitching?.era ?? null,
    width: 75,
    type: 'numericColumn',
    valueFormatter: (p) => formatERA(p.value as number | null),
  },
  {
    headerName: 'WHIP',
    valueGetter: (p) => p.data?._pitching?.whip ?? null,
    width: 78,
    type: 'numericColumn',
    valueFormatter: (p) => formatWHIP(p.value as number | null),
  },
  {
    headerName: 'K/9',
    valueGetter: (p) => p.data?._pitching?.k9 ?? null,
    width: 72,
    type: 'numericColumn',
    valueFormatter: (p) => formatK9(p.value as number | null),
  },
]

const defaultColDef: ColDef = { sortable: true, resizable: true }
</script>

<template>
  <div class="stat-table-wrap">
    <EmptyState v-if="rows.length === 0" message="No season data" />
    <div v-else class="ag-theme-alpine-dark" style="height: 100%; width: 100%">
      <AgGridVue
        :row-data="data"
        :column-defs="mode === 'batting' ? battingCols : pitchingCols"
        :default-col-def="defaultColDef"
        :animate-rows="false"
        :suppress-cell-focus="true"
        :dom-layout="'autoHeight'"
        row-height="34"
        header-height="34"
        style="width: 100%"
      />
    </div>
  </div>
</template>

<style scoped>
.stat-table-wrap {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}
</style>
