import type { Meta, StoryObj } from '@storybook/vue3'
import StatHighlightLegend from './StatHighlightLegend.vue'

const meta: Meta<typeof StatHighlightLegend> = {
  title: 'Components/StatHighlightLegend',
  component: StatHighlightLegend,
  decorators: [() => ({ template: '<div style="padding: 1.5rem"><story /></div>' })],
}
export default meta

type Story = StoryObj<typeof StatHighlightLegend>

export const Default: Story = {}

export const LeaderOnly: Story = {
  args: { showLeader: true, showRecord: false },
}

export const RecordOnly: Story = {
  args: { showLeader: false, showRecord: true },
}
