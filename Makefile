# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=go-base-ms
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/server/main.go

# Build variables
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

# Docker variables
DOCKER_IMAGE=go-base-ms
DOCKER_TAG?=latest

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: all build clean test coverage run docker-build docker-run deps fmt vet lint help

# Default target
all: clean deps fmt vet test openapi build

# Build the binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BINARY_PATH)$(NC)"

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -rf bin/
	rm -rf coverage/
	@echo "$(GREEN)Clean complete$(NC)"

# Run tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v -race -timeout 30s ./...

# Run tests with coverage
coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@mkdir -p coverage
	$(GOTEST) -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "$(GREEN)Coverage report: coverage/coverage.html$(NC)"

# Run the application
run: build
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	./$(BINARY_PATH)

# Run with environment variables
run-dev:
	@echo "$(GREEN)Running in development mode...$(NC)"
	LOG_LEVEL=debug \
	DB_HOST=localhost \
	DB_PORT=5432 \
	DB_USER=postgres \
	DB_PASSWORD=postgres \
	DB_NAME=gobase \
	DB_MAX_OPEN_CONNS=25 \
	DB_MAX_IDLE_CONNS=5 \
	DB_CONN_MAX_LIFETIME=5 \
	KAFKA_BROKERS=localhost:9092 \
	KAFKA_SECURITY_PROTOCOL=PLAINTEXT \
	SCHEMA_REGISTRY_URL=http://localhost:8081 \
	$(GOCMD) run $(MAIN_PATH)

# Docker build
docker-build:
	@echo "$(GREEN)Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Docker run
docker-run:
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -p 8080:8080 \
		-e DB_HOST=host.docker.internal \
		-e DB_PORT=5432 \
		-e DB_USER=postgres \
		-e DB_PASSWORD=postgres \
		-e DB_NAME=gobase \
		-e DB_MAX_OPEN_CONNS=25 \
		-e DB_MAX_IDLE_CONNS=5 \
		-e DB_CONN_MAX_LIFETIME=5 \
		-e KAFKA_BROKERS=host.docker.internal:9092 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOGET) github.com/lib/pq
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GOCMD) fmt ./...

# Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GOCMD) vet ./...

# Run linter (requires golangci-lint)
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)" && exit 1)
	golangci-lint run

# Generate mocks (if needed)
mocks:
	@echo "$(GREEN)Generating mocks...$(NC)"
	@echo "$(YELLOW)No mocks to generate yet$(NC)"

# Generate OpenAPI specification
openapi:
	@echo "$(GREEN)Generating OpenAPI specification...$(NC)"
	@./scripts/merge-openapi.sh

# Generate README from template
readme:
	@echo "$(GREEN)Generating README from template...$(NC)"
	@./scripts/generate-readme.sh

# Database migrations (placeholder)
migrate-up:
	@echo "$(GREEN)Running database migrations...$(NC)"
	@echo "$(YELLOW)No migrations defined yet$(NC)"

migrate-down:
	@echo "$(GREEN)Rolling back database migrations...$(NC)"
	@echo "$(YELLOW)No migrations defined yet$(NC)"

