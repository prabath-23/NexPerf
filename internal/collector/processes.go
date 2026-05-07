package collector

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"

	gprocess "github.com/shirou/gopsutil/v4/process"
)

func TopProcesses(limit int) ([]ProcessInfo, error) {
	if limit <= 0 {
		limit = 10
	}

	processes, err := gprocess.Processes()
	if err != nil {
		return fallbackCurrentProcess(), nil
	}

	items := make([]ProcessInfo, 0, len(processes))
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil || name == "" {
			continue
		}

		memInfo, err := proc.MemoryInfo()
		if err != nil {
			continue
		}

		cpuPercent, _ := proc.CPUPercent()
		user, _ := proc.Username()

		items = append(items, ProcessInfo{
			PID:        proc.Pid,
			Name:       name,
			MemoryMB:   float64(memInfo.RSS) / 1024 / 1024,
			CPUPercent: cpuPercent,
			User:       user,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].MemoryMB > items[j].MemoryMB
	})

	if len(items) > limit {
		items = items[:limit]
	}

	if len(items) == 0 {
		return fallbackCurrentProcess(), nil
	}

	return items, nil
}

func fallbackCurrentProcess() []ProcessInfo {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	name := "nexperf"
	if executable, err := os.Executable(); err == nil {
		name = filepath.Base(executable)
	}

	return []ProcessInfo{
		{
			PID:      int32(os.Getpid()),
			Name:     name,
			MemoryMB: float64(mem.Sys) / 1024 / 1024,
		},
	}
}
