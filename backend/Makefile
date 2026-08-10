.PHONY: run build test test-race test-coverage test-integration proto-gen lint fmt clean deps help migrate migrate-down migrate-version seed swagger swagger-gen

APP_NAME := approval-flow
PORT := 8080
GRPC_PORT := 9090
MIGRATIONS_DIR := ./migrations

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/aeroxe/approval-flow/internal/config.Version=$(VERSION)-$(COMMIT) \
	-X github.com/aeroxe/approval-flow/internal/config.BuildTime=$(BUILD_TIME)

run:
	go run cmd/server/main.go

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/server cmd/server/main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/server-linux-amd64 cmd/server/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/server-linux-arm64 cmd/server/main.go

test:
	go test ./... -v -count=1

test-race:
	go test ./... -v -race -count=1

test-coverage:
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration:
	go test ./... -tags=integration -v

test-short:
	go test ./... -v -short -count=1

proto-gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

fmt:
	go fmt ./...
	goimports -w .

vet:
	go vet ./...

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f routes.json

deps:
	go mod download
	go mod tidy

deps-verify:
	go mod verify

# Database migrations
migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq -digits 6 $$name

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-down-all:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down -all

migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

migrate-force:
	@read -p "Version number: " version; \
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $$version

# Docker
docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Development
dev:
	go run cmd/server/main.go

seed:
	@echo "Seeding is done automatically on server start"

swagger:
	@echo "Swagger docs served at /docs/swagger.json"

# Regenerate Swagger docs from handler annotations.
# --parseInternal is required: response/validation types live in internal packages,
# and without it swag fails on json.RawMessage and gorm soft_delete types.
swagger-gen:
	swag init -g cmd/server/main.go --parseInternal --outputTypes json,yaml,go --output docs

# Production
install-golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

install-tools: install-golangci-lint install-swag install-migrate
	@echo "All tools installed"

help:
	@echo "Available targets:"
	@echo ""
	@echo "  run              - Run the server"
	@echo "  build            - Build the server"
	@echo "  build-linux      - Build for Linux (amd64 + arm64)"
	@echo "  test             - Run unit tests"
	@echo "  test-race        - Run tests with race detector"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-integration - Run integration tests"
	@echo "  test-short       - Run short tests"
	@echo ""
	@echo "  proto-gen        - Generate protobuf files"
	@echo "  lint             - Run linter"
	@echo "  lint-fix         - Run linter with auto-fix"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  clean            - Clean build artifacts"
	@echo "  deps             - Download dependencies"
	@echo "  deps-verify      - Verify dependencies"
	@echo ""
	@echo "  migrate-create   - Create new migration"
	@echo "  migrate-up       - Run pending migrations"
	@echo "  migrate-down     - Rollback last migration"
	@echo "  migrate-down-all - Rollback all migrations"
	@echo "  migrate-version  - Show current migration version"
	@echo ""
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Start with Docker Compose"
	@echo "  docker-stop      - Stop Docker Compose"
	@echo "  docker-logs      - View Docker logs"
	@echo ""
	@echo "  dev              - Run in development mode"
	@echo "  seed             - Seed database"
	@echo "  swagger          - Show swagger docs info"
	@echo "  swagger-gen      - Regenerate swagger docs (swag init --parseInternal)"
	@echo ""
	@echo "  install-tools    - Install development tools"
