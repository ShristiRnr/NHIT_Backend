# Makefile for NHIT Backend Microservices

.PHONY: help proto build run-user run-auth run-org docker-up docker-down clean test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

proto: ## Generate protobuf code (gRPC + Gateway)
	@echo "Generating protobuf code with gRPC gateway..."
	protoc -I . -I third_party/googleapis \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		api/proto/user_management.proto
	protoc -I . -I third_party/googleapis \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		api/proto/auth.proto

proto-swagger: ## Generate OpenAPI/Swagger documentation
	@echo "Generating Swagger docs..."
	protoc -I . -I third_party/googleapis \
		--openapiv2_out=. \
		--openapiv2_opt=allow_merge=true \
		--openapiv2_opt=merge_file_name=api_docs \
		api/proto/*.proto

sqlc: ## Generate sqlc code
	@echo "Generating sqlc code..."
	sqlc generate

build-user: ## Build user service
	@echo "Building user service..."
	go build -o bin/user-service ./services/user-service/cmd/server

build-auth: ## Build auth service
	@echo "Building auth service..."
	go build -o bin/auth-service ./services/auth-service/cmd/server

build-org: ## Build organization service
	@echo "Building organization service..."
	go build -o bin/organization-service ./services/organization-service/cmd/server

build-gateway: ## Build API gateway
	@echo "Building API gateway..."
	go build -o bin/api-gateway ./services/api-gateway/cmd/server

build: build-user build-auth build-org build-gateway ## Build all services

run-user: ## Run user service
	@echo "Starting user service..."
	go run ./services/user-service/cmd/server/main.go

run-auth: ## Run auth service
	@echo "Starting auth service..."
	go run ./services/auth-service/cmd/server/main.go

run-org: ## Run organization service
	@echo "Starting organization service..."
	go run ./services/organization-service/cmd/server/main.go

run-gateway: ## Run API gateway
	@echo "Starting API gateway..."
	go run ./services/api-gateway/cmd/server/main.go

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker-compose build

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	docker-compose up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	# Add migration command here

migrate-down: ## Run database migrations down
	@echo "Rolling back migrations..."
	# Add migration rollback command here

dev-user: ## Run user service in development mode
	air -c services/user-service/.air.toml

dev-auth: ## Run auth service in development mode
	air -c services/auth-service/.air.toml

dev-org: ## Run organization service in development mode
	air -c services/organization-service/.air.toml
