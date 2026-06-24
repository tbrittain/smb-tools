import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import { EXPORT_DATASET_MAP } from '../lib/exportDatasets'
import ExportFilterPanel from './ExportFilterPanel.vue'

const meta: Meta<typeof ExportFilterPanel> = {
  title: 'Export/ExportFilterPanel',
  component: ExportFilterPanel,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ExportFilterPanel>

const BATTING_COLS = EXPORT_DATASET_MAP.batting_season.columns
const CAREER_COLS = EXPORT_DATASET_MAP.career_batting.columns
const STANDINGS_COLS = EXPORT_DATASET_MAP.standings.columns

const SAMPLE_COLUMN_OPTIONS = {
  team_name: ['Bisons', 'Marshals', 'Vipers', 'Hammers'],
  conference_name: ['East', 'West'],
  division_name: ['North', 'South'],
}

const makeRow = (column: string, op: string, value: string) => new main.FilterRowDTO({ column, op, value, value2: '' })

export const EmptyFilterList: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.batting_season,
    filterRows: [],
    availableColumns: BATTING_COLS,
    columnOptions: SAMPLE_COLUMN_OPTIONS,
    careerStatType: 'regular_season',
    qualifiedOnly: false,
  },
}

export const WithSeasonFilter: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.batting_season,
    filterRows: [makeRow('season_num', 'gte', '3')],
    availableColumns: BATTING_COLS,
    columnOptions: SAMPLE_COLUMN_OPTIONS,
    careerStatType: 'regular_season',
    qualifiedOnly: false,
  },
}

export const WithTeamFilter: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.batting_season,
    filterRows: [makeRow('team_name', 'eq', 'Bisons')],
    availableColumns: BATTING_COLS,
    columnOptions: SAMPLE_COLUMN_OPTIONS,
    careerStatType: 'regular_season',
    qualifiedOnly: false,
  },
}

export const WithMultipleRows: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.batting_season,
    filterRows: [makeRow('season_num', 'gte', '3'), makeRow('home_runs', 'gt', '20')],
    availableColumns: BATTING_COLS,
    columnOptions: SAMPLE_COLUMN_OPTIONS,
    careerStatType: 'regular_season',
    qualifiedOnly: false,
  },
}

export const CareerDataset: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.career_batting,
    filterRows: [],
    availableColumns: CAREER_COLS,
    columnOptions: {},
    careerStatType: 'regular_season',
    qualifiedOnly: false,
  },
}

export const CareerDatasetPlayoffs: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.career_batting,
    filterRows: [makeRow('smb_war', 'gt', '10')],
    availableColumns: CAREER_COLS,
    columnOptions: {},
    careerStatType: 'playoffs',
    qualifiedOnly: false,
  },
}

export const QualifiedOnlyActive: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.batting_season,
    filterRows: [],
    availableColumns: BATTING_COLS,
    columnOptions: SAMPLE_COLUMN_OPTIONS,
    careerStatType: 'regular_season',
    qualifiedOnly: true,
  },
}

export const StandingsWithEnumFilters: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.standings,
    filterRows: [makeRow('conference_name', 'eq', 'East'), makeRow('wins', 'gte', '50')],
    availableColumns: STANDINGS_COLS,
    columnOptions: SAMPLE_COLUMN_OPTIONS,
    careerStatType: 'regular_season',
    qualifiedOnly: false,
  },
}

export const NoColumnsSelected: Story = {
  args: {
    dataset: EXPORT_DATASET_MAP.batting_season,
    filterRows: [],
    availableColumns: [],
    columnOptions: SAMPLE_COLUMN_OPTIONS,
    careerStatType: 'regular_season',
    qualifiedOnly: false,
  },
}

export const AllVariants: Story = {
  render: () => ({
    components: { ExportFilterPanel },
    setup() {
      const row1 = new main.FilterRowDTO({ column: 'season_num', op: 'gte', value: '3', value2: '' })
      const row2 = new main.FilterRowDTO({ column: 'team_name', op: 'eq', value: 'Bisons', value2: '' })
      const row3 = new main.FilterRowDTO({ column: 'smb_war', op: 'gt', value: '2', value2: '' })
      return {
        EXPORT_DATASET_MAP,
        BATTING_COLS,
        CAREER_COLS,
        SAMPLE_COLUMN_OPTIONS,
        row1,
        row2,
        row3,
      }
    },
    template: `
      <div style="display:flex;flex-direction:column;gap:2rem;padding:1.5rem;background:#0d1117;width:300px">
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.75rem">Season — no filters</p>
          <ExportFilterPanel
            :dataset="EXPORT_DATASET_MAP['batting_season']"
            :filter-rows="[]"
            :available-columns="BATTING_COLS"
            :column-options="SAMPLE_COLUMN_OPTIONS"
            career-stat-type="regular_season"
            :qualified-only="false"
            @update:filter-rows="() => {}"
            @update:career-stat-type="() => {}"
            @update:qualified-only="() => {}"
          />
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.75rem">Season — season + team filters + qualified only</p>
          <ExportFilterPanel
            :dataset="EXPORT_DATASET_MAP['batting_season']"
            :filter-rows="[row1, row2]"
            :available-columns="BATTING_COLS"
            :column-options="SAMPLE_COLUMN_OPTIONS"
            career-stat-type="regular_season"
            :qualified-only="true"
            @update:filter-rows="() => {}"
            @update:career-stat-type="() => {}"
            @update:qualified-only="() => {}"
          />
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.75rem">Career — playoffs + smbWAR filter</p>
          <ExportFilterPanel
            :dataset="EXPORT_DATASET_MAP['career_batting']"
            :filter-rows="[row3]"
            :available-columns="CAREER_COLS"
            :column-options="{}"
            career-stat-type="playoffs"
            :qualified-only="false"
            @update:filter-rows="() => {}"
            @update:career-stat-type="() => {}"
            @update:qualified-only="() => {}"
          />
        </div>
      </div>
    `,
  }),
}
