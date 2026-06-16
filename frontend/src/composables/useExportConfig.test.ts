import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { EXPORT_DATASETS } from '../lib/exportDatasets'
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

function mountComposable() {
  let result: ReturnType<typeof useExportConfig> | undefined
  const component = defineComponent({
    setup() {
      result = useExportConfig()
      return () => h('div')
    },
  })
  mount(component)
  return result!
}

describe('useExportConfig', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    mockPreviewExportData.mockClear()
    mockGetTeamsForExport.mockClear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('onDatasetChange resets selectedColumnKeys to all columns for the new dataset', async () => {
    const cfg = mountComposable()
    await nextTick()

    cfg.activeDatasetId.value = 'pitching_season'
    cfg.onDatasetChange()

    const expectedKeys = EXPORT_DATASETS.find((d) => d.id === 'pitching_season')!.columns.map((c) => c.key)
    expect(cfg.selectedColumnKeys.value).toEqual(expectedKeys)
  })

  it('schedulePreview skips PreviewExportData call when no columns are selected', async () => {
    const cfg = mountComposable()
    await nextTick()
    vi.runAllTimers()
    await nextTick()
    mockPreviewExportData.mockClear()

    cfg.selectedColumnKeys.value = []
    cfg.schedulePreview()
    vi.runAllTimers()
    await nextTick()

    expect(mockPreviewExportData).not.toHaveBeenCalled()
  })

  it('buildOptions emits season range filter rows for seasonMin and seasonMax', async () => {
    const cfg = mountComposable()
    await nextTick()
    vi.runAllTimers()
    await nextTick()
    mockPreviewExportData.mockClear()

    cfg.seasonMin.value = 3
    cfg.seasonMax.value = 8
    cfg.schedulePreview()
    vi.runAllTimers()
    await nextTick()
    await nextTick()

    expect(mockPreviewExportData).toHaveBeenCalledOnce()
    const opts = mockPreviewExportData.mock.calls[0][0]
    expect(opts.filters).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ column: 'season_num', op: 'gte', value: '3' }),
        expect.objectContaining({ column: 'season_num', op: 'lte', value: '8' }),
      ]),
    )
  })

  it('buildOptions passes careerStatType only for career datasets, empty string otherwise', async () => {
    const cfg = mountComposable()
    await nextTick()
    vi.runAllTimers()
    await nextTick()
    mockPreviewExportData.mockClear()

    // Default dataset is batting_season — careerStatType must be empty in options
    cfg.schedulePreview()
    vi.runAllTimers()
    await nextTick()
    await nextTick()

    const seasonOpts = mockPreviewExportData.mock.calls[0][0]
    expect(seasonOpts.careerStatType).toBe('')

    // Switch to career_batting — careerStatType should be forwarded
    mockPreviewExportData.mockClear()
    cfg.activeDatasetId.value = 'career_batting'
    cfg.onDatasetChange()
    vi.runAllTimers()
    await nextTick()
    await nextTick()

    const careerOpts = mockPreviewExportData.mock.calls[0][0]
    expect(careerOpts.careerStatType).toBe('regular_season')
  })
})
