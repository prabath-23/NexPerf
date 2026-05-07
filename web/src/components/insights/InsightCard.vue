<script setup>
defineProps({
  insight: { type: Object, required: true }
})
</script>

<template>
  <article class="insight-card" :class="insight.severity">
    <div>
      <span>{{ insight.category }} · {{ insight.state || 'active' }}</span>
      <time>{{ new Date(insight.timestamp).toLocaleTimeString() }}</time>
    </div>
    <strong>{{ insight.title }}</strong>
    <p>{{ insight.message }}</p>
    <footer>
      <span>{{ insight.recommendation }}</span>
      <em v-if="insight.score">score {{ insight.score }}</em>
    </footer>
  </article>
</template>
