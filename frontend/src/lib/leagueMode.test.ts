import { describe, expect, it } from 'vitest'
import { isSeasonMode } from './leagueMode'

describe('isSeasonMode', () => {
  it('returns true for "season"', () => {
    expect(isSeasonMode('season')).toBe(true)
  })

  it('returns false for "franchise"', () => {
    expect(isSeasonMode('franchise')).toBe(false)
  })

  it('returns false for undefined (no active franchise)', () => {
    expect(isSeasonMode(undefined)).toBe(false)
  })

  it('returns false for an empty string', () => {
    expect(isSeasonMode('')).toBe(false)
  })
})
