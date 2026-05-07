import { reactive } from 'vue'
import { getHealthScore, getHistory, getInsights, getProcesses, getSystem } from '../services/api'

export const dashboardStore = reactive({
  system: null,
  healthScore: null,
  processes: [],
  previousProcesses: [],
  processLimit: 12,
  timeRange: '5m',
  insights: [],
  insightLifecycle: [],
  history: {
    cpu: [],
    memory: [],
    disk: []
  }
})

export async function refreshDashboard() {
  const [system, healthScore, processes, insights, cpu, memory, disk] = await Promise.all([
    getSystem(),
    getHealthScore(),
    getProcesses(dashboardStore.processLimit),
    getInsights(),
    getHistory('cpu', dashboardStore.timeRange),
    getHistory('memory', dashboardStore.timeRange),
    getHistory('disk', dashboardStore.timeRange)
  ])
  dashboardStore.system = system
  dashboardStore.healthScore = healthScore
  dashboardStore.previousProcesses = dashboardStore.processes
  dashboardStore.processes = processes
  dashboardStore.insights = mergeInsightLifecycle(insights)
  dashboardStore.history.cpu = cpu
  dashboardStore.history.memory = memory
  dashboardStore.history.disk = disk
}

function mergeInsightLifecycle(incoming) {
  const now = new Date().toISOString()
  const currentIds = new Set(incoming.map((item) => item.id))
  const previous = new Map(dashboardStore.insightLifecycle.map((item) => [item.id, item]))
  const active = incoming.map((item) => {
    const existing = previous.get(item.id)
    return {
      ...item,
      state: existing?.state === 'resolved' ? 'active' : (existing?.state || 'active'),
      first_seen: existing?.first_seen || now,
      last_seen: now
    }
  })
  const resolved = dashboardStore.insightLifecycle
    .filter((item) => !currentIds.has(item.id) && item.state !== 'expired')
    .map((item) => ({ ...item, state: 'resolved', resolved_at: item.resolved_at || now }))
    .filter((item) => Date.now() - new Date(item.resolved_at).getTime() < 120000)
  dashboardStore.insightLifecycle = [...active, ...resolved].sort((a, b) => Number(b.score || 0) - Number(a.score || 0))
  return dashboardStore.insightLifecycle
}
