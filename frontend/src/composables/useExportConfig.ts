import { useToast } from 'primevue/usetoast'
import { computed, onMounted, ref } from 'vue'
import { ExportToCSV, GetTeamsForExport, PreviewExportData } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { EXPORT_DATASET_MAP, EXPORT_DATASETS, type ExportColumnDef } from '../lib/exportDatasets'

export interface ExportPresetConfig {
  columns: string[]
  seasonMin: number | null
  seasonMax: number | null
  selectedTeamName: string
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

  const seasonMin = ref<number | null>(null)
  const seasonMax = ref<number | null>(null)
  const selectedTeamName = ref<string>('')
  const careerStatType = ref<string>('regular_season')
  const sortCol = ref<string>('')
  const sortDir = ref<'asc' | 'desc'>('asc')

  // ── Teams list (loaded once on mount) ────────────────────────────────────────

  const teams = ref<main.TeamPickerResultDTO[]>([])

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
    seasonMin.value = null
    seasonMax.value = null
    selectedTeamName.value = ''
    careerStatType.value = 'regular_season'
    sortCol.value = ''
    sortDir.value = 'asc'
    selectAllColumns()
  }

  // ── Options builder ───────────────────────────────────────────────────────────

  function buildOptions(): main.ExportOptionsDTO {
    const filters: main.FilterRowDTO[] = []
    const ds = activeDataset.value

    if (ds.supportsSeasonFilter) {
      if (seasonMin.value !== null) {
        filters.push(
          new main.FilterRowDTO({ column: 'season_num', op: 'gte', value: String(seasonMin.value), value2: '' }),
        )
      }
      if (seasonMax.value !== null) {
        filters.push(
          new main.FilterRowDTO({ column: 'season_num', op: 'lte', value: String(seasonMax.value), value2: '' }),
        )
      }
    }

    if (ds.supportsTeamFilter && selectedTeamName.value) {
      filters.push(new main.FilterRowDTO({ column: 'team_name', op: 'eq', value: selectedTeamName.value, value2: '' }))
    }

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
      seasonMin: seasonMin.value,
      seasonMax: seasonMax.value,
      selectedTeamName: selectedTeamName.value,
      careerStatType: careerStatType.value,
      sortCol: sortCol.value,
      sortDir: sortDir.value,
    }
    return JSON.stringify(cfg)
  }

  function fromConfigJSON(json: string) {
    try {
      const cfg = JSON.parse(json) as ExportPresetConfig
      selectedColumnKeys.value = cfg.columns ?? activeDataset.value.columns.map((c) => c.key)
      seasonMin.value = cfg.seasonMin ?? null
      seasonMax.value = cfg.seasonMax ?? null
      selectedTeamName.value = cfg.selectedTeamName ?? ''
      careerStatType.value = cfg.careerStatType ?? 'regular_season'
      sortCol.value = cfg.sortCol ?? ''
      sortDir.value = cfg.sortDir ?? 'asc'
    } catch {
      toast.add({ severity: 'error', summary: 'Failed to load preset', life: 4000 })
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
      // non-fatal — team filter just won't show options
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
    seasonMin,
    seasonMax,
    selectedTeamName,
    careerStatType,
    sortCol,
    sortDir,
    // teams
    teams,
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
    fromConfigJSON,
  }
}
