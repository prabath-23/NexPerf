<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  points: { type: Array, default: () => [] },
  height: { type: Number, default: 48 },
  max: { type: Number, default: 100 },
  tone: { type: String, default: 'normal' }
})

const width = 160
const active = ref(null)

const path = computed(() => {
  if (!props.points.length) return ''
  const coordinates = props.points.map((point, index) => {
    const value = Number(point.value || 0)
    const x = props.points.length === 1 ? width : (index / (props.points.length - 1)) * width
    const y = props.height - (Math.min(value, props.max) / props.max) * props.height
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
  return { x: width, y: props.height - (Math.min(value, props.max) / props.max) * props.height }
})

function move(event) {
  if (!props.points.length) return
  const rect = event.currentTarget.getBoundingClientRect()
  const ratio = Math.min(Math.max((event.clientX - rect.left) / rect.width, 0), 1)
  const index = Math.round(ratio * (props.points.length - 1))
  const point = props.points[index]
  const x = props.points.length === 1 ? width : (index / (props.points.length - 1)) * width
  const y = props.height - (Math.min(Number(point.value || 0), props.max) / props.max) * props.height
  active.value = { x, y, value: Number(point.value || 0), timestamp: point.timestamp }
}

function leave() {
  active.value = null
}
</script>

<template>
  <svg class="sparkline" :class="tone" :viewBox="`0 0 ${width} ${height}`" preserveAspectRatio="none" @mousemove="move" @mouseleave="leave">
    <path class="grid-line" :d="`M 0 ${height * 0.2} L ${width} ${height * 0.2}`" />
    <path class="grid-line" :d="`M 0 ${height * 0.8} L ${width} ${height * 0.8}`" />
    <path v-if="path" class="line" :d="path" />
    <circle v-if="latest" class="latest-point" :cx="latest.x" :cy="latest.y" r="3.5" />
    <g v-if="active" class="chart-focus">
      <line :x1="active.x" y1="0" :x2="active.x" :y2="height" />
      <circle :cx="active.x" :cy="active.y" r="3.5" />
      <text :x="Math.min(active.x + 6, width - 44)" :y="Math.max(active.y - 7, 10)">{{ active.value.toFixed(1) }}%</text>
    </g>
  </svg>
</template>
