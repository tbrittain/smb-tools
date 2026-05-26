import type { Meta, StoryObj } from '@storybook/vue3'
import AppLink from './AppLink.vue'

const meta: Meta<typeof AppLink> = {
  title: 'Components/AppLink',
  component: AppLink,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof AppLink>

export const Internal: Story = {
  args: { to: '/players/1' },
  render: (args) => ({
    components: { AppLink },
    setup: () => ({ args }),
    template: '<AppLink v-bind="args">Manny Ramirez</AppLink>',
  }),
}

export const External: Story = {
  args: { href: 'https://example.com' },
  render: (args) => ({
    components: { AppLink },
    setup: () => ({ args }),
    template: '<AppLink v-bind="args">Baseball Reference</AppLink>',
  }),
}

export const AllVariants: Story = {
  render: () => ({
    components: { AppLink },
    template: `
      <div style="display:flex;flex-direction:column;gap:1rem;padding:1.5rem;background:#0d1117;font-size:0.9375rem">
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;width:80px;font-size:0.75rem">internal</span>
          <AppLink to="/players/1">Manny Ramirez</AppLink>
        </div>
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;width:80px;font-size:0.75rem">external</span>
          <AppLink href="https://example.com">Baseball Reference ↗</AppLink>
        </div>
        <div style="display:flex;gap:2rem;align-items:center">
          <span style="color:#8b949e;width:80px;font-size:0.75rem">no route</span>
          <AppLink>plain span fallback</AppLink>
        </div>
      </div>
    `,
  }),
}
