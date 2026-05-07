<script setup>
defineProps({
  label: { type: String, required: true },
  tone: { type: String, default: 'info' },
  pulse: { type: Boolean, default: false }
})
</script>

<template>
  <span class="status-badge" :class="[tone, { pulse }]">
    <span class="dot" />
    {{ label }}
  </span>
</template>
