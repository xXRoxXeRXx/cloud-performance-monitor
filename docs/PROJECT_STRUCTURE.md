# Project Structure

## Directory Layout

```
cloud-performance-monitor/
├── .github/                    # GitHub workflows and templates
├── alertmanager/              # Alertmanager configuration
│   ├── Dockerfile            # Custom alertmanager with email support
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
│   │   ├── dashboard.json  # Enhanced monitoring dashboard
│   │   └── provisioning/   # Auto-provisioning configs
│   └── webhook-logger/     # Webhook logger deployment
│       └── Dockerfile      # Webhook logger container
├── docs/                   # Documentation
│   └── EMAIL_CONFIGURATION.md  # Email setup guide
├── internal/               # Internal packages
│   ├── agent/             # Core agent logic
│   │   ├── config.go      # Configuration management
│   │   ├── metrics.go     # Prometheus metrics
│   │   ├── tester.go      # Test orchestration
│   │   ├── hidrive_tester.go      # HiDrive-specific tests
│   │   ├── magentacloud_tester.go # MagentaCLOUD-specific tests
│   │   ├── hidrive_legacy_tester.go # HiDrive Legacy tests
│   │   └── dropbox_tester.go      # Dropbox-specific tests
│   ├── nextcloud/         # Nextcloud WebDAV client
│   │   └── client.go      # Nextcloud API implementation
│   ├── hidrive/           # HiDrive WebDAV client
│   │   └── client.go      # HiDrive API implementation
│   ├── magentacloud/      # MagentaCLOUD WebDAV client
│   │   └── client.go      # MagentaCLOUD API with ANID support
│   ├── hidrive_legacy/    # HiDrive Legacy OAuth2 client
│   │   └── client.go      # HiDrive Legacy API implementation
│   └── dropbox/           # Dropbox REST API client
│       └── client.go      # Dropbox API implementation
├── prometheus/            # Prometheus configuration
│   ├── prometheus.yml     # Prometheus config with alerting
│   └── alert_rules.yml    # Comprehensive alert rules
├── .env.example          # Environment variables template
├── docker-compose.yml    # Complete monitoring stack
├── Dockerfile            # Main agent container
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
