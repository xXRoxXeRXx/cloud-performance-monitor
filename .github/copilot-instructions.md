# Cloud Performance Monitor – AI Agent Instructions

## Project Overview

A containerized Go application that benchmarks Nextcloud, HiDrive, MagentaCLOUD, and Dropbox instances via synthetic WebDAV upload/download tests, exports Prometheus metrics (with service label), and provides a ready-to-use Grafana dashboard with service selector. Features comprehensive alerting via Alertmanager with email notifications, structured logging across all services, and HTTP uptime monitoring for all configured instances.

## Architecture

- **Go Agent** (`cmd/agent/main.go`): Spawns a goroutine per instance (Nextcloud/HiDrive/MagentaCLOUD/HiDrive Legacy/Dropbox), runs periodic tests.

- **WebDAV Clients** (`internal/nextcloud/client.go`, `internal/hidrive/client.go`, `internal/magentacloud/client.go`): Handle chunked file uploads/downloads, directory management, and cleanup.

- **OAuth2 Clients** (`internal/hidrive_legacy/client.go`, `internal/dropbox/client.go`): Handle OAuth2 authentication, refresh tokens, and REST API operations.

- **Uptime Monitoring** (`internal/agent/uptime_checker.go`): HTTP health checks every 60 seconds for all instances with status code validation (200-299).

- **Metrics** (`internal/agent/metrics.go`): Prometheus metrics with `service`, `instance`, and `type` labels, including uptime metrics.

- **Docker Compose**: Orchestrates agent, Prometheus, Alertmanager, Grafana- **Metrics** (`internal/agent/metrics.go`): Exposes Prometheus metrics (duration, speed, success) with `service`, `instance`, and `type` labels.

- **Config** (`internal/agent/config.go`): Loads instance credentials and test parameters from `.env`.

## Configuration (.env)- **Logging** (`internal/agent/logger.go`): Unified structured logging with ClientLogger interface, configurable via LOG_LEVEL and LOG_FORMAT.

```env- **Alerting** (`alertmanager/`): Email notifications for critical, performance, network, and SLA alerts with enhanced templates.

# Test settings- **Security** (`docs/PORT_SECURITY.md`): Minimal external port exposure, only Grafana accessible from outside.

TEST_FILE_SIZE_MB=100- **Docker Compose**: Orchestrates agent, Prometheus, Alertmanager, and Grafana with internal networking.

TEST_INTERVAL_SECONDS=900- **Grafana**: Multiple dashboards including enhanced analytics at `deploy/grafana/enhanced-dashboard.json`.

TEST_CHUNK_SIZE_MB=10

- **Multi-instance config**: Use numbered env vars (`NC_INSTANCE_1_URL`, etc.) in `.env`. HiDrive is supported via `HIDRIVE_INSTANCE_1_URL`, `HIDRIVE_INSTANCE_1_USER`, `HIDRIVE_INSTANCE_1_PASS` etc. MagentaCLOUD uses `MAGENTACLOUD_INSTANCE_1_URL`, `MAGENTACLOUD_INSTANCE_1_USER`, `MAGENTACLOUD_INSTANCE_1_ANID`, `MAGENTACLOUD_INSTANCE_1_PASS`. HiDrive Legacy uses OAuth2 with `HIDRIVE_LEGACY_INSTANCE_1_REFRESH_TOKEN`. Dropbox uses `DROPBOX_INSTANCE_1_TOKEN`.

# SMTP for alerts- **.env example**:

SMTP_SMARTHOST=smtp.example.com:587	```env

SMTP_FROM=monitor@example.com	# Test Configuration

SMTP_AUTH_USERNAME=monitor@example.com	TEST_FILE_SIZE_MB=100

SMTP_AUTH_PASSWORD=password	TEST_INTERVAL_SECONDS=900

EMAIL_ADMIN=admin@example.com	TEST_CHUNK_SIZE_MB=10



# Services (numbered instances)	# Email/SMTP Configuration for Alertmanager

NC_INSTANCE_1_URL=https://cloud.example.com	SMTP_SMARTHOST=smtp.example.com:587

