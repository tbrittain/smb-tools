<script lang="ts" setup>
import { computed } from 'vue'
import type { main } from '../../wailsjs/go/models'
import AppLink from './AppLink.vue'

const props = defineProps<{
  group: main.AwardGroupSummaryDTO
}>()

const isLargeGroup = computed(() => props.group.winners.length > 8)

const winnersByTeam = computed(() => {
  const map = new Map<string, main.AwardWinnerRowDTO[]>()
  for (const w of props.group.winners) {
    const key = w.teamName || 'Free Agent'
    const arr = map.get(key) ?? []
    arr.push(w)
    map.set(key, arr)
  }
  return [...map.entries()].map(([team, players]) => ({ team, players })).sort((a, b) => a.team.localeCompare(b.team))
})

function fmtBA(v: number): string {
  return v ? v.toFixed(3).replace(/^0\./, '.') : '—'
}

function fmtERA(v: number): string {
  return v ? v.toFixed(2) : '—'
}

function fmtWAR(v: number | null | undefined): string {
  return v != null ? v.toFixed(1) : '—'
}

function positionLabel(row: { pitcherRole: string; primaryPosition: string }): string {
  return row.pitcherRole || row.primaryPosition
}
</script>

<template>
  <div class="award-card">
    <div class="card-header">{{ group.awardName }}</div>

    <!-- Large-group layout (e.g. All-Star): 3-column grid grouped by team -->
    <div v-if="isLargeGroup" class="card-body card-body--large">
      <div v-for="teamGroup in winnersByTeam" :key="teamGroup.team" class="team-group">
        <div class="team-group-header">{{ teamGroup.team }}</div>
        <div class="team-group-players">
          <AppLink
            v-for="w in teamGroup.players"
            :key="w.playerSeasonId"
            :to="`/players/${w.playerId}`"
            class="large-player-row"
          >
            <span class="position-chip">{{ positionLabel(w) }}</span>
            <span class="large-player-name">{{ w.firstName }} {{ w.lastName }}</span>
          </AppLink>
        </div>
      </div>
    </div>

    <!-- Standard layout -->
    <div v-else class="card-body">
      <div
        v-for="(w, idx) in group.winners"
        :key="w.playerSeasonId"
        class="winner-row"
        :class="{ 'winner-row--first': idx === 0 }"
      >
        <div class="row-identity">
          <span class="position-chip">{{ positionLabel(w) }}</span>
          <AppLink :to="`/players/${w.playerId}`" class="player-name">
            {{ w.firstName }} {{ w.lastName }}
          </AppLink>
          <span v-if="w.teamName" class="team-name">{{ w.teamName }}</span>
          <span v-else class="team-name team-name--fa">FA</span>
        </div>
        <div class="row-stats">
          <template v-if="w.pitcherRole">
            <span class="stat"><span class="stat-label">ERA</span> {{ fmtERA(w.era) }}</span>
            <span class="stat"><span class="stat-label">W</span> {{ w.wins }}</span>
            <span class="stat"><span class="stat-label">K</span> {{ w.strikeouts }}</span>
          </template>
          <template v-else>
            <span class="stat"><span class="stat-label">BA</span> {{ fmtBA(w.ba) }}</span>
            <span class="stat"><span class="stat-label">HR</span> {{ w.hr }}</span>
            <span class="stat"><span class="stat-label">RBI</span> {{ w.rbi }}</span>
          </template>
          <span class="stat"><span class="stat-label">smbWAR</span> {{ fmtWAR(w.smbWar) }}</span>
        </div>
      </div>

      <div
        v-for="ru in group.runnerUps"
        :key="ru.playerSeasonId"
        class="runner-up-row"
      >
        <div class="runner-up-label">Runner-up ({{ ru.awardName }})</div>
        <div class="row-identity">
          <span class="position-chip">{{ positionLabel(ru) }}</span>
          <AppLink :to="`/players/${ru.playerId}`" class="player-name">
            {{ ru.firstName }} {{ ru.lastName }}
          </AppLink>
          <span v-if="ru.teamName" class="team-name">{{ ru.teamName }}</span>
          <span v-else class="team-name team-name--fa">FA</span>
        </div>
        <div class="row-stats">
          <template v-if="ru.pitcherRole">
            <span class="stat"><span class="stat-label">ERA</span> {{ fmtERA(ru.era) }}</span>
            <span class="stat"><span class="stat-label">W</span> {{ ru.wins }}</span>
            <span class="stat"><span class="stat-label">K</span> {{ ru.strikeouts }}</span>
          </template>
          <template v-else>
            <span class="stat"><span class="stat-label">BA</span> {{ fmtBA(ru.ba) }}</span>
            <span class="stat"><span class="stat-label">HR</span> {{ ru.hr }}</span>
            <span class="stat"><span class="stat-label">RBI</span> {{ ru.rbi }}</span>
          </template>
          <span class="stat"><span class="stat-label">smbWAR</span> {{ fmtWAR(ru.smbWar) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped>
.award-card {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}

.card-header {
  background: var(--color-surface-1);
  border-bottom: 1px solid var(--color-border);
  padding: 0.5rem 0.875rem;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-secondary);
}

.card-body {
  padding: 0.5rem 0.875rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.winner-row {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.row-identity {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.position-chip {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 3px;
  padding: 0.1em 0.35em;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.player-name {
  font-size: 0.9rem;
  font-weight: 600;
}

.team-name {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.team-name--fa {
  font-style: italic;
}

.row-stats {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.stat {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.stat-label {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--color-text-tertiary, var(--color-text-secondary));
  margin-right: 0.2rem;
}

.runner-up-row {
  border-top: 1px solid var(--color-border);
  padding-top: 0.4rem;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  opacity: 0.75;
}

.runner-up-label {
  font-size: 0.65rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-secondary);
}

/* ── Large-group layout (All-Star etc.) ─────────────────────────────────── */

.card-body--large {
  padding: 0.75rem 0.875rem;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.team-group-header {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-secondary);
  margin-bottom: 0.3rem;
}

.team-group-players {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.large-player-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  text-decoration: none;
  color: inherit;
}

.large-player-row:hover .large-player-name {
  text-decoration: underline;
}

.large-player-name {
  font-size: 0.85rem;
  font-weight: 500;
}
</style>
