# Document Registry Service

A production-ready document management and version control system built with Go, following clean architecture principles with CQRS, hexagonal architecture, and DDD-inspired boundaries.

## 🎯 Features

### Core Functionality
- ✅ **Document Management** - Create, update, search documents with metadata and tags
- ✅ **Version Control** - Semantic versioning with workflow states (DRAFT → SUBMITTED → APPROVED → PUBLISHED)
- ✅ **Approval Workflow** - Multi-stage approval process with role-based stakeholder management
- ✅ **Content Storage** - Multi-backend storage (S3, filesystem, database) with content validation
- ✅ **Document Types** - Typed document system with optional capabilities (validation, approval, deprecation)
- ✅ **Search** - Full-text search with filters, tags, and pagination
- ✅ **Validation** - Extensible validation registry supporting multiple document types (YAML, JSON, etc.)
- ✅ **Deprecation** - Manual and auto-deprecation workflows with approval support
- ✅ **Health Checks** - Comprehensive health endpoints for ECS deployment (Postgres, S3, SNS)

### Architecture Highlights
- **CQRS** - Logical separation of commands and queries
- **Hexagonal Architecture** - Ports and adapters pattern with clean boundaries
- **Event-Driven** - Internal domain events for all state changes
- **Type-Safe Document System** - DocumentType value objects with capability declarations
- **Validation Registry** - Pluggable validators for different document types
- **Transaction Safety** - err2-based error handling with ACID transactions
- **Multi-Backend Storage** - Abstracted storage layer supporting S3, filesystem, and database
- **ECS-Ready Health Checks** - ALB target group, liveness, readiness, and startup probes

### Technical Stack
- **Language**: Go 1.21+
- **HTTP Framework**: Echo v4
- **Database**: PostgreSQL 16
- **Object Storage**: S3-compatible (AWS S3, LocalStack, MinIO)
- **Messaging**: SNS-compatible (AWS SNS, LocalStack)
- **Migrations**: golang-migrate
- **Error Handling**: github.com/lainio/err2
- **Config**: github.com/caarlos0/env/v11 with prefix support

