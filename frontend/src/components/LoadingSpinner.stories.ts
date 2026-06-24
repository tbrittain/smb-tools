import type { Meta, StoryObj } from '@storybook/vue3'
import LoadingSpinner from './LoadingSpinner.vue'

const meta: Meta<typeof LoadingSpinner> = {
  title: 'Common/LoadingSpinner',
  component: LoadingSpinner,
}
export default meta

type Story = StoryObj<typeof LoadingSpinner>

export const Default: Story = {}
