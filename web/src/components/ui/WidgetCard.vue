<script setup>
defineProps({
  size: { type: String, default: 'small' },
  tone: { type: String, default: 'neutral' },
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  description: { type: String, default: '' },
  draggable: { type: Boolean, default: false }
})

const emit = defineEmits(['dragstart', 'dragover', 'drop'])
</script>

<template>
  <article
    class="widget-card"
    :class="[`widget-${size}`, `tone-${tone}`, { draggable }]"
    :draggable="draggable"
    @dragstart="emit('dragstart', $event)"
    @dragover.prevent="emit('dragover', $event)"
    @drop.prevent="emit('drop', $event)"
  >
    <header v-if="$slots.header || title" class="widget-chrome">
      <slot name="header">
        <div class="widget-title-block">
          <span>{{ title }}</span>
          <small v-if="subtitle">{{ subtitle }}</small>
        </div>
      </slot>
      <div v-if="$slots.actions" class="widget-actions">
        <slot name="actions" />
      </div>
    </header>
    <div class="widget-content">
      <slot />
    </div>
    <footer v-if="$slots.footer" class="widget-footer">
      <slot name="footer" />
    </footer>
  </article>
</template>
