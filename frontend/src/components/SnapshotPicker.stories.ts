import type { Meta, StoryObj } from '@storybook/vue3'
import type { main } from '../../wailsjs/go/models'
import SnapshotPicker from './SnapshotPicker.vue'

const makeSnap = (
  id: number,
  seasonNum: number,
  capturedAt: string,
  fileSizeBytes: number,
  fileExists: boolean,
): main.SnapshotDTO => ({ id, seasonNum, capturedAt, fileSizeBytes, fileExists }) as main.SnapshotDTO

const sampleSnapshots: main.SnapshotDTO[] = [
  makeSnap(1, 1, '2024-04-10T14:22:00Z', 4_300_000, true),
  makeSnap(2, 2, '2024-08-03T09:11:00Z', 4_500_000, true),
  makeSnap(3, 3, '2024-12-19T17:45:00Z', 4_800_000, true),
  makeSnap(4, 3, '2025-01-02T20:30:00Z', 4_820_000, true),
  makeSnap(5, 4, '2025-05-15T11:00:00Z', 5_100_000, false),
]

const meta: Meta<typeof SnapshotPicker> = {
  title: 'Franchise/SnapshotPicker',
  component: SnapshotPicker,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof SnapshotPicker>

export const Default: Story = {
  render: (args) => ({
    components: { SnapshotPicker },
    setup: () => ({ args }),
    template: '<SnapshotPicker v-bind="args" />',
  }),
  args: {
    snapshots: sampleSnapshots,
    loading: false,
    selectedId: 2,
  },
}

export const Empty: Story = {
  render: (args) => ({
    components: { SnapshotPicker },
    setup: () => ({ args }),
    template: '<SnapshotPicker v-bind="args" />',
  }),
  args: {
    snapshots: [],
    loading: false,
    selectedId: null,
  },
}

export const Loading: Story = {
  render: (args) => ({
    components: { SnapshotPicker },
    setup: () => ({ args }),
    template: '<SnapshotPicker v-bind="args" />',
  }),
  args: {
    snapshots: [],
    loading: true,
    selectedId: null,
  },
}

export const AllMissing: Story = {
  render: (args) => ({
    components: { SnapshotPicker },
    setup: () => ({ args }),
    template: '<SnapshotPicker v-bind="args" />',
  }),
  args: {
    snapshots: sampleSnapshots.map((s) => ({ ...s, fileExists: false })),
    loading: false,
    selectedId: null,
  },
}