# Start local development environment
dev-env:
	@echo "$(GREEN)Starting development environment...$(NC)"
	@echo "$(YELLOW)Starting PostgreSQL...$(NC)"
	docker run -d --name postgres-dev \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=gobase \
		-p 5432:5432 \
		postgres:15-alpine || echo "$(YELLOW)PostgreSQL container already exists$(NC)"
	@echo "$(YELLOW)Starting Zookeeper...$(NC)"
	docker run -d --name zookeeper-dev \
		-p 2181:2181 \
		-e ZOOKEEPER_CLIENT_PORT=2181 \
		-e ZOOKEEPER_TICK_TIME=2000 \
		confluentinc/cp-zookeeper:latest || echo "$(YELLOW)Zookeeper container already exists$(NC)"
	@echo "$(YELLOW)Starting Kafka...$(NC)"
	docker run -d --name kafka-dev \
		--link zookeeper-dev:zookeeper \
		-p 9092:9092 \
		-e KAFKA_BROKER_ID=1 \
		-e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
		-e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
		-e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
		-e KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
		-e KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 \
		confluentinc/cp-kafka:latest || echo "$(YELLOW)Kafka container already exists$(NC)"
	@echo "$(YELLOW)Starting Schema Registry...$(NC)"
	docker run -d --name schema-registry-dev \
		--link zookeeper-dev:zookeeper \
		--link kafka-dev:kafka \
		-p 8081:8081 \
		-e SCHEMA_REGISTRY_HOST_NAME=schema-registry \
		-e SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS=kafka:9092 \
		-e SCHEMA_REGISTRY_LISTENERS=http://0.0.0.0:8081 \
		confluentinc/cp-schema-registry:latest || echo "$(YELLOW)Schema Registry container already exists$(NC)"
	@echo "$(GREEN)Development environment started$(NC)"

# Stop local development environment
dev-env-stop:
	@echo "$(RED)Stopping development environment...$(NC)"
	docker stop postgres-dev zookeeper-dev kafka-dev schema-registry-dev || true
	docker rm postgres-dev zookeeper-dev kafka-dev schema-registry-dev || true
	@echo "$(GREEN)Development environment stopped$(NC)"

# Install development tools
install-tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/goreleaser/goreleaser@latest
	@echo "$(GREEN)Tools installed$(NC)"

# GoReleaser targets
.PHONY: release-init release-dry-run release-snapshot release-patch release-minor release-major

# Initialize first release (run once)
release-init:
	@echo "$(GREEN)Initializing first release...$(NC)"
	@./scripts/init-release.sh

# Dry run release (test without actually releasing)
release-dry-run:
	@echo "$(GREEN)Running GoReleaser dry run...$(NC)"
	goreleaser release --snapshot --clean --skip=publish

# Create a snapshot release (no tags, for testing)
release-snapshot:
	@echo "$(GREEN)Creating snapshot release...$(NC)"
	goreleaser release --snapshot --clean

# Create a patch release (x.x.X)
release-patch:
	@echo "$(GREEN)Creating patch release...$(NC)"
	@$(MAKE) check-git-clean
	@$(MAKE) create-tag INCREMENT=patch
	goreleaser release --clean

# Create a minor release (x.X.0)  
release-minor:
	@echo "$(GREEN)Creating minor release...$(NC)"
	@$(MAKE) check-git-clean
	@$(MAKE) create-tag INCREMENT=minor
	goreleaser release --clean

# Create a major release (X.0.0)
release-major:
	@echo "$(GREEN)Creating major release...$(NC)"
	@echo "$(RED)WARNING: This will create a major version bump!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(MAKE) check-git-clean; \
		$(MAKE) create-tag INCREMENT=major; \
		goreleaser release --clean; \
	else \
		echo "$(YELLOW)Major release cancelled$(NC)"; \
	fi

# Check if git working directory is clean
check-git-clean:
	@echo "$(GREEN)Checking git status...$(NC)"
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "$(RED)Git working directory is not clean. Please commit or stash changes.$(NC)"; \
		git status --short; \
		exit 1; \
	fi
	@echo "$(GREEN)Git working directory is clean$(NC)"

# Create and push git tag with version increment
create-tag:
	@echo "$(GREEN)Creating git tag...$(NC)"
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "Latest tag: $$LATEST_TAG"; \
	LATEST_VERSION=$$(echo $$LATEST_TAG | sed 's/v//'); \
	IFS='.' read -r MAJOR MINOR PATCH <<< "$$LATEST_VERSION"; \
	if [ "$(INCREMENT)" = "major" ]; then \
		NEW_TAG="v$$((MAJOR + 1)).0.0"; \
	elif [ "$(INCREMENT)" = "minor" ]; then \
		NEW_TAG="v$$MAJOR.$$((MINOR + 1)).0"; \
	else \
		NEW_TAG="v$$MAJOR.$$MINOR.$$((PATCH + 1))"; \
	fi; \
	echo "Creating new tag: $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG"; \
	git push origin $$NEW_TAG; \
	echo "$(GREEN)Tag $$NEW_TAG created and pushed$(NC)"

