<script lang="ts" setup>
import { RouterLink } from 'vue-router'
import type { main } from '../../wailsjs/go/models'
import LoadingSpinner from './LoadingSpinner.vue'

defineProps<{
  leaders: main.StatLeadersDTO | null
  loading: boolean
}>()

function formatBA(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(3).replace(/^0/, '')
}

function formatERA(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(2)
}

function formatInt(v: number | null | undefined): string {
  if (v == null) return '—'
  return String(Math.round(v))
}
</script>

<template>
  <div class="leaders-panel">
    <LoadingSpinner v-if="loading" />
    <template v-else-if="leaders">
      <div class="leader-grid">
        <div v-for="cat in [
          { key: 'ba',         label: 'Batting Avg',   leader: leaders.ba,         fmt: formatBA  },
          { key: 'hr',         label: 'Home Runs',     leader: leaders.hr,         fmt: formatInt },
          { key: 'rbi',        label: 'RBI',           leader: leaders.rbi,        fmt: formatInt },
          { key: 'era',        label: 'ERA',           leader: leaders.era,        fmt: formatERA },
          { key: 'wins',       label: 'Wins',          leader: leaders.wins,       fmt: formatInt },
          { key: 'strikeouts', label: 'Strikeouts',    leader: leaders.strikeouts, fmt: formatInt },
        ]" :key="cat.key" class="leader-tile">
          <span class="tile-label">{{ cat.label }}</span>
          <template v-if="cat.leader">
            <RouterLink
              :to="`/players/${cat.leader.playerId}`"
              class="tile-value"
            >
              {{ cat.fmt(cat.leader.statValue) }}
            </RouterLink>
            <span class="tile-player">{{ cat.leader.firstName }} {{ cat.leader.lastName }}</span>
          </template>
          <span v-else class="tile-value muted">—</span>
        </div>
      </div>
    </template>
    <p v-else class="empty">No season data</p>
  </div>
</template>

<style scoped>
.leaders-panel {
  display: flex;
  flex-direction: column;
}

.leader-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
}

.leader-tile {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.tile-label {
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.tile-value {
  font-size: 1.375rem;
  font-weight: 700;
  color: var(--color-text-primary);
  text-decoration: none;
  font-family: var(--font-mono);
}

.tile-value:hover {
  color: var(--color-accent);
}

.tile-value.muted {
  color: var(--color-text-secondary);
  font-size: 1rem;
  font-weight: 400;
}

.tile-player {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.empty {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}
</style>
