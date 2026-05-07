package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prabath/nexperf/internal/collector"
	"github.com/prabath/nexperf/internal/insight"
	"github.com/prabath/nexperf/internal/monitor"
	"github.com/prabath/nexperf/internal/platform"
	"github.com/prabath/nexperf/internal/server"
	"github.com/prabath/nexperf/internal/service"
	"github.com/prabath/nexperf/internal/storage"
	"github.com/prabath/nexperf/internal/version"
)

type options struct {
	host       string
	port       int
	jsonOutput bool
	privileged bool
	out        io.Writer
	err        io.Writer
}

func Run(args []string) int {
	opts := options{host: "127.0.0.1", port: 8756, out: os.Stdout, err: os.Stderr}
	remaining, err := parseOptions(args, &opts)
	if err != nil {
		fmt.Fprintf(opts.err, "nexperf: %v\n", err)
		return 2
	}

	if opts.privileged {
		fmt.Fprintln(opts.out, "Privileged diagnostics are planned for a future NexPerf release; v0.2 runs in local user mode.")
	}

	if len(remaining) == 0 {
		usage(opts.out)
		return 0
	}

	cmd := remaining[0]
	cmdArgs := remaining[1:]
	if err := runCommand(cmd, cmdArgs, opts); err != nil {
		fmt.Fprintf(opts.err, "nexperf: %v\n", err)
		return 1
	}
	return 0
}

func parseOptions(args []string, opts *options) ([]string, error) {
	remaining := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			opts.jsonOutput = true
		case arg == "--privileged":
			opts.privileged = true
		case arg == "--host":
			i++
			if i >= len(args) {
				return nil, fmt.Errorf("--host requires a value")
			}
			opts.host = args[i]
		case strings.HasPrefix(arg, "--host="):
			opts.host = strings.TrimPrefix(arg, "--host=")
		case arg == "--port":
			i++
			if i >= len(args) {
				return nil, fmt.Errorf("--port requires a value")
			}
			port, err := strconv.Atoi(args[i])
			if err != nil {
				return nil, fmt.Errorf("--port must be a number")
			}
			opts.port = port
		case strings.HasPrefix(arg, "--port="):
			port, err := strconv.Atoi(strings.TrimPrefix(arg, "--port="))
			if err != nil {
				return nil, fmt.Errorf("--port must be a number")
			}
			opts.port = port
		default:
			remaining = append(remaining, arg)
		}
	}
	return remaining, nil
}

