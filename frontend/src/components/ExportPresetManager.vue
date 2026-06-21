<script lang="ts" setup>
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'
import { onMounted, ref } from 'vue'
import { DeleteExportPreset, GetExportPresets, SaveExportPreset } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'

const props = defineProps<{
  currentConfigJSON: string
  datasetId: string
}>()

const emit = defineEmits<{
  load: [datasetId: string, configJSON: string]
}>()

const toast = useToast()
const presets = ref<main.ExportPresetDTO[]>([])
const newPresetName = ref('')
const saving = ref(false)

async function loadPresets() {
  try {
    presets.value = await GetExportPresets()
  } catch {
    // non-fatal — list just stays empty
  }
}

async function savePreset() {
  const name = newPresetName.value.trim()
  if (!name) return
  saving.value = true
  try {
    const saved = await SaveExportPreset(name, props.datasetId, props.currentConfigJSON)
    presets.value = [saved, ...presets.value]
    newPresetName.value = ''
    toast.add({ severity: 'success', summary: `Preset "${name}" saved`, life: 3000 })
  } catch (e) {
    toast.add({ severity: 'error', summary: String(e), life: 5000 })
  } finally {
    saving.value = false
  }
}

async function deletePreset(preset: main.ExportPresetDTO) {
  try {
    await DeleteExportPreset(preset.id)
    presets.value = presets.value.filter((p) => p.id !== preset.id)
    toast.add({ severity: 'success', summary: `Preset "${preset.name}" deleted`, life: 4000 })
  } catch (e) {
    toast.add({ severity: 'error', summary: String(e), life: 5000 })
  }
}

function loadPreset(preset: main.ExportPresetDTO) {
  emit('load', preset.datasetId, preset.configJson)
}

onMounted(loadPresets)
</script>

<template>
  <div class="preset-manager">
    <div class="save-row">
      <input
        v-model="newPresetName"
        type="text"
        class="preset-name-input"
        placeholder="Preset name…"
        maxlength="80"
        @keyup.enter="savePreset"
      />
      <Button
        label="Save"
        size="small"
        :disabled="!newPresetName.trim() || saving"
        @click="savePreset"
      />
    </div>

    <div v-if="presets.length === 0" class="no-presets">No saved presets.</div>

    <ul v-else class="preset-list">
      <li v-for="p in presets" :key="p.id" class="preset-item">
        <span class="preset-name" :title="p.name">{{ p.name }}</span>
        <div class="preset-actions">
          <Button
            label="Load"
            size="small"
            severity="secondary"
            text
            @click="loadPreset(p)"
          />
          <Button
            icon="pi pi-trash"
            size="small"
            severity="danger"
            text
            @click="deletePreset(p)"
          />
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.preset-manager {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.save-row {
  display: flex;
  gap: 0.375rem;
}

.preset-name-input {
  flex: 1;
  padding: 0.25rem 0.5rem;
  font-size: 0.8125rem;
  font-family: inherit;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  min-width: 0;
}

.preset-name-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.preset-name-input::placeholder {
  color: var(--color-text-secondary);
}

.no-presets {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

.preset-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.preset-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.375rem;
  border-radius: 4px;
}

.preset-item:hover {
  background: var(--color-surface-2);
}

.preset-name {
  flex: 1;
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

.preset-actions {
  display: flex;
  align-items: center;
  gap: 0;
  flex-shrink: 0;
}
</style>
