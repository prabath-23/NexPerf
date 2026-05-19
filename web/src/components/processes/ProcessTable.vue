<script setup>
import { computed, ref } from 'vue'
import DataTable from '../ui/DataTable.vue'
import SelectControl from '../ui/SelectControl.vue'

const props = defineProps({
  processes: { type: Array, default: () => [] },
  previousProcesses: { type: Array, default: () => [] },
  limit: { type: Number, default: 12 }
})

const emit = defineEmits(['update:limit'])

const query = ref('')
const activeCategory = ref('')
const sortKey = ref('memory_mb')
const sortDir = ref('desc')
const limitOptions = [
  { label: '40', value: 40 },
  { label: '80', value: 80 },
  { label: 'All', value: 500 }
]

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
    const category = process.category || 'application'
    const label = process.category_label || category
    const current = map.get(category) || { category, label, count: 0, memory: 0, cpu: 0 }
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
  const rows = props.processes.filter((process) => {
    if (activeCategory.value && process.category !== activeCategory.value) return false
    if (!needle) return true
    return `${process.pid} ${process.name} ${process.user || ''} ${process.category || ''} ${process.category_label || ''}`.toLowerCase().includes(needle)
  })
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

function toggleCategory(category) {
  activeCategory.value = activeCategory.value === category ? '' : category
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
          <SelectControl
            :model-value="limit"
            :options="limitOptions"
            aria-label="Process row limit"
            @update:model-value="emit('update:limit', Number($event))"
          />
        </label>
        <input v-model="query" type="search" placeholder="Search processes" />
      </div>
    </div>
    <div class="process-categories">
      <button
        v-for="item in categorySummary"
        :key="item.category"
        type="button"
        :class="{ active: activeCategory === item.category }"
        @click="toggleCategory(item.category)"
      >
        <span>{{ item.label }}</span>
        <strong>{{ item.count }}</strong>
        <small>{{ item.memory.toFixed(0) }} MB</small>
      </button>
    </div>
    <div class="table-wrap">
      <DataTable :columns="columns" :rows="filtered" :sort-key="sortKey" :sort-dir="sortDir" @sort="sort">
        <tr v-for="process in filtered" :key="process.pid" :class="{ hot: process.cpu_percent > 20 }">
          <td>{{ process.pid }}</td>
          <td>{{ process.name }}</td>
          <td>
            <span class="category-pill" :title="process.category_reason || process.category_label">
              {{ process.category_label || process.category || 'Application' }}
            </span>
          </td>
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
