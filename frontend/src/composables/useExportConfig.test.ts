import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
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

  it('refreshPreview skips PreviewExportData call when no columns are selected', async () => {
    const cfg = mountComposable()
    await nextTick()
    mockPreviewExportData.mockClear()

    cfg.selectedColumnKeys.value = []
    await cfg.refreshPreview()

    expect(mockPreviewExportData).not.toHaveBeenCalled()
  })

  it('buildOptions emits season range filter rows for seasonMin and seasonMax', async () => {
    const cfg = mountComposable()
    await nextTick()
    mockPreviewExportData.mockClear()

    cfg.seasonMin.value = 3
    cfg.seasonMax.value = 8
    await cfg.refreshPreview()

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
    mockPreviewExportData.mockClear()

    // Default dataset is batting_season — careerStatType must be empty in options
    await cfg.refreshPreview()

    const seasonOpts = mockPreviewExportData.mock.calls[0][0]
    expect(seasonOpts.careerStatType).toBe('')

    // Switch to career_batting — careerStatType should be forwarded
    mockPreviewExportData.mockClear()
    cfg.activeDatasetId.value = 'career_batting'
    cfg.onDatasetChange()
    await cfg.refreshPreview()

    const careerOpts = mockPreviewExportData.mock.calls[0][0]
    expect(careerOpts.careerStatType).toBe('regular_season')
  })
})
