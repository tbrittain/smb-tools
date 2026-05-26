<script lang="ts" setup>
import type { RouteLocationRaw } from 'vue-router'
import { RouterLink } from 'vue-router'

defineProps<{
  /** Internal route — renders a RouterLink. Mutually exclusive with href. */
  to?: RouteLocationRaw
  /** External URL — renders an <a> with target="_blank". Mutually exclusive with to. */
  href?: string
}>()
</script>

<template>
  <a
    v-if="href"
    :href="href"
    target="_blank"
    rel="noopener noreferrer"
    class="app-link"
  ><slot /></a>
  <RouterLink
    v-else-if="to"
    :to="to"
    class="app-link"
  ><slot /></RouterLink>
  <span v-else class="app-link"><slot /></span>
</template>

<style scoped>
.app-link {
  color: var(--color-accent);
  text-decoration: none;
}
.app-link:hover {
  text-decoration: underline;
}
</style>
