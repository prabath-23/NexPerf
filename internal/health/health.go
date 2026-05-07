package health

import (
	"math"
	"time"

	"github.com/prabath/nexperf/internal/collector"
	"github.com/prabath/nexperf/internal/storage"
)

type Score struct {
	Value       int       `json:"value"`
	Status      string    `json:"status"`
	Summary     string    `json:"summary"`
	GeneratedAt time.Time `json:"generated_at"`
	Factors     []Factor  `json:"factors"`
}

type Factor struct {
	ID      string  `json:"id"`
	Label   string  `json:"label"`
	Impact  int     `json:"impact"`
	Value   float64 `json:"value"`
	Message string  `json:"message"`
}

func Calculate(system collector.SystemSummary, processes []collector.ProcessInfo, history []storage.MetricSample) Score {
	factors := []Factor{}
	add := func(id, label string, value float64, impact int, message string) {
		if impact > 0 {
			factors = append(factors, Factor{ID: id, Label: label, Value: value, Impact: impact, Message: message})
		}
	}

	add("cpu", "CPU pressure", system.CPUPercent, pressureImpact(system.CPUPercent, 28), "Current CPU utilization is affecting the score.")
	add("memory", "Memory pressure", system.Memory.Percent, pressureImpact(system.Memory.Percent, 34), "Memory pressure is the largest stability signal.")
	add("disk", "Disk capacity", system.Disk.Percent, pressureImpact(system.Disk.Percent, 24), "Full disks reduce headroom for builds, caches, logs, and updates.")

	heavyProcesses := 0
	for _, proc := range processes {
		if proc.CPUPercent > 20 || proc.MemoryMB > 500 {
			heavyProcesses++
		}
	}
	if heavyProcesses > 0 {
		add("processes", "Process anomalies", float64(heavyProcesses), int(math.Min(float64(heavyProcesses*4), 14)), "Heavy or spiking processes are active.")
	}

	if sustained(history, func(sample storage.MetricSample) float64 { return sample.MemoryPct }, 80) {
		add("memory-sustained", "Sustained memory", system.Memory.Percent, 8, "Memory has stayed above threshold across recent samples.")
	}
	if sustained(history, func(sample storage.MetricSample) float64 { return sample.CPUPercent }, 80) {
		add("cpu-sustained", "Sustained CPU", system.CPUPercent, 8, "CPU has stayed above threshold across recent samples.")
	}

	totalImpact := 0
	for _, factor := range factors {
		totalImpact += factor.Impact
	}
	value := int(math.Max(0, 100-float64(totalImpact)))
	status := "steady"
	summary := "System has comfortable operating headroom."
	if value < 70 {
		status = "watch"
		summary = "System needs attention; one or more resources are under pressure."
	}
	if value < 45 {
		status = "strained"
		summary = "System is strained and may affect local development workflows."
	}
	return Score{Value: value, Status: status, Summary: summary, GeneratedAt: time.Now(), Factors: factors}
}

func pressureImpact(value float64, maxImpact int) int {
	if value <= 60 {
		return 0
	}
	ratio := math.Min((value-60)/40, 1)
	return int(math.Round(ratio * float64(maxImpact)))
}

func sustained(history []storage.MetricSample, pick func(storage.MetricSample) float64, threshold float64) bool {
	if len(history) < 3 {
		return false
	}
	count := 0
	for i := len(history) - 1; i >= 0 && count < 6; i-- {
		if pick(history[i]) < threshold {
			return false
		}
		count++
	}
	return count >= 3
}
