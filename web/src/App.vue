<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'

const system = ref(null)
const processes = ref([])
const insights = ref([])
const loading = ref(true)
const error = ref('')
let timer

const cards = computed(() => {
  if (!system.value) return []
  return [
    { label: 'CPU', value: percent(system.value.cpu_percent), percent: system.value.cpu_percent, detail: 'Current aggregate usage' },
    {
      label: 'Memory',
      value: percent(system.value.memory.percent),
      percent: system.value.memory.percent,
      detail: `${bytes(system.value.memory.used)} used of ${bytes(system.value.memory.total)}`
    },
    {
      label: 'Disk',
      value: percent(system.value.disk.percent),
      percent: system.value.disk.percent,
      detail: `${bytes(system.value.disk.used)} used of ${bytes(system.value.disk.total)}`
    }
  ]
})

async function refresh() {
  try {
    const [systemRes, processesRes, insightsRes] = await Promise.all([
      fetch('/api/system'),
      fetch('/api/processes/top'),
      fetch('/api/insights')
    ])
    system.value = await systemRes.json()
    processes.value = await processesRes.json()
    insights.value = await insightsRes.json()
    error.value = ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unable to refresh NexPerf data'
  } finally {
    loading.value = false
  }
}

function percent(value) {
  return `${Number(value || 0).toFixed(1)}%`
}

function bytes(value) {
  const gb = Number(value || 0) / 1024 / 1024 / 1024
  return gb >= 10 ? `${gb.toFixed(0)} GB` : `${gb.toFixed(1)} GB`
}

onMounted(() => {
  refresh()
  timer = window.setInterval(refresh, 3000)
})

onUnmounted(() => {
  window.clearInterval(timer)
})
</script>

<template>
  <main>
    <header class="topbar">
      <div>
        <h1>NexPerf</h1>
        <p>Local system intelligence for developer machines.</p>
      </div>
      <div class="machine" v-if="system">
        <strong>{{ system.os }} / {{ system.arch }}</strong>
        <span v-if="system.hostname">{{ system.hostname }}</span>
      </div>
    </header>

    <div v-if="error" class="notice">{{ error }}</div>
    <div v-if="loading" class="notice">Loading NexPerf metrics...</div>

    <section class="metrics" v-if="system">
      <article v-for="card in cards" :key="card.label" class="metric-card">
        <span>{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
        <p>{{ card.detail }}</p>
        <div class="bar">
          <div :style="{ width: `${Math.min(card.percent, 100)}%` }"></div>
        </div>
      </article>
    </section>

    <section class="panel">
      <div class="section-head">
        <h2>Top Processes</h2>
        <span>Sorted by memory</span>
      </div>
      <table>
        <thead>
          <tr>
            <th>PID</th>
            <th>Name</th>
            <th>Memory</th>
            <th>CPU</th>
            <th>User</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="process in processes" :key="process.pid">
            <td>{{ process.pid }}</td>
            <td>{{ process.name }}</td>
            <td>{{ process.memory_mb.toFixed(1) }} MB</td>
            <td>{{ percent(process.cpu_percent) }}</td>
            <td>{{ process.user }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="panel">
      <div class="section-head">
        <h2>Insights</h2>
        <span>Rule-based v0.1 checks</span>
      </div>
      <article v-for="item in insights" :key="item.id" class="insight" :class="item.severity">
        <strong>{{ item.title }}</strong>
        <p>{{ item.message }}</p>
        <span>{{ item.recommendation }}</span>
      </article>
    </section>
  </main>
</template>
