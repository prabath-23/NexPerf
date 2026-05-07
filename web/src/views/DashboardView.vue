<script setup>
import { computed } from 'vue'
import TimelineChart from '../components/charts/TimelineChart.vue'
import InsightCard from '../components/insights/InsightCard.vue'
import AppShell from '../components/layout/AppShell.vue'
import MetricCard from '../components/metrics/MetricCard.vue'
import SystemPulse from '../components/metrics/SystemPulse.vue'
import ProcessTable from '../components/processes/ProcessTable.vue'
import { usePolling } from '../composables/usePolling'
import { dashboardStore, refreshDashboard } from '../stores/dashboardStore'

const { loading, refreshing, error, lastUpdated, refresh } = usePolling(refreshDashboard, 3000)

const metricCards = computed(() => {
  const system = dashboardStore.system
  if (!system) return []
  return [
    {
      label: 'CPU',
      value: pct(system.cpu_percent),
      detail: 'Current aggregate processor utilization',
      severity: severity(system.cpu_percent),
      history: dashboardStore.history.cpu,
      trend: trend(dashboardStore.history.cpu),
      peak: peak(dashboardStore.history.cpu)
    },
    {
      label: 'Memory',
      value: pct(system.memory.percent),
      detail: `${bytes(system.memory.used)} used of ${bytes(system.memory.total)}`,
      severity: severity(system.memory.percent),
      history: dashboardStore.history.memory,
      trend: trend(dashboardStore.history.memory),
      peak: peak(dashboardStore.history.memory)
    },
    {
      label: 'Disk',
      value: pct(system.disk.percent),
      detail: `${bytes(system.disk.used)} used of ${bytes(system.disk.total)}`,
      severity: severity(system.disk.percent),
      history: dashboardStore.history.disk,
      trend: trend(dashboardStore.history.disk),
      peak: peak(dashboardStore.history.disk)
    }
  ]
})

const timeRanges = [
  { label: '5 min', value: '5m' },
  { label: '15 min', value: '15m' },
  { label: '1 hour', value: '1h' }
]

function pct(value) {
  return `${Number(value || 0).toFixed(1)}%`
}

function bytes(value) {
  const gb = Number(value || 0) / 1024 / 1024 / 1024
  return gb >= 10 ? `${gb.toFixed(0)} GB` : `${gb.toFixed(1)} GB`
}

function severity(value) {
  if (value >= 90) return 'critical'
  if (value >= 80) return 'warning'
  return 'normal'
}

function trend(points) {
  if (!points || points.length < 2) return { label: 'collecting trend', direction: 'flat', delta: 0 }
  const first = Number(points[0].value || 0)
  const last = Number(points[points.length - 1].value || 0)
  const delta = last - first
  if (Math.abs(delta) < 1) return { label: 'stable over window', direction: 'flat', delta }
  return { label: `${Math.abs(delta).toFixed(1)} pts over window`, direction: delta > 0 ? 'up' : 'down', delta }
}

function peak(points) {
  if (!points?.length) return 'peak --'
  const max = Math.max(...points.map((point) => Number(point.value || 0)))
  return `peak ${max.toFixed(1)}%`
}
</script>

<template>
  <AppShell :last-updated="lastUpdated" :refreshing="refreshing" :system="dashboardStore.system">
    <div v-if="error" class="notice">{{ error }}</div>
    <div v-if="loading" class="notice">Loading NexPerf observability data...</div>

    <section class="overview-stage" v-if="dashboardStore.system">
      <SystemPulse :system="dashboardStore.system" :score="dashboardStore.healthScore" />
      <div class="metrics-grid">
        <MetricCard
          v-for="card in metricCards"
          :key="card.label"
          :label="card.label"
          :value="card.value"
          :detail="card.detail"
          :severity="card.severity"
          :history="card.history"
          :trend="card.trend"
          :peak="card.peak"
        />
      </div>
    </section>

    <section class="control-strip" v-if="dashboardStore.system">
      <div>
        <span>Observation window</span>
        <strong>{{ timeRanges.find((item) => item.value === dashboardStore.timeRange)?.label }}</strong>
      </div>
      <div class="segmented-control">
        <button
          v-for="range in timeRanges"
          :key="range.value"
          type="button"
          :class="{ active: dashboardStore.timeRange === range.value }"
          @click="dashboardStore.timeRange = range.value; refresh()"
        >
          {{ range.label }}
        </button>
      </div>
    </section>

    <section class="timeline-grid" v-if="dashboardStore.system">
      <TimelineChart title="CPU Timeline" :points="dashboardStore.history.cpu" />
      <TimelineChart title="Memory Pressure" :points="dashboardStore.history.memory" />
      <TimelineChart title="Disk Usage History" :points="dashboardStore.history.disk" />
    </section>

    <section class="workspace-grid">
      <ProcessTable
        :processes="dashboardStore.processes"
        :previous-processes="dashboardStore.previousProcesses"
        :limit="dashboardStore.processLimit"
        @update:limit="(limit) => { dashboardStore.processLimit = limit; refresh() }"
      />

      <section class="panel insights-panel">
        <div class="section-head">
          <div>
            <h2>Insights</h2>
            <span>Context from current metrics and processes</span>
          </div>
          <button type="button" @click="refresh">Refresh</button>
        </div>
        <div class="insight-grid">
          <InsightCard v-for="item in dashboardStore.insights" :key="item.id" :insight="item" />
        </div>
      </section>
    </section>
  </AppShell>
</template>
