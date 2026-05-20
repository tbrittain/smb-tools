<script lang="ts" setup>
const props = defineProps<{
  power: number
  contact: number
  speed: number
  fielding: number
  arm: number
  velocity: number
  junk: number
  accuracy: number
  showPitching: boolean
}>()

interface Attr {
  label: string
  value: number
}

function attrs(): Attr[] {
  const hitting: Attr[] = [
    { label: 'Power', value: props.power },
    { label: 'Contact', value: props.contact },
    { label: 'Speed', value: props.speed },
    { label: 'Fielding', value: props.fielding },
    { label: 'Arm', value: props.arm },
  ]
  const pitching: Attr[] = [
    { label: 'Velocity', value: props.velocity },
    { label: 'Junk', value: props.junk },
    { label: 'Accuracy', value: props.accuracy },
  ]
  return props.showPitching ? [...hitting, ...pitching] : hitting
}

function barColor(v: number): string {
  if (v >= 80) return '#3fb950'
  if (v >= 60) return '#d29922'
  return 'var(--color-accent)'
}
</script>

<template>
  <div class="attrs-table">
    <div v-for="a in attrs()" :key="a.label" class="attr-row">
      <span class="attr-label">{{ a.label }}</span>
      <div class="bar-wrap">
        <div
          class="bar"
          :style="{ width: `${(a.value / 99) * 100}%`, background: barColor(a.value) }"
        />
      </div>
      <span class="attr-val">{{ a.value }}</span>
    </div>
  </div>
</template>

<style scoped>
.attrs-table {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  max-width: 340px;
}

.attr-row {
  display: grid;
  grid-template-columns: 72px 1fr 36px;
  align-items: center;
  gap: 0.5rem;
}

.attr-label {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  text-align: right;
}

.bar-wrap {
  height: 8px;
  background: var(--color-surface-3);
  border-radius: 4px;
  overflow: hidden;
}

.bar {
  height: 100%;
  border-radius: 4px;
  transition: width 0.2s ease;
}

.attr-val {
  font-size: 0.8125rem;
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--color-text-primary);
  text-align: right;
}
</style>
