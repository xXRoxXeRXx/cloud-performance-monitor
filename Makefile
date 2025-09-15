# Nextcloud Performance Monitor - Makefile
# Simple automation for common development tasks

.PHONY: help build clean test run logs stop restart status

# Default target
help:
	@echo "Nextcloud Performance Monitor - Available Commands:"
	@echo ""
	@echo "🏗️  Building:"
	@echo "  make build          - Build all Docker images"
	@echo "  make build-agent    - Build only the monitoring agent"
	@echo ""
	@echo "🚀 Running:"
	@echo "  make run           - Start the complete monitoring stack"
	@echo "  make stop          - Stop all services"
	@echo "  make restart       - Restart all services"
	@echo "  make status        - Show service status"
	@echo ""
	@echo "📊 Monitoring:"
	@echo "  make logs          - Show logs from all services"
	@echo "  make logs-agent    - Show only agent logs"
	@echo "  make logs-alerts   - Show alertmanager logs"
	@echo "  make metrics       - Open Prometheus metrics"
	@echo "  make alerts        - Open Alertmanager UI"
	@echo "  make dashboards    - Open Grafana dashboards"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test          - Run Go tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make test-alert    - Send test alert"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  make clean         - Remove containers and volumes"
	@echo "  make clean-images  - Remove Docker images"
	@echo "  make clean-all     - Full cleanup"

# Building
build:
	docker compose build

build-agent:
	docker compose build monitor-agent

# Running
run:
	docker compose up -d

stop:
	docker compose down

restart:
	docker compose restart

status:
	docker compose ps

# Monitoring
logs:
	docker compose logs -f

logs-agent:
	docker compose logs -f monitor-agent

logs-alerts:
	docker compose logs -f alertmanager

metrics:
	@echo "Opening Prometheus at http://localhost:9090"
	@start http://localhost:9090 2>/dev/null || open http://localhost:9090 2>/dev/null || xdg-open http://localhost:9090 2>/dev/null || echo "Please open http://localhost:9090"

alerts:
	@echo "Opening Alertmanager at http://localhost:9093"
	@start http://localhost:9093 2>/dev/null || open http://localhost:9093 2>/dev/null || xdg-open http://localhost:9093 2>/dev/null || echo "Please open http://localhost:9093"

dashboards:
	@echo "Opening Grafana at http://localhost:3003"
	@start http://localhost:3003 2>/dev/null || open http://localhost:3003 2>/dev/null || xdg-open http://localhost:3003 2>/dev/null || echo "Please open http://localhost:3003"

# Testing
test:
	go test -v -cover ./...

test-integration:
	go test -tags=integration -v ./internal/agent/

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-alert:
	@echo "Sending test alert..."
	@curl -X POST http://localhost:9093/api/v2/alerts \
		-H "Content-Type: application/json" \
		-d '[{"labels":{"alertname":"TestAlert","severity":"critical","instance":"test.example.com","service":"test"},"annotations":{"summary":"Test Alert from Makefile","description":"This is a test alert sent via make test-alert"},"startsAt":"'$$(date -u +%Y-%m-%dT%H:%M:%S)'.000Z"}]' \
		&& echo "Test alert sent successfully!" || echo "Failed to send test alert"

# Maintenance
clean:
	docker compose down -v

clean-images:
	docker compose down --rmi all

clean-all: clean-images
	docker system prune -f --volumes

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "✅ Created .env file from template"
	@echo "📝 Please edit .env with your configuration"
	@echo "🚀 Run 'make run' to start the stack"

install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed"

lint:
	golangci-lint run ./...

# Quick start for new users
quick-start: dev-setup build run
	@echo ""
	@echo "🎉 Quick start completed!"
	@echo ""
	@echo "📊 Access your monitoring stack:"
	@echo "  - Grafana:     http://localhost:3003"
	@echo "  - Prometheus:  http://localhost:9090"
	@echo "  - Alertmanager: http://localhost:9093"
	@echo ""
	@echo "📝 Next steps:"
	@echo "  1. Edit .env with your Nextcloud/HiDrive credentials"
	@echo "  2. Configure email settings in .env"
	@echo "  3. Run 'make restart' to apply changes"
	@echo "  4. Run 'make test-alert' to test notifications"
