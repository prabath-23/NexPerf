package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	PollingIntervalSeconds int        `json:"polling_interval_seconds" toml:"polling_interval_seconds"`
	RetentionHours         int        `json:"retention_hours" toml:"retention_hours"`
	DashboardRefreshMs     int        `json:"dashboard_refresh_ms" toml:"dashboard_refresh_ms"`
	EnabledCollectors      []string   `json:"enabled_collectors" toml:"enabled_collectors"`
	StorageLimits          Limits     `json:"storage_limits" toml:"storage_limits"`
	ObservationWindows     []string   `json:"observation_windows" toml:"observation_windows"`
	UsageMode              string     `json:"usage_mode" toml:"usage_mode"`
	InsightThresholds      Thresholds `json:"insight_thresholds" toml:"insight_thresholds"`
	Path                   string     `json:"path,omitempty" toml:"-"`
}

type Limits struct {
	MaxScanDepth     int `json:"max_scan_depth" toml:"max_scan_depth"`
	MaxDirectoryRows int `json:"max_directory_rows" toml:"max_directory_rows"`
}

type Thresholds struct {
	CPUWarning      float64 `json:"cpu_warning" toml:"cpu_warning"`
	MemoryWarning   float64 `json:"memory_warning" toml:"memory_warning"`
	DiskWarning     float64 `json:"disk_warning" toml:"disk_warning"`
	ProcessMemoryMB float64 `json:"process_memory_mb" toml:"process_memory_mb"`
}

type ModeProfile struct {
	Name                   string     `json:"name"`
	Label                  string     `json:"label"`
	Description            string     `json:"description"`
	PollingIntervalSeconds int        `json:"polling_interval_seconds"`
	RetentionHours         int        `json:"retention_hours"`
	DashboardRefreshMs     int        `json:"dashboard_refresh_ms"`
	StorageLimits          Limits     `json:"storage_limits"`
	InsightThresholds      Thresholds `json:"insight_thresholds"`
}

func Default() Config {
	cfg := Config{
		PollingIntervalSeconds: 5,
		RetentionHours:         24,
		DashboardRefreshMs:     3000,
		EnabledCollectors:      []string{"system", "processes", "storage"},
		StorageLimits: Limits{
			MaxScanDepth:     2,
			MaxDirectoryRows: 24,
		},
		ObservationWindows: []string{"5m", "15m", "1h"},
		UsageMode:          "balanced",
		InsightThresholds: Thresholds{
			CPUWarning:      80,
			MemoryWarning:   80,
			DiskWarning:     80,
			ProcessMemoryMB: 500,
		},
	}
	return ApplyUsageMode(cfg)
}

func Profiles() map[string]ModeProfile {
	return map[string]ModeProfile{
		"quiet": {
			Name:                   "quiet",
			Label:                  "Quiet",
			Description:            "Lower sampling pressure for battery-sensitive or low-overhead monitoring.",
			PollingIntervalSeconds: 15,
			RetentionHours:         12,
			DashboardRefreshMs:     6000,
			StorageLimits:          Limits{MaxScanDepth: 2, MaxDirectoryRows: 40},
			InsightThresholds:      Thresholds{CPUWarning: 85, MemoryWarning: 85, DiskWarning: 85, ProcessMemoryMB: 700},
		},
		"balanced": {
			Name:                   "balanced",
			Label:                  "Balanced",
			Description:            "General-purpose workstation observability for everyday work.",
			PollingIntervalSeconds: 5,
			RetentionHours:         24,
			DashboardRefreshMs:     3000,
			StorageLimits:          Limits{MaxScanDepth: 3, MaxDirectoryRows: 80},
			InsightThresholds:      Thresholds{CPUWarning: 80, MemoryWarning: 80, DiskWarning: 80, ProcessMemoryMB: 500},
		},
		"developer": {
			Name:                   "developer",
			Label:                  "Developer",
			Description:            "More attention to build caches, package folders, terminals, containers, and IDE workloads.",
			PollingIntervalSeconds: 5,
			RetentionHours:         24,
			DashboardRefreshMs:     3000,
			StorageLimits:          Limits{MaxScanDepth: 4, MaxDirectoryRows: 100},
			InsightThresholds:      Thresholds{CPUWarning: 75, MemoryWarning: 78, DiskWarning: 80, ProcessMemoryMB: 450},
		},
		"diagnostic": {
			Name:                   "diagnostic",
			Label:                  "Diagnostic",
			Description:            "Higher-detail local inspection for short troubleshooting sessions.",
			PollingIntervalSeconds: 2,
			RetentionHours:         48,
			DashboardRefreshMs:     1500,
			StorageLimits:          Limits{MaxScanDepth: 5, MaxDirectoryRows: 160},
			InsightThresholds:      Thresholds{CPUWarning: 70, MemoryWarning: 75, DiskWarning: 75, ProcessMemoryMB: 350},
		},
	}
}

