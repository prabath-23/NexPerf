export const widgetSizes = ['small', 'medium', 'large', 'wide']

export const widgetSizeLabels = {
  small: 'small',
  medium: 'medium',
  large: 'large',
  wide: 'wide'
}

export function normalizeWidgetSize(size) {
  if (size === 'xlarge') return 'wide'
  return widgetSizes.includes(size) ? size : 'small'
}

const defineWidget = (definition) => ({
  supportedSizes: [definition.defaultSize || definition.size || 'small'],
  defaultSize: definition.defaultSize || definition.size || 'small',
  previewComponent: definition.previewComponent || `${definition.kind}-preview`,
  renderComponent: definition.renderComponent || `${definition.kind}-widget`,
  ...definition,
  size: normalizeWidgetSize(definition.size || definition.defaultSize || 'small')
})

export const widgetRegistry = [
  defineWidget({
    id: 'runtime-host',
    title: 'Runtime Summary',
    description: 'Overall health, pressure state, and host context.',
    category: 'health',
    kind: 'host',
    family: 'System',
    size: 'large',
    supportedSizes: ['large'],
    defaultSize: 'large'
  }),
  defineWidget({
    id: 'health-ring',
    title: 'Runtime Health',
    description: 'Compact health score with pressure context.',
    category: 'health',
    kind: 'health',
    family: 'System',
    size: 'small',
    chart: 'ring'
  }),
  defineWidget({
    id: 'active-process-count',
    title: 'Process Load',
    description: 'Active process count with workload attribution.',
    category: 'processes',
    kind: 'processes',
    family: 'Processes',
    size: 'small'
  }),
  defineWidget({
    id: 'runtime-pressure-tile',
    title: 'Pressure State',
    description: 'Small runtime pressure signal with correlated state.',
    category: 'health',
    kind: 'pressure',
    family: 'System',
    size: 'small',
    question: 'Is runtime pressure active right now?'
  }),
  defineWidget({
    id: 'system-summary',
    title: 'Observation Summary',
    description: 'Correlates health, pressure, trend, and workload source.',
    category: 'health',
    kind: 'summary',
    family: 'System',
    size: 'wide',
    supportedSizes: ['wide'],
    defaultSize: 'wide'
  }),

  defineWidget({ id: 'cpu-signal-tile', title: 'CPU Signal', description: 'Compact CPU volatility and trend tile.', category: 'metrics', kind: 'metric', metric: 'CPU', family: 'Pressure', size: 'small', chart: 'line', question: 'Is CPU movement calm or bursty?' }),
  defineWidget({ id: 'memory-allocation-tile', title: 'Allocation Risk', description: 'Small memory pressure and allocation trend.', category: 'metrics', kind: 'metric', metric: 'Memory', family: 'Pressure', size: 'small', chart: 'area', question: 'Is allocation pressure rising?' }),
  defineWidget({ id: 'storage-saturation-tile', title: 'Storage Saturation', description: 'Compact storage saturation level and risk.', category: 'storage', kind: 'metric', metric: 'Disk', family: 'Storage', size: 'small', chart: 'progress', question: 'Is storage pressure elevated?' }),

  defineWidget({ id: 'cpu-volatility', title: 'CPU Volatility', description: 'Classifies CPU movement as stable, bursty, or sustained.', category: 'metrics', kind: 'metric', metric: 'CPU', family: 'Pressure', size: 'medium', chart: 'line', question: 'Has CPU behavior been unstable recently?' }),
  defineWidget({ id: 'cpu-bursts', title: 'Workload Bursts', description: 'Shows burst distribution and peak workload behavior.', category: 'metrics', kind: 'metric', metric: 'CPU', family: 'Workload', size: 'medium', chart: 'bar', question: 'Are CPU spikes clustered or isolated?' }),
  defineWidget({ id: 'cpu-timeline', title: 'CPU Timeline', description: 'Temporal CPU pressure with peak and stabilization context.', category: 'history', kind: 'timeline', metric: 'CPU', family: 'History', size: 'large', chart: 'line', question: 'Has CPU pressure stabilized or escalated?' }),

  defineWidget({ id: 'memory-pressure', title: 'Memory Pressure', description: 'Detects sustained memory pressure and leading contributors.', category: 'metrics', kind: 'metric', metric: 'Memory', family: 'Pressure', size: 'medium', chart: 'area', question: 'Is memory pressure sustained or rising?' }),
  defineWidget({ id: 'memory-history', title: 'Memory Timeline', description: 'Memory pressure history with stability classification.', category: 'history', kind: 'timeline', metric: 'Memory', family: 'History', size: 'large', chart: 'area', question: 'Is memory pressure changing over time?' }),

  defineWidget({ id: 'disk-growth', title: 'Storage Growth', description: 'Flags storage growth velocity and capacity risk.', category: 'storage', kind: 'metric', metric: 'Disk', family: 'Storage', size: 'medium', chart: 'progress', question: 'Is storage growing abnormally?' }),
  defineWidget({ id: 'disk-history-bars', title: 'Disk Trend', description: 'Disk utilization history with growth classification.', category: 'history', kind: 'timeline', metric: 'Disk', family: 'History', size: 'large', chart: 'bar', question: 'Is disk usage accelerating?' }),

  defineWidget({ id: 'process-stack', title: 'Workload Attribution', description: 'Top process groups causing memory and CPU pressure.', category: 'processes', kind: 'processes', family: 'Workload', size: 'wide', supportedSizes: ['wide'], defaultSize: 'wide', question: 'Which applications are causing pressure?' }),
  defineWidget({ id: 'top-memory-processes', title: 'Top Memory Contributors', description: 'Largest contributors with category attribution.', category: 'processes', kind: 'processes', family: 'Workload', size: 'large', question: 'Which workload owns memory pressure?' }),
  defineWidget({ id: 'browser-pressure-tile', title: 'Browser Pressure', description: 'Small browser renderer pressure indicator.', category: 'processes', kind: 'workload', workload: 'browser', family: 'Workload', size: 'small', question: 'Is browser workload prominent?' }),
  defineWidget({ id: 'developer-load-tile', title: 'Developer Load', description: 'Small IDE and coding-agent workload indicator.', category: 'processes', kind: 'workload', workload: 'development', family: 'Workload', size: 'small', question: 'Is development tooling prominent?' }),
  defineWidget({ id: 'browser-process-group', title: 'Browser Workload', description: 'Browser, WebKit, and renderer process pressure.', category: 'processes', kind: 'workload', workload: 'browser', family: 'Workload', size: 'medium', question: 'Are browser processes driving pressure?' }),
  defineWidget({ id: 'ide-activity', title: 'IDE Activity', description: 'Editor, extension host, and coding agent workload.', category: 'processes', kind: 'workload', workload: 'development', family: 'Workload', size: 'medium', question: 'Is development tooling driving pressure?' }),
  defineWidget({ id: 'background-services', title: 'Background Services', description: 'Service and helper process footprint.', category: 'processes', kind: 'workload', workload: 'service', family: 'Workload', size: 'medium', question: 'Are background services prominent?' }),
  defineWidget({ id: 'cpu-spike-processes', title: 'CPU Spike Attribution', description: 'Processes most likely to explain CPU spikes.', category: 'processes', kind: 'processes', family: 'Workload', size: 'medium', question: 'Which workload explains CPU bursts?' }),

  defineWidget({ id: 'next-action', title: 'Reasoned Action', description: 'Correlates metrics and workload into one next action.', category: 'insights', kind: 'actionable', family: 'Insights', size: 'wide', supportedSizes: ['wide'], defaultSize: 'wide', question: 'What should I inspect next and why?' }),
  defineWidget({ id: 'action-hint-tile', title: 'Action Hint', description: 'Small correlated next-action cue.', category: 'insights', kind: 'actionable', family: 'Insights', size: 'small', question: 'What is the next signal to inspect?' }),
  defineWidget({ id: 'pressure-signal-tile', title: 'Pressure Signal', description: 'Compact sustained-pressure classification.', category: 'insights', kind: 'pressure', family: 'Insights', size: 'small', question: 'Is pressure localized or correlated?' }),
  defineWidget({ id: 'active-insight', title: 'Active Insight', description: 'Highest-priority observation with cited signals.', category: 'insights', kind: 'actionable', family: 'Insights', size: 'medium', question: 'What changed enough to matter?' }),
  defineWidget({ id: 'insight-severity', title: 'Sustained Pressure', description: 'Detects repeated pressure across CPU, memory, and disk.', category: 'insights', kind: 'pressure', family: 'Insights', size: 'medium', question: 'Is pressure sustained or temporary?' }),
  defineWidget({ id: 'recent-insight-timeline', title: 'Insight Timeline', description: 'Recent insight activity over time.', category: 'insights', kind: 'timeline', metric: 'CPU', family: 'Insights', size: 'wide', chart: 'area', question: 'Are issues increasing over time?' }),

  defineWidget({ id: 'observation-window', title: 'Observation Window', description: 'Sampling range, coverage, and active trend window.', category: 'history', kind: 'window', family: 'History', size: 'wide', supportedSizes: ['wide'], defaultSize: 'wide' }),
  defineWidget({ id: 'multi-metric-timeline', title: 'Multi-Metric Timeline', description: 'Correlated CPU, memory, and disk trend strip.', category: 'history', kind: 'timeline', metric: 'Memory', family: 'History', size: 'wide', chart: 'area', question: 'Which signals moved together?' }),
  defineWidget({ id: 'network-signal-tile', title: 'Network Signal', description: 'Small network collector readiness indicator.', category: 'insights', kind: 'network', family: 'Network', size: 'small', question: 'Is network telemetry ready?' }),
  defineWidget({ id: 'network-activity', title: 'Network Activity', description: 'Prepared for active connections and traffic anomalies.', category: 'insights', kind: 'network', family: 'Network', size: 'medium', question: 'Is unexpected traffic occurring?' })
]

export const defaultWidgetIds = [
  'runtime-host',
  'system-summary',
  'cpu-volatility',
  'memory-pressure',
  'process-stack',
  'browser-process-group',
  'disk-growth',
  'insight-severity',
  'next-action',
  'cpu-timeline',
  'observation-window'
]

export const categoryLabels = {
  health: 'System',
  metrics: 'Metrics',
  storage: 'Storage',
  processes: 'Processes',
  insights: 'Insights',
  history: 'History'
}

export const widgetTemplateMap = new Map(widgetRegistry.map((widget) => [widget.id, widget]))

export const widgetLibraryGroups = Object.entries(categoryLabels)
  .map(([category, group]) => ({
    group,
    items: widgetRegistry.filter((widget) => widget.category === category)
  }))
  .filter((group) => group.items.length)

export function resolveWidgetTemplate(id) {
  return widgetTemplateMap.get(id)
}
