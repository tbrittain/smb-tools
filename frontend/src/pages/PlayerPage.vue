<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { GetPlayerCareer, GetPlayerCareerAwards, GetPlayerSeasonLog } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AttributesTable from '../components/AttributesTable.vue'
import CareerStatSummary from '../components/CareerStatSummary.vue'
import EmptyState from '../components/EmptyState.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import PlayerBioCard from '../components/PlayerBioCard.vue'
import PlayerStatTable from '../components/PlayerStatTable.vue'
import { useBreadcrumbs } from '../composables/useBreadcrumbs'
import { useStatHighlightsStore } from '../stores/statHighlights'

const props = defineProps<{ playerId: number }>()

const { set } = useBreadcrumbs()
const highlightsStore = useStatHighlightsStore()

const career = ref<main.PlayerCareerDTO | null>(null)
const seasonLog = ref<main.PlayerSeasonLogDTO[]>([])
const awardsBySeason = ref<Record<string, main.AwardDTO[]>>({})
const loading = ref(false)
const error = ref<string | null>(null)

// Show pitching tab if the player has any pitching stats
const hasPitching = computed(() => seasonLog.value.some((r) => r.pitching != null || r.playoffPitching != null))
const hasBatting = computed(() => seasonLog.value.some((r) => r.batting != null || r.playoffBatting != null))

const statMode = ref<'batting' | 'pitching'>('batting')
const showPlayoffs = ref(false)

// Most recent season for bio detail
const mostRecentSeason = computed(() =>
  seasonLog.value.length > 0 ? seasonLog.value[seasonLog.value.length - 1] : undefined,
)

// Most recent attributes
const latestAttrs = computed(() => mostRecentSeason.value)

const isPitcher = computed(() => {
  const s = mostRecentSeason.value
  return s != null && (s.primaryPosition === 'P' || s.pitcherRole !== '')
})

onMounted(async () => {
  loading.value = true
  error.value = null
  highlightsStore.fetch()
  try {
    const [c, log, awards] = await Promise.all([
      GetPlayerCareer(props.playerId),
      GetPlayerSeasonLog(props.playerId),
      GetPlayerCareerAwards(props.playerId),
    ])
    career.value = c
    seasonLog.value = log ?? []
    awardsBySeason.value = awards ?? {}
    statMode.value = isPitcher.value ? 'pitching' : 'batting'
    set([{ label: c ? `${c.firstName} ${c.lastName}` : 'Player' }])
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="player-page">
    <LoadingSpinner v-if="loading" />
    <p v-else-if="error" class="error-text">{{ error }}</p>

    <template v-else-if="career">
      <div class="player-content">
        <!-- Bio header -->
        <PlayerBioCard :player="career" :current-season="mostRecentSeason" :awards-by-season="awardsBySeason" />

        <!-- Career stat summary row -->
        <CareerStatSummary :batting="career.batting" :pitching="career.pitching" />

        <!-- Attributes -->
        <section v-if="latestAttrs" class="section">
          <h3>Attributes
            <span class="season-tag">Season {{ latestAttrs.seasonNum }}</span>
          </h3>
          <AttributesTable
            :power="latestAttrs.power"
            :contact="latestAttrs.contact"
            :speed="latestAttrs.speed"
            :fielding="latestAttrs.fielding"
            :arm="latestAttrs.arm"
            :velocity="latestAttrs.velocity"
            :junk="latestAttrs.junk"
            :accuracy="latestAttrs.accuracy"
            :show-pitching="isPitcher"
          />
        </section>

      </div>

      <!-- Season log — full width -->
      <section class="season-log-section">
        <div class="season-log-header">
          <h3>Season Log</h3>
          <div class="tab-bar">
            <button
              v-if="hasBatting"
              class="tab-btn"
              :class="{ active: statMode === 'batting' }"
              @click="statMode = 'batting'"
            >
              Batting
            </button>
            <button
              v-if="hasPitching"
              class="tab-btn"
              :class="{ active: statMode === 'pitching' }"
              @click="statMode = 'pitching'"
            >
              Pitching
            </button>
            <label class="playoff-toggle">
              <input v-model="showPlayoffs" type="checkbox" />
              Playoffs
            </label>
          </div>
        </div>

        <PlayerStatTable
          :rows="seasonLog"
          :mode="statMode"
          :show-playoffs="showPlayoffs"
          :awards-by-season="awardsBySeason"
          :player-id="props.playerId"
          :highlights="highlightsStore.highlights"
        />
      </section>
    </template>

    <EmptyState v-else message="Player not found" />
  </div>
</template>

<style scoped>
.player-page {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  padding-bottom: 2rem;
}

.player-content {
  padding: 2rem 2rem 0;
  display: flex;
  flex-direction: column;
  gap: 2rem;
  max-width: 1000px;
}

.season-log-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0 2rem;
}

.season-log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.season-tag {
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--color-text-secondary);
}

.section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.tab-bar {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.tab-btn {
  padding: 0.25rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: background 0.1s, color 0.1s;
}

.tab-btn.active,
.tab-btn:hover {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.tab-btn.active {
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.playoff-toggle {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  cursor: pointer;
  margin-left: 0.5rem;
}

.error-text { font-size: 0.875rem; color: var(--color-error); }
</style>
