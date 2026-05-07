package insight

import (
	"github.com/prabath/nexperf/internal/collector"
)

type Insight struct {
	ID             string `json:"id"`
	Severity       string `json:"severity"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
}

func Generate(system collector.SystemSummary) []Insight {
	insights := []Insight{}

	if system.Memory.Percent > 80 {
		insights = append(insights, Insight{
			ID:             "memory-high",
			Severity:       severity(system.Memory.Percent),
			Title:          "High memory usage",
			Message:        "Memory usage is above 80%, which can increase swap activity and slow interactive workloads.",
			Recommendation: "Review top memory processes and close or restart anything that is unexpectedly large.",
		})
	}

	if system.Disk.Percent > 80 {
		insights = append(insights, Insight{
			ID:             "disk-high",
			Severity:       severity(system.Disk.Percent),
			Title:          "High disk usage",
			Message:        "The root disk is above 80% capacity, which can affect builds, logs, caches, and OS updates.",
			Recommendation: "Clear old build artifacts, package caches, downloads, or move large files off the root volume.",
		})
	}

	if system.CPUPercent > 80 {
		insights = append(insights, Insight{
			ID:             "cpu-high",
			Severity:       severity(system.CPUPercent),
			Title:          "High CPU usage",
			Message:        "CPU usage is above 80%, which may make local development tools feel sluggish.",
			Recommendation: "Inspect active workloads and wait for expected build, indexing, or container activity to complete.",
		})
	}

	if len(insights) == 0 {
		insights = append(insights, Insight{
			ID:             "system-normal",
			Severity:       "info",
			Title:          "System resources look steady",
			Message:        "CPU, memory, and disk usage are below the v0.1 warning thresholds.",
			Recommendation: "Keep NexPerf running while reproducing slowdowns to catch spikes as they happen.",
		})
	}

	return insights
}

func severity(percent float64) string {
	if percent >= 90 {
		return "critical"
	}
	return "warning"
}
