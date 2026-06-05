import type { main } from '../../wailsjs/go/models'

export interface WinDeltaPoint {
  gameNum: number
  delta: number
  runDiff: number
  won: boolean
  opponentName: string
  myScore: number
  oppScore: number
}

/**
 * Computes a team's cumulative win delta (wins minus losses) for each played
 * game in chronological order. Games with null scores (unplayed) are skipped.
 * Delta starts at 0 before game 1: a win makes it +1, a loss makes it -1.
 */
export function computeWinDeltaSeries(schedule: main.ScheduleGameDTO[], teamHistoryId: number): WinDeltaPoint[] {
  const points: WinDeltaPoint[] = []
  let delta = 0

  const sorted = [...schedule].sort((a, b) => a.teamGameNum - b.teamGameNum)

  for (const game of sorted) {
    if (game.homeScore == null || game.awayScore == null) continue

    const isHome = game.homeTeamHistoryId === teamHistoryId
    const myScore = isHome ? game.homeScore : game.awayScore
    const oppScore = isHome ? game.awayScore : game.homeScore
    const opponentName = isHome ? game.awayTeamName : game.homeTeamName
    const won = myScore > oppScore

    delta += won ? 1 : -1

    points.push({
      gameNum: game.teamGameNum,
      delta,
      runDiff: Math.abs(myScore - oppScore),
      won,
      opponentName,
      myScore,
      oppScore,
    })
  }

  return points
}
