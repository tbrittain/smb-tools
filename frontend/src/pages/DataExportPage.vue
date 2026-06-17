<script lang="ts" setup>
import Button from 'primevue/button'
import ExportColumnSelector from '../components/ExportColumnSelector.vue'
import ExportDatasetPicker from '../components/ExportDatasetPicker.vue'
import ExportFilterPanel from '../components/ExportFilterPanel.vue'
import ExportPresetManager from '../components/ExportPresetManager.vue'
import ExportPreviewTable from '../components/ExportPreviewTable.vue'
import { useExportConfig } from '../composables/useExportConfig'
import { EXPORT_DATASETS } from '../lib/exportDatasets'

const {
  activeDatasetId,
  activeDataset,
  onDatasetChange,
  selectedColumnKeys,
  selectedColumns,
  seasonMin,
  seasonMax,
  selectedTeamName,
  careerStatType,
  sortCol,
  sortDir,
  teams,
  previewRows,
  totalCount,
  isPreviewLoading,
  refreshPreview,
  isExporting,
  downloadCSV,
  toConfigJSON,
  fromConfigJSON,
} = useExportConfig()
</script>

<template>
  <div class="export-page">
    <div class="left-panel">
      <section class="panel-section">
        <h3 class="section-title">Dataset</h3>
        <ExportDatasetPicker
          :datasets="EXPORT_DATASETS"
          v-model="activeDatasetId"
          @update:model-value="onDatasetChange"
        />
      </section>

      <section class="panel-section">
        <h3 class="section-title">Columns</h3>
        <ExportColumnSelector
          :columns="activeDataset.columns"
          v-model="selectedColumnKeys"
        />
      </section>

      <section class="panel-section">
        <h3 class="section-title">Filters</h3>
        <ExportFilterPanel
          :dataset="activeDataset"
          :season-min="seasonMin"
          :season-max="seasonMax"
          :selected-team-name="selectedTeamName"
          :teams="teams"
          :career-stat-type="careerStatType"
          @update:season-min="(v) => { seasonMin = v }"
          @update:season-max="(v) => { seasonMax = v }"
          @update:selected-team-name="(v) => { selectedTeamName = v }"
          @update:career-stat-type="(v) => { careerStatType = v }"
        />
      </section>

      <section class="panel-section">
        <h3 class="section-title">Presets</h3>
        <ExportPresetManager
          :current-config-j-s-o-n="toConfigJSON()"
          :dataset-id="activeDatasetId"
          @load="fromConfigJSON"
        />
      </section>

      <div class="apply-row">
        <Button
          label="Apply"
          icon="pi pi-refresh"
          :loading="isPreviewLoading"
          :disabled="selectedColumnKeys.length === 0"
          class="apply-btn"
          @click="refreshPreview"
        />
      </div>
    </div>

    <div class="right-panel">
      <div class="right-toolbar">
        <div class="sort-controls">
          <select
            class="sort-select"
            :value="sortCol"
            @change="sortCol = ($event.target as HTMLSelectElement).value"
          >
            <option value="">No sort</option>
            <option v-for="col in selectedColumns" :key="col.key" :value="col.key">
              {{ col.label }}
            </option>
          </select>
          <button
            class="sort-dir-btn"
            :title="sortDir === 'asc' ? 'Ascending' : 'Descending'"
            @click="sortDir = sortDir === 'asc' ? 'desc' : 'asc'"
          >
            <i :class="sortDir === 'asc' ? 'pi pi-sort-amount-up-alt' : 'pi pi-sort-amount-down'" />
          </button>
        </div>
        <Button
          label="Export CSV"
          icon="pi pi-download"
          :loading="isExporting"
          :disabled="isExporting || selectedColumnKeys.length === 0"
          @click="downloadCSV"
        />
      </div>
      <ExportPreviewTable
        :selected-columns="selectedColumns"
        :rows="previewRows"
        :loading="isPreviewLoading"
        :total-count="totalCount"
      />
    </div>
  </div>
</template>

<style scoped>
.export-page {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.left-panel {
  width: 300px;
  flex-shrink: 0;
  border-right: 1px solid var(--color-border);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.panel-section {
  padding: 1rem;
  border-bottom: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  flex-shrink: 0;
}

.section-title {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-secondary);
  margin: 0;
}

.apply-row {
  padding: 1rem;
  margin-top: auto;
  border-top: 1px solid var(--color-border);
  flex-shrink: 0;
}

.apply-btn {
  width: 100%;
}

.right-panel {
  flex: 1;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 1rem;
  gap: 0.75rem;
  overflow: hidden;
}

.right-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex-shrink: 0;
}

.sort-controls {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.sort-select {
  padding: 0.25rem 0.375rem;
  font-size: 0.8125rem;
  font-family: inherit;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  cursor: pointer;
}

.sort-dir-btn {
  padding: 0.3rem 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface-2);
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 0.875rem;
  line-height: 1;
}

.sort-dir-btn:hover {
  color: var(--color-text-primary);
}
</style>
