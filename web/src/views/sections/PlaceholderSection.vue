<script setup>
defineProps({
  activeSection: { type: String, required: true }
})
</script>

<template>
  <section class="panel placeholder-panel">
    <div class="section-head">
      <div>
        <h2>{{ activeSection === 'network' ? 'Network Telemetry' : 'Diagnostics' }}</h2>
        <span>Workspace surface reserved for v0.4 collectors and privileged diagnostics</span>
      </div>
    </div>
    <p>Collector architecture is ready; this surface is intentionally quiet until live signals are available.</p>
  </section>
</template>
