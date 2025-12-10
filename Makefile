# Cloud Performance Monitor - Makefile
# Simplified automation for common tasks

.PHONY: help build run stop logs test clean dev

# Default target
help:
	@echo "Cloud Performance Monitor - Commands:"
	@echo ""
	@echo "  make dev      - Setup & start (first time)"
	@echo "  make run      - Start monitoring stack"
	@echo "  make stop     - Stop all services"
	@echo "  make logs     - Show live logs"
	@echo "  make test     - Run Go tests"
	@echo "  make clean    - Remove containers & data"
	@echo ""
	@echo "  make build    - Rebuild Docker images"
	@echo "  make restart  - Restart services"
	@echo "  make status   - Show service status"
	@echo "  make dashboards - Open Grafana"

# Building
build:
	docker compose build

# Running
dev:
	@if not exist .env (copy .env.example .env && echo Created .env - please configure it)
	docker compose build
	docker compose up -d
	@echo ""
	@echo "Started! Open http://localhost:3003 for Grafana"

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

dashboards:
	@echo "Opening Grafana at http://localhost:3003"
	@start http://localhost:3003 2>/dev/null || open http://localhost:3003 2>/dev/null || xdg-open http://localhost:3003 2>/dev/null || echo "Open http://localhost:3003"

# Testing
test:
	go test -v -cover ./...

# Maintenance
clean:
	docker compose down -v
	@echo "Containers and volumes removed"
