<script lang="ts" setup>
import Accordion from 'primevue/accordion'
import AccordionContent from 'primevue/accordioncontent'
import AccordionHeader from 'primevue/accordionheader'
import AccordionPanel from 'primevue/accordionpanel'
import RadioButton from 'primevue/radiobutton'
import TabPanel from 'primevue/tabpanel'
import TabView from 'primevue/tabview'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { computed, onMounted, ref } from 'vue'
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

const toast = useToast()

onMounted(() => {
  loadLeagues()
})

// ── Export tab ───────────────────────────────────────────────────────────────

// SMB4 leaves an empty, never-played "shell" league-*.sav file behind
// alongside the league the player actually started, both sharing the same
// display name — so DiscoverLeagues' flat list is grouped by name here and
// presented as: the shell's Export button by default, with any real
// save(s) (Franchise/Season/Elimination) tucked behind an expand action.
interface LeagueGroup {
  name: string
  shell: main.LeagueOverviewDTO | null
  realSaves: main.LeagueOverviewDTO[]
  entries: main.LeagueOverviewDTO[]
}

const leagues = ref<main.LeagueOverviewDTO[]>([])
const loadingLeagues = ref(false)
const leaguesError = ref<string | null>(null)
const exportingGUID = ref<string | null>(null)

const leagueGroups = computed<LeagueGroup[]>(() => {
  const byName = new Map<string, main.LeagueOverviewDTO[]>()
  for (const league of leagues.value) {
    const entries = byName.get(league.name) ?? []
    entries.push(league)
    byName.set(league.name, entries)
  }
  return Array.from(byName.entries())
    .map(([name, entries]) => ({
      name,
      shell: entries.find((e) => e.mode === 'none') ?? null,
      realSaves: entries.filter((e) => e.mode !== 'none'),
      entries,
    }))
    .sort((a, b) => a.name.localeCompare(b.name))
})

function teamCount(league: main.LeagueOverviewDTO): number {
  return league.conferences.reduce((sum, c) => sum + c.divisions.reduce((dSum, d) => dSum + d.teams.length, 0), 0)
}

function divisionCount(league: main.LeagueOverviewDTO): number {
  return league.conferences.reduce((sum, c) => sum + c.divisions.length, 0)
}

function statsSummary(league: main.LeagueOverviewDTO): string {
  return `${league.conferences.length} conferences · ${divisionCount(league)} divisions · ${teamCount(league)} teams`
}

function modeLabel(mode: string): string {
  return mode.charAt(0).toUpperCase() + mode.slice(1)
}

function modeSeverity(mode: string): 'success' | 'info' | 'warn' | 'secondary' {
  switch (mode) {
    case 'franchise':
      return 'success'
    case 'season':
      return 'info'
    case 'elimination':
      return 'warn'
    default:
      return 'secondary'
  }
}

function exportButtonLabel(league: main.LeagueOverviewDTO): string {
  return league.mode === 'none' ? 'Export Empty League' : 'Export League with Save Game'
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
            Export a league from any save file on this machine so you can share it with someone else. You can
            export an empty league shell (just the teams/conferences/divisions setup, no games played) as well as
            any actual save game — Franchise, Season, or Elimination — created from it.
          </p>

          <p v-if="leaguesError" class="error-text">{{ leaguesError }}</p>

          <div v-if="loadingLeagues" class="progress-state">
            <LoadingSpinner />
            <p>Scanning save files…</p>
          </div>

          <div v-else-if="leagues.length === 0" class="empty-state">
            <p>No leagues found on this machine.</p>
          </div>

          <div v-else class="league-groups">
            <div v-for="group in leagueGroups" :key="group.name" class="league-group">
              <Accordion v-if="group.shell && group.realSaves.length > 0" multiple>
                <AccordionPanel value="shell">
                  <AccordionHeader as="div">
                    <div class="league-row league-row--in-header">
                      <div class="league-info">
                        <span class="league-name">{{ group.name }}</span>
                        <span class="league-stats">{{ statsSummary(group.shell) }}</span>
                      </div>
                      <AppButton
                        variant="secondary"
                        size="sm"
                        :disabled="exportingGUID === group.shell.guid"
                        @click.stop="exportLeague(group.shell)"
                      >
                        {{ exportingGUID === group.shell.guid ? 'Exporting…' : 'Export Empty League' }}
                      </AppButton>
                    </div>
                  </AccordionHeader>
                  <AccordionContent>
                    <div v-for="save in group.realSaves" :key="save.guid" class="league-row">
                      <div class="league-info">
                        <Tag :value="modeLabel(save.mode)" :severity="modeSeverity(save.mode)" />
                      </div>
                      <AppButton
                        variant="secondary"
                        size="sm"
                        :disabled="exportingGUID === save.guid"
                        @click="exportLeague(save)"
                      >
                        {{ exportingGUID === save.guid ? 'Exporting…' : 'Export League with Save Game' }}
                      </AppButton>
                    </div>
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>

              <template v-else>
                <div v-for="entry in group.entries" :key="entry.guid" class="league-row">
                  <div class="league-info">
                    <span class="league-info-top">
                      <span class="league-name">{{ entry.name }}</span>
                      <Tag v-if="entry.mode !== 'none'" :value="modeLabel(entry.mode)" :severity="modeSeverity(entry.mode)" />
                    </span>
                    <span class="league-stats">{{ statsSummary(entry) }}</span>
                  </div>
                  <AppButton
                    variant="secondary"
                    size="sm"
                    :disabled="exportingGUID === entry.guid"
                    @click="exportLeague(entry)"
                  >
                    {{ exportingGUID === entry.guid ? 'Exporting…' : exportButtonLabel(entry) }}
                  </AppButton>
                </div>
              </template>
            </div>
          </div>
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
              <h3>
                {{ importPreview.overview.name }}
                <Tag
                  v-if="importPreview.overview.mode !== 'none'"
                  :value="modeLabel(importPreview.overview.mode)"
                  :severity="modeSeverity(importPreview.overview.mode)"
                />
              </h3>
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
  margin: 0 auto;
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

.league-groups {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.league-group :deep(.p-accordion),
.league-group :deep(.p-accordionpanel) {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

/* The box styling (border/background/radius/padding) lives on the header
   itself, not the row inside it, so the expand/collapse caret sits inside
   the same bordered bar as the league name and Export button rather than
   floating outside it. min-width: 0 on the row lets its text shrink/wrap
   instead of forcing the header wider than its container. */
.league-group :deep(.p-accordionheader) {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.875rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

.league-group :deep(.p-accordioncontent-content) {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.5rem 0 0;
}

.league-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.875rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
}

/* Inside the accordion header, the box styling above already applies to
   the header itself — this row just needs to lay out its own children and
   push the Export button to the far right. */
.league-row--in-header {
  flex: 1;
  min-width: 0;
  padding: 0;
  background: none;
  border: none;
}

.league-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

/* In the nested real-save rows, .league-info's only child is the mode Tag —
   without this it stretches to the column's full cross-axis width instead
   of sizing to its own content. */
.league-info :deep(.p-tag) {
  align-self: flex-start;
}

.league-info-top {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.league-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.9375rem;
  color: var(--color-text-primary);
}

.league-stats {
  font-size: 0.8125rem;
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
