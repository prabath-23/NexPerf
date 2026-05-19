<script setup>
defineProps({
  active: { type: Boolean, default: false },
  x: { type: Number, default: 0 },
  y: { type: Number, default: 0 },
  align: { type: String, default: 'right' },
  title: { type: String, default: '' },
  value: { type: String, default: '' },
  timestamp: { type: String, default: '' },
  meta: { type: Array, default: () => [] },
  color: { type: String, default: 'var(--blue)' },
  fixed: { type: Boolean, default: false }
})
</script>

<template>
  <Teleport to="body" :disabled="!fixed">
    <div
      v-if="active"
      class="chart-tooltip-layer"
      :class="[`align-${align}`, { fixed }]"
      :style="fixed ? { left: `${x}px`, top: `${y}px` } : { left: `${x}%`, top: `${y}%` }"
    >
      <slot>
        <div class="chart-tooltip rich" :style="{ '--tooltip-accent': color }">
          <div class="tooltip-top">
            <span><i></i>{{ title }}</span>
            <em>{{ timestamp }}</em>
          </div>
          <div class="tooltip-value-row">
            <strong>{{ value }}</strong>
          </div>
          <div v-if="meta.length" class="tooltip-grid">
            <template v-for="item in meta" :key="item.label">
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </template>
          </div>
        </div>
      </slot>
    </div>
  </Teleport>
</template>
