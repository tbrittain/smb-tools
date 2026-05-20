<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { BrowseSaveFile, GetSaveFileCandidates, ProbeLeagues } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from './AppButton.vue'
import LoadingSpinner from './LoadingSpinner.vue'

defineProps<{
  selectedPath?: string
}>()

const emit = defineEmits<{
  // Fired immediately when the user clicks a candidate or confirms a browse selection.
  // probe is the full metadata object so callers can display franchise info.
  change: [path: string, leagueGUID: string, probe: main.SaveFileCandidateDTO]
}>()

const candidates = ref<main.SaveFileCandidateDTO[]>([])
const loading = ref(false)
const browsing = ref(false)
const error = ref<string | null>(null)

onMounted(load)

async function load() {
  loading.value = true
  error.value = null
  try {
    const all = await GetSaveFileCandidates()
    // SMB3 is deferred — only surface SMB4 save files
    candidates.value = all.filter((c) => c.gameVersion === 'smb4')
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

async function handleBrowse() {
  browsing.value = true
  error.value = null
  try {
    const path = await BrowseSaveFile()
    if (!path) return
    // Probe the file for franchise metadata so the card shows real info
    const probed = await ProbeLeagues(path)
    const match = probed[0]
    const candidate: main.SaveFileCandidateDTO = match
      ? { ...match, path, gameVersion: 'smb4' }
      : {
          path,
          gameVersion: 'smb4',
          leagueName: '',
          numSeasons: 0,
          isFranchise: false,
          playerTeamName: '',
          leagueGUID: '',
        }
    if (!candidates.value.find((c) => c.path === path)) {
      candidates.value = [...candidates.value, candidate]
    }
    select(candidate)
  } catch (e) {
    error.value = String(e)
  } finally {
    browsing.value = false
  }
}

function select(c: main.SaveFileCandidateDTO) {
  emit('change', c.path, c.leagueGUID, c)
}

function modeLabel(c: main.SaveFileCandidateDTO): string {
  if (c.isFranchise) return 'Franchise'
  if (c.numSeasons > 0) return 'Season'
  return 'No games yet'
}

function modeCssClass(c: main.SaveFileCandidateDTO): string {
  if (c.isFranchise) return 'mode-franchise'
  if (c.numSeasons > 0) return 'mode-season'
  return 'mode-empty'
}

function primaryLabel(c: main.SaveFileCandidateDTO): string {
  return c.leagueName || fileName(c)
}

function fileName(c: main.SaveFileCandidateDTO): string {
  return c.path.split(/[\\/]/).pop() ?? c.path
}

function seasonLine(c: main.SaveFileCandidateDTO): string | null {
  if (c.numSeasons === 0) return null
  return `${c.numSeasons} season${c.numSeasons === 1 ? '' : 's'} played`
}
</script>

<template>
  <div class="save-file-picker">
    <div v-if="loading" class="picker-loading">
      <LoadingSpinner size="sm" />
      <span>Scanning for save files…</span>
    </div>

    <template v-else>
      <div v-if="candidates.length === 0" class="no-candidates">
        No save files found in the default SMB4 location.
      </div>

      <!-- Scrollable candidate list -->
      <ul v-else class="candidate-list">
        <li
          v-for="c in candidates"
          :key="c.path"
          class="candidate-card"
          :class="{ 'is-selected': c.path === selectedPath }"
          @click="select(c)"
        >
          <div class="card-select">
            <input
              type="radio"
              :checked="c.path === selectedPath"
              :name="`save-file-picker-${c.path}`"
              @change="select(c)"
              @click.stop
            />
          </div>

          <div class="card-body">
            <!-- Mode badge -->
            <span class="mode-badge" :class="modeCssClass(c)">{{ modeLabel(c) }}</span>

            <!-- League name — primary identifier -->
            <span class="league-name">{{ primaryLabel(c) }}</span>

            <!-- Player team — only in franchise mode -->
            <span v-if="c.isFranchise && c.playerTeamName" class="detail-line">
              Playing as: <strong>{{ c.playerTeamName }}</strong>
            </span>

            <!-- Season count -->
            <span v-if="seasonLine(c)" class="detail-line">{{ seasonLine(c) }}</span>

            <!-- File name for disambiguation when multiple files share a league name -->
            <span class="file-path">{{ fileName(c) }}</span>
          </div>

          <div v-if="c.path === selectedPath" class="card-check" aria-hidden="true">✓</div>
        </li>
      </ul>

      <p v-if="error" class="error-text">{{ error }}</p>

      <AppButton
        variant="ghost"
        size="sm"
        class="browse-btn"
        :disabled="browsing"
        @click="handleBrowse"
      >
        {{ browsing ? 'Opening…' : 'Browse for file…' }}
      </AppButton>
    </template>
  </div>
</template>

<style scoped>
.save-file-picker {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.picker-loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  padding: 0.5rem 0;
}

.no-candidates {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  padding: 0.25rem 0;
}

/* Scrollable list — 3.5 cards visible before scroll */
.candidate-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  max-height: 320px;
  overflow-y: auto;
  padding-right: 2px; /* room for scrollbar */
}

.candidate-card {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  padding: 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.12s;
  user-select: none;
}

.candidate-card:hover {
  border-color: color-mix(in srgb, var(--color-accent) 60%, var(--color-border));
}

.candidate-card.is-selected {
  border-color: var(--color-accent);
  background: color-mix(in srgb, var(--color-accent) 6%, var(--color-surface-2));
}

.card-select {
  padding-top: 1px;
  flex-shrink: 0;
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.mode-badge {
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.07em;
  text-transform: uppercase;
}

.mode-franchise { color: var(--color-accent); }
.mode-season    { color: #60a5fa; }
.mode-empty     { color: var(--color-text-secondary); }

.league-name {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.detail-line {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.file-path {
  font-size: 0.75rem;
  font-family: var(--font-mono);
  color: var(--color-text-secondary);
  opacity: 0.7;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: 0.125rem;
}

.card-check {
  color: var(--color-accent);
  font-weight: 700;
  font-size: 0.875rem;
  flex-shrink: 0;
  align-self: center;
}

.browse-btn {
  align-self: flex-start;
}

.error-text {
  font-size: 0.8125rem;
  color: var(--color-error);
}
</style>
