import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { useSearchDebounce } from './useSearchDebounce'

describe('useSearchDebounce', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('does not call handler immediately on query change', async () => {
    const handler = vi.fn().mockResolvedValue(undefined)
    const { query } = useSearchDebounce(handler, 300)

    query.value = 'Smith'
    await nextTick()
    expect(handler).not.toHaveBeenCalled()
  })

  it('calls handler after the debounce delay', async () => {
    const handler = vi.fn().mockResolvedValue(undefined)
    const { query } = useSearchDebounce(handler, 300)

    query.value = 'Smith'
    await nextTick()
    vi.advanceTimersByTime(300)
    await nextTick()
    expect(handler).toHaveBeenCalledWith('Smith')
  })

  it('cancels pending call when query changes before delay elapses', async () => {
    const handler = vi.fn().mockResolvedValue(undefined)
    const { query } = useSearchDebounce(handler, 300)

    query.value = 'Sm'
    await nextTick()
    vi.advanceTimersByTime(100)

    query.value = 'Smith'
    await nextTick()
    vi.advanceTimersByTime(300)
    await nextTick()

    expect(handler).toHaveBeenCalledTimes(1)
    expect(handler).toHaveBeenCalledWith('Smith')
  })

  it('sets loading true while waiting and false after handler resolves', async () => {
    let resolve: () => void = () => {}
    const handler = vi.fn().mockImplementation(
      () =>
        new Promise<void>((r) => {
          resolve = r
        }),
    )
    const { query, loading } = useSearchDebounce(handler, 300)

    query.value = 'Smith'
    await nextTick()
    expect(loading.value).toBe(true)

    vi.advanceTimersByTime(300)
    await nextTick()
    resolve()
    await nextTick()
    expect(loading.value).toBe(false)
  })

  it('does not call handler for empty query', async () => {
    const handler = vi.fn().mockResolvedValue(undefined)
    const { query, loading } = useSearchDebounce(handler, 300)

    query.value = '  '
    await nextTick()
    vi.advanceTimersByTime(300)
    await nextTick()

    expect(handler).not.toHaveBeenCalled()
    expect(loading.value).toBe(false)
  })
})
