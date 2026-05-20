import type { Meta, StoryObj } from '@storybook/vue3'
import CareerStatSummary from './CareerStatSummary.vue'

const meta: Meta<typeof CareerStatSummary> = {
  title: 'Components/CareerStatSummary',
  component: CareerStatSummary,
}
export default meta

type Story = StoryObj<typeof CareerStatSummary>

const batting = {
  gamesPlayed: 2143,
  gamesBatting: 2143,
  atBats: 7844,
  runs: 1545,
  hits: 2935,
  doubles: 601,
  triples: 77,
  homeRuns: 762,
  rbi: 1996,
  stolenBases: 514,
  caughtStealing: 141,
  walks: 2558,
  strikeouts: 1539,
  hitByPitch: 107,
  sacHits: 0,
  sacFlies: 91,
  errors: 72,
  passedBalls: 0,
  ba: 0.298,
  obp: 0.444,
  slg: 0.607,
  ops: 1.051,
  iso: 0.309,
  babip: 0.312,
  kPct: 0.182,
  bbPct: 0.302,
  abPerHr: 10.3,
}

const pitching = {
  wins: 303,
  losses: 166,
  games: 618,
  gamesStarted: 603,
  completeGames: 100,
  shutouts: 28,
  saves: 0,
  outsPitched: 14163,
  hitsAllowed: 3942,
  earnedRuns: 1392,
  homeRunsAllowed: 291,
  walks: 1171,
  strikeouts: 4875,
  hitBatters: 188,
  battersFaced: 19272,
  gamesFinished: 4,
  runsAllowed: 1489,
  wildPitches: 130,
  totalPitches: 93000,
  era: 2.65,
  whip: 1.08,
  k9: 9.3,
  bb9: 2.5,
  h9: 7.5,
  hr9: 0.6,
  kPerBb: 4.2,
  kPct: 0.253,
  winPct: 0.646,
  pPerIp: 17.4,
}

export const BattingOnly: Story = {
  args: { batting, pitching: null },
}

export const PitchingOnly: Story = {
  args: { batting: null, pitching },
}

export const Both: Story = {
  args: { batting, pitching },
}

export const Empty: Story = {
  args: { batting: null, pitching: null },
}
