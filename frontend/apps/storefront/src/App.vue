<script setup lang="ts">
import NavBar from './components/NavBar.vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const hideNav = ['Login']
</script>

<template>
  <div id="layout">
    <NavBar v-if="!hideNav.includes(route.name as string)" />
    <main>
      <router-view />
    </main>
  </div>
</template>

<style>
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  background: #f1f5f9; color: #1e293b; line-height: 1.5;
}
a { color: inherit; }
#layout { min-height: 100vh; display: flex; flex-direction: column; }
main { flex: 1; }
</style>
