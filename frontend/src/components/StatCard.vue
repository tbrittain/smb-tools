<script lang="ts" setup>
import AppLink from './AppLink.vue'

defineProps<{
  label: string
  value: string
  subtext?: string
  to?: string // router-link target; omit for non-linked cards
}>()
</script>

<template>
  <div class="stat-card">
    <span class="label">{{ label }}</span>
    <AppLink v-if="to" :to="to" class="value">{{ value }}</AppLink>
    <span v-else class="value">{{ value }}</span>
    <span v-if="subtext" class="subtext">{{ subtext }}</span>
  </div>
</template>

<style scoped>
.stat-card {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 0.875rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 120px;
}

.label {
  font-size: 0.75rem;
  font-weight: 500;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.value {
  font-size: 1.25rem;
  font-weight: 600;
}

.subtext {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}
</style>
