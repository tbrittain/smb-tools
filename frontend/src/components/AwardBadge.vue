<script lang="ts" setup>
import { computed } from 'vue'
import type { main } from '../../wailsjs/go/models'
import { getAwardIcon } from '../composables/useAwardIcons'

const props = defineProps<{
  award: main.AwardDTO
  size?: 'sm' | 'md' | 'lg'
  count?: number
}>()

const sizeClass = computed(() => props.size ?? 'md')
const icon = computed(() => getAwardIcon(props.award.originalName))

// Importance → visual tier
function tierClass(importance: number): string {
  if (importance === 0) return 'tier-gold'
  if (importance <= 2) return 'tier-silver'
  if (importance <= 3) return 'tier-bronze'
  return 'tier-muted'
}
</script>

<template>
  <span class="award-badge" :class="[tierClass(award.importance), sizeClass]" :title="award.name">
{{ count && count > 1 ? `${count}x ` : '' }}{{ icon ? icon + ' ' : '' }}{{ award.name }}
  </span>
</template>

<style scoped>
.award-badge {
  display: inline-flex;
  align-items: center;
  border-radius: 4px;
  font-weight: 600;
  white-space: nowrap;
  line-height: 1;
}

.award-badge.lg {
  font-size: 0.875rem;
  padding: 0.3rem 0.65rem;
}

.award-badge.md {
  font-size: 0.72rem;
  padding: 0.25rem 0.5rem;
}

.award-badge.sm {
  font-size: 0.65rem;
  padding: 0.15rem 0.35rem;
}

/* Tier styles */
.tier-gold {
  background: #7c5c00;
  color: #ffd966;
  border: 1px solid #a07800;
}

.tier-silver {
  background: var(--color-surface-3, #3a3a3a);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

.tier-bronze {
  background: var(--color-surface-2, #2e2e2e);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}

.tier-muted {
  background: transparent;
  color: var(--color-text-muted, #888);
  border: 1px solid var(--color-border);
  font-weight: 400;
}
</style>
