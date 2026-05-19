<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, default: () => [] },
  ariaLabel: { type: String, default: 'Select option' }
})

const emit = defineEmits(['update:modelValue', 'change'])
const open = ref(false)
const root = ref(null)

const selected = computed(() => props.options.find((option) => option.value === props.modelValue) || props.options[0])

function choose(option) {
  emit('update:modelValue', option.value)
  emit('change', option.value)
  open.value = false
}

function onDocumentClick(event) {
  if (root.value && !root.value.contains(event.target)) open.value = false
}

function onKeydown(event) {
  if (event.key === 'Escape') open.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div ref="root" class="select-control" :class="{ open }">
    <button
      type="button"
      class="select-trigger"
      :aria-label="ariaLabel"
      :aria-expanded="open"
      @click="open = !open"
    >
      <span>{{ selected?.label || 'Select' }}</span>
      <svg viewBox="0 0 16 16" aria-hidden="true">
        <path d="M4 6l4 4 4-4" />
      </svg>
    </button>
    <div v-if="open" class="select-menu" role="listbox">
      <button
        v-for="option in options"
        :key="option.value"
        type="button"
        role="option"
        :aria-selected="option.value === modelValue"
        :class="{ selected: option.value === modelValue }"
        @click="choose(option)"
      >
        <span>{{ option.label }}</span>
        <small v-if="option.detail">{{ option.detail }}</small>
      </button>
    </div>
  </div>
</template>
