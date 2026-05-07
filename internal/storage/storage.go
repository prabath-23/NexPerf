package storage

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"

	"github.com/prabath/nexperf/internal/collector"
)

type Store struct {
	db *sql.DB
}

type MetricSample struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemoryUsed  uint64    `json:"memory_used"`
	MemoryTotal uint64    `json:"memory_total"`
	MemoryPct   float64   `json:"memory_percent"`
	DiskUsed    uint64    `json:"disk_used"`
	DiskTotal   uint64    `json:"disk_total"`
	DiskPct     float64   `json:"disk_percent"`
}

type HistoryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Used      uint64    `json:"used,omitempty"`
	Total     uint64    `json:"total,omitempty"`
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.Migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS metric_samples (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp TEXT NOT NULL,
  cpu_percent REAL NOT NULL,
  memory_used INTEGER NOT NULL,
  memory_total INTEGER NOT NULL,
  memory_percent REAL NOT NULL,
  disk_used INTEGER NOT NULL,
  disk_total INTEGER NOT NULL,
  disk_percent REAL NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_metric_samples_timestamp ON metric_samples(timestamp);
`)
	return err
}

func (s *Store) SaveSystem(summary collector.SystemSummary) error {
	if s == nil {
		return errors.New("storage is not configured")
	}
	_, err := s.db.Exec(`
INSERT INTO metric_samples (
  timestamp, cpu_percent, memory_used, memory_total, memory_percent, disk_used, disk_total, disk_percent
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		summary.Timestamp.Format(time.RFC3339Nano),
		summary.CPUPercent,
		summary.Memory.Used,
		summary.Memory.Total,
		summary.Memory.Percent,
		summary.Disk.Used,
		summary.Disk.Total,
		summary.Disk.Percent,
	)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DELETE FROM metric_samples WHERE timestamp < ?", time.Now().Add(-24*time.Hour).Format(time.RFC3339Nano))
	return err
}

func (s *Store) History(metric string, limit int) ([]HistoryPoint, error) {
	return s.HistorySince(metric, time.Time{}, limit)
}

func (s *Store) HistorySince(metric string, since time.Time, limit int) ([]HistoryPoint, error) {
	if limit <= 0 || limit > 720 {
		limit = 120
	}

	valueColumn := ""
	usedColumn := ""
	totalColumn := ""
	switch metric {
	case "cpu":
		valueColumn = "cpu_percent"
	case "memory":
		valueColumn = "memory_percent"
		usedColumn = "memory_used"
		totalColumn = "memory_total"
	case "disk":
		valueColumn = "disk_percent"
		usedColumn = "disk_used"
		totalColumn = "disk_total"
	default:
		return nil, errors.New("unknown history metric")
	}

	query := "SELECT timestamp, " + valueColumn
	if usedColumn != "" {
		query += ", " + usedColumn + ", " + totalColumn
	}
	query += " FROM metric_samples"
	args := []any{}
	if !since.IsZero() {
		query += " WHERE timestamp >= ?"
		args = append(args, since.Format(time.RFC3339Nano))
	}
	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reversed := []HistoryPoint{}
	for rows.Next() {
		point := HistoryPoint{}
		var ts string
		if usedColumn == "" {
			if err := rows.Scan(&ts, &point.Value); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&ts, &point.Value, &point.Used, &point.Total); err != nil {
				return nil, err
			}
		}
		parsed, err := time.Parse(time.RFC3339Nano, ts)
		if err == nil {
			point.Timestamp = parsed
		}
		reversed = append(reversed, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	points := make([]HistoryPoint, len(reversed))
	for i := range reversed {
		points[len(reversed)-1-i] = reversed[i]
	}
	return points, nil
}

func (s *Store) Recent(limit int) ([]MetricSample, error) {
	if limit <= 0 || limit > 720 {
		limit = 120
	}
	rows, err := s.db.Query(`
SELECT timestamp, cpu_percent, memory_used, memory_total, memory_percent, disk_used, disk_total, disk_percent
FROM metric_samples ORDER BY timestamp DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reversed := []MetricSample{}
	for rows.Next() {
		sample := MetricSample{}
		var ts string
		if err := rows.Scan(&ts, &sample.CPUPercent, &sample.MemoryUsed, &sample.MemoryTotal, &sample.MemoryPct, &sample.DiskUsed, &sample.DiskTotal, &sample.DiskPct); err != nil {
			return nil, err
		}
		if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			sample.Timestamp = parsed
		}
		reversed = append(reversed, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	samples := make([]MetricSample, len(reversed))
	for i := range reversed {
		samples[len(reversed)-1-i] = reversed[i]
	}
	return samples, nil
}
