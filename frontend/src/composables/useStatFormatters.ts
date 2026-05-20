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

export function formatSalary(v: number): string {
  return `$${v.toLocaleString()}`
}

export function formatCount(v: number | null | undefined): string {
  if (v == null) return '—'
  return String(Math.round(v))
}