# Get current version
version:
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "Current version: $$LATEST_TAG"

# Get next versions
next-versions:
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	LATEST_VERSION=$$(echo $$LATEST_TAG | sed 's/v//'); \
	IFS='.' read -r MAJOR MINOR PATCH <<< "$$LATEST_VERSION"; \
	echo "Current version: $$LATEST_TAG"; \
	echo "Next patch:      v$$MAJOR.$$MINOR.$$((PATCH + 1))"; \
	echo "Next minor:      v$$MAJOR.$$((MINOR + 1)).0"; \
	echo "Next major:      v$$((MAJOR + 1)).0.0"

# Clean release artifacts
release-clean:
	@echo "$(GREEN)Cleaning release artifacts...$(NC)"
	rm -rf dist/
	@echo "$(GREEN)Release artifacts cleaned$(NC)"

# Initialize new project from template
init-project:
	@echo "$(GREEN)Initializing new project from template...$(NC)"
	@./scripts/init-project.sh

# Help
help:
	@echo "$(GREEN)Available targets:$(NC)"
	@echo "  $(YELLOW)build$(NC)          - Build the binary"
	@echo "  $(YELLOW)clean$(NC)          - Remove build artifacts"
	@echo "  $(YELLOW)test$(NC)           - Run tests"
	@echo "  $(YELLOW)coverage$(NC)       - Run tests with coverage report"
	@echo "  $(YELLOW)run$(NC)            - Build and run the application"
	@echo "  $(YELLOW)run-dev$(NC)        - Run in development mode"
	@echo "  $(YELLOW)docker-build$(NC)   - Build Docker image"
	@echo "  $(YELLOW)docker-run$(NC)     - Run Docker container"
	@echo "  $(YELLOW)deps$(NC)           - Download dependencies"
	@echo "  $(YELLOW)fmt$(NC)            - Format code"
	@echo "  $(YELLOW)vet$(NC)            - Run go vet"
	@echo "  $(YELLOW)lint$(NC)           - Run linter"
	@echo "  $(YELLOW)openapi$(NC)        - Generate OpenAPI specification"
	@echo "  $(YELLOW)readme$(NC)         - Generate README from template"
	@echo "  $(YELLOW)dev-env$(NC)        - Start local development environment"
	@echo "  $(YELLOW)dev-env-stop$(NC)   - Stop local development environment"
	@echo "  $(YELLOW)install-tools$(NC)  - Install development tools"
	@echo ""
	@echo "$(GREEN)Release targets:$(NC)"
	@echo "  $(YELLOW)version$(NC)         - Show current version"
	@echo "  $(YELLOW)next-versions$(NC)   - Show next available versions"
	@echo "  $(YELLOW)release-init$(NC)    - Initialize first release (v0.1.0)"
	@echo "  $(YELLOW)release-dry-run$(NC) - Test release without publishing"
	@echo "  $(YELLOW)release-snapshot$(NC)- Create snapshot release"
	@echo "  $(YELLOW)release-patch$(NC)   - Create patch release (x.x.X)"
	@echo "  $(YELLOW)release-minor$(NC)   - Create minor release (x.X.0)"
	@echo "  $(YELLOW)release-major$(NC)   - Create major release (X.0.0)"
	@echo "  $(YELLOW)release-clean$(NC)   - Clean release artifacts"
	@echo ""
	@echo "$(GREEN)Template targets:$(NC)"
	@echo "  $(YELLOW)init-project$(NC)    - Initialize new project from this template"
	@echo "  $(YELLOW)help$(NC)           - Show this help message"