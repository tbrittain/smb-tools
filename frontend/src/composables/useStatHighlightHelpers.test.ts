import { describe, expect, it } from 'vitest'
import type { main } from '../../wailsjs/go/models'
import {
  highlightTooltip,
  isCareerRecordPO,
  isCareerRecordRS,
  isRateCareerRecordPO,
  isRateCareerRecordRS,
  isRateSeasonLeader,
  isRateSingleSeasonRecord,
  isSeasonLeader,
  isSingleSeasonRecord,
  rateHighlightTooltip,
} from './useStatHighlightHelpers'

function makeHighlights(overrides: Partial<main.StatHighlightsDTO> = {}): main.StatHighlightsDTO {
  return {
    leagueLeadersBatting: {},
    leagueLeadersPitching: {},
    singleSeasonBatting: {},
    singleSeasonPitching: {},
    careerBattingRS: {},
    careerBattingPO: {},
    careerPitchingRS: {},
    careerPitchingPO: {},
    ...overrides,
  } as main.StatHighlightsDTO
}

describe('isSeasonLeader', () => {
  it('returns false when highlights is null', () => {
    expect(isSeasonLeader(1, 1, 'homeRuns', null, 'batting')).toBe(false)
  })

  it('returns false when highlights is undefined', () => {
    expect(isSeasonLeader(1, 1, 'homeRuns', undefined, 'batting')).toBe(false)
  })

  it('returns true when player is the sole league leader', () => {
    const h = makeHighlights({ leagueLeadersBatting: { '1': { homeRuns: [42] } } })
    expect(isSeasonLeader(42, 1, 'homeRuns', h, 'batting')).toBe(true)
  })

  it('returns true for tied leader', () => {
    const h = makeHighlights({ leagueLeadersBatting: { '1': { homeRuns: [42, 99] } } })
    expect(isSeasonLeader(99, 1, 'homeRuns', h, 'batting')).toBe(true)
    expect(isSeasonLeader(42, 1, 'homeRuns', h, 'batting')).toBe(true)
  })

  it('returns false when player is not in the leader list', () => {
    const h = makeHighlights({ leagueLeadersBatting: { '1': { homeRuns: [42] } } })
    expect(isSeasonLeader(99, 1, 'homeRuns', h, 'batting')).toBe(false)
  })

  it('returns false when looking at wrong season', () => {
    const h = makeHighlights({ leagueLeadersBatting: { '1': { homeRuns: [42] } } })
    expect(isSeasonLeader(42, 2, 'homeRuns', h, 'batting')).toBe(false)
  })

  it('checks pitching leaders separately from batting', () => {
    const h = makeHighlights({
      leagueLeadersBatting: { '1': { strikeouts: [42] } },
      leagueLeadersPitching: { '1': { strikeouts: [99] } },
    })
    expect(isSeasonLeader(42, 1, 'strikeouts', h, 'batting')).toBe(true)
    expect(isSeasonLeader(42, 1, 'strikeouts', h, 'pitching')).toBe(false)
    expect(isSeasonLeader(99, 1, 'strikeouts', h, 'pitching')).toBe(true)
  })
})

describe('isSingleSeasonRecord', () => {
  it('returns false when highlights is null', () => {
    expect(isSingleSeasonRecord(1, 1, 'homeRuns', null, 'batting')).toBe(false)
  })

  it('returns true when player-season matches the record holder', () => {
    const h = makeHighlights({
      singleSeasonBatting: { homeRuns: [{ playerId: 42, seasonNum: 3 }] },
    })
    expect(isSingleSeasonRecord(42, 3, 'homeRuns', h, 'batting')).toBe(true)
  })

  it('returns false when player matches but season does not', () => {
    const h = makeHighlights({
      singleSeasonBatting: { homeRuns: [{ playerId: 42, seasonNum: 3 }] },
    })
    expect(isSingleSeasonRecord(42, 4, 'homeRuns', h, 'batting')).toBe(false)
  })

  it('returns false when season matches but player does not', () => {
    const h = makeHighlights({
      singleSeasonBatting: { homeRuns: [{ playerId: 42, seasonNum: 3 }] },
    })
    expect(isSingleSeasonRecord(99, 3, 'homeRuns', h, 'batting')).toBe(false)
  })

  it('returns true for each tied record holder', () => {
    const h = makeHighlights({
      singleSeasonBatting: {
        homeRuns: [
          { playerId: 42, seasonNum: 3 },
          { playerId: 99, seasonNum: 5 },
        ],
      },
    })
    expect(isSingleSeasonRecord(42, 3, 'homeRuns', h, 'batting')).toBe(true)
    expect(isSingleSeasonRecord(99, 5, 'homeRuns', h, 'batting')).toBe(true)
  })

  it('checks pitching records separately from batting', () => {
    const h = makeHighlights({
      singleSeasonBatting: { strikeouts: [{ playerId: 42, seasonNum: 1 }] },
      singleSeasonPitching: { strikeouts: [{ playerId: 99, seasonNum: 2 }] },
    })
    expect(isSingleSeasonRecord(42, 1, 'strikeouts', h, 'batting')).toBe(true)
    expect(isSingleSeasonRecord(42, 1, 'strikeouts', h, 'pitching')).toBe(false)
    expect(isSingleSeasonRecord(99, 2, 'strikeouts', h, 'pitching')).toBe(true)
  })
})

