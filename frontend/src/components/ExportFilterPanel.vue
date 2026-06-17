<script lang="ts" setup>
import { computed } from 'vue'
import { main } from '../../wailsjs/go/models'
import type { ExportColumnDef, ExportDatasetDef } from '../lib/exportDatasets'

const props = defineProps<{
  dataset: ExportDatasetDef
  filterRows: main.FilterRowDTO[]
  availableColumns: ExportColumnDef[]
  columnOptions: Record<string, string[]>
  careerStatType: string
}>()

const emit = defineEmits<{
  'update:filterRows': [rows: main.FilterRowDTO[]]
  'update:careerStatType': [value: string]
}>()

// ── Op metadata ───────────────────────────────────────────────────────────────

type Op = { value: string; label: string }

const STRING_OPS: Op[] = [
  { value: 'eq', label: '=' },
  { value: 'neq', label: '≠' },
  { value: 'contains', label: 'contains' },
]
const NUMERIC_OPS: Op[] = [
  { value: 'eq', label: '=' },
  { value: 'neq', label: '≠' },
  { value: 'lt', label: '<' },
  { value: 'lte', label: '≤' },
  { value: 'gt', label: '>' },
  { value: 'gte', label: '≥' },
]
const ENUM_OPS: Op[] = [
  { value: 'eq', label: '=' },
  { value: 'neq', label: '≠' },
]

function opsForType(dataType: string): Op[] {
  if (dataType === 'enum') return ENUM_OPS
  if (dataType === 'int' || dataType === 'float') return NUMERIC_OPS
  return STRING_OPS
}

function defaultOpForType(dataType: string): string {
  return opsForType(dataType)[0].value
}

// ── Column lookup ─────────────────────────────────────────────────────────────

const columnMap = computed<Record<string, ExportColumnDef>>(() =>
  Object.fromEntries(props.availableColumns.map((c) => [c.key, c])),
)

function colFor(key: string): ExportColumnDef | undefined {
  return columnMap.value[key]
}

function enumOptionsFor(col: ExportColumnDef): string[] {
  return (col.options as string[] | undefined) ?? props.columnOptions[col.key] ?? []
}

// ── Mutation helpers ──────────────────────────────────────────────────────────

function addRow() {
  if (props.availableColumns.length === 0) return
  const firstCol = props.availableColumns[0]
  emit('update:filterRows', [
    ...props.filterRows,
    new main.FilterRowDTO({ column: firstCol.key, op: defaultOpForType(firstCol.dataType), value: '', value2: '' }),
  ])
}

function removeRow(index: number) {
  emit(
    'update:filterRows',
    props.filterRows.filter((_, i) => i !== index),
  )
}

function updateColumn(index: number, columnKey: string) {
  const col = colFor(columnKey)
  const newOp = col ? defaultOpForType(col.dataType) : 'eq'
  const updated = props.filterRows.map((r, i) =>
    i === index ? new main.FilterRowDTO({ column: columnKey, op: newOp, value: '', value2: '' }) : r,
  )
  emit('update:filterRows', updated)
}

function updateOp(index: number, op: string) {
  const updated = props.filterRows.map((r, i) =>
    i === index ? new main.FilterRowDTO({ column: r.column, op, value: r.value, value2: r.value2 }) : r,
  )
  emit('update:filterRows', updated)
}

function updateValue(index: number, value: string) {
  const updated = props.filterRows.map((r, i) =>
    i === index ? new main.FilterRowDTO({ column: r.column, op: r.op, value, value2: r.value2 }) : r,
  )
  emit('update:filterRows', updated)
}
</script>

