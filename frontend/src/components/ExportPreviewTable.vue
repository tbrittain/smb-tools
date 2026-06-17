<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable, { type DataTablePageEvent } from 'primevue/datatable'
import type { ExportColumnDef } from '../lib/exportDatasets'

defineProps<{
  selectedColumns: ExportColumnDef[]
  rows: Record<string, unknown>[]
  loading: boolean
  totalCount: number
  first: number
}>()

const emit = defineEmits<{
  page: [first: number]
}>()

function onPage(event: DataTablePageEvent) {
  emit('page', event.first)
}

function formatCell(value: unknown, col: ExportColumnDef): string {
  if (value === null || value === undefined) return ''
  if (col.dataType === 'float') {
    const n = Number(value)
    return Number.isNaN(n) ? String(value) : n.toFixed(3)
  }
  return String(value)
}
</script>

<template>
  <div class="preview-table-wrap">
    <DataTable
      :value="rows"
      :loading="loading"
      lazy
      :total-records="totalCount"
      :first="first"
      paginator
      :rows="50"
      size="small"
      scrollable
      scroll-height="flex"
      class="preview-table"
      @page="onPage"
    >
      <template #empty>
        <span class="empty-msg">
          {{ selectedColumns.length === 0 ? 'Select at least one column to preview data.' : 'No data found.' }}
        </span>
      </template>
      <Column
        v-for="col in selectedColumns"
        :key="col.key"
        :field="col.key"
        style="min-width: 90px"
      >
        <template #header>
          <span :title="col.label">{{ col.label }}</span>
        </template>
        <template #body="{ data }">
          {{ formatCell(data[col.key], col) }}
        </template>
      </Column>
    </DataTable>
  </div>
</template>

<style scoped>
.preview-table-wrap {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  gap: 0.5rem;
}

.preview-table {
  flex: 1;
  min-height: 0;
}

.empty-msg {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}
</style>
