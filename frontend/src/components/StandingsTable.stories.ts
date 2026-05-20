import type { Meta, StoryObj } from '@storybook/vue3'
import StandingsTable from './StandingsTable.vue'

const meta: Meta<typeof StandingsTable> = {
  title: 'Components/StandingsTable',
  component: StandingsTable,
}
export default meta

type Story = StoryObj<typeof StandingsTable>

const row = (
  id: number,
  name: string,
  conf: string,
  div: string,
  w: number,
  l: number,
  gb: number,
  rf: number,
  ra: number,
  seed?: number,
) => ({
  historyId: id,
  teamId: id,
  teamName: name,
  conferenceName: conf,
  divisionName: div,
  wins: w,
  losses: l,
  winPct: w / (w + l),
  gamesBack: gb,
  runsFor: rf,
  runsAgainst: ra,
  runDiff: rf - ra,
  playoffSeed: seed,
})

const standings = [
  row(1, 'Red Sox', 'American', 'East', 95, 47, 0, 810, 650, 1),
  row(2, 'Yankees', 'American', 'East', 88, 54, 7, 780, 690, 3),
  row(3, 'Rays', 'American', 'East', 80, 62, 15, 720, 710),
  row(4, 'Astros', 'American', 'West', 91, 51, 0, 800, 660, 2),
  row(5, 'Dodgers', 'National', 'West', 98, 44, 0, 850, 600, 1),
  row(6, 'Giants', 'National', 'West', 90, 52, 8, 760, 680, 4),
]

export const Default: Story = {
  args: { standings },
}

export const WithActiveRow: Story = {
  args: { standings, activeHistoryId: 1 },
}

export const Empty: Story = {
  args: { standings: [] },
}
