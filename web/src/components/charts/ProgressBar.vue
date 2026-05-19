<script setup>
import { computed } from 'vue'

const props = defineProps({
  value: { type: Number, default: 0 },
  max: { type: Number, default: 100 },
  tone: { type: String, default: 'normal' },
  label: { type: String, default: '' },
  compact: { type: Boolean, default: false }
})

const percent = computed(() => Math.max(0, Math.min((Number(props.value || 0) / Number(props.max || 100)) * 100, 100)))
</script>

<template>
  <div class="progress-chart" :class="[tone, { compact }]">
    <div v-if="label || !compact" class="progress-chart-head">
      <span>{{ label }}</span>
      <strong>{{ percent.toFixed(0) }}%</strong>
    </div>
    <div class="progress-track">
      <i :style="{ width: `${percent}%` }"></i>
    </div>
  </div>
</template>
