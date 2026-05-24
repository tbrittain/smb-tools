<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { SearchPlayers, SearchTeams } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import EmptyState from '../components/EmptyState.vue'
import SearchInput from '../components/SearchInput.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'
import { useSearchDebounce } from '../composables/useSearchDebounce'

const props = defineProps<{ q: string }>()
const router = useRouter()

const playerResults = ref<main.PlayerSearchResultDTO[]>([])
const teamResults = ref<main.TeamSearchResultDTO[]>([])
const error = ref<string | null>(null)
const hasSearched = ref(false)

async function runSearch(query: string) {
  if (!query.trim()) {
    playerResults.value = []
    teamResults.value = []
    hasSearched.value = false
    return
  }
  error.value = null
  hasSearched.value = true
  try {
    const [players, teams] = await Promise.all([SearchPlayers(query), SearchTeams(query)])
    playerResults.value = players ?? []
    teamResults.value = teams ?? []
  } catch (e) {
    error.value = String(e)
  }
  // Update URL query param so the search is bookmarkable/shareable
  router.replace({ path: '/search', query: query ? { q: query } : {} })
}

const { query, loading } = useSearchDebounce(runSearch, 300)

const { set } = useBreadcrumbs()

// Seed from URL query param on mount
onMounted(() => {
  set([{ label: 'Search' }])
  if (props.q) {
    query.value = props.q
  }
})
</script>

<template>
  <div class="search-page">
    <header class="page-header">
      <h2>Search</h2>
    </header>

    <SearchInput
      v-model="query"
      :loading="loading"
      placeholder="Search players and teams…"
      class="search-bar"
    />

    <p v-if="error" class="error-text">{{ error }}</p>

    <template v-if="hasSearched && !loading">
      <!-- Player results -->
      <section class="result-section">
        <h3>
          Players
          <span class="count">{{ playerResults.length }}</span>
        </h3>
        <EmptyState
          v-if="playerResults.length === 0"
          message="No players found"
        />
        <table v-else class="result-table">
          <thead>
            <tr>
              <th>Name</th>
              <th class="col-num">Seasons</th>
              <th class="col-num">First</th>
              <th class="col-num">Last</th>
              <th class="col-num">HoF</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in playerResults" :key="p.playerId">
              <td>
                <RouterLink :to="`/players/${p.playerId}`" class="result-link">
                  {{ p.firstName }} {{ p.lastName }}
                </RouterLink>
              </td>
              <td class="col-num">{{ p.seasonsPlayed }}</td>
              <td class="col-num">{{ p.firstSeason }}</td>
              <td class="col-num">{{ p.lastSeason }}</td>
              <td class="col-num">
                <span v-if="p.isHallOfFamer" class="hof-badge">HoF</span>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      <!-- Team results -->
      <section class="result-section">
        <h3>
          Teams
          <span class="count">{{ teamResults.length }}</span>
        </h3>
        <EmptyState
          v-if="teamResults.length === 0"
          message="No teams found"
        />
        <table v-else class="result-table">
          <thead>
            <tr>
              <th>Name</th>
              <th class="col-num">Seasons</th>
              <th class="col-num">First</th>
              <th class="col-num">Last</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in teamResults" :key="t.teamId">
              <td>
                <RouterLink :to="`/teams/${t.teamId}`" class="result-link">
                  {{ t.teamName }}
                </RouterLink>
              </td>
              <td class="col-num">{{ t.seasons }}</td>
              <td class="col-num">{{ t.firstSeason }}</td>
              <td class="col-num">{{ t.lastSeason }}</td>
            </tr>
          </tbody>
        </table>
      </section>
    </template>

    <EmptyState
      v-else-if="!hasSearched && !loading"
      message="Type to search"
      subtext="Search by player name or team name."
    />
  </div>
</template>

<style scoped>
.search-page {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 800px;
}

.page-header h2 {
  font-size: 1.4rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.search-bar {
  max-width: 480px;
}

.result-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.count {
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--color-text-secondary);
  background: var(--color-surface-2);
  border-radius: 10px;
  padding: 0 6px;
}

.result-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.result-table th {
  text-align: left;
  padding: 0.25rem 0.5rem;
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
  border-bottom: 1px solid var(--color-border);
}

.result-table td {
  padding: 0.4rem 0.5rem;
  border-bottom: 1px solid color-mix(in srgb, var(--color-border) 40%, transparent);
  color: var(--color-text-primary);
}

.col-num { text-align: right; width: 5rem; }

.result-link {
  color: var(--color-text-primary);
  text-decoration: none;
}
.result-link:hover { color: var(--color-accent); }

.hof-badge {
  font-size: 0.625rem;
  font-weight: 600;
  color: #d29922;
  background: color-mix(in srgb, #d29922 15%, transparent);
  border: 1px solid color-mix(in srgb, #d29922 40%, transparent);
  border-radius: 3px;
  padding: 0 4px;
}

.error-text { font-size: 0.875rem; color: var(--color-error); }
</style>
