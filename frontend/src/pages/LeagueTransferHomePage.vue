<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import RadioButton from 'primevue/radiobutton'
import TabPanel from 'primevue/tabpanel'
import TabView from 'primevue/tabview'
import { useToast } from 'primevue/usetoast'
import { onMounted, ref } from 'vue'
import {
  BrowseLeagueImportZip,
  ConfirmLeagueImport,
  DiscoverLeagues,
  ExportLeague,
  PreviewLeagueImport,
} from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'

const toast = useToast()
const { set: setCrumbs } = useBreadcrumbs()

onMounted(() => {
  setCrumbs([{ label: 'League Transfer' }])
  loadLeagues()
})

// ── Export tab ───────────────────────────────────────────────────────────────

const leagues = ref<main.LeagueOverviewDTO[]>([])
const loadingLeagues = ref(false)
const leaguesError = ref<string | null>(null)
const exportingGUID = ref<string | null>(null)

function divisionCount(league: main.LeagueOverviewDTO): number {
  return league.conferences.reduce((sum, c) => sum + c.divisions.length, 0)
}

function teamCount(league: main.LeagueOverviewDTO): number {
  return league.conferences.reduce((sum, c) => sum + c.divisions.reduce((dSum, d) => dSum + d.teams.length, 0), 0)
}

async function loadLeagues() {
  loadingLeagues.value = true
  leaguesError.value = null
  try {
    leagues.value = (await DiscoverLeagues()) ?? []
  } catch (e) {
    leaguesError.value = String(e)
  } finally {
    loadingLeagues.value = false
  }
}

async function exportLeague(league: main.LeagueOverviewDTO) {
  exportingGUID.value = league.guid
  try {
    const outputPath = await ExportLeague(league.guid, league.sourcePath)
    toast.add({ severity: 'success', summary: `Exported "${league.name}"`, detail: outputPath, life: 5000 })
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Export failed', detail: String(e), life: 5000 })
  } finally {
    exportingGUID.value = null
  }
}

// ── Import tab ───────────────────────────────────────────────────────────────

type ImportStep = 'choose' | 'preview' | 'importing'

const importStep = ref<ImportStep>('choose')
const importZipPath = ref<string | null>(null)
const importPreview = ref<main.LeagueImportPreviewDTO | null>(null)
const importError = ref<string | null>(null)
const selectedTargetDir = ref<string | null>(null)
const loadingPreview = ref(false)
const confirmingImport = ref(false)

async function chooseImportFile() {
  importError.value = null
  const path = await BrowseLeagueImportZip()
  if (!path) return

  importZipPath.value = path
  loadingPreview.value = true
  try {
    importPreview.value = await PreviewLeagueImport(path)
    selectedTargetDir.value = importPreview.value.targets.find((t) => !t.alreadyRegistered)?.dirPath ?? null
    importStep.value = 'preview'
  } catch (e) {
    importError.value = String(e)
  } finally {
    loadingPreview.value = false
  }
}

async function confirmImport() {
  if (!importZipPath.value || !selectedTargetDir.value) return
  confirmingImport.value = true
  importError.value = null
  try {
    await ConfirmLeagueImport(importZipPath.value, selectedTargetDir.value)
    toast.add({ severity: 'success', summary: 'League imported', life: 4000 })
    resetImport()
  } catch (e) {
    importError.value = String(e)
  } finally {
    confirmingImport.value = false
  }
}

function resetImport() {
  importStep.value = 'choose'
  importZipPath.value = null
  importPreview.value = null
  selectedTargetDir.value = null
  importError.value = null
}
</script>

