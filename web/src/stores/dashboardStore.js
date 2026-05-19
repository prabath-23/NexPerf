import { reactive } from 'vue'
import { getConfig, getConfigModes, getDashboardWidgets, getHealthScore, getHistory, getInsights, getNexPerfStats, getProcesses, getProcessTree, getStorageSummary, getSystem, saveConfig, saveDashboardWidgets, runTerminalCommand } from '../services/api'

export const dashboardStore = reactive({
  system: null,
  healthScore: null,
  processes: [],
  previousProcesses: [],
  processTree: [],
  processLimit: 500,
  timeRange: '5m',
  insights: [],
  insightLifecycle: [],
  storage: null,
  config: null,
  configModes: [],
  configDirty: false,
  configStatus: '',
  widgetLayout: [],
  terminalHistory: [],
  terminalRunning: false,
  nexperf: null,
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

export async function refreshSystemOverview() {
  const [system, healthScore, cpu, memory, disk] = await Promise.all([
    getSystem(),
    getHealthScore(),
    getHistory('cpu', dashboardStore.timeRange),
    getHistory('memory', dashboardStore.timeRange),
    getHistory('disk', dashboardStore.timeRange)
  ])
  dashboardStore.system = system
  dashboardStore.healthScore = healthScore
  dashboardStore.history.cpu = cpu
  dashboardStore.history.memory = memory
  dashboardStore.history.disk = disk
}

export async function refreshProcesses() {
  const [processes, tree] = await Promise.all([
    getProcesses(dashboardStore.processLimit),
    getProcessTree(500)
  ])
  dashboardStore.previousProcesses = dashboardStore.processes
  dashboardStore.processes = processes
  dashboardStore.processTree = tree
}

export async function refreshInsights() {
  dashboardStore.insights = mergeInsightLifecycle(await getInsights())
}

export async function refreshProcessTree() {
  dashboardStore.processTree = await getProcessTree(500)
}

export async function refreshStorage(path = '') {
  dashboardStore.storage = await getStorageSummary(path)
}

export async function refreshNexPerfStats() {
  dashboardStore.nexperf = await getNexPerfStats()
}

export async function refreshConfig(options = {}) {
  if (dashboardStore.configDirty && !options.force) return
  const [config, modes] = await Promise.all([getConfig(), getConfigModes()])
  dashboardStore.config = config
  dashboardStore.configModes = modes
}

export async function persistConfig() {
  dashboardStore.config = await saveConfig(dashboardStore.config)
  dashboardStore.configDirty = false
  dashboardStore.configStatus = 'Saved and applied to API behavior. Restart service only for collector interval changes.'
}

export async function refreshWidgetLayout() {
  dashboardStore.widgetLayout = await getDashboardWidgets()
}

export async function persistWidgetLayout(widgets) {
  dashboardStore.widgetLayout = await saveDashboardWidgets(widgets)
  return dashboardStore.widgetLayout
}

export async function executeTerminal(command, cwd = '', options = {}) {
  dashboardStore.terminalRunning = true
  try {
    const result = await runTerminalCommand(command, cwd)
    if (options.record !== false) {
      dashboardStore.terminalHistory.unshift(result)
      dashboardStore.terminalHistory = dashboardStore.terminalHistory.slice(0, 20)
    }
    return result
  } finally {
    dashboardStore.terminalRunning = false
  }
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
