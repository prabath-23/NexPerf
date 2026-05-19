<script setup>
defineProps({
  commandManual: { type: Array, default: () => [] },
  manualLoading: { type: Boolean, default: false },
  manualQuery: { type: String, default: '' },
  manualResult: { type: Object, default: null },
  processManual: { type: Array, default: () => [] }
})

const emit = defineEmits(['lookup-manual', 'update:manual-query'])
</script>

<template>
  <section class="panel manual-panel">
    <div class="section-head">
      <div>
        <h2>Manual</h2>
        <span>NexPerf guidance, process categories, and local system man pages</span>
      </div>
    </div>
    <form class="manual-lookup" @submit.prevent="emit('lookup-manual')">
      <div>
        <span>Manual page</span>
        <input :value="manualQuery" placeholder="netstat, ps, launchctl, top..." @input="emit('update:manual-query', $event.target.value)" />
        <button type="submit" :disabled="manualLoading">{{ manualLoading ? 'Searching' : 'Open' }}</button>
      </div>
    </form>
    <article v-if="manualResult" class="manual-reader">
      <header>
        <strong>man {{ manualQuery }}</strong>
        <span>local manual · exit {{ manualResult.exit_code }}</span>
      </header>
      <pre>{{ manualResult.stdout || manualResult.stderr || 'No manual output returned.' }}</pre>
    </article>
    <div class="manual-grid">
      <section>
        <h3>Commands</h3>
        <article v-for="item in commandManual" :key="item.command">
          <strong>{{ item.command }}</strong>
          <span>{{ item.description }}</span>
        </article>
      </section>
      <section>
        <h3>Processes</h3>
        <article v-for="item in processManual" :key="item.label">
          <strong>{{ item.label }}</strong>
          <span>{{ item.detail }}</span>
        </article>
      </section>
    </div>
  </section>
</template>
