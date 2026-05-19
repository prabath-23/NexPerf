<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import AppShell from '../components/layout/AppShell.vue'
import { usePolling } from '../composables/usePolling'
import { metaForSection, navItems, sectionFromURL, updateSectionURL } from '../router'
import { cacheDashboardWidgets } from '../services/api'
import { dashboardStore, executeTerminal, persistWidgetLayout, refreshConfig, refreshDashboard, refreshInsights, refreshNexPerfStats, refreshProcesses, refreshStorage, refreshSystemOverview, refreshWidgetLayout } from '../stores/dashboardStore'
import { defaultWidgetIds, normalizeWidgetSize, resolveWidgetTemplate, widgetLibraryGroups, widgetTemplateMap } from '../widgets/registry'
import InsightsSection from './sections/InsightsSection.vue'
import ManualSection from './sections/ManualSection.vue'
import NexPerfSection from './sections/NexPerfSection.vue'
import PlaceholderSection from './sections/PlaceholderSection.vue'
import ProcessesSection from './sections/ProcessesSection.vue'
import SettingsSection from './sections/SettingsSection.vue'
import StorageSection from './sections/StorageSection.vue'
import TerminalSection from './sections/TerminalSection.vue'
import WidgetBoardSection from './sections/WidgetBoardSection.vue'

const activeSection = ref(sectionFromURL())
const workspaceMeta = computed(() => metaForSection(activeSection.value))
const { loading, refreshing, error, lastUpdated, refresh } = usePolling(refreshActiveSection, 3000)

const metricCards = computed(() => {
  const system = dashboardStore.system
  if (!system) return []
  return [
    {
      label: 'CPU',
      value: pct(system.cpu_percent),
      detail: 'Workload activity across active cores',
      severity: severity(system.cpu_percent),
      history: dashboardStore.history.cpu,
      trend: trend(dashboardStore.history.cpu),
      peak: peak(dashboardStore.history.cpu)
    },
    {
      label: 'Memory',
      value: pct(system.memory.percent),
      detail: `${bytes(system.memory.used)} pressure from ${bytes(system.memory.total)} total`,
      severity: severity(system.memory.percent),
      history: dashboardStore.history.memory,
      trend: trend(dashboardStore.history.memory),
      peak: peak(dashboardStore.history.memory)
    },
    {
      label: 'Disk',
      value: pct(system.disk.percent),
      detail: `${bytes(system.disk.used)} allocated of ${bytes(system.disk.total)} capacity`,
      severity: severity(system.disk.percent),
      history: dashboardStore.history.disk,
      trend: trend(dashboardStore.history.disk),
      peak: peak(dashboardStore.history.disk)
    }
  ]
})

const topProcesses = computed(() => dashboardStore.processes.slice(0, 4))

const topInsight = computed(() => dashboardStore.insights[0])

const processCategoryBreakdown = computed(() => {
  const totalMemory = Number(dashboardStore.system?.memory?.total || 0) / 1024 / 1024
  const groups = new Map()
  for (const process of dashboardStore.processes) {
    const key = process.category || 'unknown'
    const group = groups.get(key) || {
      id: key,
      label: process.category_label || titleCase(key),
      count: 0,
      memoryMB: 0,
      cpu: 0,
      topProcess: process.name
    }
    group.count += 1
    group.memoryMB += Number(process.memory_mb || 0)
    group.cpu += Number(process.cpu_percent || 0)
    if (Number(process.memory_mb || 0) > Number(group.topMemory || 0)) {
      group.topMemory = Number(process.memory_mb || 0)
      group.topProcess = process.name
    }
    groups.set(key, group)
  }
  return [...groups.values()]
    .map((group) => ({
      ...group,
      memoryShare: totalMemory ? (group.memoryMB / totalMemory) * 100 : 0,
      display: `${group.memoryMB.toFixed(0)} MB`
    }))
    .sort((a, b) => (b.memoryShare + b.cpu / 3) - (a.memoryShare + a.cpu / 3))
})

