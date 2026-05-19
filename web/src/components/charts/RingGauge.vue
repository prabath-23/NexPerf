<script setup>
import { computed, ref } from 'vue'
import ChartTooltip from '../ui/ChartTooltip.vue'

const props = defineProps({
  value: { type: Number, default: 0 },
  max: { type: Number, default: 100 },
  tone: { type: String, default: 'normal' },
  label: { type: String, default: '' },
  variant: { type: String, default: 'single' }
})

const radius = 42
const circumference = 2 * Math.PI * radius
const tooltip = ref({ active: false, x: 0, y: 0 })
const percent = computed(() => Math.max(0, Math.min((Number(props.value || 0) / Number(props.max || 100)) * 100, 100)))
const dash = computed(() => `${(percent.value / 100) * circumference} ${circumference}`)
const chartLabel = computed(() => `${percent.value.toFixed(0)} ${props.label || 'score'} · ${props.tone} state`)
const tooltipMeta = computed(() => [
  { label: 'Classification', value: props.tone },
  { label: 'Maximum', value: String(props.max) },
  { label: 'Signal', value: props.variant === 'double' ? 'dual ring' : 'single ring' }
])
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
</script>

<template>
  <div
    class="ring-gauge"
    :class="[tone, `variant-${variant}`]"
    :aria-label="chartLabel"
    role="img"
    tabindex="0"
    @pointermove="setTooltipPosition"
    @pointerleave="hideTooltip"
    @blur="hideTooltip"
  >
    <svg viewBox="0 0 108 108" aria-hidden="true">
      <circle class="ring-track" cx="54" cy="54" :r="radius" />
      <circle class="ring-value" cx="54" cy="54" :r="radius" :stroke-dasharray="dash" />
      <circle v-if="variant === 'double'" class="ring-value secondary" cx="54" cy="54" r="31" :stroke-dasharray="dash" />
    </svg>
    <div>
      <strong>{{ percent.toFixed(0) }}</strong>
      <span>{{ label || 'score' }}</span>
    </div>
    <ChartTooltip
      :active="tooltip.active"
      :x="tooltip.x"
      :y="tooltip.y"
      fixed
      :title="label || 'Runtime Score'"
      :value="`${percent.toFixed(0)}%`"
      :meta="tooltipMeta"
      color="var(--monitor-blue)"
    />
  </div>
</template>
