import type { Meta, StoryObj } from '@storybook/vue3'
import AppButton from './AppButton.vue'

const meta: Meta<typeof AppButton> = {
  title: 'Common/AppButton',
  component: AppButton,
  tags: ['autodocs'],
  argTypes: {
    variant: { control: 'select', options: ['primary', 'secondary', 'ghost'] },
    size: { control: 'select', options: ['sm', 'md'] },
    disabled: { control: 'boolean' },
    loading: { control: 'boolean' },
  },
}

export default meta
type Story = StoryObj<typeof AppButton>

export const Primary: Story = {
  args: { variant: 'primary', size: 'md', disabled: false },
  render: (args) => ({
    components: { AppButton },
    setup: () => ({ args }),
    template: '<AppButton v-bind="args">Save</AppButton>',
  }),
}

export const Secondary: Story = {
  args: { variant: 'secondary', size: 'md', disabled: false },
  render: (args) => ({
    components: { AppButton },
    setup: () => ({ args }),
    template: '<AppButton v-bind="args">Cancel</AppButton>',
  }),
}

export const Ghost: Story = {
  args: { variant: 'ghost', size: 'md', disabled: false },
  render: (args) => ({
    components: { AppButton },
    setup: () => ({ args }),
    template: '<AppButton v-bind="args">Switch franchise</AppButton>',
  }),
}

export const Small: Story = {
  args: { variant: 'primary', size: 'sm', disabled: false },
  render: (args) => ({
    components: { AppButton },
    setup: () => ({ args }),
    template: '<AppButton v-bind="args">+ New</AppButton>',
  }),
}

export const Disabled: Story = {
  args: { variant: 'primary', size: 'md', disabled: true },
  render: (args) => ({
    components: { AppButton },
    setup: () => ({ args }),
    template: '<AppButton v-bind="args">Saving…</AppButton>',
  }),
}

export const Loading: Story = {
  args: { variant: 'primary', size: 'md', loading: true },
  render: (args) => ({
    components: { AppButton },
    setup: () => ({ args }),
    template: '<AppButton v-bind="args">Exporting…</AppButton>',
  }),
}

export const AllVariants: Story = {
  render: () => ({
    components: { AppButton },
    template: `
      <div style="display:flex;gap:1rem;align-items:center;flex-wrap:wrap;padding:1rem;background:#0d1117">
        <AppButton variant="primary">Primary</AppButton>
        <AppButton variant="secondary">Secondary</AppButton>
        <AppButton variant="ghost">Ghost</AppButton>
        <AppButton variant="primary" size="sm">Small primary</AppButton>
        <AppButton variant="primary" disabled>Disabled</AppButton>
        <AppButton variant="primary" loading>Loading</AppButton>
      </div>
    `,
  }),
}
