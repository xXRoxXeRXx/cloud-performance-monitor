
# Nextcloud Performance Monitor – AI Agent Instructions

## Project Overview
A containerized Go application that benchmarks Nextcloud instances via synthetic WebDAV upload/download tests, exports Prometheus metrics, and provides a ready-to-use Grafana dashboard.

## Architecture & Data Flow
- **Go Agent** (`cmd/agent/main.go`): Spawns a goroutine per Nextcloud instance, runs periodic tests.
- **WebDAV Client** (`internal/nextcloud/client.go`): Handles chunked file uploads/downloads, directory management, and cleanup.
- **Metrics** (`internal/agent/metrics.go`): Exposes Prometheus metrics (duration, speed, success) with `instance` and `type` labels.
- **Config** (`internal/agent/config.go`): Loads instance credentials and test parameters from `.env`.
- **Docker Compose**: Orchestrates agent, Prometheus, and Grafana.
- **Grafana**: Visualizes metrics from Prometheus, dashboard at `deploy/grafana/dashboard.json`.

## Key Patterns & Conventions
- **Multi-instance config**: Use numbered env vars (`NC_INSTANCE_1_URL`, etc.) in `.env`.
- **Test logic**: Each test uploads a random file (streamed, not loaded in memory), downloads it, validates size, and deletes it.
- **Chunked uploads**: Files >10MB are split into 10MB chunks, uploaded to `/remote.php/dav/uploads/{username}/`, then assembled via MOVE.
- **Metrics**: All metrics labeled by `instance` (URL) and `type` (upload/download). Errors are logged and surfaced via Prometheus labels.
- **Error handling**: Log errors, set Prometheus error labels, continue with other instances/tests.
- **Provisioning**: Grafana and Prometheus are auto-provisioned via Dockerfile and config files.

## Developer Workflow
```bash
# Setup
cp .env.example .env
mkdir -p prometheus

# Build & Run
docker compose up -d

# Debug
docker compose logs monitor-agent
docker compose exec monitor-agent /bin/sh
```

## Integration Points
- **Prometheus**: Scrapes agent at `:8080/metrics` (see `prometheus/prometheus.yml`).
- **Grafana**: Imports dashboard from `deploy/grafana/dashboard.json`, Prometheus datasource at `http://prometheus:9090`.
- **Dashboard**: Panels use `instance`/`exported_instance` label for filtering; no variables by default for simplicity.

## Common Issues
- **Dashboard import fails**: Ensure `dashboard.json` is valid JSON, not double-wrapped or corrupted.
- **No data in Grafana**: Check agent logs, Prometheus target, and that metrics use correct labels.
- **WebDAV errors**: Confirm Nextcloud user/app password has full read/write permissions.

## Key Files/Dirs
- `cmd/agent/main.go`: Agent entrypoint, goroutine orchestration.
- `internal/nextcloud/client.go`: WebDAV logic, chunked upload/download.
- `internal/agent/metrics.go`: Prometheus metric definitions.
- `deploy/grafana/dashboard.json`: Grafana dashboard definition.
- `.env.example`: Configuration template.