<template>
  <div class="lt-page">
    <TabView>
      <TabPanel value="0" header="Export">
        <div class="tab-content">
          <p class="tab-desc">
            Export a league from any save file on this machine so you can share it with someone else.
          </p>

          <p v-if="leaguesError" class="error-text">{{ leaguesError }}</p>

          <div v-if="loadingLeagues" class="progress-state">
            <LoadingSpinner />
            <p>Scanning save files…</p>
          </div>

          <div v-else-if="leagues.length === 0" class="empty-state">
            <p>No leagues found on this machine.</p>
          </div>

          <DataTable v-else :value="leagues" data-key="guid">
            <Column field="name" header="League" style="min-width: 200px" />
            <Column header="Conferences" style="min-width: 110px">
              <template #body="{ data }">{{ data.conferences.length }}</template>
            </Column>
            <Column header="Divisions" style="min-width: 100px">
              <template #body="{ data }">{{ divisionCount(data) }}</template>
            </Column>
            <Column header="Teams" style="min-width: 90px">
              <template #body="{ data }">{{ teamCount(data) }}</template>
            </Column>
            <Column header="" style="min-width: 120px">
              <template #body="{ data }">
                <AppButton
                  variant="secondary"
                  size="sm"
                  :disabled="exportingGUID === data.guid"
                  @click="exportLeague(data)"
                >
                  {{ exportingGUID === data.guid ? 'Exporting…' : 'Export' }}
                </AppButton>
              </template>
            </Column>
          </DataTable>
        </div>
      </TabPanel>

      <TabPanel value="1" header="Import">
        <div class="tab-content">
          <p class="tab-desc">
            Import a league someone exported for you. smb-tools does not scan league files for malware —
            only import files from people you trust.
          </p>

          <p v-if="importError" class="error-text">{{ importError }}</p>

          <!-- Choose step -->
          <div v-if="importStep === 'choose'" class="import-choose">
            <AppButton variant="primary" :disabled="loadingPreview" @click="chooseImportFile">
              {{ loadingPreview ? 'Reading…' : 'Choose League File…' }}
            </AppButton>
          </div>

          <!-- Preview step -->
          <div v-else-if="importStep === 'preview' && importPreview" class="import-preview">
            <div class="preview-summary">
              <h3>{{ importPreview.overview.name }}</h3>
              <span class="preview-stats">
                {{ importPreview.overview.conferences.length }} conferences ·
                {{ divisionCount(importPreview.overview) }} divisions ·
                {{ teamCount(importPreview.overview) }} teams
              </span>
            </div>

            <div class="target-picker">
              <span class="target-label">Import into:</span>
              <div v-if="importPreview.targets.length === 0" class="empty-state">
                <p>No SMB4 save directories were found on this machine.</p>
              </div>
              <label
                v-for="target in importPreview.targets"
                :key="target.dirPath"
                class="target-option"
                :class="{ 'target-option--disabled': target.alreadyRegistered }"
              >
                <RadioButton
                  v-model="selectedTargetDir"
                  :value="target.dirPath"
                  :disabled="target.alreadyRegistered"
                  name="target-dir"
                />
                <span class="target-path">{{ target.dirPath }}</span>
                <span v-if="target.alreadyRegistered" class="target-warning">Already registered</span>
              </label>
            </div>

            <p class="reminder-text">Make sure Super Mega Baseball 4 is closed before importing.</p>

            <div class="step-actions">
              <AppButton variant="secondary" :disabled="confirmingImport" @click="resetImport">Cancel</AppButton>
              <AppButton
                variant="primary"
                :disabled="confirmingImport || !selectedTargetDir"
                @click="confirmImport"
              >
                {{ confirmingImport ? 'Importing…' : 'Import League' }}
              </AppButton>
            </div>
          </div>
        </div>
      </TabPanel>
    </TabView>
  </div>
</template>

<style scoped>
.lt-page {
  padding: 1.5rem 2rem;
  max-width: 960px;
  width: 100%;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-top: 0.5rem;
}

.tab-desc {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  margin: 0;
  line-height: 1.6;
}

.error-text {
  color: var(--color-error);
  font-size: 0.875rem;
}

.empty-state {
  text-align: center;
  padding: 2rem 0;
  color: var(--color-text-secondary);
}

.progress-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem 0;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.import-choose {
  display: flex;
  padding: 1rem 0;
}

.import-preview {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.preview-summary h3 {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.25rem;
}

.preview-stats {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.target-picker {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.target-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.target-option {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.625rem 0.875rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
}

.target-option--disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.target-path {
  flex: 1;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.target-warning {
  font-size: 0.75rem;
  color: var(--color-error);
  flex-shrink: 0;
}

.reminder-text {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  margin: 0;
}

.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
