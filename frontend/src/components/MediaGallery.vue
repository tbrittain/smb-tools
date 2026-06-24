<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import { GetMediaForPlayer, GetMediaForTeamSeason } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import AppButton from './AppButton.vue'
import MediaLightbox from './MediaLightbox.vue'
import MediaUploadDialog from './MediaUploadDialog.vue'

const props = defineProps<{
  entityType: 'team_season' | 'player'
  entityId: number
  entityLabel: string
  pageSize?: number
}>()

const effectivePageSize = computed(() => props.pageSize ?? 24)

const items = ref<main.MediaItemDTO[]>([])
const totalCount = ref(0)
const currentPage = ref(0)
const loading = ref(false)
const uploadDialogVisible = ref(false)
const lightboxVisible = ref(false)
const lightboxIndex = ref(0)

const hasMore = computed(() => items.value.length < totalCount.value)

async function loadPage(page: number) {
  loading.value = true
  try {
    const result =
      props.entityType === 'team_season'
        ? await GetMediaForTeamSeason(props.entityId, page, effectivePageSize.value)
        : await GetMediaForPlayer(props.entityId, page, effectivePageSize.value)

    if (page === 0) {
      items.value = result.items ?? []
    } else {
      items.value = [...items.value, ...(result.items ?? [])]
    }
    totalCount.value = result.totalCount
    currentPage.value = page
  } catch (e) {
    console.error('MediaGallery: failed to load media', e)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  await loadPage(currentPage.value + 1)
}

function openLightbox(index: number) {
  lightboxIndex.value = index
  lightboxVisible.value = true
}

function onUploaded() {
  uploadDialogVisible.value = false
  loadPage(0)
}

function onMediaRemoved(removedId: string) {
  lightboxVisible.value = false
  items.value = items.value.filter((item) => item.id !== removedId)
  totalCount.value = Math.max(0, totalCount.value - 1)
}

watch([() => props.entityType, () => props.entityId], () => {
  loadPage(0)
})

onMounted(() => loadPage(0))
</script>

<template>
  <section class="media-gallery">
    <div class="gallery-header">
      <h3 class="gallery-title">Media</h3>
      <AppButton size="sm" @click="uploadDialogVisible = true">Upload media</AppButton>
    </div>

    <div v-if="loading && items.length === 0" class="gallery-loading">
      <span class="loading-text">Loading…</span>
    </div>

    <div v-else-if="items.length === 0" class="gallery-empty">
      <p class="empty-text">No media yet. Upload screenshots or video highlights to get started.</p>
    </div>

    <template v-else>
      <div class="gallery-grid">
        <div
          v-for="(item, index) in items"
          :key="item.id"
          class="gallery-thumb"
          role="button"
          tabindex="0"
          :aria-label="item.name"
          @click="openLightbox(index)"
          @keydown.enter="openLightbox(index)"
        >
          <img
            v-if="item.mediaType === 'image'"
            :src="item.url"
            :alt="item.name"
            class="thumb-image"
            loading="lazy"
          />
          <div v-else class="thumb-video">
            <span class="video-play-icon" aria-hidden="true">▶</span>
            <span class="video-name">{{ item.name }}</span>
          </div>
        </div>
      </div>

      <div v-if="hasMore" class="load-more-row">
        <AppButton variant="secondary" :loading="loading" @click="loadMore">Load more</AppButton>
        <span class="count-hint">Showing {{ items.length }} of {{ totalCount }}</span>
      </div>
    </template>

    <MediaUploadDialog
      v-model:visible="uploadDialogVisible"
      :entity-type="entityType"
      :entity-id="entityId"
      :entity-label="entityLabel"
      @uploaded="onUploaded"
    />

    <MediaLightbox
      v-if="lightboxVisible && items.length > 0"
      :items="items"
      :initial-index="lightboxIndex"
      :entity-type="entityType"
      :entity-id="entityId"
      :entity-label="entityLabel"
      @close="lightboxVisible = false"
      @removed="onMediaRemoved"
    />
  </section>
</template>

<style scoped>
.media-gallery {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-top: 1.5rem;
}

.gallery-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.gallery-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.gallery-loading,
.gallery-empty {
  padding: 2rem 0;
  text-align: center;
}

.loading-text,
.empty-text {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.gallery-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 0.625rem;
}

.gallery-thumb {
  position: relative;
  aspect-ratio: 16 / 9;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  transition: border-color 0.15s, box-shadow 0.15s;
  outline: none;
}

.gallery-thumb:hover,
.gallery-thumb:focus {
  border-color: var(--color-accent);
  box-shadow: 0 0 0 2px rgba(var(--color-accent-rgb, 88, 130, 228), 0.25);
}

.thumb-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.thumb-video {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  padding: 0.5rem;
}

.video-play-icon {
  font-size: 1.25rem;
  color: var(--color-text-secondary);
}

.video-name {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  word-break: break-word;
}

.load-more-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  justify-content: center;
  padding-top: 0.5rem;
}

.count-hint {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}
</style>
