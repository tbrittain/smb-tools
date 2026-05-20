import type { Meta, StoryObj } from '@storybook/vue3'
import AttributesTable from './AttributesTable.vue'

const meta: Meta<typeof AttributesTable> = {
  title: 'Components/AttributesTable',
  component: AttributesTable,
}
export default meta

type Story = StoryObj<typeof AttributesTable>

const base = { power: 85, contact: 72, speed: 60, fielding: 78, arm: 88, velocity: 92, junk: 75, accuracy: 80 }

export const PositionPlayer: Story = {
  args: { ...base, showPitching: false },
}

export const Pitcher: Story = {
  args: { ...base, showPitching: true },
}

export const TwoWay: Story = {
  args: {
    power: 75,
    contact: 70,
    speed: 65,
    fielding: 72,
    arm: 75,
    velocity: 88,
    junk: 82,
    accuracy: 85,
    showPitching: true,
  },
}

export const LowStats: Story = {
  args: {
    power: 30,
    contact: 25,
    speed: 40,
    fielding: 35,
    arm: 28,
    velocity: 45,
    junk: 32,
    accuracy: 38,
    showPitching: false,
  },
}
