import type { Meta, StoryObj } from '@storybook/vue3'
import TeamTopPlayersTable from './TeamTopPlayersTable.vue'

const meta: Meta<typeof TeamTopPlayersTable> = {
  title: 'Components/TeamTopPlayersTable',
  component: TeamTopPlayersTable,
}
export default meta

type Story = StoryObj<typeof TeamTopPlayersTable>

const batter = (
  id: number,
  first: string,
  last: string,
  seasons: number[],
  war: number,
  opsPlus: number,
  awards: string[] = [],
  hof = false,
) => ({
  playerId: id,
  firstName: first,
  lastName: last,
  isHallOfFamer: hof,
  numSeasons: seasons.length,
  seasonNums: seasons,
  isPitcher: false,
  position: 'CF',
  smbWarWithTeam: war,
  avgOpsPlus: opsPlus,
  avgEraPlus: undefined,
  awards,
})

const pitcher = (
  id: number,
  first: string,
  last: string,
  seasons: number[],
  war: number,
  eraPlus: number,
  awards: string[] = [],
) => ({
  playerId: id,
  firstName: first,
  lastName: last,
  isHallOfFamer: false,
  numSeasons: seasons.length,
  seasonNums: seasons,
  isPitcher: true,
  position: 'SP',
  smbWarWithTeam: war,
  avgOpsPlus: undefined,
  avgEraPlus: eraPlus,
  awards,
})

export const Empty: Story = {
  args: { players: [] },
}

export const PositionPlayers: Story = {
  args: {
    players: [
      batter(
        1,
        'Mike',
        'Trout',
        [1, 2, 3, 4, 5],
        32.4,
        178,
        ['MVP', 'All-Star', 'All-Star', 'All-Star', 'League Champion'],
        true,
      ),
      batter(2, 'Aaron', 'Judge', [3, 4, 5], 18.1, 155, ['MVP', 'All-Star', 'League Champion']),
      batter(3, 'Mookie', 'Betts', [1, 2, 3], 14.7, 138, ['Gold Glove', 'All-Star']),
      batter(4, 'Freddie', 'Freeman', [4, 5], 9.2, 128, ['All-Star']),
      batter(5, 'Jose', 'Altuve', [1, 2], 5.8, 115, []),
    ],
  },
}

export const PitcherHeavy: Story = {
  args: {
    players: [
      pitcher(10, 'Jacob', 'deGrom', [1, 2, 3, 4, 5], 28.9, 168, ['Cy Young', 'Cy Young', 'All-Star']),
      pitcher(11, 'Max', 'Scherzer', [2, 3, 4], 17.3, 152, ['Cy Young', 'All-Star']),
      pitcher(12, 'Shane', 'Bieber', [1, 2], 10.1, 141, ['All-Star']),
      batter(13, 'Francisco', 'Lindor', [3, 4, 5], 12.6, 121, ['Gold Glove', 'All-Star']),
      pitcher(14, 'Sandy', 'Alcantara', [5], 4.7, 135, []),
    ],
  },
}

export const NonContiguousSeasons: Story = {
  args: {
    players: [
      batter(
        20,
        'Albert',
        'Pujols',
        [1, 2, 3, 5, 6, 7, 10],
        44.2,
        162,
        ['MVP', 'League Champion', 'League Champion'],
        true,
      ),
      batter(21, 'David', 'Ortiz', [2, 4, 5], 11.5, 138, ['League Champion', 'Conference Champion']),
    ],
  },
}
