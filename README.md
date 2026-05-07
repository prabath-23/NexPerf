# NexPerf

NexPerf is a local-first system intelligence and performance monitoring tool for developers. It pairs a CLI-first workflow with a local dashboard served by the app.

v0.1 focuses on the essentials: inspect the machine from the terminal, expose a small local API, and view current system health at `http://127.0.0.1:8756/nexperf`.

## Philosophy

NexPerf is built for local trust and useful explanations:

- Local-first by default
- Fast terminal inspection before dashboard polish
- Human-readable diagnostics alongside raw metrics
- Clean boundaries between collectors, CLI, API, and UI
- Prepared for future privileged diagnostics without enabling them in v0.1

## CLI

The binary is named `nexperf`.

```sh
go run ./cmd/nexperf start
go run ./cmd/nexperf status
go run ./cmd/nexperf processes
go run ./cmd/nexperf inspect
go run ./cmd/nexperf explain memory
go run ./cmd/nexperf explain cpu
go run ./cmd/nexperf explain disk
go run ./cmd/nexperf open
go run ./cmd/nexperf version
```

Global flags:

```sh
--host        Server host, default 127.0.0.1
--port        Server port, default 8756
--json        JSON output for supported commands
--privileged  Placeholder for planned privileged diagnostics
```

Examples:

```sh
go run ./cmd/nexperf --json status
go run ./cmd/nexperf --port 9000 start
```

## API

Start the local server:

```sh
go run ./cmd/nexperf start
```

Endpoints:

- `GET /api/system`
- `GET /api/processes/top`
- `GET /api/insights`
- `GET /api/health`
- `GET /nexperf`

`/api/system` returns CPU usage, memory usage, disk usage, OS, architecture, hostname when available, and timestamp.

`/api/processes/top` returns top processes by memory usage with PID, name, memory MB, CPU percent when available, and user when available.

`/api/insights` returns rule-based insight objects with `id`, `severity`, `title`, `message`, and `recommendation`.

## Dashboard

The dashboard source lives in `web/` and uses Vue 3 + Vite.

For v0.1 development:

```sh
cd web
npm install
npm run dev
```

The Go server also serves a lightweight dashboard at `/nexperf` today. The project is structured so a future Vue production build can be embedded into the Go binary.

## Development

Requirements:

- Go 1.22+
- Node.js 20+ for dashboard development

Run CLI checks:

```sh
go run ./cmd/nexperf status
go run ./cmd/nexperf processes
go run ./cmd/nexperf inspect
```

Run the local app:

```sh
go run ./cmd/nexperf start
```

Then open:

```txt
http://127.0.0.1:8756/nexperf
```

## Architecture

```txt
cmd/nexperf        CLI entrypoint
internal/cli      command handling and terminal formatting
internal/collector reusable system and process collectors
internal/insight  rule-based local insights
internal/server   local API server and dashboard route
internal/platform OS-specific helpers
internal/version  build version metadata
web/              Vue 3 dashboard source
docs/             project notes
```

NexPerf uses `github.com/shirou/gopsutil/v4` for v0.1 metrics because it is widely used, cross-platform, and avoids fragile OS-specific shell parsing for CPU, memory, disk, host, and process data.

## Roadmap

- Embed the production Vue build into the Go binary
- Add SQLite-backed history and snapshots
- Add charts and process trend views
- Add network and local service visibility
- Add macOS/Linux privileged diagnostics behind explicit user consent
- Add Homebrew packaging

## Future Homebrew Plan

The intended install flow is:

```sh
brew install nexperf
nexperf start
```

Before publishing, NexPerf will need release builds, checksums, a formula, and a stable tap or upstream Homebrew submission.
