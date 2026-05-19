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
	PID            int32   `json:"pid"`
	PPID           int32   `json:"ppid,omitempty"`
	Name           string  `json:"name"`
	MemoryMB       float64 `json:"memory_mb"`
	CPUPercent     float64 `json:"cpu_percent,omitempty"`
	User           string  `json:"user,omitempty"`
	Category       string  `json:"category"`
	CategoryLabel  string  `json:"category_label,omitempty"`
	CategoryReason string  `json:"category_reason,omitempty"`
	Threads        int32   `json:"threads,omitempty"`
	CreateTime     int64   `json:"create_time,omitempty"`
	Runtime        string  `json:"runtime,omitempty"`
}
