<script setup>
import SparklineChart from '../charts/SparklineChart.vue'

defineProps({
  label: { type: String, required: true },
  value: { type: String, required: true },
  detail: { type: String, required: true },
  severity: { type: String, default: 'info' },
  history: { type: Array, default: () => [] },
  trend: { type: Object, default: () => ({ label: 'collecting trend', direction: 'flat' }) },
  peak: { type: String, default: 'peak --' }
})
</script>

<template>
  <article class="metric-card" :class="severity">
    <div class="metric-topline">
      <span>{{ label }}</span>
      <em class="metric-state">{{ severity === 'normal' ? 'normal' : severity }}</em>
    </div>
    <strong>{{ value }}</strong>
    <p>{{ detail }}</p>
    <div class="metric-context">
      <span :class="trend.direction">{{ trend.direction === 'up' ? '↑' : trend.direction === 'down' ? '↓' : '→' }} {{ trend.label }}</span>
      <span>{{ peak }}</span>
    </div>
    <SparklineChart :points="history" :tone="severity" />
  </article>
</template>
