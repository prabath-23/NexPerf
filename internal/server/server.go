package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prabath/nexperf/internal/collector"
	"github.com/prabath/nexperf/internal/config"
	"github.com/prabath/nexperf/internal/health"
	"github.com/prabath/nexperf/internal/insight"
	"github.com/prabath/nexperf/internal/service"
	"github.com/prabath/nexperf/internal/storage"
	"github.com/prabath/nexperf/internal/storageintel"
	"github.com/prabath/nexperf/internal/version"
)

type Server struct {
	startedAt time.Time
	store     *storage.Store
	config    config.Config
	dbPath    string
	http      *http.Server
	mu        sync.Mutex
	timings   map[string]timingStat
}

type timingStat struct {
	Count       int           `json:"count"`
	Last        time.Duration `json:"last"`
	Average     time.Duration `json:"average"`
	Total       time.Duration `json:"-"`
	LastUpdated time.Time     `json:"last_updated"`
}

func New(store *storage.Store) *Server {
	cfg, _ := config.Load()
	return NewWithConfig(store, cfg, "")
}

func NewWithConfig(store *storage.Store, cfg config.Config, dbPath string) *Server {
	return &Server{startedAt: time.Now(), store: store, config: cfg, dbPath: dbPath, timings: map[string]timingStat{}}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/system", s.system)
	mux.HandleFunc("/api/processes/top", s.processesTop)
	mux.HandleFunc("/api/processes/detail", s.processDetail)
	mux.HandleFunc("/api/processes/tree", s.processTree)
	mux.HandleFunc("/api/insights", s.insights)
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/health-score", s.healthScore)
	mux.HandleFunc("/api/storage/summary", s.storageSummary)
	mux.HandleFunc("/api/dashboard/widgets", s.dashboardWidgets)
	mux.HandleFunc("/api/config", s.configHandler)
	mux.HandleFunc("/api/config/modes", s.configModes)
	mux.HandleFunc("/api/terminal/exec", s.terminalExec)
	mux.HandleFunc("/api/nexperf", s.self)
	mux.HandleFunc("/api/history/cpu", s.history("cpu"))
	mux.HandleFunc("/api/history/memory", s.history("memory"))
	mux.HandleFunc("/api/history/disk", s.history("disk"))
	mux.Handle("/nexperf", s.dashboard())
	mux.Handle("/nexperf/", s.dashboard())
	return s.observe(mux)
}

func (s *Server) ListenAndServe(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	s.http = &http.Server{Addr: addr, Handler: s.Handler()}
	err := s.http.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

func (s *Server) system(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	system, err := collector.CollectSystem()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, system)
}

func (s *Server) processesTop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	limit := 12
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil {
			limit = parsed
		}
	}
	processes, err := collector.TopProcesses(limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, processes)
}

func (s *Server) processDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	raw := r.URL.Query().Get("pid")
	pid, err := strconv.Atoi(raw)
	if err != nil || pid <= 0 {
		http.Error(w, "pid is required", http.StatusBadRequest)
		return
	}
	detail, err := collector.ProcessDetail(int32(pid))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, detail)
}

func (s *Server) processTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	limit := 80
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	tree, err := collector.ProcessTree(limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, tree)
}

func (s *Server) healthScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	system, err := collector.CollectSystem()
	if err != nil {
		writeError(w, err)
		return
	}
	processes, _ := collector.TopProcesses(20)
	history, _ := s.recentSamples(72)
	writeJSON(w, health.CalculateWithThresholds(system, processes, history, health.Thresholds{
		CPUWarning:      s.config.InsightThresholds.CPUWarning,
		MemoryWarning:   s.config.InsightThresholds.MemoryWarning,
		DiskWarning:     s.config.InsightThresholds.DiskWarning,
		ProcessMemoryMB: s.config.InsightThresholds.ProcessMemoryMB,
	}))
}

func (s *Server) insights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	system, err := collector.CollectSystem()
	if err != nil {
		writeError(w, err)
		return
	}
	processes, _ := collector.TopProcesses(12)
	history, _ := s.recentSamples(12)
	writeJSON(w, insight.GenerateWithThresholds(system, processes, history, insight.Thresholds{
		CPUWarning:    s.config.InsightThresholds.CPUWarning,
		MemoryWarning: s.config.InsightThresholds.MemoryWarning,
		DiskWarning:   s.config.InsightThresholds.DiskWarning,
	}))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, map[string]any{
		"service": "nexperf",
		"status":  "ok",
		"version": version.Version,
		"uptime":  time.Since(s.startedAt).String(),
		"ui":      "vue",
	})
}

func (s *Server) storageSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	root := r.URL.Query().Get("path")
	summary, err := storageintel.Analyze(root, s.config.StorageLimits.MaxScanDepth, s.config.StorageLimits.MaxDirectoryRows)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, summary)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.config)
	case http.MethodPut:
		var cfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Path == "" {
			cfg.Path = s.config.Path
		}
		saved, err := config.Save(cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.config = saved
		if s.store != nil {
			s.store.SetRetentionHours(saved.RetentionHours)
		}
		writeJSON(w, saved)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) dashboardWidgets(w http.ResponseWriter, r *http.Request) {
	const key = "dashboard.widgets"
	if s.store == nil {
		http.Error(w, "storage is not configured", http.StatusServiceUnavailable)
		return
	}
	switch r.Method {
	case http.MethodGet:
		value, ok, err := s.store.GetUIState(key)
		if err != nil {
			writeError(w, err)
			return
		}
		if !ok || strings.TrimSpace(value) == "" {
			writeJSON(w, []any{})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(value))
	case http.MethodPut:
		var raw json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !json.Valid(raw) {
			http.Error(w, "invalid widget layout", http.StatusBadRequest)
			return
		}
		if err := s.store.SaveUIState(key, string(raw)); err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	default:
		methodNotAllowed(w)
	}
}

