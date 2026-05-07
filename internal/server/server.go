package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prabath/nexperf/internal/collector"
	"github.com/prabath/nexperf/internal/health"
	"github.com/prabath/nexperf/internal/insight"
	"github.com/prabath/nexperf/internal/storage"
	"github.com/prabath/nexperf/internal/version"
)

type Server struct {
	startedAt time.Time
	store     *storage.Store
	http      *http.Server
}

func New(store *storage.Store) *Server {
	return &Server{startedAt: time.Now(), store: store}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/system", s.system)
	mux.HandleFunc("/api/processes/top", s.processesTop)
	mux.HandleFunc("/api/insights", s.insights)
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/health-score", s.healthScore)
	mux.HandleFunc("/api/history/cpu", s.history("cpu"))
	mux.HandleFunc("/api/history/memory", s.history("memory"))
	mux.HandleFunc("/api/history/disk", s.history("disk"))
	mux.Handle("/nexperf", s.dashboard())
	mux.Handle("/nexperf/", s.dashboard())
	return mux
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
	writeJSON(w, health.Calculate(system, processes, history))
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
	writeJSON(w, insight.GenerateContextual(system, processes, history))
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
