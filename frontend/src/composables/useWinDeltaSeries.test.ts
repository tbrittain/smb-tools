import { describe, expect, it } from 'vitest'
import { main } from '../../wailsjs/go/models'
import { computeWinDeltaSeries } from './useWinDeltaSeries'

function makeGame(
  teamGameNum: number,
  homeHistId: number,
  awayHistId: number,
  homeScore: number | undefined,
  awayScore: number | undefined,
  homeTeamName = 'Home',
  awayTeamName = 'Away',
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

const MY = 1
const OPP = 2

describe('computeWinDeltaSeries', () => {
  it('returns empty array for empty schedule', () => {
    expect(computeWinDeltaSeries([], MY)).toEqual([])
  })

  it('skips unplayed games (null scores)', () => {
    const games = [
      makeGame(1, MY, OPP, 5, 3),
      makeGame(2, MY, OPP, undefined, undefined),
      makeGame(3, MY, OPP, undefined, undefined),
    ]
    const result = computeWinDeltaSeries(games, MY)
    expect(result).toHaveLength(1)
    expect(result[0].gameNum).toBe(1)
  })

  it('returns empty for all unplayed games', () => {
    const games = [makeGame(1, MY, OPP, undefined, undefined), makeGame(2, OPP, MY, undefined, undefined)]
    expect(computeWinDeltaSeries(games, MY)).toEqual([])
  })

  it('all wins: delta increments by 1 each game', () => {
    const games = [makeGame(1, MY, OPP, 5, 3), makeGame(2, MY, OPP, 7, 2), makeGame(3, MY, OPP, 4, 1)]
    const result = computeWinDeltaSeries(games, MY)
    expect(result.map((p) => p.delta)).toEqual([1, 2, 3])
    expect(result.every((p) => p.won)).toBe(true)
  })

  it('all losses: delta decrements by 1 each game', () => {
    const games = [makeGame(1, MY, OPP, 1, 5), makeGame(2, MY, OPP, 2, 7), makeGame(3, MY, OPP, 0, 4)]
    const result = computeWinDeltaSeries(games, MY)
    expect(result.map((p) => p.delta)).toEqual([-1, -2, -3])
    expect(result.every((p) => !p.won)).toBe(true)
  })

  it('mixed record: delta crosses 0 and recovers', () => {
    const games = [
      makeGame(1, MY, OPP, 5, 3), // +1
      makeGame(2, MY, OPP, 1, 6), // 0
      makeGame(3, MY, OPP, 2, 4), // -1
      makeGame(4, MY, OPP, 8, 2), // 0
      makeGame(5, MY, OPP, 3, 2), // +1
    ]
    const result = computeWinDeltaSeries(games, MY)
    expect(result.map((p) => p.delta)).toEqual([1, 0, -1, 0, 1])
  })

  it('correctly identifies win when team is home side', () => {
    // MY is home, wins 6-2
    const games = [makeGame(1, MY, OPP, 6, 2, 'MyTeam', 'OppTeam')]
    const result = computeWinDeltaSeries(games, MY)
    expect(result[0].won).toBe(true)
    expect(result[0].myScore).toBe(6)
    expect(result[0].oppScore).toBe(2)
    expect(result[0].opponentName).toBe('OppTeam')
  })

  it('correctly identifies win when team is away side', () => {
    // MY is away, wins 4-1
    const games = [makeGame(1, OPP, MY, 1, 4, 'OppTeam', 'MyTeam')]
    const result = computeWinDeltaSeries(games, MY)
    expect(result[0].won).toBe(true)
    expect(result[0].myScore).toBe(4)
    expect(result[0].oppScore).toBe(1)
    expect(result[0].opponentName).toBe('OppTeam')
  })

  it('correctly identifies loss when team is away side', () => {
    // MY is away, loses 1-5
    const games = [makeGame(1, OPP, MY, 5, 1, 'OppTeam', 'MyTeam')]
    const result = computeWinDeltaSeries(games, MY)
    expect(result[0].won).toBe(false)
    expect(result[0].myScore).toBe(1)
    expect(result[0].oppScore).toBe(5)
  })

  it('computes runDiff as absolute score difference', () => {
    const games = [
      makeGame(1, MY, OPP, 10, 3), // diff 7
      makeGame(2, MY, OPP, 2, 5), // diff 3
    ]
    const result = computeWinDeltaSeries(games, MY)
    expect(result[0].runDiff).toBe(7)
    expect(result[1].runDiff).toBe(3)
  })

  it('sorts by teamGameNum regardless of input order', () => {
    const games = [makeGame(3, MY, OPP, 4, 1), makeGame(1, MY, OPP, 5, 3), makeGame(2, MY, OPP, 1, 6)]
    const result = computeWinDeltaSeries(games, MY)
    expect(result.map((p) => p.gameNum)).toEqual([1, 2, 3])
    expect(result.map((p) => p.delta)).toEqual([1, 0, 1])
  })
})
