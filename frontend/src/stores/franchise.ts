import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { main } from '../../wailsjs/go/models'
import {
  CreateFranchise,
  DeleteFranchise,
  GetActiveFranchise,
  ListFranchises,
  RenameFranchise,
  SelectFranchise,
} from '../../wailsjs/go/main/App'

export const useFranchiseStore = defineStore('franchise', () => {
  const franchises = ref<main.FranchiseDTO[]>([])
  const active = ref<main.FranchiseDTO | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function loadFranchises() {
    loading.value = true
    error.value = null
    try {
      franchises.value = await ListFranchises()
      const currentActive = await GetActiveFranchise()
      active.value = currentActive.id ? currentActive : null
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function createFranchise(name: string, gameVersion: string) {
    const created = await CreateFranchise(name, gameVersion)
    franchises.value = [...franchises.value, created]
    return created
  }

  async function selectFranchise(id: string) {
    const selected = await SelectFranchise(id)
    active.value = selected
    return selected
  }

  async function renameFranchise(id: string, newName: string) {
    await RenameFranchise(id, newName)
    franchises.value = franchises.value.map(f =>
      f.id === id ? { ...f, name: newName } : f
    )
    if (active.value?.id === id) {
      active.value = { ...active.value, name: newName }
    }
  }

  async function deleteFranchise(id: string) {
    await DeleteFranchise(id)
    franchises.value = franchises.value.filter(f => f.id !== id)
    if (active.value?.id === id) {
      active.value = null
    }
  }

  return {
    franchises,
    active,
    loading,
    error,
    loadFranchises,
    createFranchise,
    selectFranchise,
    renameFranchise,
    deleteFranchise,
  }
})
