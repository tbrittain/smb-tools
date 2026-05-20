<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { BrowseSaveFile, GetSaveFileCandidates, ProbeLeagues } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from './AppButton.vue'

const emit = defineEmits<{
  create: [name: string, gameVersion: string, saveFilePath: string, leagueGUID: string]
  cancel: []
}>()

// SMB3 support is deferred — only SMB4 is available for now.
const GAME_VERSION = 'smb4'

// ── Form state ───────────────────────────────────────────────────────────────

const name = ref('')

// ── Save file discovery ───────────────────────────────────────────────────────

const candidates = ref<main.SaveFileCandidateDTO[]>([])
const loadingCandidates = ref(false)
const selectedPath = ref('')
const selectedLeagueGUID = ref('')
const probing = ref(false)
const probeResult = ref<main.SaveFileCandidateDTO | null>(null)

// ── Form submission ───────────────────────────────────────────────────────────

const submitting = ref(false)
const error = ref<string | null>(null)

onMounted(discoverSaveFiles)

async function discoverSaveFiles() {
  loadingCandidates.value = true
  try {
    const all = await GetSaveFileCandidates()
    // Only show SMB4 save files — SMB3 support is deferred
    candidates.value = all.filter((c) => c.gameVersion === GAME_VERSION)
    const match = candidates.value.find((c) => c.isFranchise)
    if (match) {
      await selectCandidate(match.path)
    }
  } catch {
    // Non-fatal — user can still browse manually
  } finally {
    loadingCandidates.value = false
  }
}

async function selectCandidate(path: string) {
  selectedPath.value = path
  selectedLeagueGUID.value = ''
  probeResult.value = null
  if (!path) return
  probing.value = true
  try {
    const leagues = await ProbeLeagues(path)
    if (leagues.length > 0) {
      const league = leagues[0]
      selectedLeagueGUID.value = league.leagueGUID
      probeResult.value = { ...league, path, gameVersion: GAME_VERSION }
    }
  } catch {
    // Non-fatal — can proceed without league detail
  } finally {
    probing.value = false
  }
}

async function handleBrowse() {
  try {
    const path = await BrowseSaveFile()
    if (path) {
      await selectCandidate(path)
      // Add the browsed path to the candidates list if not already there
      if (!candidates.value.find((c) => c.path === path)) {
        const entry: main.SaveFileCandidateDTO = probeResult.value ?? {
          path,
          gameVersion: GAME_VERSION,
          leagueName: '',
          numSeasons: 0,
          isFranchise: false,
          playerTeamName: '',
          leagueGUID: '',
        }
        candidates.value = [...candidates.value, entry]
      }
    }
  } catch (e) {
    error.value = String(e)
  }
}

async function handleSubmit() {
  error.value = null
  if (!name.value.trim()) {
    error.value = 'Franchise name is required'
    return
  }
  submitting.value = true
  try {
    emit('create', name.value.trim(), GAME_VERSION, selectedPath.value, selectedLeagueGUID.value)
  } finally {
    submitting.value = false
  }
}

function displayLabel(c: main.SaveFileCandidateDTO): string {
  if (c.isFranchise && c.leagueName) {
    const team = c.playerTeamName ? ` · ${c.playerTeamName}` : ''
    const seasons = c.numSeasons > 0 ? ` · ${c.numSeasons} season${c.numSeasons === 1 ? '' : 's'}` : ''
    return `${c.leagueName}${team}${seasons}`
  }
  if (c.leagueName) return c.leagueName
  // Fall back to the file name
  return c.path.split(/[\\/]/).pop() ?? c.path
}
</script>