describe('isCareerRecordRS', () => {
  it('returns false when highlights is null', () => {
    expect(isCareerRecordRS(1, 'homeRuns', null, 'batting')).toBe(false)
  })

  it('returns true when player holds the career RS record', () => {
    const h = makeHighlights({ careerBattingRS: { homeRuns: [42] } })
    expect(isCareerRecordRS(42, 'homeRuns', h, 'batting')).toBe(true)
  })

  it('returns true for each player in a tied career record', () => {
    const h = makeHighlights({ careerBattingRS: { homeRuns: [42, 99] } })
    expect(isCareerRecordRS(42, 'homeRuns', h, 'batting')).toBe(true)
    expect(isCareerRecordRS(99, 'homeRuns', h, 'batting')).toBe(true)
  })

  it('returns false when player does not hold the record', () => {
    const h = makeHighlights({ careerBattingRS: { homeRuns: [42] } })
    expect(isCareerRecordRS(99, 'homeRuns', h, 'batting')).toBe(false)
  })

  it('checks pitching records separately from batting', () => {
    const h = makeHighlights({
      careerBattingRS: { strikeouts: [42] },
      careerPitchingRS: { strikeouts: [99] },
    })
    expect(isCareerRecordRS(42, 'strikeouts', h, 'batting')).toBe(true)
    expect(isCareerRecordRS(42, 'strikeouts', h, 'pitching')).toBe(false)
    expect(isCareerRecordRS(99, 'strikeouts', h, 'pitching')).toBe(true)
  })
})

describe('isCareerRecordPO', () => {
  it('returns false when highlights is null', () => {
    expect(isCareerRecordPO(1, 'homeRuns', null, 'batting')).toBe(false)
  })

  it('returns true when player holds the career PO record', () => {
    const h = makeHighlights({ careerBattingPO: { homeRuns: [42] } })
    expect(isCareerRecordPO(42, 'homeRuns', h, 'batting')).toBe(true)
  })

  it('returns false when player does not hold the PO record', () => {
    const h = makeHighlights({ careerBattingPO: { homeRuns: [42] } })
    expect(isCareerRecordPO(99, 'homeRuns', h, 'batting')).toBe(false)
  })

  it('checks pitching PO records separately from batting', () => {
    const h = makeHighlights({
      careerBattingPO: { strikeouts: [42] },
      careerPitchingPO: { strikeouts: [99] },
    })
    expect(isCareerRecordPO(42, 'strikeouts', h, 'batting')).toBe(true)
    expect(isCareerRecordPO(99, 'strikeouts', h, 'pitching')).toBe(true)
    expect(isCareerRecordPO(42, 'strikeouts', h, 'pitching')).toBe(false)
  })
})

