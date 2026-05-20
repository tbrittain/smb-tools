import type { Meta, StoryObj } from '@storybook/vue3'
import type { main } from '../../wailsjs/go/models'
import PlayerStatTable from './PlayerStatTable.vue'

const meta: Meta<typeof PlayerStatTable> = {
  title: 'Components/PlayerStatTable',
  component: PlayerStatTable,
}
export default meta

type Story = StoryObj<typeof PlayerStatTable>

const makeBatting = (ab: number, h: number, hr: number, rbi: number): main.CareerBattingStatsDTO => ({
  gamesPlayed: ab > 0 ? Math.round(ab / 4) : 0,
  gamesBatting: ab > 0 ? Math.round(ab / 4) : 0,
  atBats: ab,
  runs: Math.round(h * 0.6),
  hits: h,
  doubles: Math.round(h * 0.2),
  triples: Math.round(h * 0.03),
  homeRuns: hr,
  rbi,
  stolenBases: Math.round(ab * 0.05),
  caughtStealing: 5,
  walks: Math.round(ab * 0.1),
  strikeouts: Math.round(ab * 0.2),
  hitByPitch: 3,
  sacHits: 2,
  sacFlies: 4,
  errors: 2,
  passedBalls: 0,
  ba: h / ab,
  obp: (h + Math.round(ab * 0.1)) / (ab + Math.round(ab * 0.1) + 4),
  slg: (h + hr * 3) / ab,
  ops: null,
  iso: null,
  babip: null,
  kPct: null,
  bbPct: null,
  abPerHr: ab / hr,
})

const rows: main.PlayerSeasonLogDTO[] = [1, 2, 3].map((i) => ({
  seasonNum: i,
  seasonId: i,
  teamName: i < 3 ? 'Red Sox' : 'Cubs',
  age: 25 + i,
  salary: 3000000 + i * 1000000,
  primaryPosition: 'SS',
  secondaryPosition: '',
  pitcherRole: '',
  batHand: 'R',
  throwHand: 'R',
  chemistryType: 'Competitive',
  traitsJson: '[]',
  pitchesJson: '[]',
  power: 70 + i,
  contact: 75 + i,
  speed: 65,
  fielding: 80,
  arm: 72,
  velocity: 0,
  junk: 0,
  accuracy: 0,
  batting: makeBatting(500 + i * 20, 140 + i * 5, 15 + i, 75 + i * 3),
  pitching: null,
  playoffBatting: makeBatting(60, 18, 2, 10),
  playoffPitching: null,
}))

export const BattingRegular: Story = {
  args: { rows, mode: 'batting', showPlayoffs: false },
}

export const BattingPlayoffs: Story = {
  args: { rows, mode: 'batting', showPlayoffs: true },
}

export const Empty: Story = {
  args: { rows: [], mode: 'batting', showPlayoffs: false },
}
