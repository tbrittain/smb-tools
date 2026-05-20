<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'
import { formatBA, formatERA, formatIP, formatK9, formatWHIP } from '../composables/useStatFormatters'

defineProps<{
  batting: main.CareerBattingStatsDTO | null | undefined
  pitching: main.CareerPitchingStatsDTO | null | undefined
}>()
</script>

<template>
  <div class="career-summary">
    <div v-if="batting && batting.atBats > 0" class="stat-group">
      <h4 class="group-label">Career Batting</h4>
      <div class="stat-row">
        <div class="stat-cell">
          <span class="stat-label">G</span>
          <span class="stat-val">{{ batting.gamesPlayed }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">AB</span>
          <span class="stat-val">{{ batting.atBats }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">H</span>
          <span class="stat-val">{{ batting.hits }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">HR</span>
          <span class="stat-val">{{ batting.homeRuns }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">RBI</span>
          <span class="stat-val">{{ batting.rbi }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">SB</span>
          <span class="stat-val">{{ batting.stolenBases }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">BB</span>
          <span class="stat-val">{{ batting.walks }}</span>
        </div>
        <div class="stat-cell highlight">
          <span class="stat-label">BA</span>
          <span class="stat-val">{{ formatBA(batting.ba) }}</span>
        </div>
        <div class="stat-cell highlight">
          <span class="stat-label">OBP</span>
          <span class="stat-val">{{ formatBA(batting.obp) }}</span>
        </div>
        <div class="stat-cell highlight">
          <span class="stat-label">SLG</span>
          <span class="stat-val">{{ formatBA(batting.slg) }}</span>
        </div>
        <div class="stat-cell highlight">
          <span class="stat-label">OPS</span>
          <span class="stat-val">{{ formatBA(batting.ops) }}</span>
        </div>
      </div>
    </div>

    <div v-if="pitching && pitching.outsPitched > 0" class="stat-group">
      <h4 class="group-label">Career Pitching</h4>
      <div class="stat-row">
        <div class="stat-cell">
          <span class="stat-label">G</span>
          <span class="stat-val">{{ pitching.games }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">GS</span>
          <span class="stat-val">{{ pitching.gamesStarted }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">W</span>
          <span class="stat-val">{{ pitching.wins }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">L</span>
          <span class="stat-val">{{ pitching.losses }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">SV</span>
          <span class="stat-val">{{ pitching.saves }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">IP</span>
          <span class="stat-val">{{ formatIP(pitching.outsPitched) }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">K</span>
          <span class="stat-val">{{ pitching.strikeouts }}</span>
        </div>
        <div class="stat-cell">
          <span class="stat-label">BB</span>
          <span class="stat-val">{{ pitching.walks }}</span>
        </div>
        <div class="stat-cell highlight">
          <span class="stat-label">ERA</span>
          <span class="stat-val">{{ formatERA(pitching.era) }}</span>
        </div>
        <div class="stat-cell highlight">
          <span class="stat-label">WHIP</span>
          <span class="stat-val">{{ formatWHIP(pitching.whip) }}</span>
        </div>
        <div class="stat-cell highlight">
          <span class="stat-label">K/9</span>
          <span class="stat-val">{{ formatK9(pitching.k9) }}</span>
        </div>
      </div>
    </div>

    <p v-if="(!batting || batting.atBats === 0) && (!pitching || pitching.outsPitched === 0)" class="empty">
      No career stats
    </p>
  </div>
</template>

<style scoped>
.career-summary {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.stat-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.group-label {
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
  margin: 0;
}

.stat-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}

.stat-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.5rem 0.875rem;
  border-right: 1px solid var(--color-border);
  min-width: 52px;
}

.stat-cell:last-child { border-right: none; }

.stat-cell.highlight { background: color-mix(in srgb, var(--color-accent) 5%, transparent); }

.stat-label {
  font-size: 0.625rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.stat-val {
  font-size: 0.9375rem;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--color-text-primary);
}

.empty {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}
</style>
