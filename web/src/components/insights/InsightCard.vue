<script setup>
defineProps({
  insight: { type: Object, required: true }
})
</script>

<template>
  <article class="insight-card" :class="insight.severity">
    <span class="insight-rail" aria-hidden="true"></span>
    <div class="insight-body">
      <div class="insight-meta">
        <span>{{ insight.category }}</span>
        <time>{{ new Date(insight.timestamp).toLocaleTimeString() }}</time>
      </div>
      <div class="insight-main">
        <strong>{{ insight.title }}</strong>
        <em v-if="insight.score">score {{ insight.score }}</em>
      </div>
      <p>{{ insight.message }}</p>
      <footer>
        <span>Recommendation</span>
        {{ insight.recommendation }}
      </footer>
    </div>
  </article>
</template>
