import type { Meta, StoryObj } from '@storybook/vue3'
import { EXPORT_DATASET_MAP, EXPORT_DATASETS } from '../lib/exportDatasets'
import ExportFilterPanel from './ExportFilterPanel.vue'

const meta: Meta<typeof ExportFilterPanel> = {
  title: 'Components/ExportFilterPanel',
  component: ExportFilterPanel,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ExportFilterPanel>

const SAMPLE_TEAMS = [
  { teamName: 'Bisons', teamId: 1 },
  { teamName: 'Marshals', teamId: 2 },
  { teamName: 'Vipers', teamId: 3 },
]

export const SeasonDataset: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.batting_season,
    seasonMin: null,
    seasonMax: null,
    selectedTeamName: '',
    teams: SAMPLE_TEAMS,
    careerStatType: 'regular_season',
  },
}

export const SeasonDatasetWithFilters: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.batting_season,
    seasonMin: 3,
    seasonMax: 8,
    selectedTeamName: 'Bisons',
    teams: SAMPLE_TEAMS,
    careerStatType: 'regular_season',
  },
}

export const CareerDataset: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.career_batting,
    seasonMin: null,
    seasonMax: null,
    selectedTeamName: '',
    teams: SAMPLE_TEAMS,
    careerStatType: 'regular_season',
  },
}

export const CareerDatasetPlayoffs: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.career_batting,
    seasonMin: null,
    seasonMax: null,
    selectedTeamName: '',
    teams: SAMPLE_TEAMS,
    careerStatType: 'playoffs',
  },
}

export const AllVariants: Story = {
  render: () => ({
    components: { ExportFilterPanel },
    setup() {
      return { EXPORT_DATASET_MAP, SAMPLE_TEAMS, EXPORT_DATASETS }
    },
    template: `
      <div style="display:flex;flex-direction:column;gap:2rem;padding:1.5rem;background:#0d1117;width:300px">
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.75rem">Season dataset — no filters applied</p>
          <ExportFilterPanel
            :dataset="EXPORT_DATASET_MAP['batting_season']"
            :season-min="null"
            :season-max="null"
            selected-team-name=""
            :teams="SAMPLE_TEAMS"
            career-stat-type="regular_season"
          />
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.75rem">Career dataset — stat type toggle</p>
          <ExportFilterPanel
            :dataset="EXPORT_DATASET_MAP['career_batting']"
            :season-min="null"
            :season-max="null"
            selected-team-name=""
            :teams="SAMPLE_TEAMS"
            career-stat-type="playoffs"
          />
        </div>
      </div>
    `,
  }),
}
