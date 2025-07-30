# 🚀 Go Microservice Template

A comprehensive Go microservice template with customizable dependencies and automated project initialization.

## ✨ Features

- **Smart Dependency Selection**: Choose PostgreSQL, Kafka, and Schema Registry support
- **Automated Project Setup**: One-command initialization with custom naming
- **GitHub Integration**: Automatic repository creation and initial commit
- **Production Ready**: Health checks, structured logging, OpenAPI docs, and Docker support
- **Clean Architecture**: Well-organized project structure with comprehensive tests

## 🎯 Quick Start

### 1. Clone and Initialize

```bash
git clone https://github.com/sksmith/go-base-ms.git my-new-service
cd my-new-service
./scripts/init-project.sh
```

The initialization script will:
- Collect project details (name, description, GitHub settings)
- Let you choose dependencies (PostgreSQL, Kafka, Schema Registry)
- Remove unwanted code and dependencies
- Generate project-specific documentation
- Optionally create and push to a GitHub repository

### 2. Start Development

After initialization, your new project will be ready:

```bash
# Start development environment
make dev-env

# Run the application
make run-dev

# Run tests
make test

# Create your first release
make release-init
```

## 🛠️ What You Get

### Core Features (Always Included)
- HTTP server with graceful shutdown
- Structured logging with `slog` and dynamic log level control
- Health check endpoints (`/health/live`, `/health/ready`)
- Version information endpoint (`/version`)
- OpenAPI 3.0 specification (modular YAML structure)
- Comprehensive test suite with mocks
- Docker support with multi-stage builds
- Kubernetes manifests
- GoReleaser configuration for automated releases
- Makefile with common development tasks

### Optional Dependencies
- **PostgreSQL**: Connection pooling, health checks, Docker Compose integration
- **Kafka**: Confluent Go client with SASL/SSL support, producer/consumer patterns
- **Schema Registry**: Avro serialization with automatic schema evolution

### Generated Project Structure
```
your-project/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── api/            # HTTP handlers and routing
│   ├── config/         # Configuration management
│   ├── db/             # Database layer (if PostgreSQL selected)
│   ├── kafka/          # Kafka client (if Kafka selected)
│   ├── health/         # Health checks
│   ├── logger/         # Structured logging
│   └── version/        # Version information
├── api/openapi/         # OpenAPI specification sources
├── k8s/                # Kubernetes manifests
├── docker-compose.yml  # Development environment
├── Dockerfile          # Production container
├── Makefile           # Development tasks
└── README.md          # Project-specific documentation
```

## 🎛️ Configuration Examples

### Minimal Service (No External Dependencies)
```
Include PostgreSQL? (Y/n): n
Include Kafka? (Y/n): n
```
Result: HTTP service with health checks, logging, and OpenAPI docs

### Database Service
```
Include PostgreSQL? (Y/n): y
Include Kafka? (Y/n): n
```
Result: HTTP service + PostgreSQL integration with connection pooling

### Full Event-Driven Service
```
Include PostgreSQL? (Y/n): y
Include Kafka? (Y/n): y
Include Schema Registry? (Y/n): y
```
Result: Complete microservice with database, messaging, and Avro serialization

## 📋 Initialization Process

The `./scripts/init-project.sh` script performs:

1. **Prerequisites Check**: Validates Go version and required tools
2. **Project Configuration**: Collects name, description, and GitHub settings
3. **Dependency Selection**: Choose PostgreSQL, Kafka, and Schema Registry
4. **Code Customization**: Updates imports, module names, and all references
5. **Dependency Cleanup**: Removes unwanted code and dependencies cleanly
6. **Documentation Generation**: Creates project-specific README with selected features
7. **Git Setup**: Initializes clean repository with initial commit
8. **GitHub Integration**: Optionally creates repository and pushes code
9. **Finalization**: Removes template-specific files and validates setup

## 🚢 Deployment Ready

Generated projects include:
- **Docker**: Multi-stage builds with security best practices
- **Docker Compose**: Full development environment with dependencies
- **Kubernetes**: Production-ready manifests with health checks
- **GoReleaser**: Automated releases with version management
- **CI/CD**: GitHub Actions workflow templates

## 📚 Documentation

Each generated project includes:
- **Comprehensive README**: Tailored to selected features and dependencies
- **OpenAPI Specification**: Interactive API documentation
- **Development Guide**: Local setup, testing, and deployment instructions
- **Environment Variables**: Complete configuration reference

## 🤝 Contributing

1. Fork this template repository
2. Make your improvements
3. Test with `./scripts/init-project.sh`
4. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.