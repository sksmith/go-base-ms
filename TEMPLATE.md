# 🚀 Go Microservice Template

This is a comprehensive Go microservice template with customizable dependencies and automated setup.

## ✨ Features

- **Configurable Dependencies**: Choose PostgreSQL, Kafka, and Schema Registry support
- **Automated Setup**: One-command project initialization
- **GitHub Integration**: Automatic repository creation and push
- **GoReleaser**: Built-in release management
- **Production Ready**: Health checks, logging, metrics, and Docker support
- **Clean Architecture**: Well-organized project structure

## 🎯 Quick Start

### 1. Clone the Template
```bash
git clone https://github.com/your-username/go-base-ms.git my-new-service
cd my-new-service
```

### 2. Run the Initialization Script
```bash
./scripts/init-project.sh
```

The script will guide you through:
- **Project Configuration**: Name, description, GitHub settings
- **Dependency Selection**: PostgreSQL, Kafka, Schema Registry
- **GitHub Repository**: Automatic creation and push (optional)

### 3. Start Development
```bash
# Start development environment
make dev-env

# Run the application
make run-dev

# Run tests
make test
```

## 🛠️ What Gets Configured

### Project Structure
```
your-project/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── api/            # HTTP handlers and routing
│   ├── config/         # Configuration management
│   ├── db/             # Database layer (optional)
│   ├── kafka/          # Kafka client (optional)
│   ├── health/         # Health checks
│   ├── logger/         # Structured logging
│   └── version/        # Version information
├── pkg/                # Public packages
├── k8s/                # Kubernetes manifests
├── scripts/            # Utility scripts
└── docker-compose.yml  # Development environment
```

### Dependency Options

#### PostgreSQL Support
- ✅ Connection pooling and health checks
- ✅ Configuration via environment variables
- ✅ Docker Compose integration
- ❌ Removed if not selected

#### Kafka Support
- ✅ Confluent's official Go client
- ✅ Producer/Consumer with delivery guarantees
- ✅ SASL authentication and SSL support
- ✅ Docker Compose with Zookeeper
- ❌ Removed if not selected

#### Schema Registry (Avro)
- ✅ Automatic schema evolution
- ✅ Serialization/deserialization
- ✅ API key and basic auth support
- ❌ Simplified Kafka client if not selected

### GitHub Integration
- 🔄 Automatic repository creation
- 🔄 Initial commit and push
- 🔄 Public/private repository options
- 🔄 GitHub Actions CI/CD setup

### Dynamic Documentation
- 📝 Template-based README generation
- 📝 Conditional sections based on selected features
- 📝 Automatic project name and description replacement
- 📝 Clean removal of unused dependency documentation

## 📋 Initialization Process

The initialization script performs:

1. **Prerequisites Check**: Go version, git, required tools
2. **User Input Collection**: Project details and dependencies
3. **Module Renaming**: Updates all imports and references
4. **Dependency Removal**: Removes unwanted code and dependencies
5. **File Updates**: Configs, Docker files, Makefiles
6. **README Generation**: Creates project-specific documentation
7. **Git Initialization**: New repository with clean history
8. **GitHub Creation**: Optional repository creation and push
9. **Cleanup**: Removes template-specific files

## 🔧 Configuration Examples

### Basic Web Service (No External Dependencies)
```bash
# During initialization:
Include PostgreSQL? (Y/n): n
Include Kafka? (Y/n): n
```
Result: Minimal HTTP service with health checks and logging

### Database Service
```bash
# During initialization:
Include PostgreSQL? (Y/n): y
Include Kafka? (Y/n): n
```
Result: HTTP service + PostgreSQL integration

### Event-Driven Service
```bash
# During initialization:
Include PostgreSQL? (Y/n): y
Include Kafka? (Y/n): y
Include Schema Registry? (Y/n): y
```
Result: Full-featured microservice with database and messaging

## 🎛️ Environment Variables

The generated service supports:

### Common
- `PORT` - HTTP server port (default: 8080)
- `LOG_LEVEL` - Log level (debug/info/warn/error)

### PostgreSQL (if enabled)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `DB_SSLMODE` - SSL mode

### Kafka (if enabled)
- `KAFKA_BROKERS` - Comma-separated broker list
- `KAFKA_TOPIC`, `KAFKA_GROUP_ID` - Topic and consumer group
- `KAFKA_SECURITY_PROTOCOL`, `KAFKA_SASL_MECHANISM` - Security settings
- `KAFKA_SASL_USERNAME`, `KAFKA_SASL_PASSWORD` - Authentication

### Schema Registry (if enabled)
- `SCHEMA_REGISTRY_URL` - Registry endpoint
- `SCHEMA_REGISTRY_USERNAME`, `SCHEMA_REGISTRY_PASSWORD` - Basic auth
- `SCHEMA_REGISTRY_API_KEY`, `SCHEMA_REGISTRY_API_SECRET` - API auth

## 🚢 Deployment Options

### Local Development
```bash
make dev-env     # Start dependencies
make run-dev     # Run with development config
```

### Docker
```bash
make docker-build
make docker-run
```

### Docker Compose
```bash
docker-compose up
```

### Kubernetes
```bash
kubectl apply -f k8s/
```

## 📦 Release Management

Built-in GoReleaser configuration:

```bash
# Initialize versioning
make release-init

# Create releases
make release-patch    # v1.0.0 → v1.0.1
make release-minor    # v1.0.0 → v1.1.0
make release-major    # v1.0.0 → v2.0.0
```

## 🧪 Generated Tests

The template includes comprehensive tests:
- Unit tests for all packages
- Integration tests for HTTP endpoints
- Mock implementations for external dependencies
- Test coverage reporting

## 📚 Documentation

Generated projects include:
- **README.md** - Dynamically generated project-specific documentation
- **RELEASES.md** - Release management guide
- **CHANGELOG.md** - Version history
- **OpenAPI Spec** - API documentation at `/openapi.yaml` and `/openapi.json`

### Dynamic README Generation

The template uses a sophisticated README generation system:

- **Template Location**: `templates/README.md.template`
- **Generator Script**: `scripts/generate-readme.sh`
- **Conditional Sections**: Uses `{{#FEATURE}}...{{/FEATURE}}` syntax
- **Variable Replacement**: Replaces `PROJECT_NAME` and `PROJECT_DESCRIPTION`

**Supported Conditionals:**
- `{{#USE_POSTGRES}}...{{/USE_POSTGRES}}` - PostgreSQL documentation
- `{{#USE_KAFKA}}...{{/USE_KAFKA}}` - Kafka integration documentation  
- `{{#USE_SCHEMA_REGISTRY}}...{{/USE_SCHEMA_REGISTRY}}` - Schema Registry documentation

**Manual README Generation:**
```bash
# Generate README with current template defaults
make readme

# Generate README with custom settings
PROJECT_NAME="my-service" USE_KAFKA=false ./scripts/generate-readme.sh
```