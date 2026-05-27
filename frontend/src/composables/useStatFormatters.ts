/**
 * Pure stat formatting functions used across player and team stat tables.
 * All functions return '—' for null/undefined/zero-denominator cases.
 */

export function formatBA(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(3).replace(/^0/, '') // ".304" not "0.304"
}

export function formatERA(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(2)
}

export function formatWHIP(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(2)
}

export function formatK9(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(1)
}

export function formatPct(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(3).replace(/^0/, '')
}

/** Converts outs_pitched to the display string "32.1" (X whole innings + remainder outs). */
export function formatIP(outsPitched: number): string {
  const whole = Math.floor(outsPitched / 3)
  const remainder = outsPitched % 3
  return `${whole}.${remainder}`
}

/** Formats OPS+, ERA+, FIP- as an integer (e.g. 115, 87). */
export function formatAdjustedStat(v: number | null | undefined): string {
  if (v == null) return '—'
  return Math.round(v).toString()
}

/** Formats FIP like ERA: two decimal places. */
export function formatFIP(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(2)
}

/** Formats smbWAR to one decimal place. */
export function formatWAR(v: number | null | undefined): string {
  if (v == null) return '—'
  return v.toFixed(1)
}

export function formatSalary(v: number): string {
  return `$${v.toLocaleString()}`
}

export function formatCount(v: number | null | undefined): string {
  if (v == null) return '—'
  return String(Math.round(v))
}

// [1,2,3,5,6] → "1–3, 5–6"   [1,3,5] → "1, 3, 5"
export function formatSeasonRanges(nums: number[]): string {
  if (nums.length === 0) return '—'
  const sorted = [...nums].sort((a, b) => a - b)
  const ranges: string[] = []
  let start = sorted[0]
  let end = sorted[0]
  for (let i = 1; i < sorted.length; i++) {
    if (sorted[i] === end + 1) {
      end = sorted[i]
    } else {
      ranges.push(start === end ? `${start}` : `${start}–${end}`)
      start = sorted[i]
      end = sorted[i]
    }
  }
  ranges.push(start === end ? `${start}` : `${start}–${end}`)
  return ranges.join(', ')
}