describe('highlightTooltip', () => {
  it('returns empty string when highlights is null', () => {
    expect(highlightTooltip(1, 1, 'homeRuns', 'HR', null, 'batting', 'season')).toBe('')
  })

  it('returns empty string when player has no highlight for the stat', () => {
    const h = makeHighlights()
    expect(highlightTooltip(1, 1, 'homeRuns', 'HR', h, 'batting', 'season')).toBe('')
  })

  it('returns leader tooltip for season leader only', () => {
    const h = makeHighlights({ leagueLeadersBatting: { '3': { homeRuns: [42] } } })
    expect(highlightTooltip(42, 3, 'homeRuns', 'HR', h, 'batting', 'season')).toBe('Led the league in HR (Season 3)')
  })

  it('returns record tooltip for single-season record only', () => {
    const h = makeHighlights({
      singleSeasonBatting: { homeRuns: [{ playerId: 42, seasonNum: 3 }] },
    })
    expect(highlightTooltip(42, 3, 'homeRuns', 'HR', h, 'batting', 'season')).toBe(
      'All-time single-season record in HR (Season 3)',
    )
  })

  it('shows only record tooltip when player is both leader and all-time record holder', () => {
    const h = makeHighlights({
      leagueLeadersBatting: { '3': { homeRuns: [42] } },
      singleSeasonBatting: { homeRuns: [{ playerId: 42, seasonNum: 3 }] },
    })
    expect(highlightTooltip(42, 3, 'homeRuns', 'HR', h, 'batting', 'season')).toBe(
      'All-time single-season record in HR (Season 3)',
    )
  })

  it('returns career RS tooltip', () => {
    const h = makeHighlights({ careerBattingRS: { homeRuns: [42] } })
    expect(highlightTooltip(42, 0, 'homeRuns', 'HR', h, 'batting', 'careerRS')).toBe(
      'All-time career record (Regular Season): HR',
    )
  })

  it('returns career PO tooltip', () => {
    const h = makeHighlights({ careerBattingPO: { homeRuns: [42] } })
    expect(highlightTooltip(42, 0, 'homeRuns', 'HR', h, 'batting', 'careerPO')).toBe(
      'All-time career record (Playoffs): HR',
    )
  })

  it('returns empty string for careerRS context when player does not hold the record', () => {
    const h = makeHighlights({ careerBattingRS: { homeRuns: [99] } })
    expect(highlightTooltip(42, 0, 'homeRuns', 'HR', h, 'batting', 'careerRS')).toBe('')
  })

  it('uses pitching highlights when type is pitching', () => {
    const h = makeHighlights({
      leagueLeadersPitching: { '2': { strikeouts: [7] } },
    })
    expect(highlightTooltip(7, 2, 'strikeouts', 'K', h, 'pitching', 'season')).toBe('Led the league in K (Season 2)')
    expect(highlightTooltip(7, 2, 'strikeouts', 'K', h, 'batting', 'season')).toBe('')
  })
})

// ── Rate stat helpers ─────────────────────────────────────────────────────────

describe('isRateSeasonLeader', () => {
  it('returns false when highlights is null', () => {
    expect(isRateSeasonLeader(1, 1, 'ba', null, 'batting')).toBe(false)
  })

  it('returns false when highlights is undefined', () => {
    expect(isRateSeasonLeader(1, 1, 'ba', undefined, 'batting')).toBe(false)
  })

  it('returns true when player is in leagueLeadersBattingRate for the season', () => {
    const h = makeHighlights({ leagueLeadersBattingRate: { '1': { ba: [42] } } })
    expect(isRateSeasonLeader(42, 1, 'ba', h, 'batting')).toBe(true)
  })

  it('returns false when player is not in the rate leader list', () => {
    const h = makeHighlights({ leagueLeadersBattingRate: { '1': { ba: [99] } } })
    expect(isRateSeasonLeader(42, 1, 'ba', h, 'batting')).toBe(false)
  })

  it('uses pitching leaders when type is pitching', () => {
    const h = makeHighlights({ leagueLeadersPitchingRate: { '3': { era: [10] } } })
    expect(isRateSeasonLeader(10, 3, 'era', h, 'pitching')).toBe(true)
    expect(isRateSeasonLeader(10, 3, 'era', h, 'batting')).toBe(false)
  })
})

describe('isRateSingleSeasonRecord', () => {
  it('returns true when playerId and seasonNum both match', () => {
    const h = makeHighlights({
      singleSeasonBattingRate: { ba: [{ playerId: 42, seasonNum: 5 }] },
    })
    expect(isRateSingleSeasonRecord(42, 5, 'ba', h, 'batting')).toBe(true)
  })

  it('returns false when seasonNum does not match', () => {
    const h = makeHighlights({
      singleSeasonBattingRate: { ba: [{ playerId: 42, seasonNum: 5 }] },
    })
    expect(isRateSingleSeasonRecord(42, 6, 'ba', h, 'batting')).toBe(false)
  })

  it('returns false when highlights is null', () => {
    expect(isRateSingleSeasonRecord(1, 1, 'ba', null, 'batting')).toBe(false)
  })
})

describe('isRateCareerRecordRS', () => {
  it('returns true when player is in careerBattingRSRate', () => {
    const h = makeHighlights({ careerBattingRSRate: { obp: [7] } })
    expect(isRateCareerRecordRS(7, 'obp', h, 'batting')).toBe(true)
  })

  it('returns false when player is not in the list', () => {
    const h = makeHighlights({ careerBattingRSRate: { obp: [7] } })
    expect(isRateCareerRecordRS(99, 'obp', h, 'batting')).toBe(false)
  })

  it('uses pitching map when type is pitching', () => {
    const h = makeHighlights({ careerPitchingRSRate: { era: [10] } })
    expect(isRateCareerRecordRS(10, 'era', h, 'pitching')).toBe(true)
    expect(isRateCareerRecordRS(10, 'era', h, 'batting')).toBe(false)
  })
})

