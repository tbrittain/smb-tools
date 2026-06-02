<script lang="ts" setup>
import Button from 'primevue/button'
import { onMounted, ref } from 'vue'
import { GetAllMediaForTeam } from '../../wailsjs/go/main/App'
import type { main } from '../../wailsjs/go/models'
import MediaLightbox from './MediaLightbox.vue'
import MediaUploadDialog from './MediaUploadDialog.vue'

const props = defineProps<{
  teamId: number
}>()

const groups = ref<main.TeamSeasonMediaGroupDTO[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

type LightboxState = {
  items: main.MediaItemDTO[]
  teamHistoryId: number
  entityLabel: string
}
const lightboxState = ref<LightboxState | null>(null)
const lightboxIndex = ref(0)

type UploadState = {
  teamHistoryId: number
  entityLabel: string
}
const uploadState = ref<UploadState | null>(null)
const uploadVisible = ref(false)

async function load() {
  loading.value = true
  error.value = null
  try {
    groups.value = (await GetAllMediaForTeam(props.teamId)) ?? []
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function openLightbox(group: main.TeamSeasonMediaGroupDTO, index: number) {
  lightboxState.value = {
    items: group.items,
    teamHistoryId: group.teamHistoryId,
    entityLabel: `${group.teamName} S${group.seasonNum}`,
  }
  lightboxIndex.value = index
}

function openUpload(group: main.TeamSeasonMediaGroupDTO) {
  uploadState.value = {
    teamHistoryId: group.teamHistoryId,
    entityLabel: `${group.teamName} S${group.seasonNum}`,
  }
  uploadVisible.value = true
}

function onUploaded() {
  uploadVisible.value = false
  load()
}

function onMediaRemoved(removedId: string) {
  for (const group of groups.value) {
    const idx = group.items.findIndex((i) => i.id === removedId)
    if (idx !== -1) {
      group.items.splice(idx, 1)
    }
  }
  groups.value = groups.value.filter((g) => g.items.length > 0)
  if (lightboxState.value) {
    const remaining = lightboxState.value.items.filter((i) => i.id !== removedId)
    lightboxState.value = remaining.length === 0 ? null : { ...lightboxState.value, items: remaining }
  }
}

onMounted(load)
</script>

<template>
  <section class="team-media-gallery">
    <h3 class="gallery-title">Media</h3>

    <div v-if="loading" class="gallery-status">
      <span class="status-text">Loading…</span>
    </div>
    <div v-else-if="error" class="gallery-status">
      <span class="status-text error-text">{{ error }}</span>
    </div>
    <div v-else-if="groups.length === 0" class="gallery-status">
      <span class="status-text">No media uploaded yet. Upload from an individual season page to get started.</span>
    </div>

    <template v-else>
      <div v-for="group in groups" :key="group.teamHistoryId" class="season-group">
        <div class="season-group-header">
          <span class="season-label">Season {{ group.seasonNum }}</span>
          <Button
            label="Upload media"
            size="small"
            severity="secondary"
            outlined
            @click="openUpload(group)"
          />
        </div>
        <div class="gallery-grid">
          <div
            v-for="(item, index) in group.items"
            :key="item.id"
            class="gallery-thumb"
            role="button"
            tabindex="0"
            :aria-label="item.name"
            @click="openLightbox(group, index)"
            @keydown.enter="openLightbox(group, index)"
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
      </div>
    </template>

    <MediaLightbox
      v-if="lightboxState !== null"
      :items="lightboxState.items"
      :initial-index="lightboxIndex"
      entity-type="team_season"
      :entity-id="lightboxState.teamHistoryId"
      :entity-label="lightboxState.entityLabel"
      @close="lightboxState = null"
      @removed="onMediaRemoved"
    />

    <MediaUploadDialog
      v-if="uploadState !== null"
      v-model:visible="uploadVisible"
      entity-type="team_season"
      :entity-id="uploadState.teamHistoryId"
      :entity-label="uploadState.entityLabel"
      @uploaded="onUploaded"
    />
  </section>
</template>

<style scoped>
.team-media-gallery {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.gallery-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.gallery-status {
  padding: 1rem 0;
}

.status-text {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.error-text {
  color: var(--color-error);
}

.season-group {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.season-group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.season-label {
  font-size: 0.875rem;
  font-weight: 600;
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
</style>
