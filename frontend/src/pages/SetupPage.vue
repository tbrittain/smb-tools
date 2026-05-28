<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { ListFranchiseSources, ListSnapshots, ReimportSeasonFromSnapshot, SyncSeason } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import SaveFilePicker from '../components/SaveFilePicker.vue'
import SnapshotPicker from '../components/SnapshotPicker.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'
import { useFranchiseStore } from '../stores/franchise'
import { useStatHighlightsStore } from '../stores/statHighlights'

const franchiseStore = useFranchiseStore()
const highlightsStore = useStatHighlightsStore()

// ── Source history ────────────────────────────────────────────────────────────

const sources = ref<main.FranchiseSourceDTO[]>([])

const sortedSources = computed(() => [...sources.value].sort((a, b) => a.seasonOffset - b.seasonOffset))

const lastSeason = computed(() => franchiseStore.active?.lastSeason ?? 0)

const sourcesWithRanges = computed(() => {
  const s = sortedSources.value
  return s.map((src, i) => {
    const start = src.seasonOffset + 1
    const nextOffset = i + 1 < s.length ? s[i + 1].seasonOffset : null
    const end = nextOffset ?? lastSeason.value
    return { ...src, start, end }
  })
})

// Passed to SaveFilePicker so previously-used paths show a usage label.
const usedSourceLabels = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {}
  for (const s of sourcesWithRanges.value) {
    if (s.isLegacy) continue
    const label =
      s.end === 0 || s.start > s.end
        ? 'Previously used — no seasons synced yet'
        : s.start === s.end
          ? `Previously used · Season ${s.start}`
          : `Previously used · Seasons ${s.start}–${s.end}`
    map[s.saveFilePath] = label
  }
  return map
})

async function loadSources() {
  if (!franchiseStore.active) return
  try {
    sources.value = await ListFranchiseSources(franchiseStore.active.id)
  } catch {
    // non-fatal — source history just won't display
  }
}

const { set } = useBreadcrumbs()
onMounted(() => set([{ label: 'Setup' }]))
onMounted(loadSources)
onMounted(loadSnapshots)

function sourceDisplayName(src: (typeof sourcesWithRanges.value)[number]): string {
  if (src.isLegacy) return 'Legacy import'
  return src.saveFilePath.split(/[\\/]/).pop() ?? src.saveFilePath
}

function formatRange(start: number, end: number): string {
  if (end === 0 || start > end) return 'No seasons synced yet'
  if (start === end) return `Season ${start}`
  return `Seasons ${start}–${end}`
}

function formatDate(iso: string): string {
  return iso ? new Date(iso).toLocaleDateString() : ''
}

// ── Replace active source ────────────────────────────────────────────────────

const showReplacePicker = ref(false)
const replaceError = ref<string | null>(null)

async function handleSaveFileChange(path: string, leagueGUID: string) {
  if (!franchiseStore.active) return
  replaceError.value = null
  try {
    if (franchiseStore.active.hasActiveSource) {
      await franchiseStore.replaceActiveFranchiseSource(franchiseStore.active.id, path, leagueGUID)
    } else {
      await franchiseStore.setInitialSource(franchiseStore.active.id, path, leagueGUID)
    }
    showReplacePicker.value = false
    await loadSources()
  } catch (e) {
    replaceError.value = String(e)
  }
}

// ── Legacy continuation (first real source after a legacy import) ────────────

const pendingLegacySource = ref<{ path: string; leagueGUID: string; numSeasons: number } | null>(null)
const legacyFranchiseSeason = ref(0)
const legacyConnectError = ref<string | null>(null)

const legacyOffsetValid = computed(
  () => pendingLegacySource.value !== null && legacyFranchiseSeason.value >= pendingLegacySource.value.numSeasons,
)

function handleLegacyFileSelected(path: string, leagueGUID: string, probe: main.SaveFileCandidateDTO) {
  pendingLegacySource.value = { path, leagueGUID, numSeasons: probe.numSeasons }
  legacyFranchiseSeason.value = lastSeason.value
  legacyConnectError.value = null
}

