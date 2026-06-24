<script lang="ts" setup>
withDefaults(
  defineProps<{
    icon: string
    variant?: 'secondary' | 'danger'
    size?: 'sm' | 'md'
    rounded?: boolean
    disabled?: boolean
  }>(),
  {
    variant: 'secondary',
    size: 'sm',
    rounded: false,
    disabled: false,
  },
)

// No custom emits — click is a native DOM event and flows through $attrs.
// Callers use @click directly on <IconButton>.
</script>

<template>
  <button
    type="button"
    :disabled="disabled"
    :class="['icon-btn', `icon-btn--${variant}`, `icon-btn--${size}`, { 'icon-btn--rounded': rounded }]"
    v-bind="$attrs"
  >
    <i :class="icon" />
  </button>
</template>

<style scoped>
.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.icon-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.icon-btn--rounded {
  border-radius: 50%;
}

/* Size variants */
.icon-btn--sm {
  width: 1.75rem;
  height: 1.75rem;
  font-size: 0.8125rem;
}

.icon-btn--md {
  width: 2.25rem;
  height: 2.25rem;
  font-size: 0.9375rem;
}

/* Style variants */
.icon-btn--secondary {
  color: var(--color-text-secondary);
}

.icon-btn--secondary:hover:not(:disabled) {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.icon-btn--danger {
  color: var(--color-error, #dc2626);
}

.icon-btn--danger:hover:not(:disabled) {
  background: color-mix(in srgb, var(--color-error, #dc2626) 15%, transparent);
}
</style>
