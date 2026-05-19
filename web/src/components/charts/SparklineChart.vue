<script setup>
import { computed, ref } from 'vue'
import ChartTooltip from '../ui/ChartTooltip.vue'

const props = defineProps({
  points: { type: Array, default: () => [] },
  height: { type: Number, default: 48 },
  max: { type: Number, default: 100 },
  tone: { type: String, default: 'normal' },
  variant: { type: String, default: 'line' },
  interactive: { type: Boolean, default: true },
  compact: { type: Boolean, default: false }
})

const width = 160
const uid = `spark-${Math.random().toString(36).slice(2, 9)}`
const plotPadding = computed(() => {
  if (props.compact) return { top: 0, right: 0, bottom: 0, left: 0 }
  return props.interactive ? { top: 8, right: 18, bottom: 18, left: 8 } : { top: 5, right: 4, bottom: 5, left: 4 }
})
const plotWidth = computed(() => Math.max(width - plotPadding.value.left - plotPadding.value.right, 1))
const plotHeight = computed(() => Math.max(props.height - plotPadding.value.top - plotPadding.value.bottom, 1))
const active = ref(null)

const yTicks = computed(() => props.interactive && !props.compact ? [
  { value: props.max, y: plotPadding.value.top },
  { value: props.max / 2, y: plotPadding.value.top + plotHeight.value / 2 },
  { value: 0, y: plotPadding.value.top + plotHeight.value }
] : [])

const xTicks = computed(() => props.interactive && !props.compact ? [
  { label: 'start', x: plotPadding.value.left },
  { label: 'now', x: plotPadding.value.left + plotWidth.value }
] : [])

const path = computed(() => {
  if (!props.points.length) return ''
  const coordinates = props.points.map((point, index) => {
    const value = Number(point.value || 0)
    const x = props.points.length === 1 ? plotPadding.value.left + plotWidth.value : plotPadding.value.left + (index / (props.points.length - 1)) * plotWidth.value
    const y = plotPadding.value.top + plotHeight.value - (Math.min(value, props.max) / props.max) * plotHeight.value
    return { x, y }
  })
  if (coordinates.length === 1) return `M ${coordinates[0].x.toFixed(2)} ${coordinates[0].y.toFixed(2)}`
  let output = `M ${coordinates[0].x.toFixed(2)} ${coordinates[0].y.toFixed(2)}`
  for (let index = 1; index < coordinates.length; index++) {
    const previous = coordinates[index - 1]
    const current = coordinates[index]
    const midX = (previous.x + current.x) / 2
    output += ` Q ${previous.x.toFixed(2)} ${previous.y.toFixed(2)} ${midX.toFixed(2)} ${((previous.y + current.y) / 2).toFixed(2)}`
  }
  const last = coordinates[coordinates.length - 1]
  return `${output} T ${last.x.toFixed(2)} ${last.y.toFixed(2)}`
})

const latest = computed(() => {
  if (!props.points.length) return null
  const point = props.points[props.points.length - 1]
  const value = Number(point.value || 0)
  return { x: plotPadding.value.left + plotWidth.value, y: plotPadding.value.top + plotHeight.value - (Math.min(value, props.max) / props.max) * plotHeight.value }
})

const bars = computed(() => {
  if (!props.points.length) return []
  const gap = 2
  const barWidth = Math.max(2, (plotWidth.value - gap * (props.points.length - 1)) / props.points.length)
  return props.points.map((point, index) => {
    const value = Math.min(Number(point.value || 0), props.max)
    const barHeight = Math.max(2, (value / props.max) * plotHeight.value)
    return {
      x: plotPadding.value.left + index * (barWidth + gap),
      y: plotPadding.value.top + plotHeight.value - barHeight,
      width: barWidth,
      height: barHeight
    }
  })
})

