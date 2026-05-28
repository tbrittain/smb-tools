import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import BattingLeaderboardTable from './BattingLeaderboardTable.vue'

// Player IDs used in stories
const P0 = 0
const P1 = 1
const P2 = 2

const meta: Meta<typeof BattingLeaderboardTable> = {
  title: 'Components/BattingLeaderboardTable',
  component: BattingLeaderboardTable,
  decorators: [() => ({ template: '<div style="padding: 1.5rem"><story /></div>' })],
}
export default meta

type Story = StoryObj<typeof BattingLeaderboardTable>

function makeCareerRow(i: number): main.BattingLeaderRowDTO {
  const ab = 500 + i * 50
  const h = Math.round(ab * (0.27 + i * 0.005))
  const hr = 15 + i * 3
  const bb = Math.round(ab * 0.1)
  const sf = 5
  return new main.BattingLeaderRowDTO({
    playerId: i,
    firstName: ['Aaron', 'Babe', 'Cal', 'Derek', 'Eddie'][i % 5],
    lastName: ['Judge', 'Ruth', 'Ripken', 'Jeter', 'Murray'][i % 5],
    isHallOfFamer: i % 3 === 0,
    seasonsPlayed: 5 + i,
    seasonNum: 0,
    teamName: '',
    age: 0,
    primaryPosition: '',
    batHand: '',
    chemistryType: '',
    gamesPlayed: Math.round(ab / 4),
    gamesBatting: Math.round(ab / 4),
    atBats: ab,
    runs: Math.round(h * 0.65),
    hits: h,
    doubles: Math.round(h * 0.2),
    triples: Math.round(h * 0.02),
    homeRuns: hr,
    rbi: Math.round(hr * 3.2),
    stolenBases: Math.round(ab * 0.04),
    caughtStealing: 4,
    walks: bb,
    strikeouts: Math.round(ab * 0.2),
    hitByPitch: 5,
    sacHits: 2,
    sacFlies: sf,
    errors: 3,
    passedBalls: 0,
    ba: h / ab,
    obp: (h + bb + 5) / (ab + bb + 5 + sf),
    slg: (h + hr * 3) / ab,
    ops: (h + bb + 5) / (ab + bb + 5 + sf) + (h + hr * 3) / ab,
    iso: (hr * 3) / ab,
  })
}

function makeSeasonRow(i: number): main.BattingLeaderRowDTO {
  const ab = 480 + i * 20
  const h = Math.round(ab * 0.275)
  const hr = 12 + i * 2
  const bb = Math.round(ab * 0.09)
  const sf = 4
  return new main.BattingLeaderRowDTO({
    playerId: i,
    firstName: ['Aaron', 'Babe', 'Cal'][i % 3],
    lastName: ['Judge', 'Ruth', 'Ripken'][i % 3],
    isHallOfFamer: i === 1,
    seasonsPlayed: 0,
    seasonNum: (i % 5) + 1,
    teamName: i % 2 === 0 ? 'Yankees' : 'Red Sox',
    age: 28 + (i % 5),
    primaryPosition: ['SS', '1B', 'CF', 'RF', '3B'][i % 5],
    batHand: i % 2 === 0 ? 'R' : 'L',
    chemistryType: 'Competitive',
    gamesPlayed: Math.round(ab / 4),
    gamesBatting: Math.round(ab / 4),
    atBats: ab,
    runs: Math.round(h * 0.6),
    hits: h,
    doubles: Math.round(h * 0.2),
    triples: 3,
    homeRuns: hr,
    rbi: Math.round(hr * 3),
    stolenBases: 8,
    caughtStealing: 3,
    walks: bb,
    strikeouts: Math.round(ab * 0.18),
    hitByPitch: 4,
    sacHits: 1,
    sacFlies: sf,
    errors: 2,
    passedBalls: 0,
    ba: h / ab,
    obp: (h + bb) / (ab + bb + sf),
    slg: (h + hr * 3) / ab,
    ops: (h + bb) / (ab + bb + sf) + (h + hr * 3) / ab,
    iso: (hr * 3) / ab,
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

// Mock highlights: P0 is both season leader AND single-season record holder in HR (season 1);
// P1 is season leader only in hits (season 2); P2 holds the career RS HR record.
const mockHighlights: main.StatHighlightsDTO = new main.StatHighlightsDTO({
  leagueLeadersBatting: {
    '1': { homeRuns: [P0] },
    '2': { hits: [P1], homeRuns: [P1] },
  },
  leagueLeadersPitching: {},
  singleSeasonBatting: {
    homeRuns: [{ playerId: P0, seasonNum: 1 }],
  },
  singleSeasonPitching: {},
  careerBattingRS: { homeRuns: [P2] },
  careerBattingPO: {},
  careerPitchingRS: {},
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
