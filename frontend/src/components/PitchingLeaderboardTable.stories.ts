import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import PitchingLeaderboardTable from './PitchingLeaderboardTable.vue'

const P0 = 0
const P1 = 1
const P2 = 2

const meta: Meta<typeof PitchingLeaderboardTable> = {
  title: 'Tables/PitchingLeaderboardTable',
  component: PitchingLeaderboardTable,
  decorators: [() => ({ template: '<div style="padding: 1.5rem"><story /></div>' })],
}
export default meta

type Story = StoryObj<typeof PitchingLeaderboardTable>

function makeCareerRow(i: number): main.PitchingLeaderRowDTO {
  const outs = 600 + i * 90
  const er = Math.round(outs * (0.12 + i * 0.005))
  const k = Math.round(outs * 0.28)
  const bb = Math.round(outs * 0.09)
  const h = Math.round(outs * 0.7)
  return new main.PitchingLeaderRowDTO({
    playerId: i,
    firstName: ['Greg', 'Pedro', 'Randy', 'Roger', 'Nolan'][i % 5],
    lastName: ['Maddux', 'Martinez', 'Johnson', 'Clemens', 'Ryan'][i % 5],
    isHallOfFamer: i % 2 === 0,
    seasonsPlayed: 4 + i,
    seasonNum: 0,
    teamName: '',
    age: 0,
    pitcherRole: '',
    throwHand: '',
    chemistryType: '',
    wins: 10 + i * 2,
    losses: 6 + i,
    games: 30 + i * 2,
    gamesStarted: 28 + i,
    completeGames: Math.round(i * 0.5),
    shutouts: i % 3,
    saves: 0,
    outsPitched: outs,
    hitsAllowed: h,
    earnedRuns: er,
    homeRunsAllowed: Math.round(er * 0.15),
    walks: bb,
    strikeouts: k,
    hitBatters: 4,
    battersFaced: Math.round(outs * 1.15),
    gamesFinished: 2,
    runsAllowed: Math.round(er * 1.1),
    wildPitches: 3,
    totalPitches: Math.round(outs * 16),
    era: (er * 27) / outs,
    whip: ((h + bb) * 3) / outs,
    k9: (k * 27) / outs,
    bb9: (bb * 27) / outs,
    winPct: (10 + i * 2) / (16 + i * 3),
  })
}

function makeSeasonRow(i: number): main.PitchingLeaderRowDTO {
  const outs = 540 + i * 30
  const er = Math.round(outs * 0.13)
  const k = Math.round(outs * 0.27)
  const bb = Math.round(outs * 0.09)
  const h = Math.round(outs * 0.72)
  return new main.PitchingLeaderRowDTO({
    playerId: i,
    firstName: ['Greg', 'Pedro', 'Randy'][i % 3],
    lastName: ['Maddux', 'Martinez', 'Johnson'][i % 3],
    isHallOfFamer: i === 0,
    seasonsPlayed: 0,
    seasonNum: (i % 4) + 1,
    teamName: i % 2 === 0 ? 'Braves' : 'Red Sox',
    age: 29 + (i % 4),
    pitcherRole: i % 3 === 2 ? 'RP' : 'SP',
    throwHand: i % 4 === 0 ? 'L' : 'R',
    chemistryType: 'Scholarly',
    wins: 14 + i,
    losses: 7 + i,
    games: 32,
    gamesStarted: 30,
    completeGames: 3,
    shutouts: 1,
    saves: 0,
    outsPitched: outs,
    hitsAllowed: h,
    earnedRuns: er,
    homeRunsAllowed: Math.round(er * 0.18),
    walks: bb,
    strikeouts: k,
    hitBatters: 5,
    battersFaced: Math.round(outs * 1.13),
    gamesFinished: 2,
    runsAllowed: Math.round(er * 1.08),
    wildPitches: 2,
    totalPitches: Math.round(outs * 16),
    era: (er * 27) / outs,
    whip: ((h + bb) * 3) / outs,
    k9: (k * 27) / outs,
    bb9: (bb * 27) / outs,
    winPct: (14 + i) / (21 + i * 2),
  })
}

export const CareerView: Story = {
  args: {
    rows: [0, 1, 2, 3, 4].map(makeCareerRow),
    isCareer: true,
  },
}

export const SeasonView: Story = {
  args: {
    rows: [0, 1, 2].map(makeSeasonRow),
    isCareer: false,
  },
}

export const Empty: Story = {
  args: { rows: [], isCareer: true },
}

export const LargeSet: Story = {
  args: {
    rows: Array.from({ length: 80 }, (_, i) => makeCareerRow(i)),
    isCareer: true,
  },
}

// Mock highlights: P0 is both season leader AND single-season record holder in K (season 1);
// P1 is season leader only in wins (season 2); P2 holds the career RS strikeout record.
const mockHighlights: main.StatHighlightsDTO = new main.StatHighlightsDTO({
  leagueLeadersBatting: {},
  leagueLeadersPitching: {
    '1': { strikeouts: [P0] },
    '2': { wins: [P1], strikeouts: [P1] },
  },
  singleSeasonBatting: {},
  singleSeasonPitching: {
    strikeouts: [{ playerId: P0, seasonNum: 1 }],
  },
  careerBattingRS: {},
  careerBattingPO: {},
  careerPitchingRS: { strikeouts: [P2] },
  careerPitchingPO: {},
})

export const SeasonViewWithHighlights: Story = {
  args: {
    rows: [0, 1, 2].map(makeSeasonRow),
    isCareer: false,
    highlights: mockHighlights,
    totalRecords: 3,
    first: 0,
    sortField: 'smbWar',
    sortOrder: -1,
  },
}

export const CareerViewWithHighlights: Story = {
  args: {
    rows: [0, 1, 2, 3, 4].map(makeCareerRow),
    isCareer: true,
    highlights: mockHighlights,
  },
}
