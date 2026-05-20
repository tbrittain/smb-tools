import type { Meta, StoryObj } from '@storybook/vue3'
import PlayerBioCard from './PlayerBioCard.vue'

const meta: Meta<typeof PlayerBioCard> = {
  title: 'Components/PlayerBioCard',
  component: PlayerBioCard,
}
export default meta

type Story = StoryObj<typeof PlayerBioCard>

const basePlayer = {
  playerId: 1,
  firstName: 'Barry',
  lastName: 'Bonds',
  isHallOfFamer: false,
  batting: null,
  pitching: null,
}

const baseSeason = {
  seasonNum: 8,
  seasonId: 8,
  teamName: 'Giants',
  age: 34,
  salary: 15000000,
  primaryPosition: 'LF',
  secondaryPosition: 'OF',
  pitcherRole: '',
  batHand: 'L',
  throwHand: 'L',
  chemistryType: 'Competitive',
  traitsJson: '[]',
  pitchesJson: '[]',
  power: 99,
  contact: 85,
  speed: 72,
  fielding: 78,
  arm: 80,
  velocity: 0,
  junk: 0,
  accuracy: 0,
  batting: null,
  pitching: null,
  playoffBatting: null,
  playoffPitching: null,
}

export const PositionPlayer: Story = {
  args: { player: basePlayer, currentSeason: baseSeason },
}

export const HallOfFamer: Story = {
  args: {
    player: { ...basePlayer, isHallOfFamer: true },
    currentSeason: baseSeason,
  },
}

export const Pitcher: Story = {
  args: {
    player: { ...basePlayer, firstName: 'Randy', lastName: 'Johnson' },
    currentSeason: {
      ...baseSeason,
      primaryPosition: 'P',
      secondaryPosition: '',
      pitcherRole: 'SP',
      batHand: 'R',
      throwHand: 'L',
      teamName: 'Diamondbacks',
    },
  },
}

export const NoBioDetail: Story = {
  args: { player: basePlayer },
}
