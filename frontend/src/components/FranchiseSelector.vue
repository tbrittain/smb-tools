<script lang="ts" setup>
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Menu from 'primevue/menu'
import { computed, onMounted, ref, watch } from 'vue'
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
  delete: [id: string]
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

// ── Delete flow ───────────────────────────────────────────────────────────────

const franchiseMenu = ref<{ toggle: (event: Event) => void } | null>(null)
const menuTargetId = ref<string | null>(null)
const showDeleteDialog = ref(false)
const deleteNameInput = ref('')

const franchiseToDelete = computed(() => props.franchises.find((f) => f.id === menuTargetId.value) ?? null)

const menuItems = computed(() => [
  {
    label: 'Delete franchise',
    command: () => {
      deleteNameInput.value = ''
      showDeleteDialog.value = true
    },
  },
])

function openMenu(event: MouseEvent, id: string) {
  menuTargetId.value = id
  franchiseMenu.value?.toggle(event)
}

const deleteNameMatches = computed(
  () => !!franchiseToDelete.value && deleteNameInput.value === franchiseToDelete.value.name,
)

function submitDelete() {
  if (!deleteNameMatches.value || !menuTargetId.value) return
  emit('delete', menuTargetId.value)
  showDeleteDialog.value = false
}

function onDialogHide() {
  menuTargetId.value = null
  deleteNameInput.value = ''
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
        <AppButton variant="primary" size="sm" @click="emit('create')">+ New Franchise</AppButton>
      </div>
    </div>

    <div v-if="props.franchises.length === 0" class="empty-state">
      <p>No franchises yet.</p>
    </div>

    <ul v-else class="franchise-list">
      <li
        v-for="f in props.franchises"
        :key="f.id"
        class="franchise-item"
        :class="{ active: f.id === props.activeId }"
        @click="emit('select', f.id)"
      >
        <!-- Top row: name + version badge + menu -->
        <div class="item-header">
          <span class="franchise-name">{{ f.name }}</span>
          <div class="item-right">
            <div class="item-badges">
              <span class="version-badge">{{ gameVersionLabel(f.gameVersion) }}</span>
              <span v-if="f.id === props.activeId" class="active-badge">Active</span>
            </div>
            <button
              class="menu-btn"
              title="Franchise options"
              aria-label="Franchise options"
              @click.stop="openMenu($event, f.id)"
            >
              ⋮
            </button>
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

    <div class="import-footer">
      <span class="import-hint">Have a SmbExplorerCompanion database?</span>
      <button class="import-link" @click="emit('import')">Import franchise</button>
    </div>

    <!-- Shared popup menu for franchise actions -->
    <Menu ref="franchiseMenu" :model="menuItems" popup />

    <!-- Delete confirmation dialog -->
    <Dialog
      v-model:visible="showDeleteDialog"
      modal
      header="Delete franchise"
      :style="{ width: '26rem' }"
      @hide="onDialogHide"
    >
      <div class="delete-dialog-body">
        <p class="delete-warning">
          This permanently deletes
          <strong>{{ franchiseToDelete?.name }}</strong>
          and all its seasons, stats, and save game snapshots. This cannot be undone.
        </p>

        <div class="delete-confirm-field">
          <label class="delete-confirm-label" for="delete-name-input">
            Type the franchise name to confirm:
          </label>
          <InputText
            id="delete-name-input"
            v-model="deleteNameInput"
            class="delete-name-input"
            autocomplete="off"
            @keydown.enter.prevent="submitDelete"
          />
        </div>
      </div>

      <template #footer>
        <div class="delete-dialog-footer">
          <AppButton variant="secondary" size="sm" @click="showDeleteDialog = false">Cancel</AppButton>
          <AppButton variant="danger" size="sm" :disabled="!deleteNameMatches" @click="submitDelete">
            Delete franchise
          </AppButton>
        </div>
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.franchise-selector {
  max-width: 100%;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  gap: 2rem;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
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

.item-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.item-badges {
  display: flex;
  gap: 0.375rem;
  align-items: center;
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

/* Kebab menu button */
.menu-btn {
  background: none;
  border: none;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  color: var(--color-text-secondary);
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
  opacity: 0.45;
  transition: opacity 0.15s, background 0.15s;
}

.franchise-item:hover .menu-btn,
.menu-btn:hover,
.menu-btn:focus-visible {
  opacity: 1;
  background: var(--color-surface-1);
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

.import-footer {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1.25rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--color-border);
  font-size: 0.8125rem;
}

.import-hint {
  color: var(--color-text-secondary);
}

.import-link {
  background: none;
  border: none;
  padding: 0;
  color: var(--color-accent);
  font-size: 0.8125rem;
  font-family: inherit;
  cursor: pointer;
}

.import-link:hover {
  text-decoration: underline;
}

/* Delete dialog */
.delete-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.delete-warning {
  font-size: 0.9375rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.delete-warning strong {
  color: var(--color-text-primary);
}

.delete-confirm-field {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.delete-confirm-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.delete-name-input {
  width: 100%;
}

.delete-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.625rem;
}
</style>
