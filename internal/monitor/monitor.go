package monitor

import (
	"context"
	"log"
	"time"

	"github.com/prabath/nexperf/internal/collector"
	"github.com/prabath/nexperf/internal/storage"
)

type Monitor struct {
	store    *storage.Store
	interval time.Duration
	logger   *log.Logger
}

func New(store *storage.Store, logger *log.Logger) *Monitor {
	return &Monitor{store: store, interval: 5 * time.Second, logger: logger}
}

func (m *Monitor) SetInterval(interval time.Duration) {
	if interval > 0 {
		m.interval = interval
	}
}

func (m *Monitor) Run(ctx context.Context) {
	m.collectOnce()
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collectOnce()
		}
	}
}

func (m *Monitor) collectOnce() {
	summary, err := collector.CollectSystem()
	if err != nil {
		m.logf("collector error: %v", err)
		return
	}
	if err := m.store.SaveSystem(summary); err != nil {
		m.logf("storage error: %v", err)
	}
}

func (m *Monitor) logf(format string, args ...any) {
	if m.logger != nil {
		m.logger.Printf(format, args...)
	}
}
