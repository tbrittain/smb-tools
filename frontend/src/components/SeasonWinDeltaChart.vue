<script lang="ts" setup>
import type { CustomSeriesRenderItemAPI, CustomSeriesRenderItemParams } from 'echarts'
import { CustomChart, LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, MarkLineComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { computed, ref, watch } from 'vue'
import VChart from 'vue-echarts'
import { GetStandings, GetTeamSeasonScheduleByHistoryID } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import type { WinDeltaPoint } from '../composables/useWinDeltaSeries'
import { computeWinDeltaSeries } from '../composables/useWinDeltaSeries'
import LoadingSpinner from './LoadingSpinner.vue'

use([CanvasRenderer, LineChart, CustomChart, GridComponent, TooltipComponent, LegendComponent, MarkLineComponent])

const props = defineProps<{
  currentTeamHistoryId: number
  currentTeamName: string
  currentTeamDivisionName: string
  currentTeamSeasonId: number
  schedule: main.ScheduleGameDTO[]
}>()

const showDivision = ref(false)
const showMarginOfVictory = ref(false)
const loadingDivision = ref(false)
const divisionError = ref<string | null>(null)

// Map from historyId → { name, points }. Populated lazily on first division toggle.
const divisionData = ref<Map<number, { name: string; points: WinDeltaPoint[] }>>(new Map())
const divisionFetched = ref(false)

// Team colors used when division mode is active (current team always uses accent).
const DIVISION_COLORS = ['#e06c75', '#98c379', '#e5c07b', '#61afef', '#c678dd', '#56b6c2']

const currentTeamPoints = computed(() => computeWinDeltaSeries(props.schedule, props.currentTeamHistoryId))

watch(showDivision, async (enabled) => {
  if (!enabled || divisionFetched.value) return
  loadingDivision.value = true
  divisionError.value = null
  try {
    const standings = await GetStandings(props.currentTeamSeasonId)
    const peers = standings.filter(
      (s) => s.divisionName === props.currentTeamDivisionName && s.historyId !== props.currentTeamHistoryId,
    )
    const results = await Promise.all(
      peers.map(async (peer) => {
        const games = await GetTeamSeasonScheduleByHistoryID(peer.historyId)
        const points = computeWinDeltaSeries(games, peer.historyId)
        return { historyId: peer.historyId, name: peer.teamName, points }
      }),
    )
    const map = new Map<number, { name: string; points: WinDeltaPoint[] }>()
    for (const r of results) {
      map.set(r.historyId, { name: r.name, points: r.points })
    }
    divisionData.value = map
    divisionFetched.value = true
  } catch (e) {
    divisionError.value = String(e)
  } finally {
    loadingDivision.value = false
  }
})

function makeLineSeries(name: string, points: WinDeltaPoint[], color: string, isCurrentTeam: boolean): object {
  return {
    name,
    type: 'line',
    step: 'end',
    data: points.map((p) => [p.gameNum, p.delta]),
    symbol: 'circle',
    symbolSize: isCurrentTeam ? 5 : 4,
    lineStyle: { color, width: isCurrentTeam ? 2.5 : 1.5 },
    itemStyle: { color },
    markLine: isCurrentTeam
      ? {
          silent: true,
          symbol: 'none',
          data: [{ yAxis: 0 }],
          lineStyle: { color: 'rgba(255,255,255,0.35)', type: 'dashed', width: 1 },
          label: {
            show: true,
            position: 'insideEndTop',
            formatter: '.500',
            color: 'rgba(255,255,255,0.5)',
            fontSize: 11,
          },
        }
      : undefined,
  }
}

function makeErrorBarSeries(name: string, points: WinDeltaPoint[]): object {
  return {
    name: `${name}_bars`,
    type: 'custom',
    renderItem(params: CustomSeriesRenderItemParams, api: CustomSeriesRenderItemAPI) {
      const point = points[params.dataIndex]
      if (!point) return { type: 'group', children: [] }

      const baseX = api.coord([point.gameNum, point.delta])[0]
      const baseY = api.coord([point.gameNum, point.delta])[1]
      const tipY = api.coord([point.gameNum, point.delta + (point.won ? point.runDiff : -point.runDiff)])[1]
      const barColor = point.won ? '#4ade80' : '#f87171'

      return {
        type: 'rect',
        shape: { x: baseX - 1, y: Math.min(baseY, tipY), width: 2, height: Math.abs(baseY - tipY) },
        style: { fill: barColor, opacity: 0.7 },
      }
    },
    data: points.map((p) => [p.gameNum, p.delta]),
    z: 5,
    tooltip: { show: false },
    legendHoverLink: false,
  }
}

const allTeams = computed<{ historyId: number; name: string; points: WinDeltaPoint[]; color: string }[]>(() => {
  const teams: { historyId: number; name: string; points: WinDeltaPoint[]; color: string }[] = [
    {
      historyId: props.currentTeamHistoryId,
      name: props.currentTeamName,
      points: currentTeamPoints.value,
      color: '#ffffff',
    },
  ]
  if (showDivision.value && divisionFetched.value) {
    let colorIdx = 0
    for (const [histId, data] of divisionData.value) {
      teams.push({
        historyId: histId,
        name: data.name,
        points: data.points,
        color: DIVISION_COLORS[colorIdx % DIVISION_COLORS.length],
      })
      colorIdx++
    }
  }
  return teams
})

const chartOption = computed(() => {
  const series: object[] = []
  for (const team of allTeams.value) {
    const isCurrentTeam = team.historyId === props.currentTeamHistoryId
    series.push(makeLineSeries(team.name, team.points, team.color, isCurrentTeam))
    if (showMarginOfVictory.value) {
      series.push(makeErrorBarSeries(team.name, team.points))
    }
  }

  const showLegend = showDivision.value && allTeams.value.length > 1

  const maxGameNum = allTeams.value.reduce(
    (max, team) => Math.max(max, team.points.length > 0 ? team.points[team.points.length - 1].gameNum : 0),
    0,
  )

  return {
    backgroundColor: 'transparent',
    grid: { left: 48, right: 24, top: showLegend ? 48 : 24, bottom: showLegend ? 56 : 32, containLabel: false },
    xAxis: {
      type: 'value',
      name: 'Game',
      nameLocation: 'middle',
      nameGap: 20,
      nameTextStyle: { color: '#abb2bf', fontSize: 11 },
      axisLabel: { color: '#abb2bf', fontSize: 11 },
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.12)' } },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.05)' } },
      minInterval: 1,
      min: 1,
      max: maxGameNum > 0 ? maxGameNum : undefined,
    },
    yAxis: {
      type: 'value',
      name: 'Games above .500',
      nameLocation: 'middle',
      nameGap: 36,
      nameTextStyle: { color: '#abb2bf', fontSize: 11 },
      axisLabel: { color: '#abb2bf', fontSize: 11, formatter: (v: number) => (v > 0 ? `+${v}` : String(v)) },
      axisLine: { show: false },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.08)' } },
    },
    legend: showLegend
      ? {
          data: allTeams.value.map((t) => t.name),
          bottom: 4,
          textStyle: { color: '#abb2bf', fontSize: 11 },
          icon: 'circle',
          itemWidth: 10,
          itemHeight: 10,
        }
      : { show: false },
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1e2030',
      borderColor: 'rgba(255,255,255,0.12)',
      textStyle: { color: '#cdd6f4', fontSize: 12 },
      formatter(params: object[]) {
        const items = params as { seriesName: string; data: [number, number]; color: string }[]
        const lineItems = items.filter((p) => !p.seriesName.endsWith('_bars'))
        if (lineItems.length === 0) return ''
        const gameNum = lineItems[0].data[0]
        let html = `<div style="font-weight:600;margin-bottom:4px">Game ${gameNum}</div>`
        for (const item of lineItems) {
          const team = allTeams.value.find((t) => t.name === item.seriesName)
          const point = team?.points.find((p) => p.gameNum === gameNum)
          if (!point) continue
          const wl = point.won ? 'W' : 'L'
          const wlColor = point.won ? '#4ade80' : '#f87171'
          html += `<div style="display:flex;align-items:center;gap:6px">`
          html += `<span style="color:${item.color}">●</span>`
          html += `<span>${item.seriesName}</span>`
          html += `<span style="color:${wlColor};font-weight:600">${wl}</span>`
          html += `<span style="opacity:0.7">${point.myScore}–${point.oppScore} vs ${point.opponentName}</span>`
          html += `</div>`
        }
        return html
      },
    },
    series,
  }
})

const noPlayedGames = computed(() => currentTeamPoints.value.length === 0)
</script>

<template>
  <div class="win-delta-chart">
    <div class="chart-controls">
      <div class="toggle-group">
        <button class="mode-btn" :class="{ active: showDivision }" @click="showDivision = !showDivision">
          Division
        </button>
        <button
          class="mode-btn"
          :class="{ active: showMarginOfVictory }"
          @click="showMarginOfVictory = !showMarginOfVictory"
        >
          Margin of Victory
        </button>
      </div>
    </div>

    <div v-if="loadingDivision" class="chart-loading">
      <LoadingSpinner />
    </div>
    <p v-else-if="divisionError" class="error-text">{{ divisionError }}</p>
    <p v-else-if="noPlayedGames" class="empty-text">No games played yet.</p>
    <VChart v-else class="chart" :option="chartOption" autoresize />
  </div>
</template>

<style scoped>
.win-delta-chart {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.chart-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
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
  height: 360px;
  width: 100%;
}

.chart-loading {
  height: 360px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-text {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}
</style>