const areaPath = computed(() => path.value ? `${path.value} L ${plotPadding.value.left + plotWidth.value} ${plotPadding.value.top + plotHeight.value} L ${plotPadding.value.left} ${plotPadding.value.top + plotHeight.value} Z` : '')

const stats = computed(() => {
  const values = props.points.map((point) => Number(point.value || 0)).filter((value) => Number.isFinite(value))
  if (!values.length) {
    return {
      avg: 0,
      peak: 0,
      low: 0,
      delta: 0,
      volatility: 0,
      anomalies: 0,
      state: 'collecting',
      direction: 'flat'
    }
  }
  const avg = values.reduce((sum, value) => sum + value, 0) / values.length
  const peak = Math.max(...values)
  const low = Math.min(...values)
  const delta = values[values.length - 1] - values[0]
  const volatility = Math.sqrt(values.reduce((sum, value) => sum + ((value - avg) ** 2), 0) / values.length)
  const anomalies = values.filter((value) => value >= Math.max(80, avg + 18)).length
  const state = anomalies >= 3 ? 'bursty' : volatility >= 12 ? 'volatile' : Math.abs(delta) >= 3 ? (delta > 0 ? 'rising' : 'recovering') : 'steady'
  return {
    avg,
    peak,
    low,
    delta,
    volatility,
    anomalies,
    state,
    direction: delta > 0.75 ? 'up' : delta < -0.75 ? 'down' : 'flat'
  }
})

const tooltip = computed(() => active.value ? {
  x: active.value.clientX,
  y: active.value.clientY,
  align: active.value.x > width * 0.72 ? 'left' : 'right'
} : null)

function move(event) {
  if (!props.interactive || !props.points.length) return
  const rect = event.currentTarget.getBoundingClientRect()
  const relativeX = ((event.clientX - rect.left) / rect.width) * width
  const ratio = Math.min(Math.max((relativeX - plotPadding.value.left) / plotWidth.value, 0), 1)
  const index = Math.round(ratio * (props.points.length - 1))
  const point = props.points[index]
  const x = props.points.length === 1 ? plotPadding.value.left + plotWidth.value : plotPadding.value.left + (index / (props.points.length - 1)) * plotWidth.value
  const y = plotPadding.value.top + plotHeight.value - (Math.min(Number(point.value || 0), props.max) / props.max) * plotHeight.value
  const tooltipWidth = props.compact ? 190 : 240
  const tooltipHeight = props.compact ? 72 : 150
  const tooltipOffsetX = props.compact ? 12 : 18
  const tooltipOffsetY = props.compact ? 14 : -74
  const compactY = rect.bottom + 8 + tooltipHeight > window.innerHeight
    ? rect.top - tooltipHeight - 8
    : rect.bottom + 8
  active.value = {
    x,
    y,
    value: Number(point.value || 0),
    time: point.timestamp ? new Date(point.timestamp).toLocaleTimeString() : '',
    index: index + 1,
    total: props.points.length,
    stats: stats.value,
    clientX: Math.min(Math.max(event.clientX + tooltipOffsetX, 18), window.innerWidth - tooltipWidth),
    clientY: props.compact
      ? Math.min(Math.max(compactY, 18), window.innerHeight - tooltipHeight)
      : Math.min(Math.max(event.clientY + tooltipOffsetY, 18), window.innerHeight - tooltipHeight)
  }
}

function leave() {
  active.value = null
}
</script>

