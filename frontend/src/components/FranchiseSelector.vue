<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue'
import { ProbeFranchiseSaveFile } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from './AppButton.vue'
import LoadingSpinner from './LoadingSpinner.vue'

const props = defineProps<{
  franchises: main.FranchiseDTO[]
  activeId?: string
}>()

const emit = defineEmits<{
  select: [id: string]
  create: []
  import: []
}>()

// ── Live save file probe state ────────────────────────────────────────────────

const probeResults = ref<Record<string, main.SaveFileCandidateDTO>>({})
const probing = ref<Record<string, boolean>>({})
const anyProbing = ref(false)

async function probeAll() {
  const toProbe = props.franchises.filter((f) => f.hasActiveSource)
  if (toProbe.length === 0) return

  anyProbing.value = true
  const startState: Record<string, boolean> = {}
  for (const f of toProbe) startState[f.id] = true
  probing.value = { ...probing.value, ...startState }

  await Promise.allSettled(
    toProbe.map(async (f) => {
      try {
        const result = await ProbeFranchiseSaveFile(f.id)
        probeResults.value = { ...probeResults.value, [f.id]: result }
      } catch {
        // ignore — card shows static data only
      } finally {
        probing.value = { ...probing.value, [f.id]: false }
      }
    }),
  )
  anyProbing.value = false
}

onMounted(probeAll)

// Probe any newly added franchises when the list changes
watch(
  () => props.franchises.map((f) => f.id).join(','),
  (next, prev) => {
    if (next !== prev) probeAll()
  },
)

// ── Display helpers ───────────────────────────────────────────────────────────

function gameVersionLabel(version: string): string {
  return version === 'smb4' ? 'SMB4' : 'SMB3'
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

interface SyncInfo {
  syncedLine: string
  gameLine: string | null
  behind: number
  noSaveFile: boolean
}

function syncInfo(f: main.FranchiseDTO): SyncInfo {
  if (!f.hasActiveSource) {
    return {
      syncedLine: f.lastSynced ? `Synced Season ${f.lastSeason} · ${formatDate(f.lastSynced)}` : 'Never synced',
      gameLine: null,
      behind: 0,
      noSaveFile: true,
    }
  }

  const probe = probeResults.value[f.id]
  const syncedLine = f.lastSynced ? `Synced Season ${f.lastSeason} · ${formatDate(f.lastSynced)}` : 'Never synced'

  if (!probe || probe.numSeasons === 0) {
    return { syncedLine, gameLine: null, behind: 0, noSaveFile: false }
  }

  const behind = probe.numSeasons - (f.lastSeason ?? 0)
  const seasonWord = probe.numSeasons === 1 ? 'season' : 'seasons'
  const gameLine = `${probe.numSeasons} ${seasonWord} in game`

  return { syncedLine, gameLine, behind: Math.max(0, behind), noSaveFile: false }
}

function liveLabel(f: main.FranchiseDTO): string | null {
  const probe = probeResults.value[f.id]
  if (!probe) return null
  const parts: string[] = []
  if (probe.leagueName) parts.push(probe.leagueName)
  if (probe.playerTeamName) parts.push(probe.playerTeamName)
  return parts.length > 0 ? parts.join(' · ') : null
}
</script>

<template>
  <div class="franchise-selector">
    <div class="header">
      <h2>Select Franchise</h2>
      <div class="header-actions">
        <AppButton
          variant="ghost"
          size="sm"
          :disabled="anyProbing"
          title="Refresh live game data from save files"
          @click="probeAll"
        >
          {{ anyProbing ? 'Refreshing…' : 'Refresh' }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="emit('import')">Import legacy…</AppButton>
        <AppButton variant="primary" size="sm" @click="emit('create')">+ New Franchise</AppButton>
      </div>
    </div>

    <div v-if="props.franchises.length === 0" class="empty-state">
      <p>No franchises yet. Create one to get started, or import from SmbExplorerCompanion.</p>
    </div>

    <ul v-else class="franchise-list">
      <li
        v-for="f in props.franchises"
        :key="f.id"
        class="franchise-item"
        :class="{ active: f.id === props.activeId }"
        @click="emit('select', f.id)"
      >
        <!-- Top row: name + version badge -->
        <div class="item-header">
          <span class="franchise-name">{{ f.name }}</span>
          <div class="item-badges">
            <span class="version-badge">{{ gameVersionLabel(f.gameVersion) }}</span>
            <span v-if="f.id === props.activeId" class="active-badge">Active</span>
          </div>
        </div>

        <!-- Live probe line: league · team (or placeholder) -->
        <div class="live-line">
          <template v-if="probing[f.id]">
            <LoadingSpinner size="xs" />
            <span class="hint">Reading save file…</span>
          </template>
          <template v-else-if="liveLabel(f)">
            <span class="live-label">{{ liveLabel(f) }}</span>
          </template>
          <template v-else-if="!f.hasActiveSource">
            <span class="hint warn-text">No save file configured</span>
          </template>
        </div>

        <!-- Sync status row -->
        <div class="sync-row">
          <span
            class="sync-line"
            :class="{ unsync: !syncInfo(f).noSaveFile && syncInfo(f).behind > 0 }"
          >
            {{ syncInfo(f).syncedLine }}
          </span>

          <template v-if="syncInfo(f).gameLine">
            <span class="dot-sep">·</span>
            <span class="game-line">{{ syncInfo(f).gameLine }}</span>
            <span v-if="syncInfo(f).behind > 0" class="behind-pill">
              {{ syncInfo(f).behind }}
              {{ syncInfo(f).behind === 1 ? 'season' : 'seasons' }} behind
            </span>
            <span v-else-if="syncInfo(f).behind === 0 && f.lastSynced" class="uptodate-pill">
              Up to date
            </span>
          </template>
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.franchise-selector {
  max-width: 620px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
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
  flex-direction: column;
  gap: 0.375rem;
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
  background: var(--color-surface-3, var(--color-surface-2));
}

/* Top row */
.item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.franchise-name {
  font-size: 1rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.item-badges {
  display: flex;
  gap: 0.375rem;
  align-items: center;
  flex-shrink: 0;
}

.version-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  padding: 0.15rem 0.45rem;
  border-radius: 3px;
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
}

.active-badge {
  font-size: 0.75rem;
  padding: 0.2rem 0.6rem;
  background: var(--color-accent);
  color: #fff;
  border-radius: 99px;
}

/* Live probe line */
.live-line {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  min-height: 1.1rem;
}

.live-label {
  font-size: 0.875rem;
  color: var(--color-text-primary);
}

/* Sync status row */
.sync-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.375rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.sync-line {
  color: var(--color-text-secondary);
}

.dot-sep {
  color: var(--color-border);
}

.game-line {
  color: var(--color-text-secondary);
}

.behind-pill {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 99px;
  background: color-mix(in srgb, #f59e0b 15%, transparent);
  color: #f59e0b;
}

.uptodate-pill {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 99px;
  background: color-mix(in srgb, var(--color-accent) 15%, transparent);
  color: var(--color-accent);
}

.hint {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.warn-text {
  color: color-mix(in srgb, #f59e0b 80%, var(--color-text-secondary));
}
</style>