func UsageModes() []ModeProfile {
	modes := []ModeProfile{}
	for _, key := range []string{"quiet", "balanced", "developer", "diagnostic"} {
		modes = append(modes, Profiles()[key])
	}
	return modes
}

func ApplyUsageMode(cfg Config) Config {
	profile, ok := Profiles()[cfg.UsageMode]
	if !ok {
		cfg.UsageMode = "balanced"
		profile = Profiles()[cfg.UsageMode]
	}
	cfg.PollingIntervalSeconds = profile.PollingIntervalSeconds
	cfg.RetentionHours = profile.RetentionHours
	cfg.DashboardRefreshMs = profile.DashboardRefreshMs
	cfg.StorageLimits = profile.StorageLimits
	cfg.InsightThresholds = profile.InsightThresholds
	return cfg
}

func Load() (Config, error) {
	cfg := Default()
	paths, err := candidatePaths()
	if err != nil {
		return cfg, err
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return cfg, err
		}
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return cfg, err
		}
		cfg.Path = path
		return cfg, nil
	}
	cfg.Path = paths[len(paths)-1]
	return cfg, nil
}

func Save(cfg Config) (Config, error) {
	if err := Validate(cfg); err != nil {
		return cfg, err
	}
	if cfg.Path == "" {
		paths, err := candidatePaths()
		if err != nil {
			return cfg, err
		}
		cfg.Path = paths[len(paths)-1]
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return cfg, err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, os.WriteFile(cfg.Path, data, 0o644)
}

func Validate(cfg Config) error {
	if cfg.PollingIntervalSeconds < 1 || cfg.PollingIntervalSeconds > 300 {
		return errors.New("polling interval must be between 1 and 300 seconds")
	}
	if cfg.RetentionHours < 1 || cfg.RetentionHours > 720 {
		return errors.New("retention hours must be between 1 and 720")
	}
	if cfg.DashboardRefreshMs < 500 || cfg.DashboardRefreshMs > 60000 {
		return errors.New("dashboard refresh must be between 500 and 60000 ms")
	}
	if cfg.StorageLimits.MaxScanDepth < 1 || cfg.StorageLimits.MaxScanDepth > 6 {
		return errors.New("max scan depth must be between 1 and 6")
	}
	if cfg.StorageLimits.MaxDirectoryRows < 5 || cfg.StorageLimits.MaxDirectoryRows > 200 {
		return errors.New("max directory rows must be between 5 and 200")
	}
	if _, ok := Profiles()[cfg.UsageMode]; !ok {
		return errors.New("usage mode must be one of quiet, balanced, developer, diagnostic")
	}
	return nil
}

func PollingInterval(cfg Config) time.Duration {
	return time.Duration(cfg.PollingIntervalSeconds) * time.Second
}

func candidatePaths() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return []string{
		"/etc/nexperf/config.toml",
		filepath.Join(home, ".config", "nexperf", "config.toml"),
	}, nil
}
