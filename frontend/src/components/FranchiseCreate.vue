<script lang="ts" setup>
import { ref } from 'vue'

const emit = defineEmits<{
  create: [name: string, gameVersion: string]
  cancel: []
}>()

const name = ref('')
const gameVersion = ref('smb4')
const submitting = ref(false)
const error = ref<string | null>(null)

async function handleSubmit() {
  error.value = null
  if (!name.value.trim()) {
    error.value = 'Franchise name is required'
    return
  }
  submitting.value = true
  try {
    emit('create', name.value.trim(), gameVersion.value)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="franchise-create">
    <h2>New Franchise</h2>

    <div class="field">
      <label for="franchise-name">Name</label>
      <input
        id="franchise-name"
        v-model="name"
        type="text"
        placeholder="e.g. Super Mega League Season 1"
        autocomplete="off"
        @keyup.enter="handleSubmit"
      />
    </div>

    <div class="field">
      <label>Game Version</label>
      <div class="radio-group">
        <label class="radio-option">
          <input v-model="gameVersion" type="radio" value="smb4" />
          Super Mega Baseball 4
        </label>
        <label class="radio-option">
          <input v-model="gameVersion" type="radio" value="smb3" />
          Super Mega Baseball 3
        </label>
      </div>
    </div>

    <p v-if="error" class="error-text">{{ error }}</p>

    <div class="actions">
      <button class="btn-secondary" @click="emit('cancel')">Cancel</button>
      <button class="btn-primary" :disabled="submitting" @click="handleSubmit">
        Create Franchise
      </button>
    </div>
  </div>
</template>

<style scoped>
.franchise-create {
  max-width: 480px;
  margin: 0 auto;
}

h2 {
  font-size: 1.4rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
  color: var(--color-text-primary);
}

.field {
  margin-bottom: 1.25rem;
}

label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 0.4rem;
}

input[type='text'] {
  width: 100%;
  padding: 0.5rem 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
  outline: none;
  box-sizing: border-box;
}

input[type='text']:focus {
  border-color: var(--color-accent);
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.radio-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9375rem;
  color: var(--color-text-primary);
  cursor: pointer;
}

.error-text {
  color: var(--color-error);
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.btn-primary {
  padding: 0.5rem 1.25rem;
  background: var(--color-accent);
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 0.9375rem;
  cursor: pointer;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  padding: 0.5rem 1.25rem;
  background: transparent;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 0.9375rem;
  cursor: pointer;
}
</style>
