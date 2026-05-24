<script lang="ts" setup>
import { computed } from 'vue'
import type { main } from '../../wailsjs/go/models'
import AwardBadge from './AwardBadge.vue'

interface AwardGroup {
  award: main.AwardDTO
  count: number
}

const props = defineProps<{
  player: main.PlayerCareerDTO
  currentSeason?: main.PlayerSeasonLogDTO
  awardsBySeason?: Record<string, main.AwardDTO[]>
}>()

const careerAwardGroups = computed((): AwardGroup[] => {
  if (!props.awardsBySeason) return []
  const allAwards = Object.values(props.awardsBySeason)
    .flat()
    .filter((a) => !a.omitFromGroupings)
  const groups = new Map<string, AwardGroup>()
  for (const award of allAwards) {
    const existing = groups.get(award.name)
    if (existing) {
      existing.count++
    } else {
      groups.set(award.name, { award, count: 1 })
    }
  }
  return [...groups.values()].sort((a, b) => a.award.importance - b.award.importance)
})
</script>

<template>
  <div class="bio-card">
    <div class="name-row">
      <h2 class="player-name">{{ player.firstName }} {{ player.lastName }}</h2>
      <span v-if="player.isHallOfFamer" class="hof-badge">Hall of Famer</span>
    </div>
    <div v-if="careerAwardGroups.length > 0" class="awards-row">
      <AwardBadge
        v-for="group in careerAwardGroups"
        :key="group.award.name"
        :award="group.award"
        :count="group.count"
        size="lg"
      />
    </div>

    <div v-if="currentSeason" class="bio-details">
      <div v-if="currentSeason.primaryPosition" class="bio-item">
        <span class="bio-label">Position</span>
        <span class="bio-val">
          {{ currentSeason.primaryPosition }}
          <span v-if="currentSeason.secondaryPosition" class="secondary-pos">
            / {{ currentSeason.secondaryPosition }}
          </span>
        </span>
      </div>
      <div v-if="currentSeason.pitcherRole" class="bio-item">
        <span class="bio-label">Role</span>
        <span class="bio-val">{{ currentSeason.pitcherRole }}</span>
      </div>
      <div v-if="currentSeason.batHand" class="bio-item">
        <span class="bio-label">Bats / Throws</span>
        <span class="bio-val">{{ currentSeason.batHand }} / {{ currentSeason.throwHand }}</span>
      </div>
      <div v-if="currentSeason.chemistryType" class="bio-item">
        <span class="bio-label">Chemistry</span>
        <span class="bio-val">{{ currentSeason.chemistryType }}</span>
      </div>
      <div v-if="currentSeason.teams.length > 0" class="bio-item">
        <span class="bio-label">Last Team</span>
        <span class="bio-val">{{ currentSeason.teams[0].teamName }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bio-card {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.awards-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}

.player-name {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
}

.hof-badge {
  font-size: 0.875rem;
  font-weight: 600;
  color: #d29922;
  background: color-mix(in srgb, #d29922 15%, transparent);
  border: 1px solid color-mix(in srgb, #d29922 40%, transparent);
  border-radius: 4px;
  padding: 0.3rem 0.65rem;
  white-space: nowrap;
}

.bio-details {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem 2rem;
}

.bio-item {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.bio-label {
  font-size: 0.6875rem;
  font-weight: 500;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.bio-val {
  font-size: 0.9375rem;
  color: var(--color-text-primary);
}

.secondary-pos {
  color: var(--color-text-secondary);
}
</style>
