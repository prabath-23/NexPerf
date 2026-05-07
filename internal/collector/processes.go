package collector

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

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
			Category:   CategoryForProcess(name),
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
			Category: CategoryForProcess(name),
		},
	}
}

func CategoryForProcess(name string) string {
	normalized := strings.ToLower(name)
	switch {
	case strings.Contains(normalized, "chrome") || strings.Contains(normalized, "safari") || strings.Contains(normalized, "webkit") || strings.Contains(normalized, "firefox") || strings.Contains(normalized, "browser"):
		return "browser"
	case strings.Contains(normalized, "code") || strings.Contains(normalized, "cursor") || strings.Contains(normalized, "intellij") || strings.Contains(normalized, "goland") || strings.Contains(normalized, "xcode") || strings.Contains(normalized, "codex"):
		return "ide"
	case strings.Contains(normalized, "terminal") || strings.Contains(normalized, "iterm") || strings.Contains(normalized, "zsh") || strings.Contains(normalized, "bash") || strings.Contains(normalized, "fish"):
		return "terminal"
	case strings.Contains(normalized, "docker") || strings.Contains(normalized, "container") || strings.Contains(normalized, "colima") || strings.Contains(normalized, "podman"):
		return "container"
	case strings.Contains(normalized, "helper") || strings.Contains(normalized, "agent") || strings.Contains(normalized, "service") || strings.Contains(normalized, "daemon"):
		return "service"
	case strings.HasPrefix(normalized, "com.apple") || strings.Contains(normalized, "kernel") || strings.Contains(normalized, "launchd") || strings.Contains(normalized, "system"):
		return "system"
	default:
		return "app"
	}
}
