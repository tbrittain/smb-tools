import type { Meta, StoryObj } from '@storybook/vue3'
import AppHelpButton from './AppHelpButton.vue'

const meta: Meta<typeof AppHelpButton> = {
  title: 'Common/AppHelpButton',
  component: AppHelpButton,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof AppHelpButton>

export const Default: Story = {
  args: { docsPath: 'save-game-setup.html#syncing-a-season' },
}

export const InContext: Story = {
  render: (args) => ({
    components: { AppHelpButton },
    setup: () => ({ args }),
    template: `
      <div style="display:flex;align-items:center;gap:0.5rem;padding:1.5rem;background:#0d1117">
        <h3 style="margin:0;color:#e6edf3;font-size:1.0625rem">Sync Season</h3>
        <AppHelpButton v-bind="args" />
      </div>
    `,
  }),
  args: { docsPath: 'save-game-setup.html#syncing-a-season' },
}
