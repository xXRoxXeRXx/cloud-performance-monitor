# Project Structure

## Directory Layout

```
cloud-performance-monitor/
├── .github/                    # GitHub workflows and templates
│   └── copilot-instructions.md # AI agent instructions
├── alertmanager/              # Alertmanager configuration
│   ├── Dockerfile            # Custom alertmanager with email support
│   ├── alertmanager.yml      # Active alertmanager config
│   ├── alertmanager.yml.template  # Email configuration template
│   └── start-alertmanager.sh # Startup script with env substitution
├── cmd/                      # Application entry points
│   ├── agent/               # Main monitoring agent
│   │   └── main.go         # Agent entrypoint
│   └── webhook-logger/      # Alert webhook logger
│       └── main.go         # Webhook logger entrypoint
├── deploy/                  # Deployment configurations
│   ├── grafana/            # Grafana setup
│   │   ├── Dockerfile      # Custom grafana with dashboards
│   │   ├── dashboard.json  # Main performance dashboard
│   │   ├── enhanced-dashboard.json  # Enhanced analytics
│   │   ├── dashboard-enhanced.json  # Enhanced alt
│   │   ├── errors-dashboard.json    # Error analysis
│   │   ├── daily-performance-dashboard.json   # Daily raw data
│   │   ├── monthly-performance-dashboard.json # Monthly trends
│   │   ├── README.md       # Dashboard documentation
│   │   └── provisioning/   # Auto-provisioning configs
│   │       ├── dashboards/
│   │       └── datasources/
│   └── webhook-logger/     # Webhook logger deployment
│       └── Dockerfile      # Webhook logger container
├── docs/                   # Documentation
│   ├── ALERT_ERROR_CODES.md    # Error code reference
│   ├── ALERT_OVERVIEW.md       # Alert documentation
│   ├── DATA_RETENTION.md       # Data retention guide
│   ├── DROPBOX_OAUTH2_SETUP.md # Dropbox OAuth2 setup
│   ├── EMAIL_CONFIGURATION.md  # Email/SMTP setup guide
│   ├── HIDRIVE_OAUTH2_SETUP.md # HiDrive Legacy OAuth2 setup
│   ├── HISTORICAL_ANALYTICS.md # Analytics documentation
│   ├── LOGGING_SCHEMA.md       # Logging schema reference
│   ├── MAGENTACLOUD_SETUP.md   # MagentaCLOUD setup guide
│   ├── PORT_SECURITY.md        # Security documentation
│   ├── PROJECT_STRUCTURE.md    # This file
│   └── wiki/               # GitHub Wiki content
│       ├── Home.md             # Wiki homepage
│       ├── Error-Code-Reference.md # Error codes
│       ├── Runbook-CriticalUploadDuration.md
│       ├── Runbook-ServiceDown.md
│       └── Runbook-ServiceTestFailure.md
├── internal/               # Internal packages
│   ├── agent/             # Core agent logic
│   │   ├── config.go      # Configuration management
│   │   ├── config_test.go
│   │   ├── metrics.go     # Prometheus metrics
│   │   ├── tester.go      # Test orchestration
│   │   ├── hidrive_tester.go      # HiDrive-specific tests
│   │   ├── magentacloud_tester.go # MagentaCLOUD-specific tests
│   │   ├── hidrive_legacy_tester.go # HiDrive Legacy tests
│   │   ├── dropbox_tester.go      # Dropbox-specific tests
│   │   ├── dropbox_test.go
│   │   ├── error_codes.go # Error classification
│   │   ├── error_codes_test.go
│   │   ├── health.go      # Health check endpoints
│   │   ├── health_test.go
│   │   ├── logger.go      # Structured logging
│   │   ├── logger_test.go
│   │   ├── network_monitoring.go  # Network metrics
│   │   ├── shutdown.go    # Graceful shutdown
│   │   ├── validation.go  # Config validation
│   │   └── integration_test.go
│   ├── nextcloud/         # Nextcloud WebDAV client
│   │   ├── client.go      # Nextcloud API implementation
│   │   └── client_test.go
│   ├── hidrive/           # HiDrive WebDAV client
│   │   ├── client.go      # HiDrive API implementation
│   │   └── client_test.go
│   ├── magentacloud/      # MagentaCLOUD WebDAV client
│   │   ├── client.go      # MagentaCLOUD API with ANID support
│   │   └── client_test.go
│   ├── hidrive_legacy/    # HiDrive Legacy OAuth2 client
│   │   ├── client.go      # HiDrive Legacy API implementation
│   │   └── client_test.go
│   ├── dropbox/           # Dropbox REST API client
│   │   ├── client.go      # Dropbox API implementation
│   │   └── client_test.go
│   └── utils/             # Shared utilities
│       ├── circuit_breaker.go # Circuit breaker pattern
│       ├── client_logger.go   # Client logger interface
│       ├── retry.go       # Retry logic
│       └── retry_test.go
├── prometheus/            # Prometheus configuration
│   ├── prometheus.yml     # Prometheus config with alerting
│   ├── alert_rules.yml    # Comprehensive alert rules
│   └── recording_rules.yml # Recording rules for aggregations
├── .env.example          # Environment variables template
├── docker-compose.yml    # Complete monitoring stack
├── Dockerfile            # Main agent container
├── Makefile              # Development commands
├── go.mod               # Go module definition
├── go.sum               # Go dependencies
├── LICENSE              # MIT License
└── README.md            # Main documentation
```

## Key Components

### 🔧 **Core Application**
- **cmd/agent/main.go**: Main monitoring agent that tests Nextcloud/HiDrive instances
- **internal/**: Business logic separated by service type
- **Dockerfile**: Containerized agent for easy deployment

### 📊 **Monitoring Stack**
- **prometheus/**: Metrics collection and alerting rules
- **alertmanager/**: Email notifications with environment variable configuration
- **deploy/grafana/**: Enhanced dashboards with performance visualization

### 🔔 **Alerting & Notifications**
- **15+ alert rules** covering availability, performance, errors, network, and SLA
- **Email notifications** via SMTP with dynamic configuration
- **Webhook logger** for alert testing and debugging

### 📋 **Configuration**
- **Environment-based configuration** via `.env` files
- **Template-based** alertmanager setup for flexibility
- **Auto-provisioned** Grafana dashboards

## Development Guidelines

### 🏗️ **Building**
```bash
# Build agent
go build -o bin/agent cmd/agent/main.go

# Build with Docker
docker compose build
```

### 🚀 **Deployment**
```bash
# Start complete stack
docker compose up -d

# View logs
docker compose logs -f [service-name]
```

### 🧪 **Testing**
- All test files follow `*_test.go` pattern
- Unit tests in `internal/` packages
- Integration tests with real WebDAV endpoints

### 📝 **Configuration Management**
- Use `.env.example` as template for new environments
- Never commit actual `.env` files with credentials
- Environment variables are documented in `.env.example`

## Services & Ports

| Service | Port | Purpose |
|---------|------|---------|
| monitor-agent | 8080 | Metrics endpoint |
| prometheus | 9090 | Metrics collection UI |
| alertmanager | 9093 | Alert management UI |
| grafana | 3003 | Dashboard UI |
| webhook-logger | 8081 | Alert webhook testing |

## Monitoring Targets

- **Nextcloud instances**: WebDAV upload/download performance
- **HiDrive instances**: Cloud storage performance testing
- **Network latency**: Connection quality monitoring
- **Error rates**: Service availability tracking
- **SLA compliance**: Uptime and performance thresholds
