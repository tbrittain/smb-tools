<script lang="ts" setup>
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
    size?: 'sm' | 'md'
    disabled?: boolean
    loading?: boolean
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    variant: 'primary',
    size: 'md',
    disabled: false,
    loading: false,
    type: 'button',
  },
)

// No custom emits — click is a native DOM event and flows through $attrs.
// Callers use @click directly on <AppButton>.
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="['app-btn', `app-btn--${variant}`, `app-btn--${size}`]"
    v-bind="$attrs"
  >
    <i v-if="loading" class="pi pi-spinner pi-spin" />
    <slot />
  </button>
</template>

<style scoped>
.app-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  border: none;
  border-radius: 6px;
  font-family: inherit;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.15s, background 0.15s;
  white-space: nowrap;
}

.app-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* Size variants */
.app-btn--sm {
  padding: 0.3rem 0.75rem;
  font-size: 0.8125rem;
}

.app-btn--md {
  padding: 0.5rem 1.25rem;
  font-size: 0.9375rem;
}

/* Style variants */
.app-btn--primary {
  background: var(--color-accent);
  color: #fff;
}

.app-btn--primary:hover:not(:disabled) {
  opacity: 0.9;
}

.app-btn--secondary {
  background: transparent;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}

.app-btn--secondary:hover:not(:disabled) {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.app-btn--ghost {
  background: transparent;
  color: var(--color-accent);
  border: none;
  padding-left: 0;
  padding-right: 0;
}

.app-btn--ghost:hover:not(:disabled) {
  text-decoration: underline;
}

.app-btn--danger {
  background: var(--color-error, #dc2626);
  color: #fff;
}

.app-btn--danger:hover:not(:disabled) {
  opacity: 0.9;
}
</style>
