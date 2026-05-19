<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'
import AppButton from './AppButton.vue'

const props = defineProps<{
  franchises: main.FranchiseDTO[]
  activeId?: string
}>()

const emit = defineEmits<{
  select: [id: string]
  create: []
}>()

function formatLastSynced(dto: main.FranchiseDTO): string {
  if (!dto.lastSynced) return 'Never synced'
  const d = new Date(dto.lastSynced)
  return `Season ${dto.lastSeason} · ${d.toLocaleDateString()}`
}

function gameVersionLabel(version: string): string {
  return version === 'smb4' ? 'SMB4' : 'SMB3'
}
</script>

<template>
  <div class="franchise-selector">
    <div class="header">
      <h2>Select Franchise</h2>
      <AppButton variant="primary" size="sm" @click="emit('create')">+ New Franchise</AppButton>
    </div>

    <div v-if="props.franchises.length === 0" class="empty-state">
      <p>No franchises yet. Create one to get started.</p>
    </div>

    <ul v-else class="franchise-list">
      <li
        v-for="f in props.franchises"
        :key="f.id"
        class="franchise-item"
        :class="{ active: f.id === props.activeId }"
        @click="emit('select', f.id)"
      >
        <div class="franchise-info">
          <span class="franchise-name">{{ f.name }}</span>
          <span class="franchise-meta">{{ gameVersionLabel(f.gameVersion) }} · {{ formatLastSynced(f) }}</span>
        </div>
        <span v-if="f.id === props.activeId" class="active-badge">Active</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.franchise-selector {
  max-width: 600px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

h2 {
  font-size: 1.4rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.empty-state {
  text-align: center;
  padding: 3rem 0;
  color: var(--color-text-secondary);
}

.franchise-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.franchise-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.15s;
}

.franchise-item:hover {
  border-color: var(--color-accent);
}

.franchise-item.active {
  border-color: var(--color-accent);
  background: var(--color-surface-3);
}

.franchise-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.franchise-name {
  font-size: 1rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.franchise-meta {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.active-badge {
  font-size: 0.75rem;
  padding: 0.2rem 0.6rem;
  background: var(--color-accent);
  color: #fff;
  border-radius: 99px;
}

</style>
