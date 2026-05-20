import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import LeaderboardFilters from './LeaderboardFilters.vue'

const meta: Meta<typeof LeaderboardFilters> = {
  title: 'Components/LeaderboardFilters',
  component: LeaderboardFilters,
  decorators: [() => ({ template: '<div style="padding: 1.5rem; max-width: 900px"><story /></div>' })],
}
export default meta

type Story = StoryObj<typeof LeaderboardFilters>

const mockSeasons = [1, 2, 3, 4, 5].map(
  (n) =>
    new main.SeasonSummaryDTO({
      id: n,
      seasonNum: n,
      numGames: 100,
      importedAt: '2024-01-01T00:00:00Z',
      championTeamName: '',
      championHistoryId: null,
    }),
)

function defaultFilters(): main.LeaderboardFiltersDTO {
  return new main.LeaderboardFiltersDTO({
    isPlayoffs: false,
    onlyHallOfFamers: false,
    position: '',
    batHand: '',
    throwHand: '',
    chemistryType: '',
    seasonStart: 0,
    seasonEnd: 0,
  })
}

export const BattingMode: Story = {
  args: {
    mode: 'batting',
    seasons: mockSeasons,
    modelValue: defaultFilters(),
  },
}

export const PitchingMode: Story = {
  args: {
    mode: 'pitching',
    seasons: mockSeasons,
    modelValue: defaultFilters(),
  },
}

export const WithActiveFilters: Story = {
  args: {
    mode: 'batting',
    seasons: mockSeasons,
    modelValue: new main.LeaderboardFiltersDTO({
      ...defaultFilters(),
      position: 'SS',
      onlyHallOfFamers: true,
      seasonStart: 2,
      seasonEnd: 4,
    }),
  },
}
