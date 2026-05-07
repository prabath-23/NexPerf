const json = async (path) => {
  const response = await fetch(path)
  if (!response.ok) {
    throw new Error(`${path} returned ${response.status}`)
  }
  return response.json()
}

export function getSystem() {
  return json('/api/system')
}

export function getProcesses(limit = 12) {
  return json(`/api/processes/top?limit=${limit}`)
}

export function getInsights() {
  return json('/api/insights')
}

export function getHealthScore() {
  return json('/api/health-score')
}

export function getHistory(metric, range = '5m') {
  return json(`/api/history/${metric}?range=${range}`)
}
