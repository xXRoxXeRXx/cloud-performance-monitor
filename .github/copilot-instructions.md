

# Nextcloud & HiDrive Performance Monitor – AI Agent Instructions

## Project Overview
A containerized Go application that benchmarks Nextcloud and HiDrive instances via synthetic WebDAV upload/download tests, exports Prometheus metrics (with service label), and provides a ready-to-use Grafana dashboard with service selector.

- **Go Agent** (`cmd/agent/main.go`): Spawns a goroutine per instance (Nextcloud/HiDrive), runs periodic tests.
- **WebDAV Clients** (`internal/nextcloud/client.go`, `internal/hidrive/client.go`): Handle chunked file uploads/downloads, directory management, and cleanup.
- **Metrics** (`internal/agent/metrics.go`): Exposes Prometheus metrics (duration, speed, success) with `service`, `instance`, and `type` labels.
- **Config** (`internal/agent/config.go`): Loads instance credentials and test parameters from `.env`.
- **Docker Compose**: Orchestrates agent, Prometheus, and Grafana.
- **Grafana**: Visualizes metrics from Prometheus, dashboard at `deploy/grafana/dashboard.json` (with service selector).

- **Multi-instance config**: Use numbered env vars (`NC_INSTANCE_1_URL`, etc.) in `.env`. HiDrive is supported via `HIDRIVE_INSTANCE_1_URL`, `HIDRIVE_INSTANCE_1_USER`, `HIDRIVE_INSTANCE_1_PASS` etc.
- **.env example**:
	```env
	# Nextcloud
	NC_INSTANCE_1_URL=https://cloud.company-a.com
	NC_INSTANCE_1_USER=monitor_user_a
	NC_INSTANCE_1_PASS=super-secret-password-a

	# HiDrive
	HIDRIVE_INSTANCE_1_URL=https://storage.ionos.fr
	HIDRIVE_INSTANCE_1_USER=monitor_user_hidrive
	HIDRIVE_INSTANCE_1_PASS=super-secret-password-hidrive
	```
- **Test logic**: Each test uploads a random file (streamed, not loaded in memory), downloads it, validates size, and deletes it.
- **Chunked uploads**: Files >10MB are split into 10MB chunks, uploaded to `/remote.php/dav/uploads/{username}/`, then assembled via MOVE.
- **Metrics**: All metrics labeled by `service` (nextcloud/hidrive), `instance` (URL) and `type` (upload/download). Errors are logged and surfaced via Prometheus labels.
- **Prometheus metric example**:
	```
	nextcloud_test_duration_seconds{service="nextcloud",instance="https://cloud.company-a.com",type="upload"} 2.5
	nextcloud_test_duration_seconds{service="hidrive",instance="https://storage.ionos.fr",type="upload"} 12.3
	```
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

# Test & CI
go test -v -cover ./...
# GitHub Actions: Go-Module- und Docker-Layer-Caching, Coverage, automatischer Build/Push bei Tags
```

## Integration Points
- **Prometheus**: Scrapes agent at `:8080/metrics` (see `prometheus/prometheus.yml`).
- **Grafana**: Imports dashboard from `deploy/grafana/dashboard.json`, Prometheus datasource at `http://prometheus:9090`.
- **Dashboard**: Panels use a `service` selector (Nextcloud/HiDrive) für das Filtern; Instanz-Filter ist optional.

## Common Issues
- **Dashboard import fails**: Ensure `dashboard.json` is valid JSON, not double-wrapped or corrupted.
- **No data in Grafana**: Check agent logs, Prometheus target, and that metrics use correct labels (especially `service`).
- **WebDAV errors**: Confirm Nextcloud/HiDrive user/app password has full read/write permissions.

## Key Files/Dirs
- `cmd/agent/main.go`: Agent entrypoint, goroutine orchestration.
- `internal/nextcloud/client.go`: WebDAV logic, chunked upload/download.
- `internal/hidrive/client.go`: HiDrive WebDAV logic, chunked upload/download.
- `internal/agent/metrics.go`: Prometheus metric definitions.
- `deploy/grafana/dashboard.json`: Grafana dashboard definition (mit Service-Selector).
- `.env.example`: Configuration template.
