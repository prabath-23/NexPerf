<script setup>
import SelectControl from '../../components/ui/SelectControl.vue'
import { dashboardStore, persistConfig } from '../../stores/dashboardStore'

defineProps({
  configModeOptions: { type: Array, default: () => [] },
  modeCopy: { type: Object, default: null }
})

const emit = defineEmits(['apply-usage-mode', 'mark-config-dirty'])
</script>

<template>
  <section class="panel settings-panel">
    <div class="section-head">
      <div>
        <h2>Settings</h2>
        <span>{{ dashboardStore.config?.path }}</span>
      </div>
      <button type="button" @click="persistConfig">Save Config</button>
    </div>
    <div
      v-if="dashboardStore.config"
      class="settings-layout"
      @input="emit('mark-config-dirty')"
      @change="emit('mark-config-dirty')"
    >
      <aside>
        <strong>Runtime config</strong>
        <span>{{ modeCopy?.label || dashboardStore.config.usage_mode }} mode</span>
        <p>{{ modeCopy?.description || 'Changes are saved to the local user config.' }}</p>
        <div v-if="modeCopy" class="mode-stack">
          <small v-for="item in modeCopy.stats" :key="item">{{ item }}</small>
        </div>
      </aside>
      <div class="settings-sections">
        <fieldset>
          <legend>Collection</legend>
          <label>
            <span>Polling interval <button type="button" class="info-button" aria-label="Polling interval help">i<span>Seconds between backend metric samples. Lower values feel livelier but use more local resources.</span></button></span>
            <input v-model.number="dashboardStore.config.polling_interval_seconds" type="number" min="1" max="300" />
            <small>Backend sample cadence in seconds.</small>
          </label>
          <label>
            <span>Retention hours <button type="button" class="info-button" aria-label="Retention help">i<span>How long historical CPU, memory, and disk samples stay in local SQLite storage.</span></button></span>
            <input v-model.number="dashboardStore.config.retention_hours" type="number" min="1" max="720" />
            <small>Historical metric window kept locally.</small>
          </label>
          <label>
            <span>Dashboard refresh <button type="button" class="info-button" aria-label="Dashboard refresh help">i<span>Frontend polling cadence in milliseconds for the current workspace section.</span></button></span>
            <input v-model.number="dashboardStore.config.dashboard_refresh_ms" type="number" min="500" max="60000" />
            <small>UI refresh interval in milliseconds.</small>
          </label>
          <label>
            <span>Usage mode <button type="button" class="info-button" aria-label="Usage mode help">i<span>Modes are applied by the backend and tune sampling interval, retention, storage scan depth, refresh cadence, and insight thresholds.</span></button></span>
            <SelectControl
              v-model="dashboardStore.config.usage_mode"
              :options="configModeOptions"
              aria-label="Usage mode"
              @change="emit('apply-usage-mode')"
            />
            <small>{{ dashboardStore.configModes.find((mode) => mode.name === dashboardStore.config.usage_mode)?.description || 'Select how aggressively NexPerf observes this workstation.' }}</small>
          </label>
        </fieldset>
        <fieldset>
          <legend>Thresholds</legend>
          <label>
            <span>CPU warning <button type="button" class="info-button" aria-label="CPU warning help">i<span>CPU percentage where NexPerf starts treating workload pressure as elevated.</span></button></span>
            <input v-model.number="dashboardStore.config.insight_thresholds.cpu_warning" type="number" />
            <small>Insight threshold for CPU pressure.</small>
          </label>
          <label>
            <span>Memory warning <button type="button" class="info-button" aria-label="Memory warning help">i<span>Memory percentage where pressure insights and warning states begin.</span></button></span>
            <input v-model.number="dashboardStore.config.insight_thresholds.memory_warning" type="number" />
            <small>Insight threshold for memory pressure.</small>
          </label>
          <label>
            <span>Disk warning <button type="button" class="info-button" aria-label="Disk warning help">i<span>Disk allocation percentage where storage pressure becomes actionable.</span></button></span>
            <input v-model.number="dashboardStore.config.insight_thresholds.disk_warning" type="number" />
            <small>Insight threshold for storage pressure.</small>
          </label>
          <label>
            <span>Storage rows <button type="button" class="info-button" aria-label="Storage rows help">i<span>Maximum number of top-level directories returned by filesystem intelligence scans.</span></button></span>
            <input v-model.number="dashboardStore.config.storage_limits.max_directory_rows" type="number" />
            <small>Directory rows shown in Storage.</small>
          </label>
        </fieldset>
      </div>
    </div>
    <p v-if="dashboardStore.configStatus" class="settings-status">{{ dashboardStore.configStatus }}</p>
  </section>
</template>
