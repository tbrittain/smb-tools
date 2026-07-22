import type { main } from '../../wailsjs/go/models'

type SourceIdentity = Pick<main.FranchiseSourceDTO, 'saveFilePath' | 'leagueGUID'>
type CandidateIdentity = Pick<main.SaveFileCandidateDTO, 'path' | 'leagueGUID'>

export const FORK_DUPLICATE_SOURCE_REASON = 'Already connected. Fork requires a different SMB4 league/save.'

export function normalizeSourcePath(path: string): string {
  const isWindowsPath = /^[A-Za-z]:[\\/]/.test(path) || /^[\\/]{2}/.test(path)
  const slashPath = path.replace(/\\/g, '/')
  const hasLeadingSlash = slashPath.startsWith('/')
  const parts: string[] = []

  for (const part of slashPath.split('/')) {
    if (part === '' || part === '.') continue
    if (part === '..') {
      if (parts.length > 0 && parts[parts.length - 1] !== '..') {
        parts.pop()
      } else if (!hasLeadingSlash) {
        parts.push(part)
      }
    } else {
      parts.push(part)
    }
  }

  const normalized = `${hasLeadingSlash ? '/' : ''}${parts.join('/')}`
  return isWindowsPath ? normalized.toLowerCase() : normalized
}

export function forkDuplicateReason(candidate: CandidateIdentity, connectedSources: SourceIdentity[]): string | null {
  const candidatePath = normalizeSourcePath(candidate.path)
  const candidateGUID = candidate.leagueGUID.trim().toLowerCase()

  for (const source of connectedSources) {
    if (normalizeSourcePath(source.saveFilePath) === candidatePath) {
      return FORK_DUPLICATE_SOURCE_REASON
    }
    if (source.leagueGUID.trim().toLowerCase() === candidateGUID) {
      return FORK_DUPLICATE_SOURCE_REASON
    }
  }
  return null
}
