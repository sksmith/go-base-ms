# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial Go microservice template with PostgreSQL and Kafka support
- Confluent Kafka Go client integration with Schema Registry support
- Kubernetes-compatible health endpoints (liveness/readiness)  
- OpenAPI 3.0 specification
- Structured logging with slog and remote log level control
- GoReleaser configuration for automated releases
- Comprehensive unit tests
- Complete Makefile with development workflows
- Multi-stage Dockerfile
- Docker Compose setup with all services
- SASL authentication and SSL encryption support
- Version endpoint for build information

### Features
- PostgreSQL connection with connection pooling
- Kafka producer/consumer with delivery guarantees
- Avro serialization via Schema Registry
- Health checks for external dependencies
- Dynamic log level configuration via API
- Automated version management and tagging
- Multi-architecture Docker builds (amd64/arm64)
- GitHub Actions CI/CD pipeline