NC_INSTANCE_1_USER=user	SMTP_FROM=cloud-monitor@company.com

NC_INSTANCE_1_PASS=password	SMTP_AUTH_USERNAME=cloud-monitor@company.com

	SMTP_AUTH_PASSWORD=your-smtp-password

HIDRIVE_INSTANCE_1_URL=https://storage.ionos.fr	SMTP_REQUIRE_TLS=true

HIDRIVE_INSTANCE_1_USER=user	EMAIL_ADMIN=admin@company.com

HIDRIVE_INSTANCE_1_PASS=password	EMAIL_DEVOPS=devops@company.com

	EMAIL_NETWORK=network@company.com

DROPBOX_INSTANCE_1_REFRESH_TOKEN=token	EMAIL_MANAGEMENT=management@company.com

DROPBOX_INSTANCE_1_APP_KEY=key

DROPBOX_INSTANCE_1_APP_SECRET=secret	# Nextcloud

DROPBOX_INSTANCE_1_NAME=name	NC_INSTANCE_1_URL=https://cloud.company-a.com

```	NC_INSTANCE_1_USER=monitor_user_a

	NC_INSTANCE_1_PASS=super-secret-password-a

## Alerts (9 Total)

| Alert | Severity | Trigger |
|-------|----------|---------|
| ServiceDown | Critical | Agent not responding |
| CloudServiceUnavailable | Warning | No tests successful in 15min |
| ServiceUptimeDown | Critical | HTTP uptime check failing for 3min |
| ServiceUptimeDegraded | Warning | HTTP response time >5s for 5min |
| LowServiceAvailability | Warning | Service availability <95% over last hour |
| SlowUploadSpeed | Warning | Upload < 1 MB/s |
| HighErrorRate | Warning | > 20% errors in 1 hour |
| CircuitBreakerOpen | Warning | Service protection active |
| PrometheusStorageNearFull | Warning | Storage > 80% |

## Dashboards (4 Total)

- **Daily Performance** - 24-hour upload/download speeds
- **Monthly Performance** - 30-day trends
- **Errors** - Error tracking and analysis
- **Uptime** - HTTP uptime status, response times, and availability statistics

## Dashboards (3 Total)	MAGENTACLOUD_INSTANCE_1_PASS=app-password

- **Daily Performance** - 24-hour upload/download speeds

- **Monthly Performance** - 30-day trends	# HiDrive Legacy (OAuth2)

- **Errors** - Error tracking and analysis	HIDRIVE_LEGACY_INSTANCE_1_URL=https://api.hidrive.strato.com/2.1

	HIDRIVE_LEGACY_INSTANCE_1_CLIENT_ID=your-oauth2-client-id

## Quick Commands	HIDRIVE_LEGACY_INSTANCE_1_CLIENT_SECRET=your-oauth2-client-secret

```bash	HIDRIVE_LEGACY_INSTANCE_1_REFRESH_TOKEN=your-refresh-token

make dev        # Start all services	HIDRIVE_LEGACY_INSTANCE_1_NAME=hidrive-legacy-main

make stop       # Stop all services

make logs       # View agent logs	# Dropbox

make test       # Run Go tests	DROPBOX_INSTANCE_1_REFRESH_TOKEN=your-refresh-token

make dashboards # Open Grafana	DROPBOX_INSTANCE_1_APP_KEY=your-app-key

```	DROPBOX_INSTANCE_1_APP_SECRET=your-app-secret

	DROPBOX_INSTANCE_1_NAME=user@example.com

## Key Files	```

- `cmd/agent/main.go` - Agent entrypoint- **Test logic**: Each test uploads a random file (streamed, not loaded in memory), downloads it, validates size, and deletes it.

- `internal/agent/metrics.go` - Prometheus metrics- **Chunked uploads**: Files >10MB are split into 10MB chunks, uploaded to `/remote.php/dav/uploads/{username}/` for Nextcloud/HiDrive or `/remote.php/dav/uploads/{ANID}/` for MagentaCLOUD, then assembled via MOVE.

