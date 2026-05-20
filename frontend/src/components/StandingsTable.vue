<script lang="ts" setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { main } from '../../wailsjs/go/models'

const props = defineProps<{
  standings: main.TeamStandingDTO[]
  activeHistoryId?: number
}>()

// Group by conference then division
const grouped = computed(() => {
  const map = new Map<string, Map<string, main.TeamStandingDTO[]>>()
  for (const row of props.standings) {
    const conf = row.conferenceName || 'League'
    const div = row.divisionName || ''
    if (!map.has(conf)) map.set(conf, new Map())
    const confMap = map.get(conf) ?? new Map<string, main.TeamStandingDTO[]>()
    if (!confMap.has(div)) confMap.set(div, [])
    ;(confMap.get(div) ?? []).push(row)
    map.set(conf, confMap)
  }
  return map
})

function fmtPct(v: number): string {
  return v.toFixed(3).replace(/^0/, '')
}
</script>

<template>
  <div class="standings-wrap">
    <div
      v-for="[confName, divMap] in grouped"
      :key="confName"
      class="conference"
    >
      <h4 v-if="confName !== 'League'" class="conf-name">{{ confName }}</h4>
      <div v-for="[divName, rows] in divMap" :key="divName" class="division">
        <h5 v-if="divName" class="div-name">{{ divName }}</h5>
        <table class="standings-table">
          <thead>
            <tr>
              <th class="col-team">Team</th>
              <th class="col-num">W</th>
              <th class="col-num">L</th>
              <th class="col-num">PCT</th>
              <th class="col-num">GB</th>
              <th class="col-num">R</th>
              <th class="col-num">RA</th>
              <th class="col-num">DIFF</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in rows"
              :key="row.historyId"
              :class="{ active: row.historyId === activeHistoryId }"
            >
              <td class="col-team">
                <RouterLink :to="`/teams/${row.teamId}/seasons/${row.historyId}`" class="team-link">
                  {{ row.teamName }}
                </RouterLink>
                <span v-if="row.playoffSeed" class="playoff-badge">P{{ row.playoffSeed }}</span>
              </td>
              <td class="col-num">{{ row.wins }}</td>
              <td class="col-num">{{ row.losses }}</td>
              <td class="col-num mono">{{ fmtPct(row.winPct) }}</td>
              <td class="col-num">{{ row.gamesBack === 0 ? '—' : row.gamesBack.toFixed(1) }}</td>
              <td class="col-num">{{ row.runsFor }}</td>
              <td class="col-num">{{ row.runsAgainst }}</td>
              <td class="col-num" :class="row.runDiff >= 0 ? 'pos' : 'neg'">
                {{ row.runDiff > 0 ? '+' : '' }}{{ row.runDiff }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.standings-wrap {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.conference {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.conf-name {
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.div-name {
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
  margin-bottom: 0.25rem;
}

.standings-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.standings-table th {
  color: var(--color-text-secondary);
  font-weight: 500;
  font-size: 0.6875rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  text-align: left;
  padding: 0.25rem 0.5rem;
  border-bottom: 1px solid var(--color-border);
}

.standings-table td {
  padding: 0.375rem 0.5rem;
  border-bottom: 1px solid color-mix(in srgb, var(--color-border) 40%, transparent);
  color: var(--color-text-primary);
}

.standings-table tr.active td {
  background: color-mix(in srgb, var(--color-accent) 8%, transparent);
}

.col-team { min-width: 160px; }
.col-num { text-align: right; width: 3rem; }
.mono { font-family: var(--font-mono); }

.team-link {
  color: var(--color-text-primary);
  text-decoration: none;
}
.team-link:hover { color: var(--color-accent); }

.playoff-badge {
  margin-left: 0.375rem;
  font-size: 0.625rem;
  font-weight: 600;
  background: var(--color-accent);
  color: #fff;
  border-radius: 3px;
  padding: 0 3px;
  vertical-align: middle;
}

.pos { color: #3fb950; }
.neg { color: var(--color-error); }
</style>