const observability = computed(() => {
  const metrics = Object.fromEntries(metricCards.value.map((card) => [card.label, metricIntelligence(card)]))
  const topGroup = processCategoryBreakdown.value[0]
  const browserGroup = processCategoryBreakdown.value.find((group) => group.id === 'browser') || null
  const developmentGroup = processCategoryBreakdown.value.find((group) => group.id === 'development') || null
  const serviceGroup = processCategoryBreakdown.value.find((group) => group.id === 'service') || null
  const leadGroup = browserGroup || developmentGroup || topGroup
  const activeInsight = topInsight.value
  const pressureSignals = [metrics.CPU, metrics.Memory, metrics.Disk].filter((item) => item?.isPressure)
  const reasoning = activeInsight
    ? `${activeInsight.message} ${activeInsight.recommendation}`
    : leadGroup
      ? `${leadGroup.label} workload leads current attribution with ${leadGroup.count} processes, ${leadGroup.memoryShare.toFixed(1)}% memory share, and ${leadGroup.cpu.toFixed(1)}% sampled CPU.`
      : 'No dominant workload yet. Keep observing while reproducing the slowdown.'

  return {
    metrics,
    categories: processCategoryBreakdown.value,
    workloads: {
      browser: workloadInsight(browserGroup, 'Browser', 'browser, WebKit, or renderer workload'),
      development: workloadInsight(developmentGroup, 'IDE', 'editor, extension host, build, or coding-agent workload'),
      service: workloadInsight(serviceGroup, 'Background services', 'helper, agent, updater, or daemon workload')
    },
    pressure: {
      label: pressureSignals.length ? `${pressureSignals.length} pressure signals` : 'Pressure normal',
      detail: pressureSignals.length
        ? pressureSignals.map((item) => `${item.label} ${item.current}`).join(' · ')
        : 'No sustained pressure across the active observation window.',
      state: pressureSignals.length >= 2 ? 'correlated' : pressureSignals.length ? 'localized' : 'stable'
    },
    action: {
      title: activeInsight?.title || (leadGroup ? `${leadGroup.label} workload is most prominent` : 'System steady'),
      message: reasoning,
      recommendation: activeInsight?.recommendation || (leadGroup ? `Inspect ${leadGroup.topProcess} and related ${leadGroup.label.toLowerCase()} processes before ending anything.` : 'Keep the board live while reproducing the workload.'),
      citedSignals: [
        metrics.CPU ? `CPU ${metrics.CPU.current}, ${metrics.CPU.state}` : '',
        metrics.Memory ? `Memory ${metrics.Memory.current}, ${metrics.Memory.state}` : '',
        metrics.Disk ? `Disk ${metrics.Disk.current}, ${metrics.Disk.state}` : '',
        leadGroup ? `${leadGroup.label}: ${leadGroup.count} processes, ${leadGroup.memoryShare.toFixed(1)}% memory` : ''
      ].filter(Boolean)
    }
  }
})

const draggedWidget = ref('')
const widgetLibraryOpen = ref(false)
const layoutLoaded = ref(false)
let widgetPersistTimer = 0
const widgetTemplates = widgetTemplateMap
const defaultWidgets = defaultWidgetIds.map((id) => normalizeWidget(resolveWidgetTemplate(id))).filter(Boolean)
const widgets = ref(defaultWidgets.map((widget) => ({ ...widget })))

const widgetLibrary = widgetLibraryGroups

const processTreeParents = computed(() => new Map(dashboardStore.processTree.map((process) => [process.pid, process])))

const timeRanges = [
  { label: '5 min', value: '5m' },
  { label: '15 min', value: '15m' },
  { label: '1 hour', value: '1h' }
]

const terminalCommand = ref('nexperf status')
const terminalCwd = ref('')
const terminalTab = ref('terminal')
const terminalResult = ref(null)
const manualQuery = ref('netstat')
const manualResult = ref(null)
const manualLoading = ref(false)

