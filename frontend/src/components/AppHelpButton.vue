<script lang="ts" setup>
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import IconButton from './IconButton.vue'

const DOCS_BASE_URL = 'https://tbrittain.github.io/smb-tools/'

const props = defineProps<{
  /** Relative path (and optional anchor) on the user-docs site, e.g. "save-game-setup.html#syncing-a-season". */
  docsPath: string
}>()

function openDocs() {
  BrowserOpenURL(DOCS_BASE_URL + props.docsPath)
}
</script>

<template>
  <IconButton
    v-tooltip.top="'Learn more in the docs'"
    icon="pi pi-question-circle"
    rounded
    aria-label="Learn more in the docs"
    @click="openDocs"
  />
</template>
