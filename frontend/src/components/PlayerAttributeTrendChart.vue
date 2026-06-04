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
const mode = ref<Mode>('raw')

// Batter attributes always shown; pitcher-specific ones added for pitchers.
interface AttrDef {
  key: keyof main.PlayerAttributeSeasonDTO
  pctKey: keyof main.PlayerAttributeSeasonDTO
  lgKey: keyof main.PlayerAttributeSeasonDTO
  label: string
  color: string
}

const BATTER_ATTRS: AttrDef[] = [
  { key: 'power', pctKey: 'powerPct', lgKey: 'lgAvgPower', label: 'Power', color: '#e06c75' },
  { key: 'contact', pctKey: 'contactPct', lgKey: 'lgAvgContact', label: 'Contact', color: '#98c379' },
  { key: 'speed', pctKey: 'speedPct', lgKey: 'lgAvgSpeed', label: 'Speed', color: '#61afef' },
  { key: 'fielding', pctKey: 'fieldingPct', lgKey: 'lgAvgFielding', label: 'Fielding', color: '#c678dd' },
  { key: 'arm', pctKey: 'armPct', lgKey: 'lgAvgArm', label: 'Arm', color: '#e5c07b' },
]

const PITCHER_ATTRS: AttrDef[] = [
  { key: 'velocity', pctKey: 'velocityPct', lgKey: 'lgAvgVelocity', label: 'Velocity', color: '#56b6c2' },
  { key: 'junk', pctKey: 'junkPct', lgKey: 'lgAvgJunk', label: 'Junk', color: '#d19a66' },
  { key: 'accuracy', pctKey: 'accuracyPct', lgKey: 'lgAvgAccuracy', label: 'Accuracy', color: '#abb2bf' },
]

const activeAttrs = computed<AttrDef[]>(() => (props.isPitcher ? [...BATTER_ATTRS, ...PITCHER_ATTRS] : BATTER_ATTRS))

const xAxisData = computed(() => props.seasons.map((s) => `S${s.seasonNum}`))

const chartOption = computed(() => {
  const isPercentile = mode.value === 'percentile'

  const series: object[] = []

  for (const attr of activeAttrs.value) {
    // Player line
    const playerData = props.seasons.map((s) => {
      if (isPercentile) {
        const pct = s[attr.pctKey] as number | undefined
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

    // League average line — dashed, no symbols, same color but lighter
    const lgData = props.seasons.map((s) => {
      const raw = s[attr.lgKey] as number
      if (raw === 0) return null
      if (isPercentile) return 50 // league avg is always the 50th percentile by definition
      return Math.round(raw * 10) / 10
    })

    series.push({
      name: `Lg Avg ${attr.label}`,
      type: 'line',
      data: lgData,
      smooth: false,
      symbol: 'none',
      lineStyle: { color: attr.color, width: 1, type: 'dashed', opacity: 0.45 },
      itemStyle: { color: attr.color },
      // Hide from legend — it tracks with its paired player line
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
      // Show only player lines (not the league avg dashed lines)
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
        const items = (params as { seriesName: string; value: number | null; color: string }[])
          // Only show player lines (skip league avg lines)
          .filter((p) => !p.seriesName.startsWith('Lg Avg') && p.value != null)
        if (items.length === 0) return ''
        const season =
          props.seasons[xAxisData.value.indexOf((params as { axisValueLabel: string }[])[0]?.axisValueLabel ?? '')]
        const label = isPercentile ? 'Percentile' : 'Rating'
        let html = `<div style="font-weight:600;margin-bottom:4px">Season ${season?.seasonNum ?? ''}</div>`
        for (const item of items) {
          const lgAttr = activeAttrs.value.find((a) => a.label === item.seriesName)
          const lgVal = lgAttr && season ? (season[lgAttr.lgKey] as number) : null
          const lgStr =
            lgVal && lgVal > 0
              ? isPercentile
                ? ' <span style="opacity:0.55">(lg avg: 50th)</span>'
                : ` <span style="opacity:0.55">(lg avg: ${lgVal.toFixed(1)})</span>`
              : ''
          html += `<div><span style="color:${item.color}">●</span> ${item.seriesName}: <strong>${item.value}</strong>${lgStr}</div>`
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
      <button
        class="mode-btn"
        :class="{ active: mode === 'raw' }"
        @click="mode = 'raw'"
      >
        Raw (1–99)
      </button>
      <button
        class="mode-btn"
        :class="{ active: mode === 'percentile' }"
        @click="mode = 'percentile'"
      >
        Percentile
      </button>
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
