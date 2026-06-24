import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import TeamLogoManager from './TeamLogoManager.vue'

const SAMPLE_LOGO =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='

const meta: Meta<typeof TeamLogoManager> = {
  title: 'Team/TeamLogoManager',
  component: TeamLogoManager,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof TeamLogoManager>

const BASE_PROPS = {
  teamId: 1,
  latestSeason: 8,
  availableSeasons: [1, 2, 3, 4, 5, 6, 7, 8],
}

// Renders the dialog in an always-open state for Storybook inspection.
function makeStory(logos: object[]): Story {
  return {
    render: (args) => ({
      components: { TeamLogoManager },
      setup() {
        const visible = ref(true)
        // Stub Wails bindings so the component renders without a real app.
        // @ts-expect-error – deliberate window stub for Storybook
        window.go = {
          main: {
            App: {
              GetTeamLogos: () => Promise.resolve(logos),
              BrowseLogoFile: () => Promise.resolve(''),
              UploadAndAssignTeamLogo: () => Promise.resolve({}),
              AssignExistingTeamLogo: () => Promise.resolve({}),
              DeleteTeamLogoAssignment: () => Promise.resolve(),
            },
          },
        }
        return { args, visible }
      },
      template: '<TeamLogoManager v-bind="args" v-model:visible="visible" />',
    }),
    args: { ...BASE_PROPS },
  }
}

export const EmptyState: Story = makeStory([])

export const OneLogoOneAssignment: Story = makeStory([
  {
    id: 'logo-1',
    teamId: 1,
    logoUrl: SAMPLE_LOGO,
    uploadedAt: '2024-03-01T00:00:00Z',
    assignments: [
      {
        id: 'assign-1',
        logoId: 'logo-1',
        startSeason: null,
        endSeason: null,
        assignedAt: '2024-03-01T00:00:00Z',
      },
    ],
  },
])

export const OneLogoTwoDisjointAssignments: Story = makeStory([
  {
    id: 'logo-1',
    teamId: 1,
    logoUrl: SAMPLE_LOGO,
    uploadedAt: '2024-03-01T00:00:00Z',
    assignments: [
      {
        id: 'assign-1',
        logoId: 'logo-1',
        startSeason: 1,
        endSeason: 5,
        assignedAt: '2024-03-01T00:00:00Z',
      },
      {
        id: 'assign-2',
        logoId: 'logo-1',
        startSeason: 10,
        endSeason: null,
        assignedAt: '2024-03-02T00:00:00Z',
      },
    ],
  },
])

export const MultipleLogos: Story = makeStory([
  {
    id: 'logo-1',
    teamId: 1,
    logoUrl: SAMPLE_LOGO,
    uploadedAt: '2024-01-01T00:00:00Z',
    assignments: [
      {
        id: 'assign-1',
        logoId: 'logo-1',
        startSeason: 1,
        endSeason: 4,
        assignedAt: '2024-01-01T00:00:00Z',
      },
    ],
  },
  {
    id: 'logo-2',
    teamId: 1,
    logoUrl: SAMPLE_LOGO,
    uploadedAt: '2024-06-01T00:00:00Z',
    assignments: [
      {
        id: 'assign-2',
        logoId: 'logo-2',
        startSeason: 5,
        endSeason: null,
        assignedAt: '2024-06-01T00:00:00Z',
      },
    ],
  },
])

export const UseExistingTab: Story = {
  render: (args) => ({
    components: { TeamLogoManager },
    setup() {
      const visible = ref(true)
      // @ts-expect-error – deliberate window stub for Storybook
      window.go = {
        main: {
          App: {
            GetTeamLogos: () =>
              Promise.resolve([
                {
                  id: 'logo-1',
                  teamId: 1,
                  logoUrl: SAMPLE_LOGO,
                  uploadedAt: '2024-01-01T00:00:00Z',
                  assignments: [],
                },
              ]),
            BrowseLogoFile: () => Promise.resolve(''),
            UploadAndAssignTeamLogo: () => Promise.resolve({}),
            AssignExistingTeamLogo: () => Promise.resolve({}),
            DeleteTeamLogoAssignment: () => Promise.resolve(),
          },
        },
      }
      return { args, visible }
    },
    template: '<TeamLogoManager v-bind="args" v-model:visible="visible" />',
  }),
  args: { ...BASE_PROPS },
}