<template>
  <div class="filter-panel">
    <!-- Career stat type toggle — dedicated control, not a generic filter row -->
    <div v-if="dataset.supportsCareerStatType" class="filter-section">
      <span class="filter-label">Stat Type</span>
      <div class="toggle-group">
        <button
          class="toggle-btn"
          :class="{ active: !careerStatType || careerStatType === 'regular_season' }"
          @click="$emit('update:careerStatType', 'regular_season')"
        >
          Reg Season
        </button>
        <button
          class="toggle-btn"
          :class="{ active: careerStatType === 'playoffs' }"
          @click="$emit('update:careerStatType', 'playoffs')"
        >
          Playoffs
        </button>
        <button
          class="toggle-btn"
          :class="{ active: careerStatType === 'total_career' }"
          @click="$emit('update:careerStatType', 'total_career')"
        >
          Total
        </button>
      </div>
    </div>

    <!-- Dynamic filter rows -->
    <div v-if="filterRows.length > 0" class="filter-rows">
      <div v-for="(row, idx) in filterRows" :key="idx" class="filter-row">
        <!-- Column picker -->
        <select
          class="row-select row-col"
          :value="row.column"
          @change="updateColumn(idx, ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="col in availableColumns" :key="col.key" :value="col.key">
            {{ col.label }}
          </option>
        </select>

        <!-- Op picker -->
        <select
          class="row-select row-op"
          :value="row.op"
          @change="updateOp(idx, ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="op in opsForType(colFor(row.column)?.dataType ?? 'string')" :key="op.value" :value="op.value">
            {{ op.label }}
          </option>
        </select>

        <!-- Value input — varies by column type -->
        <template v-if="colFor(row.column)?.dataType === 'enum'">
          <select
            class="row-select row-val"
            :value="row.value"
            @change="updateValue(idx, ($event.target as HTMLSelectElement).value)"
          >
            <option value="">— select —</option>
            <option
              v-for="opt in enumOptionsFor(colFor(row.column)!)"
              :key="opt"
              :value="opt"
            >
              {{ opt }}
            </option>
          </select>
        </template>
        <template v-else-if="colFor(row.column)?.dataType === 'int'">
          <input
            type="number"
            step="1"
            class="row-input row-val"
            :value="row.value"
            @change="updateValue(idx, ($event.target as HTMLInputElement).value)"
          />
        </template>
        <template v-else-if="colFor(row.column)?.dataType === 'float'">
          <input
            type="number"
            step="any"
            class="row-input row-val"
            :value="row.value"
            @change="updateValue(idx, ($event.target as HTMLInputElement).value)"
          />
        </template>
        <template v-else>
          <input
            type="text"
            class="row-input row-val"
            :value="row.value"
            @change="updateValue(idx, ($event.target as HTMLInputElement).value)"
          />
        </template>

        <!-- Remove button -->
        <button class="remove-btn" title="Remove filter" @click="removeRow(idx)">×</button>
      </div>
    </div>

    <p
      v-if="filterRows.length === 0 && availableColumns.length === 0 && !dataset.supportsCareerStatType"
      class="no-filters"
    >
      No columns selected.
    </p>

    <button
      class="add-filter-btn"
      :disabled="availableColumns.length === 0"
      @click="addRow"
    >
      + Add Filter
    </button>
  </div>
</template>

<style scoped>
.filter-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.filter-section {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
  min-width: 64px;
}

/* careerStatType toggle */
.toggle-group {
  display: flex;
}

.toggle-btn {
  padding: 0.2rem 0.5rem;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.75rem;
  font-family: inherit;
  cursor: pointer;
  border-radius: 0;
  transition:
    background 0.1s,
    color 0.1s;
}

.toggle-btn:first-child {
  border-radius: 4px 0 0 4px;
}

.toggle-btn:last-child {
  border-radius: 0 4px 4px 0;
}

.toggle-btn + .toggle-btn {
  margin-left: -1px;
}

.toggle-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.toggle-btn.active {
  background: var(--color-surface-2);
  border-color: var(--color-accent);
  color: var(--color-accent);
  position: relative;
  z-index: 1;
}

/* Filter rows */
.filter-rows {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.row-select,
.row-input {
  padding: 0.25rem 0.375rem;
  font-size: 0.75rem;
  font-family: inherit;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  min-width: 0;
}

.row-select:focus,
.row-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.row-col {
  flex: 2;
}

.row-op {
  flex: 1;
}

.row-val {
  flex: 2;
}

.remove-btn {
  flex-shrink: 0;
  padding: 0.2rem 0.4rem;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  border-radius: 4px;
  line-height: 1;
}

.remove-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-error, #dc2626);
}

.add-filter-btn {
  align-self: flex-start;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-family: inherit;
  border: 1px dashed var(--color-border);
  border-radius: 4px;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: color 0.1s, border-color 0.1s;
}

.add-filter-btn:hover:not(:disabled) {
  color: var(--color-accent);
  border-color: var(--color-accent);
}

.add-filter-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.no-filters {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  font-style: italic;
}
</style>
