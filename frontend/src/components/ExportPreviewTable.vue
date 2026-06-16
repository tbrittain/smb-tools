<script lang="ts" setup>
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import type { ExportColumnDef } from '../lib/exportDatasets'

defineProps<{
  selectedColumns: ExportColumnDef[]
  rows: Record<string, unknown>[]
  loading: boolean
  totalCount: number
}>()

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
      size="small"
      scrollable
      scroll-height="flex"
      class="preview-table"
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
    <div v-if="totalCount > 0" class="preview-caption">
      <template v-if="rows.length < totalCount">
        Showing {{ rows.length }} of {{ totalCount }} rows
      </template>
      <template v-else>
        {{ totalCount }} row{{ totalCount === 1 ? '' : 's' }}
      </template>
    </div>
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

.preview-caption {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-align: right;
  padding-right: 0.25rem;
}
</style>
