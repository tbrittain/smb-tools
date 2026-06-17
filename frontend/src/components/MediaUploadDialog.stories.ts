import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import MediaUploadDialog from './MediaUploadDialog.vue'

function stubWails() {
  // @ts-expect-error – deliberate window stub for Storybook
  window.go = {
    main: {
      App: {
        BrowseMediaFile: () => Promise.resolve(''),
        UploadMedia: () => Promise.resolve({}),
        SearchTeamsForMediaPicker: () =>
          Promise.resolve([
            { teamId: 1, teamName: 'Heaters', conferenceName: 'East', divisionName: 'North' },
            { teamId: 2, teamName: 'Icebreakers', conferenceName: 'West', divisionName: 'South' },
          ]),
        GetTeamSeasonsForMediaPicker: () =>
          Promise.resolve([
            { teamHistoryId: 10, seasonNum: 3 },
            { teamHistoryId: 11, seasonNum: 4 },
          ]),
        SearchPlayers: () =>
          Promise.resolve([
            {
              playerId: 100,
              firstName: 'John',
              lastName: 'Doe',
              isHallOfFamer: false,
              seasonsPlayed: 5,
              firstSeason: 1,
              lastSeason: 5,
            },
          ]),
      },
    },
  }
}

const meta: Meta<typeof MediaUploadDialog> = {
  title: 'Components/MediaUploadDialog',
  component: MediaUploadDialog,
}
export default meta

type Story = StoryObj<typeof MediaUploadDialog>

function makeStory(entityType: 'team_season' | 'player', entityLabel: string): Story {
  return {
    render: (args) => ({
      components: { MediaUploadDialog },
      setup() {
        stubWails()
        const visible = ref(true)
        return { args, visible }
      },
      template: '<MediaUploadDialog v-bind="args" v-model:visible="visible" />',
    }),
    args: { entityType, entityId: 1, entityLabel },
  }
}

export const TeamSeasonContext: Story = makeStory('team_season', 'Heaters S4')
export const PlayerContext: Story = makeStory('player', 'John Doe')