<template>
  <div class="chart-frame" :class="{ interacting: active }" @mouseleave="leave">
    <svg class="sparkline" :class="[tone, `variant-${variant}`, { interacting: active, inert: !interactive, compact }]" :viewBox="`0 0 ${width} ${height}`" preserveAspectRatio="none" @mousemove="move">
      <defs>
        <linearGradient :id="`${uid}-stroke`" x1="0%" x2="100%" y1="0%" y2="0%">
          <stop offset="0%" stop-color="#0a84ff" />
          <stop offset="54%" stop-color="#00c7be" />
          <stop offset="100%" stop-color="#30d158" />
        </linearGradient>
        <linearGradient :id="`${uid}-fill`" x1="0%" x2="0%" y1="0%" y2="100%">
          <stop offset="0%" stop-color="#0a84ff" stop-opacity="0.32" />
          <stop offset="100%" stop-color="#0a84ff" stop-opacity="0.02" />
        </linearGradient>
        <filter :id="`${uid}-glow`" x="-30%" y="-80%" width="160%" height="260%">
          <feGaussianBlur stdDeviation="2.2" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      <rect class="plot-area" :x="plotPadding.left" :y="plotPadding.top" :width="plotWidth" :height="plotHeight" />
      <rect class="warning-zone" :x="plotPadding.left" :y="plotPadding.top" :width="plotWidth" :height="plotHeight * 0.2" />
      <path class="grid-line major" :d="`M ${plotPadding.left} ${plotPadding.top + plotHeight * 0.2} L ${plotPadding.left + plotWidth} ${plotPadding.top + plotHeight * 0.2}`" />
      <path class="grid-line" :d="`M ${plotPadding.left} ${plotPadding.top + plotHeight * 0.5} L ${plotPadding.left + plotWidth} ${plotPadding.top + plotHeight * 0.5}`" />
      <path class="grid-line" :d="`M ${plotPadding.left} ${plotPadding.top + plotHeight * 0.8} L ${plotPadding.left + plotWidth} ${plotPadding.top + plotHeight * 0.8}`" />
      <path v-for="tick in xTicks" :key="tick.label" class="grid-line vertical" :d="`M ${tick.x} ${plotPadding.top} L ${tick.x} ${plotPadding.top + plotHeight}`" />
      <path v-if="variant === 'area' && areaPath" class="area" :d="areaPath" :fill="`url(#${uid}-fill)`" />
      <g v-if="variant === 'bar'" class="spark-bars">
        <rect v-for="bar in bars" :key="`${bar.x}-${bar.height}`" :x="bar.x" :y="bar.y" :width="bar.width" :height="bar.height" rx="1.5" />
      </g>
      <path v-if="path && variant !== 'bar'" class="line telemetry-glow" :d="path" :stroke="`url(#${uid}-stroke)`" :filter="`url(#${uid}-glow)`" />
      <path v-if="path && variant !== 'bar'" class="line" :d="path" :stroke="`url(#${uid}-stroke)`" />
      <path v-if="path && active && variant !== 'bar'" class="line active-line" :d="path" />
      <circle v-if="latest && variant !== 'bar'" class="latest-point" :cx="latest.x" :cy="latest.y" r="2.5" />
      <circle v-if="latest && variant !== 'bar'" class="latest-ring" :cx="latest.x" :cy="latest.y" r="4.5" />
      <g v-if="interactive && active" class="chart-focus">
        <line :x1="active.x" y1="0" :x2="active.x" :y2="height" />
        <circle :cx="active.x" :cy="active.y" r="3.5" />
      </g>
      <g v-if="interactive && !compact" class="axis-labels">
        <text v-for="tick in yTicks" :key="tick.value" :x="width - 4" :y="tick.y + 3" text-anchor="end">{{ Math.round(tick.value) }}</text>
        <text :x="plotPadding.left" :y="height - 10">start</text>
        <text :x="plotPadding.left + plotWidth" :y="height - 10" text-anchor="end">now</text>
      </g>
    </svg>
    <ChartTooltip v-if="tooltip" :active="interactive && !!active" :x="tooltip.x" :y="tooltip.y" :align="tooltip.align" fixed>
      <slot name="tooltip" :point="active" :stats="stats">
        <div class="chart-tooltip mini">
          <strong>{{ active.value.toFixed(1) }}%</strong>
          <small>{{ active.time }}</small>
        </div>
      </slot>
    </ChartTooltip>
  </div>
</template>
