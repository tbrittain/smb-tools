<script lang="ts" setup>
import { onMounted } from 'vue'
import type { main } from '../../wailsjs/go/models'
import AppButton from '../components/AppButton.vue'
import ExportColumnSelector from '../components/ExportColumnSelector.vue'
import ExportDatasetPicker from '../components/ExportDatasetPicker.vue'
import ExportFilterPanel from '../components/ExportFilterPanel.vue'
import ExportPresetManager from '../components/ExportPresetManager.vue'
import ExportPreviewTable from '../components/ExportPreviewTable.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'
import { useExportConfig } from '../composables/useExportConfig'
import { EXPORT_DATASETS } from '../lib/exportDatasets'

const {
  activeDatasetId,
  activeDataset,
  onDatasetChange,
  selectedColumnKeys,
  selectedColumns,
  filterRows,
  careerStatType,
  qualifiedOnly,
  columnOptions,
  sortCol,
  sortDir,
  previewRows,
  totalCount,
  isPreviewLoading,
  previewFirst,
  appliedColumns,
  applyAndPreview,
  onPreviewPage,
  isExporting,
  downloadCSV,
  toConfigJSON,
  fromPreset,
} = useExportConfig()

const { set } = useBreadcrumbs()
onMounted(() => set([{ label: 'Stat Explorer' }]))

function setFilterRows(rows: main.FilterRowDTO[]) {
  filterRows.value = rows
}
function setCareerStatType(v: string) {
  careerStatType.value = v
}
function setQualifiedOnly(v: boolean) {
  qualifiedOnly.value = v
}
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
          :filter-rows="filterRows"
          :available-columns="selectedColumns"
          :column-options="columnOptions"
          :career-stat-type="careerStatType"
          :qualified-only="qualifiedOnly"
          @update:filter-rows="setFilterRows"
          @update:career-stat-type="setCareerStatType"
          @update:qualified-only="setQualifiedOnly"
        />
      </section>

      <section class="panel-section">
        <h3 class="section-title">Presets</h3>
        <ExportPresetManager
          :current-config-j-s-o-n="toConfigJSON()"
          :dataset-id="activeDatasetId"
          @load="fromPreset"
        />
      </section>

      <div class="apply-row">
        <AppButton
          icon="pi pi-refresh"
          :loading="isPreviewLoading"
          :disabled="selectedColumnKeys.length === 0"
          class="apply-btn"
          @click="applyAndPreview"
        >
          Apply
        </AppButton>
        <AppButton
          variant="secondary"
          icon="pi pi-download"
          :loading="isExporting"
          :disabled="isExporting || selectedColumnKeys.length === 0"
          class="apply-btn"
          @click="downloadCSV"
        >
          Export CSV
        </AppButton>
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
        <AppButton
          icon="pi pi-download"
          :loading="isExporting"
          :disabled="isExporting || selectedColumnKeys.length === 0"
          @click="downloadCSV"
        >
          Export CSV
        </AppButton>
      </div>
      <ExportPreviewTable
        :selected-columns="appliedColumns"
        :rows="previewRows"
        :loading="isPreviewLoading"
        :total-count="totalCount"
        :first="previewFirst"
        @page="onPreviewPage"
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
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
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
