.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build all services
	docker-compose build

.PHONY: up
up: ## Start all services
	docker-compose up -d

.PHONY: down
down: ## Stop all services
	docker-compose down

.PHONY: logs
logs: ## Show logs from all services
	docker-compose logs -f

.PHONY: ps
ps: ## Show running services
	docker-compose ps

.PHONY: migrate
migrate: ## Run database migrations
	docker-compose exec postgres psql -U postgres -d crm_dialer -f /docker-entrypoint-initdb.d/001_initial_schema.sql
	docker-compose exec postgres psql -U postgres -d crm_dialer -f /docker-entrypoint-initdb.d/002_add_entity_type.sql
	docker-compose exec postgres psql -U postgres -d crm_dialer -f /docker-entrypoint-initdb.d/003_add_users_table.sql

.PHONY: migrate-create
migrate-create: ## Create a new migration file (usage: make migrate-create name=add_new_table)
	@if [ -z "$(name)" ]; then echo "Error: name parameter is required. Usage: make migrate-create name=add_new_table"; exit 1; fi
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="migrations/$${timestamp}_$(name).sql"; \
	touch $$filename; \
	echo "Created migration file: $$filename"

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -cover ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	go mod tidy

.PHONY: mod-download
mod-download: ## Download go modules
	go mod download

.PHONY: run-local-gateway
run-local-gateway: ## Run API Gateway locally
	go run cmd/api-gateway/main.go

.PHONY: run-local-crm
run-local-crm: ## Run CRM Service locally
	go run cmd/crm-service/main.go

.PHONY: run-local-webhook
run-local-webhook: ## Run Webhook Service locally
	go run cmd/webhook-service/main.go

.PHONY: run-local-flow
run-local-flow: ## Run Flow Engine Service locally
	go run cmd/flow-engine-service/main.go

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	docker-compose down -v
	docker system prune -f

.PHONY: frontend-install
frontend-install: ## Install frontend dependencies
	cd web && npm install

.PHONY: frontend-dev
frontend-dev: ## Run frontend in development mode
	cd web && npm start

.PHONY: frontend-build
frontend-build: ## Build frontend for production
	cd web && npm run build

.PHONY: frontend-test
frontend-test: ## Run frontend tests
	cd web && npm test

.PHONY: setup
setup: ## Initial setup of the project
	cp .env.example .env
	@echo "Please edit .env file with your configuration"
	make mod-download
	make frontend-install
	make up
	sleep 5
	make migrate

.PHONY: restart
restart: ## Restart all services
	make down
	make up

.PHONY: restart-service
restart-service: ## Restart specific service (usage: make restart-service service=api-gateway)
	@if [ -z "$(service)" ]; then echo "Error: service parameter is required. Usage: make restart-service service=api-gateway"; exit 1; fi
	docker-compose restart $(service)

.PHONY: shell
shell: ## Open shell in service container (usage: make shell service=api-gateway)
	@if [ -z "$(service)" ]; then echo "Error: service parameter is required. Usage: make shell service=api-gateway"; exit 1; fi
	docker-compose exec $(service) /bin/sh

.PHONY: db-shell
db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U postgres -d crm_dialer

.PHONY: redis-cli
redis-cli: ## Open Redis CLI
	docker-compose exec redis redis-cli

.PHONY: nats-cli
nats-cli: ## Open NATS CLI
	docker-compose exec nats nats-cli

.PHONY: backup-db
backup-db: ## Backup database
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	docker-compose exec postgres pg_dump -U postgres crm_dialer > backups/db_backup_$$timestamp.sql; \
	echo "Database backed up to backups/db_backup_$$timestamp.sql"

.PHONY: restore-db
restore-db: ## Restore database from backup (usage: make restore-db file=backups/db_backup_20240101_120000.sql)
	@if [ -z "$(file)" ]; then echo "Error: file parameter is required. Usage: make restore-db file=backups/db_backup_20240101_120000.sql"; exit 1; fi
	docker-compose exec -T postgres psql -U postgres crm_dialer < $(file)

.PHONY: monitoring
monitoring: ## Open monitoring dashboards
	@echo "Opening monitoring dashboards..."
	@echo "Grafana: http://localhost:3001 (admin/admin)"
	@echo "Prometheus: http://localhost:9090"
	@open http://localhost:3001 || xdg-open http://localhost:3001 || echo "Please open http://localhost:3001 in your browser"