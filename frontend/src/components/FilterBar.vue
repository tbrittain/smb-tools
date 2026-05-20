<template>
  <div class="filter-bar">
    <slot />
  </div>
</template>

<style scoped>
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem 1.25rem;
  padding: 0.75rem 1rem;
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}
</style>