<template>
  <div class="franchise-create">
    <h2>New Franchise</h2>

    <!-- Name -->
    <div class="field">
      <label for="franchise-name">Name</label>
      <input
        id="franchise-name"
        v-model="name"
        type="text"
        placeholder="e.g. Super Mega League"
        autocomplete="off"
        @keyup.enter="handleSubmit"
      />
    </div>

    <!-- Save file -->
    <div class="field">
      <label>Save File</label>

      <div v-if="loadingCandidates" class="hint-text">Scanning for save files…</div>

      <div v-else>
        <!-- Discovered candidates -->
        <div v-if="candidates.length > 0" class="candidates">
          <label
            v-for="c in candidates"
            :key="c.path"
            class="candidate-option"
            :class="{ selected: c.path === selectedPath }"
          >
            <input
              v-model="selectedPath"
              type="radio"
              :value="c.path"
              @change="selectCandidate(c.path)"
            />
            <div class="candidate-body">
              <span class="candidate-label">{{ displayLabel(c) }}</span>
              <span v-if="c.isFranchise" class="badge">Franchise</span>
              <span v-else-if="c.numSeasons > 0" class="badge badge-season">Season mode</span>
              <span class="candidate-path">{{ c.path }}</span>
            </div>
          </label>
        </div>

        <p v-else class="hint-text">No save files found in the default location.</p>

        <AppButton variant="ghost" size="sm" class="browse-btn" @click="handleBrowse">
          Browse for file…
        </AppButton>

        <!-- Probe result for selected file -->
        <div v-if="probing" class="hint-text probe-status">Reading save file…</div>
        <div v-else-if="selectedPath && probeResult" class="probe-result">
          <span class="probe-icon">✓</span>
          <span>
            <strong>{{ probeResult.leagueName || 'Save file found' }}</strong>
            <template v-if="probeResult.playerTeamName"> · {{ probeResult.playerTeamName }}</template>
            <template v-if="probeResult.numSeasons > 0">
              · {{ probeResult.numSeasons }} season{{ probeResult.numSeasons === 1 ? '' : 's' }}
            </template>
          </span>
        </div>
        <div v-else-if="selectedPath && !probing" class="hint-text probe-status">
          Selected: {{ selectedPath.split(/[\\/]/).pop() }}
        </div>
      </div>
    </div>

    <p v-if="error" class="error-text">{{ error }}</p>

    <div class="actions">
      <AppButton variant="secondary" @click="emit('cancel')">Cancel</AppButton>
      <AppButton variant="primary" :disabled="submitting" @click="handleSubmit">
        Create Franchise
      </AppButton>
    </div>
  </div>
</template>

<style scoped>
.franchise-create {
  max-width: 520px;
  margin: 0 auto;
}

h2 {
  font-size: 1.4rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
  color: var(--color-text-primary);
}

.field {
  margin-bottom: 1.25rem;
}

label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 0.4rem;
}

input[type='text'] {
  width: 100%;
  padding: 0.5rem 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
  outline: none;
  box-sizing: border-box;
}

input[type='text']:focus {
  border-color: var(--color-accent);
}

/* Candidate list */
.candidates {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin-bottom: 0.75rem;
}

.candidate-option {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  padding: 0.625rem 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.1s;
}

.candidate-option:hover {
  border-color: var(--color-accent);
}

.candidate-option.selected {
  border-color: var(--color-accent);
  background: color-mix(in srgb, var(--color-accent) 8%, var(--color-surface-2));
}

.candidate-option input[type='radio'] {
  margin-top: 0.125rem;
  flex-shrink: 0;
}

.candidate-body {
  display: flex;
  flex-direction: column;
  gap: 0.1875rem;
  min-width: 0;
}

.candidate-label {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.candidate-path {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  font-family: var(--font-mono);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.badge {
  display: inline-block;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
  background: color-mix(in srgb, var(--color-accent) 20%, transparent);
  color: var(--color-accent);
  width: fit-content;
}

.badge-season {
  background: color-mix(in srgb, var(--color-text-secondary) 15%, transparent);
  color: var(--color-text-secondary);
}

.browse-btn {
  margin-top: 0.25rem;
}

/* Probe result */
.probe-result {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
  padding: 0.5rem 0.75rem;
  background: color-mix(in srgb, var(--color-accent) 8%, var(--color-surface-2));
  border: 1px solid color-mix(in srgb, var(--color-accent) 30%, transparent);
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--color-text-primary);
}

.probe-icon {
  color: var(--color-accent);
  font-weight: 700;
}

.probe-status {
  margin-top: 0.5rem;
}

.hint-text {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.error-text {
  color: var(--color-error);
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
}
</style>
