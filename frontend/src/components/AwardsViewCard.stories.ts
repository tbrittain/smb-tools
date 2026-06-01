import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import AwardsViewCard from './AwardsViewCard.vue'

const meta: Meta<typeof AwardsViewCard> = {
  title: 'Components/AwardsViewCard',
  component: AwardsViewCard,
}
export default meta

type Story = StoryObj<typeof AwardsViewCard>

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeBatter(id: number, first: string, last: string, smbWar: number | undefined = 3.2): main.AwardWinnerRowDTO {
  return new main.AwardWinnerRowDTO({
    playerSeasonId: id * 10,
    playerId: id,
    firstName: first,
    lastName: last,
    teamName: 'Springfield Isotopes',
    primaryPosition: 'CF',
    pitcherRole: '',
    ba: 0.312,
    hr: 28,
    rbi: 95,
    era: 0,
    wins: 0,
    strikeouts: 0,
    smbWar,
  })
}

function makePitcher(
  id: number,
  first: string,
  last: string,
  smbWar: number | undefined = 4.1,
): main.AwardWinnerRowDTO {
  return new main.AwardWinnerRowDTO({
    playerSeasonId: id * 10,
    playerId: id,
    firstName: first,
    lastName: last,
    teamName: 'Capital City Goofballs',
    primaryPosition: 'P',
    pitcherRole: 'SP',
    ba: 0,
    hr: 0,
    rbi: 0,
    era: 2.87,
    wins: 17,
    strikeouts: 198,
    smbWar,
  })
}

// ── Stories ───────────────────────────────────────────────────────────────────

export const BattingAwardSingleWinner: Story = {
  args: {
    group: new main.AwardGroupSummaryDTO({
      awardId: 1,
      awardName: 'MVP',
      winners: [makeBatter(101, 'Mike', 'Troutman')],
      runnerUp: undefined,
    }),
  },
}

export const PitchingAwardWithRunnerUp: Story = {
  args: {
    group: new main.AwardGroupSummaryDTO({
      awardId: 2,
      awardName: 'Cy Young',
      winners: [makePitcher(201, 'Sandy', 'Koufaux')],
      runnerUp: makePitcher(202, 'Roger', 'Clemons', 3.0),
    }),
  },
}

export const MultiWinnerAllStar: Story = {
  args: {
    group: new main.AwardGroupSummaryDTO({
      awardId: 3,
      awardName: 'All-Star',
      winners: [
        makeBatter(301, 'Albert', 'Pujolson'),
        makeBatter(302, 'Barry', 'Bonds Jr.'),
        makeBatter(303, 'Cal', 'Ripkenning'),
        makePitcher(304, 'Roger', 'Clemons'),
        makePitcher(305, 'Pedro', 'Martinique'),
      ],
      runnerUp: undefined,
    }),
  },
}

export const NullSmbWAR: Story = {
  name: 'Null smbWAR (pre-Phase-8.5 season)',
  args: {
    group: new main.AwardGroupSummaryDTO({
      awardId: 4,
      awardName: 'Silver Slugger',
      winners: [makeBatter(401, 'Ken', 'Griffey III', undefined)],
      runnerUp: makeBatter(402, 'Tony', 'Gwynne Jr.', undefined),
    }),
  },
}
