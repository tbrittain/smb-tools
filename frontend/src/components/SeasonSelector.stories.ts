import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import SeasonSelector from './SeasonSelector.vue'

const meta: Meta<typeof SeasonSelector> = {
  title: 'Components/SeasonSelector',
  component: SeasonSelector,
}
export default meta

type Story = StoryObj<typeof SeasonSelector>

const seasons = [
  new main.SeasonSummaryDTO({ id: 1, seasonNum: 1, numGames: 40, importedAt: '', championTeamName: 'Red Sox', championHistoryId: 10 }),
  new main.SeasonSummaryDTO({ id: 2, seasonNum: 2, numGames: 40, importedAt: '', championTeamName: 'Cubs', championHistoryId: 20 }),
  new main.SeasonSummaryDTO({ id: 3, seasonNum: 3, numGames: 40, importedAt: '', championTeamName: '' }),
]

export const Populated: Story = {
  args: { seasons, modelValue: 2 },
}

export const NoneSelected: Story = {
  args: { seasons, modelValue: null },
}

export const Empty: Story = {
  args: { seasons: [], modelValue: null },
}
