<script lang="ts" setup>
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { computed, ref } from 'vue'
import { DeleteMediaEverywhere, RemoveMediaAssociation } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'

const props = defineProps<{
  mediaItem: main.MediaItemDTO
  contextEntityType: 'team_season' | 'player'
  contextEntityId: number
  contextEntityLabel: string
}>()

const emit = defineEmits<{
  removed: []
}>()

const visible = defineModel<boolean>('visible', { required: true })

const toast = useToast()
const removing = ref(false)

const hasOtherAssociations = computed(() => props.mediaItem.totalAssociationCount > 1)

async function removeFromContext() {
  removing.value = true
  try {
    await RemoveMediaAssociation(props.mediaItem.id, props.contextEntityType, props.contextEntityId)
    toast.add({
      severity: 'success',
      summary: 'Removed',
      detail: `Removed from ${props.contextEntityLabel}`,
      life: 3000,
    })
    visible.value = false
    emit('removed')
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Error', detail: String(e), life: 5000 })
  } finally {
    removing.value = false
  }
}

async function deleteEverywhere() {
  removing.value = true
  try {
    await DeleteMediaEverywhere(props.mediaItem.id)
    toast.add({ severity: 'success', summary: 'Deleted', detail: `"${props.mediaItem.name}" deleted`, life: 3000 })
    visible.value = false
    emit('removed')
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Error', detail: String(e), life: 5000 })
  } finally {
    removing.value = false
  }
}
</script>

<template>
  <Dialog
    v-model:visible="visible"
    :header="hasOtherAssociations ? 'Remove or delete media?' : 'Delete media?'"
    :modal="true"
    :closable="!removing"
    :style="{ width: '420px' }"
  >
    <div class="confirm-body">
      <p class="item-name">"{{ mediaItem.name }}"</p>

      <template v-if="hasOtherAssociations">
        <p class="confirm-text">
          This item is linked to {{ mediaItem.totalAssociationCount }} places. What would you like to do?
        </p>
        <div class="confirm-actions">
          <Button
            label="Remove from this page only"
            :loading="removing"
            @click="removeFromContext"
          />
          <Button
            label="Delete everywhere"
            severity="danger"
            :loading="removing"
            @click="deleteEverywhere"
          />
        </div>
      </template>

      <template v-else>
        <p class="confirm-text">This will permanently delete the file. This cannot be undone.</p>
        <div class="confirm-actions">
          <Button label="Cancel" severity="secondary" :disabled="removing" @click="visible = false" />
          <Button label="Delete" severity="danger" :loading="removing" @click="deleteEverywhere" />
        </div>
      </template>
    </div>
  </Dialog>
</template>

<style scoped>
.confirm-body {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding-top: 0.25rem;
}

.item-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.confirm-text {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  margin: 0;
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding-top: 0.25rem;
}
</style>
