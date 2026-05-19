<script setup>
import { dashboardStore, refreshNexPerfStats } from '../../stores/dashboardStore'

defineProps({
  formatBytes: { type: Function, required: true }
})
</script>

<template>
  <section class="panel nexperf-panel">
    <div class="section-head">
      <div>
        <h2>NexPerf Runtime</h2>
        <span>Overhead, storage, samples, and API timings</span>
      </div>
      <button type="button" @click="refreshNexPerfStats">Reload</button>
    </div>
    <div class="runtime-board">
      <article class="runtime-hero">
        <span>version {{ dashboardStore.nexperf?.version }}</span>
        <strong>{{ dashboardStore.nexperf?.database?.sample_count || 0 }}</strong>
        <p>stored samples · {{ formatBytes(dashboardStore.nexperf?.database?.database_size_bytes) }} database</p>
      </article>
      <div class="runtime-stats">
        <article><span>Uptime</span><strong>{{ dashboardStore.nexperf?.uptime }}</strong></article>
        <article><span>Memory</span><strong>{{ formatBytes(dashboardStore.nexperf?.memory_alloc_bytes) }}</strong></article>
        <article><span>Goroutines</span><strong>{{ dashboardStore.nexperf?.goroutines || 0 }}</strong></article>
      </div>
    </div>
    <div class="endpoint-list">
      <article v-for="(timing, path) in dashboardStore.nexperf?.api_response_timings || {}" :key="path">
        <strong>{{ path }}</strong>
        <span>{{ timing.count }} calls</span>
        <em>avg {{ Math.round((timing.average || 0) / 1000000) }}ms</em>
        <small>last {{ Math.round((timing.last || 0) / 1000000) }}ms</small>
      </article>
    </div>
  </section>
</template>
