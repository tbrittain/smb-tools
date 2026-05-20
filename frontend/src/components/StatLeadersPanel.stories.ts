import type { Meta, StoryObj } from '@storybook/vue3'
import StatLeadersPanel from './StatLeadersPanel.vue'

const meta: Meta<typeof StatLeadersPanel> = {
  title: 'Components/StatLeadersPanel',
  component: StatLeadersPanel,
}
export default meta

type Story = StoryObj<typeof StatLeadersPanel>

const leader = (playerId: number, first: string, last: string, val: number) => ({
  playerId,
  firstName: first,
  lastName: last,
  teamName: 'Cubs',
  statValue: val,
})

export const Loading: Story = {
  args: { leaders: null, loading: true },
}

export const Populated: Story = {
  args: {
    loading: false,
    leaders: {
      seasonNum: 5,
      ba: leader(1, 'Ted', 'Williams', 0.406),
      hr: leader(2, 'Barry', 'Bonds', 73),
      rbi: leader(3, 'Hack', 'Wilson', 191),
      era: leader(4, 'Pedro', 'Martinez', 1.74),
      wins: leader(5, 'Cy', 'Young', 24),
      strikeouts: leader(6, 'Randy', 'Johnson', 372),
    },
  },
}

export const SomeNull: Story = {
  args: {
    loading: false,
    leaders: {
      seasonNum: 1,
      ba: leader(1, 'Ted', 'Williams', 0.406),
      hr: null,
      rbi: null,
      era: leader(4, 'Pedro', 'Martinez', 1.74),
      wins: null,
      strikeouts: null,
    },
  },
}

export const NoData: Story = {
  args: { leaders: null, loading: false },
}
