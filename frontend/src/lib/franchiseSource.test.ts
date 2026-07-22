import { describe, expect, it } from 'vitest'
import { FORK_DUPLICATE_SOURCE_REASON, forkDuplicateReason } from './franchiseSource'

const connectedSource = {
  saveFilePath: 'C:\\SMB4\\Saves\\league-original.sav',
  leagueGUID: 'B9B0F849-480C-4B01-B9D8-CB632C739A9B',
}

describe('forkDuplicateReason', () => {
  it('rejects the exact connected path and GUID', () => {
    expect(
      forkDuplicateReason(
        {
          path: connectedSource.saveFilePath,
          leagueGUID: connectedSource.leagueGUID,
        },
        [connectedSource],
      ),
    ).toBe(FORK_DUPLICATE_SOURCE_REASON)
  })

  it('rejects the same GUID from a copied file', () => {
    expect(
      forkDuplicateReason(
        {
          path: 'C:\\SMB4\\Copied Saves\\league-copy.sav',
          leagueGUID: connectedSource.leagueGUID.toLowerCase(),
        },
        [connectedSource],
      ),
    ).toBe(FORK_DUPLICATE_SOURCE_REASON)
  })

  it('rejects the same normalized Windows path with different metadata', () => {
    expect(
      forkDuplicateReason(
        {
          path: 'c:/smb4/saves/unused/../league-original.sav',
          leagueGUID: '2fc4219b-9da9-4e62-b6aa-a962f9d677ee',
        },
        [connectedSource],
      ),
    ).toBe(FORK_DUPLICATE_SOURCE_REASON)
  })

  it('allows a distinct path and league GUID', () => {
    expect(
      forkDuplicateReason(
        {
          path: 'C:\\SMB4\\Saves\\league-fork.sav',
          leagueGUID: '2fc4219b-9da9-4e62-b6aa-a962f9d677ee',
        },
        [connectedSource],
      ),
    ).toBeNull()
  })
})
