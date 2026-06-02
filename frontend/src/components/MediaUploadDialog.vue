<script lang="ts" setup>
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import { useToast } from 'primevue/usetoast'
import { ref } from 'vue'
import { BrowseMediaFile, UploadMedia } from '../../wailsjs/go/main/App'
import MediaAssociationPicker from './MediaAssociationPicker.vue'

const props = defineProps<{
  entityType: 'team_season' | 'player'
  entityId: number
  entityLabel: string
}>()

const emit = defineEmits<{
  uploaded: []
}>()

const visible = defineModel<boolean>('visible', { required: true })

const toast = useToast()

const filePath = ref('')
const mediaName = ref('')
const description = ref('')
const uploading = ref(false)

type AssocChip = { type: 'team_season' | 'player'; id: number; label: string }
const extraAssociations = ref<AssocChip[]>([])

const showTeamSeasonPicker = ref(false)
const showPlayerPicker = ref(false)

async function browseFile() {
  const path = await BrowseMediaFile()
  if (!path) return
  filePath.value = path
  // Default name from filename without extension
  const base = path.replace(/\\/g, '/').split('/').pop() ?? path
  const dotIdx = base.lastIndexOf('.')
  mediaName.value = dotIdx >= 0 ? base.slice(0, dotIdx) : base
}

function onPicked(type: 'team_season' | 'player', id: number, label: string) {
  const existing = extraAssociations.value.findIndex((a) => a.type === type && a.id === id)
  if (existing === -1) {
    extraAssociations.value.push({ type, id, label })
  }
  showTeamSeasonPicker.value = false
  showPlayerPicker.value = false
}

function removeExtraAssoc(index: number) {
  extraAssociations.value.splice(index, 1)
}

function alreadySelectedTeamHistoryIds(): number[] {
  const ids = extraAssociations.value.filter((a) => a.type === 'team_season').map((a) => a.id)
  if (props.entityType === 'team_season') ids.push(props.entityId)
  return ids
}

function alreadySelectedPlayerIds(): number[] {
  const ids = extraAssociations.value.filter((a) => a.type === 'player').map((a) => a.id)
  if (props.entityType === 'player') ids.push(props.entityId)
  return ids
}

function reset() {
  filePath.value = ''
  mediaName.value = ''
  description.value = ''
  extraAssociations.value = []
  showTeamSeasonPicker.value = false
  showPlayerPicker.value = false
}

async function upload() {
  if (!filePath.value || !mediaName.value.trim()) return
  uploading.value = true
  try {
    const teamHistoryIds: number[] = [
      ...(props.entityType === 'team_season' ? [props.entityId] : []),
      ...extraAssociations.value.filter((a) => a.type === 'team_season').map((a) => a.id),
    ]
    const playerIds: number[] = [
      ...(props.entityType === 'player' ? [props.entityId] : []),
      ...extraAssociations.value.filter((a) => a.type === 'player').map((a) => a.id),
    ]
    await UploadMedia({
      name: mediaName.value.trim(),
      description: description.value,
      filePath: filePath.value,
      teamHistoryIds,
      playerIds,
    })
    toast.add({ severity: 'success', summary: 'Uploaded', detail: `"${mediaName.value}" added`, life: 3000 })
    visible.value = false
    reset()
    emit('uploaded')
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Upload failed', detail: String(e), life: 5000 })
  } finally {
    uploading.value = false
  }
}

function onHide() {
  if (!uploading.value) reset()
}
</script>

<template>
  <Dialog
    v-model:visible="visible"
    header="Upload media"
    :modal="true"
    :closable="!uploading"
    :style="{ width: '480px' }"
    @hide="onHide"
  >
    <div class="upload-form">
      <!-- File selection -->
      <div class="form-row">
        <div class="file-row">
          <Button label="Browse…" size="small" severity="secondary" @click="browseFile" />
          <span v-if="filePath" class="file-name">{{ filePath.replace(/\\/g, '/').split('/').pop() }}</span>
          <span v-else class="file-placeholder">No file selected</span>
        </div>
      </div>

      <!-- Name -->
      <div class="form-row">
        <label class="form-label" for="media-name">Name</label>
        <InputText id="media-name" v-model="mediaName" class="form-input" placeholder="Screenshot or video name" />
      </div>

      <!-- Description -->
      <div class="form-row">
        <label class="form-label" for="media-desc">Description <span class="optional">(optional)</span></label>
        <Textarea
          id="media-desc"
          v-model="description"
          class="form-input"
          placeholder="Add a note about this moment…"
          :rows="2"
          auto-resize
        />
      </div>

      <!-- Associations -->
      <div class="form-row">
        <div class="form-label">Linked to</div>
        <div class="assoc-chips">
          <!-- Locked context chip -->
          <span class="assoc-chip assoc-chip--locked">
            <span class="chip-icon" aria-hidden="true">{{ entityType === 'team_season' ? '📅' : '👤' }}</span>
            {{ entityLabel }}
          </span>

          <!-- Extra association chips -->
          <span v-for="(assoc, i) in extraAssociations" :key="`${assoc.type}-${assoc.id}`" class="assoc-chip">
            <span class="chip-icon" aria-hidden="true">{{ assoc.type === 'team_season' ? '📅' : '👤' }}</span>
            {{ assoc.label }}
            <button class="chip-remove" aria-label="Remove" @click="removeExtraAssoc(i)">×</button>
          </span>
        </div>

        <!-- Picker toggles -->
        <div class="assoc-add-row">
          <button class="add-assoc-btn" @click="showTeamSeasonPicker = !showTeamSeasonPicker; showPlayerPicker = false">
            + Add team season
          </button>
          <button class="add-assoc-btn" @click="showPlayerPicker = !showPlayerPicker; showTeamSeasonPicker = false">
            + Add player
          </button>
        </div>

        <MediaAssociationPicker
          v-if="showTeamSeasonPicker"
          mode="team_season"
          :already-selected-team-history-ids="alreadySelectedTeamHistoryIds()"
          @picked="onPicked"
        />
        <MediaAssociationPicker
          v-if="showPlayerPicker"
          mode="player"
          :already-selected-player-ids="alreadySelectedPlayerIds()"
          @picked="onPicked"
        />
      </div>
    </div>

    <template #footer>
      <Button label="Cancel" severity="secondary" :disabled="uploading" @click="visible = false" />
      <Button
        label="Upload"
        :loading="uploading"
        :disabled="!filePath || !mediaName.trim()"
        @click="upload"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.upload-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-top: 0.25rem;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.optional {
  font-weight: 400;
  opacity: 0.7;
}

.form-input {
  width: 100%;
}

.file-row {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.file-name {
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 280px;
}

.file-placeholder {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.assoc-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  min-height: 28px;
}

.assoc-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.1875rem 0.5rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  font-size: 0.8125rem;
  color: var(--color-text-primary);
}

.assoc-chip--locked {
  border-color: var(--color-accent);
  opacity: 0.85;
}

.chip-icon {
  font-size: 0.75rem;
}

.chip-remove {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
  color: var(--color-text-secondary);
  padding: 0 0 0 0.125rem;
  font-family: var(--font-sans);
}

.chip-remove:hover {
  color: var(--color-text-primary);
}

.assoc-add-row {
  display: flex;
  gap: 0.5rem;
}

.add-assoc-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.8125rem;
  color: var(--color-accent);
  font-family: var(--font-sans);
  padding: 0;
  text-decoration: underline;
}

.add-assoc-btn:hover {
  opacity: 0.8;
}
</style>
