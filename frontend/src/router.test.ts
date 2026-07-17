import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import type { main } from '../wailsjs/go/models'
import router from './router'
import { useFranchiseStore } from './stores/franchise'

function franchiseWithMode(leagueMode: string): main.FranchiseDTO {
  return { id: 'f1', name: 'Test', leagueMode } as main.FranchiseDTO
}

describe('router: /hall-of-fame guard', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('redirects to / for a season-mode active franchise', async () => {
    const store = useFranchiseStore()
    store.active = franchiseWithMode('season')

    await router.push('/hall-of-fame')
    expect(router.currentRoute.value.path).toBe('/')
  })

  it('allows navigation for a franchise-mode active franchise', async () => {
    const store = useFranchiseStore()
    store.active = franchiseWithMode('franchise')

    await router.push('/hall-of-fame')
    expect(router.currentRoute.value.path).toBe('/hall-of-fame')
  })

  it('allows navigation when no franchise is active', async () => {
    const store = useFranchiseStore()
    store.active = null

    await router.push('/hall-of-fame')
    expect(router.currentRoute.value.path).toBe('/hall-of-fame')
  })
})