- `internal/*/client.go` - Service clients- **Performance optimizations**: 

- `prometheus/alert_rules.yml` - Alert definitions  - HiDrive uses optimized HTTP transport with MaxIdleConns=100, MaxConnsPerHost=100 for better connection reuse

- `deploy/grafana/*.json` - Dashboard definitions  - MagentaCLOUD includes 2-second delay between upload completion and download start due to backend file availability timing

- `alertmanager/alertmanager.yml.template` - Email templates  - All clients use 10-minute timeout for MOVE operations to handle large file assembly

  - Progressive backoff retry logic for chunk upload conflicts (409 errors)
- **Protocol compliance**: 
  - MagentaCLOUD follows Nextcloud Chunking v2 protocol with mandatory OC-Total-Length and Destination headers
  - User-Agent mimics official Nextcloud desktop client for better compatibility
  - OAuth2 token refresh handling for HiDrive Legacy and Dropbox services
- **Metrics**: All metrics labeled by `service` (nextcloud/hidrive/magentacloud/hidrive_legacy/dropbox), `instance` (URL) and `type` (upload/download). Errors are logged and surfaced via Prometheus labels.
- **Prometheus metric example**:
	```
	cloud_test_duration_seconds{service="nextcloud",instance="https://cloud.company-a.com",type="upload"} 2.5
	cloud_test_duration_seconds{service="hidrive",instance="https://storage.ionos.fr",type="upload"} 12.3
	cloud_test_duration_seconds{service="magentacloud",instance="https://magentacloud.de",type="upload"} 3.9
	cloud_test_duration_seconds{service="hidrive_legacy",instance="hidrive-legacy-main",type="upload"} 1.8
	cloud_test_duration_seconds{service="dropbox",instance="user@example.com",type="upload"} 3.2
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
- **Dashboard**: Panels use a `service` selector (Nextcloud/HiDrive/MagentaCLOUD/HiDrive Legacy/Dropbox) für das Filtern; Instanz-Filter ist optional.

## Common Issues
- **Dashboard import fails**: Ensure `dashboard.json` is valid JSON, not double-wrapped or corrupted.
- **No data in Grafana**: Check agent logs, Prometheus target, and that metrics use correct labels (especially `service`).
- **WebDAV errors**: Confirm Nextcloud/HiDrive user/app password has full read/write permissions.
- **MagentaCLOUD 404 download errors**: Fixed by implementing 2-second delay after upload completion to allow backend file availability.
- **409 Conflict errors**: Handled by progressive backoff retry logic with If-Match headers for chunk overwrites.
- **Large file timeouts**: MOVE operations use extended 10-minute timeout for file assembly.
- **OAuth2 token expiry**: HiDrive Legacy and Dropbox clients automatically refresh tokens using refresh_token grants.

## Key Files/Dirs
- `cmd/agent/main.go`: Agent entrypoint, goroutine orchestration, uptime checker initialization.
- `internal/agent/uptime_checker.go`: HTTP uptime monitoring implementation with 60s check interval.
- `internal/nextcloud/client.go`: WebDAV logic, chunked upload/download.
- `internal/hidrive/client.go`: HiDrive WebDAV logic, chunked upload/download with optimized HTTP transport.
- `internal/magentacloud/client.go`: MagentaCLOUD WebDAV client with Nextcloud Chunking v2 protocol implementation.
- `internal/hidrive_legacy/client.go`: HiDrive Legacy OAuth2 REST API client with token refresh handling.
- `internal/dropbox/client.go`: Dropbox OAuth2 client with automatic token refresh.
- `internal/agent/*_tester.go`: Service-specific test implementations with error handling and timing logic.
- `internal/agent/metrics.go`: Prometheus metric definitions including uptime metrics.
- `deploy/grafana/uptime-dashboard.json`: Grafana uptime monitoring dashboard.

- `internal/agent/metrics.go`: Prometheus metric definitions.
- `deploy/grafana/dashboard.json`: Grafana dashboard definition (mit Service-Selector).
- `.env.example`: Configuration template.
