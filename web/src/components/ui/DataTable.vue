<script setup>
defineProps({
  columns: { type: Array, required: true },
  rows: { type: Array, required: true },
  sortKey: { type: String, default: '' },
  sortDir: { type: String, default: 'desc' }
})

const emit = defineEmits(['sort'])
</script>

<template>
  <table class="data-table">
    <thead>
      <tr>
        <th v-for="column in columns" :key="column.key">
          <button v-if="column.sortable" type="button" @click="emit('sort', column.key)">
            {{ column.label }}
            <span v-if="sortKey === column.key">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
          </button>
          <span v-else>{{ column.label }}</span>
        </th>
      </tr>
    </thead>
    <tbody>
      <slot />
    </tbody>
  </table>
</template>
