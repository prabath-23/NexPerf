<script setup>
import BarChart from '../../components/charts/BarChart.vue'
import DonutChart from '../../components/charts/DonutChart.vue'
import ProgressBar from '../../components/charts/ProgressBar.vue'
import RingGauge from '../../components/charts/RingGauge.vue'
import SparklineChart from '../../components/charts/SparklineChart.vue'
import WidgetCard from '../../components/ui/WidgetCard.vue'
import { dashboardStore } from '../../stores/dashboardStore'

defineProps({
  addWidget: { type: Function, required: true },
  canAddWidget: { type: Function, required: true },
  dragLibraryWidget: { type: Function, required: true },
  dragWidget: { type: Function, required: true },
  dropOnBoard: { type: Function, required: true },
  dropWidget: { type: Function, required: true },
  healthTone: { type: Function, required: true },
  metricByLabel: { type: Function, required: true },
  metricCards: { type: Array, default: () => [] },
  metricSegments: { type: Function, required: true },
  observability: { type: Object, default: () => ({}) },
  processBarItems: { type: Function, required: true },
  refresh: { type: Function, required: true },
  removeWidget: { type: Function, required: true },
  timeRanges: { type: Array, default: () => [] },
  timelinePoints: { type: Function, required: true },
  topInsight: { type: Object, default: null },
  topProcesses: { type: Array, default: () => [] },
  widgetDescription: { type: Function, required: true },
  widgetLibrary: { type: Array, default: () => [] },
  widgetLibraryOpen: { type: Boolean, default: false },
  widgetTone: { type: Function, required: true },
  widgets: { type: Array, default: () => [] },
  windowLabel: { type: Function, required: true }
})

const emit = defineEmits(['change-section', 'update:widget-library-open'])

const telemetryLabels = {
  CPU: 'CPU Volatility',
  Memory: 'Allocation Pressure',
  Disk: 'Storage Saturation'
}

function runtimeStateLabel(healthScore, observability) {
  const pressure = observability?.pressure?.state
  const actionTitle = observability?.action?.title || ''
  if (pressure === 'correlated') return 'Pressure stabilizing'
  if (actionTitle.toLowerCase().includes('memory')) return 'Memory recovery detected'
  if (actionTitle.toLowerCase().includes('disk')) return 'Storage contention rising'
  if (Number(healthScore?.value || 0) >= 80) return 'Runtime stable'
  if (Number(healthScore?.value || 0) >= 75) return 'Workload balanced'
  return 'Pressure stabilizing'
}

function runtimeStateSummary(healthScore, observability) {
  const label = runtimeStateLabel(healthScore, observability)
  if (label.includes('Memory')) return 'Memory pressure is easing, but allocation remains the leading signal.'
  if (label.includes('Storage')) return 'Storage saturation is elevated and may affect local workload flow.'
  if (label.includes('balanced')) return 'Runtime signals are balanced across CPU, memory, and storage.'
  if (label.includes('stable')) return 'System telemetry is calm with comfortable operating headroom.'
  return 'Runtime pressure is active; NexPerf is tracking stabilization across key signals.'
}

function summaryLeadMetric(cards) {
  return cards
    .map((card) => ({
      ...card,
      numericValue: Number(String(card?.value || '0').replace('%', ''))
    }))
    .filter((card) => Number.isFinite(card.numericValue))
    .sort((a, b) => b.numericValue - a.numericValue)[0]
}

function summaryLeadUsage(cards) {
  const lead = summaryLeadMetric(cards)
  return lead ? `${Math.round(lead.numericValue)}%` : '--'
}

function summaryLeadLabel(cards) {
  const lead = summaryLeadMetric(cards)
  if (!lead) return 'leading signal'
  if (lead.label === 'Disk') return 'storage saturation'
  if (lead.label === 'Memory') return 'allocation pressure'
  if (lead.label === 'CPU') return 'cpu volatility'
  return `${String(lead.label).toLowerCase()} signal`
}
</script>

