import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import AwardBadge from './AwardBadge.vue'

const meta: Meta<typeof AwardBadge> = {
  title: 'Components/AwardBadge',
  component: AwardBadge,
}
export default meta

type Story = StoryObj<typeof AwardBadge>

const leagueChampion = new main.AwardDTO({
  id: 1,
  name: 'League Champion',
  originalName: 'League Champion',
  importance: 1,
  omitFromGroupings: false,
  isBattingAward: false,
  isPitchingAward: false,
  isFieldingAward: false,
  isPlayoffAward: true,
  isUserAssignable: false,
  isBuiltIn: true,
})

const conferenceChampion = new main.AwardDTO({
  id: 2,
  name: 'Conference Champion',
  originalName: 'Conference Champion',
  importance: 1,
  omitFromGroupings: true,
  isBattingAward: false,
  isPitchingAward: false,
  isFieldingAward: false,
  isPlayoffAward: true,
  isUserAssignable: false,
  isBuiltIn: true,
})

const mvp = new main.AwardDTO({
  id: 3,
  name: 'MVP',
  originalName: 'MVP',
  importance: 0,
  omitFromGroupings: false,
  isBattingAward: false,
  isPitchingAward: false,
  isFieldingAward: false,
  isPlayoffAward: false,
  isUserAssignable: true,
  isBuiltIn: true,
})

export const LeagueChampion: Story = {
  args: { award: leagueChampion, count: 1 },
}

export const LeagueChampionMultiple: Story = {
  args: { award: leagueChampion, count: 3 },
}

export const ConferenceChampion: Story = {
  args: { award: conferenceChampion, count: 1 },
}

export const MVP: Story = {
  args: { award: mvp, count: 1 },
}
