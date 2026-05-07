<script setup>
defineProps({
  system: { type: Object, default: null },
  score: { type: Object, default: null }
})

function pct(value) {
  return `${Number(value || 0).toFixed(1)}%`
}

function tone(value) {
  if (value >= 90) return 'critical'
  if (value >= 80) return 'warning'
  return 'normal'
}
</script>

<template>
  <section class="system-pulse" v-if="system">
    <div class="pulse-copy">
      <span>System Pulse</span>
      <h2>{{ system.hostname || 'Local machine' }}</h2>
      <p>{{ system.os }} / {{ system.arch }} · live resource signal</p>
    </div>

    <div class="pulse-orb" aria-hidden="true">
      <div class="orb-ring ring-one" />
      <div class="orb-ring ring-two" />
      <div class="orb-core">
        <strong>{{ score?.value ?? Math.round(100 - ((system.cpu_percent + system.memory.percent + system.disk.percent) / 3) / 2) }}</strong>
        <span>health</span>
      </div>
    </div>

    <div class="pulse-readouts">
      <div :class="tone(system.cpu_percent)">
        <span>CPU</span>
        <strong>{{ pct(system.cpu_percent) }}</strong>
      </div>
      <div :class="tone(system.memory.percent)">
        <span>Memory</span>
        <strong>{{ pct(system.memory.percent) }}</strong>
      </div>
      <div :class="tone(system.disk.percent)">
        <span>Disk</span>
        <strong>{{ pct(system.disk.percent) }}</strong>
      </div>
    </div>

    <div class="health-note" v-if="score">
      <strong>{{ score.status }}</strong>
      <span>{{ score.summary }}</span>
    </div>
  </section>
</template>
