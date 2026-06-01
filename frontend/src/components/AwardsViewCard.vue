<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'
import AppLink from './AppLink.vue'

defineProps<{
  group: main.AwardGroupSummaryDTO
}>()

function fmtBA(v: number): string {
  return v ? v.toFixed(3).replace(/^0\./, '.') : '—'
}

function fmtERA(v: number): string {
  return v ? v.toFixed(2) : '—'
}

function fmtWAR(v: number | null | undefined): string {
  return v != null ? v.toFixed(1) : '—'
}
</script>

<template>
  <div class="award-card">
    <div class="card-header">{{ group.awardName }}</div>

    <div class="card-body">
      <div
        v-for="(w, idx) in group.winners"
        :key="w.playerSeasonId"
        class="winner-row"
        :class="{ 'winner-row--first': idx === 0 }"
      >
        <div class="row-identity">
          <AppLink :to="`/players/${w.playerId}`" class="player-name">
            {{ w.firstName }} {{ w.lastName }}
          </AppLink>
          <span class="team-name">{{ w.teamName }}</span>
        </div>
        <div class="row-stats">
          <template v-if="w.ba || w.hr || w.rbi">
            <span class="stat"><span class="stat-label">BA</span> {{ fmtBA(w.ba) }}</span>
            <span class="stat"><span class="stat-label">HR</span> {{ w.hr }}</span>
            <span class="stat"><span class="stat-label">RBI</span> {{ w.rbi }}</span>
          </template>
          <template v-else-if="w.era || w.wins || w.strikeouts">
            <span class="stat"><span class="stat-label">ERA</span> {{ fmtERA(w.era) }}</span>
            <span class="stat"><span class="stat-label">W</span> {{ w.wins }}</span>
            <span class="stat"><span class="stat-label">K</span> {{ w.strikeouts }}</span>
          </template>
          <span class="stat"><span class="stat-label">WAR</span> {{ fmtWAR(w.smbWar) }}</span>
        </div>
      </div>

      <div v-if="group.runnerUp" class="runner-up-row">
        <div class="runner-up-label">Runner-up</div>
        <div class="row-identity">
          <AppLink :to="`/players/${group.runnerUp.playerId}`" class="player-name">
            {{ group.runnerUp.firstName }} {{ group.runnerUp.lastName }}
          </AppLink>
          <span class="team-name">{{ group.runnerUp.teamName }}</span>
        </div>
        <div class="row-stats">
          <template v-if="group.runnerUp.ba || group.runnerUp.hr || group.runnerUp.rbi">
            <span class="stat"><span class="stat-label">BA</span> {{ fmtBA(group.runnerUp.ba) }}</span>
            <span class="stat"><span class="stat-label">HR</span> {{ group.runnerUp.hr }}</span>
            <span class="stat"><span class="stat-label">RBI</span> {{ group.runnerUp.rbi }}</span>
          </template>
          <template v-else-if="group.runnerUp.era || group.runnerUp.wins || group.runnerUp.strikeouts">
            <span class="stat"><span class="stat-label">ERA</span> {{ fmtERA(group.runnerUp.era) }}</span>
            <span class="stat"><span class="stat-label">W</span> {{ group.runnerUp.wins }}</span>
            <span class="stat"><span class="stat-label">K</span> {{ group.runnerUp.strikeouts }}</span>
          </template>
          <span class="stat"><span class="stat-label">WAR</span> {{ fmtWAR(group.runnerUp.smbWar) }}</span>
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

.player-name {
  font-size: 0.9rem;
  font-weight: 600;
}

.team-name {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
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
</style>