<template>
<section class="widget-board">
      <div class="widget-board-toolbar margintop" :class="{ open: widgetLibraryOpen }">
        <div class="widget-board-heading">
          <div>
            <h2>Telemetry Board</h2>
            <span>Live system widgets for health, workload, and history.</span>
          </div>
          <button type="button" @click="emit('update:widget-library-open', !widgetLibraryOpen)">
            {{ widgetLibraryOpen ? 'Close Library' : 'Widget Library' }}
          </button>
        </div>

        <aside v-if="widgetLibraryOpen" class="widget-gallery-panel">
          <section v-for="group in widgetLibrary" :key="group.group" class="widget-gallery-group">
            <header>
              <h3>{{ group.group }}</h3>
              <span>{{ group.items.length }} widget styles</span>
            </header>
            <div class="widget-gallery-strip">
              <article
                v-for="item in group.items"
                :key="item.id"
                class="widget-gallery-item"
                :class="[`preview-${item.size}`, `preview-kind-${item.kind}`, item.chart ? `preview-chart-${item.chart}` : '', { disabled: !canAddWidget(item) }]"
                @dblclick="addWidget(item)"
              >
                <div
                  class="gallery-preview-card"
                  :draggable="canAddWidget(item)"
                  @dragstart.stop="canAddWidget(item) && dragLibraryWidget(item, $event)"
                >
                  <span>{{ item.title }}</span>
                  <RingGauge
                    v-if="item.chart === 'ring' || item.kind === 'health'"
                    :value="item.kind === 'health' ? dashboardStore.healthScore?.value : Number(String(metricByLabel(item.metric)?.value || 0).replace('%', ''))"
                    :tone="widgetTone(item)"
                    :label="item.kind === 'health' ? 'score' : item.metric"
                    variant="double"
                  />
                  <DonutChart
                    v-else-if="item.chart === 'donut'"
                    :segments="metricSegments(item.metric)"
                    :value="metricByLabel(item.metric)?.value"
                    :label="item.metric"
                  />
                  <BarChart
                    v-else-if="item.chart === 'bar'"
                    :items="item.id === 'cpu-bursts'
                      ? [18, 42, 25, 76, 34, 92, 46, 68].map((value, index) => ({ label: `${index + 1}`, value, display: `${value}%` }))
                      : item.kind === 'processes'
                        ? processBarItems().slice(0, 3)
                        : (metricByLabel(item.metric)?.history || []).slice(-8).map((point, index) => ({ label: `${index + 1}`, value: point.value, display: `${Number(point.value || 0).toFixed(0)}%` }))"
                    :max="item.id === 'cpu-bursts' ? 100 : 0"
                    variant="compact"
                    :tone="widgetTone(item)"
                  />
                  <ProgressBar
                    v-else-if="item.chart === 'progress'"
                    :value="Number(String(metricByLabel(item.metric)?.value || 0).replace('%', ''))"
                    :tone="widgetTone(item)"
                    :label="item.metric"
                    compact
                  />
                  <div
                    v-else-if="item.kind === 'metric' && item.chart !== 'value'"
                    class="preview-metric-stream"
                  >
                    <strong>{{ metricByLabel(item.metric)?.value || '0.0%' }}</strong>
                    <span>{{ observability.metrics?.[item.metric]?.state || widgetTone(item) }}</span>
                    <SparklineChart
                      :points="(metricByLabel(item.metric)?.history || []).slice(-18)"
                      :tone="widgetTone(item)"
                      :variant="item.chart || 'line'"
                      :interactive="false"
                      compact
                      :height="46"
                    />
                  </div>
                  <SparklineChart
                    v-else-if="item.kind === 'timeline'"
                    :points="timelinePoints(item.metric).slice(-24)"
                    :tone="widgetTone(item)"
                    :variant="item.chart || 'area'"
                    :interactive="false"
                    :height="34"
                  />
                  <div
                    v-else-if="item.kind === 'host'"
                    class="preview-host"
                    :style="{ '--score-value': Number(dashboardStore.healthScore?.value || 0) }"
                  >
                    <strong>{{ dashboardStore.healthScore?.value || '--' }}</strong>
                    <span>{{ dashboardStore.system.os }} / {{ dashboardStore.system.arch }}</span>
                  </div>
                  <div v-else-if="item.kind === 'summary'" class="preview-summary">
                    <strong>{{ dashboardStore.healthScore?.value || '--' }}</strong>
                    <span>Health score</span>
                    <i v-for="card in metricCards" :key="`preview-summary-${card.label}`">
                      <b :class="card.severity" :style="{ width: card.value }"></b>
                    </i>
                  </div>
              <div v-else-if="item.kind === 'processes'" class="preview-processes">
                <strong>{{ dashboardStore.processes.length }}</strong>
                <span>tracked processes</span>
                <i v-for="process in processBarItems().slice(0, 3)" :key="`preview-process-${process.label}`">
                  <b :style="{ width: `${Math.max(14, process.value)}%` }"></b>
                </i>
                  </div>
                  <div v-else-if="item.kind === 'workload'" class="preview-processes">
                    <strong>{{ observability.workloads?.[item.workload]?.count || 0 }}</strong>
                    <span>{{ observability.workloads?.[item.workload]?.status || 'collecting' }}</span>
                    <i>
                      <b :style="{ width: `${Math.max(10, observability.workloads?.[item.workload]?.memoryShare || 0)}%` }"></b>
                    </i>
                  </div>
                  <div v-else-if="item.kind === 'actionable'" class="preview-insight">
                    <em :class="widgetTone(item)"></em>
                    <strong>{{ observability.action?.title || topInsight?.title || 'No pressure' }}</strong>
                  </div>
                  <div v-else-if="item.kind === 'pressure'" class="preview-insight">
                    <em :class="observability.pressure?.state"></em>
                    <strong>{{ observability.pressure?.label || 'Pressure normal' }}</strong>
                  </div>
                  <div v-else-if="item.kind === 'network'" class="preview-window">
                    <strong>Prepared</strong>
                    <span>collector pending</span>
                  </div>
                  <div v-else-if="item.kind === 'window'" class="preview-window">
                    <strong>{{ windowLabel() }}</strong>
                    <span>{{ dashboardStore.history.cpu.length }} samples</span>
                  </div>
                  <strong v-else>{{ item.title }}</strong>
                </div>
                <div class="gallery-copy">
                  <strong>{{ item.title }}</strong>
                  <span>{{ item.size }} · {{ item.family }}</span>
                </div>
                <small class="gallery-status">{{ canAddWidget(item) ? item.description : 'Added' }}</small>
              </article>
            </div>
          </section>
        </aside>
      </div>

      <section class="widget-grid" @dragover.prevent @drop.self="dropOnBoard">
        <WidgetCard
          v-for="widget in widgets"
          :key="widget.id"
          :size="widget.size"
          :tone="widgetTone(widget)"
          :title="widget.title"
          :subtitle="widget.size === 'small' || ['processes', 'workload', 'pressure', 'actionable'].includes(widget.kind) ? '' : widget.description"
          :description="widgetDescription(widget)"
          :class="[`widget-kind-${widget.kind}`, widget.chart ? `widget-chart-${widget.chart}` : '']"
          draggable
          @dragstart="dragWidget(widget.id, $event)"
          @drop="dropWidget(widget.id)"
        >
          <template v-if="widgetLibraryOpen" #actions>
              <button type="button" class="remove" title="Remove widget" @click.stop="removeWidget(widget.id)">×</button>
          </template>

          <div v-if="widget.kind === 'host'" class="host-widget">
            <div>
              <h2>{{ dashboardStore.system.hostname || 'Local machine' }}</h2>
              <p>{{ dashboardStore.system.os }} / {{ dashboardStore.system.arch }}</p>
            </div>
            <div
              class="host-score"
              :class="healthTone(dashboardStore.healthScore)"
              :style="{ '--score-value': Number(dashboardStore.healthScore?.value || 0) }"
            >
              <svg class="telemetry-mark" viewBox="0 0 132 132" aria-hidden="true">
                <defs>
                  <linearGradient id="runtime-ring-gradient" x1="14%" x2="86%" y1="12%" y2="88%">
                    <stop offset="0%" stop-color="#0a84ff" />
                    <stop offset="48%" stop-color="#00c7be" />
                    <stop offset="100%" stop-color="#8b5cf6" />
                  </linearGradient>
                  <radialGradient id="runtime-donut-field" cx="48%" cy="42%" r="58%">
                    <stop offset="0%" stop-color="#e8fbff" stop-opacity="0.16" />
                    <stop offset="45%" stop-color="#0a84ff" stop-opacity="0.2" />
                    <stop offset="76%" stop-color="#00c7be" stop-opacity="0.18" />
                    <stop offset="100%" stop-color="#8b5cf6" stop-opacity="0.12" />
                  </radialGradient>
                  <filter id="runtime-ring-glow" x="-35%" y="-35%" width="170%" height="170%">
                    <feGaussianBlur stdDeviation="2.4" />
                  </filter>
                </defs>
                <circle class="mark-aura" cx="66" cy="66" r="47" />
                <circle class="mark-donut-field" cx="66" cy="66" r="42" />
                <circle class="mark-disc" cx="66" cy="66" r="35" />
                <circle class="mark-track" cx="66" cy="66" r="48" pathLength="100" />
                <circle class="mark-progress" cx="66" cy="66" r="48" pathLength="100" />
                <circle class="mark-flow flow-a" cx="66" cy="66" r="45" />
                <circle class="mark-flow flow-b" cx="66" cy="66" r="42" />
                <circle class="mark-energy" cx="66" cy="66" r="40" />
                <circle class="mark-orbit" cx="66" cy="66" r="56" />
                <g class="mark-segments">
                  <path d="M66 12a54 54 0 0 1 18 3" />
                  <path d="M120 66a54 54 0 0 1-3 18" />
                  <path d="M66 120a54 54 0 0 1-18-3" />
                  <path d="M12 66a54 54 0 0 1 3-18" />
                </g>
                <g class="mark-trails">
                  <path d="M39 66h12" />
                  <path d="M81 66h12" />
                </g>
              </svg>
              <strong>{{ dashboardStore.healthScore?.value || '--' }}</strong>
              <span>score</span>
            </div>
            <!-- <div class="host-readouts">
              <div v-for="item in metricCards" :key="`host-${item.label}`">
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </div>
            </div> -->
            <div class="host-note" v-if="dashboardStore.healthScore">
              <strong>{{ runtimeStateLabel(dashboardStore.healthScore, observability) }}</strong>
              <span>{{ runtimeStateSummary(dashboardStore.healthScore, observability) }}</span>
            </div>
          </div>

          <div v-else-if="widget.kind === 'metric'" class="metric-widget" :class="[metricByLabel(widget.metric)?.severity, `chart-${widget.chart || 'line'}`]">
            <div class="metric-topline">
              <em class="metric-state">{{ observability.metrics?.[widget.metric]?.state || metricByLabel(widget.metric)?.severity || 'normal' }}</em>
              <span class="telemetry-dot"></span>
            </div>
            <strong class="metric-value"><span>{{ metricByLabel(widget.metric)?.value?.replace('%', '') }}</span><small>%</small></strong>
            <p>{{ observability.metrics?.[widget.metric]?.interpretation || metricByLabel(widget.metric)?.detail }}</p>
            <div class="metric-context">
              <span :class="metricByLabel(widget.metric)?.trend.direction">
                {{ metricByLabel(widget.metric)?.trend.direction === 'up' ? '↑' : metricByLabel(widget.metric)?.trend.direction === 'down' ? '↓' : '→' }}
                {{ metricByLabel(widget.metric)?.trend.label }}
              </span>
              <span>{{ observability.metrics?.[widget.metric]?.peak || metricByLabel(widget.metric)?.peak }}</span>
            </div>
            <div class="telemetry-chip-row">
              <span>{{ observability.metrics?.[widget.metric]?.volatilityLabel }}</span>
              <span>{{ observability.metrics?.[widget.metric]?.confidence }}</span>
            </div>
            <RingGauge
              v-if="widget.chart === 'ring'"
              :value="Number(String(metricByLabel(widget.metric)?.value || 0).replace('%', ''))"
              :tone="metricByLabel(widget.metric)?.severity"
              :label="widget.metric"
              variant="double"
            />
            <DonutChart
              v-else-if="widget.chart === 'donut'"
              :segments="metricSegments(widget.metric)"
              :value="metricByLabel(widget.metric)?.value"
              :label="widget.metric"
            />
            <BarChart
              v-else-if="widget.chart === 'bar'"
              :items="(metricByLabel(widget.metric)?.history || []).slice(-12).map((point, index) => ({ label: `${index + 1}`, value: point.value, display: `${Number(point.value || 0).toFixed(0)}%` }))"
              :tone="metricByLabel(widget.metric)?.severity"
              variant="compact"
            />
            <ProgressBar
              v-else-if="widget.chart === 'progress'"
              :value="Number(String(metricByLabel(widget.metric)?.value || 0).replace('%', ''))"
              :tone="metricByLabel(widget.metric)?.severity"
              :label="widget.metric"
            />
            <SparklineChart
              v-else-if="widget.chart !== 'value'"
              :points="metricByLabel(widget.metric)?.history"
              :tone="metricByLabel(widget.metric)?.severity"
              :height="widget.size === 'small' ? 64 : 68"
              :variant="widget.chart || 'line'"
              :interactive="true"
              compact
            >
              <template #tooltip="{ point, stats }">
                <div class="chart-tooltip rich telemetry-tooltip compact metric-tooltip">
                  <div class="tooltip-top">
                    <span><i></i>{{ widget.title }}</span>
                    <em>{{ point.time }}</em>
                  </div>
                  <div class="tooltip-value-row">
                    <strong>{{ point.value.toFixed(1) }}%</strong>
                    <span>{{ stats.state }}</span>
                  </div>
                  <div class="tooltip-strip">
                    <span>Avg {{ stats.avg.toFixed(1) }}%</span>
                    <span>Peak {{ stats.peak.toFixed(1) }}%</span>
                    <span :class="stats.direction === 'up' ? 'positive' : stats.direction === 'down' ? 'negative' : ''">{{ stats.delta > 0 ? '+' : '' }}{{ stats.delta.toFixed(1) }} pts</span>
                  </div>
                </div>
              </template>
            </SparklineChart>
          </div>

          <div v-else-if="widget.kind === 'processes'" class="process-widget">
            <div class="widget-head process-headline">
              <strong>{{ dashboardStore.processes.length }}</strong>
              <span>tracked processes</span>
              <em v-if="observability.categories?.[0]">
                {{ observability.categories[0].label }} leads by memory
              </em>
            </div>
            <BarChart :items="processBarItems().slice(0, widget.size === 'wide' ? 3 : widget.size === 'medium' ? 3 : 4)" variant="horizontal" tone="process" />
            <p class="widget-detail">
              {{ observability.categories?.[0]
                ? `${observability.categories[0].label} leads with ${observability.categories[0].memoryShare.toFixed(1)}% memory share; top process ${observability.categories[0].topProcess}.`
                : 'Collecting process attribution.' }}
            </p>
          </div>

          <div v-else-if="widget.kind === 'workload'" class="workload-widget">
            <span class="widget-kicker">{{ observability.workloads?.[widget.workload]?.status || 'collecting' }}</span>
            <strong>{{ observability.workloads?.[widget.workload]?.label || widget.title }}</strong>
            <p>{{ observability.workloads?.[widget.workload]?.detail || widget.description }}</p>
            <div class="dominance-strip" :style="{ '--dominance': `${Math.min(100, observability.workloads?.[widget.workload]?.memoryShare || 0)}%` }"></div>
            <div class="workload-readouts">
              <span>{{ observability.workloads?.[widget.workload]?.count || 0 }} processes</span>
              <span>{{ Number(observability.workloads?.[widget.workload]?.memoryShare || 0).toFixed(1) }}% memory</span>
              <span>{{ Number(observability.workloads?.[widget.workload]?.cpu || 0).toFixed(1) }}% CPU</span>
            </div>
            <small>{{ observability.workloads?.[widget.workload]?.topProcess || 'No dominant process' }}</small>
          </div>

          <div v-else-if="widget.kind === 'health'" class="health-widget">
            <div class="widget-head compact">
              <strong>{{ dashboardStore.healthScore?.value || '--' }}</strong>
            </div>
            <RingGauge :value="dashboardStore.healthScore?.value" :tone="healthTone(dashboardStore.healthScore)" label="score" variant="double" />
            <p class="widget-detail">{{ dashboardStore.healthScore?.summary || 'Collecting host state.' }}</p>
            <div class="health-bars">
              <div v-for="card in metricCards" :key="`health-${card.label}`">
                <span>{{ card.label }}</span>
                <i><b :class="card.severity" :style="{ width: card.value }"></b></i>
                <strong>{{ card.value }}</strong>
              </div>
            </div>
          </div>

          <div v-else-if="widget.kind === 'actionable'" class="action-widget">
            <span class="widget-kicker">Reasoned action</span>
            <strong>{{ observability.action?.title || topInsight?.title || 'No active pressure' }}</strong>
            <p>{{ observability.action?.message || topInsight?.recommendation || 'System state is steady. Keep the board live for changes.' }}</p>
            <div class="signal-list">
              <span v-for="signal in (observability.action?.citedSignals || []).slice(0, 3)" :key="signal">{{ signal }}</span>
            </div>
            <button type="button" @click="emit('change-section', topInsight ? 'insights' : 'processes')">
              {{ topInsight ? 'Review actionables' : 'Review processes' }}
            </button>
          </div>

          <div v-else-if="widget.kind === 'pressure'" class="pressure-widget">
            <span class="widget-kicker">{{ observability.pressure?.state || 'stable' }}</span>
            <strong>{{ observability.pressure?.label || 'Pressure normal' }}</strong>
            <p>{{ observability.pressure?.detail || 'No sustained pressure across the active observation window.' }}</p>
            <div class="pressure-grid">
              <span v-for="card in metricCards" :key="`pressure-${card.label}`">
                <b>{{ telemetryLabels[card.label] || card.label }}</b>
                <em>{{ observability.metrics?.[card.label]?.state || card.severity }}</em>
              </span>
            </div>
          </div>

          <div v-else-if="widget.kind === 'window'" class="window-widget">
            <div class="widget-head compact">
              <span>Observation window</span>
              <strong>{{ windowLabel() }}</strong>
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
            <div class="window-stats">
              <span>{{ dashboardStore.history.cpu.length }} CPU samples</span>
              <span>{{ dashboardStore.history.memory.length }} memory samples</span>
              <span>{{ dashboardStore.history.disk.length }} disk samples</span>
            </div>
          </div>

          <div v-else-if="widget.kind === 'timeline'" class="timeline-widget">
            <SparklineChart
              :points="timelinePoints(widget.metric)"
              :tone="widgetTone(widget)"
              :variant="widget.chart || 'area'"
              :interactive="true"
              :height="widget.size === 'wide' ? 132 : 122"
            >
              <template #tooltip="{ point, stats }">
                <div class="chart-tooltip rich telemetry-tooltip compact timeline-tooltip">
                  <div class="tooltip-top">
                    <span><i></i>{{ widget.title }}</span>
                    <em>{{ point.time }}</em>
                  </div>
                  <div class="tooltip-value-row">
                    <strong>{{ point.value.toFixed(1) }}%</strong>
                    <span>{{ stats.state }}</span>
                  </div>
                  <div class="tooltip-grid">
                    <span>Avg window</span><strong>{{ stats.avg.toFixed(1) }}%</strong>
                    <span>Peak</span><strong>{{ stats.peak.toFixed(1) }}%</strong>
                    <span>Volatility</span><strong>{{ stats.volatility.toFixed(1) }}</strong>
                    <span>Delta</span><strong :class="stats.direction === 'up' ? 'positive' : stats.direction === 'down' ? 'negative' : ''">{{ stats.delta > 0 ? '+' : '' }}{{ stats.delta.toFixed(1) }} pts</strong>
                    <span>Anomalies</span><strong>{{ stats.anomalies }}</strong>
                    <span>Sample</span><strong>{{ point.index }} / {{ point.total }}</strong>
                  </div>
                </div>
              </template>
            </SparklineChart>
            <div class="timeline-meta">
              <span>{{ timelinePoints(widget.metric).length }} samples</span>
              <span>{{ windowLabel() }}</span>
              <span>{{ observability.metrics?.[widget.metric]?.state }}</span>
            </div>
          </div>

          <div v-else-if="widget.kind === 'summary'" class="summary-widget">
            <div class="widget-head">
              <strong>{{ dashboardStore.healthScore?.value || '--' }}</strong>
              <span>{{ observability.pressure?.state || 'stable' }}</span>
            </div>
            <div class="summary-mosaic">
              <DonutChart
                :segments="metricCards.map((card) => ({ label: card.label, value: Number(String(card.value).replace('%', '')), color: card.label === 'CPU' ? 'var(--blue)' : card.label === 'Memory' ? 'var(--violet)' : 'var(--teal)' }))"
                :total="300"
                :value="summaryLeadUsage(metricCards)"
                :label="summaryLeadLabel(metricCards)"
              />
              <div class="summary-signals">
                <div
                  v-for="card in metricCards"
                  :key="`summary-signal-${card.label}`"
                  class="summary-signal-row"
                  :class="observability.metrics?.[card.label]?.state"
                >
                  <div>
                    <strong>{{ telemetryLabels[card.label] || card.label }}</strong>
                    <span>{{ observability.metrics?.[card.label]?.state || card.severity }}</span>
                  </div>
                  <i :style="{ '--bar-pct': `${Math.min(Number(String(card.value).replace('%', '')), 100)}%` }">
                    <b></b>
                  </i>
                  <em>{{ card.value }}</em>
                </div>
              </div>
              <div class="summary-processes">
                <strong>{{ dashboardStore.processes.length }}</strong>
                <span>active processes</span>
                <small>{{ observability.categories?.[0] ? `${observability.categories[0].label} leads memory` : topProcesses[0]?.name || 'collecting workload data' }}</small>
              </div>
            </div>
            <p class="widget-detail">{{ observability.action?.recommendation || 'No dominant pressure source detected.' }}</p>
          </div>

          <div v-else-if="widget.kind === 'network'" class="network-widget">
            <span class="widget-kicker">collector pending</span>
            <strong>Network signals prepared</strong>
            <p>Active connections, upload/download trend, DNS activity, and remote host anomalies will appear when the network collector is enabled.</p>
          </div>
        </WidgetCard>
      </section>
    </section>
</template>
