import type { Meta, StoryObj } from '@storybook/vue3'
import SearchInput from './SearchInput.vue'

const meta: Meta<typeof SearchInput> = {
  title: 'Components/SearchInput',
  component: SearchInput,
}
export default meta

type Story = StoryObj<typeof SearchInput>

export const Empty: Story = {
  args: { modelValue: '', loading: false, placeholder: 'Search players and teams…' },
}

export const WithValue: Story = {
  args: { modelValue: 'John Smith', loading: false },
}

export const Loading: Story = {
  args: { modelValue: 'Smith', loading: true },
}
