<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { useFranchiseStore } from './stores/franchise'
import FranchiseSelector from './components/FranchiseSelector.vue'
import FranchiseCreate from './components/FranchiseCreate.vue'

const franchiseStore = useFranchiseStore()
const showCreate = ref(false)
const error = ref<string | null>(null)

onMounted(async () => {
  await franchiseStore.loadFranchises()
})

async function handleCreate(name: string, gameVersion: string) {
  error.value = null
  try {
    const created = await franchiseStore.createFranchise(name, gameVersion)
    showCreate.value = false
    await franchiseStore.selectFranchise(created.id)
  } catch (e) {
    error.value = String(e)
  }
}

async function handleSelect(id: string) {
  error.value = null
  try {
    await franchiseStore.selectFranchise(id)
  } catch (e) {
    error.value = String(e)
  }
}
</script>

<template>
  <div id="app-root" class="dark">
    <!-- Loading state -->
    <div v-if="franchiseStore.loading" class="fullscreen-center">
      <span class="loading-text">Loading…</span>
    </div>

    <!-- Franchise selection / creation -->
    <div v-else-if="!franchiseStore.active" class="fullscreen-center">
      <div class="franchise-setup-panel">
        <div class="app-brand">
          <h1>smb-tools</h1>
          <p>Super Mega Baseball franchise history tracker</p>
        </div>

        <p v-if="error" class="error-text">{{ error }}</p>

        <FranchiseCreate
          v-if="showCreate"
          @create="handleCreate"
          @cancel="showCreate = false"
        />

        <FranchiseSelector
          v-else
          :franchises="franchiseStore.franchises"
          @select="handleSelect"
          @create="showCreate = true"
        />
      </div>
    </div>

    <!-- Main app shell — franchise is selected -->
    <div v-else class="app-shell">
      <aside class="sidebar">
        <div class="sidebar-brand">
          <span class="brand-name">smb-tools</span>
        </div>
        <nav class="sidebar-nav">
          <router-link to="/">Dashboard</router-link>
          <!-- Additional nav items added in Phase 5-6 -->
        </nav>
        <div class="sidebar-footer">
          <span class="active-franchise-name">{{ franchiseStore.active.name }}</span>
          <button class="btn-link" @click="franchiseStore.active = null">Switch</button>
        </div>
      </aside>
      <main class="main-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style>
:root {
  --color-bg:           #0d1117;
  --color-surface-1:    #161b22;
  --color-surface-2:    #21262d;
  --color-surface-3:    #30363d;
  --color-border:       #30363d;
  --color-text-primary: #e6edf3;
  --color-text-secondary: #8b949e;
  --color-accent:       #388bfd;
  --color-error:        #f85149;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  background: var(--color-bg);
  color: var(--color-text-primary);
  height: 100vh;
  overflow: hidden;
}

#app-root {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.fullscreen-center {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.franchise-setup-panel {
  width: 100%;
  max-width: 560px;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.app-brand h1 {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.app-brand p {
  font-size: 0.9375rem;
  color: var(--color-text-secondary);
  margin-top: 0.25rem;
}

.error-text {
  color: var(--color-error);
  font-size: 0.875rem;
}

.loading-text {
  color: var(--color-text-secondary);
}

/* App shell */
.app-shell {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: 220px;
  flex-shrink: 0;
  background: var(--color-surface-1);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
}

.sidebar-brand {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border);
}

.brand-name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.sidebar-nav {
  flex: 1;
  padding: 0.75rem 0;
  display: flex;
  flex-direction: column;
}

.sidebar-nav a {
  display: block;
  padding: 0.5rem 1.25rem;
  color: var(--color-text-secondary);
  text-decoration: none;
  font-size: 0.9375rem;
  border-radius: 4px;
  margin: 0 0.5rem;
}

.sidebar-nav a:hover,
.sidebar-nav a.router-link-active {
  background: var(--color-surface-2);
  color: var(--color-text-primary);
}

.sidebar-footer {
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.active-franchise-name {
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-link {
  background: none;
  border: none;
  color: var(--color-accent);
  font-size: 0.8125rem;
  cursor: pointer;
  padding: 0;
  text-align: left;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  background: var(--color-bg);
}
</style>
