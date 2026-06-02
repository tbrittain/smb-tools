<script lang="ts" setup>
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import Tab from 'primevue/tab'
import TabList from 'primevue/tablist'
import TabPanel from 'primevue/tabpanel'
import TabPanels from 'primevue/tabpanels'
import Tabs from 'primevue/tabs'
import { computed, ref, watch } from 'vue'
import {
  AssignExistingTeamLogo,
  BrowseLogoFile,
  DeleteTeamLogoAssignment,
  GetTeamLogos,
  UploadAndAssignTeamLogo,
} from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import TeamLogoDisplay from './TeamLogoDisplay.vue'

const props = defineProps<{
  teamId: number
  latestSeason: number
  availableSeasons: number[]
}>()

const visible = defineModel<boolean>('visible', { required: true })

const logos = ref<main.TeamLogoDTO[]>([])
const loadingLogos = ref(false)
const activeTab = ref('upload')

// Upload tab state
const pendingFilePath = ref('')
const pendingFileName = ref('')
const uploading = ref(false)
const uploadError = ref<string | null>(null)

// Use-existing tab state
const selectedLogoId = ref<string | null>(null)
const assigning = ref(false)
const assignError = ref<string | null>(null)

// Shared range form state
const allSeasons = ref(true)
const startSeason = ref<number | null>(props.latestSeason)
const endSeason = ref<number | null>(null)
const noEnd = ref(true)

const resolvedRange = computed<[number | null, number | null]>(() => {
  if (allSeasons.value) return [null, null]
  return [startSeason.value, noEnd.value ? null : endSeason.value]
})

function applyQuickRange(mode: 'single' | 'onwards') {
  allSeasons.value = false
  startSeason.value = props.latestSeason
  if (mode === 'single') {
    noEnd.value = false
    endSeason.value = props.latestSeason
  } else {
    noEnd.value = true
    endSeason.value = null
  }
}

async function loadLogos() {
  loadingLogos.value = true
  try {
    logos.value = (await GetTeamLogos(props.teamId)) ?? []
  } catch {
    logos.value = []
  } finally {
    loadingLogos.value = false
  }
}

watch(visible, (v) => {
  if (v) {
    loadLogos()
    resetForm()
  }
})

function resetForm() {
  pendingFilePath.value = ''
  pendingFileName.value = ''
  uploadError.value = null
  assignError.value = null
  selectedLogoId.value = null
  allSeasons.value = true
  startSeason.value = props.latestSeason
  endSeason.value = null
  noEnd.value = true
  activeTab.value = 'upload'
}

async function browseFile() {
  try {
    const path = await BrowseLogoFile()
    if (path) {
      pendingFilePath.value = path
      pendingFileName.value = path.split(/[\\/]/).pop() ?? path
      uploadError.value = null
    }
  } catch (e) {
    uploadError.value = String(e)
  }
}

async function upload() {
  if (!pendingFilePath.value) return
  uploading.value = true
  uploadError.value = null
  try {
    const [start, end] = resolvedRange.value
    await UploadAndAssignTeamLogo(props.teamId, pendingFilePath.value, start, end)
    pendingFilePath.value = ''
    pendingFileName.value = ''
    await loadLogos()
  } catch (e) {
    uploadError.value = String(e)
  } finally {
    uploading.value = false
  }
}

async function assignExisting() {
  if (!selectedLogoId.value) return
  assigning.value = true
  assignError.value = null
  try {
    const [start, end] = resolvedRange.value
    await AssignExistingTeamLogo(selectedLogoId.value, start, end)
    selectedLogoId.value = null
    await loadLogos()
  } catch (e) {
    assignError.value = String(e)
  } finally {
    assigning.value = false
  }
}

async function deleteAssignment(assignmentId: string) {
  try {
    await DeleteTeamLogoAssignment(assignmentId)
    await loadLogos()
  } catch (e) {
    uploadError.value = String(e)
  }
}

function rangeLabel(a: main.TeamLogoAssignmentDTO): string {
  if (a.startSeason == null && a.endSeason == null) return 'All seasons'
  if (a.startSeason != null && a.endSeason != null && a.startSeason === a.endSeason)
    return `Season ${a.startSeason} only`
  const start = a.startSeason != null ? `Season ${a.startSeason}` : 'Beginning'
  const end = a.endSeason != null ? `Season ${a.endSeason}` : 'ongoing'
  return `${start} – ${end}`
}

