<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { BrowseLegacyDB, DetectLegacyDB, ListLegacyFranchises, MigrateLegacyFranchise } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import AppHelpButton from '../components/AppHelpButton.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

type Step = 'source' | 'select' | 'names' | 'confirm' | 'progress' | 'done'

const router = useRouter()

const step = ref<Step>('source')
const detectedPath = ref<string | null>(null)
const chosenPath = ref('')
const error = ref<string | null>(null)
const loading = ref(false)

const franchises = ref<main.LegacyFranchiseDTO[]>([])
const selected = ref<Set<number>>(new Set())
const customNames = ref<Record<number, string>>({})

type MigrationEntry = { legacyID: number; legacyName: string; newName: string; gameVersion: string }
const pendingMigrations = ref<MigrationEntry[]>([])

type ResultEntry = { result: main.MigrateLegacyResult; error: string | null }
const results = ref<ResultEntry[]>([])

onMounted(async () => {
  const p = await DetectLegacyDB()
  if (p) detectedPath.value = p
})

async function useDetectedPath() {
  if (!detectedPath.value) return
  chosenPath.value = detectedPath.value
  await loadFranchises()
}

async function browseForFile() {
  error.value = null
  const p = await BrowseLegacyDB()
  if (!p) return
  chosenPath.value = p
  await loadFranchises()
}

