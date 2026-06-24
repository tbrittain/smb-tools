import type { Meta, StoryObj } from '@storybook/vue3'
import { main } from '../../wailsjs/go/models'
import SeasonWinDeltaChart from './SeasonWinDeltaChart.vue'

const meta: Meta<typeof SeasonWinDeltaChart> = {
  title: 'Charts/SeasonWinDeltaChart',
  component: SeasonWinDeltaChart,
}
export default meta

type Story = StoryObj<typeof SeasonWinDeltaChart>

const MY_HIST = 1
const OPP1_HIST = 2

function makeGame(
  teamGameNum: number,
  homeHistId: number,
  awayHistId: number,
  homeScore: number | undefined,
  awayScore: number | undefined,
  homeTeamName = 'MyTeam',
  awayTeamName = 'Opponent',
): main.ScheduleGameDTO {
  return new main.ScheduleGameDTO({
    teamGameNum,
    gameNumber: teamGameNum,
    day: teamGameNum,
    homeTeamHistoryId: homeHistId,
    homeTeamName,
    homeTeamId: homeHistId,
    awayTeamHistoryId: awayHistId,
    awayTeamName,
    awayTeamId: awayHistId,
    homeScore,
    awayScore,
    homePitcherName: '',
    awayPitcherName: '',
  })
}

/** Build a schedule where the team wins `wins` games and loses `losses` games
 *  in an alternating pattern, starting with a win. Scores are fixed. */
function buildSchedule(wins: number, losses: number): main.ScheduleGameDTO[] {
  const total = wins + losses
  const games: main.ScheduleGameDTO[] = []
  let w = 0
  let l = 0
  for (let i = 1; i <= total; i++) {
    const shouldWin = w < wins && (l >= losses || w <= l)
    if (shouldWin) {
      games.push(makeGame(i, MY_HIST, OPP1_HIST, 5, 3))
      w++
    } else {
      games.push(makeGame(i, MY_HIST, OPP1_HIST, 2, 7))
      l++
    }
  }
  return games
}

// 50-game season — balanced record
const balancedSchedule = buildSchedule(25, 25)

// Dominant season — 40 wins, 10 losses
const dominantSchedule = buildSchedule(40, 10)

// Poor season — 10 wins, 40 losses
const poorSchedule = buildSchedule(10, 40)

// Unplayed games (season in progress — first 20 played)
const inProgressSchedule = [
  ...buildSchedule(12, 8),
  ...Array.from({ length: 10 }, (_, i) => makeGame(21 + i, MY_HIST, OPP1_HIST, undefined, undefined)),
]

export const Balanced: Story = {
  args: {
    currentTeamHistoryId: MY_HIST,
    currentTeamName: 'Riverside Rockets',
    currentTeamDivisionName: 'East',
    currentTeamSeasonId: 1,
    schedule: balancedSchedule,
  },
}

export const DominantSeason: Story = {
  args: {
    currentTeamHistoryId: MY_HIST,
    currentTeamName: 'Riverside Rockets',
    currentTeamDivisionName: 'East',
    currentTeamSeasonId: 1,
    schedule: dominantSchedule,
  },
}

export const PoorSeason: Story = {
  args: {
    currentTeamHistoryId: MY_HIST,
    currentTeamName: 'Riverside Rockets',
    currentTeamDivisionName: 'East',
    currentTeamSeasonId: 1,
    schedule: poorSchedule,
  },
}

export const SeasonInProgress: Story = {
  args: {
    currentTeamHistoryId: MY_HIST,
    currentTeamName: 'Riverside Rockets',
    currentTeamDivisionName: 'East',
    currentTeamSeasonId: 1,
    schedule: inProgressSchedule,
  },
}

export const NoPlayedGames: Story = {
  args: {
    currentTeamHistoryId: MY_HIST,
    currentTeamName: 'Riverside Rockets',
    currentTeamDivisionName: 'East',
    currentTeamSeasonId: 1,
    schedule: Array.from({ length: 5 }, (_, i) => makeGame(i + 1, MY_HIST, OPP1_HIST, undefined, undefined)),
  },
}

export const AllVariants: Story = {
  render: () => ({
    components: { SeasonWinDeltaChart },
    template: `
      <div style="background:#1e1e2e;padding:2rem;display:flex;flex-direction:column;gap:3rem">
        <div>
          <h3 style="color:#cdd6f4;font-size:0.875rem;margin:0 0 0.75rem">Balanced season (25–25)</h3>
          <SeasonWinDeltaChart
            :current-team-history-id="myHist"
            current-team-name="Riverside Rockets"
            current-team-division-name="East"
            :current-team-season-id="1"
            :schedule="balanced"
          />
        </div>
        <div>
          <h3 style="color:#cdd6f4;font-size:0.875rem;margin:0 0 0.75rem">Dominant season (40–10)</h3>
          <SeasonWinDeltaChart
            :current-team-history-id="myHist"
            current-team-name="Riverside Rockets"
            current-team-division-name="East"
            :current-team-season-id="2"
            :schedule="dominant"
          />
        </div>
        <div>
          <h3 style="color:#cdd6f4;font-size:0.875rem;margin:0 0 0.75rem">Poor season (10–40)</h3>
          <SeasonWinDeltaChart
            :current-team-history-id="myHist"
            current-team-name="Riverside Rockets"
            current-team-division-name="East"
            :current-team-season-id="3"
            :schedule="poor"
          />
        </div>
      </div>
    `,
    setup() {
      return { myHist: MY_HIST, balanced: balancedSchedule, dominant: dominantSchedule, poor: poorSchedule }
    },
  }),
}
