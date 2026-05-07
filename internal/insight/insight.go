package insight

import (
	"fmt"
	"strings"
	"time"

	"github.com/prabath/nexperf/internal/collector"
	"github.com/prabath/nexperf/internal/storage"
)

type Insight struct {
	ID             string    `json:"id"`
	Severity       string    `json:"severity"`
	Category       string    `json:"category"`
	Score          int       `json:"score"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	Recommendation string    `json:"recommendation"`
	Timestamp      time.Time `json:"timestamp"`
}

func Generate(system collector.SystemSummary) []Insight {
	return GenerateContextual(system, nil, nil)
}

func GenerateContextual(system collector.SystemSummary, processes []collector.ProcessInfo, history []storage.MetricSample) []Insight {
	insights := []Insight{}
	now := time.Now()

	if system.Memory.Percent > 80 {
		processHint := memoryProcessHint(processes)
		insights = append(insights, Insight{
			ID:             "memory-high",
			Severity:       severity(system.Memory.Percent),
			Category:       "Performance",
			Score:          score(system.Memory.Percent, 34),
			Title:          "High memory usage",
			Message:        fmt.Sprintf("Memory usage is %.1f%%. High memory pressure can increase swap activity and slow interactive workloads.%s", system.Memory.Percent, processHint),
			Recommendation: "Review top memory processes and close or restart anything that is unexpectedly large.",
			Timestamp:      now,
		})
	}

	if system.Disk.Percent > 80 {
		insights = append(insights, Insight{
			ID:             "disk-high",
			Severity:       severity(system.Disk.Percent),
			Category:       "Storage",
			Score:          score(system.Disk.Percent, 28),
			Title:          "High disk usage",
			Message:        fmt.Sprintf("Disk usage is %.1f%% on the root filesystem, which may impact builds, caches, logs, databases, and OS updates.", system.Disk.Percent),
			Recommendation: "Clear old build artifacts, package caches, downloads, or move large files off the root volume.",
			Timestamp:      now,
		})
	}

	if system.CPUPercent > 80 {
		insights = append(insights, Insight{
			ID:             "cpu-high",
			Severity:       severity(system.CPUPercent),
			Category:       "Performance",
			Score:          score(system.CPUPercent, 30),
			Title:          "High CPU usage",
			Message:        fmt.Sprintf("CPU usage is %.1f%%, which may make local development tools feel sluggish.", system.CPUPercent),
			Recommendation: "Inspect active workloads and wait for expected build, indexing, or container activity to complete.",
			Timestamp:      now,
		})
	}

	if sustainedCPU(history) {
		insights = append(insights, Insight{
			ID:             "cpu-sustained",
			Severity:       "warning",
			Category:       "Performance",
			Score:          26,
			Title:          "Sustained CPU pressure",
			Message:        "CPU usage has remained elevated across recent samples rather than appearing as a single spike.",
			Recommendation: "Sort processes by CPU and check for long-running build, indexing, or container workloads.",
			Timestamp:      now,
		})
	}

	if sustainedMemory(history) {
		duration := sustainedDuration(history, func(sample storage.MetricSample) float64 { return sample.MemoryPct }, 80)
		insights = append(insights, Insight{
			ID:             "memory-sustained",
			Severity:       "warning",
			Category:       "Performance",
			Score:          30,
			Title:          "Memory pressure is persistent",
			Message:        fmt.Sprintf("Memory pressure has remained above 80%% for roughly %s across recent samples.", duration),
			Recommendation: "Review browsers, IDE helpers, and background services before starting memory-heavy work.",
			Timestamp:      now,
		})
	}

	if diskGrowth(history) > 1.0 {
		insights = append(insights, Insight{
			ID:             "disk-growing",
			Severity:       "info",
			Category:       "Storage",
			Score:          14,
			Title:          "Disk usage is trending upward",
			Message:        fmt.Sprintf("Disk usage increased by %.1f percentage points over the selected history window.", diskGrowth(history)),
			Recommendation: "Watch build artifacts, caches, downloads, and logs if this trend continues.",
			Timestamp:      now,
		})
	}

	insights = append(insights, processInsights(processes, system.Memory.Total, now)...)

	if len(insights) == 0 {
		insights = append(insights, Insight{
			ID:             "system-normal",
			Severity:       "info",
			Category:       "System Health",
			Score:          1,
			Title:          "System resources look steady",
			Message:        "CPU, memory, and disk usage are below the v0.2 warning thresholds.",
			Recommendation: "Keep NexPerf running while reproducing slowdowns to catch spikes as they happen.",
			Timestamp:      now,
		})
	}

	sortInsights(insights)
	return insights
}

func severity(percent float64) string {
	if percent >= 90 {
		return "critical"
	}
	return "warning"
}

func score(percent float64, max int) int {
	if percent <= 60 {
		return 1
	}
	value := int(((percent - 60) / 40) * float64(max))
	if value < 1 {
		return 1
	}
	if value > max {
		return max
	}
	return value
}

func sustainedCPU(history []storage.MetricSample) bool {
	if len(history) < 3 {
		return false
	}
	count := 0
	for i := len(history) - 1; i >= 0 && count < 6; i-- {
		if history[i].CPUPercent <= 80 {
			return false
		}
		count++
	}
	return count >= 3
}

func sustainedMemory(history []storage.MetricSample) bool {
	return sustainedDuration(history, func(sample storage.MetricSample) float64 { return sample.MemoryPct }, 80) != "0s"
}

func sustainedDuration(history []storage.MetricSample, pick func(storage.MetricSample) float64, threshold float64) string {
	if len(history) < 2 {
		return "0s"
	}
	latest := time.Time{}
	earliest := time.Time{}
	for i := len(history) - 1; i >= 0; i-- {
		if pick(history[i]) < threshold {
			break
		}
		if latest.IsZero() {
			latest = history[i].Timestamp
		}
		earliest = history[i].Timestamp
	}
	if latest.IsZero() || earliest.IsZero() || latest.Equal(earliest) {
		return "0s"
	}
	return latest.Sub(earliest).Round(time.Second).String()
}

func diskGrowth(history []storage.MetricSample) float64 {
	if len(history) < 2 {
		return 0
	}
	return history[len(history)-1].DiskPct - history[0].DiskPct
}

func processInsights(processes []collector.ProcessInfo, memoryTotal uint64, now time.Time) []Insight {
	if len(processes) == 0 {
		return nil
	}
	byCategory := map[string]struct {
		memory float64
		cpu    float64
		count  int
	}{}
	for _, proc := range processes {
		item := byCategory[proc.Category]
		item.memory += proc.MemoryMB
		item.cpu += proc.CPUPercent
		item.count++
		byCategory[proc.Category] = item
	}

	insights := []Insight{}
	totalMB := float64(memoryTotal) / 1024 / 1024
	for category, item := range byCategory {
		if item.count < 2 {
			continue
		}
		memoryShare := 0.0
		if totalMB > 0 {
			memoryShare = item.memory / totalMB * 100
		}
		if memoryShare >= 8 || item.cpu >= 20 {
			title := fmt.Sprintf("%s processes are prominent", strings.Title(category))
			message := fmt.Sprintf("%d %s processes account for %.1f%% of physical memory and %.1f%% sampled CPU.", item.count, category, memoryShare, item.cpu)
			insights = append(insights, Insight{
				ID:             "process-" + category,
				Severity:       "info",
				Category:       "Processes",
				Score:          int(memoryShare) + int(item.cpu/4),
				Title:          title,
				Message:        message,
				Recommendation: "Use the process table categories and trend markers to identify whether this is expected workload activity.",
				Timestamp:      now,
			})
		}
	}
	return insights
}

func sortInsights(insights []Insight) {
	for i := 0; i < len(insights); i++ {
		for j := i + 1; j < len(insights); j++ {
			if insights[j].Score > insights[i].Score {
				insights[i], insights[j] = insights[j], insights[i]
			}
		}
	}
}

func memoryProcessHint(processes []collector.ProcessInfo) string {
	counts := map[string]int{}
	for _, proc := range processes {
		name := strings.ToLower(proc.Name)
		switch {
		case strings.Contains(name, "chrome"):
			counts["Chrome"]++
		case strings.Contains(name, "webkit") || strings.Contains(name, "safari"):
			counts["Safari/WebKit"]++
		case strings.Contains(name, "code"):
			counts["VS Code"]++
		case strings.Contains(name, "docker"):
			counts["Docker"]++
		}
	}
	for label, count := range counts {
		if count >= 2 {
			return fmt.Sprintf(" Several high-memory entries are %s-related processes.", label)
		}
	}
	return ""
}
