<script lang="ts" setup>
import { RouterLink } from 'vue-router'

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
    <component
      :is="to ? RouterLink : 'span'"
      :to="to"
      class="value"
      :class="{ linked: !!to }"
    >
      {{ value }}
    </component>
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
  color: var(--color-text-primary);
  text-decoration: none;
}

.value.linked:hover {
  color: var(--color-accent);
}

.subtext {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}
</style>
