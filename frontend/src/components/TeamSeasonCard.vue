<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'
import AppLink from './AppLink.vue'

defineProps<{
  season: main.TeamSeasonSummaryDTO
}>()

function fmtPct(v: number): string {
  return v.toFixed(3).replace(/^0/, '')
}
</script>

<template>
  <div class="team-season-card" :class="{ champion: season.isChampion }">
    <div class="card-main">
      <div class="name-row">
        <AppLink
          :to="`/teams/${season.historyId}/seasons/${season.historyId}`"
          class="team-name"
        >
          {{ season.teamName }}
        </AppLink>
        <span v-if="season.isChampion" class="champ-badge">★ Champion</span>
        <span v-else-if="season.playoffSeed" class="playoff-badge">Playoffs #{{ season.playoffSeed }}</span>
      </div>
      <span class="season-label">Season {{ season.seasonNum }}</span>
    </div>

    <div class="card-stats">
      <span class="record">{{ season.wins }}–{{ season.losses }}</span>
      <span class="pct">{{ fmtPct(season.winPct) }}</span>
    </div>

    <div v-if="season.playoffWins != null" class="card-playoff">
      <span class="playoff-record">
        Playoffs: {{ season.playoffWins }}–{{ season.playoffLosses }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.team-season-card {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.team-season-card.champion {
  border-color: #d29922;
  background: color-mix(in srgb, #d29922 5%, var(--color-surface-1));
}

.card-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.team-name {
  font-size: 0.9375rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.season-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.champ-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  color: #d29922;
  background: color-mix(in srgb, #d29922 15%, transparent);
  border: 1px solid color-mix(in srgb, #d29922 40%, transparent);
  border-radius: 4px;
  padding: 0 5px;
  white-space: nowrap;
}

.playoff-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--color-accent);
  background: color-mix(in srgb, var(--color-accent) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-accent) 30%, transparent);
  border-radius: 4px;
  padding: 0 5px;
  white-space: nowrap;
}

.card-stats {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.125rem;
}

.record {
  font-size: 0.9375rem;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--color-text-primary);
}

.pct {
  font-size: 0.75rem;
  font-family: var(--font-mono);
  color: var(--color-text-secondary);
}

.card-playoff {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.playoff-record { font-family: var(--font-mono); }
</style>
