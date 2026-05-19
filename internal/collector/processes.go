package collector

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

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
		ppid, _ := proc.Ppid()
		threads, _ := proc.NumThreads()
		createTime, _ := proc.CreateTime()
		exe, _ := proc.Exe()
		classification := ClassifyProcess(name, exe, user)

		items = append(items, ProcessInfo{
			PID:            proc.Pid,
			PPID:           ppid,
			Name:           name,
			MemoryMB:       float64(memInfo.RSS) / 1024 / 1024,
			CPUPercent:     cpuPercent,
			User:           user,
			Category:       classification.ID,
			CategoryLabel:  classification.Label,
			CategoryReason: classification.Reason,
			Threads:        threads,
			CreateTime:     createTime,
			Runtime:        runtimeSince(createTime),
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

func ProcessDetail(pid int32) (ProcessInfo, error) {
	proc, err := gprocess.NewProcess(pid)
	if err != nil {
		return ProcessInfo{}, err
	}
	name, err := proc.Name()
	if err != nil {
		return ProcessInfo{}, err
	}
	memInfo, _ := proc.MemoryInfo()
	memoryMB := 0.0
	if memInfo != nil {
		memoryMB = float64(memInfo.RSS) / 1024 / 1024
	}
	cpuPercent, _ := proc.CPUPercent()
	user, _ := proc.Username()
	ppid, _ := proc.Ppid()
	threads, _ := proc.NumThreads()
	createTime, _ := proc.CreateTime()
	exe, _ := proc.Exe()
	classification := ClassifyProcess(name, exe, user)
	return ProcessInfo{
		PID:            proc.Pid,
		PPID:           ppid,
		Name:           name,
		MemoryMB:       memoryMB,
		CPUPercent:     cpuPercent,
		User:           user,
		Category:       classification.ID,
		CategoryLabel:  classification.Label,
		CategoryReason: classification.Reason,
		Threads:        threads,
		CreateTime:     createTime,
		Runtime:        runtimeSince(createTime),
	}, nil
}

func ProcessTree(limit int) ([]ProcessInfo, error) {
	items, err := TopProcesses(limit)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].PPID == items[j].PPID {
			return items[i].MemoryMB > items[j].MemoryMB
		}
		return items[i].PPID < items[j].PPID
	})
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
			PID:            int32(os.Getpid()),
			Name:           name,
			MemoryMB:       float64(mem.Sys) / 1024 / 1024,
			Category:       CategoryForProcess(name),
			CategoryLabel:  ClassifyProcess(name, "", "").Label,
			CategoryReason: ClassifyProcess(name, "", "").Reason,
		},
	}
}

func runtimeSince(createTime int64) string {
	if createTime <= 0 {
		return ""
	}
	start := time.UnixMilli(createTime)
	if start.After(time.Now()) {
		return ""
	}
	return time.Since(start).Round(time.Second).String()
}

func CategoryForProcess(name string) string {
	return ClassifyProcess(name, "", "").ID
}

type ProcessClassification struct {
	ID     string
	Label  string
	Reason string
}

func ClassifyProcess(name string, exe string, user string) ProcessClassification {
	normalized := strings.ToLower(name)
	context := strings.ToLower(name + " " + exe + " " + user)
	switch {
	case containsAny(context, "chrome", "safari", "webkit", "firefox", "edge", "browser", "arc", "brave"):
		return ProcessClassification{ID: "browser", Label: "Browser", Reason: "Browser or renderer process"}
	case containsAny(context, "slack", "teams", "zoom", "meet", "discord", "telegram", "whatsapp", "facetime"):
		return ProcessClassification{ID: "communication", Label: "Communication", Reason: "Messaging, meeting, or collaboration workload"}
	case containsAny(context, "word", "excel", "powerpoint", "keynote", "pages", "numbers", "notion", "obsidian", "notes", "preview", "acrobat"):
		return ProcessClassification{ID: "productivity", Label: "Productivity", Reason: "Documents, notes, or office application"}
	case containsAny(context, "photoshop", "illustrator", "figma", "sketch", "final cut", "logic pro", "garageband", "blender", "resolve"):
		return ProcessClassification{ID: "creative", Label: "Creative", Reason: "Design, media, or creative production workload"}
	case containsAny(context, "code", "cursor", "intellij", "goland", "xcode", "codex", "eclipse", "android studio", "pycharm", "webstorm"):
		return ProcessClassification{ID: "development", Label: "Development", Reason: "Editor, IDE, or coding assistant"}
	case containsAny(context, "terminal", "iterm", "warp", "zsh", "bash", "fish", "tmux", "ssh"):
		return ProcessClassification{ID: "terminal", Label: "Terminal", Reason: "Shell or terminal session"}
	case containsAny(context, "docker", "container", "colima", "podman", "kubernetes", "kubectl", "containerd"):
		return ProcessClassification{ID: "container", Label: "Containers", Reason: "Container or local infrastructure workload"}
	case containsAny(context, "postgres", "mysql", "redis", "mongodb", "sqlite", "database"):
		return ProcessClassification{ID: "database", Label: "Database", Reason: "Local database or datastore process"}
	case containsAny(context, "dropbox", "onedrive", "icloud", "drive", "sync"):
		return ProcessClassification{ID: "sync", Label: "Sync", Reason: "File sync or cloud storage helper"}
	case containsAny(context, "security", "defender", "sentinel", "crowdstrike", "falcon", "jamf", "vpn"):
		return ProcessClassification{ID: "security", Label: "Security", Reason: "Security, device management, or VPN agent"}
	case containsAny(context, "music", "spotify", "vlc", "quicktime", "media"):
		return ProcessClassification{ID: "media", Label: "Media", Reason: "Playback or media utility"}
	case containsAny(context, "helper", "agent", "service", "daemon", "updater"):
		return ProcessClassification{ID: "service", Label: "Service", Reason: "Background helper or service process"}
	case strings.HasPrefix(normalized, "com.apple") || strings.Contains(normalized, "kernel") || strings.Contains(normalized, "launchd") || strings.Contains(normalized, "system"):
		return ProcessClassification{ID: "system", Label: "System", Reason: "Operating system process"}
	default:
		return ProcessClassification{ID: "application", Label: "Application", Reason: "User or application workload"}
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
