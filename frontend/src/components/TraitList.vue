<script lang="ts" setup>
// Negative trait names — everything else is positive. Includes legacy SMB3 names
// so migrated franchise data renders correctly.
const NEGATIVE_TRAITS = new Set([
  'Bad Jumps',
  'Base Jogger',
  'BB Prone',
  'Butter Fingers',
  'Choker',
  'Crossed Up',
  'Easy Jumps',
  'Easy Target',
  'Falls Behind',
  'First Pitch Prayer',
  'Injury Prone',
  'K Dud',
  'K Neglector',
  'Meltdown',
  'Noodle Arm',
  'RBI Dud',
  'RBI Zero',
  'Slow Poke',
  'Surrounded',
  'Whiffer',
  'Wild Thing',
  'Wild Thrower',
])

defineProps<{
  traits: string[]
}>()

function traitClass(trait: string): 'trait-pos' | 'trait-neg' {
  return NEGATIVE_TRAITS.has(trait) ? 'trait-neg' : 'trait-pos'
}
</script>

<template>
  <span v-if="traits.length > 0" class="trait-list">
    <template v-for="(trait, i) in traits" :key="trait">
      <span v-if="i" class="trait-sep">, </span>
      <span :class="traitClass(trait)">{{ trait }}</span>
    </template>
  </span>
  <span v-else class="trait-empty">—</span>
</template>

<style scoped>
.trait-list {
  font-size: 0.8125rem;
}

.trait-pos {
  color: #4a9eff;
}

.trait-neg {
  color: var(--color-error, #e05252);
}

.trait-sep {
  color: var(--color-text-secondary);
}

.trait-empty {
  color: var(--color-text-secondary);
}
</style>
