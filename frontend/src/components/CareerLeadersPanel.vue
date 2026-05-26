<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'
import AppLink from './AppLink.vue'

defineProps<{
  leaders: main.CareerLeadersDTO | null
}>()

const groups: { label: string; categories: { key: keyof main.CareerLeadersDTO; label: string }[] }[] = [
  {
    label: 'Batting',
    categories: [
      { key: 'hr', label: 'Home Runs' },
      { key: 'hits', label: 'Hits' },
      { key: 'rbi', label: 'RBI' },
    ],
  },
  {
    label: 'Pitching',
    categories: [
      { key: 'wins', label: 'Wins' },
      { key: 'strikeouts', label: 'Strikeouts' },
      { key: 'saves', label: 'Saves' },
    ],
  },
]
</script>

<template>
  <div v-if="leaders" class="career-leaders">
    <div v-for="group in groups" :key="group.label" class="career-group">
      <span class="group-label">{{ group.label }}</span>
      <div class="group-grid">
        <div v-for="cat in group.categories" :key="cat.key" class="category">
          <h4 class="cat-label">{{ cat.label }}</h4>
          <ol class="leader-list">
            <li
              v-for="(row, i) in leaders[cat.key]"
              :key="row.playerId"
              class="leader-row"
            >
              <span class="rank">{{ i + 1 }}</span>
              <AppLink :to="`/players/${row.playerId}`" class="player-name">
                {{ row.firstName }} {{ row.lastName }}
              </AppLink>
              <span class="stat-val">{{ Math.round(row.statValue) }}</span>
            </li>
            <li v-if="!leaders[cat.key]?.length" class="leader-row empty-row">
              <span class="empty">—</span>
            </li>
          </ol>
        </div>
      </div>
    </div>
  </div>
  <p v-else class="empty-text">No career data</p>
</template>

<style scoped>
.career-leaders {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.career-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.group-label {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.07em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.group-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.category {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 0.75rem 1rem;
}

.cat-label {
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
  margin-bottom: 0.5rem;
}

.leader-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.leader-row {
  display: grid;
  grid-template-columns: 1rem 1fr auto;
  align-items: baseline;
  gap: 0.5rem;
  font-size: 0.8125rem;
}

.rank {
  color: var(--color-text-secondary);
  font-variant-numeric: tabular-nums;
}

.player-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.stat-val {
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  font-weight: 600;
}

.empty-row .empty {
  color: var(--color-text-secondary);
  grid-column: 1 / -1;
}

.empty-text {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}
</style>
