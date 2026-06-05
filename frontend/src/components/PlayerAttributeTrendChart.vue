<script lang="ts" setup>
import { BarChart, LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, MarkLineComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { computed, ref } from 'vue'
import VChart from 'vue-echarts'
import type { main } from '../../wailsjs/go/models'

use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent, MarkLineComponent])

const props = defineProps<{
  seasons: main.PlayerAttributeSeasonDTO[]
  isPitcher: boolean
}>()

type Mode = 'raw' | 'percentile'
type ComparisonMode = 'role' | 'league'

const mode = ref<Mode>('raw')
const comparisonMode = ref<ComparisonMode>('role')

// AttrDef carries both league-wide and role-specific keys so the chart can
// switch between them based on the comparison toggle.
interface AttrDef {
  key: keyof main.PlayerAttributeSeasonDTO
  // Percentile keys: pctLeagueKey = league-wide, pctRoleKey = own role group.
  // For role-exclusive stats (arm, velocity, junk, accuracy) both point to the
  // same field because no meaningful league-wide comparison exists.
  pctLeagueKey: keyof main.PlayerAttributeSeasonDTO
  pctRoleKey: keyof main.PlayerAttributeSeasonDTO
  // Average keys: lgAvgKey for league-wide overlay, roleAvgKey for role overlay.
  lgAvgKey: keyof main.PlayerAttributeSeasonDTO
  roleAvgKey: keyof main.PlayerAttributeSeasonDTO
  label: string
  color: string
}

const BATTER_ATTRS: AttrDef[] = [
  {
    key: 'power',
    pctLeagueKey: 'powerPct',
    pctRoleKey: 'powerPctRole',
    lgAvgKey: 'lgAvgPower',
    roleAvgKey: 'roleAvgPower',
    label: 'Power',
    color: '#e06c75',
  },
  {
    key: 'contact',
    pctLeagueKey: 'contactPct',
    pctRoleKey: 'contactPctRole',
    lgAvgKey: 'lgAvgContact',
    roleAvgKey: 'roleAvgContact',
    label: 'Contact',
    color: '#98c379',
  },
  {
    key: 'speed',
    pctLeagueKey: 'speedPct',
    pctRoleKey: 'speedPctRole',
    lgAvgKey: 'lgAvgSpeed',
    roleAvgKey: 'roleAvgSpeed',
    label: 'Speed',
    color: '#61afef',
  },
  {
    key: 'fielding',
    pctLeagueKey: 'fieldingPct',
    pctRoleKey: 'fieldingPctRole',
    lgAvgKey: 'lgAvgFielding',
    roleAvgKey: 'roleAvgFielding',
    label: 'Fielding',
    color: '#c678dd',
  },
  // Arm is batter-only; both pct keys point to armPct (role-specific after migration 0014).
  {
    key: 'arm',
    pctLeagueKey: 'armPct',
    pctRoleKey: 'armPct',
    lgAvgKey: 'lgAvgArm',
    roleAvgKey: 'roleAvgArm',
    label: 'Arm',
    color: '#e5c07b',
  },
]

const PITCHER_ATTRS: AttrDef[] = [
  // Pitching stats are pitcher-only; both pct keys point to the role-specific field.
  {
    key: 'velocity',
    pctLeagueKey: 'velocityPct',
    pctRoleKey: 'velocityPct',
    lgAvgKey: 'lgAvgVelocity',
    roleAvgKey: 'roleAvgVelocity',
    label: 'Velocity',
    color: '#56b6c2',
  },
  {
    key: 'junk',
    pctLeagueKey: 'junkPct',
    pctRoleKey: 'junkPct',
    lgAvgKey: 'lgAvgJunk',
    roleAvgKey: 'roleAvgJunk',
    label: 'Junk',
    color: '#d19a66',
  },
  {
    key: 'accuracy',
    pctLeagueKey: 'accuracyPct',
    pctRoleKey: 'accuracyPct',
    lgAvgKey: 'lgAvgAccuracy',
    roleAvgKey: 'roleAvgAccuracy',
    label: 'Accuracy',
    color: '#abb2bf',
  },
]

const activeAttrs = computed<AttrDef[]>(() =>
  props.isPitcher ? [...BATTER_ATTRS.filter((a) => a.key !== 'arm'), ...PITCHER_ATTRS] : BATTER_ATTRS,
)

const xAxisData = computed(() => props.seasons.map((s) => `S${s.seasonNum}`))

