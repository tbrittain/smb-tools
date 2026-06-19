import type { Meta, StoryObj } from '@storybook/vue3'
import { EXPORT_DATASETS } from '../lib/exportDatasets'
import ExportPreviewTable from './ExportPreviewTable.vue'

const meta: Meta<typeof ExportPreviewTable> = {
  title: 'Components/ExportPreviewTable',
  component: ExportPreviewTable,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ExportPreviewTable>

const BATTING_COLS = EXPORT_DATASETS[0].columns.slice(0, 10)

const SAMPLE_ROWS = [
  {
    player_name: 'Mike Trout',
    first_name: 'Mike',
    last_name: 'Trout',
    season_num: 8,
    team_name: 'Bisons',
    age: 28,
    primary_position: 'CF',
    bat_hand: 'R',
    throw_hand: 'R',
    chemistry_type: 'Competitive',
    _player_id: 101,
    _team_id: 5,
    _team_history_id: 55,
  },
  {
    player_name: 'Jose Ramirez',
    first_name: 'Jose',
    last_name: 'Ramirez',
    season_num: 8,
    team_name: 'Marshals',
    age: 30,
    primary_position: '3B',
    bat_hand: 'L',
    throw_hand: 'R',
    chemistry_type: 'Fun',
    _player_id: 102,
    _team_id: 6,
    _team_history_id: 56,
  },
  {
    player_name: 'Freddie Freeman',
    first_name: 'Freddie',
    last_name: 'Freeman',
    season_num: 8,
    team_name: 'Vipers',
    age: 33,
    primary_position: '1B',
    bat_hand: 'L',
    throw_hand: 'R',
    chemistry_type: 'Competitive',
    _player_id: 103,
    _team_id: 7,
    _team_history_id: 57,
  },
]

export const Populated: Story = {
  args: {
    selectedColumns: BATTING_COLS,
    rows: SAMPLE_ROWS,
    loading: false,
    totalCount: 3,
  },
}

export const Loading: Story = {
  args: {
    selectedColumns: BATTING_COLS,
    rows: [],
    loading: true,
    totalCount: 0,
  },
}

export const EmptyNoColumns: Story = {
  args: {
    selectedColumns: [],
    rows: [],
    loading: false,
    totalCount: 0,
  },
}

export const EmptyNoResults: Story = {
  args: {
    selectedColumns: BATTING_COLS,
    rows: [],
    loading: false,
    totalCount: 0,
  },
}

export const FreeAgentMissingTeamLink: Story = {
  args: {
    selectedColumns: BATTING_COLS,
    rows: [
      {
        player_name: 'Free Agent',
        first_name: 'Free',
        last_name: 'Agent',
        season_num: 8,
        team_name: '',
        age: 25,
        primary_position: 'SS',
        bat_hand: 'R',
        throw_hand: 'R',
        chemistry_type: 'Competitive',
        _player_id: 104,
        // No _team_id/_team_history_id — team_name renders as plain text, not a link.
      },
    ],
    loading: false,
    totalCount: 1,
  },
}

export const TruncatedPreview: Story = {
  args: {
    selectedColumns: BATTING_COLS,
    rows: SAMPLE_ROWS,
    loading: false,
    totalCount: 847,
  },
}

export const AllVariants: Story = {
  render: () => ({
    components: { ExportPreviewTable },
    setup() {
      return { BATTING_COLS, SAMPLE_ROWS }
    },
    template: `
      <div style="display:flex;flex-direction:column;gap:2rem;padding:1.5rem;background:#0d1117">
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">Populated (truncated — 3 of 847)</p>
          <div style="height:200px;display:flex;flex-direction:column">
            <ExportPreviewTable :selected-columns="BATTING_COLS" :rows="SAMPLE_ROWS" :loading="false" :total-count="847" />
          </div>
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">Loading</p>
          <div style="height:200px;display:flex;flex-direction:column">
            <ExportPreviewTable :selected-columns="BATTING_COLS" :rows="[]" :loading="true" :total-count="0" />
          </div>
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">No columns selected</p>
          <div style="height:200px;display:flex;flex-direction:column">
            <ExportPreviewTable :selected-columns="[]" :rows="[]" :loading="false" :total-count="0" />
          </div>
        </div>
      </div>
    `,
  }),
}