const commandManual = [
  { command: 'nexperf start', description: 'Start the local NexPerf service and dashboard.' },
  { command: 'nexperf stop', description: 'Stop the running service cleanly.' },
  { command: 'nexperf status', description: 'Print CPU, memory, disk, OS, and architecture summary.' },
  { command: 'nexperf processes', description: 'List top processes by memory and sampled CPU.' },
  { command: 'nexperf inspect', description: 'Show current system inspection and warnings.' },
  { command: 'nexperf explain storage', description: 'Explain storage state with live disk context.' },
  { command: 'nexperf manual storage', description: 'Print directory guidance for common filesystem locations.' }
]

const processManual = [
  { label: 'Browser', detail: 'Browser apps and renderer/helper processes.' },
  { label: 'Communication', detail: 'Messaging, meetings, voice/video, and collaboration tools.' },
  { label: 'Productivity', detail: 'Documents, notes, PDFs, office apps, and planning tools.' },
  { label: 'Creative', detail: 'Design, media, editing, rendering, and audio/video tools.' },
  { label: 'Development', detail: 'Editors, IDEs, coding agents, compilers, and project tooling.' },
  { label: 'Service', detail: 'Background helpers, agents, updaters, and daemon-like processes.' },
  { label: 'System', detail: 'Operating-system managed processes and platform services.' }
]

const modeCopy = computed(() => {
  const mode = dashboardStore.configModes.find((item) => item.name === dashboardStore.config?.usage_mode)
  if (!mode) return null
  const rows = mode.storage_limits?.max_directory_rows || 0
  const depth = mode.storage_limits?.max_scan_depth || 0
  return {
    label: mode.label,
    description: mode.description,
    stats: [
      `${mode.polling_interval_seconds}s sampling`,
      `${mode.dashboard_refresh_ms}ms UI refresh`,
      `${mode.retention_hours}h retention`,
      `${depth} level storage scan`,
      `${rows} storage rows`
    ]
  }
})

const configModeOptions = computed(() => dashboardStore.configModes.map((mode) => ({
  label: mode.label,
  value: mode.name,
  detail: mode.description
})))

const hierarchyGroups = computed(() => {
  const groups = new Map()
  for (const process of dashboardStore.processTree) {
    const parentKey = process.ppid || 0
    const parent = processTreeParents.value.get(process.ppid)
    const group = groups.get(parentKey) || {
      key: parentKey,
      parentName: parent ? parent.name : (parentKey === 0 ? 'Root processes' : `PID ${parentKey}`),
      parentPid: parent?.pid || parentKey || null,
      children: []
    }
    group.children.push(process)
    groups.set(parentKey, group)
  }
  return [...groups.values()]
    .map((group) => ({ ...group, children: group.children.sort((a, b) => Number(b.memory_mb || 0) - Number(a.memory_mb || 0)) }))
    .sort((a, b) => b.children.length - a.children.length)
})

const actionableItems = computed(() => {
  const actionMap = new Map()
  const appendAction = (key, action, insight) => {
    const existing = actionMap.get(key)
    if (existing) {
      existing.count += 1
      existing.sources.push(insight.title || insight.message)
      return
    }
    actionMap.set(key, {
      ...action,
      id: key,
      count: 1,
      sources: [insight.title || insight.message]
    })
  }

  for (const insight of dashboardStore.insights) {
    const category = String(insight.category || insight.id || '').toLowerCase()
    if (category.includes('storage') || category.includes('disk')) {
      appendAction('storage-pressure', {
        title: 'Inspect storage pressure',
        detail: insight.recommendation || insight.message,
        command: 'du -sh /private/var/folders /private/tmp /private/var/tmp ~/Downloads ~/Library/Caches 2>/dev/null | sort -h',
        primary: 'Open Storage',
        target: 'storage'
      }, insight)
      continue
    }
    if (category.includes('process')) {
      appendAction('process-contributors', {
        title: 'Inspect process contributors',
        detail: insight.recommendation || insight.message,
        command: 'ps -axo pid,ppid,%cpu,%mem,rss,comm | sort -k3 -nr | head -30',
        primary: 'Open Processes',
        target: 'processes'
      }, insight)
      continue
    }
    appendAction('runtime-pressure', {
      title: 'Inspect runtime pressure',
      detail: insight.recommendation || insight.message,
      command: 'top -l 1 -o cpu | head -35',
      primary: 'Open Terminal',
      target: 'terminal'
    }, insight)
  }

  const items = [...actionMap.values()].map((action) => ({
    ...action,
    detail: action.count > 1
      ? `${action.detail} ${action.count} related observations are grouped here.`
      : action.detail
  }))
  if (items.length) return items
  return [
    {
      id: 'process-review',
      title: 'Review active workloads',
      detail: 'Inspect CPU and memory contributors before taking process actions.',
      command: 'ps -axo pid,ppid,%cpu,%mem,rss,comm | sort -k4 -nr | head -30',
      primary: 'Open Processes',
      target: 'processes'
    },
    {
      id: 'storage-review',
      title: 'Review cleanup candidates',
      detail: 'Check caches, downloads, temporary folders, and generated data before deleting files.',
      command: 'du -sh /private/var/folders /private/tmp /private/var/tmp ~/Downloads ~/Library/Caches 2>/dev/null | sort -h',
      primary: 'Open Storage',
      target: 'storage'
    }
  ]
})

