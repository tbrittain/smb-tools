<script lang="ts" setup>
withDefaults(defineProps<{ size?: 'xs' | 'sm' | 'md' }>(), { size: 'md' })
</script>

<template>
  <div class="spinner" :class="`spinner--${size}`" aria-label="Loading" role="status">
    <div class="spinner-ring" />
  </div>
</template>

<style scoped>
.spinner {
  display: flex;
  justify-content: center;
  align-items: center;
}

.spinner--md {
  padding: 2rem;
}

.spinner--sm {
  padding: 0.75rem;
}

.spinner--xs {
  padding: 0;
}

.spinner-ring {
  border: 3px solid var(--color-surface-3);
  border-top-color: var(--color-accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.spinner--md .spinner-ring {
  width: 28px;
  height: 28px;
}

.spinner--sm .spinner-ring {
  width: 18px;
  height: 18px;
  border-width: 2px;
}

.spinner--xs .spinner-ring {
  width: 10px;
  height: 10px;
  border-width: 1.5px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
