export const appRoutes = [
  { id: 'dashboard', label: 'Overview', icon: 'overview', nav: true },
  { id: 'processes', label: 'Processes', icon: 'processes', nav: true },
  { id: 'storage', label: 'Storage', icon: 'storage', nav: true },
  { id: 'insights', label: 'Actionables', icon: 'actionables', nav: true },
  { id: 'terminal', label: 'Terminal', icon: 'terminal', nav: true },
  { id: 'manual', label: 'Manual', icon: 'manual', nav: true },
  { id: 'network', label: 'Network', icon: 'network', nav: true, soon: true },
  { id: 'diagnostics', label: 'Diagnostics', icon: 'diagnostics', nav: true, soon: true },
  { id: 'nexperf', label: 'NexPerf', icon: 'nexperf', nav: true },
  { id: 'settings', label: 'Settings', icon: 'settings', nav: true }
]

export const workspaceMeta = {
  dashboard: {
    title: 'Overview',
    subtitle: 'Live system telemetry and health'
  },
  overview: {
    title: 'Overview',
    subtitle: 'Live system telemetry and health'
  },
  processes: {
    title: 'Processes',
    subtitle: 'Runtime workload and memory activity'
  },
  storage: {
    title: 'Storage',
    subtitle: 'Capacity, allocation, and usage history'
  },
  insights: {
    title: 'Insights',
    subtitle: 'System analysis and workload recommendations'
  },
  history: {
    title: 'History',
    subtitle: 'Historical telemetry and runtime trends'
  },
  terminal: {
    title: 'Terminal',
    subtitle: 'Local commands and execution history'
  },
  manual: {
    title: 'Manual',
    subtitle: 'Command reference and operational guidance'
  },
  network: {
    title: 'Network',
    subtitle: 'Connectivity signals and interface activity'
  },
  diagnostics: {
    title: 'Diagnostics',
    subtitle: 'Runtime checks and system signals'
  },
  nexperf: {
    title: 'NexPerf',
    subtitle: 'Agent status and collector health'
  },
  settings: {
    title: 'Settings',
    subtitle: 'Sampling, retention, and workspace preferences'
  }
}

export const routeAliases = {
  history: 'dashboard',
  overview: 'dashboard'
}

export const routeSections = new Set(appRoutes.map((route) => route.id))

export const navItems = appRoutes
  .filter((route) => route.nav)
  .map(({ id, label, icon, soon }) => ({ id, label, icon, soon }))

export function metaForSection(section) {
  return workspaceMeta[section] || workspaceMeta.dashboard
}

export function sectionFromURL(location = window.location) {
  const params = new URLSearchParams(location.search)
  const raw = params.get('section') || location.hash.replace(/^#\/?/, '') || 'dashboard'
  const normalized = routeAliases[raw] || raw
  return routeSections.has(normalized) ? normalized : 'dashboard'
}

export function updateSectionURL(section, history = window.history, location = window.location) {
  const params = new URLSearchParams(location.search)
  params.set('section', routeSections.has(section) ? section : 'dashboard')
  params.delete('rev')

  const next = `${location.pathname}?${params.toString()}`
  if (next !== `${location.pathname}${location.search}`) {
    history.pushState({ section }, '', next)
  }
}
