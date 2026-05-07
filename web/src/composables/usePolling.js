import { onMounted, onUnmounted, ref } from 'vue'

export function usePolling(callback, interval = 3000) {
  const loading = ref(true)
  const refreshing = ref(false)
  const error = ref('')
  const lastUpdated = ref(null)
  let timer

  async function refresh() {
    refreshing.value = true
    try {
      await callback()
      error.value = ''
      lastUpdated.value = new Date()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unable to refresh NexPerf data'
    } finally {
      loading.value = false
      refreshing.value = false
    }
  }

  onMounted(() => {
    refresh()
    timer = window.setInterval(refresh, interval)
  })

  onUnmounted(() => window.clearInterval(timer))

  return { loading, refreshing, error, lastUpdated, refresh }
}
