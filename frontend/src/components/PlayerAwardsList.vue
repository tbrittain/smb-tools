<script lang="ts" setup>
import type { main } from '../../wailsjs/go/models'
import AwardBadge from './AwardBadge.vue'

defineProps<{
  // Keys are season numbers (as strings from the Go map[string][]AwardDTO).
  awardsBySeason: Record<string, main.AwardDTO[]>
}>()

function sortedSeasons(awardsBySeason: Record<string, main.AwardDTO[]>): number[] {
  return Object.keys(awardsBySeason)
    .map(Number)
    .sort((a, b) => a - b)
}

function primaryAwards(awards: main.AwardDTO[]): main.AwardDTO[] {
  return awards.filter((a) => !a.omitFromGroupings)
}

function runnerUpAwards(awards: main.AwardDTO[]): main.AwardDTO[] {
  return awards.filter((a) => a.omitFromGroupings)
}
</script>

<template>
  <div class="awards-list">
    <div v-if="Object.keys(awardsBySeason).length === 0" class="no-awards">
      No awards on record.
    </div>
    <div v-for="season in sortedSeasons(awardsBySeason)" :key="season" class="season-row">
      <span class="season-label">Season {{ season }}</span>
      <div class="badges">
        <AwardBadge
          v-for="award in primaryAwards(awardsBySeason[String(season)])"
          :key="award.id"
          :award="award"
          size="sm"
        />
        <AwardBadge
          v-for="award in runnerUpAwards(awardsBySeason[String(season)])"
          :key="award.id"
          :award="award"
          size="sm"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.awards-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.no-awards {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

.season-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.season-label {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  min-width: 6rem;
  white-space: nowrap;
}

.badges {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}
</style>
