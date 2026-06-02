import type { Meta, StoryObj } from '@storybook/vue3'
import MediaGallery from './MediaGallery.vue'

const SAMPLE_IMAGE_URL =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='

function makeItem(id: string, name: string, type: 'image' | 'video' = 'image') {
  return {
    id,
    name,
    description: '',
    mediaType: type,
    url: type === 'image' ? SAMPLE_IMAGE_URL : '',
    uploadedAt: '2025-03-01T12:00:00Z',
    totalAssociationCount: 1,
    teamSeasonAssocs: [],
    playerAssocs: [],
  }
}

function stubWails(items: object[], totalCount?: number) {
  // @ts-expect-error – deliberate window stub for Storybook
  window.go = {
    main: {
      App: {
        GetMediaForTeamSeason: () =>
          Promise.resolve({ items, totalCount: totalCount ?? items.length, page: 0, pageSize: 24 }),
        GetMediaForPlayer: () =>
          Promise.resolve({ items, totalCount: totalCount ?? items.length, page: 0, pageSize: 24 }),
        BrowseMediaFile: () => Promise.resolve(''),
        UploadMedia: () => Promise.resolve({}),
        SearchTeamsForMediaPicker: () => Promise.resolve([]),
        GetTeamSeasonsForMediaPicker: () => Promise.resolve([]),
        SearchPlayers: () => Promise.resolve([]),
        RemoveMediaAssociation: () => Promise.resolve(),
        DeleteMediaEverywhere: () => Promise.resolve(),
      },
    },
  }
}

const meta: Meta<typeof MediaGallery> = {
  title: 'Components/MediaGallery',
  component: MediaGallery,
}
export default meta

type Story = StoryObj<typeof MediaGallery>

export const Empty: Story = {
  render: (args) => ({
    components: { MediaGallery },
    setup() {
      stubWails([])
      return { args }
    },
    template: '<MediaGallery v-bind="args" />',
  }),
  args: { entityType: 'team_season', entityId: 1, entityLabel: 'Heaters S4' },
}

export const MixedMedia: Story = {
  render: (args) => ({
    components: { MediaGallery },
    setup() {
      stubWails([
        makeItem('1', 'Walk-off HR', 'image'),
        makeItem('2', 'No-hitter highlights', 'video'),
        makeItem('3', 'Championship celebration', 'image'),
        makeItem('4', 'Season opener', 'image'),
        makeItem('5', 'Playoff game 7', 'video'),
      ])
      return { args }
    },
    template: '<MediaGallery v-bind="args" />',
  }),
  args: { entityType: 'team_season', entityId: 1, entityLabel: 'Heaters S4' },
}

export const WithLoadMore: Story = {
  render: (args) => ({
    components: { MediaGallery },
    setup() {
      const firstPage = Array.from({ length: 6 }, (_, i) => makeItem(`img-${i}`, `Screenshot ${i + 1}`))
      stubWails(firstPage, 15)
      return { args }
    },
    template: '<MediaGallery v-bind="args" />',
  }),
  args: { entityType: 'player', entityId: 42, entityLabel: 'Jane Smith' },
}
