<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import type { main } from '../../wailsjs/go/models'
import { useSaveFileCandidates } from '../composables/useSaveFileCandidates'
import AppButton from './AppButton.vue'
import AppHelpButton from './AppHelpButton.vue'
import SaveFilePicker from './SaveFilePicker.vue'

const emit = defineEmits<{
  create: [name: string, gameVersion: string, saveFilePath: string, leagueGUID: string]
  cancel: []
}>()

// SMB3 support is deferred — only SMB4 is available for now.
const GAME_VERSION = 'smb4'

const name = ref('')
const selectedPath = ref('')
const selectedLeagueGUID = ref('')
const selectedProbe = ref<main.SaveFileCandidateDTO | null>(null)
const submitting = ref(false)
const error = ref<string | null>(null)

const {
  candidates,
  loading,
  scanning,
  browsing,
  error: pickerError,
  load,
  scanDirectory,
  browseFile,
} = useSaveFileCandidates()

onMounted(load)

function onSaveFileChange(path: string, leagueGUID: string, probe?: main.SaveFileCandidateDTO) {
  selectedPath.value = path
  selectedLeagueGUID.value = leagueGUID
  selectedProbe.value = probe ?? null
}

async function handleBrowseFile() {
  const c = await browseFile()
  if (c) onSaveFileChange(c.path, c.leagueGUID, c)
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
</script>

<template>
  <div class="franchise-create">
    <div class="heading-row">
      <h2>New Franchise</h2>
      <AppHelpButton docs-path="save-game-setup.html#creating-a-franchise" />
    </div>

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

    <div class="field">
      <label>Save File</label>
      <p class="field-hint">
        Franchise and Season mode saves are supported — this Franchise will track whichever mode your save uses.
        Elimination mode is not supported.
      </p>
      <SaveFilePicker
        :candidates="candidates"
        :loading="loading"
        :scanning="scanning"
        :browsing="browsing"
        :error="pickerError"
        :selected-path="selectedPath"
        @change="onSaveFileChange"
        @scan-directory="scanDirectory"
        @browse-file="handleBrowseFile"
      />
    </div>

    <!-- Confirmation: show what will be connected -->
    <div v-if="selectedPath" class="selection-summary">
      <span class="summary-icon">✓</span>
      <span>
        <template v-if="selectedProbe?.leagueName">
          <strong>{{ selectedProbe.leagueName }}</strong>
          <template v-if="selectedProbe.playerTeamName"> · {{ selectedProbe.playerTeamName }}</template>
        </template>
        <template v-else>
          {{ selectedPath.split(/[\\/]/).pop() }}
        </template>
      </span>
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

.heading-row {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  margin-bottom: 1.5rem;
}

h2 {
  font-size: 1.4rem;
  font-weight: 600;
  margin: 0;
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

.field-hint {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  margin-bottom: 0.6rem;
  line-height: 1.4;
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

.selection-summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: color-mix(in srgb, var(--color-accent) 8%, var(--color-surface-2));
  border: 1px solid color-mix(in srgb, var(--color-accent) 30%, transparent);
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--color-text-primary);
  margin-bottom: 1rem;
}

.summary-icon {
  color: var(--color-accent);
  font-weight: 700;
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
