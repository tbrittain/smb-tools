import { defineStore } from 'pinia'
import { ref } from 'vue'
import { GetStatHighlights } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'

export const useStatHighlightsStore = defineStore('statHighlights', () => {
  const highlights = ref<main.StatHighlightsDTO | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetch() {
    if (highlights.value) return
    loading.value = true
    error.value = null
    try {
      highlights.value = await GetStatHighlights()
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  function invalidate() {
    highlights.value = null
  }

  return { highlights, loading, error, fetch, invalidate }
})
