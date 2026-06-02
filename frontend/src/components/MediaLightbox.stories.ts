import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import { main } from '../../wailsjs/go/models'
import MediaLightbox from './MediaLightbox.vue'

const SAMPLE_IMAGE_URL =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='

function stubWails() {
  // @ts-expect-error – deliberate window stub for Storybook
  window.go = {
    main: {
      App: {
        RemoveMediaAssociation: () => Promise.resolve(),
        DeleteMediaEverywhere: () => Promise.resolve(),
      },
    },
  }
}

function makeItem(id: string, name: string, type: 'image' | 'video', description = ''): main.MediaItemDTO {
  return main.MediaItemDTO.createFrom({
    id,
    name,
    description,
    mediaType: type,
    url: type === 'image' ? SAMPLE_IMAGE_URL : '',
    uploadedAt: '2025-03-01T12:00:00Z',
    totalAssociationCount: 2,
    teamSeasonAssocs: [{ teamHistoryId: 1, teamName: 'Heaters', seasonNum: 4 }],
    playerAssocs: [],
  })
}

const MULTI_ITEMS = [
  makeItem('img-1', 'Walk-off home run', 'image', 'Bottom of the 9th, season 4 clincher'),
  makeItem('vid-1', 'No-hitter highlights', 'video'),
  makeItem('img-2', 'Championship celebration', 'image'),
]

const meta: Meta<typeof MediaLightbox> = {
  title: 'Components/MediaLightbox',
  component: MediaLightbox,
}
export default meta

type Story = StoryObj<typeof MediaLightbox>

export const SingleImage: Story = {
  render: (args) => ({
    components: { MediaLightbox },
    setup() {
      stubWails()
      const visible = ref(true)
      return { args, visible }
    },
    template: `<MediaLightbox v-if="visible" v-bind="args" @close="visible = false" @removed="visible = false" />`,
  }),
  args: {
    items: [makeItem('img-1', 'Walk-off home run', 'image', 'Bottom of the 9th')],
    initialIndex: 0,
    entityType: 'team_season',
    entityId: 1,
    entityLabel: 'Heaters S4',
  },
}

export const SingleVideo: Story = {
  render: (args) => ({
    components: { MediaLightbox },
    setup() {
      stubWails()
      const visible = ref(true)
      return { args, visible }
    },
    template: `<MediaLightbox v-if="visible" v-bind="args" @close="visible = false" @removed="visible = false" />`,
  }),
  args: {
    items: [makeItem('vid-1', 'No-hitter highlights', 'video', 'Full 9-inning no-hitter, season 4')],
    initialIndex: 0,
    entityType: 'team_season',
    entityId: 1,
    entityLabel: 'Heaters S4',
  },
}

export const MultiItemNavigation: Story = {
  render: (args) => ({
    components: { MediaLightbox },
    setup() {
      stubWails()
      const visible = ref(true)
      return { args, visible }
    },
    template: `<MediaLightbox v-if="visible" v-bind="args" @close="visible = false" @removed="visible = false" />`,
  }),
  args: {
    items: MULTI_ITEMS,
    initialIndex: 1,
    entityType: 'team_season',
    entityId: 1,
    entityLabel: 'Heaters S4',
  },
}
