<script lang="ts" setup>
defineProps<{
  value?: number | string | null
  classMap?: Record<string, boolean>
  tooltip?: string
}>()
</script>

<template>
  <span v-if="value != null" :class="classMap" v-tooltip.top="tooltip">{{ value }}</span>
  <span v-else>—</span>
</template>

<style scoped>
.stat-leader {
  font-weight: 700;
  color: #000000;
  background: color-mix(in srgb, #ffffff 85%, transparent);
  border-radius: 3px;
  padding: 1px 3px;
}

.stat-record {
  background: color-mix(in srgb, #d4a017 25%, transparent);
  border-radius: 3px;
  padding: 1px 3px;
}
</style>