type terminalRequest struct {
	Command string `json:"command"`
	CWD     string `json:"cwd"`
}

func (s *Server) terminalExec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req terminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Command = strings.TrimSpace(req.Command)
	if req.Command == "" {
		http.Error(w, "command is required", http.StatusBadRequest)
		return
	}
	cwd := req.CWD
	if cwd == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cwd = home
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	start := time.Now()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-lc", req.Command)
	cmd.Dir = cwd
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	writeJSON(w, map[string]any{
		"command":     req.Command,
		"cwd":         cwd,
		"stdout":      truncateOutput(stdout.String()),
		"stderr":      truncateOutput(stderr.String()),
		"exit_code":   exitCode,
		"duration_ms": time.Since(start).Milliseconds(),
		"timed_out":   ctx.Err() == context.DeadlineExceeded,
	})
}

func truncateOutput(value string) string {
	const max = 20000
	if len(value) <= max {
		return value
	}
	return value[:max] + "\n... output truncated ..."
}

func (s *Server) configModes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, config.UsageModes())
}

func (s *Server) self(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	paths, _ := service.DefaultPaths()
	dbPath := s.dbPath
	if dbPath == "" {
		dbPath = paths.DB
	}
	storeStats := map[string]any{}
	if s.store != nil {
		if stats, err := s.store.Stats(dbPath); err == nil {
			storeStats = stats
		}
	}
	process, _ := collector.ProcessDetail(int32(os.Getpid()))
	writeJSON(w, map[string]any{
		"service":               "nexperf",
		"version":               version.Version,
		"usage_mode":            s.config.UsageMode,
		"uptime":                time.Since(s.startedAt).String(),
		"memory_alloc_bytes":    mem.Alloc,
		"memory_sys_bytes":      mem.Sys,
		"goroutines":            runtime.NumGoroutine(),
		"process":               process,
		"active_collectors":     s.config.EnabledCollectors,
		"polling_interval":      s.config.PollingIntervalSeconds,
		"database":              storeStats,
		"api_response_timings":  s.snapshotTimings(),
		"insight_generation_ms": s.timingMs("/api/insights"),
	})
}

func (s *Server) history(metric string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		if s.store == nil {
			writeJSON(w, []storage.HistoryPoint{})
			return
		}
		limit := 120
		since := time.Time{}
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err == nil {
				limit = parsed
			}
		}
		if raw := r.URL.Query().Get("range"); raw != "" {
			if duration, ok := parseRange(raw); ok {
				since = time.Now().Add(-duration)
				switch raw {
				case "5m":
					limit = 60
				case "15m":
					limit = 180
				case "1h":
					limit = 720
				}
			}
		}
		points, err := s.store.HistorySince(metric, since, limit)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, points)
	}
}

func parseRange(value string) (time.Duration, bool) {
	switch value {
	case "5m":
		return 5 * time.Minute, true
	case "15m":
		return 15 * time.Minute, true
	case "1h":
		return time.Hour, true
	default:
		return 0, false
	}
}

func (s *Server) dashboard() http.Handler {
	if dist, ok := dashboardDist(); ok {
		return spaHandler{root: dist}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fallbackHTML))
	})
}

func dashboardDist() (string, bool) {
	candidates := []string{filepath.Join("web", "dist")}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "web", "dist"),
			filepath.Join(exeDir, "..", "web", "dist"),
		)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "web", "dist"))
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil || seen[abs] {
			continue
		}
		seen[abs] = true
		if _, err := os.Stat(filepath.Join(abs, "index.html")); err == nil {
			return abs, true
		}
	}
	return "", false
}

func (s *Server) recentSamples(limit int) ([]storage.MetricSample, error) {
	if s.store == nil {
		return nil, nil
	}
	return s.store.Recent(limit)
}

func (s *Server) observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			s.recordTiming(r.URL.Path, time.Since(start))
		}
	})
}

func (s *Server) recordTiming(path string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stat := s.timings[path]
	stat.Count++
	stat.Last = duration
	stat.Total += duration
	stat.Average = stat.Total / time.Duration(stat.Count)
	stat.LastUpdated = time.Now()
	s.timings[path] = stat
}

func (s *Server) snapshotTimings() map[string]timingStat {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]timingStat{}
	for key, value := range s.timings {
		out[key] = value
	}
	return out
}

func (s *Server) timingMs(path string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return float64(s.timings[path].Last.Microseconds()) / 1000
}

type spaHandler struct {
	root string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/nexperf")
	trimmed = strings.TrimPrefix(trimmed, "/")
	if trimmed != "" && !strings.Contains(trimmed, "..") {
		candidate := filepath.Join(h.root, filepath.Clean(trimmed))
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			http.ServeFile(w, r, candidate)
			return
		}
	}
	http.ServeFile(w, r, filepath.Join(h.root, "index.html"))
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
