package service

import (
	"os"
	"path/filepath"
)

type Paths struct {
	Dir string
	PID string
	Log string
	DB  string
	Bin string
}

func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}
	dir := filepath.Join(home, ".nexperf")
	exe, err := os.Executable()
	if err != nil {
		return Paths{}, err
	}
	return Paths{
		Dir: dir,
		PID: filepath.Join(dir, "nexperf.pid"),
		Log: filepath.Join(dir, "nexperf.log"),
		DB:  filepath.Join(dir, "nexperf.db"),
		Bin: exe,
	}, nil
}

func (p Paths) Ensure() error {
	return os.MkdirAll(p.Dir, 0o755)
}
