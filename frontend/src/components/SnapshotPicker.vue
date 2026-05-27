<script lang="ts" setup>
import { computed } from 'vue'
import type { main } from '../../wailsjs/go/models'

const props = defineProps<{
  snapshots: main.SnapshotDTO[]
  loading: boolean
  selectedId: number | null
}>()

const emit = defineEmits<{
  'update:selectedId': [id: number]
}>()

const sorted = computed(() =>
  [...props.snapshots].sort((a, b) => new Date(b.capturedAt).getTime() - new Date(a.capturedAt).getTime()),
)

function formatDate(iso: string): string {
  if (!iso) return ''
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function handleSelect(snap: main.SnapshotDTO) {
  if (!snap.fileExists) return
  emit('update:selectedId', snap.id)
}
</script>

<template>
  <div class="snapshot-picker">
    <div v-if="loading" class="snapshot-empty">Loading snapshots…</div>

    <div v-else-if="!sorted.length" class="snapshot-empty">No snapshots found.</div>

    <div v-else class="snapshot-list">
      <button
        v-for="snap in sorted"
        :key="snap.id"
        class="snapshot-row"
        :class="{
          'snapshot-row--selected': snap.id === selectedId,
          'snapshot-row--missing': !snap.fileExists,
        }"
        :disabled="!snap.fileExists"
        @click="handleSelect(snap)"
      >
        <div class="snapshot-info">
          <span class="snapshot-season">Season {{ snap.seasonNum }}</span>
          <span class="snapshot-meta">
            {{ formatDate(snap.capturedAt) }} · {{ formatSize(snap.fileSizeBytes) }}
          </span>
        </div>
        <span v-if="!snap.fileExists" class="snapshot-badge-missing">File missing</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.snapshot-picker {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.snapshot-empty {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  padding: 0.5rem 0;
}

.snapshot-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  max-height: 260px;
  overflow-y: auto;
}

.snapshot-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.625rem 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  width: 100%;
  transition: border-color 0.1s, background 0.1s;
}

.snapshot-row:hover:not(:disabled):not(.snapshot-row--selected) {
  border-color: var(--color-accent);
}

.snapshot-row--selected {
  border-color: var(--color-accent);
  background: color-mix(in srgb, var(--color-accent) 8%, var(--color-surface-2));
}

.snapshot-row--missing {
  opacity: 0.55;
  cursor: not-allowed;
}

.snapshot-info {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.snapshot-season {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.snapshot-meta {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.snapshot-badge-missing {
  flex-shrink: 0;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-error, #dc2626);
  border: 1px solid var(--color-error, #dc2626);
  border-radius: 3px;
  padding: 1px 5px;
}
</style>
