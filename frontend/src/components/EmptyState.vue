<script lang="ts" setup>
defineProps<{
  message: string
  subtext?: string
}>()
</script>

<template>
  <div class="empty-state">
    <p class="empty-message">{{ message }}</p>
    <p v-if="subtext" class="empty-subtext">{{ subtext }}</p>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 3rem 2rem;
  text-align: center;
}

.empty-message {
  font-size: 0.9375rem;
  color: var(--color-text-secondary);
}

.empty-subtext {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  opacity: 0.7;
}
</style>