async function loadFranchises() {
  loading.value = true
  error.value = null
  try {
    const list = await ListLegacyFranchises(chosenPath.value)
    franchises.value = list ?? []
    selected.value = new Set()
    customNames.value = {}
    for (const f of franchises.value) customNames.value[f.id] = f.name
    step.value = 'select'
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function toggleFranchise(id: number) {
  if (selected.value.has(id)) selected.value.delete(id)
  else selected.value.add(id)
}

function proceedToNames() {
  if (selected.value.size === 0) return
  step.value = 'names'
}

function proceedToConfirm() {
  pendingMigrations.value = franchises.value
    .filter((f) => selected.value.has(f.id))
    .map((f) => ({
      legacyID: f.id,
      legacyName: f.name,
      newName: customNames.value[f.id] ?? f.name,
      gameVersion: f.isSmb3 ? 'smb3' : 'smb4',
    }))
  step.value = 'confirm'
}

async function runMigrations() {
  step.value = 'progress'
  results.value = []
  for (const entry of pendingMigrations.value) {
    try {
      const r = await MigrateLegacyFranchise(chosenPath.value, entry.legacyID, entry.newName, entry.gameVersion)
      results.value.push({ result: r, error: null })
    } catch (e) {
      results.value.push({
        result: {
          franchiseId: '',
          franchiseName: entry.newName,
          seasonsMigrated: 0,
          teamsMigrated: 0,
          playersMigrated: 0,
          awardsMigrated: 0,
          logosSkipped: 0,
        },
        error: String(e),
      })
    }
  }
  step.value = 'done'
}

function backToSelector() {
  router.push('/')
}
</script>

<template>
  <div class="migration-page">
    <!-- Header -->
    <div class="page-header">
      <button class="back-btn" @click="backToSelector">← Back</button>
      <div class="heading-row">
        <h1>Import from SmbExplorerCompanion</h1>
        <AppHelpButton docs-path="legacy-migration.html#starting-the-import" />
      </div>
    </div>

    <!-- ── Source step ─────────────────────────────────────────────── -->
    <div v-if="step === 'source'" class="step">
      <p class="step-desc">
        Locate your <code>SmbExplorerCompanion.db</code> file to bring your franchise history into smb-tools.
      </p>

      <p v-if="error" class="error-text">{{ error }}</p>

      <!-- Detected path banner -->
      <div v-if="detectedPath" class="detected-card">
        <div class="detected-info">
          <span class="detected-label">Found at default location</span>
          <code class="detected-path">{{ detectedPath }}</code>
        </div>
        <AppButton variant="primary" size="sm" :disabled="loading" @click="useDetectedPath">
          {{ loading ? 'Loading…' : 'Use this file' }}
        </AppButton>
      </div>

      <div class="or-divider" v-if="detectedPath">
        <span>or choose a different file</span>
      </div>

      <AppButton
        variant="secondary"
        :disabled="loading"
        @click="browseForFile"
      >
        {{ detectedPath ? 'Browse for file…' : 'Browse for SmbExplorerCompanion.db…' }}
      </AppButton>
    </div>

    <!-- ── Franchise selection step ───────────────────────────────── -->
    <div v-else-if="step === 'select'" class="step">
      <p class="step-desc">Select the franchises to import. Each becomes a separate franchise in smb-tools.</p>

      <ul class="franchise-list">
        <li
          v-for="f in franchises"
          :key="f.id"
          class="franchise-item"
          :class="{ 'franchise-item--selected': selected.has(f.id) }"
          @click="toggleFranchise(f.id)"
        >
          <input type="checkbox" :checked="selected.has(f.id)" @click.stop="toggleFranchise(f.id)" />
          <span class="franchise-name">{{ f.name }}</span>
          <span class="version-badge">{{ f.isSmb3 ? 'SMB3' : 'SMB4' }}</span>
        </li>
      </ul>

      <div class="step-actions">
        <AppButton variant="secondary" @click="step = 'source'">Back</AppButton>
        <AppButton variant="primary" :disabled="selected.size === 0" @click="proceedToNames">
          Next
        </AppButton>
      </div>
    </div>

    <!-- ── Name step ──────────────────────────────────────────────── -->
    <div v-else-if="step === 'names'" class="step">
      <p class="step-desc">Confirm or rename each franchise before importing.</p>

      <div class="names-form">
        <div
          v-for="f in franchises.filter((x) => selected.has(x.id))"
          :key="f.id"
          class="field"
        >
          <label :for="`name-${f.id}`">{{ f.name }}</label>
          <input
            :id="`name-${f.id}`"
            v-model="customNames[f.id]"
            type="text"
            autocomplete="off"
          />
        </div>
      </div>

      <div class="step-actions">
        <AppButton variant="secondary" @click="step = 'select'">Back</AppButton>
        <AppButton variant="primary" @click="proceedToConfirm">Review & Import</AppButton>
      </div>
    </div>

    <!-- ── Confirm step ───────────────────────────────────────────── -->
    <div v-else-if="step === 'confirm'" class="step">
      <p class="step-desc">The following will be imported as new franchises. Your original data is not modified.</p>

      <ul class="confirm-list">
        <li v-for="entry in pendingMigrations" :key="entry.legacyID" class="confirm-item">
          <span class="confirm-name">{{ entry.newName }}</span>
          <span class="version-badge">{{ entry.gameVersion.toUpperCase() }}</span>
        </li>
      </ul>

      <div class="step-actions">
        <AppButton variant="secondary" @click="step = 'names'">Back</AppButton>
        <AppButton variant="primary" @click="runMigrations">Import</AppButton>
      </div>
    </div>

    <!-- ── Progress step ──────────────────────────────────────────── -->
    <div v-else-if="step === 'progress'" class="progress-state">
      <LoadingSpinner />
      <p>Importing franchises…</p>
    </div>

    <!-- ── Done step ──────────────────────────────────────────────── -->
    <div v-else-if="step === 'done'" class="step">
      <h2>Import complete</h2>

      <ul class="results-list">
        <li v-for="entry in results" :key="entry.result.franchiseName" class="result-item">
          <template v-if="entry.error">
            <span class="result-name">{{ entry.result.franchiseName }}</span>
            <span class="error-text">{{ entry.error }}</span>
          </template>
          <template v-else>
            <div class="result-header">
              <span class="result-name">{{ entry.result.franchiseName }}</span>
              <span v-if="entry.result.logosSkipped > 0" class="logos-note">
                {{ entry.result.logosSkipped }} logo(s) skipped
              </span>
            </div>
            <div class="result-stats">
              <span>{{ entry.result.seasonsMigrated }} seasons</span>
              <span>{{ entry.result.teamsMigrated }} teams</span>
              <span>{{ entry.result.playersMigrated }} players</span>
              <span>{{ entry.result.awardsMigrated }} awards</span>
            </div>
          </template>
        </li>
      </ul>

      <div class="step-actions">
        <AppButton variant="primary" @click="backToSelector">Go to franchise selector</AppButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.migration-page {
  width: 100%;
  max-width: 520px;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.step {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.page-header {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.back-btn {
  background: none;
  border: none;
  padding: 0;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  font-family: inherit;
  cursor: pointer;
  align-self: flex-start;
}

.back-btn:hover {
  color: var(--color-text-primary);
}

.heading-row {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

h1 {
  font-size: 1.4rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

h2 {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.step-desc {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  margin: 0;
  line-height: 1.6;
}

.error-text {
  color: var(--color-error);
  font-size: 0.875rem;
}

/* Source step */
.detected-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.875rem 1rem;
  background: var(--color-surface-2);
  border: 1px solid color-mix(in srgb, var(--color-accent) 40%, transparent);
  border-radius: 8px;
}

.detected-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.detected-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.detected-path {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.or-divider {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.or-divider::before,
.or-divider::after {
  content: '';
  flex: 1;
  border-top: 1px solid var(--color-border);
}

/* Franchise selection */
.franchise-list,
.confirm-list,
.results-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.franchise-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.15s;
  user-select: none;
}

.franchise-item:hover {
  border-color: var(--color-accent);
}

.franchise-item--selected {
  border-color: var(--color-accent);
  background: color-mix(in srgb, var(--color-accent) 8%, var(--color-surface-2));
}

.franchise-item input[type='checkbox'] {
  accent-color: var(--color-accent);
  width: 15px;
  height: 15px;
  flex-shrink: 0;
  cursor: pointer;
}

.franchise-name {
  flex: 1;
  font-size: 0.9375rem;
  color: var(--color-text-primary);
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
  flex-shrink: 0;
}

/* Name step */
.names-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

input[type='text'] {
  width: 100%;
  padding: 0.5rem 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
}

input[type='text']:focus {
  border-color: var(--color-accent);
}

/* Confirm step */
.confirm-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.confirm-name {
  flex: 1;
  font-size: 0.9375rem;
  color: var(--color-text-primary);
}

.notice-text {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  margin: 0;
}

/* Progress */
.progress-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem 0;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

/* Results */
.result-item {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  padding: 0.875rem 1rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.result-name {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.result-stats {
  display: flex;
  gap: 1rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.logos-note {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

/* Shared */
.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

code {
  font-family: var(--font-mono);
  font-size: 0.9em;
}
</style>
