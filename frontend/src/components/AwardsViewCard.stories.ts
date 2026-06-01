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

function makeBatter(
  id: number,
  first: string,
  last: string,
  smbWar: number | undefined = 3.2,
  awardName = '',
): main.AwardWinnerRowDTO {
  return new main.AwardWinnerRowDTO({
    playerSeasonId: id * 10,
    playerId: id,
    firstName: first,
    lastName: last,
    teamName: 'Springfield Isotopes',
    primaryPosition: 'CF',
    pitcherRole: '',
    awardName,
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
  awardName = '',
): main.AwardWinnerRowDTO {
  return new main.AwardWinnerRowDTO({
    playerSeasonId: id * 10,
    playerId: id,
    firstName: first,
    lastName: last,
    teamName: 'Capital City Goofballs',
    primaryPosition: 'P',
    pitcherRole: 'SP',
    awardName,
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
      runnerUps: [],
    }),
  },
}

export const PitchingAwardWithRunnerUp: Story = {
  args: {
    group: new main.AwardGroupSummaryDTO({
      awardId: 2,
      awardName: 'Cy Young',
      winners: [makePitcher(201, 'Sandy', 'Koufaux', 4.1, 'Cy Young')],
      runnerUps: [makePitcher(202, 'Roger', 'Clemons', 3.0, 'Cy Young-2')],
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
      runnerUps: [],
    }),
  },
}

export const MultipleRunnerUps: Story = {
  name: 'Multiple runner-ups (MVP-2 and MVP-3 both assigned)',
  args: {
    group: new main.AwardGroupSummaryDTO({
      awardId: 1,
      awardName: 'MVP',
      winners: [makeBatter(401, 'Mike', 'Troutman', 3.2, 'MVP')],
      runnerUps: [
        makeBatter(402, 'Bryce', 'Harpington', 2.8, 'MVP-2'),
        makeBatter(403, 'Mookie', 'Bettsford', 2.1, 'MVP-3'),
      ],
    }),
  },
}

export const NullSmbWAR: Story = {
  name: 'Null smbWAR (pre-Phase-8.5 season)',
  args: {
    group: new main.AwardGroupSummaryDTO({
      awardId: 5,
      awardName: 'Silver Slugger',
      winners: [makeBatter(501, 'Ken', 'Griffey III', undefined, 'Silver Slugger')],
      runnerUps: [makeBatter(502, 'Tony', 'Gwynne Jr.', undefined, 'Silver Slugger-2')],
    }),
  },
}