watch(activeSection, async (section) => {
  updateSectionURL(section)
  if (section === 'processes') {
    dashboardStore.processLimit = 500
    await refreshProcesses()
  } else if (section === 'storage') {
    await refreshStorage()
  } else if (section === 'insights') {
    await refreshInsights()
  } else if (section === 'nexperf') {
    await refreshNexPerfStats()
  } else if (section === 'settings') {
    await refreshConfig()
  } else if (section === 'dashboard') {
    await refreshDashboard()
  }
}, { immediate: false })

watch(widgets, () => {
  if (!layoutLoaded.value) return
  const nextLayout = widgets.value.map((widget) => ({ ...widget }))
  cacheDashboardWidgets(nextLayout)
  clearTimeout(widgetPersistTimer)
  widgetPersistTimer = window.setTimeout(() => {
    persistWidgetLayout(nextLayout).catch(() => {})
  }, 350)
}, { deep: true })

onMounted(async () => {
  window.addEventListener('popstate', syncSectionFromURL)
  try {
    await refreshWidgetLayout()
    if (Array.isArray(dashboardStore.widgetLayout) && dashboardStore.widgetLayout.length) {
      const rawLayout = dashboardStore.widgetLayout
      const storedWidgets = rawLayout.map(normalizeWidget).filter(Boolean)
      if (shouldReplaceLegacyWidgetLayout(rawLayout, storedWidgets)) {
        widgets.value = defaultWidgets.map((widget) => ({ ...widget }))
        persistWidgetLayout(widgets.value.map((widget) => ({ ...widget }))).catch(() => {})
      } else {
        widgets.value = storedWidgets
      }
    }
  } finally {
    layoutLoaded.value = true
  }
})

onUnmounted(() => {
  window.removeEventListener('popstate', syncSectionFromURL)
  clearTimeout(widgetPersistTimer)
})

function changeSection(section) {
  activeSection.value = section
}

async function refreshActiveSection() {
  switch (activeSection.value) {
    case 'dashboard':
      await refreshDashboard()
      break
    case 'processes':
      await refreshProcesses()
      break
    case 'storage':
      await refreshStorage()
      break
    case 'insights':
      await refreshInsights()
      break
    case 'nexperf':
      await refreshNexPerfStats()
      break
    case 'settings':
      if (!dashboardStore.config) await refreshConfig()
      break
    case 'terminal':
    case 'manual':
      break
    default:
      await refreshSystemOverview()
  }
}

function syncSectionFromURL() {
  activeSection.value = sectionFromURL()
}

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

function healthTone(score) {
  const value = Number(score?.value ?? 100)
  if (value < 50) return 'critical'
  if (value < 75) return 'warning'
  return 'normal'
}

function metricByLabel(label) {
  return metricCards.value.find((card) => card.label === label) || metricCards.value[0]
}

