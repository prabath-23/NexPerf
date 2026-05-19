<script setup>
import { dashboardStore, refreshStorage } from '../../stores/dashboardStore'

defineProps({
  childStorageShare: { type: Function, required: true },
  formatBytes: { type: Function, required: true },
  storageShare: { type: Function, required: true }
})
</script>

<template>
  <section class="panel storage-panel">
    <div class="section-head">
      <div>
        <h2>Storage Map</h2>
        <span>{{ dashboardStore.storage?.root || 'Scanning filesystem root' }}</span>
      </div>
      <button type="button" @click="refreshStorage()">Rescan</button>
    </div>
    <div class="directory-map" aria-label="Directory size map">
      <article
        v-for="entry in (dashboardStore.storage?.entries || []).slice(0, 10)"
        :key="`map-${entry.path}`"
        :class="entry.heat"
        :style="{ '--share': `${Math.max(storageShare(entry), 7)}%` }"
      >
        <span>{{ entry.name }}</span>
        <strong>{{ formatBytes(entry.size_bytes) }}</strong>
        <small>{{ entry.path }}</small>
      </article>
    </div>
    <div class="directory-browser">
      <div class="directory-root">
        <strong>{{ dashboardStore.storage?.root || '~' }}</strong>
        <span>{{ formatBytes(dashboardStore.storage?.total_bytes) }} scanned · {{ dashboardStore.storage?.entries?.length || 0 }} directories</span>
      </div>
      <details v-for="entry in dashboardStore.storage?.entries || []" :key="entry.path" class="directory-tree-node" :class="entry.heat" open>
        <summary>
          <div class="directory-node">
            <span class="branch"></span>
            <div>
              <strong>{{ entry.name }}</strong>
              <span>{{ entry.kind }} · {{ entry.path }}</span>
              <small>{{ entry.description }}</small>
            </div>
          </div>
          <div class="storage-meter">
            <i :style="{ width: `${storageShare(entry)}%` }"></i>
          </div>
          <em>{{ formatBytes(entry.size_bytes) }}</em>
        </summary>
        <div v-if="entry.children?.length" class="directory-children">
          <article v-for="child in entry.children" :key="child.path" :class="child.heat">
            <div class="directory-node child">
              <span class="branch"></span>
              <div>
                <strong>{{ child.name }}</strong>
                <span>{{ child.kind }} · {{ child.path }}</span>
                <small>{{ child.description }}</small>
              </div>
            </div>
            <div class="storage-meter">
              <i :style="{ width: `${childStorageShare(child, entry)}%` }"></i>
            </div>
            <em>{{ formatBytes(child.size_bytes) }}</em>
          </article>
        </div>
      </details>
    </div>
    <div class="section-head subhead">
      <div>
        <h2>Storage Artifacts</h2>
        <span>Largest known caches, dependencies, downloads, logs, and generated data</span>
      </div>
    </div>
    <div class="artifact-list">
      <article v-for="entry in dashboardStore.storage?.detected || []" :key="entry.path">
        <span class="artifact-dot" :class="entry.heat"></span>
        <div>
          <strong>{{ entry.name }}</strong>
          <small>{{ entry.path }}</small>
          <p>{{ entry.description }}</p>
        </div>
        <span>{{ entry.kind }}</span>
        <em>{{ formatBytes(entry.size_bytes) }}</em>
      </article>
    </div>
  </section>
</template>
