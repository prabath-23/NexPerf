<script setup>
import { computed, ref } from 'vue'
import ChartTooltip from '../ui/ChartTooltip.vue'

const props = defineProps({
  title: { type: String, required: true },
  metricLabel: { type: String, default: '' },
  windowLabel: { type: String, default: '5 min' },
  points: { type: Array, default: () => [] },
  tone: { type: String, default: 'normal' }
})

const width = 760
const height = 190
const active = ref(null)

const path = computed(() => {
  if (!props.points.length) return ''
  const coordinates = props.points.map((point, index) => {
    const value = Number(point.value || 0)
    const x = props.points.length === 1 ? width : (index / (props.points.length - 1)) * width
    const y = height - (Math.min(value, 100) / 100) * height
    return { x, y }
  })
  return smoothPath(coordinates)
})

const latest = computed(() => {
  if (!props.points.length) return null
  const point = props.points[props.points.length - 1]
  const x = width
  const value = Number(point.value || 0)
  return { x, y: height - (Math.min(value, 100) / 100) * height, value }
})

const tooltip = computed(() => {
  if (!active.value) return null
  return {
    x: (active.value.x / width) * 100,
    y: (active.value.y / height) * 100,
    align: active.value.x > width * 0.72 ? 'left' : 'right'
  }
})

function smoothPath(points) {
  if (!points.length) return ''
  if (points.length === 1) return `M ${points[0].x.toFixed(2)} ${points[0].y.toFixed(2)}`
  let output = `M ${points[0].x.toFixed(2)} ${points[0].y.toFixed(2)}`
  for (let index = 1; index < points.length; index++) {
    const previous = points[index - 1]
    const current = points[index]
    const midX = (previous.x + current.x) / 2
    output += ` Q ${previous.x.toFixed(2)} ${previous.y.toFixed(2)} ${midX.toFixed(2)} ${((previous.y + current.y) / 2).toFixed(2)}`
  }
  const last = points[points.length - 1]
  output += ` T ${last.x.toFixed(2)} ${last.y.toFixed(2)}`
  return output
}

function move(event) {
  if (!props.points.length) return
  const rect = event.currentTarget.getBoundingClientRect()
  const ratio = Math.min(Math.max((event.clientX - rect.left) / rect.width, 0), 1)
  const index = Math.round(ratio * (props.points.length - 1))
  const point = props.points[index]
  const previous = props.points[Math.max(0, index - 1)]
  const x = props.points.length === 1 ? width : (index / (props.points.length - 1)) * width
  const value = Number(point.value || 0)
  const delta = value - Number(previous?.value || value)
  const y = height - (Math.min(value, 100) / 100) * height
  active.value = {
    x,
    y,
    value,
    delta,
    time: point.timestamp ? new Date(point.timestamp).toLocaleTimeString() : '',
    index: index + 1,
    total: props.points.length
  }
}

function leave() {
  active.value = null
}
</script>

<template>
  <section class="panel timeline-panel">
    <div class="section-head">
      <h2>{{ title }}</h2>
      <span>{{ points.length }} samples</span>
    </div>
    <div class="chart-frame" :class="{ interacting: active }">
    <svg class="timeline-chart" :class="[tone, { interacting: active }]" :viewBox="`0 0 ${width} ${height}`" preserveAspectRatio="none" @mousemove="move" @mouseleave="leave">
      <rect class="warning-zone" x="0" y="0" :width="width" :height="height * 0.2" />
      <path class="band warn" :d="`M 0 ${height * 0.2} L ${width} ${height * 0.2}`" />
      <path class="threshold" :d="`M 0 ${height * 0.2} L ${width} ${height * 0.2}`" />
      <path class="band" :d="`M 0 ${height * 0.5} L ${width} ${height * 0.5}`" />
      <path class="band" :d="`M 0 ${height * 0.8} L ${width} ${height * 0.8}`" />
      <path v-if="path" class="area" :d="`${path} L ${width} ${height} L 0 ${height} Z`" />
      <path v-if="path" class="line" :d="path" />
      <path v-if="path && active" class="line active-line" :d="path" />
      <circle v-if="latest" class="latest-point" :cx="latest.x" :cy="latest.y" r="5" />
      <circle v-if="latest" class="latest-ring" :cx="latest.x" :cy="latest.y" r="8" />
      <g v-if="active" class="chart-focus">
        <line :x1="active.x" y1="0" :x2="active.x" :y2="height" />
        <circle :cx="active.x" :cy="active.y" r="4.5" />
      </g>
    </svg>
    <ChartTooltip
      v-if="tooltip"
      :active="!!active"
      :x="tooltip.x"
      :y="tooltip.y"
      :align="tooltip.align"
      :title="metricLabel || title"
      :value="`${active.value.toFixed(1)}%`"
      :timestamp="active.time"
      :meta="[
        { label: 'delta', value: `${active.delta >= 0 ? '+' : ''}${active.delta.toFixed(1)} pts` },
        { label: 'sample', value: `#${active.index} of ${active.total}` },
        { label: 'window', value: windowLabel },
        { label: 'boundary', value: '80%' }
      ]"
    />
    </div>
  </section>
</template>