function timelinePoints(metric) {
  if (metric === 'Memory') return dashboardStore.history.memory
  if (metric === 'Disk') return dashboardStore.history.disk
  return dashboardStore.history.cpu
}

function windowLabel() {
  return timeRanges.find((item) => item.value === dashboardStore.timeRange)?.label || '5 min'
}

function widgetTone(widget) {
  if (widget.kind === 'host' || widget.kind === 'health') return healthTone(dashboardStore.healthScore)
  if (widget.kind === 'metric') return metricByLabel(widget.metric)?.severity || 'normal'
  return widget.kind
}

function processBarItems() {
  return topProcesses.value.map((process) => ({
    label: process.name,
    value: process.memory_mb,
    display: `${process.memory_mb.toFixed(0)} MB`
  }))
}

function metricSegments(metric) {
  const card = metricByLabel(metric)
  const used = Number(String(card?.value || '0').replace('%', ''))
  return [
    { label: card?.label || metric, value: used, color: used >= 80 ? 'var(--amber)' : 'var(--blue)' },
    { label: 'free', value: Math.max(0, 100 - used), color: 'rgba(148, 163, 184, 0.34)' }
  ]
}

function widgetDescription(widget) {
  if (widget.description) return widget.description
  if (widget.subtitle) return widget.subtitle
  if (widget.kind === 'metric') return `${widget.metric} ${widget.chart}`
  if (widget.kind === 'timeline') return `${widget.metric} history`
  if (widget.kind === 'processes') return 'Top memory contributors'
  if (widget.kind === 'window') return `${windowLabel()} sampling`
  return widget.family || widget.kind
}

function widgetKey(widget) {
  const key = widget?.defId || widget?.id
  const legacyKeys = {
    'runtime-compact': 'runtime-host',
    'cpu-line': 'cpu-volatility',
    'cpu-sparkline': 'cpu-volatility',
    'cpu-ring': 'cpu-volatility',
    'cpu-value': 'cpu-volatility',
    'cpu-bars': 'cpu-bursts',
    'cpu-history': 'cpu-timeline',
    'memory-donut': 'memory-pressure',
    'memory-value': 'memory-pressure',
    'memory-area': 'memory-pressure',
    'disk-progress': 'disk-growth',
    'disk-bars': 'disk-growth',
    'wide-history': 'multi-metric-timeline'
  }
  return legacyKeys[key] || key
}

function shouldReplaceLegacyWidgetLayout(rawLayout, layout) {
  const rawKeys = rawLayout.map((widget) => widgetKey(widget))
  const missingDefault = defaultWidgetIds.some((id) => !rawKeys.includes(id))
  const deprecatedIds = new Set([
    'cpu-ring',
    'cpu-value',
    'cpu-sparkline',
    'cpu-bars',
    'memory-donut',
    'memory-value',
    'memory-area',
    'disk-progress',
    'disk-bars',
    'health-ring',
    'active-process-count'
  ])
  const hasDeprecated = rawLayout.some((widget) => deprecatedIds.has(widget?.defId || widget?.id))
  const legacyMetricCount = layout.filter((widget) => ['CPU', 'Memory', 'Disk'].includes(widget.metric) && widget.kind === 'metric').length
  const intelligenceCount = layout.filter((widget) => ['workload', 'pressure', 'summary', 'actionable'].includes(widget.kind)).length
  return hasDeprecated || (missingDefault && legacyMetricCount >= 3) || (legacyMetricCount >= 4 && intelligenceCount < 3)
}

function canAddWidget(template) {
  const key = widgetKey(template)
  return !!key && !widgets.value.some((widget) => widgetKey(widget) === key)
}

function addWidget(template) {
  if (!canAddWidget(template)) return false
  widgets.value.push(normalizeWidget({ ...template, id: widgetKey(template), defId: widgetKey(template) }))
  return true
}

function removeWidget(id) {
  widgets.value = widgets.value.filter((widget) => widget.id !== id)
}

