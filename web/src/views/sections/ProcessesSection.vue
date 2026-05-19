<script setup>
import { computed } from 'vue'
import ProcessTable from '../../components/processes/ProcessTable.vue'
import { dashboardStore, refreshProcessTree } from '../../stores/dashboardStore'

const props = defineProps({
  hierarchyGroups: { type: Array, default: () => [] },
  parentLabel: { type: Function, required: true }
})

const emit = defineEmits(['refresh'])

function updateLimit(limit) {
  dashboardStore.processLimit = limit
  emit('refresh')
  refreshProcessTree()
}

const treeStats = computed(() => {
  const groups = props.hierarchyGroups || []
  const children = groups.flatMap((group) => group.children || [])
  const memory = children.reduce((sum, process) => sum + Number(process.memory_mb || 0), 0)
  const leadGroup = groups
    .map((group) => ({
      name: group.parentName,
      count: group.children?.length || 0
    }))
    .sort((a, b) => b.count - a.count)[0]

  return {
    groups: groups.length,
    children: children.length,
    memory,
    leadGroup
  }
})
</script>

<template>
  <section class="workspace-grid single">
    <ProcessTable
      :processes="dashboardStore.processes"
      :previous-processes="dashboardStore.previousProcesses"
      :limit="dashboardStore.processLimit"
      @update:limit="updateLimit"
    />
    <section class="panel process-tree-panel">
      <div class="section-head">
        <div>
          <h2>Process Hierarchy</h2>
          <span>Grouped by parent process so launch paths are visible</span>
        </div>
        <button type="button" @click="refreshProcessTree">Reload</button>
      </div>
      <div class="process-tree-summary">
        <article>
          <span>Parent groups</span>
          <strong>{{ treeStats.groups }}</strong>
        </article>
        <article>
          <span>Child workload</span>
          <strong>{{ treeStats.children }}</strong>
        </article>
        <article>
          <span>Mapped memory</span>
          <strong>{{ treeStats.memory.toFixed(0) }} MB</strong>
        </article>
        <article>
          <span>Dominant parent</span>
          <strong>{{ treeStats.leadGroup?.name || 'Collecting' }}</strong>
        </article>
      </div>
      <div class="process-tree">
        <details v-for="group in hierarchyGroups.slice(0, 18)" :key="group.key" class="process-group" open>
          <summary>
            <span class="tree-stem root"><i></i></span>
            <strong>{{ group.parentName }}</strong>
            <small v-if="group.parentPid">PID {{ group.parentPid }}</small>
            <em>{{ group.children.length }} child processes</em>
          </summary>
          <article v-for="process in group.children.slice(0, 28)" :key="process.pid" :class="{ child: process.ppid }">
            <span class="tree-stem"><i></i></span>
            <div>
              <strong>{{ process.name }}</strong>
              <span>{{ process.category || 'app' }} · {{ process.runtime || 'runtime unknown' }}</span>
              <small>{{ parentLabel(process) }}</small>
            </div>
            <dl>
              <div><dt>PID</dt><dd>{{ process.pid }}</dd></div>
              <div><dt>PPID</dt><dd>{{ process.ppid || '-' }}</dd></div>
              <div><dt>Threads</dt><dd>{{ process.threads || 0 }}</dd></div>
              <div><dt>Memory</dt><dd>{{ process.memory_mb.toFixed(0) }} MB</dd></div>
            </dl>
          </article>
        </details>
      </div>
    </section>
  </section>
</template>
