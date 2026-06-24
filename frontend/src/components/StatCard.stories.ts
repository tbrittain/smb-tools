import type { Meta, StoryObj } from '@storybook/vue3'
import StatCard from './StatCard.vue'

const meta: Meta<typeof StatCard> = {
  title: 'Common/StatCard',
  component: StatCard,
}
export default meta

type Story = StoryObj<typeof StatCard>

export const Default: Story = {
  args: { label: 'Home Runs', value: '47' },
}

export const WithSubtext: Story = {
  args: { label: 'Home Runs', value: '47', subtext: 'John Smith · Season 3' },
}

export const WithLink: Story = {
  args: { label: 'Champion', value: 'Red Sox', subtext: 'Season 5', to: '/teams/1' },
}

export const EmptyValue: Story = {
  args: { label: 'Champion', value: '—' },
}
