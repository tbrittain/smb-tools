<script lang="ts" setup>
const props = withDefaults(
  defineProps<{
    showLeader?: boolean
    showRecord?: boolean
  }>(),
  {
    showLeader: true,
    showRecord: true,
  },
)
</script>

<template>
  <div class="stat-legend">
    <span v-if="props.showLeader" class="legend-item">
      <span class="swatch swatch-leader">B</span>
      League leader that season
    </span>
    <span v-if="props.showRecord" class="legend-item">
      <span class="swatch swatch-record">R</span>
      All-time record
    </span>
  </div>
</template>

<style scoped>
.stat-legend {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  margin-top: 0.5rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.swatch {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 3px;
  font-size: 0.6875rem;
  font-weight: 700;
}

.swatch-leader {
  font-weight: 700;
  color: #000000;
  background: color-mix(in srgb, #ffffff 85%, transparent);
}

.swatch-record {
  background: color-mix(in srgb, #d4a017 25%, transparent);
  border: 1px solid color-mix(in srgb, #d4a017 50%, transparent);
  color: var(--color-text-primary);
}
</style>
