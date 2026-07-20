// Season Mode franchises have no aging/retirement, so features that depend on
// player career length (Hall of Fame induction) are meaningless and hidden.
export function isSeasonMode(leagueMode: string | undefined): boolean {
  return leagueMode === 'season'
}
