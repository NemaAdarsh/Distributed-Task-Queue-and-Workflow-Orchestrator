.PHONY: build test clean run-scheduler run-worker build-docker run-compose

# Build targets
build:
	go build -o bin/scheduler cmd/scheduler/main.go
	go build -o bin/worker cmd/worker/main.go

build-web:
	cd web/dashboard && npm install && npm run build

build-all: build build-web

# Test targets
test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Development targets
run-scheduler:
	./bin/scheduler \
		-postgres="postgres://flowctl:password@localhost/flowctl?sslmode=disable" \
		-redis="localhost:6379" \
		-api=":8080"

run-worker-etl:
	./bin/worker -types="etl" -addr="localhost:9001"

run-worker-ml:
	./bin/worker -types="ml_training" -addr="localhost:9002"

run-worker-ci:
	./bin/worker -types="ci" -addr="localhost:9003"

run-worker-generic:
	./bin/worker -types="generic" -addr="localhost:9004"

# Docker targets
build-docker:
	docker build -t flowctl-scheduler -f Dockerfile.scheduler .
	docker build -t flowctl-worker -f Dockerfile.worker .

run-compose:
	docker-compose up -d

stop-compose:
	docker-compose down

# Database targets
db-setup:
	createdb flowctl
	psql flowctl -c "CREATE USER flowctl WITH PASSWORD 'password';"
	psql flowctl -c "GRANT ALL PRIVILEGES ON DATABASE flowctl TO flowctl;"

db-migrate:
	go run cmd/migrate/main.go

# Development setup
dev-setup: db-setup build build-web
	@echo "Development environment ready!"

# Clean targets
clean:
	rm -rf bin/
	rm -rf web/dashboard/build/
	rm -f coverage.out

# Linting and formatting
fmt:
	go fmt ./...

lint:
	golangci-lint run

# Dependencies
deps:
	go mod download
	go mod tidy

# Example workflows
run-example-etl:
	curl -X POST http://localhost:8080/api/v1/workflows \
		-H "Content-Type: application/json" \
		-d @examples/etl_pipeline_api.json

run-example-ml:
	curl -X POST http://localhost:8080/api/v1/workflows \
		-H "Content-Type: application/json" \
		-d @examples/ml_training_api.json

# Monitoring
logs-scheduler:
	docker-compose logs -f scheduler

logs-worker:
	docker-compose logs -f worker

logs-redis:
	docker-compose logs -f redis

logs-postgres:
	docker-compose logs -f postgres

# Production targets
release:
	goreleaser release --rm-dist

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build scheduler and worker binaries"
	@echo "  build-web      - Build web dashboard"
	@echo "  build-all      - Build everything"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  run-scheduler  - Run scheduler locally"
	@echo "  run-worker-*   - Run specific worker type"
	@echo "  build-docker   - Build Docker images"
	@echo "  run-compose    - Start with Docker Compose"
	@echo "  db-setup       - Setup local database"
	@echo "  dev-setup      - Complete development setup"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter"
	@echo "  help           - Show this help"
