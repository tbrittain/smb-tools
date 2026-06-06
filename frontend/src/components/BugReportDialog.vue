<script lang="ts" setup>
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import { ref } from 'vue'
import { OpenBugReport } from '../../wailsjs/go/main/App'

const visible = defineModel<boolean>('visible', { required: true })

const includeSystemInfo = ref(false)
const loading = ref(false)

async function handleOpen() {
  loading.value = true
  try {
    await OpenBugReport(includeSystemInfo.value)
  } finally {
    loading.value = false
    visible.value = false
  }
}
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Report a Bug" :style="{ width: '480px' }">
    <div class="bug-report-body">
      <p class="description">
        This will open a pre-filled GitHub issue in your browser. You can review and edit it before
        submitting.
      </p>

      <div class="system-info-opt-in">
        <Checkbox v-model="includeSystemInfo" input-id="include-system-info" binary />
        <label for="include-system-info">Include system info (OS and app version)</label>
      </div>

      <p class="privacy-note">
        Log files may include file paths from your computer (save game location, franchise
        directories). The last portion of the current session log will be included in the issue
        body. Review the issue before submitting if you have privacy concerns.
      </p>
    </div>

    <template #footer>
      <Button label="Cancel" severity="secondary" text @click="visible = false" />
      <Button
        label="Open in GitHub"
        icon="pi pi-external-link"
        :loading="loading"
        @click="handleOpen"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.bug-report-body {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.description {
  color: var(--color-text-primary);
  font-size: 0.9375rem;
  margin: 0;
}

.system-info-opt-in {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.system-info-opt-in label {
  font-size: 0.9375rem;
  color: var(--color-text-primary);
  cursor: pointer;
}

.privacy-note {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  margin: 0;
}
</style>
