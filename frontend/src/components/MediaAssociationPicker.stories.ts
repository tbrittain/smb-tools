import type { Meta, StoryObj } from '@storybook/vue3'
import MediaAssociationPicker from './MediaAssociationPicker.vue'

function stubWails() {
  // @ts-expect-error – deliberate window stub for Storybook
  window.go = {
    main: {
      App: {
        SearchTeamsForMediaPicker: (q: string) =>
          Promise.resolve(
            q
              ? [
                  { teamId: 1, teamName: 'Heaters', conferenceName: 'East', divisionName: 'North' },
                  { teamId: 2, teamName: 'Icebreakers', conferenceName: 'West', divisionName: 'South' },
                ]
              : [],
          ),
        GetTeamSeasonsForMediaPicker: () =>
          Promise.resolve([
            { teamHistoryId: 10, seasonNum: 3 },
            { teamHistoryId: 11, seasonNum: 4 },
            { teamHistoryId: 12, seasonNum: 5 },
          ]),
        SearchPlayers: (q: string) =>
          Promise.resolve(
            q
              ? [
                  {
                    playerId: 100,
                    firstName: 'John',
                    lastName: 'Doe',
                    isHallOfFamer: false,
                    seasonsPlayed: 5,
                    firstSeason: 1,
                    lastSeason: 5,
                  },
                  {
                    playerId: 101,
                    firstName: 'Jane',
                    lastName: 'Smith',
                    isHallOfFamer: true,
                    seasonsPlayed: 8,
                    firstSeason: 1,
                    lastSeason: 8,
                  },
                ]
              : [],
          ),
      },
    },
  }
}

const meta: Meta<typeof MediaAssociationPicker> = {
  title: 'Components/MediaAssociationPicker',
  component: MediaAssociationPicker,
}
export default meta

type Story = StoryObj<typeof MediaAssociationPicker>

export const TeamSeasonMode: Story = {
  render: (args) => ({
    components: { MediaAssociationPicker },
    setup() {
      stubWails()
      return { args }
    },
    template: '<MediaAssociationPicker v-bind="args" @picked="(t, id, label) => console.log(t, id, label)" />',
  }),
  args: { mode: 'team_season', alreadySelectedTeamHistoryIds: [11] },
}

export const PlayerMode: Story = {
  render: (args) => ({
    components: { MediaAssociationPicker },
    setup() {
      stubWails()
      return { args }
    },
    template: '<MediaAssociationPicker v-bind="args" @picked="(t, id, label) => console.log(t, id, label)" />',
  }),
  args: { mode: 'player', alreadySelectedPlayerIds: [100] },
}
