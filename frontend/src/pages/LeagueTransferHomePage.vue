<script lang="ts" setup>
import Accordion from 'primevue/accordion'
import AccordionContent from 'primevue/accordioncontent'
import AccordionHeader from 'primevue/accordionheader'
import AccordionPanel from 'primevue/accordionpanel'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import type { MenuItem } from 'primevue/menuitem'
import RadioButton from 'primevue/radiobutton'
import SplitButton from 'primevue/splitbutton'
import TabPanel from 'primevue/tabpanel'
import TabView from 'primevue/tabview'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { computed, onMounted, ref } from 'vue'
import {
  BrowseLeagueExportDirectory,
  BrowseLeagueImportZip,
  ConfirmLeagueImport,
  DiscoverLeagues,
  ExportLeague,
  ExportLeagueWithRename,
  ExportSnapshotAsLeague,
  ListSnapshotExportCandidates,
  OpenLeagueExportDir,
  PreviewLeagueImport,
} from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const toast = useToast()

onMounted(() => {
  loadLeagues()
  loadSnapshotCandidates()
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
  return league.mode === 'none' ? 'Export Empty League' : 'Export Save Game Only'
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

// Lets a user point export discovery at a folder outside the default Steam
// save locations — results are merged into the same discovered-leagues
// list, deduplicated by GUID, rather than replacing it.
const browsingFolder = ref(false)

async function browseExportFolder() {
  browsingFolder.value = true
  try {
    const found = await BrowseLeagueExportDirectory()
    if (found.length === 0) return

    const existingGUIDs = new Set(leagues.value.map((l) => l.guid))
    const newLeagues = found.filter((l) => !existingGUIDs.has(l.guid))
    leagues.value = [...leagues.value, ...newLeagues]
    toast.add({
      severity: 'success',
      summary: newLeagues.length > 0 ? `Found ${newLeagues.length} league(s)` : 'No new leagues found in that folder',
      life: 3000,
    })
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Could not scan folder', detail: String(e), life: 5000 })
  } finally {
    browsingFolder.value = false
  }
}

async function exportLeague(league: main.LeagueOverviewDTO) {
  exportingGUID.value = league.guid
  try {
    const outputPath = await ExportLeague(league.guid, league.sourcePath)
    toast.add({ severity: 'success', summary: `Exported "${league.name}"`, detail: outputPath, life: 5000 })
    await OpenLeagueExportDir(outputPath)
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Export failed', detail: String(e), life: 5000 })
  } finally {
    exportingGUID.value = null
  }
}

// Renaming lets a user disambiguate an exported league before sharing it —
// the export itself is unaffected, only the display name baked into the
// copy that gets zipped up. The same dialog is reused for snapshot export
// (renameMode 'snapshot'), where supplying a name is mandatory rather than
// an optional disambiguation step — a snapshot has no reliable "current"
// name to default to.
type RenameMode = 'discovered' | 'snapshot'
const renameDialogVisible = ref(false)
const renameMode = ref<RenameMode>('discovered')
const renameTargetLeague = ref<main.LeagueOverviewDTO | null>(null)
const renameTargetSnapshot = ref<main.SnapshotExportCandidateDTO | null>(null)
const renameValue = ref('')

function openRenameDialog(league: main.LeagueOverviewDTO) {
  renameMode.value = 'discovered'
  renameTargetLeague.value = league
  renameValue.value = league.name
  renameDialogVisible.value = true
}

function openSnapshotExportDialog(candidate: main.SnapshotExportCandidateDTO) {
  renameMode.value = 'snapshot'
  renameTargetSnapshot.value = candidate
  renameValue.value = ''
  renameDialogVisible.value = true
}

function exportMenuItems(league: main.LeagueOverviewDTO): MenuItem[] {
  return [
    {
      label: 'Export with New Name…',
      icon: 'pi pi-pencil',
      command: () => openRenameDialog(league),
    },
  ]
}

async function confirmRenameExport() {
  if (renameMode.value === 'snapshot') {
    await confirmSnapshotExport()
    return
  }

  const league = renameTargetLeague.value
  if (!league) return

  exportingGUID.value = league.guid
  try {
    const outputPath = await ExportLeagueWithRename(league.guid, league.sourcePath, renameValue.value)
    toast.add({
      severity: 'success',
      summary: `Exported as "${renameValue.value.trim()}"`,
      detail: outputPath,
      life: 5000,
    })
    renameDialogVisible.value = false
    await OpenLeagueExportDir(outputPath)
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Export failed', detail: String(e), life: 5000 })
  } finally {
    exportingGUID.value = null
  }
}

// ── From Snapshot tab ────────────────────────────────────────────────────────

const snapshotCandidates = ref<main.SnapshotExportCandidateDTO[]>([])
const loadingSnapshots = ref(false)
const snapshotsError = ref<string | null>(null)
const exportingSnapshotId = ref<number | null>(null)

interface SnapshotGroup {
  franchiseId: string
  franchiseName: string
  snapshots: main.SnapshotExportCandidateDTO[]
}

const snapshotGroups = computed<SnapshotGroup[]>(() => {
  const byFranchise = new Map<string, SnapshotGroup>()
  for (const candidate of snapshotCandidates.value) {
    const group = byFranchise.get(candidate.franchiseId) ?? {
      franchiseId: candidate.franchiseId,
      franchiseName: candidate.franchiseName,
      snapshots: [],
    }
    group.snapshots.push(candidate)
    byFranchise.set(candidate.franchiseId, group)
  }
  for (const group of byFranchise.values()) {
    group.snapshots.sort((a, b) => b.seasonNum - a.seasonNum)
  }
  return Array.from(byFranchise.values()).sort((a, b) => a.franchiseName.localeCompare(b.franchiseName))
})

function formatSnapshotCapturedAt(iso: string): string {
  if (!iso) return ''
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

function formatSnapshotSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function loadSnapshotCandidates() {
  loadingSnapshots.value = true
  snapshotsError.value = null
  try {
    snapshotCandidates.value = (await ListSnapshotExportCandidates()) ?? []
  } catch (e) {
    snapshotsError.value = String(e)
  } finally {
    loadingSnapshots.value = false
  }
}

async function confirmSnapshotExport() {
  const candidate = renameTargetSnapshot.value
  if (!candidate || !renameValue.value.trim()) return

  exportingSnapshotId.value = candidate.snapshotId
  try {
    const outputPath = await ExportSnapshotAsLeague(candidate.franchiseId, candidate.snapshotId, renameValue.value)
    toast.add({
      severity: 'success',
      summary: `Exported as "${renameValue.value.trim()}"`,
      detail: outputPath,
      life: 5000,
    })
    renameDialogVisible.value = false
    await OpenLeagueExportDir(outputPath)
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Export failed', detail: String(e), life: 5000 })
  } finally {
    exportingSnapshotId.value = null
  }
}

// ── Import tab ───────────────────────────────────────────────────────────────

type ImportStep = 'choose' | 'preview' | 'importing'

const importStep = ref<ImportStep>('choose')
const importZipPath = ref<string | null>(null)
const importPreview = ref<main.LeagueImportPreviewDTO | null>(null)
const selectedTargetDir = ref<string | null>(null)
const loadingPreview = ref(false)
const confirmingImport = ref(false)

// Backend errors are written for logs, not players — translate the ones a
// user can actually hit by picking the wrong file into plain language before
// they reach a toast. Anything unrecognized falls back to the raw message.
function describeImportError(e: unknown): string {
  const message = e instanceof Error ? e.message : String(e)

  if (message.includes('missing manifest.json') || message.includes('opening zip')) {
    return "That doesn't look like a league export. Choose the .zip file created by smb-tools' Export tab."
  }
  if (
    message.includes('missing the league .sav file') ||
    message.includes('missing the league .sav.bak file') ||
    message.includes('does not look like a zlib-compressed save file') ||
    message.includes('manifest is malformed') ||
    message.includes('manifest is missing leagueName') ||
    message.includes("does not match the manifest's league GUID")
  ) {
    return 'This export package looks incomplete or corrupted.'
  }
  if (message.includes('SMB4 is currently running')) {
    return 'Close Super Mega Baseball 4 before importing a league.'
  }
  if (message.includes('already registered')) {
    return 'This league has already been imported into that save directory.'
  }
  if (message.includes('master.sav changed during import')) {
    return 'The save file changed during import (the game may have been running), please try again.'
  }
  return message
}

async function chooseImportFile() {
  const path = await BrowseLeagueImportZip()
  if (!path) return

  importZipPath.value = path
  loadingPreview.value = true
  try {
    importPreview.value = await PreviewLeagueImport(path)
    selectedTargetDir.value = importPreview.value.targets.find((t) => !t.alreadyRegistered)?.dirPath ?? null
    importStep.value = 'preview'
    toast.add({ severity: 'success', summary: `Loaded "${importPreview.value.overview.name}"`, life: 3000 })
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Could not read league file', detail: describeImportError(e), life: 6000 })
  } finally {
    loadingPreview.value = false
  }
}

async function confirmImport() {
  if (!importZipPath.value || !selectedTargetDir.value) return
  confirmingImport.value = true
  try {
    await ConfirmLeagueImport(importZipPath.value, selectedTargetDir.value)
    toast.add({ severity: 'success', summary: 'League imported', life: 4000 })
    resetImport()
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Import failed', detail: describeImportError(e), life: 6000 })
  } finally {
    confirmingImport.value = false
  }
}

function resetImport() {
  importStep.value = 'choose'
  importZipPath.value = null
  importPreview.value = null
  selectedTargetDir.value = null
}
</script>

<template>
  <div class="lt-page">
    <TabView>
      <TabPanel value="0" header="Export">
        <TabView class="export-subtabs">
          <TabPanel value="0" header="From Save File">
            <div class="tab-content">
              <p class="tab-desc">
                Export a league from any save file on this machine so you can share it with someone else. You can
                export an empty league shell (just the teams/conferences/divisions setup, no games played) as well
                as any actual save game (Franchise, Season, or Elimination) created from it.
              </p>
              <p class="tab-desc">
                <strong>These are not the same thing.</strong> The shell is what shows up under Customizations in
                SMB4. Exporting a save game only shares the in-progress Franchise/Season/Elimination itself. It does
                not include the shell, so the recipient won't see a Customizations entry for it unless you export
                and send the shell too.
              </p>
              <p class="tab-desc">
                If you export the save game, the recipient can re-create the league shell by using the "Export to
                League" option in-game.
              </p>

              <div class="browse-folder-row">
                <AppButton variant="secondary" :disabled="browsingFolder" @click="browseExportFolder">
                  {{ browsingFolder ? 'Scanning…' : 'Browse Folder…' }}
                </AppButton>
              </div>

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
                          <div class="split-button-stop" @click.stop>
                            <SplitButton
                              :label="exportingGUID === group.shell.guid ? 'Exporting…' : 'Export Empty League'"
                              size="small"
                              :disabled="exportingGUID === group.shell.guid"
                              :model="exportMenuItems(group.shell)"
                              @click="exportLeague(group.shell)"
                            />
                          </div>
                        </div>
                        <template #toggleicon>
                          <span class="toggle-slot">
                            <i class="pi pi-chevron-down" />
                          </span>
                        </template>
                      </AccordionHeader>
                      <AccordionContent>
                        <div v-for="save in group.realSaves" :key="save.guid" class="league-row">
                          <div class="league-info">
                            <Tag :value="modeLabel(save.mode)" :severity="modeSeverity(save.mode)" />
                          </div>
                          <SplitButton
                            :label="exportingGUID === save.guid ? 'Exporting…' : 'Export Save Game Only'"
                            size="small"
                            :disabled="exportingGUID === save.guid"
                            :model="exportMenuItems(save)"
                            @click="exportLeague(save)"
                          />
                        </div>
                      </AccordionContent>
                    </AccordionPanel>
                  </Accordion>

                  <template v-else>
                    <div v-for="entry in group.entries" :key="entry.guid" class="league-row">
                      <div class="league-info">
                        <span class="league-info-top">
                          <span class="league-name">{{ entry.name }}</span>
                          <Tag
                            v-if="entry.mode !== 'none'"
                            :value="modeLabel(entry.mode)"
                            :severity="modeSeverity(entry.mode)"
                          />
                        </span>
                        <span class="league-stats">{{ statsSummary(entry) }}</span>
                      </div>
                      <SplitButton
                        :label="exportingGUID === entry.guid ? 'Exporting…' : exportButtonLabel(entry)"
                        size="small"
                        :disabled="exportingGUID === entry.guid"
                        :model="exportMenuItems(entry)"
                        @click="exportLeague(entry)"
                      />
                      <span class="toggle-slot" />
                    </div>
                  </template>
                </div>
              </div>
            </div>
          </TabPanel>

          <TabPanel value="1" header="From Franchise Snapshot">
            <div class="tab-content">
              <p class="tab-desc">
                Export a franchise snapshot — a point-in-time save captured automatically while tracking a
                franchise — as a shareable league save game. The export gets a freshly generated league identity
                and a name you choose; it does not affect the franchise or its snapshots in any way.
              </p>

              <p v-if="snapshotsError" class="error-text">{{ snapshotsError }}</p>

              <div v-if="loadingSnapshots" class="progress-state">
                <LoadingSpinner />
                <p>Loading franchise snapshots…</p>
              </div>

              <div v-else-if="snapshotGroups.length === 0" class="empty-state">
                <p>No franchise snapshots found.</p>
              </div>

              <div v-else class="snapshot-groups">
                <div v-for="group in snapshotGroups" :key="group.franchiseId" class="snapshot-group">
                  <h3 class="snapshot-group-name">{{ group.franchiseName }}</h3>
                  <div v-for="snap in group.snapshots" :key="snap.snapshotId" class="league-row">
                    <div class="league-info">
                      <span class="league-name">Season {{ snap.seasonNum }}</span>
                      <span class="league-stats">
                        {{ formatSnapshotCapturedAt(snap.capturedAt) }} · {{ formatSnapshotSize(snap.fileSizeBytes) }}
                      </span>
                    </div>
                    <AppButton
                      variant="primary"
                      size="sm"
                      :disabled="exportingSnapshotId === snap.snapshotId"
                      @click="openSnapshotExportDialog(snap)"
                    >
                      {{ exportingSnapshotId === snap.snapshotId ? 'Exporting…' : 'Export Save Game Only' }}
                    </AppButton>
                  </div>
                </div>
              </div>
            </div>
          </TabPanel>
        </TabView>
      </TabPanel>

      <TabPanel value="1" header="Import">
        <div class="tab-content">
          <p class="tab-desc">
            Import a league someone exported for you. smb-tools does not scan league files for malware.
            Only import files from sources you trust.
          </p>

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
              <p v-if="importPreview.overview.mode !== 'none'" class="save-only-note">
                This is a save game export — it will not add a Customizations entry in SMB4, only the
                {{ modeLabel(importPreview.overview.mode).toLowerCase() }} game itself.
              </p>
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
                <span
                  v-if="target.alreadyRegistered"
                  class="target-warning"
                  title="This league has already been imported into this save directory."
                >
                  Already imported here
                </span>
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

    <Dialog
      v-model:visible="renameDialogVisible"
      modal
      :header="renameMode === 'snapshot' ? 'Export Snapshot' : 'Export with New Name'"
      :style="{ width: '420px' }"
    >
      <div class="rename-dialog-body">
        <label class="rename-label" for="rename-input">New league name</label>
        <InputText id="rename-input" v-model="renameValue" autofocus class="rename-input" />
      </div>

      <template #footer>
        <AppButton
          variant="secondary"
          :disabled="exportingGUID !== null || exportingSnapshotId !== null"
          @click="renameDialogVisible = false"
        >
          Cancel
        </AppButton>
        <AppButton
          variant="primary"
          :disabled="exportingGUID !== null || exportingSnapshotId !== null || !renameValue.trim()"
          @click="confirmRenameExport"
        >
          {{ exportingGUID !== null || exportingSnapshotId !== null ? 'Exporting…' : 'Export' }}
        </AppButton>
      </template>
    </Dialog>
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

.export-subtabs :deep(.p-tabview-nav) {
  margin-top: -0.5rem;
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

.browse-folder-row {
  display: flex;
  justify-content: flex-end;
}

.snapshot-groups {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.snapshot-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.snapshot-group-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
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

/* Wraps the shell row's SplitButton inside the accordion header so that a
   click anywhere on it (including the dropdown toggle, which doesn't emit a
   Vue 'click' event smb-tools can attach .stop to) never bubbles up to the
   AccordionHeader's native click listener and toggles the panel. */
.split-button-stop {
  display: flex;
}

/* A fixed-size slot for the expand/collapse chevron, present on every row
   (accordion header or flat) so Export buttons always land in the same
   column. Accordion headers fill it via the #toggleicon slot; flat rows
   (which have no nested content to expand) leave it empty. */
.toggle-slot {
  flex-shrink: 0;
  width: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
}

.toggle-slot i {
  transition: transform 0.2s ease;
}

.league-group :deep(.p-accordionheader[aria-expanded='true']) .toggle-slot i {
  transform: rotate(180deg);
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

.save-only-note {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  font-style: italic;
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

.rename-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.rename-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.rename-input {
  width: 100%;
}

.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
