import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { main } from '../../wailsjs/go/models'
import { useSaveFileCandidates } from './useSaveFileCandidates'

const mockGetSaveFileCandidates = vi.fn()
const mockBrowseSaveDirectory = vi.fn()
const mockBrowseSaveFile = vi.fn()
const mockProbeLeagues = vi.fn()

vi.mock('../../wailsjs/go/main/App', () => ({
  GetSaveFileCandidates: () => mockGetSaveFileCandidates(),
  BrowseSaveDirectory: () => mockBrowseSaveDirectory(),
  BrowseSaveFile: () => mockBrowseSaveFile(),
  ProbeLeagues: (path: string) => mockProbeLeagues(path),
}))

function candidate(overrides: Partial<main.SaveFileCandidateDTO>): main.SaveFileCandidateDTO {
  return {
    path: '/save.sav',
    gameVersion: 'smb4',
    leagueName: 'League',
    numSeasons: 1,
    mode: 'franchise',
    isFranchise: true,
    playerTeamName: 'Team',
    leagueGUID: 'guid',
    ...overrides,
  } as main.SaveFileCandidateDTO
}

describe('useSaveFileCandidates', () => {
  beforeEach(() => {
    mockGetSaveFileCandidates.mockReset()
    mockBrowseSaveDirectory.mockReset()
    mockBrowseSaveFile.mockReset()
    mockProbeLeagues.mockReset()
  })

  describe('load', () => {
    it('includes both franchise and season mode candidates', async () => {
      mockGetSaveFileCandidates.mockResolvedValue([
        candidate({ path: '/a.sav', mode: 'franchise' }),
        candidate({ path: '/b.sav', mode: 'season', isFranchise: false }),
      ])

      const { candidates, load } = useSaveFileCandidates()
      await load()

      expect(candidates.value).toHaveLength(2)
      expect(candidates.value.map((c) => c.mode).sort()).toEqual(['franchise', 'season'])
    })

    it('excludes elimination and none mode candidates', async () => {
      mockGetSaveFileCandidates.mockResolvedValue([
        candidate({ path: '/a.sav', mode: 'franchise' }),
        candidate({ path: '/c.sav', mode: 'elimination' }),
        candidate({ path: '/d.sav', mode: 'none' }),
      ])

      const { candidates, load } = useSaveFileCandidates()
      await load()

      expect(candidates.value).toHaveLength(1)
      expect(candidates.value[0].mode).toBe('franchise')
    })
  })

  describe('scanDirectory', () => {
    it('merges both franchise and season mode candidates found in the directory', async () => {
      mockBrowseSaveDirectory.mockResolvedValue([
        candidate({ path: '/a.sav', mode: 'franchise' }),
        candidate({ path: '/b.sav', mode: 'season', isFranchise: false }),
      ])

      const { candidates, scanDirectory, error } = useSaveFileCandidates()
      await scanDirectory()

      expect(candidates.value).toHaveLength(2)
      expect(error.value).toBeNull()
    })

    it('surfaces an error when only unsupported modes are found', async () => {
      mockBrowseSaveDirectory.mockResolvedValue([candidate({ path: '/e.sav', mode: 'elimination' })])

      const { candidates, scanDirectory, error } = useSaveFileCandidates()
      await scanDirectory()

      expect(candidates.value).toHaveLength(0)
      expect(error.value).toContain('Franchise or Season mode')
    })
  })
})
