import type { Meta, StoryObj } from '@storybook/vue3'
import IconButton from './IconButton.vue'

const meta: Meta<typeof IconButton> = {
  title: 'Common/IconButton',
  component: IconButton,
  tags: ['autodocs'],
  argTypes: {
    variant: { control: 'select', options: ['secondary', 'danger'] },
    size: { control: 'select', options: ['sm', 'md'] },
    rounded: { control: 'boolean' },
    disabled: { control: 'boolean' },
  },
}

export default meta
type Story = StoryObj<typeof IconButton>

export const Secondary: Story = {
  args: { icon: 'pi pi-pencil', variant: 'secondary', size: 'sm' },
  render: (args) => ({
    components: { IconButton },
    setup: () => ({ args }),
    template: '<IconButton v-bind="args" aria-label="Edit" />',
  }),
}

export const Danger: Story = {
  args: { icon: 'pi pi-trash', variant: 'danger', size: 'sm' },
  render: (args) => ({
    components: { IconButton },
    setup: () => ({ args }),
    template: '<IconButton v-bind="args" aria-label="Delete" />',
  }),
}

export const Rounded: Story = {
  args: { icon: 'pi pi-question-circle', variant: 'secondary', rounded: true },
  render: (args) => ({
    components: { IconButton },
    setup: () => ({ args }),
    template: '<IconButton v-bind="args" aria-label="Help" />',
  }),
}

export const Medium: Story = {
  args: { icon: 'pi pi-trash', variant: 'danger', size: 'md' },
  render: (args) => ({
    components: { IconButton },
    setup: () => ({ args }),
    template: '<IconButton v-bind="args" aria-label="Delete" />',
  }),
}

export const Disabled: Story = {
  args: { icon: 'pi pi-trash', variant: 'danger', size: 'sm', disabled: true },
  render: (args) => ({
    components: { IconButton },
    setup: () => ({ args }),
    template: '<IconButton v-bind="args" aria-label="Delete" />',
  }),
}

export const AllVariants: Story = {
  render: () => ({
    components: { IconButton },
    template: `
      <div style="display:flex;gap:1rem;align-items:center;flex-wrap:wrap;padding:1rem;background:#0d1117">
        <IconButton icon="pi pi-pencil" variant="secondary" aria-label="Edit" />
        <IconButton icon="pi pi-trash" variant="danger" aria-label="Delete" />
        <IconButton icon="pi pi-question-circle" variant="secondary" rounded aria-label="Help" />
        <IconButton icon="pi pi-trash" variant="danger" size="md" aria-label="Delete" />
        <IconButton icon="pi pi-trash" variant="danger" disabled aria-label="Delete" />
      </div>
    `,
  }),
}
