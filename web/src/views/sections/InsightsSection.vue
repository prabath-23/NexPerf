<script setup>
import InsightCard from '../../components/insights/InsightCard.vue'
import { dashboardStore } from '../../stores/dashboardStore'

defineProps({
  actionableItems: { type: Array, default: () => [] }
})

const emit = defineEmits(['open-action-target', 'refresh', 'stage-action'])
</script>

<template>
  <section class="panel insights-panel wide">
    <div class="section-head">
      <div>
        <h2>Actionables</h2>
        <span>Operational next steps staged from runtime, storage, and process pressure</span>
      </div>
      <button type="button" @click="emit('refresh')">Refresh</button>
    </div>
    <div class="actionable-grid">
      <article v-for="action in actionableItems" :key="action.id">
        <div>
          <strong>{{ action.title }}</strong>
          <small v-if="action.count > 1">{{ action.count }} observations</small>
          <span>{{ action.detail }}</span>
          <code>{{ action.command }}</code>
        </div>
        <footer>
          <button type="button" @click="emit('open-action-target', action)">{{ action.primary }}</button>
          <button type="button" @click="emit('stage-action', action)">Stage command</button>
        </footer>
      </article>
    </div>
    <div class="insight-grid">
      <InsightCard v-for="item in dashboardStore.insights" :key="item.id" :insight="item" />
    </div>
  </section>
</template>
