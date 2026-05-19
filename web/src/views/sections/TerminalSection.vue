<script setup>
import { dashboardStore } from '../../stores/dashboardStore'

defineProps({
  terminalCwd: { type: String, default: '' },
  terminalCommand: { type: String, default: '' },
  terminalResult: { type: Object, default: null },
  terminalTab: { type: String, default: 'terminal' }
})

const emit = defineEmits(['run-terminal', 'update:terminal-command', 'update:terminal-cwd', 'update:terminal-tab'])
</script>

<template>
  <section class="panel terminal-panel">
    <div class="section-head">
      <div>
        <h2>Terminal</h2>
        <span>Local command surface with execution history</span>
      </div>
    </div>
    <div class="terminal-tabs">
      <button type="button" :class="{ active: terminalTab === 'terminal' }" @click="emit('update:terminal-tab', 'terminal')">Terminal</button>
      <button type="button" :class="{ active: terminalTab === 'history' }" @click="emit('update:terminal-tab', 'history')">Command history</button>
    </div>
    <form v-if="terminalTab === 'terminal'" class="terminal-window" @submit.prevent="emit('run-terminal')">
      <header>
        <span></span><span></span><span></span>
        <strong>{{ terminalCwd || '~' }}</strong>
      </header>
      <div class="terminal-screen" aria-label="Interactive terminal">
        <label class="terminal-cwd">
          <span>cwd</span>
          <input :value="terminalCwd" placeholder="/Users/prabath" @input="emit('update:terminal-cwd', $event.target.value)" />
        </label>
        <div v-if="terminalResult" class="terminal-session-output">
          <p><span>$</span> {{ terminalResult.command }}</p>
          <pre v-if="terminalResult.stdout">{{ terminalResult.stdout }}</pre>
          <pre v-if="terminalResult.stderr" class="stderr">{{ terminalResult.stderr }}</pre>
          <small>exit {{ terminalResult.exit_code }} · {{ terminalResult.duration_ms }}ms</small>
        </div>
        <div v-else class="terminal-hint">
          <span>NexPerf shell</span>
          <p>Run local inspection commands from the current service. Output appears in this terminal and is saved to history.</p>
        </div>
        <div class="terminal-line">
          <span class="prompt-mark">$</span>
          <input
            :value="terminalCommand"
            autocomplete="off"
            spellcheck="false"
            autofocus
            placeholder="nexperf status"
            @input="emit('update:terminal-command', $event.target.value)"
          />
          <button type="submit" :disabled="dashboardStore.terminalRunning">{{ dashboardStore.terminalRunning ? 'Running' : 'Run' }}</button>
        </div>
      </div>
    </form>
    <div v-if="terminalTab === 'history'" class="terminal-output">
      <article v-for="item in dashboardStore.terminalHistory" :key="`${item.command}-${item.duration_ms}-${item.exit_code}`">
        <header>
          <strong>$ {{ item.command }}</strong>
          <span>exit {{ item.exit_code }} · {{ item.duration_ms }}ms</span>
        </header>
        <pre v-if="item.stdout">{{ item.stdout }}</pre>
        <pre v-if="item.stderr" class="stderr">{{ item.stderr }}</pre>
      </article>
      <p v-if="!dashboardStore.terminalHistory.length" class="terminal-empty">No commands run yet.</p>
    </div>
  </section>
</template>
