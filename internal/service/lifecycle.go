package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Host string
	Port int
}

func DashboardURL(cfg Config) string {
	return fmt.Sprintf("http://%s:%d/nexperf", cfg.Host, cfg.Port)
}

func HealthURL(cfg Config) string {
	return fmt.Sprintf("http://%s:%d/api/health", cfg.Host, cfg.Port)
}

func IsServerRunning(cfg Config) bool {
	client := http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(HealthURL(cfg))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func WaitForServerReady(ctx context.Context, cfg Config) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		if IsServerRunning(cfg) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func StartBackgroundServer(cfg Config) error {
	if IsServerRunning(cfg) {
		return nil
	}
	paths, err := DefaultPaths()
	if err != nil {
		return err
	}
	if err := paths.Ensure(); err != nil {
		return err
	}

	logFile, err := os.OpenFile(paths.Log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command(paths.Bin, "--host", cfg.Host, "--port", strconv.Itoa(cfg.Port), "serve")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(), "NEXPERF_DAEMON=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := os.WriteFile(paths.PID, []byte(strconv.Itoa(cmd.Process.Pid)), 0o644); err != nil {
		_ = cmd.Process.Kill()
		return err
	}
	return cmd.Process.Release()
}

func StopBackgroundServer(cfg Config) error {
	paths, err := DefaultPaths()
	if err != nil {
		return err
	}
	pid, err := readPID(paths.PID)
	if err != nil {
		if IsServerRunning(cfg) {
			return errors.New("service is running but pid file is unavailable")
		}
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		_ = os.Remove(paths.PID)
		return err
	}
	if err := process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		if !IsServerRunning(cfg) {
			_ = os.Remove(paths.PID)
			return nil
		}
		select {
		case <-ctx.Done():
			_ = process.Kill()
			_ = os.Remove(paths.PID)
			return nil
		case <-ticker.C:
		}
	}
}

func readPID(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}