function dragWidget(id, event) {
  draggedWidget.value = id
  event?.dataTransfer?.setData('text/plain', id)
  if (event?.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}

function dragLibraryWidget(widget, event) {
  draggedWidget.value = `library:${JSON.stringify(widget)}`
  event?.dataTransfer?.setData('text/plain', draggedWidget.value)
  if (event?.dataTransfer) event.dataTransfer.effectAllowed = 'copy'
}

function dropWidget(targetId) {
  const sourceId = draggedWidget.value
  draggedWidget.value = ''
  if (!sourceId || sourceId === targetId) return
  if (sourceId.startsWith('library:')) {
    const template = JSON.parse(sourceId.replace('library:', ''))
    const targetIndex = widgets.value.findIndex((widget) => widget.id === targetId)
    if (!canAddWidget(template)) return
    const nextWidget = normalizeWidget({ ...template, id: widgetKey(template), defId: widgetKey(template) })
    widgets.value.splice(targetIndex >= 0 ? targetIndex : widgets.value.length, 0, nextWidget)
    return
  }
  const sourceIndex = widgets.value.findIndex((widget) => widget.id === sourceId)
  const targetIndex = widgets.value.findIndex((widget) => widget.id === targetId)
  if (sourceIndex < 0 || targetIndex < 0) return
  const [moved] = widgets.value.splice(sourceIndex, 1)
  widgets.value.splice(targetIndex, 0, moved)
}

function dropOnBoard() {
  const sourceId = draggedWidget.value
  draggedWidget.value = ''
  if (!sourceId?.startsWith('library:')) return
  addWidget(JSON.parse(sourceId.replace('library:', '')))
}

function normalizeWidget(widget) {
  const key = widgetKey(widget)
  const template = widgetTemplates.get(key)
  if (!template) return { ...widget, id: key, defId: key }
  const requestedSize = normalizeWidgetSize(widget.size || template.defaultSize || template.size)
  const size = template.supportedSizes?.includes(requestedSize)
    ? requestedSize
    : normalizeWidgetSize(template.defaultSize || template.size)
  return {
    ...template,
    ...widget,
    ...template,
    id: key,
    defId: key,
    size
  }
}

function trend(points) {
  if (!points || points.length < 2) return { label: 'collecting trend', direction: 'flat', delta: 0 }
  const first = Number(points[0].value || 0)
  const last = Number(points[points.length - 1].value || 0)
  const delta = last - first
  if (Math.abs(delta) < 1) return { label: 'stable over window', direction: 'flat', delta }
  return { label: `${Math.abs(delta).toFixed(1)} pts over window`, direction: delta > 0 ? 'up' : 'down', delta }
}

function metricIntelligence(card) {
  const values = (card.history || []).map((point) => Number(point.value || 0)).filter((value) => Number.isFinite(value))
  const current = Number(String(card.value || '0').replace('%', ''))
  const max = values.length ? Math.max(...values) : current
  const min = values.length ? Math.min(...values) : current
  const avg = values.length ? values.reduce((sum, value) => sum + value, 0) / values.length : current
  const variance = values.length ? values.reduce((sum, value) => sum + ((value - avg) ** 2), 0) / values.length : 0
  const volatility = Math.sqrt(variance)
  const burstCount = values.filter((value) => value >= Math.max(70, avg + 18)).length
  const sustainedCount = values.slice(-6).filter((value) => value >= 80).length
  const state = sustainedCount >= 3
    ? 'sustained pressure'
    : burstCount >= 3
      ? 'bursty'
      : volatility >= 12
        ? 'unstable'
        : Math.abs(card.trend?.delta || 0) >= 3
          ? (card.trend.delta > 0 ? 'rising' : 'recovering')
          : 'stable'
  return {
    label: card.label,
    current: card.value,
    detail: card.detail,
    avg: `avg ${avg.toFixed(1)}%`,
    averageValue: avg,
    peak: `peak ${max.toFixed(1)}%`,
    low: `low ${min.toFixed(1)}%`,
    delta: card.trend?.delta || 0,
    deltaLabel: `${card.trend?.delta > 0 ? '+' : ''}${(card.trend?.delta || 0).toFixed(1)} pts`,
    trendLabel: card.trend?.label || 'collecting trend',
    direction: card.trend?.direction || 'flat',
    volatility,
    volatilityLabel: volatility < 4 ? 'low volatility' : volatility < 12 ? 'moderate volatility' : 'high volatility',
    burstCount,
    anomalyCount: values.filter((value) => value >= Math.max(85, avg + 22)).length,
    sustainedCount,
    sampleDensity: `${values.length} samples`,
    confidence: values.length >= 20 ? 'high confidence' : values.length >= 8 ? 'medium confidence' : 'warming up',
    state,
    isPressure: current >= 80 || sustainedCount >= 3,
    interpretation: metricInterpretation(card.label, state, current, max, card.trend?.delta || 0)
  }
}

function metricInterpretation(label, state, current, max, delta) {
  if (state === 'sustained pressure') return `${label} has stayed elevated, not just spiked.`
  if (state === 'bursty') return `${label} shows repeated bursts; inspect workload attribution.`
  if (state === 'unstable') return `${label} is volatile across the observation window.`
  if (state === 'rising') return `${label} is up ${Math.abs(delta).toFixed(1)} pts in the active window.`
  if (state === 'recovering') return `${label} is recovering from a ${max.toFixed(1)}% peak.`
  return `${label} is stable at ${current.toFixed(1)}%.`
}

function workloadInsight(group, label, fallback) {
  if (!group) {
    return {
      label,
      status: 'not prominent',
      detail: `No ${fallback} is prominent in the current top process sample.`,
      count: 0,
      memoryShare: 0,
      cpu: 0,
      topProcess: 'none'
    }
  }
  const status = group.memoryShare >= 8 || group.cpu >= 20 ? 'prominent' : 'normal'
  return {
    label,
    status,
    detail: `${group.count} processes account for ${group.memoryShare.toFixed(1)}% memory and ${group.cpu.toFixed(1)}% sampled CPU.`,
    count: group.count,
    memoryShare: group.memoryShare,
    cpu: group.cpu,
    topProcess: group.topProcess
  }
}

function titleCase(value) {
  return String(value || 'unknown').replace(/[-_]/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase())
}

function peak(points) {
  if (!points?.length) return 'peak --'
  const max = Math.max(...points.map((point) => Number(point.value || 0)))
  return `peak ${max.toFixed(1)}%`
}

function formatBytes(value) {
  const bytesValue = Number(value || 0)
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytesValue
  let unit = 0
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024
    unit++
  }
  return `${size >= 10 ? size.toFixed(0) : size.toFixed(1)} ${units[unit]}`
}

