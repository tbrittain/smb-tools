<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'
import AppButton from './AppButton.vue'
import LoadingSpinner from './LoadingSpinner.vue'

const props = defineProps<{
  candidates: main.SaveFileCandidateDTO[]
  loading?: boolean
  scanning?: boolean
  browsing?: boolean
  error?: string | null
  selectedPath?: string
  usedSourceLabels?: Record<string, string>
}>()

const emit = defineEmits<{
  change: [path: string, leagueGUID: string, probe: main.SaveFileCandidateDTO]
  scanDirectory: []
  browseFile: []
}>()

function select(c: main.SaveFileCandidateDTO) {
  emit('change', c.path, c.leagueGUID, c)
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

function modeLabel(c: main.SaveFileCandidateDTO): string | null {
  return c.mode === 'season' ? 'Season Mode' : null
}
</script>

<template>
  <div class="save-file-picker">
    <div v-if="props.loading" class="picker-loading">
      <LoadingSpinner size="sm" />
      <span>Scanning for save files…</span>
    </div>

    <template v-else>
      <div v-if="props.candidates.length === 0" class="no-candidates">
        No Franchise or Season mode save files found in the default SMB4 location.
        Elimination saves are not supported.
      </div>

      <!-- Scrollable candidate list -->
      <ul v-else class="candidate-list">
        <li
          v-for="c in props.candidates"
          :key="c.path"
          class="candidate-card"
          :class="{ 'is-selected': c.path === props.selectedPath }"
          @click="select(c)"
        >
          <div class="card-select">
            <input
              type="radio"
              :checked="c.path === props.selectedPath"
              :name="`save-file-picker-${c.path}`"
              @change="select(c)"
              @click.stop
            />
          </div>

          <div class="card-body">
            <span class="league-name">
              {{ primaryLabel(c) }}
              <span v-if="modeLabel(c)" class="mode-chip">{{ modeLabel(c) }}</span>
            </span>

            <span v-if="c.playerTeamName" class="detail-line">
              Playing as: <strong>{{ c.playerTeamName }}</strong>
            </span>

            <span v-if="seasonLine(c)" class="detail-line">{{ seasonLine(c) }}</span>

            <span v-if="props.usedSourceLabels?.[c.path]" class="detail-line used-label">
              {{ props.usedSourceLabels[c.path] }}
            </span>

            <span class="file-path">{{ fileName(c) }}</span>
          </div>

          <div v-if="c.path === props.selectedPath" class="card-check" aria-hidden="true">✓</div>
        </li>
      </ul>

      <p v-if="props.error" class="error-text">{{ props.error }}</p>

      <!-- Primary fallback: scan a folder -->
      <div class="action-group">
        <AppButton
          variant="ghost"
          size="sm"
          class="action-btn"
          :disabled="props.scanning"
          @click="emit('scanDirectory')"
        >
          {{ props.scanning ? 'Scanning…' : 'Scan a folder…' }}
        </AppButton>
        <p class="action-hint">
          Points to a folder and identifies all Franchise and Season mode saves inside (the
          easiest way to find your saves if they are not in the default location)
        </p>
      </div>

      <!-- Advanced fallback: browse for a single file -->
      <div class="action-group">
        <div class="advanced-row">
          <span class="advanced-chip">Advanced</span>
          <AppButton
            variant="ghost"
            size="sm"
            class="action-btn"
            :disabled="props.browsing"
            @click="emit('browseFile')"
          >
            {{ props.browsing ? 'Opening…' : 'Browse for file…' }}
          </AppButton>
        </div>
        <p class="action-hint">
          Select a single save file directly if you already know which one to use.
        </p>
      </div>
    </template>
  </div>
</template>

<style scoped>
.save-file-picker {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
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

.mode-chip {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--color-text-secondary);
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 0.05rem 0.375rem;
  margin-left: 0.4rem;
  vertical-align: middle;
}

.used-label {
  color: var(--color-accent);
  font-size: 0.75rem;
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

.error-text {
  font-size: 0.8125rem;
  color: var(--color-error);
  margin: 0;
}

.action-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.action-hint {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  margin: 0;
  line-height: 1.4;
  padding-left: 0.125rem;
}

.advanced-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.advanced-chip {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-secondary);
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 0.1rem 0.375rem;
  flex-shrink: 0;
}

.action-btn {
  align-self: flex-start;
}
</style>
