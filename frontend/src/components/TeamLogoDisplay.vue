<script lang="ts" setup>
defineProps<{
  logoUrl: string
  size: 'sm' | 'lg'
}>()
</script>

<template>
  <div :class="['logo-display', `logo-display--${size}`]">
    <img v-if="logoUrl" :src="logoUrl" alt="Team logo" class="logo-img" />
    <div v-else class="logo-placeholder">
      <span class="placeholder-icon">🏟</span>
    </div>
  </div>
</template>

<style scoped>
.logo-display {
  flex-shrink: 0;
  border-radius: 6px;
  overflow: hidden;
}

.logo-display--sm {
  width: 40px;
  height: 40px;
}

.logo-display--lg {
  width: 80px;
  height: 80px;
}

.logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.logo-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px dashed var(--color-border, #30363d);
  border-radius: 6px;
  background: transparent;
  box-sizing: border-box;
}

.placeholder-icon {
  font-size: 1.25rem;
  opacity: 0.35;
  /* Emoji icon used as a muted visual hint; no semantic meaning. */
  user-select: none;
}

.logo-display--sm .placeholder-icon {
  font-size: 0.875rem;
}
</style>
