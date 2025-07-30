# go-base-ms

A minimal Go microservice template with PostgreSQL, Kafka, and Kubernetes health endpoints

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/                 # HTTP handlers and routing
│   ├── config/              # Configuration management
│   ├── db/                  # Database connection and operations
│   ├── health/              # Health check implementation
│   ├── kafka/               # Kafka client implementation
│   ├── logger/              # Structured logging setup
│   └── version/             # Version information
├── api/
│   ├── openapi.yaml         # Generated OpenAPI specification
│   ├── openapi.json         # Generated OpenAPI specification (JSON)
│   └── openapi/             # OpenAPI specification sources
│       ├── base.yaml        # Base specification with schemas and info
│       ├── standard.yaml    # Standard endpoints (health, version, admin)
│       └── application.yaml # Application-specific endpoints
├── k8s/                     # Kubernetes manifests
├── test/                    # Additional test files
├── Dockerfile               # Multi-stage Docker build
├── Makefile                 # Build and development tasks
├── go.mod                   # Go module definition
└── README.md                # This file
```

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker (optional)

- PostgreSQL (for local development)


- Kafka (for local development)

- Schema Registry (for Avro serialization)
### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-base-ms
```

2. Install dependencies:
```bash
make deps
```

3. Build the application:
```bash
make build
```

## Running the Service

### Local Development

Start the required services:
```bash
make dev-env
```

Run the application:
```bash
make run-dev
```

The service will be available at `http://localhost:8080`.

### Using Docker

Build and run with Docker:
```bash
make docker-build
make docker-run
```

### Using Docker Compose

Run the entire stack:
```bash
docker-compose up
```

This starts:
- The application on port 8080

- PostgreSQL on port 5432


- Kafka on port 9092
- Zookeeper on port 2181

- Schema Registry on port 8081
## Available Endpoints

### Health Checks
- `GET /health/live` - Liveness probe (always returns 200)
- `GET /health/ready` - Readiness probe (checks dependencies)

### Information
- `GET /version` - Get build version information
- `GET /api/v1/admin/log-level` - Get current log level
- `PUT /api/v1/admin/log-level` - Change log level dynamically

### API Examples
- `GET /api/v1/hello` - Simple hello endpoint
- `POST /api/v1/echo` - Echo request body

### Documentation
- `GET /openapi.yaml` - OpenAPI 3.0 specification (YAML)
- `GET /openapi.json` - OpenAPI 3.0 specification (JSON)

## Usage Examples

### Changing Log Level Remotely

Get current log level:
```bash
curl http://localhost:8080/api/v1/admin/log-level
```

Change to debug level:
```bash
curl -X PUT http://localhost:8080/api/v1/admin/log-level \
  -H "Content-Type: application/json" \
  -d '{"level": "debug"}'
```

### Basic API Usage

Hello endpoint:
```bash
curl http://localhost:8080/api/v1/hello
```

Echo endpoint:
```bash
curl -X POST http://localhost:8080/api/v1/echo \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello World", "timestamp": "2024-01-01T00:00:00Z"}'
```


## Database

The service connects to PostgreSQL using the following environment variables:

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (default: gobase)
- `DB_SSLMODE` - SSL mode (default: disable)
- `DB_MAX_OPEN_CONNS` - Maximum open connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Maximum idle connections (default: 5)
- `DB_CONN_MAX_LIFETIME` - Connection lifetime in minutes (default: 5)

### Database Usage Example

The database connection is available through the health check and can be extended for your application needs.
## Kafka Integration

The service includes Kafka integration using Confluent's official Go client.

### Configuration

Kafka settings via environment variables:

- `KAFKA_BROKERS` - Comma-separated broker list (default: localhost:9092)
- `KAFKA_TOPIC` - Default topic name (default: events)
- `KAFKA_GROUP_ID` - Consumer group ID (default: go-base-ms)
- `KAFKA_SECURITY_PROTOCOL` - Security protocol (default: PLAINTEXT)
- `KAFKA_SASL_MECHANISM` - SASL mechanism (PLAIN, SCRAM-SHA-256, etc.)
- `KAFKA_SASL_USERNAME` - SASL username
- `KAFKA_SASL_PASSWORD` - SASL password


### Schema Registry Configuration

