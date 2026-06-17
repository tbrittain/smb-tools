import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import { EXPORT_DATASETS } from '../lib/exportDatasets'
import ExportColumnSelector from './ExportColumnSelector.vue'

const meta: Meta<typeof ExportColumnSelector> = {
  title: 'Components/ExportColumnSelector',
  component: ExportColumnSelector,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ExportColumnSelector>

const BATTING_COLS = EXPORT_DATASETS[0].columns

export const AllSelected: Story = {
  render: (args) => ({
    components: { ExportColumnSelector },
    setup() {
      const selected = ref(BATTING_COLS.map((c) => c.key))
      return { args, selected, BATTING_COLS }
    },
    template: '<ExportColumnSelector :columns="BATTING_COLS" v-model="selected" />',
  }),
}

export const PartialSelection: Story = {
  render: (args) => ({
    components: { ExportColumnSelector },
    setup() {
      const selected = ref(['player_name', 'season_num', 'team_name', 'home_runs', 'ops_plus', 'smb_war'])
      return { args, selected, BATTING_COLS }
    },
    template: '<ExportColumnSelector :columns="BATTING_COLS" v-model="selected" />',
  }),
}

export const NoneSelected: Story = {
  render: (args) => ({
    components: { ExportColumnSelector },
    setup() {
      const selected = ref<string[]>([])
      return { args, selected, BATTING_COLS }
    },
    template: '<ExportColumnSelector :columns="BATTING_COLS" v-model="selected" />',
  }),
}

export const AllVariants: Story = {
  render: () => ({
    components: { ExportColumnSelector },
    setup() {
      const allSelected = ref(BATTING_COLS.map((c) => c.key))
      const partial = ref(['player_name', 'home_runs', 'ops_plus'])
      const none = ref<string[]>([])
      return { BATTING_COLS, allSelected, partial, none }
    },
    template: `
      <div style="display:flex;gap:2rem;padding:1.5rem;background:#0d1117;align-items:flex-start">
        <div style="width:200px">
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">All selected</p>
          <ExportColumnSelector :columns="BATTING_COLS" v-model="allSelected" />
        </div>
        <div style="width:200px">
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">Partial</p>
          <ExportColumnSelector :columns="BATTING_COLS" v-model="partial" />
        </div>
        <div style="width:200px">
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">None selected</p>
          <ExportColumnSelector :columns="BATTING_COLS" v-model="none" />
        </div>
      </div>
    `,
  }),
}
