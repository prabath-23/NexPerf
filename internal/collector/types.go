package collector

import "time"

type MemoryInfo struct {
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

type DiskInfo struct {
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

type SystemSummary struct {
	CPUPercent float64    `json:"cpu_percent"`
	Memory     MemoryInfo `json:"memory"`
	Disk       DiskInfo   `json:"disk"`
	OS         string     `json:"os"`
	Arch       string     `json:"arch"`
	Hostname   string     `json:"hostname,omitempty"`
	Timestamp  time.Time  `json:"timestamp"`
}

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	MemoryMB   float64 `json:"memory_mb"`
	CPUPercent float64 `json:"cpu_percent,omitempty"`
	User       string  `json:"user,omitempty"`
}