const allAssignments = computed(() =>
  logos.value.flatMap((logo) => logo.assignments.map((a) => ({ ...a, logoUrl: logo.logoUrl }))),
)
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Manage Team Logos" :style="{ width: '600px' }">
    <!-- Existing assignments -->
    <section v-if="allAssignments.length > 0" class="assignments-section">
      <p class="section-label">Current assignments</p>
      <div class="assignment-list">
        <div v-for="a in allAssignments" :key="a.id" class="assignment-row">
          <TeamLogoDisplay :logoUrl="a.logoUrl" size="sm" />
          <span class="range-label">{{ rangeLabel(a) }}</span>
          <Button
            icon="pi pi-trash"
            severity="danger"
            text
            size="small"
            aria-label="Delete assignment"
            @click="deleteAssignment(a.id)"
          />
        </div>
      </div>
    </section>

    <p v-else-if="!loadingLogos" class="no-logos-hint">
      No logos uploaded yet. Use the Upload tab below to add one.
    </p>

    <!-- Tabs -->
    <Tabs v-model:value="activeTab" class="logo-tabs">
      <TabList>
        <Tab value="upload">Upload new logo</Tab>
        <Tab value="existing" :disabled="logos.length === 0">Use existing logo</Tab>
      </TabList>

      <TabPanels>
        <!-- Upload tab -->
        <TabPanel value="upload">
          <div class="tab-content">
            <div class="file-row">
              <Button label="Browse…" severity="secondary" size="small" @click="browseFile" />
              <span class="file-name">{{ pendingFileName || 'No file selected' }}</span>
            </div>

            <div class="range-form">
              <div class="range-row">
                <Checkbox v-model="allSeasons" :binary="true" input-id="allSeasons" />
                <label for="allSeasons">Apply to all seasons (past &amp; future)</label>
              </div>

              <div v-if="!allSeasons" class="range-detail">
                <div class="range-row">
                  <label>Start season</label>
                  <InputNumber v-model="startSeason" :min="1" show-buttons />
                </div>
                <div class="range-row">
                  <Checkbox v-model="noEnd" :binary="true" input-id="noEnd" />
                  <label for="noEnd">No end (apply to all future seasons)</label>
                </div>
                <div v-if="!noEnd" class="range-row">
                  <label>End season</label>
                  <InputNumber v-model="endSeason" :min="1" show-buttons />
                </div>
              </div>

              <div class="quick-btns">
                <Button
                  :label="`Just season ${latestSeason}`"
                  severity="secondary"
                  size="small"
                  text
                  @click="applyQuickRange('single')"
                />
                <Button
                  :label="`Season ${latestSeason} onwards`"
                  severity="secondary"
                  size="small"
                  text
                  @click="applyQuickRange('onwards')"
                />
              </div>
            </div>

            <p v-if="uploadError" class="error-text">{{ uploadError }}</p>

            <Button
              label="Upload &amp; assign"
              :disabled="!pendingFilePath || uploading"
              :loading="uploading"
              @click="upload"
            />
          </div>
        </TabPanel>

        <!-- Use existing tab -->
        <TabPanel value="existing">
          <div class="tab-content">
            <p v-if="logos.length === 0" class="hint-text">No logos uploaded for this team yet.</p>

            <div v-else class="logo-grid">
              <button
                v-for="logo in logos"
                :key="logo.id"
                :class="['logo-grid-item', { selected: selectedLogoId === logo.id }]"
                type="button"
                @click="selectedLogoId = logo.id"
              >
                <TeamLogoDisplay :logoUrl="logo.logoUrl" size="lg" />
              </button>
            </div>

            <div v-if="selectedLogoId" class="range-form">
              <div class="range-row">
                <Checkbox v-model="allSeasons" :binary="true" input-id="allSeasonsExisting" />
                <label for="allSeasonsExisting">Apply to all seasons (past &amp; future)</label>
              </div>

              <div v-if="!allSeasons" class="range-detail">
                <div class="range-row">
                  <label>Start season</label>
                  <InputNumber v-model="startSeason" :min="1" show-buttons />
                </div>
                <div class="range-row">
                  <Checkbox v-model="noEnd" :binary="true" input-id="noEndExisting" />
                  <label for="noEndExisting">No end (apply to all future seasons)</label>
                </div>
                <div v-if="!noEnd" class="range-row">
                  <label>End season</label>
                  <InputNumber v-model="endSeason" :min="1" show-buttons />
                </div>
              </div>

              <div class="quick-btns">
                <Button
                  :label="`Just season ${latestSeason}`"
                  severity="secondary"
                  size="small"
                  text
                  @click="applyQuickRange('single')"
                />
                <Button
                  :label="`Season ${latestSeason} onwards`"
                  severity="secondary"
                  size="small"
                  text
                  @click="applyQuickRange('onwards')"
                />
              </div>
            </div>

            <p v-if="assignError" class="error-text">{{ assignError }}</p>

            <Button
              v-if="selectedLogoId"
              label="Assign to range"
              :disabled="assigning"
              :loading="assigning"
              @click="assignExisting"
            />
          </div>
        </TabPanel>
      </TabPanels>
    </Tabs>

    <template #footer>
      <Button label="Close" severity="secondary" text @click="visible = false" />
    </template>
  </Dialog>
</template>

<style scoped>
.assignments-section {
  margin-bottom: 1.25rem;
}

.section-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.5rem;
}

.assignment-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.assignment-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.range-label {
  flex: 1;
  font-size: 0.875rem;
}

.no-logos-hint,
.hint-text {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  margin-bottom: 1rem;
}

.logo-tabs {
  margin-top: 0.5rem;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-top: 1rem;
}

.file-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.file-name {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

.range-form {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.range-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.range-detail {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-left: 1.25rem;
  border-left: 2px solid var(--color-border, #30363d);
}

.quick-btns {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.logo-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.logo-grid-item {
  background: none;
  border: 2px solid transparent;
  border-radius: 8px;
  padding: 4px;
  cursor: pointer;
  transition: border-color 0.15s;
}

.logo-grid-item:hover {
  border-color: var(--color-border, #30363d);
}

.logo-grid-item.selected {
  border-color: var(--p-primary-color, #4a9eff);
}

.error-text {
  font-size: 0.8125rem;
  color: var(--color-error, #f85149);
}
</style>
