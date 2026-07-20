import { ref } from 'vue'
import { BrowseSaveDirectory, BrowseSaveFile, GetSaveFileCandidates, ProbeLeagues } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'

export function useSaveFileCandidates() {
  const candidates = ref<main.SaveFileCandidateDTO[]>([])
  const loading = ref(false)
  const scanning = ref(false)
  const browsing = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      const all = await GetSaveFileCandidates()
      candidates.value = (all ?? [])
        .filter((c) => c.gameVersion === 'smb4' && (c.mode === 'franchise' || c.mode === 'season'))
        .sort((a, b) => b.numSeasons - a.numSeasons)
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function scanDirectory() {
    scanning.value = true
    error.value = null
    try {
      const found = await BrowseSaveDirectory()
      const supported = found.filter((c) => c.gameVersion === 'smb4' && (c.mode === 'franchise' || c.mode === 'season'))
      const existing = new Set(candidates.value.map((c) => c.path))
      const merged = [...candidates.value, ...supported.filter((c) => !existing.has(c.path))]
      candidates.value = merged.sort((a, b) => b.numSeasons - a.numSeasons)
      if (found.length === 0) {
        error.value =
          'No save files found. Make sure you are pointing at the folder that directly contains your league-*.sav files, or its parent.'
      } else if (supported.length === 0) {
        error.value = 'No Franchise or Season mode saves found in that folder. Elimination saves are not supported.'
      }
    } catch (e) {
      const msg = String(e)
      if (msg) error.value = msg
    } finally {
      scanning.value = false
    }
  }

  async function browseFile(): Promise<main.SaveFileCandidateDTO | null> {
    browsing.value = true
    error.value = null
    try {
      const path = await BrowseSaveFile()
      if (!path) return null
      const probed = await ProbeLeagues(path)
      const match = probed[0]
      const candidate: main.SaveFileCandidateDTO = match
        ? { ...match, path, gameVersion: 'smb4' }
        : {
            path,
            gameVersion: 'smb4',
            leagueName: '',
            numSeasons: 0,
            mode: 'unknown',
            isFranchise: false,
            playerTeamName: '',
            leagueGUID: '',
          }
      if (!candidates.value.find((c) => c.path === path)) {
        candidates.value = [...candidates.value, candidate]
      }
      return candidate
    } catch (e) {
      error.value = String(e)
      return null
    } finally {
      browsing.value = false
    }
  }

  return { candidates, loading, scanning, browsing, error, load, scanDirectory, browseFile }
}
