# NexPerf

NexPerf is a local-first observability and system intelligence tool for developer machines. It provides a CLI-first workflow, a local Go API service, SQLite-backed historical metrics, and a Vue dashboard served at `http://127.0.0.1:8756/nexperf`.

v0.2.0 turns the prototype dashboard into the foundation of a real local monitoring service.

## Philosophy

- Local-first system visibility with no cloud account
- Terminal workflows before dashboard ceremony
- Observability UX over toy charts
- Reusable collectors, insight rules, APIs, and frontend components
- A path toward privileged diagnostics without enabling them by default

## Architecture

```txt
Collectors
    ↓
Insight Engine
    ↓
SQLite Historical Storage
    ↓
Go HTTP API Server
    ↓
Vue Observability Dashboard
```

Project layout:

```txt
cmd/nexperf         CLI entrypoint
internal/cli       command routing and terminal output
internal/service   lifecycle helpers and background service controls
internal/collector system and process collectors
internal/monitor   5-second historical collection loop
internal/storage   SQLite persistence
internal/insight   rule-based contextual insights
internal/server    local API and Vue asset serving
internal/platform  OS-specific helpers
web/               Vue 3 dashboard
docs/              release notes and design notes
```

NexPerf currently uses `github.com/shirou/gopsutil/v4` for cross-platform CPU, memory, disk, host, and process metrics. SQLite is provided by `modernc.org/sqlite`, which keeps local persistence self-contained for development and future packaging.

## CLI

Build the local binary:

```sh
go build -o bin/nexperf ./cmd/nexperf
```

Run the service lifecycle:

```sh
nexperf start
nexperf open
nexperf stop
```

Inspection commands:

```sh
nexperf status
nexperf processes
nexperf inspect
nexperf explain memory
nexperf explain cpu
nexperf explain disk
nexperf version
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
nexperf --json status
nexperf --port 9000 start
nexperf --port 9000 stop
```

## Service Lifecycle

`nexperf start` launches a background local monitoring service, starts the HTTP API, and enables historical metric collection every 5 seconds.

Service state is stored under:

```txt
~/.nexperf/nexperf.pid
~/.nexperf/nexperf.log
~/.nexperf/nexperf.db
```

`nexperf open` checks `/api/health`. If the service is offline, it starts NexPerf, waits for readiness, and opens the dashboard.

## API

Start the service:

```sh
nexperf start
```

Endpoints:

- `GET /api/health`
- `GET /api/system`
- `GET /api/processes/top`
- `GET /api/insights`
- `GET /api/history/cpu`
- `GET /api/history/memory`
- `GET /api/history/disk`
- `GET /nexperf`

Example:

```sh
curl http://127.0.0.1:8756/api/system
curl http://127.0.0.1:8756/api/history/cpu
```

## Dashboard

Vue is the primary dashboard. Go owns APIs and serves the production Vue build from `web/dist` at `/nexperf`.

The dashboard includes:

- live CPU, memory, and disk cards
- historical sparklines
- CPU, memory, and disk timeline charts
- process search, sorting, live updates, and CPU highlighting
- categorized insights with timestamps and recommendations
- live badge, polling state, and last updated time

The frontend component system lives under:

```txt
web/src/components/ui
web/src/components/charts
web/src/components/metrics
web/src/components/processes
web/src/components/insights
web/src/components/layout
web/src/views
web/src/composables
web/src/services
web/src/stores
```

## Development Workflow

Run the backend:

```sh
go run ./cmd/nexperf --port 8756 serve
```

Run the Vue dev server:

```sh
npm --prefix web install
npm --prefix web run dev
```

Vite proxies `/api` to `http://127.0.0.1:8756`.

## Production Workflow

Build Vue:

```sh
npm --prefix web run build
```

Build NexPerf:

```sh
go build -o bin/nexperf ./cmd/nexperf
```

Run:

```sh
bin/nexperf start
open http://127.0.0.1:8756/nexperf
```

If `web/dist` is missing, Go serves a small fallback status page instead of duplicating the dashboard.

## Screenshots

Screenshots are not checked into v0.2.0 yet. After starting NexPerf, capture the local dashboard at:

```txt
http://127.0.0.1:8756/nexperf
```

## Roadmap

- Embed Vue assets into the Go binary
- Add richer process diagnostics
- Add network, ports, and local service visibility
- Add swap and scheduling explainers
- Add anomaly detection over historical data
- Add explicit privileged diagnostics mode
- Add Homebrew release packaging
- Expand reusable observability components

## Future Homebrew Plan

The intended install flow is:

```sh
brew install nexperf
nexperf start
nexperf open
```

Before publishing, NexPerf needs signed release builds, checksums, a formula, and a stable tap or upstream Homebrew submission.
