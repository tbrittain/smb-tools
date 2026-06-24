<script lang="ts" setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { SearchPlayers, SearchTeams } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import { useSearchDebounce } from '../composables/useSearchDebounce'
import HofBadge from './HofBadge.vue'

const router = useRouter()
const containerRef = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const hasSearched = ref(false)
const playerResults = ref<main.PlayerSearchResultDTO[]>([])
const teamResults = ref<main.TeamSearchResultDTO[]>([])

async function runSearch(q: string) {
  hasSearched.value = true
  try {
    const [players, teams] = await Promise.all([SearchPlayers(q), SearchTeams(q)])
    playerResults.value = (players ?? []).slice(0, 5)
    teamResults.value = (teams ?? []).slice(0, 5)
    isOpen.value = true
  } catch {
    isOpen.value = false
  }
}

const { query, loading } = useSearchDebounce(runSearch, 300)

watch(query, (q) => {
  if (!q.trim()) {
    playerResults.value = []
    teamResults.value = []
    hasSearched.value = false
    isOpen.value = false
  }
})

function close() {
  isOpen.value = false
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    query.value = ''
    hasSearched.value = false
    close()
  }
}

function handleFocus() {
  if (hasSearched.value && query.value.trim()) {
    isOpen.value = true
  }
}

function handleClickOutside(e: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    close()
  }
}

function navigateTo(path: string) {
  router.push(path)
  query.value = ''
  hasSearched.value = false
  close()
}

onMounted(() => document.addEventListener('mousedown', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('mousedown', handleClickOutside))
</script>

<template>
  <div ref="containerRef" class="global-search">
    <div class="search-wrap">
      <span class="search-icon" aria-hidden="true">
        <svg width="13" height="13" viewBox="0 0 16 16" fill="none">
          <circle cx="6.5" cy="6.5" r="5" stroke="currentColor" stroke-width="1.5" />
          <line x1="10.5" y1="10.5" x2="14" y2="14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
      </span>
      <input
        v-model="query"
        type="search"
        class="search-input"
        placeholder="Search players and teams…"
        autocomplete="off"
        @keydown="handleKeydown"
        @focus="handleFocus"
      />
      <span v-if="loading" class="search-spinner" aria-label="Searching" />
    </div>

    <div v-if="isOpen && hasSearched" class="search-dropdown" aria-label="Search results">
      <section class="dropdown-section">
        <div class="section-label">Players</div>
        <p v-if="playerResults.length === 0" class="no-results">No players found</p>
        <button
          v-for="p in playerResults"
          :key="p.playerId"
          class="result-item"
          @click="navigateTo(`/players/${p.playerId}`)"
        >
          <span class="result-name">{{ p.firstName }} {{ p.lastName }}</span>
          <HofBadge v-if="p.isHallOfFamer" />
          <span class="result-meta">{{ p.seasonsPlayed }} {{ p.seasonsPlayed === 1 ? 'season' : 'seasons' }}</span>
        </button>
      </section>
      <div class="section-divider" aria-hidden="true" />
      <section class="dropdown-section">
        <div class="section-label">Teams</div>
        <p v-if="teamResults.length === 0" class="no-results">No teams found</p>
        <button
          v-for="t in teamResults"
          :key="t.teamId"
          class="result-item"
          @click="navigateTo(`/teams/${t.teamId}`)"
        >
          <span class="result-name">{{ t.teamName }}</span>
          <span class="result-meta">{{ t.seasons }} {{ t.seasons === 1 ? 'season' : 'seasons' }}</span>
        </button>
      </section>
    </div>
  </div>
</template>

<style scoped>
.global-search {
  position: relative;
}

.search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.5rem;
  color: var(--color-text-secondary);
  display: flex;
  pointer-events: none;
}

.search-input {
  width: 460px;
  padding: 0.3125rem 1.75rem 0.3125rem 1.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.8125rem;
  font-family: var(--font-sans);
  outline: none;
  appearance: none;
}

.search-input:focus {
  border-color: var(--color-accent);
}

.search-input::placeholder {
  color: var(--color-text-secondary);
}

.search-input::-webkit-search-cancel-button {
  display: none;
}

.search-spinner {
  position: absolute;
  right: 0.5rem;
  width: 12px;
  height: 12px;
  border: 2px solid var(--color-surface-3);
  border-top-color: var(--color-accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.search-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  width: 100%;
  background: var(--color-surface-1);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
  z-index: 100;
  overflow: hidden;
}

.dropdown-section {
  padding: 0.5rem 0;
}

.section-label {
  padding: 0.25rem 0.75rem;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.no-results {
  padding: 0.25rem 0.75rem 0.375rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  margin: 0;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.375rem 0.75rem;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  font-family: var(--font-sans);
  font-size: 0.875rem;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
}

.result-item:hover,
.result-item:focus {
  background: var(--color-surface-2);
  outline: none;
}

.result-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}

.result-meta {
  flex-shrink: 0;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.section-divider {
  height: 1px;
  background: var(--color-border);
}
</style>
