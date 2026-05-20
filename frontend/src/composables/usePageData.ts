import { ref, type Ref } from 'vue'

/**
 * Generic composable for async page-level data fetching.
 * Pages call `load()` in onMounted; components receive data via props.
 */
export function usePageData<T>(fetcher: () => Promise<T>): {
  data: Ref<T | null>
  loading: Ref<boolean>
  error: Ref<string | null>
  load: () => Promise<void>
} {
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

  return { data: data as Ref<T | null>, loading, error, load }
}
