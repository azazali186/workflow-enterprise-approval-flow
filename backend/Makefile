.PHONY: run build test proto-gen lint fmt clean deps help

APP_NAME := approval-flow
PORT := 8080
GRPC_PORT := 9090

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./... -v -cover

test-integration:
	go test ./... -tags=integration -v

proto-gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

lint:
	golangci-lint run

fmt:
	go fmt ./...

clean:
	rm -rf bin/
	rm -f coverage.out

deps:
	go mod tidy

help:
	@echo "Available targets:"
	@echo "  run           - Run the server"
	@echo "  build         - Build the server"
	@echo "  test          - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  proto-gen     - Generate protobuf files"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