async function confirmLegacyConnection() {
  if (!franchiseStore.active || !pendingLegacySource.value) return
  legacyConnectError.value = null
  const { path, leagueGUID, numSeasons } = pendingLegacySource.value
  const offset = legacyFranchiseSeason.value - numSeasons
  try {
    await franchiseStore.addFranchiseSource(franchiseStore.active.id, path, leagueGUID, offset)
    pendingLegacySource.value = null
    await loadSources()
  } catch (e) {
    legacyConnectError.value = String(e)
  }
}

function cancelLegacyConnection() {
  pendingLegacySource.value = null
  legacyConnectError.value = null
}

// ── Fork source ──────────────────────────────────────────────────────────────

const showForkForm = ref(false)
const forkSeasonOffset = ref(0)
const forkError = ref<string | null>(null)

function openForkForm() {
  forkSeasonOffset.value = franchiseStore.active?.lastSeason ?? 0
  showForkForm.value = true
}

async function handleForkSourceChange(path: string, leagueGUID: string) {
  if (!franchiseStore.active) return
  forkError.value = null
  try {
    await franchiseStore.addFranchiseSource(franchiseStore.active.id, path, leagueGUID, forkSeasonOffset.value)
    showForkForm.value = false
    await loadSources()
  } catch (e) {
    forkError.value = String(e)
  }
}

// ── Sync ────────────────────────────────────────────────────────────────────

const syncing = ref(false)
const syncError = ref<string | null>(null)
const lastResult = ref<main.SyncSeasonResult | null>(null)

async function handleSync() {
  syncing.value = true
  syncError.value = null
  lastResult.value = null
  try {
    lastResult.value = await SyncSeason()
    highlightsStore.invalidate()
    if (franchiseStore.active) {
      await franchiseStore.selectFranchise(franchiseStore.active.id)
    }
  } catch (e) {
    syncError.value = String(e)
  } finally {
    syncing.value = false
  }
}

// ── Reimport from snapshot ────────────────────────────────────────────────────

const snapshots = ref<main.SnapshotDTO[]>([])
const snapshotsLoading = ref(false)
const selectedSnapshotId = ref<number | null>(null)
const reimporting = ref(false)
const reimportError = ref<string | null>(null)
const reimportResult = ref<main.ReimportSeasonResult | null>(null)

const selectedSnapshot = computed(() => snapshots.value.find((s) => s.id === selectedSnapshotId.value) ?? null)

async function loadSnapshots() {
  if (!franchiseStore.active) return
  snapshotsLoading.value = true
  try {
    snapshots.value = await ListSnapshots()
  } catch {
    // non-fatal — snapshot list just won't display
  } finally {
    snapshotsLoading.value = false
  }
}

async function handleReimport() {
  if (!selectedSnapshot.value) return
  reimporting.value = true
  reimportError.value = null
  reimportResult.value = null
  try {
    reimportResult.value = await ReimportSeasonFromSnapshot(selectedSnapshot.value.id, selectedSnapshot.value.seasonNum)
    highlightsStore.invalidate()
  } catch (e) {
    reimportError.value = String(e)
  } finally {
    reimporting.value = false
  }
}
</script>

