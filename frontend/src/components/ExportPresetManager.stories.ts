import type { Meta, StoryObj } from '@storybook/vue3'
import ExportPresetManager from './ExportPresetManager.vue'

const meta: Meta<typeof ExportPresetManager> = {
  title: 'Components/ExportPresetManager',
  component: ExportPresetManager,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ExportPresetManager>

const SAMPLE_CONFIG = JSON.stringify({
  columns: ['player_name', 'home_runs', 'ops_plus', 'smb_war'],
  seasonMin: 3,
  seasonMax: 8,
  selectedTeamName: '',
  careerStatType: 'regular_season',
  sortCol: 'smb_war',
  sortDir: 'desc',
})

const SAMPLE_PRESETS = [
  {
    id: 'abc123',
    name: 'HR Leaders S3–8',
    datasetId: 'batting_season',
    configJson: SAMPLE_CONFIG,
    createdAt: '2025-06-10T12:00:00Z',
  },
  {
    id: 'def456',
    name: 'Full career pitching',
    datasetId: 'career_pitching',
    configJson: '{}',
    createdAt: '2025-06-08T09:00:00Z',
  },
]

function makeStory(presets: object[]): Story {
  return {
    render: (args) => ({
      components: { ExportPresetManager },
      setup() {
        // @ts-expect-error – deliberate window stub for Storybook
        window.go = {
          main: {
            App: {
              GetExportPresets: () => Promise.resolve(presets),
              SaveExportPreset: (_name: string, _datasetId: string, configJson: string) =>
                Promise.resolve({
                  id: `new-${Date.now()}`,
                  name: _name,
                  datasetId: _datasetId,
                  configJson,
                  createdAt: new Date().toISOString(),
                }),
              DeleteExportPreset: () => Promise.resolve(),
            },
          },
        }
        return { args }
      },
      template: '<ExportPresetManager v-bind="args" @load="() => {}" />',
    }),
    args: {
      currentConfigJSON: SAMPLE_CONFIG,
      datasetId: 'batting_season',
    },
  }
}

export const EmptyState: Story = makeStory([])

export const WithPresets: Story = makeStory(SAMPLE_PRESETS)

export const AllVariants: Story = {
  render: () => ({
    components: { ExportPresetManager },
    setup() {
      // @ts-expect-error – deliberate window stub for Storybook
      window.go = {
        main: {
          App: {
            GetExportPresets: () => Promise.resolve(SAMPLE_PRESETS),
            SaveExportPreset: (_n: string, _d: string, c: string) =>
              Promise.resolve({ id: 'x', name: _n, datasetId: _d, configJson: c, createdAt: '' }),
            DeleteExportPreset: () => Promise.resolve(),
          },
        },
      }
      return { SAMPLE_CONFIG }
    },
    template: `
      <div style="display:flex;gap:2rem;padding:1.5rem;background:#0d1117;align-items:flex-start">
        <div style="width:260px">
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">With presets</p>
          <ExportPresetManager :current-config-j-s-o-n="SAMPLE_CONFIG" dataset-id="batting_season" @load="() => {}" />
        </div>
      </div>
    `,
  }),
}
