import { describe, expect, it, vi } from 'vitest'
import { usePageData } from './usePageData'

describe('usePageData', () => {
  it('starts with null data, not loading, no error', () => {
    const { data, loading, error } = usePageData(() => Promise.resolve('x'))
    expect(data.value).toBeNull()
    expect(loading.value).toBe(false)
    expect(error.value).toBeNull()
  })

  it('sets loading true during fetch, populates data on success', async () => {
    let resolve: (v: string) => void = () => {}
    const { data, loading, load } = usePageData(
      () =>
        new Promise<string>((r) => {
          resolve = r
        }),
    )

    const promise = load()
    expect(loading.value).toBe(true)
    resolve('hello')
    await promise
    expect(loading.value).toBe(false)
    expect(data.value).toBe('hello')
  })

  it('sets error and clears loading on failure', async () => {
    const { error, loading, load } = usePageData(() => Promise.reject(new Error('fetch failed')))

    await load()
    expect(loading.value).toBe(false)
    expect(error.value).toBe('Error: fetch failed')
  })

  it('clears previous error on re-fetch', async () => {
    let shouldFail = true
    const { error, load } = usePageData(() => (shouldFail ? Promise.reject(new Error('oops')) : Promise.resolve('ok')))

    await load()
    expect(error.value).not.toBeNull()

    shouldFail = false
    await load()
    expect(error.value).toBeNull()
  })

  it('calls the fetcher each time load() is called', async () => {
    const fetcher = vi.fn().mockResolvedValue(42)
    const { load } = usePageData(fetcher)

    await load()
    await load()
    expect(fetcher).toHaveBeenCalledTimes(2)
  })
})
