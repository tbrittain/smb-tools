import { ref } from 'vue'

export interface PageDataState<T> {
  data: ReturnType<typeof ref<T | null>>
  loading: ReturnType<typeof ref<boolean>>
  error: ReturnType<typeof ref<string | null>>
  load: () => Promise<void>
}

/**
 * Generic composable for async page-level data fetching.
 * Pages call `load()` in onMounted; components receive data via props.
 */
export function usePageData<T>(fetcher: () => Promise<T>): PageDataState<T> {
  const data = ref<T | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      data.value = (await fetcher()) as T
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  return { data, loading, error, load }
}