describe('isRateCareerRecordPO', () => {
  it('returns true when player is in careerPitchingPORate', () => {
    const h = makeHighlights({ careerPitchingPORate: { whip: [5] } })
    expect(isRateCareerRecordPO(5, 'whip', h, 'pitching')).toBe(true)
  })

  it('returns false when highlights is null', () => {
    expect(isRateCareerRecordPO(1, 'whip', null, 'pitching')).toBe(false)
  })
})

describe('rateHighlightTooltip', () => {
  it('returns season leader text when player leads in rate stat', () => {
    const h = makeHighlights({ leagueLeadersBattingRate: { '2': { ba: [42] } } })
    expect(rateHighlightTooltip(42, 2, 'ba', 'BA', h, 'batting', 'season')).toBe('Led the league in BA (Season 2)')
  })

  it('returns single-season record text when player holds the record', () => {
    const h = makeHighlights({
      singleSeasonBattingRate: { ba: [{ playerId: 42, seasonNum: 3 }] },
    })
    expect(rateHighlightTooltip(42, 3, 'ba', 'BA', h, 'batting', 'season')).toBe(
      'All-time single-season record in BA (Season 3)',
    )
  })

  it('single-season record takes precedence over league leader in tooltip', () => {
    const h = makeHighlights({
      leagueLeadersBattingRate: { '3': { ba: [42] } },
      singleSeasonBattingRate: { ba: [{ playerId: 42, seasonNum: 3 }] },
    })
    const tip = rateHighlightTooltip(42, 3, 'ba', 'BA', h, 'batting', 'season')
    expect(tip).toBe('All-time single-season record in BA (Season 3)')
  })

  it('returns career RS record text', () => {
    const h = makeHighlights({ careerBattingRSRate: { ops: [7] } })
    expect(rateHighlightTooltip(7, 0, 'ops', 'OPS', h, 'batting', 'careerRS')).toBe(
      'All-time career record (Regular Season): OPS',
    )
  })

  it('returns empty string when no highlight applies', () => {
    const h = makeHighlights()
    expect(rateHighlightTooltip(1, 1, 'ba', 'BA', h, 'batting', 'season')).toBe('')
  })
})

describe('opsPlus/eraPlus/fipMinus season rate highlights', () => {
  it('isRateSeasonLeader returns true for opsPlus in batting rate leaders', () => {
    const h = makeHighlights({ leagueLeadersBattingRate: { '5': { opsPlus: [42] } } })
    expect(isRateSeasonLeader(42, 5, 'opsPlus', h, 'batting')).toBe(true)
    expect(isRateSeasonLeader(99, 5, 'opsPlus', h, 'batting')).toBe(false)
  })

  it('isRateSingleSeasonRecord returns true for eraPlus in pitching rate records', () => {
    const h = makeHighlights({
      singleSeasonPitchingRate: { eraPlus: [{ playerId: 10, seasonNum: 3 }] },
    })
    expect(isRateSingleSeasonRecord(10, 3, 'eraPlus', h, 'pitching')).toBe(true)
    expect(isRateSingleSeasonRecord(10, 4, 'eraPlus', h, 'pitching')).toBe(false)
  })

  it('isRateSeasonLeader returns true for fipMinus in pitching rate leaders', () => {
    const h = makeHighlights({ leagueLeadersPitchingRate: { '2': { fipMinus: [10] } } })
    expect(isRateSeasonLeader(10, 2, 'fipMinus', h, 'pitching')).toBe(true)
  })

  it('rateHighlightTooltip formats OPS+ correctly', () => {
    const h = makeHighlights({ leagueLeadersBattingRate: { '4': { opsPlus: [42] } } })
    expect(rateHighlightTooltip(42, 4, 'opsPlus', 'OPS+', h, 'batting', 'season')).toBe(
      'Led the league in OPS+ (Season 4)',
    )
  })

  it('rateHighlightTooltip formats ERA+ record correctly', () => {
    const h = makeHighlights({
      singleSeasonPitchingRate: { eraPlus: [{ playerId: 10, seasonNum: 6 }] },
    })
    expect(rateHighlightTooltip(10, 6, 'eraPlus', 'ERA+', h, 'pitching', 'season')).toBe(
      'All-time single-season record in ERA+ (Season 6)',
    )
  })

  it('opsPlus does not appear in careerBattingRSRate (not tracked at career level)', () => {
    const h = makeHighlights({ careerBattingRSRate: { ops: [7] } })
    expect(isRateCareerRecordRS(7, 'opsPlus', h, 'batting')).toBe(false)
  })
})
