import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { main } from '../../wailsjs/go/models'
import { EXPORT_DATASET_MAP, EXPORT_DATASETS } from '../lib/exportDatasets'
import { useExportConfig } from './useExportConfig'

const mockPreviewExportData = vi.fn().mockResolvedValue({ rows: [], totalCount: 0 })
const mockGetTeamsForExport = vi.fn().mockResolvedValue([])

vi.mock('../../wailsjs/go/main/App', () => ({
  GetTeamsForExport: () => mockGetTeamsForExport(),
  PreviewExportData: (opts: unknown) => mockPreviewExportData(opts),
  ExportToCSV: vi.fn().mockResolvedValue(''),
}))

vi.mock('primevue/usetoast', () => ({
  useToast: () => ({ add: vi.fn() }),
}))

function mountComposable(): ReturnType<typeof useExportConfig> {
  const captured: { value: ReturnType<typeof useExportConfig> | undefined } = { value: undefined }
  mount(
    defineComponent({
      setup() {
        captured.value = useExportConfig()
        return () => h('div')
      },
    }),
  )
  if (captured.value === undefined) throw new Error('composable did not run in setup')
  return captured.value
}

describe('useExportConfig', () => {
  beforeEach(() => {
    mockPreviewExportData.mockClear()
    mockGetTeamsForExport.mockClear()
  })

  it('onDatasetChange resets selectedColumnKeys to all columns for the new dataset', async () => {
    const cfg = mountComposable()
    await nextTick()

    cfg.activeDatasetId.value = 'pitching_season'
    cfg.onDatasetChange()

    const expectedKeys = EXPORT_DATASETS.find((d) => d.id === 'pitching_season')?.columns.map((c) => c.key)
    expect(cfg.selectedColumnKeys.value).toEqual(expectedKeys)
  })

  it('onDatasetChange clears filterRows', async () => {
    const cfg = mountComposable()
    await nextTick()

    cfg.filterRows.value = [new main.FilterRowDTO({ column: 'season_num', op: 'gte', value: '3', value2: '' })]
    cfg.onDatasetChange()

    expect(cfg.filterRows.value).toEqual([])
  })

  it('refreshPreview skips PreviewExportData call when no columns are selected', async () => {
    const cfg = mountComposable()
    await nextTick()
    mockPreviewExportData.mockClear()

    cfg.selectedColumnKeys.value = []
    await cfg.refreshPreview()

    expect(mockPreviewExportData).not.toHaveBeenCalled()
  })

  it('buildOptions passes non-empty filterRows to PreviewExportData', async () => {
    const cfg = mountComposable()
    await nextTick()
    mockPreviewExportData.mockClear()

    cfg.filterRows.value = [new main.FilterRowDTO({ column: 'home_runs', op: 'gt', value: '20', value2: '' })]
    await cfg.refreshPreview()

    expect(mockPreviewExportData).toHaveBeenCalledOnce()
    const opts = mockPreviewExportData.mock.calls[0][0]
    expect(opts.filters).toEqual(
      expect.arrayContaining([expect.objectContaining({ column: 'home_runs', op: 'gt', value: '20' })]),
    )
  })

  it('buildOptions skips filterRows with empty value field', async () => {
    const cfg = mountComposable()
    await nextTick()
    mockPreviewExportData.mockClear()

    cfg.filterRows.value = [
      new main.FilterRowDTO({ column: 'team_name', op: 'eq', value: '', value2: '' }),
      new main.FilterRowDTO({ column: 'home_runs', op: 'gt', value: '10', value2: '' }),
    ]
    await cfg.refreshPreview()

    expect(mockPreviewExportData).toHaveBeenCalledOnce()
    const opts = mockPreviewExportData.mock.calls[0][0]
    expect(opts.filters).toHaveLength(1)
    expect(opts.filters[0]).toMatchObject({ column: 'home_runs', op: 'gt', value: '10' })
  })

  it('orphaned filterRows are dropped when selectedColumnKeys shrinks', async () => {
    const cfg = mountComposable()
    await nextTick()

    // Ensure both columns are selected
    const battingDataset = EXPORT_DATASET_MAP.batting_season
    const hrKey = battingDataset.columns.find((c) => c.key === 'home_runs')?.key ?? 'home_runs'
    const abKey = battingDataset.columns.find((c) => c.key === 'at_bats')?.key ?? 'at_bats'
    cfg.selectedColumnKeys.value = [hrKey, abKey]

    cfg.filterRows.value = [
      new main.FilterRowDTO({ column: hrKey, op: 'gt', value: '20', value2: '' }),
      new main.FilterRowDTO({ column: abKey, op: 'gte', value: '100', value2: '' }),
    ]

    // Deselect the home_runs column
    cfg.selectedColumnKeys.value = [abKey]
    await nextTick()

    // The home_runs filter row should be dropped; at_bats row survives
    expect(cfg.filterRows.value).toHaveLength(1)
    expect(cfg.filterRows.value[0].column).toBe(abKey)
  })

  it('fromConfigJSON with legacy preset (no filters field) defaults filterRows to []', async () => {
    const cfg = mountComposable()
    await nextTick()

    cfg.filterRows.value = [new main.FilterRowDTO({ column: 'season_num', op: 'gte', value: '2', value2: '' })]

    // Simulate a legacy preset JSON that has no filters field
    const legacyJson = JSON.stringify({
      columns: cfg.selectedColumnKeys.value,
      careerStatType: 'regular_season',
      sortCol: '',
      sortDir: 'asc',
    })

    cfg.fromPreset('batting_season', legacyJson)

    expect(cfg.filterRows.value).toEqual([])
  })

  it('buildOptions passes careerStatType only for datasets with a stat-type toggle, empty string otherwise', async () => {
    const cfg = mountComposable()
    await nextTick()
    mockPreviewExportData.mockClear()

    // Default dataset is batting_season (statTypeOptions: 'season') — careerStatType is forwarded
    await cfg.refreshPreview()

    const seasonOpts = mockPreviewExportData.mock.calls[0][0]
    expect(seasonOpts.careerStatType).toBe('regular_season')

    // Switch to standings (statTypeOptions: 'none') — careerStatType must be empty
    mockPreviewExportData.mockClear()
    cfg.activeDatasetId.value = 'standings'
    cfg.onDatasetChange()
    await cfg.refreshPreview()

    const standingsOpts = mockPreviewExportData.mock.calls[0][0]
    expect(standingsOpts.careerStatType).toBe('')

    // Switch to career_batting (statTypeOptions: 'career') — careerStatType should be forwarded
    mockPreviewExportData.mockClear()
    cfg.activeDatasetId.value = 'career_batting'
    cfg.onDatasetChange()
    await cfg.refreshPreview()

    const careerOpts = mockPreviewExportData.mock.calls[0][0]
    expect(careerOpts.careerStatType).toBe('regular_season')
  })

  it('appliedColumns only updates when a fetch actually runs, not on every checkbox toggle', async () => {
    const cfg = mountComposable()
    await nextTick()

    const battingDataset = EXPORT_DATASET_MAP.batting_season
    const hrKey = battingDataset.columns.find((c) => c.key === 'home_runs')?.key ?? 'home_runs'

    // Initial mount already ran a fetch with all columns selected.
    const initialKeys = cfg.appliedColumns.value.map((c) => c.key)
    expect(initialKeys).toEqual(cfg.selectedColumnKeys.value)

    // Deselecting a column updates the live selection immediately, but must NOT
    // change what the preview table renders until a fetch actually runs.
    cfg.selectedColumnKeys.value = cfg.selectedColumnKeys.value.filter((k) => k !== hrKey)
    await nextTick()
    expect(cfg.appliedColumns.value.map((c) => c.key)).toEqual(initialKeys)

    // Clicking Apply runs the fetch and snapshots the new column selection.
    cfg.applyAndPreview()
    await nextTick()
    expect(cfg.appliedColumns.value.map((c) => c.key)).toEqual(cfg.selectedColumnKeys.value)
    expect(cfg.appliedColumns.value.some((c) => c.key === hrKey)).toBe(false)
  })
})