<template>
  <div class="setup-page">
    <header class="page-header">
      <h2>Setup</h2>
      <span class="subtitle">{{ franchiseStore.active?.name }}</span>
    </header>

    <!-- Connected save files -->
    <section class="card-section">
      <h3>Connected Save Files</h3>

      <!-- No source yet — show picker to connect one -->
      <template v-if="!franchiseStore.active?.hasActiveSource && !sourcesWithRanges.length">
        <p class="hint-text">Connect a save file to enable syncing.</p>
        <SaveFilePicker @change="handleSaveFileChange" />
        <p v-if="replaceError" class="error-text">{{ replaceError }}</p>
      </template>

      <!-- Source history list -->
      <template v-else>
        <div class="source-list">
          <div
            v-for="(src, i) in sourcesWithRanges"
            :key="src.id"
            class="source-row"
          >
            <div class="source-info">
              <span class="source-name" :class="{ 'source-name--legacy': src.isLegacy }">
                {{ sourceDisplayName(src) }}
              </span>
              <span class="source-meta">
                <template v-if="!src.isLegacy">
                  {{ formatRange(src.start, src.end) }} ·
                </template>
                Added {{ formatDate(src.addedAt) }}
              </span>
            </div>
            <!-- Replace link only on the active (last) real source -->
            <button
              v-if="i === sourcesWithRanges.length - 1 && !src.isLegacy && !showReplacePicker"
              class="replace-link"
              @click="showReplacePicker = true"
            >
              replace
            </button>
          </div>
        </div>

        <!-- No real source yet (legacy-only franchise) — connect with season mapping -->
        <template v-if="franchiseStore.active?.hasLegacySource && !franchiseStore.active?.hasActiveSource && !showReplacePicker">
          <template v-if="!pendingLegacySource">
            <p class="hint-text">Connect a save file to enable syncing.</p>
            <SaveFilePicker :used-source-labels="usedSourceLabels" @change="handleLegacyFileSelected" />
          </template>
          <template v-else>
            <p class="hint-text">
              This save game has <strong>{{ pendingLegacySource.numSeasons }}</strong> season(s).
              Which franchise season does Season&nbsp;{{ pendingLegacySource.numSeasons }} correspond to?
            </p>
            <div class="fork-offset-row">
              <label class="fork-label">Franchise season</label>
              <input
                v-model.number="legacyFranchiseSeason"
                type="number"
                :min="pendingLegacySource.numSeasons"
                class="fork-offset-input"
              />
            </div>
            <p class="hint-text">
              Season {{ pendingLegacySource.numSeasons + 1 }} of this save game will become
              franchise Season {{ legacyFranchiseSeason + 1 }}.
            </p>
            <p v-if="legacyConnectError" class="error-text">{{ legacyConnectError }}</p>
            <div class="legacy-actions">
              <AppButton variant="primary" size="sm" :disabled="!legacyOffsetValid" @click="confirmLegacyConnection">
                Connect
              </AppButton>
              <AppButton variant="ghost" size="sm" @click="cancelLegacyConnection">Cancel</AppButton>
            </div>
          </template>
        </template>

        <!-- Inline replace picker -->
        <template v-if="showReplacePicker">
          <p class="hint-text">
            Select the corrected save file. This replaces the current active source in-place —
            use this only to fix a wrong file selection, not to add a new league.
          </p>
          <SaveFilePicker
            :selected-path="franchiseStore.active?.activeSourcePath"
            :used-source-labels="usedSourceLabels"
            @change="handleSaveFileChange"
          />
          <p v-if="replaceError" class="error-text">{{ replaceError }}</p>
          <AppButton variant="ghost" size="sm" @click="showReplacePicker = false">Cancel</AppButton>
        </template>
      </template>
    </section>

    <!-- Fork to new league — deliberately separate, below, with explanation -->
    <section
      v-if="franchiseStore.active?.hasActiveSource && !showReplacePicker"
      class="card-section card-section--fork"
    >
      <div class="fork-header">
        <div class="fork-title-group">
          <h3>Fork Franchise From New Save Game</h3>
          <span class="advanced-badge">Advanced</span>
        </div>
        <AppButton v-if="!showForkForm" variant="ghost" size="sm" @click="openForkForm">
          Add Fork Source
        </AppButton>
      </div>

      <p class="fork-description">
        SMB4 lets you export an existing franchise to a new league at any point in time,
        effectively forking it into a fresh save game. If you want smb-tools to treat that
        new league as a continuation of this franchise — rather than a separate one — connect
        its save file here. Imported seasons from the new save will be numbered sequentially
        after your last synced season, keeping your complete franchise history in one place.
      </p>

      <template v-if="showForkForm">
        <div class="fork-offset-row">
          <label class="fork-label">Season offset</label>
          <input v-model.number="forkSeasonOffset" type="number" min="0" class="fork-offset-input" />
        </div>
        <p class="hint-text">
          Seasons from this source will be numbered starting after Season {{ forkSeasonOffset }}.
          This should match your last synced season ({{ lastSeason }}).
        </p>
        <SaveFilePicker :used-source-labels="usedSourceLabels" @change="handleForkSourceChange" />
        <p v-if="forkError" class="error-text">{{ forkError }}</p>
        <AppButton variant="ghost" size="sm" @click="showForkForm = false">Cancel</AppButton>
      </template>
    </section>

    <!-- Sync -->
    <section class="card-section">
      <h3>Sync Season</h3>
      <p class="sync-help">
        Reads the current season from your save file. Sync once after the regular season
        ends, then again after the playoffs conclude — <strong>before</strong> progressing
        to the offseason. Advancing to the offseason triggers in-game data compaction that
        can cause stat loss.
      </p>
      <p v-if="syncError" class="error-text">{{ syncError }}</p>
      <div v-if="lastResult" class="sync-result">
        <span>✓ Season {{ lastResult.seasonNum }} synced —</span>
        <span>{{ lastResult.players }} players,</span>
        <span>{{ lastResult.teams }} teams,</span>
        <span>{{ lastResult.games }} games</span>
        <span v-if="lastResult.playoffGames">, {{ lastResult.playoffGames }} playoff games</span>
      </div>
      <AppButton
        variant="primary"
        :disabled="syncing || !franchiseStore.active?.hasActiveSource"
        @click="handleSync"
      >
        {{ syncing ? 'Syncing…' : 'Sync Season' }}
      </AppButton>
    </section>

    <!-- Reimport from snapshot -->
    <section v-if="franchiseStore.active" class="card-section">
      <h3>Reimport Season from Snapshot</h3>
      <p class="sync-help">
        Select a previously captured snapshot and reimport its season data. Use this to
        recover from a bad sync or correct data after the playoffs. Your awards for that
        season will not be affected.
      </p>
      <SnapshotPicker
        :snapshots="snapshots"
        :loading="snapshotsLoading"
        :selected-id="selectedSnapshotId"
        @update:selected-id="selectedSnapshotId = $event"
      />
      <p v-if="reimportError" class="error-text">{{ reimportError }}</p>
      <div v-if="reimportResult" class="sync-result">
        <span>✓ Season {{ reimportResult.seasonNum }} reimported —</span>
        <span>{{ reimportResult.players }} players,</span>
        <span>{{ reimportResult.teams }} teams,</span>
        <span>{{ reimportResult.games }} games</span>
        <span v-if="reimportResult.playoffGames">, {{ reimportResult.playoffGames }} playoff games</span>
      </div>
      <AppButton
        variant="secondary"
        :disabled="reimporting || selectedSnapshotId === null"
        @click="handleReimport"
      >
        {{ reimporting ? 'Reimporting…' : 'Reimport Selected Snapshot' }}
      </AppButton>
    </section>
  </div>
