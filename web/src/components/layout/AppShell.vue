<script setup>
import StatusBadge from '../ui/StatusBadge.vue'

defineProps({
  lastUpdated: { type: [Date, null], default: null },
  refreshing: { type: Boolean, default: false },
  system: { type: Object, default: null }
})

const brandMark = `${import.meta.env.BASE_URL}brand/nexperf-mark.svg`
</script>

<template>
  <main>
    <header class="topbar">
      <div class="brand-lockup">
        <img :src="brandMark" alt="" class="brand-mark" />
        <div>
          <h1 class="wordmark" aria-label="NexPerf">
            <span>N</span><span>E</span><span>X</span><span class="blue">P</span><span class="blue">E</span><span class="blue">R</span><span class="blue">F</span>
          </h1>
          <p class="brand-tagline">SYSTEM INTELLIGENCE. REAL TIME.</p>
        </div>
      </div>
      <div class="runtime">
        <StatusBadge :label="refreshing ? 'POLLING' : 'LIVE'" tone="success" :pulse="refreshing" />
        <strong v-if="system">{{ system.os }} / {{ system.arch }}</strong>
        <span v-if="system?.hostname">{{ system.hostname }}</span>
        <span v-if="lastUpdated">Updated {{ lastUpdated.toLocaleTimeString() }}</span>
      </div>
    </header>
    <slot />
  </main>
</template>