function storageShare(entry) {
  const total = Number(dashboardStore.storage?.total_bytes || 0)
  if (!total) return 0
  return Math.min((Number(entry.size_bytes || 0) / total) * 100, 100)
}

function childStorageShare(child, parent) {
  const total = Number(parent?.size_bytes || 0)
  if (!total) return 0
  return Math.min((Number(child.size_bytes || 0) / total) * 100, 100)
}

function parentLabel(process) {
  if (!process?.ppid) return 'root process'
  const parent = processTreeParents.value.get(process.ppid)
  return parent ? `${parent.name} (${process.ppid})` : `parent PID ${process.ppid}`
}

async function runTerminal() {
  const command = terminalCommand.value.trim()
  if (!command) return
  terminalResult.value = await executeTerminal(command, terminalCwd.value.trim())
  terminalCommand.value = ''
}

function shellQuote(value) {
  return `'${String(value).replaceAll("'", "'\\''")}'`
}

async function lookupManual() {
  const topic = manualQuery.value.trim()
  if (!topic) return
  manualLoading.value = true
  try {
    manualResult.value = await executeTerminal(`MANPAGER=cat man ${shellQuote(topic)} 2>&1 | col -b | head -n 260`, '', { record: false })
  } finally {
    manualLoading.value = false
  }
}

function stageAction(action) {
  terminalCommand.value = action.command
  terminalTab.value = 'terminal'
  activeSection.value = 'terminal'
}

