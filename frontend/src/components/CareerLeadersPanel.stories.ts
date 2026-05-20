import type { Meta, StoryObj } from '@storybook/vue3'
import CareerLeadersPanel from './CareerLeadersPanel.vue'

const meta: Meta<typeof CareerLeadersPanel> = {
  title: 'Components/CareerLeadersPanel',
  component: CareerLeadersPanel,
}
export default meta

type Story = StoryObj<typeof CareerLeadersPanel>

const makeLeaders = (vals: number[]) =>
  vals.map((v, i) => ({
    playerId: i + 1,
    firstName: ['Ted', 'Barry', 'Hack', 'Babe', 'Hank'][i] ?? 'Player',
    lastName: ['Williams', 'Bonds', 'Wilson', 'Ruth', 'Aaron'][i] ?? String(i),
    statValue: v,
    seasonsPlayed: 8,
  }))

export const Populated: Story = {
  args: {
    leaders: {
      hr: makeLeaders([580, 520, 490, 450, 420]),
      hits: makeLeaders([3000, 2900, 2800, 2700, 2600]),
      rbi: makeLeaders([1800, 1700, 1600, 1500, 1400]),
      wins: makeLeaders([200, 180, 160, 140, 120]),
      strikeouts: makeLeaders([2800, 2600, 2400, 2200, 2000]),
      saves: makeLeaders([400, 350, 300, 250, 200]),
    },
  },
}

export const Empty: Story = {
  args: { leaders: null },
}

export const SomeEmpty: Story = {
  args: {
    leaders: {
      hr: makeLeaders([580]),
      hits: [],
      rbi: makeLeaders([1800, 1700]),
      wins: [],
      strikeouts: [],
      saves: [],
    },
  },
}
