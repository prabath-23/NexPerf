<script setup>
import { onMounted, ref, watch } from 'vue'
import WorkspaceNav from '../navigation/WorkspaceNav.vue'
import StatusBadge from '../ui/StatusBadge.vue'

defineProps({
  lastUpdated: { type: [Date, null], default: null },
  refreshing: { type: Boolean, default: false },
  system: { type: Object, default: null },
  activeSection: { type: String, default: 'dashboard' },
  navItems: { type: Array, default: () => [] },
  workspaceMeta: {
    type: Object,
    default: () => ({
      title: 'Overview',
      subtitle: 'Live system telemetry and health'
    })
  }
})

const emit = defineEmits(['change-section'])
const brandMark = `${import.meta.env.BASE_URL}brand/nexperf-mark.svg`
const navMode = ref('sidebar')
const sidebarCollapsed = ref(false)
const theme = ref('light')

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

onMounted(() => {
  const saved = window.localStorage.getItem('nexperf-theme')
  const preferred = window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  theme.value = saved || preferred
  sidebarCollapsed.value = window.localStorage.getItem('nexperf-sidebar-collapsed') === 'true'
})

watch(theme, (value) => {
  document.documentElement.dataset.theme = value
  document.documentElement.style.colorScheme = value
  window.localStorage.setItem('nexperf-theme', value)
}, { immediate: true })

watch(sidebarCollapsed, (value) => {
  window.localStorage.setItem('nexperf-sidebar-collapsed', String(value))
})
</script>

<template>
  <main :class="['workspace-shell', `nav-${navMode}`, { 'sidebar-collapsed': sidebarCollapsed }]">
    <aside v-if="navItems.length" class="workspace-sidebar" aria-label="NexPerf workspace sidebar">
      <div class="sidebar-brand">
        <img :src="brandMark" alt="" class="brand-mark" />
        <div>
          <strong>NexPerf</strong>
        </div>
      </div>
      <button
        class="sidebar-collapse"
        type="button"
        @click="toggleSidebar"
        :aria-label="sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        :title="sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path :d="sidebarCollapsed ? 'M9 6l6 6-6 6' : 'M15 6l-6 6 6 6'" />
        </svg>
      </button>
      <WorkspaceNav :items="navItems" :active="activeSection" mode="sidebar" @change="emit('change-section', $event)" />
    </aside>

    <section class="workspace-main">
      <header class="topbar">
        <div class="topbar-surface">
          <div class="workspace-header">
            <div class="workspace-title-block">
              <h1 class="section-title">{{ workspaceMeta.title }}</h1>
              <p class="section-subtitle">{{ workspaceMeta.subtitle }}</p>
            </div>
          </div>
          <div class="topbar-right">
            <button class="theme-toggle" type="button" @click="toggleTheme" :aria-label="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'">
              <span></span>
              {{ theme === 'dark' ? 'Light' : 'Dark' }}
            </button>
            <div class="runtime" aria-label="System status">
              <StatusBadge :label="refreshing ? 'POLLING' : 'LIVE'" tone="success" :pulse="refreshing" />
              <div class="runtime-meta">
                <strong v-if="system">{{ system.os }} / {{ system.arch }}</strong>
                <span v-if="system?.hostname">{{ system.hostname }}</span>
                <span v-if="lastUpdated">Updated {{ lastUpdated.toLocaleTimeString() }}</span>
              </div>
            </div>
          </div>
        </div>
      </header>
      <div class="workspace-scroll">
        <WorkspaceNav v-if="navItems.length && navMode === 'topbar'" :items="navItems" :active="activeSection" mode="topbar" @change="emit('change-section', $event)" />
        <slot />
      </div>
    </section>
  </main>
</template>
