import type { Meta, StoryObj } from '@storybook/vue3'
import { EXPORT_DATASETS } from '../lib/exportDatasets'
import ExportDatasetPicker from './ExportDatasetPicker.vue'

const meta: Meta<typeof ExportDatasetPicker> = {
  title: 'Components/ExportDatasetPicker',
  component: ExportDatasetPicker,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ExportDatasetPicker>

export const Default: Story = {
  args: {
    datasets: EXPORT_DATASETS,
    modelValue: 'batting_season',
  },
}

export const PitchingSelected: Story = {
  args: {
    datasets: EXPORT_DATASETS,
    modelValue: 'pitching_season',
  },
}

export const CareerSelected: Story = {
  args: {
    datasets: EXPORT_DATASETS,
    modelValue: 'career_batting',
  },
}

export const AllVariants: Story = {
  render: () => ({
    components: { ExportDatasetPicker },
    setup() {
      return { EXPORT_DATASETS }
    },
    template: `
      <div style="display:flex;flex-direction:column;gap:2rem;padding:1.5rem;background:#0d1117">
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">Season dataset selected</p>
          <ExportDatasetPicker :datasets="EXPORT_DATASETS" model-value="batting_season" />
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin-bottom:0.5rem">Career dataset selected</p>
          <ExportDatasetPicker :datasets="EXPORT_DATASETS" model-value="career_batting" />
        </div>
      </div>
    `,
  }),
}
