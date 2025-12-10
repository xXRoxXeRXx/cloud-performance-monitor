
# 📊 Cloud Performance Monitor

[![Build Status](https://img.shields.io/github/actions/workflow/status/xXRoxXeRXx/cloud-performance-monitor/docker-image.yml?branch=main)](https://github.com/xXRoxXeRXx/cloud-performance-monitor/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue)](https://www.docker.com/)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8)](https://golang.org/)

Ein professionelles, containerisiertes Monitoring-System zur kontinuierlichen Überwachung der Performance von Nextcloud-, HiDrive-, MagentaCLOUD- und Dropbox-Instanzen mit vollständigem Alerting, E-Mail-Benachrichtigungen und erweiterten Dashboards.

## ✨ Features

### 🎯 **Core Monitoring**
- **Multi-Instance Support**: Überwache beliebig viele Nextcloud-, HiDrive-, MagentaCLOUD- und Dropbox-Instanzen gleichzeitig
- **Real Performance Testing**: Synthetische Upload/Download-Tests mit Chunked-Upload-Support
- **Service Labeling**: Automatische Unterscheidung zwischen nextcloud/hidrive/magentacloud/hidrive_legacy/dropbox Services

### 📈 **Monitoring Stack**
- **Prometheus**: Metriken-Sammlung mit 6 fokussierten Alert-Regeln
- **Grafana**: 3 spezialisierte Dashboards (Daily, Monthly, Errors)
- **Alertmanager**: E-Mail-Benachrichtigungen mit modernem HTML-Template

### 🔔 **Alerting (6 Alerts)**
| Alert | Severity | Beschreibung |
|-------|----------|--------------|
| `ServiceDown` | critical | Monitor Agent offline |
| `CloudServiceUnavailable` | critical | Cloud-Service 5xx Fehler |
| `SlowUploadSpeed` | warning | Upload <1 MB/s |
| `HighErrorRate` | warning | Fehlerrate >20% |
| `CircuitBreakerOpen` | critical | Circuit Breaker ausgelöst |
| `PrometheusStorageNearFull` | warning | Speicher >80% voll |

### 🔒 **Production-Ready**
- **Health Checks**: /health, /health/live, /health/ready Endpoints
- **Structured Logging**: JSON/Text-Format mit konfigurierbaren Levels
- **Graceful Shutdown**: Signal-basierter Shutdown mit Cleanup
- **Docker Health Checks**: Container-Monitoring für alle Services

## 🚀 Quick Start

```bash
# Repository klonen
git clone https://github.com/xXRoxXeRXx/cloud-performance-monitor.git
cd cloud-performance-monitor

# 1. Konfiguration erstellen
cp .env.example .env

# 2. .env-Datei mit deinen Credentials anpassen
nano .env

# 3. Stack bauen und starten  
make dev

# 4. Grafana öffnen
make dashboards
```

## ⚙️ Konfiguration

### `.env` Datei Beispiel
```bash
# Test-Konfiguration
TEST_FILE_SIZE_MB=100
TEST_INTERVAL_SECONDS=300
TEST_CHUNK_SIZE_MB=10

# E-Mail-Benachrichtigungen
SMTP_SMARTHOST=smtp.gmail.com:587
SMTP_FROM=alerts@your-domain.com
SMTP_AUTH_USERNAME=alerts@your-domain.com
SMTP_AUTH_PASSWORD=your-app-password
SMTP_REQUIRE_TLS=true

# E-Mail-Empfänger
EMAIL_ADMIN=admin@your-domain.com

# Nextcloud Instanzen
NC_INSTANCE_1_URL=https://cloud.example.com
NC_INSTANCE_1_USER=monitor_user
NC_INSTANCE_1_PASS=app-password

# HiDrive Instanzen
HIDRIVE_INSTANCE_1_URL=https://storage.ionos.fr
HIDRIVE_INSTANCE_1_USER=your-username
HIDRIVE_INSTANCE_1_PASS=your-password

# MagentaCLOUD Instanzen (WebDAV mit ANID)
MAGENTACLOUD_INSTANCE_1_URL=https://magentacloud.de
MAGENTACLOUD_INSTANCE_1_USER=your-email@t-online.de
MAGENTACLOUD_INSTANCE_1_ANID=120049010000000114279134
MAGENTACLOUD_INSTANCE_1_PASS=your-app-password

# HiDrive Legacy (OAuth2)
HIDRIVE_LEGACY_INSTANCE_1_URL=https://api.hidrive.strato.com/2.1
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_ID=your-oauth2-client-id
HIDRIVE_LEGACY_INSTANCE_1_CLIENT_SECRET=your-oauth2-client-secret
HIDRIVE_LEGACY_INSTANCE_1_REFRESH_TOKEN=your-refresh-token
HIDRIVE_LEGACY_INSTANCE_1_NAME=hidrive-legacy-main

# Dropbox Instanzen (OAuth2)
DROPBOX_INSTANCE_1_REFRESH_TOKEN=sl.your-dropbox-refresh-token
DROPBOX_INSTANCE_1_APP_KEY=your-app-key
DROPBOX_INSTANCE_1_APP_SECRET=your-app-secret
DROPBOX_INSTANCE_1_NAME=user@example.com
```

### Unterstützte Cloud-Services

| Service | Protokoll | Konfiguration | Setup-Anleitung |
|---------|-----------|---------------|------------------|
| **Nextcloud** | WebDAV | Username/Password | Standard WebDAV-Konfiguration |
| **HiDrive Next** | WebDAV | Username/Password | Optimiert für IONOS HiDrive |
| **MagentaCLOUD** | WebDAV | Username/Password/ANID | [📖 MagentaCLOUD Setup Guide](docs/MAGENTACLOUD_SETUP.md) |
| **HiDrive Legacy** | OAuth2 REST API | Refresh Token | [📖 HiDrive OAuth2 Setup Guide](docs/HIDRIVE_OAUTH2_SETUP.md) |
| **Dropbox** | OAuth2 REST API | Refresh Token | [📖 Dropbox Setup Guide](docs/DROPBOX_OAUTH2_SETUP.md) |

### E-Mail Provider Konfiguration
Unterstützte Provider: **Gmail**, **Outlook**, **Yahoo**, **Strato**, und andere SMTP-Server.

📧 Detaillierte Anleitung: [docs/EMAIL_CONFIGURATION.md](docs/EMAIL_CONFIGURATION.md)

## 📊 Monitoring Dashboards

Nach dem Start stehen folgende Interfaces zur Verfügung:

| Service | URL | Beschreibung |
|---------|-----|--------------|
| **Grafana** | http://localhost:3003 | Haupt-Dashboard mit Performance-Metriken |

> **Hinweis**: Prometheus (Port 9090) und Alertmanager (Port 9093) sind aus Sicherheitsgründen nur intern zugänglich. Der Zugriff erfolgt über Grafana oder Docker-interne Verbindungen.

### Grafana Dashboard Features
- **Performance Overview**: Upload/Download Geschwindigkeiten und Latenz
- **Error Analysis**: Fehlerquoten und Service-Verfügbarkeit
- **Chunk Statistics**: Detaillierte Upload-Chunk-Metriken
- **Network Analysis**: Netzwerk-Performance und Verbindungsqualität

## 🛠️ Development Commands

```bash
make dev        # Setup & start (first time)
make run        # Start monitoring stack
make stop       # Stop all services
make logs       # Show live logs
make test       # Run Go tests
make clean      # Remove containers & data
make build      # Rebuild Docker images
make restart    # Restart services
make status     # Show service status
make dashboards # Open Grafana
```

## API Endpoints

### Monitor Agent (Port 8080 - nur intern zugänglich)
```bash
# Core Endpoints (nur über Docker-internes Netzwerk)
GET /metrics              # Prometheus metrics
GET /health              # Complete health status with all services
GET /health/live         # Liveness probe (simple alive check)
GET /health/ready        # Readiness probe (ready to serve traffic)

# Zugriff für Debugging über Docker:
docker exec monitor-agent curl http://localhost:8080/health

# Example Health Response
{
  "status": "healthy",
  "timestamp": "2025-09-15T13:30:00Z",
  "uptime_seconds": 3600,
  "services": [
    {
      "name": "nextcloud-instance1",
      "status": "healthy",
      "last_check": "2025-09-15T13:29:45Z",
      "response_time_ms": 250
    }
  ],
  "version": "1.0.0"
}
```

### Environment Variables

#### Logging Configuration
```bash
# Logging Level (DEBUG, INFO, WARN, ERROR)
LOG_LEVEL=INFO

# Logging Format (text or json)
LOG_FORMAT=json
```

#### Health Check Configuration
```bash
# Health check intervals (automatically configured in docker-compose.yml)
# - interval: 30s (how often to check)
# - timeout: 10s (max time to wait for response)
# - retries: 3 (failures before marking unhealthy)
# - start_period: 10s (grace period after container start)
```

## 📈 Metrics & Alerts

### Available Metrics
```prometheus
# Performance Metrics
cloud_test_duration_seconds{service="nextcloud|hidrive|magentacloud|hidrive_legacy|dropbox",instance="url",type="upload|download"}
cloud_test_speed_mbytes_per_sec{service="nextcloud|hidrive|magentacloud|hidrive_legacy|dropbox",instance="url",type="upload|download"}
cloud_test_success{service="nextcloud|hidrive|magentacloud|hidrive_legacy|dropbox",instance="url",type="upload|download"}
cloud_test_errors_total{service="nextcloud|hidrive|magentacloud|hidrive_legacy|dropbox",instance="url",type="upload|download",error_type="..."}

# Advanced Metrics
cloud_chunks_uploaded_total{service="nextcloud|hidrive|magentacloud|hidrive_legacy|dropbox",instance="url"}
cloud_chunk_retries_total{service="nextcloud|hidrive|magentacloud|hidrive_legacy|dropbox",instance="url"}
cloud_network_latency_ms{service="nextcloud|hidrive|magentacloud|hidrive_legacy|dropbox",instance="url"}
cloud_circuit_breaker_state{service="nextcloud|hidrive|magentacloud|hidrive_legacy|dropbox",instance="url"}
```

### Alert Categories

- **🚨 Critical**: ServiceDown, CloudServiceUnavailable, CircuitBreakerOpen
- **⚠️ Warning**: SlowUploadSpeed, HighErrorRate, PrometheusStorageNearFull

## 🔒 Security Features

### Port Security
- **Minimal External Exposure**: Nur Grafana (Port 3003) extern zugänglich
- **Internal Networking**: Alle Services kommunizieren über Docker-internes Netzwerk
- **Secure by Default**: Prometheus, Alertmanager und Monitor-Agent nicht extern erreichbar

### Accessing Internal Services
```bash
# Prometheus Metrics (nur intern)
docker exec prometheus wget -qO- http://monitor-agent:8080/metrics

# Alertmanager Status (nur intern)  
docker exec alertmanager wget -qO- http://localhost:9093/api/v1/status

# Service Health Checks
docker exec monitor-agent curl http://localhost:8080/health
```

📖 Detaillierte Sicherheitsdokumentation: [docs/PORT_SECURITY.md](docs/PORT_SECURITY.md)

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Monitor Agent │───▶│   Prometheus    │───▶│     Grafana     │
│                 │    │                 │    │                 │
│ • Upload Tests  │    │ • Metrics Store │    │ • Dashboards    │
│ • Download Tests│    │ • Alert Rules   │    │ • Visualisation│
│ • Chunked Upload│    │ • Scraping      │    │ • Service Filter│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       
         │              ┌─────────────────┐              
         │              │  Alertmanager   │              
         │              │                 │              
         │              │ • Email Alerts  │              
         └──────────────▶│ • Smart Routing │              
                        │ • Suppression   │              
                        └─────────────────┘              
```

## 📚 Documentation

- � [Email Configuration](docs/EMAIL_CONFIGURATION.md) - SMTP setup guide
- 🔒 [Port Security](docs/PORT_SECURITY.md) - Security documentation
- � [Data Retention](docs/DATA_RETENTION.md) - Prometheus storage guide
- � [Dropbox OAuth2](docs/DROPBOX_OAUTH2_SETUP.md) - Dropbox setup
- 🔐 [HiDrive OAuth2](docs/HIDRIVE_OAUTH2_SETUP.md) - HiDrive Legacy setup
- � [MagentaCLOUD](docs/MAGENTACLOUD_SETUP.md) - MagentaCLOUD setup

## 🔧 Requirements

- **Docker** & **Docker Compose**
- **Go 1.22+** (for development)
- **Nextcloud/HiDrive** instances with WebDAV access
- **SMTP Server** access (for email notifications)

## 🐛 Troubleshooting

### Common Issues

**Services not starting?**
```bash
make status         # Check service status
make logs          # View logs
docker system prune # Clean up Docker
```

**No metrics in Grafana?**
```bash
make logs          # Check agent logs
docker exec prometheus wget -qO- http://monitor-agent:8080/metrics | head -10
```

**Email notifications not working?**
```bash
docker compose logs alertmanager
# Check docs/EMAIL_CONFIGURATION.md for provider-specific settings
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/agent/
go test ./internal/nextcloud/

# Run with coverage
go test -cover ./...
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⭐ Support

If this project helps you, please consider giving it a star! ⭐

For issues and feature requests, please use the [GitHub Issues](https://github.com/xXRoxXeRXx/cloud-performance-monitor/issues) page.