const chartOption = computed(() => {
  const isPercentile = mode.value === 'percentile'
  const isRole = comparisonMode.value === 'role'

  const series: object[] = []

  for (const attr of activeAttrs.value) {
    // Player line
    const playerData = props.seasons.map((s) => {
      if (isPercentile) {
        const pctKey = isRole ? attr.pctRoleKey : attr.pctLeagueKey
        const pct = s[pctKey] as number | undefined
        return pct != null ? Math.round(pct) : null
      }
      return s[attr.key] as number
    })

    series.push({
      name: attr.label,
      type: 'line',
      data: playerData,
      smooth: false,
      symbol: 'circle',
      symbolSize: 6,
      lineStyle: { color: attr.color, width: 2 },
      itemStyle: { color: attr.color },
      connectNulls: false,
    })

    // Average overlay — dashed, no symbols, same color but lighter.
    // In raw mode: shows the league-wide average.
    // In percentile mode: shows 50 (any average group is at the 50th percentile
    // of that group by definition), but switches key so the tooltip context changes.
    const avgKey = isPercentile && isRole ? attr.roleAvgKey : attr.lgAvgKey
    const lgData = props.seasons.map((s) => {
      const raw = s[avgKey] as number
      if (raw === 0) return null
      if (isPercentile) return 50
      return Math.round(raw * 10) / 10
    })

    series.push({
      name: `Avg ${attr.label}`,
      type: 'line',
      data: lgData,
      smooth: false,
      symbol: 'none',
      lineStyle: { color: attr.color, width: 1, type: 'dashed', opacity: 0.45 },
      itemStyle: { color: attr.color },
      legendHoverLink: false,
      tooltip: { show: false },
    })
  }

  return {
    backgroundColor: 'transparent',
    grid: { left: 48, right: 24, top: 40, bottom: 64, containLabel: false },
    xAxis: {
      type: 'category',
      data: xAxisData.value,
      axisLabel: { color: '#abb2bf', fontSize: 11 },
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.12)' } },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: isPercentile ? 100 : 99,
      axisLabel: {
        color: '#abb2bf',
        fontSize: 11,
        formatter: isPercentile ? '{value}th' : '{value}',
      },
      axisLine: { show: false },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.08)' } },
    },
    legend: {
      data: activeAttrs.value.map((a) => a.label),
      bottom: 8,
      textStyle: { color: '#abb2bf', fontSize: 11 },
      icon: 'circle',
      itemWidth: 10,
      itemHeight: 10,
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1e2030',
      borderColor: 'rgba(255,255,255,0.12)',
      textStyle: { color: '#cdd6f4', fontSize: 12 },
      formatter(params: object[]) {
        const items = (params as { seriesName: string; value: number | null; color: string }[]).filter(
          (p) => !p.seriesName.startsWith('Avg ') && p.value != null,
        )
        if (items.length === 0) return ''
        const season =
          props.seasons[xAxisData.value.indexOf((params as { axisValueLabel: string }[])[0]?.axisValueLabel ?? '')]
        const label = isPercentile ? 'Percentile' : 'Rating'
        const avgLabel = isPercentile && isRole ? 'role avg' : 'lg avg'
        let html = `<div style="font-weight:600;margin-bottom:4px">Season ${season?.seasonNum ?? ''}</div>`
        for (const item of items) {
          const lgAttr = activeAttrs.value.find((a) => a.label === item.seriesName)
          const avgKey = isPercentile && isRole ? lgAttr?.roleAvgKey : lgAttr?.lgAvgKey
          const avgVal = avgKey && season ? (season[avgKey] as number) : null
          const avgStr =
            avgVal && avgVal > 0
              ? isPercentile
                ? ` <span style="opacity:0.55">(${avgLabel}: ${avgVal.toFixed(1)})</span>`
                : ` <span style="opacity:0.55">(${avgLabel}: ${avgVal.toFixed(1)})</span>`
              : ''
          html += `<div><span style="color:${item.color}">●</span> ${item.seriesName}: <strong>${item.value}</strong>${avgStr}</div>`
        }
        return html
      },
    },
    series,
  }
})
</script>

<template>
  <div class="attr-trend-chart">
    <div class="chart-controls">
      <div class="toggle-group">
        <button class="mode-btn" :class="{ active: mode === 'raw' }" @click="mode = 'raw'">Raw (1–99)</button>
        <button class="mode-btn" :class="{ active: mode === 'percentile' }" @click="mode = 'percentile'">
          Percentile
        </button>
      </div>
      <div v-if="mode === 'percentile'" class="toggle-group">
        <button
          class="mode-btn"
          :class="{ active: comparisonMode === 'role' }"
          @click="comparisonMode = 'role'"
        >
          vs. My Role
        </button>
        <button
          class="mode-btn"
          :class="{ active: comparisonMode === 'league' }"
          @click="comparisonMode = 'league'"
        >
          vs. League
        </button>
      </div>
    </div>
    <VChart class="chart" :option="chartOption" autoresize />
  </div>
</template>

<style scoped>
.attr-trend-chart {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.chart-controls {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.toggle-group {
  display: flex;
  gap: 0.25rem;
}

.mode-btn {
  padding: 0.25rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: background 0.1s, color 0.1s, border-color 0.1s;
}

.mode-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.mode-btn.active {
  background: var(--color-surface-2);
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.chart {
  height: 320px;
  width: 100%;
}
</style>
