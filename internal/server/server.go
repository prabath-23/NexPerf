package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prabath/nexperf/internal/collector"
	"github.com/prabath/nexperf/internal/insight"
	"github.com/prabath/nexperf/internal/version"
)

type Server struct {
	startedAt time.Time
}

func New() *Server {
	return &Server{startedAt: time.Now()}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/system", s.system)
	mux.HandleFunc("/api/processes/top", s.processesTop)
	mux.HandleFunc("/api/insights", s.insights)
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/nexperf", s.dashboard)
	mux.HandleFunc("/nexperf/", s.dashboard)
	return mux
}

func (s *Server) ListenAndServe(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	return http.ListenAndServe(addr, s.Handler())
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
	processes, err := collector.TopProcesses(12)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, processes)
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
	writeJSON(w, insight.Generate(system))
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
	})
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(dashboardHTML))
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