function openActionTarget(action) {
  activeSection.value = action.target
}

function markConfigDirty() {
  dashboardStore.configDirty = true
  dashboardStore.configStatus = 'Unsaved changes'
}

function applyUsageMode() {
  markConfigDirty()
  const mode = dashboardStore.configModes.find((item) => item.name === dashboardStore.config?.usage_mode)
  if (!mode || !dashboardStore.config) return
  dashboardStore.config.polling_interval_seconds = mode.polling_interval_seconds
  dashboardStore.config.retention_hours = mode.retention_hours
  dashboardStore.config.dashboard_refresh_ms = mode.dashboard_refresh_ms
  dashboardStore.config.storage_limits = { ...mode.storage_limits }
  dashboardStore.config.insight_thresholds = { ...mode.insight_thresholds }
}
</script>

<template>
  <AppShell 
    :last-updated="lastUpdated"
    :refreshing="refreshing"
    :system="dashboardStore.system"
    :active-section="activeSection"
    :nav-items="navItems"
    :workspace-meta="workspaceMeta"
    @change-section="changeSection"
  >
    <div v-if="error" class="notice">{{ error }}</div>
    <div v-if="loading" class="notice">Loading NexPerf observability data...</div>

    <WidgetBoardSection
      v-if="activeSection === 'dashboard' && dashboardStore.system"
      v-model:widget-library-open="widgetLibraryOpen"
      :add-widget="addWidget"
      :can-add-widget="canAddWidget"
      :drag-library-widget="dragLibraryWidget"
      :drag-widget="dragWidget"
      :drop-on-board="dropOnBoard"
      :drop-widget="dropWidget"
      :health-tone="healthTone"
      :metric-by-label="metricByLabel"
      :metric-cards="metricCards"
      :metric-segments="metricSegments"
      :observability="observability"
      :process-bar-items="processBarItems"
      :refresh="refresh"
      :remove-widget="removeWidget"
      :time-ranges="timeRanges"
      :timeline-points="timelinePoints"
      :top-insight="topInsight"
      :top-processes="topProcesses"
      :widget-description="widgetDescription"
      :widget-library="widgetLibrary"
      :widget-tone="widgetTone"
      :widgets="widgets"
      :window-label="windowLabel"
      @change-section="changeSection"
    />

    <ProcessesSection
      v-if="activeSection === 'processes'"
      :hierarchy-groups="hierarchyGroups"
      :parent-label="parentLabel"
      @refresh="refresh"
    />

    <StorageSection
      v-if="activeSection === 'storage'"
      :child-storage-share="childStorageShare"
      :format-bytes="formatBytes"
      :storage-share="storageShare"
    />

    <InsightsSection
      v-if="activeSection === 'insights'"
      :actionable-items="actionableItems"
      @open-action-target="openActionTarget"
      @refresh="refresh"
      @stage-action="stageAction"
    />

    <TerminalSection
      v-if="activeSection === 'terminal'"
      v-model:terminal-command="terminalCommand"
      v-model:terminal-cwd="terminalCwd"
      v-model:terminal-tab="terminalTab"
      :terminal-result="terminalResult"
      @run-terminal="runTerminal"
    />

    <ManualSection
      v-if="activeSection === 'manual'"
      v-model:manual-query="manualQuery"
      :command-manual="commandManual"
      :manual-loading="manualLoading"
      :manual-result="manualResult"
      :process-manual="processManual"
      @lookup-manual="lookupManual"
    />

    <PlaceholderSection
      v-if="activeSection === 'network' || activeSection === 'diagnostics'"
      :active-section="activeSection"
    />

    <NexPerfSection v-if="activeSection === 'nexperf'" :format-bytes="formatBytes" />

    <SettingsSection
      v-if="activeSection === 'settings'"
      :config-mode-options="configModeOptions"
      :mode-copy="modeCopy"
      @apply-usage-mode="applyUsageMode"
      @mark-config-dirty="markConfigDirty"
    />
  </AppShell>
</template>
