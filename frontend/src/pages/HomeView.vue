<script setup lang="ts">
import HelloWorld from "@/components/HelloWorld.vue";
import { useGameContextStore } from "@/stores/game-context";

const gameContextStore = useGameContextStore();
</script>

<template>
  <HelloWorld />
  <p>
    Current game title: {{ gameContextStore.gameTitle }}
  </p>
  <p>
    Current game ID: {{ gameContextStore.gameId }}
  </p>
</template>

<style scoped>

</style>