<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import type { main } from '../../wailsjs/go/models'
import RemoveMediaConfirmDialog from './RemoveMediaConfirmDialog.vue'

const props = defineProps<{
  items: main.MediaItemDTO[]
  initialIndex: number
  entityType: 'team_season' | 'player'
  entityId: number
  entityLabel: string
}>()

const emit = defineEmits<{
  close: []
  removed: [mediaId: string]
}>()

const currentIndex = ref(props.initialIndex)
const confirmDialogVisible = ref(false)

const currentItem = computed(() => props.items[currentIndex.value])

function prev() {
  if (currentIndex.value > 0) currentIndex.value--
}

function next() {
  if (currentIndex.value < props.items.length - 1) currentIndex.value++
}

function handleKeydown(e: KeyboardEvent) {
  switch (e.key) {
    case 'Escape':
      emit('close')
      break
    case 'ArrowLeft':
      prev()
      break
    case 'ArrowRight':
      next()
      break
  }
}

function onMediaRemoved() {
  confirmDialogVisible.value = false
  if (currentItem.value) {
    emit('removed', currentItem.value.id)
  }
  emit('close')
}

onMounted(() => window.addEventListener('keydown', handleKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', handleKeydown))
</script>

<template>
  <Teleport to="body">
    <div class="lightbox-overlay" role="dialog" aria-modal="true" @click.self="emit('close')">
      <div class="lightbox-content">
        <!-- Close -->
        <button class="lightbox-close" aria-label="Close" @click="emit('close')">✕</button>

        <!-- Media -->
        <div class="media-frame">
          <img
            v-if="currentItem?.mediaType === 'image'"
            :src="currentItem.url"
            :alt="currentItem.name"
            class="media-image"
          />
          <video
            v-else-if="currentItem?.mediaType === 'video'"
            :src="currentItem?.url"
            class="media-video"
            controls
            :key="currentItem?.id"
          />
        </div>

        <!-- Navigation arrows -->
        <button
          v-if="currentIndex > 0"
          class="nav-btn nav-btn--prev"
          aria-label="Previous"
          @click="prev"
        >
          ‹
        </button>
        <button
          v-if="currentIndex < items.length - 1"
          class="nav-btn nav-btn--next"
          aria-label="Next"
          @click="next"
        >
          ›
        </button>

        <!-- Info bar -->
        <div class="info-bar">
          <div class="info-text">
            <span class="media-name">{{ currentItem?.name }}</span>
            <span v-if="currentItem?.description" class="media-desc">{{ currentItem.description }}</span>
          </div>
          <div class="info-actions">
            <span class="counter">{{ currentIndex + 1 }} / {{ items.length }}</span>
            <button class="delete-btn" aria-label="Delete media" @click="confirmDialogVisible = true">
              Delete
            </button>
          </div>
        </div>
      </div>

      <RemoveMediaConfirmDialog
        v-if="currentItem"
        v-model:visible="confirmDialogVisible"
        :media-item="currentItem"
        :context-entity-type="entityType"
        :context-entity-id="entityId"
        :context-entity-label="entityLabel"
        @removed="onMediaRemoved"
      />
    </div>
  </Teleport>
</template>

<style scoped>
.lightbox-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.lightbox-content {
  position: relative;
  display: flex;
  flex-direction: column;
  max-width: 90vw;
  max-height: 90vh;
  gap: 0.75rem;
}

.lightbox-close {
  position: absolute;
  top: -2.25rem;
  right: 0;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.75);
  font-size: 1.25rem;
  cursor: pointer;
  padding: 0.25rem;
  line-height: 1;
}

.lightbox-close:hover {
  color: #fff;
}

.media-frame {
  display: flex;
  align-items: center;
  justify-content: center;
  max-height: 75vh;
}

.media-image {
  max-width: 90vw;
  max-height: 75vh;
  object-fit: contain;
  border-radius: 4px;
}

.media-video {
  max-width: 90vw;
  max-height: 75vh;
  outline: none;
  border-radius: 4px;
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(0, 0, 0, 0.5);
  border: none;
  color: #fff;
  font-size: 2rem;
  line-height: 1;
  cursor: pointer;
  padding: 0.5rem 0.75rem;
  border-radius: 4px;
}

.nav-btn--prev {
  left: -3.5rem;
}

.nav-btn--next {
  right: -3.5rem;
}

.nav-btn:hover {
  background: rgba(0, 0, 0, 0.8);
}

.info-bar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0 0.25rem;
}

.info-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.media-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: #fff;
}

.media-desc {
  font-size: 0.8125rem;
  color: rgba(255, 255, 255, 0.65);
}

.info-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
}

.counter {
  font-size: 0.8125rem;
  color: rgba(255, 255, 255, 0.55);
}

.delete-btn {
  background: none;
  border: 1px solid rgba(255, 255, 255, 0.25);
  color: rgba(255, 255, 255, 0.75);
  font-size: 0.8125rem;
  font-family: var(--font-sans);
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.delete-btn:hover {
  border-color: #e87040;
  color: #e87040;
}
</style>
