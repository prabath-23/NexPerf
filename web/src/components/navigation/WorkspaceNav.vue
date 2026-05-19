<script setup>
defineProps({
  active: { type: String, required: true },
  items: { type: Array, required: true },
  mode: { type: String, default: 'topbar' }
})

const emit = defineEmits(['change'])

const icons = {
  overview: ['M4 13h3l2.2-5 4 9 2.1-4H20'],
  processes: ['M5 7h14', 'M5 12h14', 'M5 17h14'],
  storage: ['M4 7c0-1.7 3.6-3 8-3s8 1.3 8 3-3.6 3-8 3-8-1.3-8-3Z', 'M4 7v5c0 1.7 3.6 3 8 3s8-1.3 8-3V7', 'M4 12v5c0 1.7 3.6 3 8 3s8-1.3 8-3v-5'],
  actionables: ['M5 12l4 4L19 6', 'M5 19h14'],
  history: ['M4 12a8 8 0 1 0 2.35-5.65', 'M4 5v5h5', 'M12 8v5l3 2'],
  terminal: ['M4 6h16v12H4Z', 'M7 10l2 2-2 2', 'M12 14h4'],
  manual: [
    'M6 3h8l4 4v14H6Z',
    'M14 3v5h5',
    'M9 11h6',
    'M9 15h6',
    'M9 19h4'
  ],
  network: [
    'M6 9.5a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z',
    'M18 9.5a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z',
    'M12 20a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z',
    'M8.2 8.2l2.6 6.1',
    'M15.8 8.2l-2.6 6.1',
    'M8.5 7h7'
  ],
  diagnostics: ['M12 3l2.2 5.4 5.8.5-4.4 3.8 1.3 5.7-4.9-3-4.9 3 1.3-5.7L4 8.9l5.8-.5Z'],
  nexperf: ['M4 16l5-10 4 8 3-6 4 10'],
  settings: [
    'M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.38a2 2 0 0 0-.73-2.73l-.15-.09a2 2 0 0 1-1-1.74v-.51a2 2 0 0 1 1-1.72l.15-.1a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2Z',
    'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z'
  ]
}

function iconFor(item) {
  return icons[item.icon] || icons.overview
}
</script>

<template>
  <nav class="workspace-nav" :class="`mode-${mode}`" aria-label="NexPerf workspace">
    <button
      v-for="item in items"
      :key="item.id"
      type="button"
      :class="{ active: active === item.id }"
      @click="emit('change', item.id)"
      :title="mode === 'sidebar' ? item.label : undefined"
      :aria-current="active === item.id ? 'page' : undefined"
    >
      <svg class="nav-icon" viewBox="0 0 24 24" aria-hidden="true">
        <path v-for="path in iconFor(item)" :key="path" :d="path" />
      </svg>
      <span class="nav-label">{{ item.label }}</span>
      <small v-if="item.soon">soon</small>
    </button>
  </nav>
</template>
