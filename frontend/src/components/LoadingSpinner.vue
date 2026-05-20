<script lang="ts" setup></script>

<template>
  <div class="spinner" aria-label="Loading" role="status">
    <div class="spinner-ring" />
  </div>
</template>

<style scoped>
.spinner {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 2rem;
}

.spinner-ring {
  width: 28px;
  height: 28px;
  border: 3px solid var(--color-surface-3);
  border-top-color: var(--color-accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
