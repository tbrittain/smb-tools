import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import PlayerAttributesTable from './PlayerAttributesTable.vue'

const meta: Meta<typeof PlayerAttributesTable> = {
  title: 'Components/PlayerAttributesTable',
  component: PlayerAttributesTable,
  decorators: [() => ({ template: '<div style="padding: 1.5rem"><story /></div>' })],
}
export default meta

type Story = StoryObj<typeof PlayerAttributesTable>

const makeTeam = (id: number, name: string, sortOrder: number) =>
  new main.TeamRefDTO({ teamId: id, teamHistoryId: id * 10, teamName: name, sortOrder })

const makeBatterRow = (i: number) =>
  new main.PlayerSeasonLogDTO({
    seasonNum: i,
    seasonId: i,
    teams: [makeTeam(1, 'Red Sox', 0)],
    age: 22 + i,
    salary: 3000000 + i * 500000,
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
    speed: 65 + i,
    fielding: 80,
    arm: 72,
    velocity: 0,
    junk: 0,
    accuracy: 0,
  })

const makePitcherRow = (i: number) =>
  new main.PlayerSeasonLogDTO({
    seasonNum: i,
    seasonId: i,
    teams: [makeTeam(2, 'Cubs', 0)],
    age: 24 + i,
    salary: 5000000 + i * 1000000,
    primaryPosition: 'P',
    secondaryPosition: '',
    pitcherRole: 'SP',
    batHand: 'R',
    throwHand: 'R',
    chemistryType: 'Loyal',
    traitsJson: '[]',
    pitchesJson: '[]',
    power: 40,
    contact: 35,
    speed: 50,
    fielding: 45,
    arm: 60,
    velocity: 85 + i,
    junk: 78 + i,
    accuracy: 80 + i,
  })

export const BatterSeasons: Story = {
  args: {
    rows: [1, 2, 3, 4, 5].map(makeBatterRow),
    isPitcher: false,
  },
}

export const PitcherSeasons: Story = {
  args: {
    rows: [1, 2, 3, 4, 5].map(makePitcherRow),
    isPitcher: true,
  },
}

export const WithFARow: Story = {
  args: {
    rows: [
      makeBatterRow(1),
      // Season 2: played for Wolves then ended season as FA (no sortOrder=0 entry)
      new main.PlayerSeasonLogDTO({
        ...makeBatterRow(2),
        teams: [makeTeam(3, 'Wolves', 1)],
      }),
      makeBatterRow(3),
    ],
    isPitcher: false,
  },
}

export const Empty: Story = {
  args: { rows: [], isPitcher: false },
}
