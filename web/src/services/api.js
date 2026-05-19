const DASHBOARD_WIDGETS_STORAGE_KEY = 'nexperf-dashboard-widgets'

const json = async (path) => {
  const response = await fetch(path)
  if (!response.ok) {
    throw new Error(`${path} returned ${response.status}`)
  }
  return response.json()
}

function readLocalDashboardWidgets() {
  if (typeof window === 'undefined') return []
  try {
    const value = window.localStorage.getItem(DASHBOARD_WIDGETS_STORAGE_KEY)
    if (!value) return []
    const widgets = JSON.parse(value)
    return Array.isArray(widgets) ? widgets : []
  } catch {
    return []
  }
}

export function cacheDashboardWidgets(widgets) {
  if (typeof window === 'undefined') return widgets
  try {
    window.localStorage.setItem(DASHBOARD_WIDGETS_STORAGE_KEY, JSON.stringify(widgets))
  } catch {
    // Local persistence is best-effort; API persistence still runs below.
  }
  return widgets
}

export function getSystem() {
  return json('/api/system')
}

export function getProcesses(limit = 12) {
  return json(`/api/processes/top?limit=${limit}`)
}

export function getInsights() {
  return json('/api/insights')
}

export function getHealthScore() {
  return json('/api/health-score')
}

export function getHistory(metric, range = '5m') {
  return json(`/api/history/${metric}?range=${range}`)
}

export function getProcessTree(limit = 80) {
  return json(`/api/processes/tree?limit=${limit}`)
}

export function getStorageSummary(path = '') {
  const suffix = path ? `?path=${encodeURIComponent(path)}` : ''
  return json(`/api/storage/summary${suffix}`)
}

export function getConfig() {
  return json('/api/config')
}

export function getConfigModes() {
  return json('/api/config/modes')
}

export async function getDashboardWidgets() {
  const localWidgets = readLocalDashboardWidgets()
  try {
    const widgets = await json('/api/dashboard/widgets')
    if (Array.isArray(widgets) && widgets.length) {
      cacheDashboardWidgets(widgets)
      return widgets
    }
    return localWidgets
  } catch {
    return localWidgets
  }
}

export async function saveDashboardWidgets(widgets) {
  const cached = cacheDashboardWidgets(widgets)
  try {
    const response = await fetch('/api/dashboard/widgets', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(widgets)
    })
    if (!response.ok) throw new Error(await response.text())
    const saved = await response.json()
    cacheDashboardWidgets(saved)
    return saved
  } catch {
    return cached
  }
}

export async function saveConfig(config) {
  const response = await fetch('/api/config', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config)
  })
  if (!response.ok) throw new Error(await response.text())
  return response.json()
}

export function getNexPerfStats() {
  return json('/api/nexperf')
}

export async function runTerminalCommand(command, cwd = '') {
  const response = await fetch('/api/terminal/exec', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ command, cwd })
  })
  if (!response.ok) throw new Error(await response.text())
  return response.json()
}
