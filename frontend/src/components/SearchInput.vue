<script lang="ts" setup>
defineProps<{
  modelValue: string
  loading: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  search: [value: string]
}>()

function handleInput(e: Event) {
  emit('update:modelValue', (e.target as HTMLInputElement).value)
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    emit('search', (e.target as HTMLInputElement).value)
  }
}
</script>

<template>
  <div class="search-wrap">
    <span class="search-icon" aria-hidden="true">
      <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
        <circle cx="6.5" cy="6.5" r="5" stroke="currentColor" stroke-width="1.5" />
        <line x1="10.5" y1="10.5" x2="14" y2="14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
      </svg>
    </span>
    <input
      :value="modelValue"
      type="search"
      class="search-input"
      :placeholder="placeholder ?? 'Search…'"
      autocomplete="off"
      @input="handleInput"
      @keydown="handleKeydown"
    />
    <span v-if="loading" class="search-spinner" aria-label="Searching" />
  </div>
</template>

<style scoped>
.search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.625rem;
  color: var(--color-text-secondary);
  display: flex;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 0.5rem 2.25rem 0.5rem 2rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
  font-family: var(--font-sans);
  outline: none;
  /* Remove browser default clear button in search inputs */
  appearance: none;
}

.search-input:focus {
  border-color: var(--color-accent);
}

.search-input::placeholder {
  color: var(--color-text-secondary);
}

/* Remove WebKit search cancel button */
.search-input::-webkit-search-cancel-button { display: none; }

.search-spinner {
  position: absolute;
  right: 0.625rem;
  width: 14px;
  height: 14px;
  border: 2px solid var(--color-surface-3);
  border-top-color: var(--color-accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
