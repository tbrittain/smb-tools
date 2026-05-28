import type { main } from '../../wailsjs/go/models'

export const BATTING_COUNT_STATS = [
  'gamesPlayed',
  'atBats',
  'hits',
  'doubles',
  'triples',
  'homeRuns',
  'rbi',
  'stolenBases',
  'walks',
  'strikeouts',
] as const

export const PITCHING_COUNT_STATS = [
  'games',
  'gamesStarted',
  'wins',
  'losses',
  'saves',
  'outsPitched',
  'strikeouts',
  'walks',
  'hitsAllowed',
  'earnedRuns',
] as const

type StatHighlights = main.StatHighlightsDTO | null | undefined

function includesId(ids: number[] | undefined, playerId: number): boolean {
  return ids?.includes(playerId) ?? false
}

/**
 * Returns true if the player led the league in the given counting stat for the specified
 * regular season. Never applies to playoff data.
 */
export function isSeasonLeader(
  playerId: number,
  seasonNum: number,
  statKey: string,
  highlights: StatHighlights,
  type: 'batting' | 'pitching',
): boolean {
  if (!highlights) return false
  const leaders = type === 'batting' ? highlights.leagueLeadersBatting : highlights.leagueLeadersPitching
  return includesId(leaders?.[String(seasonNum)]?.[statKey], playerId)
}

/**
 * Returns true if the player-season holds the all-time single-season record for the
 * given counting stat (regular season only).
 */
export function isSingleSeasonRecord(
  playerId: number,
  seasonNum: number,
  statKey: string,
  highlights: StatHighlights,
  type: 'batting' | 'pitching',
): boolean {
  if (!highlights) return false
  const records = type === 'batting' ? highlights.singleSeasonBatting : highlights.singleSeasonPitching
  const holders = records?.[statKey]
  if (!holders) return false
  return holders.some((h) => h.playerId === playerId && h.seasonNum === seasonNum)
}

/**
 * Returns true if the player holds the all-time career record for the given counting stat
 * in the regular season.
 */
export function isCareerRecordRS(
  playerId: number,
  statKey: string,
  highlights: StatHighlights,
  type: 'batting' | 'pitching',
): boolean {
  if (!highlights) return false
  const records = type === 'batting' ? highlights.careerBattingRS : highlights.careerPitchingRS
  return includesId(records?.[statKey], playerId)
}

/**
 * Returns true if the player holds the all-time career record for the given counting stat
 * in the playoffs.
 */
export function isCareerRecordPO(
  playerId: number,
  statKey: string,
  highlights: StatHighlights,
  type: 'batting' | 'pitching',
): boolean {
  if (!highlights) return false
  const records = type === 'batting' ? highlights.careerBattingPO : highlights.careerPitchingPO
  return includesId(records?.[statKey], playerId)
}

/**
 * Builds the tooltip string for a highlighted stat cell. Returns an empty string
 * when the cell has no highlight for the given context.
 */
export function highlightTooltip(
  playerId: number,
  seasonNum: number,
  statKey: string,
  statLabel: string,
  highlights: StatHighlights,
  type: 'batting' | 'pitching',
  context: 'season' | 'careerRS' | 'careerPO',
): string {
  const parts: string[] = []

  if (context === 'season') {
    if (isSingleSeasonRecord(playerId, seasonNum, statKey, highlights, type)) {
      parts.push(`All-time single-season record in ${statLabel} (Season ${seasonNum})`)
    } else if (isSeasonLeader(playerId, seasonNum, statKey, highlights, type)) {
      parts.push(`Led the league in ${statLabel} (Season ${seasonNum})`)
    }
  } else if (context === 'careerRS') {
    if (isCareerRecordRS(playerId, statKey, highlights, type)) {
      parts.push(`All-time career record (Regular Season): ${statLabel}`)
    }
  } else if (context === 'careerPO') {
    if (isCareerRecordPO(playerId, statKey, highlights, type)) {
      parts.push(`All-time career record (Playoffs): ${statLabel}`)
    }
  }

  return parts.join(' · ')
}
