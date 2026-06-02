import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import PlayerStatTable from './PlayerStatTable.vue'

const meta: Meta<typeof PlayerStatTable> = {
  title: 'Components/PlayerStatTable',
  component: PlayerStatTable,
  decorators: [() => ({ template: '<div style="padding: 1.5rem"><story /></div>' })],
}
export default meta

type Story = StoryObj<typeof PlayerStatTable>

const makeBatting = (ab: number, h: number, hr: number, rbi: number) =>
  new main.CareerBattingStatsDTO({
    gamesPlayed: Math.round(ab / 4),
    gamesBatting: Math.round(ab / 4),
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
    abPerHr: ab / hr,
  })

const makePitching = (w: number, l: number, outs: number, er: number, k: number) =>
  new main.CareerPitchingStatsDTO({
    wins: w,
    losses: l,
    games: w + l + 2,
    gamesStarted: w + l,
    completeGames: Math.round((w + l) * 0.1),
    shutouts: 1,
    saves: 0,
    outsPitched: outs,
    hitsAllowed: Math.round(outs * 0.6),
    earnedRuns: er,
    homeRunsAllowed: Math.round(er * 0.2),
    walks: Math.round(outs * 0.09),
    strikeouts: k,
    hitBatters: 5,
    battersFaced: Math.round(outs * 1.2),
    gamesFinished: 2,
    runsAllowed: Math.round(er * 1.1),
    wildPitches: 4,
    totalPitches: Math.round(outs * 16),
    era: (er * 27) / outs,
    whip: ((Math.round(outs * 0.09) + Math.round(outs * 0.6)) * 3) / outs,
    k9: (k * 27) / outs,
    winPct: w / (w + l),
  })

const makeTeam = (id: number, name: string, sortOrder: number) =>
  new main.TeamRefDTO({ teamId: id, teamHistoryId: id * 10, teamName: name, sortOrder })

const makeRow = (i: number) =>
  new main.PlayerSeasonLogDTO({
    seasonNum: i,
    seasonId: i,
    teams: [makeTeam(i < 3 ? 1 : 2, i < 3 ? 'Red Sox' : 'Cubs', 0)],
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
    playoffBatting: makeBatting(60, 18, 2, 10),
  })

const makePitchRow = (i: number) =>
  new main.PlayerSeasonLogDTO({
    ...makeRow(i),
    primaryPosition: 'P',
    pitcherRole: 'SP',
    batting: undefined,
    playoffBatting: undefined,
    pitching: makePitching(12 + i, 8, 600 + i * 30, 60 + i * 3, 180 + i * 15),
    playoffPitching: makePitching(2, 1, 80, 8, 25),
  })

const rows = [1, 2, 3].map(makeRow)
const pitcherRows = [1, 2, 3].map(makePitchRow)

export const BattingRegular: Story = {
  args: { rows, mode: 'batting', showPlayoffs: false },
}

export const BattingPlayoffs: Story = {
  args: { rows, mode: 'batting', showPlayoffs: true },
}

export const PitchingRegular: Story = {
  args: { rows: pitcherRows, mode: 'pitching', showPlayoffs: false },
}

export const PitchingPlayoffs: Story = {
  args: { rows: pitcherRows, mode: 'pitching', showPlayoffs: true },
}

export const Empty: Story = {
  args: { rows: [], mode: 'batting', showPlayoffs: false },
}

// ── FA display scenarios ───────────────────────────────────────────────────────

const faWholeSeasonRow = new main.PlayerSeasonLogDTO({
  ...makeRow(1),
  teams: [],
})

const faAfterTeamRow = new main.PlayerSeasonLogDTO({
  ...makeRow(2),
  // Player played for Wolves then ended season as FA (no sortOrder=0 entry)
  teams: [makeTeam(5, 'Honey Badgers', 1)],
})

const fullSeasonTeamRow = new main.PlayerSeasonLogDTO({
  ...makeRow(3),
  teams: [makeTeam(6, 'Wolves', 0)],
})

export const BattingFAWholeSeason: Story = {
  args: { rows: [faWholeSeasonRow], mode: 'batting', showPlayoffs: false },
}

export const BattingFAAfterTeam: Story = {
  args: { rows: [faAfterTeamRow], mode: 'batting', showPlayoffs: false },
}

export const BattingFullSeasonTeam: Story = {
  args: { rows: [fullSeasonTeamRow], mode: 'batting', showPlayoffs: false },
}

export const BattingAllFAVariants: Story = {
  args: {
    rows: [faWholeSeasonRow, faAfterTeamRow, fullSeasonTeamRow],
    mode: 'batting',
    showPlayoffs: false,
  },
}
