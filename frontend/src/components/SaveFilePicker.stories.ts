import type { Meta, StoryObj } from '@storybook/vue3'
import type { main } from '../../wailsjs/go/models'
import SaveFilePicker from './SaveFilePicker.vue'

const makeSave = (
  path: string,
  leagueName: string,
  playerTeamName: string,
  numSeasons: number,
  mode: main.SaveFileCandidateDTO['mode'] = 'franchise',
): main.SaveFileCandidateDTO =>
  ({
    path,
    gameVersion: 'smb4',
    leagueName,
    numSeasons,
    mode,
    isFranchise: mode === 'franchise',
    playerTeamName,
    leagueGUID: path,
  }) as main.SaveFileCandidateDTO

const sampleCandidates: main.SaveFileCandidateDTO[] = [
  makeSave(
    'C:\\Users\\Player\\AppData\\Local\\Metalhead\\Super Mega Baseball 4\\76561198000000001\\league-aaa.sav',
    'Super Mega League',
    'Honey Badgers',
    12,
  ),
  makeSave(
    'C:\\Users\\Player\\AppData\\Local\\Metalhead\\Super Mega Baseball 4\\76561198000000001\\league-bbb.sav',
    'Dynasty Mode',
    'Vapors',
    4,
  ),
  makeSave(
    'C:\\Users\\Player\\AppData\\Local\\Metalhead\\Super Mega Baseball 4\\76561198000000001\\league-ccc.sav',
    '',
    '',
    0,
  ),
]

// Season Mode saves have no player-controlled team — the player never gets
// assigned to a franchise-managed roster, so playerTeamName is always empty.
const seasonModeCandidates: main.SaveFileCandidateDTO[] = [
  makeSave(
    'C:\\Users\\Player\\AppData\\Local\\Metalhead\\Super Mega Baseball 4\\76561198000000001\\league-ddd.sav',
    'Weekend League',
    '',
    6,
    'season',
  ),
  ...sampleCandidates,
]

const meta: Meta<typeof SaveFilePicker> = {
  title: 'Franchise/SaveFilePicker',
  component: SaveFilePicker,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof SaveFilePicker>

export const AutoDetected: Story = {
  render: (args) => ({
    components: { SaveFilePicker },
    setup: () => ({ args }),
    template: '<SaveFilePicker v-bind="args" />',
  }),
  args: {
    candidates: sampleCandidates,
    loading: false,
    scanning: false,
    browsing: false,
    error: null,
    selectedPath: sampleCandidates[0].path,
  },
}

export const WithSeasonMode: Story = {
  render: (args) => ({
    components: { SaveFilePicker },
    setup: () => ({ args }),
    template: '<SaveFilePicker v-bind="args" />',
  }),
  args: {
    candidates: seasonModeCandidates,
    loading: false,
    scanning: false,
    browsing: false,
    error: null,
    selectedPath: seasonModeCandidates[0].path,
  },
}

export const Empty: Story = {
  render: (args) => ({
    components: { SaveFilePicker },
    setup: () => ({ args }),
    template: '<SaveFilePicker v-bind="args" />',
  }),
  args: {
    candidates: [],
    loading: false,
    scanning: false,
    browsing: false,
    error: null,
  },
}

export const Loading: Story = {
  render: (args) => ({
    components: { SaveFilePicker },
    setup: () => ({ args }),
    template: '<SaveFilePicker v-bind="args" />',
  }),
  args: {
    candidates: [],
    loading: true,
    scanning: false,
    browsing: false,
    error: null,
  },
}

export const Scanning: Story = {
  render: (args) => ({
    components: { SaveFilePicker },
    setup: () => ({ args }),
    template: '<SaveFilePicker v-bind="args" />',
  }),
  args: {
    candidates: sampleCandidates,
    loading: false,
    scanning: true,
    browsing: false,
    error: null,
  },
}

export const WithError: Story = {
  render: (args) => ({
    components: { SaveFilePicker },
    setup: () => ({ args }),
    template: '<SaveFilePicker v-bind="args" />',
  }),
  args: {
    candidates: [],
    loading: false,
    scanning: false,
    browsing: false,
    error: 'No Franchise or Season mode saves found in that folder. Elimination saves are not supported.',
  },
}

export const WithUsedLabels: Story = {
  render: (args) => ({
    components: { SaveFilePicker },
    setup: () => ({ args }),
    template: '<SaveFilePicker v-bind="args" />',
  }),
  args: {
    candidates: sampleCandidates,
    loading: false,
    scanning: false,
    browsing: false,
    error: null,
    selectedPath: sampleCandidates[1].path,
    usedSourceLabels: {
      [sampleCandidates[0].path]: 'Previously used · Seasons 1–8',
      [sampleCandidates[1].path]: 'Previously used · Seasons 9–12',
    },
  },
}

export const WithDisabledCandidates: Story = {
  render: (args) => ({
    components: { SaveFilePicker },
    setup: () => ({ args }),
    template: '<SaveFilePicker v-bind="args" />',
  }),
  args: {
    candidates: sampleCandidates,
    loading: false,
    scanning: false,
    browsing: false,
    error: null,
    disabledCandidateReasons: {
      [sampleCandidates[0].path]: 'Already connected. Fork requires a different SMB4 league/save.',
    },
  },
}

export const AllVariants: Story = {
  render: () => ({
    components: { SaveFilePicker },
    setup: () => ({ sampleCandidates }),
    template: `
      <div style="display:flex;flex-direction:column;gap:2.5rem;padding:1.5rem;background:#0d1117;max-width:520px">
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin:0 0 0.75rem">Auto-detected (with selection)</p>
          <SaveFilePicker
            :candidates="sampleCandidates"
            :selected-path="sampleCandidates[0].path"
          />
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin:0 0 0.75rem">Empty — no saves found</p>
          <SaveFilePicker :candidates="[]" />
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin:0 0 0.75rem">Scanning state</p>
          <SaveFilePicker :candidates="sampleCandidates" :scanning="true" />
        </div>
        <div>
          <p style="color:#8b949e;font-size:0.75rem;margin:0 0 0.75rem">Error state</p>
          <SaveFilePicker
            :candidates="[]"
            error="No Franchise or Season mode saves found in that folder."
          />
        </div>
      </div>
    `,
  }),
}