---

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Running Locally](#-running-locally)
- [Health Checks](#-health-checks)
- [API Documentation](#-api-documentation)
- [Architecture](#-architecture)
- [Document Types](#-document-types)
- [Development](#-development)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Migrations](#-migrations)
- [Troubleshooting](#-troubleshooting)

---

## 🚀 Quick Start

**TL;DR** - Get running in 5 minutes:

```bash
# 1. Clone the repository
git clone https://github.com/bbridges_11/document-registry.git
cd document-registry

# 2. Start all infrastructure (PostgreSQL, LocalStack)
make docker-up

# 3. Create environment file
cp .env.example .env

# 4. Run migrations
make db-migrate

# 5. Start the application
make run

# 6. Verify it's running
curl http://localhost:8080/health
```

The API will be available at `http://localhost:8080`

See [QUICKSTART.md](QUICKSTART.md) for detailed step-by-step instructions.

---

## 📦 Prerequisites

### Required
- **Go 1.21+** - [Download](https://golang.org/doc/install)
- **Docker Desktop** - [Download](https://docs.docker.com/get-docker/)
- **Make** - Pre-installed on macOS/Linux, [Windows instructions](https://gnuwin32.sourceforge.net/packages/make.htm)

### Optional
- **golang-migrate** - For manual database migrations ([Install](https://github.com/golang-migrate/migrate))
- **AWS CLI** - For production S3/SNS operations ([Install](https://aws.amazon.com/cli/))

---

## 🔧 Configuration

### Environment Variables

All environment variables use the `DOCUMENT_REGISTRY_` prefix for namespace isolation:

```bash
# Application
DOCUMENT_REGISTRY_APP_NAME=document-registry
DOCUMENT_REGISTRY_APP_VERSION=1.0.0
DOCUMENT_REGISTRY_PROFILE=local  # local, cloud-dev, cloud-prod

# Logging
DOCUMENT_REGISTRY_LOG_LEVEL=info  # debug, info, warn, error
DOCUMENT_REGISTRY_LOG_FORMAT=json  # json, console

# Server
DOCUMENT_REGISTRY_SERVER_DRIVER=http  # http, grpc
DOCUMENT_REGISTRY_HTTP_HOST=0.0.0.0
DOCUMENT_REGISTRY_HTTP_PORT=8080

# Database
DOCUMENT_REGISTRY_DB_DRIVER=postgres  # postgres, dynamo
DOCUMENT_REGISTRY_POSTGRES_HOST=localhost
DOCUMENT_REGISTRY_POSTGRES_PORT=5432
DOCUMENT_REGISTRY_POSTGRES_DATABASE=document_registry
DOCUMENT_REGISTRY_POSTGRES_USER=postgres
DOCUMENT_REGISTRY_POSTGRES_PASSWORD=postgres
DOCUMENT_REGISTRY_POSTGRES_SSL_MODE=disable
DOCUMENT_REGISTRY_POSTGRES_MAX_CONNS=25

# AWS - Region can use either standard AWS SDK variable OR prefixed version
AWS_REGION=us-east-1
# OR
DOCUMENT_REGISTRY_AWS_REGION=us-east-1

# S3
DOCUMENT_REGISTRY_S3_BUCKET=document-registry
DOCUMENT_REGISTRY_S3_ENDPOINT=http://localhost:4566  # For LocalStack

# SNS
DOCUMENT_REGISTRY_SNS_TOPIC_ARN=arn:aws:sns:us-east-1:000000000000:document-registry
DOCUMENT_REGISTRY_SNS_ENABLED=false
DOCUMENT_REGISTRY_SNS_ENDPOINT=http://localhost:4566  # For LocalStack
```

### Configuration Profiles

**local** - Development with LocalStack
- PostgreSQL with password auth
- LocalStack for S3/SNS
- Console logging

**cloud-dev** - AWS development environment
- RDS with IAM auth
- AWS S3/SNS
- JSON logging

**cloud-prod** - Production environment
- RDS with IAM auth
- AWS S3/SNS with encryption
- JSON logging with sampling

See [ENV_MIGRATION.md](ENV_MIGRATION.md) for migration guide from old variable names.

---

## 🏃 Running Locally

### Quick Start (All-in-One)

```bash
make dev-full
```

This runs: docker-up → db-migrate → s3-create-bucket → sns-create-topic

### Step-by-Step

```bash
# 1. Start infrastructure
make docker-up

# 2. Run migrations (consolidated schema)
make db-migrate

# 3. Create S3 bucket
make s3-create-bucket

# 4. Create SNS topic (optional - only if using notifications)
make sns-create-topic

# 5. Start the application
make run
```

### Verify Health

```bash
# Check all services
make health

# Individual health checks
curl http://localhost:8080/health          # ALB health (Postgres + S3)
curl http://localhost:8080/health/live     # Liveness probe
curl http://localhost:8080/health/ready    # Readiness (Postgres + S3 + SNS)
curl http://localhost:8080/health/startup  # Startup probe
```

---

## 🏥 Health Checks

The service provides four health check endpoints optimized for ECS deployment:

### `/health` - ALB Target Group Health Check
Checks critical dependencies required for operation:
- **Postgres** - Database connectivity
- **S3** - Object storage accessibility

**Use:** ALB target group health checks
**Timeout:** 5 seconds
**Returns:** 200 if healthy, 503 if unhealthy

### `/health/live` - Container Liveness Probe
Indicates the container process is running.

**Use:** ECS container `healthCheck` or Docker HEALTHCHECK
**Timeout:** Instant (<100ms)
**Returns:** Always 200

### `/health/ready` - Deep Readiness Check
Checks all dependencies including optional services:
- **Postgres** - Database connectivity
- **S3** - Object storage accessibility
- **SNS** - Notification service (if enabled)

**Use:** Monitoring dashboards, debugging
**Timeout:** 5 seconds
**Returns:** 200 if ready, 503 if not ready

### `/health/startup` - Startup Health Check
Same as readiness but with longer timeout for cold starts.

**Use:** ECS startup checks
**Timeout:** 10 seconds
**Returns:** 200 if started, 503 if starting

### ECS Task Definition Example

```json
{
  "healthCheck": {
    "command": ["CMD-SHELL", "curl -f http://localhost:8080/health/live || exit 1"],
    "interval": 30,
    "timeout": 5,
    "retries": 3,
    "startPeriod": 60
  }
}
```

### ALB Target Group Configuration

```
Protocol: HTTP
Path: /health
Port: traffic-port
Healthy threshold: 2
Unhealthy threshold: 2
Timeout: 5 seconds
Interval: 30 seconds
```

---

## 📚 API Documentation

### Core Endpoints

#### Documents
- `POST /documents` - Create document with first version
- `GET /documents` - List all documents (paginated)
- `GET /documents/:id` - Get document details
- `PUT /documents/:id` - Update document metadata
- `GET /documents/search` - Search documents with filters

#### Versions
- `POST /documents/:documentId/versions` - Create new version
- `GET /documents/:documentId/versions` - List document versions
- `GET /versions/:id` - Get version details
- `POST /versions/:id/submit` - Submit for review
- `POST /versions/:id/approve` - Approve version (requires approval role)
- `POST /versions/:id/reject` - Reject version
- `POST /versions/:id/publish` - Publish approved version

#### Stakeholders
- `POST /documents/:documentId/stakeholders` - Add stakeholder
- `GET /documents/:documentId/stakeholders` - List stakeholders
- `DELETE /stakeholders/:id` - Remove stakeholder

#### Approvals
- `POST /versions/:versionId/approvals` - Submit approval
- `GET /versions/:versionId/approvals` - List approvals

#### Deprecation
- `POST /versions/:id/deprecate` - Request deprecation
- `POST /deprecations/:id/approve` - Approve deprecation
- `POST /deprecations/:id/reject` - Reject deprecation

#### Validation
- `POST /validation/content` - Validate document content

#### Health
- `GET /health` - ALB health check
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness check
- `GET /health/startup` - Startup check

### Example: Create Pattern Document

```bash
# Encode YAML content as base64
CONTENT=$(cat pattern.yaml | base64)

# Create document
curl -X POST http://localhost:8080/documents \
  -H "X-User-ID: user-123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Gateway Pattern",
    "description": "Standard API gateway pattern",
    "document_type": "pattern",
    "tags": ["api", "gateway", "microservices"],
    "version": "1.0.0",
    "content": "'$CONTENT'"
  }'
```

---

## 🏛️ Architecture

### Project Structure

```
document-registry/
├── cmd/
│   └── api/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── application/             # Use cases and application logic
│   │   ├── document/           # Document commands/queries
│   │   ├── version/            # Version commands/queries
│   │   ├── approval/           # Approval workflows
│   │   ├── deprecation/        # Deprecation workflows
│   │   └── validation/         # Content validation
│   ├── domain/                  # Domain entities and business logic
│   │   ├── document/           # Document aggregate + DocumentType
│   │   ├── version/            # Version entity
│   │   ├── workflow/           # Workflow state machines
│   │   ├── approval/           # Approval policies
│   │   └── shared/             # Shared value objects
│   ├── ports/
│   │   └── outbound/           # Outbound dependency interfaces
│   ├── adapters/
│   │   ├── inbound/
│   │   │   └── http/           # HTTP handlers (Echo)
│   │   └── outbound/
│   │       ├── persistence/    # Repositories (Postgres)
│   │       ├── storage/        # Storage backends (S3, filesystem)
│   │       ├── notification/   # SNS adapter
│   │       └── validation/     # Validation registry + validators
│   ├── platform/                # Technical infrastructure
│   │   ├── config/             # Configuration loading
│   │   ├── database/           # Database connections
│   │   ├── aws/                # AWS client factories
│   │   ├── events/             # Event bus
│   │   └── server/             # HTTP server setup
│   ├── bootstrap/               # Dependency wiring
│   └── events/                  # Event contracts
├── pkg/
│   └── errors/                  # Shared error codes
├── migrations/                  # Historical migrations
├── migrations_consolidated/     # Consolidated initial schema
└── docs/                        # Additional documentation
```

### Key Design Patterns

**Hexagonal Architecture:**
- Application layer defines use cases
- Ports define dependency interfaces
- Adapters implement infrastructure concerns

**CQRS:**
- Commands modify state
- Queries read state
- Explicit input/output models

**Event-Driven:**
- Domain events for state changes
- Internal event bus
- Event handlers for side effects

**Type-Safe Document System:**
- DocumentType value objects
- Capability-based features (validation, approval, deprecation)
- Registry pattern for extensibility

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed design decisions.

---

## 📄 Document Types

The service supports a typed document system where each document type declares its capabilities:

### Supported Document Types

#### Pattern (`"pattern"`)
- **Workflow**: Approval-based (Draft → Submitted → Under Review → Approved → Published)
- **Approval**: Required (multi-stakeholder approval)
- **Validation**: YAML schema validation
- **Deprecation**: Supported (manual requires approval, auto doesn't)

### Adding New Document Types

See the [feature walkthrough](MIGRATIONS.md#adding-new-document-types) for step-by-step instructions on adding new types like:
- **Policy** - Auto-publish with JSON validation
- **Contract** - Strict approval with legal review
- **Template** - No approval, no deprecation

---

## 🛠️ Development

### Prerequisites

```bash
# Install development tools
make dev-setup

# Install linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Available Commands

```bash
make help                 # Show all commands

# Development
make run                  # Run application
make build                # Build binary
make test                 # Run tests
make test-coverage        # Generate coverage report
make lint                 # Run linter
make fmt                  # Format code
make tidy                 # Tidy dependencies

# Docker Infrastructure
make docker-up            # Start all services
make docker-down          # Stop all services
make docker-clean         # Remove containers and volumes
make docker-postgres      # Start only Postgres
make docker-localstack    # Start only LocalStack

# Database
make db-migrate           # Run migrations (consolidated)
make db-migrate-down      # Rollback one migration
make db-reset             # Drop and recreate database
make db-shell             # Open psql shell

# AWS/LocalStack
make s3-create-bucket     # Create S3 bucket
make s3-list-buckets      # List S3 buckets
make s3-list-objects      # List objects in bucket
make sns-create-topic     # Create SNS topic
make sns-list-topics      # List SNS topics
make sns-subscribe        # Subscribe email to topic
make sns-list-subscriptions # List SNS subscriptions

# Health & Monitoring
make health               # Check all services
make logs                 # View application logs
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage
open coverage.html

# Specific package
go test -v ./internal/domain/document/...

# Integration tests
go test -v -tags=integration ./...
```

---

## 📦 Deployment

### Building for Production

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o bin/document-registry-linux-amd64 \
  cmd/api/main.go
```

### Docker Image

```bash
# Build image
docker build -t document-registry:latest .

# Run container
docker run -p 8080:8080 \
  --env-file .env.prod \
  document-registry:latest
```

### ECS Deployment

1. **Task Definition** - Set health check to `/health/live`
2. **Target Group** - Set health check to `/health`
3. **Environment Variables** - Use `DOCUMENT_REGISTRY_` prefix
4. **IAM Roles** - Grant S3, SNS, and RDS permissions

Example task definition snippet:

```json
{
  "containerDefinitions": [{
    "name": "document-registry",
    "image": "123456789012.dkr.ecr.us-east-1.amazonaws.com/document-registry:latest",
    "portMappings": [{
      "containerPort": 8080,
      "protocol": "tcp"
    }],
    "environment": [
      {"name": "DOCUMENT_REGISTRY_PROFILE", "value": "cloud-prod"},
      {"name": "DOCUMENT_REGISTRY_HTTP_PORT", "value": "8080"},
      {"name": "AWS_REGION", "value": "us-east-1"}
    ],
    "healthCheck": {
      "command": ["CMD-SHELL", "curl -f http://localhost:8080/health/live || exit 1"],
      "interval": 30,
      "timeout": 5,
      "retries": 3,
      "startPeriod": 60
    }
  }]
}
```

---

## 🗄️ Migrations

### Using Consolidated Migrations (New Deployments)

For fresh databases, use the consolidated schema:

```bash
# Apply consolidated migration
migrate -path migrations_consolidated \
        -database "postgres://user:pass@host:5432/dbname?sslmode=disable" \
        up
```

### Using Historical Migrations (Existing Databases)

For databases with existing data:

```bash
# Apply incremental migrations
migrate -path migrations \
        -database "postgres://user:pass@host:5432/dbname?sslmode=disable" \
        up
```

See [MIGRATIONS.md](MIGRATIONS.md) for complete migration documentation.

---

## 🐛 Troubleshooting

### Common Issues

**Port already in use**
```bash
lsof -i :8080
# Or use different port
DOCUMENT_REGISTRY_HTTP_PORT=8081 make run
```

**Database connection failed**
```bash
# Check Postgres is running
make docker-postgres-logs

# Restart Postgres
make docker-down
make docker-up
```

**S3 connection failed**
```bash
# Check LocalStack
make docker-localstack-logs

# Recreate bucket
make s3-create-bucket
```

**Migration failed**
```bash
# Check migration status
make db-status

# Reset and retry
make db-reset
make db-migrate
```

### Health Check Debugging

```bash
# Check specific service health
curl http://localhost:8080/health/ready

# View detailed health status
curl http://localhost:8080/health/ready | jq .

# Check logs
make logs
```

---

## 📖 Additional Documentation

- [QUICKSTART.md](QUICKSTART.md) - 5-minute getting started guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - Detailed architecture decisions
- [MIGRATIONS.md](MIGRATIONS.md) - Database migration guide
- [ENV_MIGRATION.md](ENV_MIGRATION.md) - Environment variable migration
- [API.md](docs/API.md) - Complete API reference

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

## 🙏 Acknowledgments

- Built with [Echo](https://echo.labstack.com/) HTTP framework
- Error handling with [err2](https://github.com/lainio/err2)
- Configuration with [env](https://github.com/caarlos0/env)
- Migrations with [golang-migrate](https://github.com/golang-migrate/migrate)