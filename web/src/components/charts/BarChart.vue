<script setup>
import { computed } from 'vue'

const props = defineProps({
  items: { type: Array, default: () => [] },
  max: { type: Number, default: 0 },
  variant: { type: String, default: 'horizontal' },
  tone: { type: String, default: 'normal' }
})

const maxValue = computed(() => Number(props.max || Math.max(...props.items.map((item) => Number(item.value || 0)), 1)))
</script>

<template>
  <div class="bar-chart" :class="[tone, `variant-${variant}`]">
    <div v-for="item in items" :key="item.label" class="bar-chart-row">
      <span>{{ item.label }}</span>
      <i :style="{ '--bar-pct': `${Math.min((Number(item.value || 0) / maxValue) * 100, 100)}%` }">
        <b :style="{ width: `${Math.min((Number(item.value || 0) / maxValue) * 100, 100)}%` }"></b>
      </i>
      <strong>{{ item.display || item.value }}</strong>
    </div>
  </div>
</template>