- `SCHEMA_REGISTRY_URL` - Registry endpoint (default: http://localhost:8081)
- `SCHEMA_REGISTRY_USERNAME` - Basic auth username
- `SCHEMA_REGISTRY_PASSWORD` - Basic auth password
- `SCHEMA_REGISTRY_API_KEY` - API key authentication
- `SCHEMA_REGISTRY_API_SECRET` - API secret authentication

### Avro Support

The service includes full Avro serialization support:
- Automatic schema evolution
- Backward/forward compatibility
- Built-in serializer/deserializer
- Schema Registry integration


### Features

The service automatically connects to:
- **Kafka brokers** using Confluent's Go client with support for:
  - SASL authentication (PLAIN, SCRAM-SHA-256, etc.)
  - SSL/TLS encryption
  - Producer with idempotence and delivery guarantees
  - Consumer with automatic offset management

- **Schema Registry** for Avro serialization:
  - Automatic schema evolution
  - Backward/forward compatibility
  - Built-in serializer/deserializer
## API Documentation

The service provides a comprehensive OpenAPI 3.0 specification using a modular approach:

### OpenAPI Structure

- **`api/openapi/base.yaml`** - Base specification with schemas, info, and common components
- **`api/openapi/standard.yaml`** - Standard endpoints (health checks, version, admin)
- **`api/openapi/application.yaml`** - Application-specific business endpoints
- **`api/openapi.yaml`** - Generated merged specification (YAML)
- **`api/openapi.json`** - Generated merged specification (JSON)

### Available Endpoints

- `GET /openapi.yaml` - OpenAPI 3.0 specification (YAML format)
- `GET /openapi.json` - OpenAPI 3.0 specification (JSON format)

### Modifying the API Specification

1. Edit the source files in `api/openapi/`:
   - Add new schemas to `base.yaml`
   - Add standard endpoints to `standard.yaml`
   - Add business endpoints to `application.yaml`

2. Regenerate the merged specification:
```bash
make openapi
```

3. The merged files are automatically included in Docker builds and releases.

## Development

### Running Tests
```bash
make test
```

### Running Tests with Coverage
```bash
make coverage
```

### Code Formatting
```bash
make fmt
```

### Linting
```bash
make lint
```

### Generate OpenAPI Spec
```bash
make openapi
```

## Environment Variables

### Application Settings
- `PORT` - HTTP server port (default: 8080)
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: info)


### Database Settings
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (default: gobase)
- `DB_SSLMODE` - SSL mode (default: disable)
- `DB_MAX_OPEN_CONNS` - Maximum open connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Maximum idle connections (default: 5)
- `DB_CONN_MAX_LIFETIME` - Connection lifetime in minutes (default: 5)
### Kafka Settings
- `KAFKA_BROKERS` - Comma-separated broker list (default: localhost:9092)
- `KAFKA_TOPIC` - Default topic name (default: events)
- `KAFKA_GROUP_ID` - Consumer group ID (default: go-base-ms)
- `KAFKA_SECURITY_PROTOCOL` - Security protocol (default: PLAINTEXT)
- `KAFKA_SASL_MECHANISM` - SASL mechanism
- `KAFKA_SASL_USERNAME` - SASL username
- `KAFKA_SASL_PASSWORD` - SASL password


### Schema Registry Settings
- `SCHEMA_REGISTRY_URL` - Registry endpoint (default: http://localhost:8081)
- `SCHEMA_REGISTRY_USERNAME` - Basic auth username
- `SCHEMA_REGISTRY_PASSWORD` - Basic auth password
- `SCHEMA_REGISTRY_API_KEY` - API key
- `SCHEMA_REGISTRY_API_SECRET` - API secret
## Deployment

### Kubernetes

Apply the Kubernetes manifests:
```bash
kubectl apply -f k8s/
```

### Docker

The service includes optimized Docker builds:

**Development build:**
```bash
make docker-build
```

**Production build (via GoReleaser):**
```bash
make release-snapshot
```

## Release Management

This project uses [GoReleaser](https://goreleaser.com/) for automated releases:

```bash
# Initialize first release
make release-init

# Create patch release
make release-patch

# Create minor release  
make release-minor

# Create major release
make release-major
```

## Monitoring and Observability

### Health Checks

The service provides Kubernetes-compatible health endpoints:

- **Liveness**: `/health/live` - Always returns 200 if service is running
- **Readiness**: `/health/ready` - Returns 200 only if all dependencies are healthy

### Logging

Structured logging using Go's `slog` package:
- JSON format in production
- Configurable log levels
- Dynamic log level changes via API

### Metrics

The service is designed to be easily extended with metrics collection:
- Ready for Prometheus integration
- Health check status monitoring
- Custom application metrics

## Configuration

All configuration is done via environment variables with sensible defaults for local development.

Create a `.env` file for local development:

```bash
# Application
PORT=8080
LOG_LEVEL=debug


# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go-base-ms
DB_SSLMODE=disable
# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=events
KAFKA_GROUP_ID=go-base-ms
KAFKA_SECURITY_PROTOCOL=PLAINTEXT


# Schema Registry
SCHEMA_REGISTRY_URL=http://localhost:8081
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run the full test suite: `make test`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
