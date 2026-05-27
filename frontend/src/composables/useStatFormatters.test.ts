import { describe, expect, it } from 'vitest'
import {
  formatBA,
  formatCount,
  formatERA,
  formatIP,
  formatK9,
  formatPct,
  formatSalary,
  formatSeasonRanges,
  formatWHIP,
} from './useStatFormatters'

describe('formatBA', () => {
  it('formats to 3 decimal places without leading zero', () => {
    expect(formatBA(0.304)).toBe('.304')
    expect(formatBA(0.406)).toBe('.406')
  })
  it('returns — for null', () => {
    expect(formatBA(null)).toBe('—')
  })
  it('returns — for undefined', () => {
    expect(formatBA(undefined)).toBe('—')
  })
})

describe('formatERA', () => {
  it('formats to 2 decimal places', () => {
    expect(formatERA(3.14)).toBe('3.14')
  })
  it('returns — for null', () => {
    expect(formatERA(null)).toBe('—')
  })
})

describe('formatIP', () => {
  it('converts outs to innings display', () => {
    expect(formatIP(0)).toBe('0.0')
    expect(formatIP(3)).toBe('1.0')
    expect(formatIP(97)).toBe('32.1')
    expect(formatIP(99)).toBe('33.0')
    expect(formatIP(100)).toBe('33.1')
    expect(formatIP(101)).toBe('33.2')
    expect(formatIP(102)).toBe('34.0')
  })
  it('rounds correctly — 1 out is 0.1, 2 outs is 0.2, 3 outs is 1.0', () => {
    expect(formatIP(1)).toBe('0.1')
    expect(formatIP(2)).toBe('0.2')
  })
})

describe('formatSalary', () => {
  it('formats with dollar sign and locale separators', () => {
    expect(formatSalary(4200)).toBe('$4,200')
    expect(formatSalary(1000000)).toBe('$1,000,000')
  })
})

describe('formatWHIP', () => {
  it('formats to 2 decimals', () => {
    expect(formatWHIP(1.18)).toBe('1.18')
  })
  it('returns — for null', () => {
    expect(formatWHIP(null)).toBe('—')
  })
})

describe('formatK9', () => {
  it('formats to 1 decimal', () => {
    expect(formatK9(9.2)).toBe('9.2')
  })
  it('returns — for null', () => {
    expect(formatK9(null)).toBe('—')
  })
})

describe('formatPct', () => {
  it('removes leading zero', () => {
    expect(formatPct(0.623)).toBe('.623')
  })
  it('returns — for null', () => {
    expect(formatPct(null)).toBe('—')
  })
})

describe('formatCount', () => {
  it('converts to string', () => {
    expect(formatCount(42)).toBe('42')
  })
  it('returns — for null', () => {
    expect(formatCount(null)).toBe('—')
  })
  it('rounds floats', () => {
    expect(formatCount(42.7)).toBe('43')
  })
})

describe('formatSeasonRanges', () => {
  it('returns — for empty array', () => {
    expect(formatSeasonRanges([])).toBe('—')
  })
  it('formats single season', () => {
    expect(formatSeasonRanges([5])).toBe('5')
  })
  it('formats contiguous range', () => {
    expect(formatSeasonRanges([1, 2, 3])).toBe('1–3')
  })
  it('formats non-contiguous seasons as comma-separated values', () => {
    expect(formatSeasonRanges([1, 3, 5])).toBe('1, 3, 5')
  })
  it('formats mixed contiguous and non-contiguous', () => {
    expect(formatSeasonRanges([1, 2, 4, 5, 7])).toBe('1–2, 4–5, 7')
  })
  it('sorts unsorted input before formatting', () => {
    expect(formatSeasonRanges([3, 1, 2])).toBe('1–3')
  })
})
