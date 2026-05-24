import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useBreadcrumbs } from './useBreadcrumbs'

function mockHistory(position: number, current: string) {
  vi.stubGlobal('history', { state: { position, current } })
}

describe('useBreadcrumbs', () => {
  beforeEach(() => {
    useBreadcrumbs().clear()
    vi.unstubAllGlobals()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  describe('set() / crumbs', () => {
    it('produces a single crumb for the first entry', () => {
      mockHistory(1, '/teams/1')
      const { set, crumbs } = useBreadcrumbs()
      set([{ label: 'United' }])
      expect(crumbs.value.map((c) => c.label)).toEqual(['United'])
    })

    it('flattens multiple trail entries into a single crumb list', () => {
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United' }])

      mockHistory(2, '/teams/1/seasons/3')
      set([{ label: 'United Season 3' }])

      mockHistory(3, '/players/7')
      set([{ label: 'Julian Huber' }])

      expect(crumbs.value.map((c) => c.label)).toEqual(['United', 'United Season 3', 'Julian Huber'])
    })

    it('attaches historyPosition to all non-terminal crumbs', () => {
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United' }])

      mockHistory(2, '/teams/1/seasons/3')
      set([{ label: 'United Season 3' }])

      mockHistory(3, '/players/7')
      set([{ label: 'Julian Huber' }])

      expect(crumbs.value[0].historyPosition).toBe(1)
      expect(crumbs.value[1].historyPosition).toBe(2)
      expect(crumbs.value[2].historyPosition).toBeUndefined()
    })

    it('injects entry path as `to` on non-terminal last items that lack an explicit to', () => {
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United' }])

      mockHistory(2, '/players/7')
      set([{ label: 'Julian Huber' }])

      expect(crumbs.value[0].to).toBe('/teams/1')
    })

    it('preserves an explicit `to` on non-terminal items', () => {
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United', to: '/teams/1?custom=1' }])

      mockHistory(2, '/players/7')
      set([{ label: 'Julian Huber' }])

      expect(crumbs.value[0].to).toBe('/teams/1?custom=1')
    })

    it('does not attach historyPosition to the terminal crumb', () => {
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/players/7')
      set([{ label: 'Julian Huber' }])

      expect(crumbs.value[0].historyPosition).toBeUndefined()
    })
  })

  describe('root path reset', () => {
    it('resets the trail when set() is called on a root path', () => {
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United' }])

      mockHistory(2, '/players/7')
      set([{ label: 'Julian Huber' }])

      // Navigate to a root path — trail should reset to just this entry
      mockHistory(3, '/teams')
      set([{ label: 'Teams' }])

      expect(crumbs.value.map((c) => c.label)).toEqual(['Teams'])
    })

    it('treats the dashboard path / as a root path', () => {
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United' }])

      mockHistory(2, '/')
      set([{ label: 'Dashboard' }])

      expect(crumbs.value.map((c) => c.label)).toEqual(['Dashboard'])
    })
  })

  describe('back-navigation trim', () => {
    it('trims entries at or beyond the target position when navigating back', () => {
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United' }])

      mockHistory(2, '/teams/1/seasons/3')
      set([{ label: 'United Season 3' }])

      mockHistory(3, '/players/7')
      set([{ label: 'Julian Huber' }])

      // Simulate browser back — history position drops to 2
      mockHistory(2, '/teams/1/seasons/3')
      set([{ label: 'United Season 3' }])

      expect(crumbs.value.map((c) => c.label)).toEqual(['United', 'United Season 3'])
    })

    it('clicking a breadcrumb link then arriving at that page does not grow the trail', () => {
      // This is the regression case: clicking "United Season 3" while on Julian Huber
      // should not append a new "United Season 3" entry — it should trim back to it.
      const { set, crumbs } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United' }])

      mockHistory(2, '/teams/1/seasons/3')
      set([{ label: 'United Season 3' }])

      mockHistory(3, '/players/7')
      set([{ label: 'Julian Huber' }])

      // User clicks "United Season 3" breadcrumb → router.go(-1) → history pos 2
      mockHistory(2, '/teams/1/seasons/3')
      set([{ label: 'United Season 3' }])

      // Trail must NOT be: United, United Season 3, Julian Huber, United Season 3
      expect(crumbs.value.map((c) => c.label)).toEqual(['United', 'United Season 3'])
      expect(crumbs.value[crumbs.value.length - 1].historyPosition).toBeUndefined()
    })
  })

  describe('clear()', () => {
    it('empties the trail', () => {
      const { set, crumbs, clear } = useBreadcrumbs()

      mockHistory(1, '/teams/1')
      set([{ label: 'United' }])

      clear()
      expect(crumbs.value).toHaveLength(0)
    })
  })
})
