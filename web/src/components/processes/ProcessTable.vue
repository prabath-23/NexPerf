<script setup>
import { computed, ref } from 'vue'
import DataTable from '../ui/DataTable.vue'

const props = defineProps({
  processes: { type: Array, default: () => [] },
  previousProcesses: { type: Array, default: () => [] },
  limit: { type: Number, default: 12 }
})

const emit = defineEmits(['update:limit'])

const query = ref('')
const sortKey = ref('memory_mb')
const sortDir = ref('desc')
const limitOptions = [8, 12, 20, 40, 80]

const columns = [
  { key: 'pid', label: 'PID', sortable: true },
  { key: 'name', label: 'Name', sortable: true },
  { key: 'category', label: 'Type', sortable: true },
  { key: 'memory_mb', label: 'Memory', sortable: true },
  { key: 'cpu_percent', label: 'CPU', sortable: true },
  { key: 'user', label: 'User', sortable: true }
]

const categorySummary = computed(() => {
  const map = new Map()
  for (const process of props.processes) {
    const category = process.category || 'app'
    const current = map.get(category) || { category, count: 0, memory: 0, cpu: 0 }
    current.count += 1
    current.memory += Number(process.memory_mb || 0)
    current.cpu += Number(process.cpu_percent || 0)
    map.set(category, current)
  }
  return [...map.values()].sort((a, b) => b.memory - a.memory).slice(0, 5)
})

const previousByPid = computed(() => new Map(props.previousProcesses.map((process) => [process.pid, process])))

const maxMemory = computed(() => Math.max(1, ...props.processes.map((process) => Number(process.memory_mb || 0))))

const filtered = computed(() => {
  const needle = query.value.toLowerCase().trim()
  const rows = needle
    ? props.processes.filter((process) => `${process.pid} ${process.name} ${process.user || ''}`.toLowerCase().includes(needle))
    : [...props.processes]
  rows.sort((a, b) => {
    const av = a[sortKey.value]
    const bv = b[sortKey.value]
    const result = typeof av === 'number' ? av - bv : String(av || '').localeCompare(String(bv || ''))
    return sortDir.value === 'asc' ? result : -result
  })
  return rows
})

function sort(key) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'desc'
  }
}

function pct(value) {
  return `${Number(value || 0).toFixed(1)}%`
}

function trendFor(process) {
  const previous = previousByPid.value.get(process.pid)
  if (!previous) return { label: 'new', tone: 'new' }
  const memoryDelta = Number(process.memory_mb || 0) - Number(previous.memory_mb || 0)
  const cpuDelta = Number(process.cpu_percent || 0) - Number(previous.cpu_percent || 0)
  if (memoryDelta > 20) return { label: `mem +${memoryDelta.toFixed(0)} MB`, tone: 'up' }
  if (cpuDelta > 8) return { label: `cpu +${cpuDelta.toFixed(1)}%`, tone: 'up' }
  if (memoryDelta < -20) return { label: `mem ${memoryDelta.toFixed(0)} MB`, tone: 'down' }
  return { label: 'steady', tone: 'flat' }
}
</script>

<template>
  <section class="panel process-panel">
    <div class="section-head process-head">
      <div>
        <h2>Process Observability</h2>
        <span>{{ filtered.length }} of {{ processes.length }} processes · categorized by workload type</span>
      </div>
      <div class="process-controls">
        <label>
          Show
          <select :value="limit" @change="emit('update:limit', Number($event.target.value))">
            <option v-for="option in limitOptions" :key="option" :value="option">{{ option }}</option>
          </select>
        </label>
        <input v-model="query" type="search" placeholder="Search processes" />
      </div>
    </div>
    <div class="process-categories">
      <div v-for="item in categorySummary" :key="item.category">
        <span>{{ item.category }}</span>
        <strong>{{ item.count }}</strong>
        <small>{{ item.memory.toFixed(0) }} MB</small>
      </div>
    </div>
    <div class="table-wrap">
      <DataTable :columns="columns" :rows="filtered" :sort-key="sortKey" :sort-dir="sortDir" @sort="sort">
        <tr v-for="process in filtered" :key="process.pid" :class="{ hot: process.cpu_percent > 20 }">
          <td>{{ process.pid }}</td>
          <td>{{ process.name }}</td>
          <td><span class="category-pill">{{ process.category || 'app' }}</span></td>
          <td>
            <div class="memory-cell">
              <span>{{ process.memory_mb.toFixed(1) }} MB</span>
              <i :style="{ width: `${Math.min((process.memory_mb / maxMemory) * 100, 100)}%` }" />
            </div>
          </td>
          <td>{{ pct(process.cpu_percent) }}</td>
          <td>
            <span>{{ process.user }}</span>
            <em :class="trendFor(process).tone">{{ trendFor(process).label }}</em>
          </td>
        </tr>
      </DataTable>
    </div>
  </section>
</template>
