import { computed, ref } from 'vue'

export interface BreadcrumbItem {
  label: string
  to?: string
}

interface TrailEntry {
  historyPosition: number
  path: string
  items: BreadcrumbItem[]
}

// Paths that represent top-level navigation (sidebar links). Arriving at one
// resets the trail rather than pushing onto it.
const ROOT_PATHS = new Set([
  '/',
  '/teams',
  '/leaderboards',
  '/awards',
  '/hall-of-fame',
  '/search',
  '/setup',
  '/migrate-legacy',
])

const trail = ref<TrailEntry[]>([])

export function useBreadcrumbs() {
  function set(items: BreadcrumbItem[]) {
    const pos: number = window.history.state?.position ?? 0
    const path: string = window.history.state?.current ?? window.location.hash.slice(1)

    if (ROOT_PATHS.has(path)) {
      trail.value = [{ historyPosition: pos, path, items }]
      return
    }

    // Filter out any entries at or beyond this position (covers back-navigation re-entry)
    const ancestors = trail.value.filter((e) => e.historyPosition < pos)
    trail.value = [...ancestors, { historyPosition: pos, path, items }]
  }

  // Build the flat crumb list from the accumulated trail. Non-terminal entries
  // whose last item lacks a `to` get the entry's path injected so they're linkable.
  const crumbs = computed<BreadcrumbItem[]>(() =>
    trail.value.flatMap((entry, entryIdx) => {
      const isLast = entryIdx === trail.value.length - 1
      return entry.items.map((item, itemIdx) => {
        const isLastItem = itemIdx === entry.items.length - 1
        if (!isLast && isLastItem && !item.to) {
          return { ...item, to: entry.path }
        }
        return item
      })
    }),
  )

  function clear() {
    trail.value = []
  }

  return { crumbs, set, clear }
}
