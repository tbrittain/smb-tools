import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import { main } from '../../wailsjs/go/models'
import RemoveMediaConfirmDialog from './RemoveMediaConfirmDialog.vue'

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

function makeItem(totalAssociationCount: number): main.MediaItemDTO {
  return main.MediaItemDTO.createFrom({
    id: 'media-1',
    name: 'Walk-off home run',
    description: '',
    mediaType: 'image',
    url: '',
    uploadedAt: '2025-03-01T12:00:00Z',
    totalAssociationCount,
    teamSeasonAssocs: [],
    playerAssocs: [],
  })
}

const meta: Meta<typeof RemoveMediaConfirmDialog> = {
  title: 'Components/RemoveMediaConfirmDialog',
  component: RemoveMediaConfirmDialog,
}
export default meta

type Story = StoryObj<typeof RemoveMediaConfirmDialog>

export const SingleAssociationDeleteOnly: Story = {
  render: (args) => ({
    components: { RemoveMediaConfirmDialog },
    setup() {
      stubWails()
      const visible = ref(true)
      return { args, visible }
    },
    template: '<RemoveMediaConfirmDialog v-bind="args" v-model:visible="visible" @removed="visible = false" />',
  }),
  args: {
    mediaItem: makeItem(1),
    contextEntityType: 'team_season',
    contextEntityId: 1,
    contextEntityLabel: 'Heaters S4',
  },
}

export const MultipleAssociationsChoice: Story = {
  render: (args) => ({
    components: { RemoveMediaConfirmDialog },
    setup() {
      stubWails()
      const visible = ref(true)
      return { args, visible }
    },
    template: '<RemoveMediaConfirmDialog v-bind="args" v-model:visible="visible" @removed="visible = false" />',
  }),
  args: {
    mediaItem: makeItem(3),
    contextEntityType: 'player',
    contextEntityId: 42,
    contextEntityLabel: 'John Doe',
  },
}
