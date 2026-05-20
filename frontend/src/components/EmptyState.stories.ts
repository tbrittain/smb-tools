import type { Meta, StoryObj } from '@storybook/vue3'
import EmptyState from './EmptyState.vue'

const meta: Meta<typeof EmptyState> = {
  title: 'Components/EmptyState',
  component: EmptyState,
}
export default meta

type Story = StoryObj<typeof EmptyState>

export const Default: Story = {
  args: { message: 'No data yet' },
}

export const WithSubtext: Story = {
  args: {
    message: 'No seasons synced',
    subtext: 'Use the Sync Season button to import your first season.',
  },
}