func runCommand(cmd string, args []string, opts options) error {
	switch cmd {
	case "start":
		return start(opts)
	case "serve":
		return serve(opts)
	case "stop":
		return stop(opts)
	case "status":
		return status(opts)
	case "processes":
		return processes(opts)
	case "inspect":
		return inspect(opts)
	case "explain":
		return explain(args, opts)
	case "open":
		return open(opts)
	case "version":
		return printVersion(opts)
	case "help", "-h", "--help":
		usage(opts.out)
		return nil
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func start(opts options) error {
	cfg := service.Config{Host: opts.host, Port: opts.port}
	if service.IsServerRunning(cfg) {
		fmt.Fprintln(opts.out, "NexPerf is already running")
		fmt.Fprintf(opts.out, "Dashboard: %s\n", service.DashboardURL(cfg))
		return nil
	}

	fmt.Fprintln(opts.out, "Starting NexPerf...")
	if err := service.StartBackgroundServer(cfg); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := service.WaitForServerReady(ctx, cfg); err != nil {
		return fmt.Errorf("service did not become ready: %w", err)
	}
	fmt.Fprintf(opts.out, "API server listening on %s:%d\n", opts.host, opts.port)
	fmt.Fprintln(opts.out, "Historical metrics collection enabled")
	fmt.Fprintln(opts.out, "NexPerf is running")
	fmt.Fprintf(opts.out, "Dashboard: %s\n", service.DashboardURL(cfg))
	return nil
}

func serve(opts options) error {
	paths, err := service.DefaultPaths()
	if err != nil {
		return err
	}
	if err := paths.Ensure(); err != nil {
		return err
	}

	store, err := storage.Open(paths.DB)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	logger := log.New(opts.err, "nexperf: ", log.LstdFlags)
	collector := monitor.New(store, logger)
	go collector.Run(ctx)

	srv := server.New(store)
	errCh := make(chan error, 1)
	go func() {
		fmt.Fprintf(opts.out, "NexPerf service listening on http://%s:%d\n", opts.host, opts.port)
		fmt.Fprintln(opts.out, "Historical metrics collection enabled")
		errCh <- srv.ListenAndServe(opts.host, opts.port)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func stop(opts options) error {
	cfg := service.Config{Host: opts.host, Port: opts.port}
	fmt.Fprintln(opts.out, "Stopping NexPerf...")
	if err := service.StopBackgroundServer(cfg); err != nil {
		return err
	}
	fmt.Fprintln(opts.out, "NexPerf stopped successfully")
	return nil
}

func status(opts options) error {
	system, err := collector.CollectSystem()
	if err != nil {
		return err
	}
	if opts.jsonOutput {
		return writeJSON(opts.out, system)
	}
	fmt.Fprintln(opts.out, "NexPerf system status")
	fmt.Fprintf(opts.out, "CPU:     %.1f%%\n", system.CPUPercent)
	fmt.Fprintf(opts.out, "Memory:  %.1f%% (%s / %s)\n", system.Memory.Percent, formatBytes(system.Memory.Used), formatBytes(system.Memory.Total))
	fmt.Fprintf(opts.out, "Disk:    %.1f%% (%s / %s)\n", system.Disk.Percent, formatBytes(system.Disk.Used), formatBytes(system.Disk.Total))
	fmt.Fprintf(opts.out, "OS/Arch: %s/%s\n", system.OS, system.Arch)
	if system.Hostname != "" {
		fmt.Fprintf(opts.out, "Host:    %s\n", system.Hostname)
	}
	return nil
}

func processes(opts options) error {
	processes, err := collector.TopProcesses(10)
	if err != nil {
		return err
	}
	if opts.jsonOutput {
		return writeJSON(opts.out, processes)
	}
	fmt.Fprintf(opts.out, "%-8s %-32s %12s %8s %s\n", "PID", "NAME", "MEMORY", "CPU", "USER")
	for _, proc := range processes {
		fmt.Fprintf(opts.out, "%-8d %-32.32s %11.1fM %7.1f%% %s\n", proc.PID, proc.Name, proc.MemoryMB, proc.CPUPercent, proc.User)
	}
	return nil
}

func inspect(opts options) error {
	system, err := collector.CollectSystem()
	if err != nil {
		return err
	}
	processes, _ := collector.TopProcesses(12)
	insights := insight.GenerateContextual(system, processes, nil)
	if opts.jsonOutput {
		return writeJSON(opts.out, insights)
	}
	fmt.Fprintln(opts.out, "NexPerf inspection")
	fmt.Fprintf(opts.out, "CPU %.1f%%, memory %.1f%%, disk %.1f%% on %s/%s\n\n", system.CPUPercent, system.Memory.Percent, system.Disk.Percent, system.OS, system.Arch)
	for _, item := range insights {
		fmt.Fprintf(opts.out, "[%s] %s\n%s\nRecommendation: %s\n\n", strings.ToUpper(item.Severity), item.Title, item.Message, item.Recommendation)
	}
	return nil
}

func explain(args []string, opts options) error {
	if len(args) == 0 {
		return fmt.Errorf("explain requires one of: memory, cpu, disk")
	}
	system, err := collector.CollectSystem()
	if err != nil {
		return err
	}

	topic := strings.ToLower(args[0])
	exp := explanation(topic, system)
	if exp == nil {
		return fmt.Errorf("unknown explain topic %q", topic)
	}
	if opts.jsonOutput {
		return writeJSON(opts.out, exp)
	}

	fmt.Fprintf(opts.out, "%s\n\n%s\n\nCurrent: %s\n", exp["title"], exp["message"], exp["current"])
	return nil
}

func open(opts options) error {
	cfg := service.Config{Host: opts.host, Port: opts.port}
	if service.IsServerRunning(cfg) {
		fmt.Fprintln(opts.out, "NexPerf service detected.")
	} else {
		fmt.Fprintln(opts.out, "NexPerf service not running.")
		fmt.Fprintln(opts.out, "Starting NexPerf...")
		if err := service.StartBackgroundServer(cfg); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := service.WaitForServerReady(ctx, cfg); err != nil {
			return fmt.Errorf("service did not become ready: %w", err)
		}
		fmt.Fprintln(opts.out, "Service ready.")
	}
	fmt.Fprintln(opts.out, "Opening dashboard...")
	url := service.DashboardURL(cfg)
	return platform.OpenBrowser(url)
}

func printVersion(opts options) error {
	info := version.Info()
	if opts.jsonOutput {
		return writeJSON(opts.out, info)
	}
	fmt.Fprintf(opts.out, "nexperf %s\ncommit: %s\nos/arch: %s/%s\n", info["version"], info["commit"], info["os"], info["arch"])
	return nil
}

func explanation(topic string, system collector.SystemSummary) map[string]string {
	switch topic {
	case "memory":
		return map[string]string{
			"topic":   "memory",
			"title":   "Memory usage",
			"message": "Memory usage shows how much physical RAM is actively in use. High memory pressure can force the OS to swap to disk, which is much slower than RAM.",
			"current": fmt.Sprintf("%.1f%% (%s / %s)", system.Memory.Percent, formatBytes(system.Memory.Used), formatBytes(system.Memory.Total)),
		}
	case "cpu":
		return map[string]string{
			"topic":   "cpu",
			"title":   "CPU usage",
			"message": "CPU usage shows how busy the processor is across sampled work. Sustained high CPU can come from builds, indexing, browsers, containers, or runaway processes.",
			"current": fmt.Sprintf("%.1f%%", system.CPUPercent),
		}
	case "disk":
		return map[string]string{
			"topic":   "disk",
			"title":   "Disk usage",
			"message": "Disk usage shows how much capacity is consumed on the root filesystem. Very full disks can break updates, logs, databases, and development caches.",
			"current": fmt.Sprintf("%.1f%% (%s / %s)", system.Disk.Percent, formatBytes(system.Disk.Used), formatBytes(system.Disk.Total)),
		}
	default:
		return nil
	}
}

func dashboardURL(opts options) string {
	return service.DashboardURL(service.Config{Host: opts.host, Port: opts.port})
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func formatBytes(bytes uint64) string {
	gb := float64(bytes) / 1024 / 1024 / 1024
	if gb >= 10 {
		return fmt.Sprintf("%.0fGB", gb)
	}
	return fmt.Sprintf("%.1fGB", gb)
}

func usage(w io.Writer) {
	fmt.Fprintln(w, `NexPerf v0.2.0

Usage:
  nexperf [--host 127.0.0.1] [--port 8756] [--json] [--privileged] <command>

Commands:
  start              Start the local monitoring service
  stop               Stop the local monitoring service
  status             Print current system summary
  processes          List top processes by memory usage
  inspect            Print rule-based system inspection
  explain <topic>    Explain memory, cpu, or disk usage
  open               Open the dashboard URL in a browser
  version            Print version information`)
}
