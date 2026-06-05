import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import BugReportDialog from './BugReportDialog.vue'

const meta: Meta<typeof BugReportDialog> = {
  title: 'Components/BugReportDialog',
  component: BugReportDialog,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof BugReportDialog>

export const Default: Story = {
  render: () => ({
    components: { BugReportDialog },
    setup() {
      const visible = ref(true)
      return { visible }
    },
    template: '<BugReportDialog v-model:visible="visible" />',
  }),
}

export const SystemInfoChecked: Story = {
  render: () => ({
    components: { BugReportDialog },
    setup() {
      const visible = ref(true)
      return { visible }
    },
    template: '<BugReportDialog v-model:visible="visible" />',
  }),
}
