import { useToast } from 'primevue/usetoast'
import { computed, onMounted, ref, watch } from 'vue'
import { ExportToCSV, GetTeamsForExport, PreviewExportData } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { EXPORT_DATASET_MAP, EXPORT_DATASETS, type ExportColumnDef } from '../lib/exportDatasets'

export interface ExportPresetConfig {
  columns: string[]
  filters: Array<{ column: string; op: string; value: string; value2: string }>
  careerStatType: string
  sortCol: string
  sortDir: 'asc' | 'desc'
}

export function useExportConfig() {
  const toast = useToast()

  // ── Dataset ───────────────────────────────────────────────────────────────────

  const activeDatasetId = ref<string>(EXPORT_DATASETS[0].id)
  const activeDataset = computed(() => EXPORT_DATASET_MAP[activeDatasetId.value])

  // ── Column selection ─────────────────────────────────────────────────────────

  const selectedColumnKeys = ref<string[]>([])

  const selectedColumns = computed<ExportColumnDef[]>(() =>
    activeDataset.value.columns.filter((c) => selectedColumnKeys.value.includes(c.key)),
  )

  function selectAllColumns() {
    selectedColumnKeys.value = activeDataset.value.columns.map((c) => c.key)
  }

  // ── Filters ───────────────────────────────────────────────────────────────────

  const filterRows = ref<main.FilterRowDTO[]>([])
  const careerStatType = ref<string>('regular_season')
  const sortCol = ref<string>('')
  const sortDir = ref<'asc' | 'desc'>('asc')

  // Drop filter rows whose column was deselected from the column picker.
  watch(selectedColumnKeys, (keys) => {
    filterRows.value = filterRows.value.filter((r) => keys.includes(r.column))
  })

  // ── Teams list + dynamic column options ──────────────────────────────────────

  const teams = ref<main.TeamPickerResultDTO[]>([])

  const columnOptions = computed<Record<string, string[]>>(() => ({
    team_name: teams.value.map((t) => t.teamName),
    conference_name: [...new Set(teams.value.map((t) => t.conferenceName))].filter(Boolean).sort(),
    division_name: [...new Set(teams.value.map((t) => t.divisionName))].filter(Boolean).sort(),
  }))

  // ── Preview state ─────────────────────────────────────────────────────────────

  const previewRows = ref<Record<string, unknown>[]>([])
  const totalCount = ref<number>(0)
  const isPreviewLoading = ref<boolean>(false)

  // ── Export state ─────────────────────────────────────────────────────────────

  const isExporting = ref<boolean>(false)

  // ── Preview ───────────────────────────────────────────────────────────────────

  async function refreshPreview() {
    if (selectedColumnKeys.value.length === 0) {
      previewRows.value = []
      totalCount.value = 0
      return
    }
    isPreviewLoading.value = true
    try {
      const result = await PreviewExportData(buildOptions())
      previewRows.value = result.rows ?? []
      totalCount.value = result.totalCount
    } catch (e) {
      toast.add({ severity: 'error', summary: String(e), life: 5000 })
    } finally {
      isPreviewLoading.value = false
    }
  }

  // ── Dataset change ────────────────────────────────────────────────────────────

  function onDatasetChange() {
    filterRows.value = []
    careerStatType.value = 'regular_season'
    sortCol.value = ''
    sortDir.value = 'asc'
    selectAllColumns()
    refreshPreview()
  }

  // ── Options builder ───────────────────────────────────────────────────────────

  function buildOptions(): main.ExportOptionsDTO {
    const ds = activeDataset.value
    const filters = filterRows.value
      .filter((r) => r.value !== '')
      .map((r) => new main.FilterRowDTO({ column: r.column, op: r.op, value: r.value, value2: '' }))

    return new main.ExportOptionsDTO({
      datasetId: activeDatasetId.value,
      columns: selectedColumnKeys.value,
      filters,
      sortCol: sortCol.value,
      sortDir: sortDir.value,
      careerStatType: ds.supportsCareerStatType ? careerStatType.value : '',
    })
  }

  // ── Preset serialisation ──────────────────────────────────────────────────────

  function toConfigJSON(): string {
    const cfg: ExportPresetConfig = {
      columns: selectedColumnKeys.value,
      filters: filterRows.value.map((r) => ({
        column: r.column,
        op: r.op,
        value: r.value,
        value2: r.value2,
      })),
      careerStatType: careerStatType.value,
      sortCol: sortCol.value,
      sortDir: sortDir.value,
    }
    return JSON.stringify(cfg)
  }

  function fromConfigJSON(json: string): boolean {
    try {
      const cfg = JSON.parse(json) as ExportPresetConfig
      selectedColumnKeys.value = cfg.columns ?? activeDataset.value.columns.map((c) => c.key)
      filterRows.value = (cfg.filters ?? []).map(
        (f) => new main.FilterRowDTO({ column: f.column, op: f.op, value: f.value, value2: f.value2 ?? '' }),
      )
      careerStatType.value = cfg.careerStatType ?? 'regular_season'
      sortCol.value = cfg.sortCol ?? ''
      sortDir.value = cfg.sortDir ?? 'asc'
      return true
    } catch {
      toast.add({ severity: 'error', summary: 'Failed to load preset', life: 4000 })
      return false
    }
  }

  function fromPreset(datasetId: string, configJSON: string) {
    activeDatasetId.value = datasetId
    if (fromConfigJSON(configJSON)) {
      refreshPreview()
    }
  }

  // ── CSV download ──────────────────────────────────────────────────────────────

  async function downloadCSV() {
    isExporting.value = true
    try {
      const b64 = await ExportToCSV(buildOptions())
      const bytes = atob(b64)
      const arr = new Uint8Array(bytes.length)
      for (let i = 0; i < bytes.length; i++) arr[i] = bytes.charCodeAt(i)
      const blob = new Blob([arr], { type: 'text/csv' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${activeDatasetId.value}_export.csv`
      a.click()
      URL.revokeObjectURL(url)
      toast.add({ severity: 'success', summary: 'CSV exported', life: 3000 })
    } catch (e) {
      toast.add({ severity: 'error', summary: String(e), life: 5000 })
    } finally {
      isExporting.value = false
    }
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────────

  onMounted(async () => {
    try {
      teams.value = await GetTeamsForExport()
    } catch {
      // non-fatal — enum options for team/conference/division just won't populate
    }
    selectAllColumns()
    await refreshPreview()
  })

  return {
    // dataset
    activeDatasetId,
    activeDataset,
    onDatasetChange,
    // columns
    selectedColumnKeys,
    selectedColumns,
    selectAllColumns,
    // filters
    filterRows,
    careerStatType,
    columnOptions,
    sortCol,
    sortDir,
    // preview
    previewRows,
    totalCount,
    isPreviewLoading,
    refreshPreview,
    // export
    isExporting,
    downloadCSV,
    // presets
    toConfigJSON,
    fromPreset,
  }
}
