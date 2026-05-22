<script lang="ts" setup>
import Button from 'primevue/button'
import Card from 'primevue/card'
import Checkbox from 'primevue/checkbox'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import ProgressSpinner from 'primevue/progressspinner'
import Tag from 'primevue/tag'
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { BrowseLegacyDB, DetectLegacyDB, ListLegacyFranchises, MigrateLegacyFranchise } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'

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

type MigrationEntry = {
  legacyID: number
  legacyName: string
  newName: string
  gameVersion: string
}
const pendingMigrations = ref<MigrationEntry[]>([])

type ResultEntry = {
  result: main.MigrateLegacyResult
  error: string | null
}
const results = ref<ResultEntry[]>([])

onMounted(async () => {
  const p = await DetectLegacyDB()
  if (p) {
    detectedPath.value = p
  }
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
    for (const f of franchises.value) {
      customNames.value[f.id] = f.name
    }
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

function openFranchise(_franchiseId: string) {
  router.push('/')
}
</script>

<template>
  <div class="migration-page">
    <div class="migration-header">
      <Button
        icon="pi pi-arrow-left"
        label="Back"
        text
        size="small"
        @click="router.push('/')"
      />
      <h1 class="migration-title">Import from SmbExplorerCompanion</h1>
    </div>

    <!-- ── Source step ──────────────────────────────────────────────── -->
    <div v-if="step === 'source'" class="step-container">
      <p class="step-description">
        Locate your <code>SmbExplorerCompanion.db</code> file to import its
        franchise history into smb-tools.
      </p>

      <Message v-if="error" severity="error" class="mb-4">{{ error }}</Message>

      <Card v-if="detectedPath" class="detected-card mb-4">
        <template #content>
          <div class="detected-content">
            <div>
              <div class="detected-label">Database found at default location</div>
              <code class="detected-path">{{ detectedPath }}</code>
            </div>
            <Button
              label="Use this file"
              :loading="loading"
              @click="useDetectedPath"
            />
          </div>
        </template>
      </Card>

      <div class="divider-row" v-if="detectedPath">
        <span class="divider-text">or choose a different file</span>
      </div>

      <Button
        :label="detectedPath ? 'Browse for file…' : 'Browse for SmbExplorerCompanion.db…'"
        icon="pi pi-folder-open"
        :outlined="!!detectedPath"
        :loading="loading"
        @click="browseForFile"
      />
    </div>

    <!-- ── Franchise selection step ─────────────────────────────────── -->
    <div v-else-if="step === 'select'" class="step-container">
      <p class="step-description">
        Select the franchises to import. Each will become a separate franchise in
        smb-tools.
      </p>
      <div class="franchise-list">
        <div
          v-for="f in franchises"
          :key="f.id"
          class="franchise-row"
          @click="toggleFranchise(f.id)"
        >
          <Checkbox :modelValue="selected.has(f.id)" :binary="true" />
          <span class="franchise-name">{{ f.name }}</span>
          <Tag :value="f.isSmb3 ? 'SMB3' : 'SMB4'" :severity="f.isSmb3 ? 'secondary' : 'info'" />
        </div>
      </div>
      <div class="step-actions">
        <Button label="Back" text @click="step = 'source'" />
        <Button
          label="Next"
          :disabled="selected.size === 0"
          @click="proceedToNames"
        />
      </div>
    </div>

    <!-- ── Name customisation step ──────────────────────────────────── -->
    <div v-else-if="step === 'names'" class="step-container">
      <p class="step-description">
        Confirm the name for each franchise. You can rename it before importing.
      </p>
      <div class="names-list">
        <div
          v-for="f in franchises.filter((x) => selected.has(x.id))"
          :key="f.id"
          class="name-row"
        >
          <label :for="`name-${f.id}`" class="name-label">{{ f.name }}</label>
          <InputText
            :id="`name-${f.id}`"
            v-model="customNames[f.id]"
            class="name-input"
          />
        </div>
      </div>
      <div class="step-actions">
        <Button label="Back" text @click="step = 'select'" />
        <Button label="Review & Import" @click="proceedToConfirm" />
      </div>
    </div>

    <!-- ── Confirm step ──────────────────────────────────────────────── -->
    <div v-else-if="step === 'confirm'" class="step-container">
      <p class="step-description">
        The following franchises will be imported. This creates new franchise
        records — your original data is not modified.
      </p>
      <div class="confirm-list">
        <Card
          v-for="entry in pendingMigrations"
          :key="entry.legacyID"
          class="confirm-card"
        >
          <template #content>
            <div class="confirm-row">
              <strong>{{ entry.newName }}</strong>
              <Tag :value="entry.gameVersion.toUpperCase()" severity="info" />
            </div>
          </template>
        </Card>
      </div>
      <Message severity="warn" class="mt-4">
        Team logos are not migrated (logo storage is not yet supported in this
        version of smb-tools).
      </Message>
      <div class="step-actions">
        <Button label="Back" text @click="step = 'names'" />
        <Button label="Import" icon="pi pi-upload" @click="runMigrations" />
      </div>
    </div>

    <!-- ── Progress step ─────────────────────────────────────────────── -->
    <div v-else-if="step === 'progress'" class="step-container centered">
      <ProgressSpinner />
      <p>Importing franchises…</p>
    </div>

    <!-- ── Done step ─────────────────────────────────────────────────── -->
    <div v-else-if="step === 'done'" class="step-container">
      <h2>Import complete</h2>
      <div class="results-list">
        <Card
          v-for="entry in results"
          :key="entry.result.franchiseName"
          class="result-card"
        >
          <template #content>
            <div v-if="entry.error" class="result-error">
              <strong>{{ entry.result.franchiseName }}</strong>
              <Message severity="error">{{ entry.error }}</Message>
            </div>
            <div v-else class="result-success">
              <div class="result-name">{{ entry.result.franchiseName }}</div>
              <div class="result-stats">
                <span>{{ entry.result.seasonsMigrated }} seasons</span>
                <span>{{ entry.result.teamsMigrated }} teams</span>
                <span>{{ entry.result.playersMigrated }} players</span>
                <span>{{ entry.result.awardsMigrated }} awards</span>
                <span v-if="entry.result.logosSkipped > 0" class="logos-note">
                  {{ entry.result.logosSkipped }} logo(s) skipped
                </span>
              </div>
              <Button
                label="Open franchise"
                size="small"
                text
                @click="openFranchise(entry.result.franchiseId)"
              />
            </div>
          </template>
        </Card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.migration-page {
  max-width: 680px;
  margin: 0 auto;
  padding: 2rem 1.5rem;
}

.migration-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.migration-title {
  font-size: 1.4rem;
  font-weight: 600;
  margin: 0;
}

.step-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.step-description {
  color: var(--p-text-muted-color);
  margin: 0;
}

.detected-card {
  border: 1px solid var(--p-primary-color);
}

.detected-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.detected-label {
  font-weight: 500;
  margin-bottom: 0.25rem;
}

.detected-path {
  font-size: 0.8rem;
  word-break: break-all;
  color: var(--p-text-muted-color);
}

.divider-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--p-text-muted-color);
  font-size: 0.85rem;
}

.franchise-list,
.names-list,
.confirm-list,
.results-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.franchise-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid var(--p-surface-border);
}

.franchise-row:hover {
  background: var(--p-surface-hover);
}

.franchise-name {
  flex: 1;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.name-label {
  min-width: 180px;
  color: var(--p-text-muted-color);
  font-size: 0.9rem;
}

.name-input {
  flex: 1;
}

.confirm-card,
.result-card {
  border: 1px solid var(--p-surface-border);
}

.confirm-row,
.result-success {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.result-name {
  font-weight: 600;
  flex: 1;
}

.result-stats {
  display: flex;
  gap: 0.75rem;
  font-size: 0.85rem;
  color: var(--p-text-muted-color);
  flex-wrap: wrap;
}

.logos-note {
  color: var(--p-yellow-500);
}

.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.centered {
  align-items: center;
  justify-content: center;
  padding: 3rem;
}

code {
  font-family: monospace;
}

.mb-4 {
  margin-bottom: 1rem;
}

.mt-4 {
  margin-top: 1rem;
}
</style>
