<script setup>
import { computed, ref } from 'vue'
import ChartTooltip from '../ui/ChartTooltip.vue'

const props = defineProps({
  segments: { type: Array, default: () => [] },
  total: { type: Number, default: 100 },
  label: { type: String, default: '' },
  value: { type: [Number, String], default: '' },
  variant: { type: String, default: 'stacked' }
})

const radius = 42
const circumference = 2 * Math.PI * radius
const tooltip = ref({ active: false, x: 0, y: 0 })
const segmentGap = computed(() => props.variant === 'stacked' && props.segments.length > 1 ? 5.5 : 0)
const chartLabel = computed(() => {
  const parts = props.segments
    .filter((segment) => segment?.label)
    .slice(0, 3)
    .map((segment) => `${segment.label}: ${Number(segment.value || 0).toFixed(1)}`)
  const headline = props.value ? `${props.value} ${props.label}` : props.label
  return [headline, ...parts].filter(Boolean).join(' · ')
})
const tooltipMeta = computed(() =>
  props.segments
    .filter((segment) => segment?.label)
    .slice(0, 4)
    .map((segment) => ({
      label: segment.label,
      value: `${Number(segment.value || 0).toFixed(1)}`
    }))
)
function setTooltipPosition(event) {
  const width = typeof window === 'undefined' ? 1200 : window.innerWidth
  const height = typeof window === 'undefined' ? 800 : window.innerHeight
  tooltip.value = {
    active: true,
    x: Math.min(Math.max(event.clientX + 16, 18), width - 286),
    y: Math.min(Math.max(event.clientY - 28, 18), height - 220)
  }
}
function hideTooltip() {
  tooltip.value = { ...tooltip.value, active: false }
}
const normalized = computed(() => {
  let offset = 0
  const total = Number(props.total || props.segments.reduce((sum, item) => sum + Number(item.value || 0), 0) || 100)
  const gap = segmentGap.value
  const drawable = Math.max(circumference - gap * props.segments.length, 1)
  return props.segments.map((segment, index) => {
    const value = Math.max(0, Number(segment.value || 0))
    const length = Math.max((value / total) * drawable, 0)
    const current = {
      ...segment,
      key: `${segment.label || index}-${value}`,
      dash: `${length} ${circumference}`,
      offset: -offset
    }
    offset += length + gap
    return current
  })
})
</script>

<template>
  <div
    class="donut-chart"
    :class="`variant-${variant}`"
    :aria-label="chartLabel"
    role="img"
    tabindex="0"
    @pointermove="setTooltipPosition"
    @pointerleave="hideTooltip"
    @blur="hideTooltip"
  >
    <svg viewBox="0 0 108 108" aria-hidden="true">
      <circle class="donut-track" cx="54" cy="54" :r="radius" />
      <circle
        v-for="segment in normalized"
        :key="segment.key"
        class="donut-segment"
        cx="54"
        cy="54"
        :r="radius"
        :stroke-dasharray="segment.dash"
        :stroke-dashoffset="segment.offset"
        :style="{ '--segment-color': segment.color }"
      />
    </svg>
    <div>
      <strong>{{ value }}</strong>
      <span>{{ label }}</span>
    </div>
    <ChartTooltip
      :active="tooltip.active"
      :x="tooltip.x"
      :y="tooltip.y"
      fixed
      title="Runtime State"
      :value="String(value || label || 'Telemetry')"
      :meta="tooltipMeta"
      color="var(--monitor-cyan)"
    />
  </div>
</template>
