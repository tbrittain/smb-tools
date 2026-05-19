<script lang="ts" setup>
import { useFranchiseStore } from '../stores/franchise'

const franchiseStore = useFranchiseStore()
</script>

<template>
  <div class="dashboard">
    <p class="placeholder">
      Dashboard for <strong>{{ franchiseStore.active?.name }}</strong> — content coming in Phase 5.
    </p>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 2rem;
}

.placeholder {
  color: var(--color-text-secondary);
  font-size: 0.9375rem;
}
</style>
