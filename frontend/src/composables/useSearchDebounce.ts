import { ref, watch } from 'vue'

/**
 * Wraps a search query string with debounced execution. The handler is called
 * after the user stops typing for `delayMs` milliseconds. The `loading` ref is
 * set to true immediately when the query changes and back to false once the
 * handler resolves.
 */
export function useSearchDebounce(
  handler: (query: string) => Promise<void>,
  delayMs = 300,
): { query: ReturnType<typeof ref<string>>; loading: ReturnType<typeof ref<boolean>> } {
  const query = ref('')
  const loading = ref(false)

  let timer: ReturnType<typeof setTimeout> | null = null

  watch(query, (q) => {
    if (timer) clearTimeout(timer)
    if (!q.trim()) {
      loading.value = false
      return
    }
    loading.value = true
    timer = setTimeout(async () => {
      try {
        await handler(q)
      } finally {
        loading.value = false
      }
    }, delayMs)
  })

  return { query, loading }
}