</template>

<style scoped>
.setup-page {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
  max-width: 680px;
  margin: 0 auto;
  width: 100%;
}

.page-header {
  display: flex;
  align-items: baseline;
  gap: 1rem;
}

h2 {
  font-size: 1.4rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.subtitle {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.card-section {
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
}

.card-section--fork {
  border-style: dashed;
}

/* Source history */
.source-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.source-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.625rem 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.source-info {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.source-name {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.source-name--legacy {
  color: var(--color-text-secondary);
  font-style: italic;
}

.source-meta {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.replace-link {
  background: none;
  border: none;
  padding: 0;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.replace-link:hover {
  color: var(--color-accent);
}

/* Fork section */
.fork-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.fork-title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.advanced-badge {
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
  border-radius: 3px;
  padding: 1px 5px;
}

.fork-description {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.55;
}

.fork-offset-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.fork-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.fork-offset-input {
  width: 80px;
  padding: 0.25rem 0.5rem;
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  color: var(--color-text-primary);
  font-size: 0.875rem;
}

/* Sync */
.sync-help {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.sync-result {
  display: flex;
  gap: 0.375rem;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  flex-wrap: wrap;
}

.legacy-actions {
  display: flex;
  gap: 0.5rem;
}

.error-text { font-size: 0.875rem; color: var(--color-error); }
.hint-text  { font-size: 0.8125rem; color: var(--color-text-secondary); line-height: 1.5; }
</style>
