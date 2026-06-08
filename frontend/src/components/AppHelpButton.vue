<script lang="ts" setup>
import Button from 'primevue/button'
import { OpenDocsPage } from '../../wailsjs/go/main/App'

const props = defineProps<{
  /** Relative path (and optional anchor) on the user-docs site, e.g. "save-game-setup.html#syncing-a-season". */
  docsPath: string
}>()

async function openDocs() {
  await OpenDocsPage(props.docsPath)
}
</script>

<template>
  <Button
    v-tooltip.top="'Learn more in the docs'"
    icon="pi pi-question-circle"
    severity="secondary"
    text
    rounded
    size="small"
    aria-label="Learn more in the docs"
    @click="openDocs"
  />
</template>
