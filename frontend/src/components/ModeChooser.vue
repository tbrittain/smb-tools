<script lang="ts" setup>
defineEmits<{
  select: [mode: 'franchise' | 'league-transfer']
}>()
</script>

<template>
  <div class="mode-chooser">
    <div class="mode-chooser-brand">
      <h1>smb-tools</h1>
      <p>What would you like to do?</p>
    </div>

    <div class="mode-cards">
      <button class="mode-card" @click="$emit('select', 'franchise')">
        <i class="pi pi-chart-line mode-card-icon" />
        <span class="mode-card-title">Franchise Tracker</span>
        <span class="mode-card-desc">Track stats, standings, and history for one of your SMB4 franchises.</span>
      </button>

      <button class="mode-card" @click="$emit('select', 'league-transfer')">
        <i class="pi pi-share-alt mode-card-icon" />
        <span class="mode-card-title">League Transfer</span>
        <span class="mode-card-desc">Export a league to share with someone else, or import one they sent you.</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.mode-chooser {
  width: 100%;
  max-width: 680px;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.mode-chooser-brand h1 {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.mode-chooser-brand p {
  font-size: 0.9375rem;
  color: var(--color-text-secondary);
  margin-top: 0.25rem;
}

.mode-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.mode-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 1.5rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  transition: border-color 0.15s;
}

.mode-card:hover,
.mode-card:focus-visible {
  border-color: var(--color-accent);
}

.mode-card-icon {
  font-size: 1.5rem;
  color: var(--color-accent);
}

.mode-card-title {
  font-size: 1.0625rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.mode-card-desc {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}
</style>
