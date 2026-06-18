import type { Meta, StoryObj } from '@storybook/vue3'
import ModeChooser from './ModeChooser.vue'

const meta: Meta<typeof ModeChooser> = {
  title: 'Components/ModeChooser',
  component: ModeChooser,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof ModeChooser>

export const Default: Story = {
  render: () => ({
    components: { ModeChooser },
    template: '<div style="background:#0d1117;padding:2rem"><ModeChooser @select="() => {}" /></div>',
  }),
}
