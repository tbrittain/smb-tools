import type { Meta, StoryObj } from '@storybook/vue3'
import type { main } from '../../wailsjs/go/models'
import TeamSeasonCard from './TeamSeasonCard.vue'

const meta: Meta<typeof TeamSeasonCard> = {
  title: 'Components/TeamSeasonCard',
  component: TeamSeasonCard,
}
export default meta

type Story = StoryObj<typeof TeamSeasonCard>

const base: main.TeamSeasonSummaryDTO = {
  historyId: 1,
  seasonId: 1,
  seasonNum: 5,
  teamName: 'Red Sox',
  divisionName: 'East',
  conferenceName: 'American',
  wins: 95,
  losses: 47,
  winPct: 0.669,
  gamesBack: 0,
  runsFor: 810,
  runsAgainst: 650,
  budget: 50000000,
  payroll: 48000000,
  playoffSeed: undefined,
  playoffWins: undefined,
  playoffLosses: undefined,
  playoffRunsFor: undefined,
  playoffRunsAgainst: undefined,
  totalPower: 1200,
  totalContact: 1100,
  totalSpeed: 900,
  totalFielding: 1050,
  totalArm: 950,
  totalVelocity: 800,
  totalJunk: 750,
  totalAccuracy: 820,
  isChampion: false,
}

export const NonPlayoff: Story = {
  args: { season: { ...base, wins: 72, losses: 70, winPct: 0.507 } },
}

export const PlayoffTeam: Story = {
  args: {
    season: {
      ...base,
      playoffSeed: 3,
      playoffWins: 4,
      playoffLosses: 2,
    },
  },
}

export const Champion: Story = {
  args: {
    season: {
      ...base,
      isChampion: true,
      playoffSeed: 1,
      playoffWins: 8,
      playoffLosses: 1,
    },
  },
}
