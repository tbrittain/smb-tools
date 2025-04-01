<script lang="ts" setup>
</script>

<template>
  <h1>SMB Tools</h1>
  <p>
    <strong>Current route path:</strong> <span style="font-family: monospace;">{{ $route.fullPath }}</span>
  </p>
  <nav style="display: flex; flex-direction: column; gap: 1rem;">
    <RouterLink to="/">Go to Home</RouterLink>
    <RouterLink to="/about">Go to About</RouterLink>
  </nav>
  <main>
    <RouterView />
  </main>
</template>

