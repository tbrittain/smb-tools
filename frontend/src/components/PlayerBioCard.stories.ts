import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import PlayerBioCard from './PlayerBioCard.vue'

const meta: Meta<typeof PlayerBioCard> = {
  title: 'Player/PlayerBioCard',
  component: PlayerBioCard,
}
export default meta

type Story = StoryObj<typeof PlayerBioCard>

const baseSeason = new main.PlayerSeasonLogDTO({
  seasonNum: 8,
  seasonId: 8,
  teams: [new main.TeamRefDTO({ teamId: 1, teamHistoryId: 1, teamName: 'Giants', sortOrder: 0 })],
  age: 34,
  salary: 15000000,
  primaryPosition: 'LF',
  secondaryPosition: 'OF',
  pitcherRole: '',
  batHand: 'L',
  throwHand: 'L',
  chemistryType: 'Competitive',
  traits: [],
  pitches: [],
  power: 99,
  contact: 85,
  speed: 72,
  fielding: 78,
  arm: 80,
  velocity: 0,
  junk: 0,
  accuracy: 0,
})

const basePlayer = new main.PlayerCareerDTO({
  playerId: 1,
  firstName: 'Barry',
  lastName: 'Bonds',
  isHallOfFamer: false,
})

export const PositionPlayer: Story = {
  args: { player: basePlayer, currentSeason: baseSeason },
}

export const HallOfFamer: Story = {
  args: {
    player: new main.PlayerCareerDTO({ ...basePlayer, isHallOfFamer: true }),
    currentSeason: baseSeason,
  },
}

export const Pitcher: Story = {
  args: {
    player: new main.PlayerCareerDTO({ playerId: 2, firstName: 'Randy', lastName: 'Johnson', isHallOfFamer: false }),
    currentSeason: new main.PlayerSeasonLogDTO({
      ...baseSeason,
      primaryPosition: 'P',
      secondaryPosition: '',
      pitcherRole: 'SP',
      batHand: 'R',
      throwHand: 'L',
      teams: [new main.TeamRefDTO({ teamId: 2, teamHistoryId: 2, teamName: 'Diamondbacks', sortOrder: 0 })],
    }),
  },
}

export const NoBioDetail: Story = {
  args: { player: basePlayer },
}

export const WithCareerEarnings: Story = {
  args: {
    player: new main.PlayerCareerDTO({ ...basePlayer, isHallOfFamer: true }),
    currentSeason: baseSeason,
    careerEarnings: 127500000,
  },
}

export const ZeroCareerEarnings: Story = {
  args: {
    player: basePlayer,
    currentSeason: baseSeason,
    careerEarnings: 0,
  },
}
