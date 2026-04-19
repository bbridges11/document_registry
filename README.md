# Document Registry Service

A production-ready document management and version control system built with Go, following clean architecture principles with CQRS, hexagonal architecture, and DDD-inspired boundaries.

## 🎯 Features

### Core Functionality
- ✅ **Document Management** - Create, update, search documents with metadata
- ✅ **Version Control** - Semantic versioning with workflow states (DRAFT → SUBMITTED → APPROVED → PUBLISHED)
- ✅ **Approval Workflow** - Multi-stage approval process with stakeholder management
- ✅ **Content Storage** - S3-compatible storage for document content
- ✅ **Authorization** - Fine-grained access control with OpenFGA
- ✅ **Search** - Full-text search with filters and pagination
- ✅ **Validation** - Custom validation rules per document type
- ✅ **Deprecation** - Manual and auto-deprecation workflows for versions

### Architecture Highlights
- **CQRS** - Logical separation of commands and queries
- **Hexagonal Architecture** - Ports and adapters pattern
- **Event-Driven** - Domain events for all state changes
- **Clean Boundaries** - Application → Domain → Ports → Adapters
- **Transaction Safety** - err2-based error handling with ACID transactions
- **Minimal main.go** - All logic in appropriate layers

### Technical Stack
- **Language**: Go 1.25+
- **HTTP Framework**: Echo v4
- **Database**: PostgreSQL 16
- **Object Storage**: S3-compatible (LocalStack for local dev)
- **Authorization**: OpenFGA
- **Migrations**: golang-migrate
- **Error Handling**: github.com/lainio/err2

---

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Running Locally](#-running-locally)
- [API Documentation](#-api-documentation)
- [Architecture](#-architecture)
- [Development](#-development)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Troubleshooting](#-troubleshooting)

---

## 🚀 Quick Start

**TL;DR** - Get running in 5 minutes:

```bash
# 1. Clone the repository
git clone https://github.com/bbridges_11/document-registry.git
cd document-registry

# 2. Start all infrastructure (PostgreSQL, OpenFGA, LocalStack)
make docker-up

# 3. Setup OpenFGA store and S3 bucket
make openfga-setup
make s3-create-bucket

# 4. Copy the OPENFGA_STORE_ID from step 3 output
# Create .env file:
cat > .env << EOF
# Server
HTTP_PORT=8080
ENV=local

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=document_registry

# OpenFGA (paste your store ID from step 3)
OPENFGA_ENABLED=false
OPENFGA_URL=http://localhost:8081
OPENFGA_STORE_ID=<paste-store-id-here>

# AWS/S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_S3_ENDPOINT=http://localhost:4566
AWS_S3_BUCKET=document-registry
AWS_S3_FORCE_PATH_STYLE=true
EOF

# 5. Run migrations
make db-migrate

# 6. Start the application
make run

# 7. Verify it's running
curl http://localhost:8080/health
```

The API will be available at `http://localhost:8080`

---

## 📦 Prerequisites

### Required

- **Go 1.25+** - [Install Go](https://golang.org/doc/install)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- **Make** - Usually pre-installed on macOS/Linux

### Optional (for development)

- **golangci-lint** - Code linting
  ```bash
  brew install golangci-lint  # macOS
  ```

- **golang-migrate** - Database migrations
  ```bash
  make migrate-install
  ```

- **AWS CLI** - For S3 operations
  ```bash
  brew install awscli  # macOS
  ```

- **jq** - JSON processing (for scripts)
  ```bash
  brew install jq  # macOS
  ```

- **Postman** - API testing (optional, collection provided)

---

## 📥 Installation

### 1. Clone the Repository

```bash
git clone https://github.com/bbridges_11/document-registry.git
cd document-registry
```

### 2. Install Go Dependencies

```bash
go mod download
go mod tidy
```

### 3. Install Development Tools (Optional)

```bash
# Database migration tool
make migrate-install

# Code linter
brew install golangci-lint

# AWS CLI for S3 operations
brew install awscli
```

---

## ⚙️ Configuration

### Environment Variables

The application uses environment variables for configuration. See `.env.example` for all options.

#### Required Variables

```bash
# Server Configuration
HTTP_PORT=8080
ENV=local  # local, dev, staging, prod

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=document_registry

# OpenFGA Authorization
OPENFGA_ENABLED=false  # true for production
OPENFGA_URL=http://localhost:8081
OPENFGA_STORE_ID=<your-store-id>

# AWS S3 Storage
AWS_REGION=us-east-1
AWS_S3_BUCKET=document-registry
AWS_S3_ENDPOINT=http://localhost:4566  # LocalStack
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_S3_FORCE_PATH_STYLE=true
```

#### Optional Variables

```bash
# Logging
LOG_LEVEL=info  # debug, info, warn, error
LOG_FORMAT=json  # json, console

# CORS
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_SECOND=100
```

### Configuration Profiles

The application supports multiple configuration profiles:

- **local** - Local development with bypassed auth
- **dev** - Development environment
- **staging** - Pre-production environment
- **prod** - Production environment

Set via `ENV` environment variable.

---

## 🏃 Running Locally

### Method 1: All-in-One (Recommended)

```bash
# Start everything
make dev-full

# Copy the OPENFGA_STORE_ID to .env
# Then start the app
make run
```

### Method 2: Step-by-Step

#### Step 1: Start Infrastructure

```bash
# Start PostgreSQL, OpenFGA, and LocalStack
make docker-up
```

This starts:
- **PostgreSQL** on port 5432
- **OpenFGA** on ports 8081 (API) and 3000 (Playground)
- **LocalStack** on port 4566

#### Step 2: Setup OpenFGA

```bash
make openfga-setup
```

Copy the `OPENFGA_STORE_ID` from the output and add it to your `.env` file.

#### Step 3: Create S3 Bucket

```bash
make s3-create-bucket
```

#### Step 4: Run Database Migrations

```bash
make db-migrate
```

This creates all tables:
- `documents` - Document metadata
- `versions` - Version history
- `stakeholders` - Document stakeholders
- `approvals` - Approval records
- `users` - User accounts
- `deprecations` - Deprecation tracking

#### Step 5: Start the Application

```bash
make run
```

Or build and run:

```bash
make build
./bin/document-registry
```

#### Step 6: Verify

```bash
# Health check
curl http://localhost:8080/health

# Should return: {"status":"healthy"}
```

---

## 🌐 API Documentation

### Postman Collection

A complete Postman collection is provided with all 48 endpoints:

```
postman/
├── document-registry.postman_collection.json  # Main collection (42 endpoints)
├── deprecation.postman_collection.json        # Deprecation feature (6 endpoints)
└── local.postman_environment.json             # Local environment
```

**Import to Postman**:
1. Open Postman
2. Click **Import**
3. Select the collection files
4. Import the environment file
5. Set your Bearer token in the environment

### API Endpoints (48 total)

#### Documents (8 endpoints)
- `POST /documents` - Create document
- `GET /documents` - List/search documents
- `GET /documents/:id` - Get document by ID
- `PUT /documents/:id` - Update document
- `DELETE /documents/:id` - Delete document
- `GET /documents/:id/versions` - List document versions
- `POST /documents/:id/versions` - Create new version
- `POST /documents/search` - Advanced search

#### Versions (12 endpoints)
- `GET /versions/:id` - Get version by ID
- `PUT /versions/:id` - Update version
- `DELETE /versions/:id` - Delete version
- `POST /versions/:id/submit` - Submit for approval
- `POST /versions/:id/approve` - Approve version
- `POST /versions/:id/reject` - Reject version
- `POST /versions/:id/publish` - Publish version
- `GET /versions/:id/content` - Get version content
- `PUT /versions/:id/content` - Upload version content
- `POST /versions/:id/validate` - Validate version
- `GET /versions/:id/approvals` - List approvals
- `POST /versions/:id/approvals` - Create approval

#### Stakeholders (5 endpoints)
- `GET /documents/:id/stakeholders` - List stakeholders
- `POST /documents/:id/stakeholders` - Add stakeholder
- `DELETE /documents/:id/stakeholders/:userID` - Remove stakeholder
- `PUT /documents/:id/stakeholders/:userID` - Update stakeholder role
- `GET /stakeholders/me` - My stakeholder relationships

#### Approvals (6 endpoints)
- `GET /approvals` - List all approvals
- `GET /approvals/:id` - Get approval details
- `POST /approvals/:id/approve` - Approve
- `POST /approvals/:id/reject` - Reject
- `GET /approvals/pending` - Pending approvals
- `GET /approvals/me` - My approval tasks

#### Deprecation (6 endpoints)
- `POST /versions/:id/deprecate` - Request deprecation
- `POST /versions/:id/deprecation/approve` - Approve deprecation
- `POST /versions/:id/deprecation/reject` - Reject deprecation
- `POST /versions/:id/deprecation/cancel` - Cancel deprecation
- `GET /versions/:id/deprecation` - Get deprecation details
- `GET /versions/:id/deprecation/status` - Get status with permissions

#### Users (8 endpoints)
- `POST /users` - Create user
- `GET /users` - List users
- `GET /users/:id` - Get user
- `PUT /users/:id` - Update user
- `DELETE /users/:id` - Delete user
- `POST /users/:id/roles` - Assign role
- `DELETE /users/:id/roles/:role` - Remove role
- `GET /users/me` - Get current user

#### Validation (2 endpoints)
- `POST /validation/rules` - Create validation rule
- `GET /validation/rules/:documentType` - Get rules for document type

#### Health (1 endpoint)
- `GET /health` - Service health check

### Authentication

All endpoints (except `/health`) require a Bearer token:

```bash
curl -H "Authorization: Bearer <your-token>" \
  http://localhost:8080/documents
```

For local development with `OPENFGA_ENABLED=false`, you can use any token:

```bash
curl -H "Authorization: Bearer test-token" \
  http://localhost:8080/documents
```

---

## 🏗️ Architecture

### Project Structure

```
document-registry/
├── cmd/
│   └── api/
│       └── main.go                    # Application entry point
├── internal/
│   ├── application/                   # Use cases and application logic
│   │   ├── approval/                  # Approval workflows
│   │   ├── deprecation/               # Deprecation workflows
│   │   ├── document/                  # Document operations
│   │   ├── stakeholder/               # Stakeholder management
│   │   ├── user/                      # User management
│   │   ├── validation/                # Validation rules
│   │   └── version/                   # Version control
│   ├── domain/                        # Business logic and entities
│   │   ├── approval/                  # Approval aggregate
│   │   ├── deprecation/               # Deprecation aggregate
│   │   ├── document/                  # Document aggregate
│   │   ├── shared/                    # Shared domain types
│   │   ├── stakeholder/               # Stakeholder value objects
│   │   ├── user/                      # User aggregate
│   │   ├── version/                   # Version aggregate
│   │   └── workflow/                  # Workflow states
│   ├── adapters/
│   │   ├── inbound/                   # HTTP handlers
│   │   │   └── http/
│   │   │       ├── approval/
│   │   │       ├── deprecation/
│   │   │       ├── document/
│   │   │       ├── health/
│   │   │       ├── stakeholder/
│   │   │       ├── user/
│   │   │       ├── validation/
│   │   │       └── version/
│   │   └── outbound/                  # Repository implementations
│   │       ├── authorization/
│   │       │   └── openfga/
│   │       ├── persistence/
│   │       │   └── postgres/
│   │       └── storage/
│   │           └── s3/
│   ├── ports/                         # Interface definitions
│   │   └── outbound/
│   ├── platform/                      # Infrastructure setup
│   │   ├── config/                    # Configuration
│   │   ├── events/                    # Event bus
│   │   ├── logger/                    # Logging
│   │   ├── postgres/                  # Database client
│   │   └── server/                    # HTTP server
│   ├── events/                        # Domain events
│   └── bootstrap/                     # Dependency injection
├── migrations/                        # Database migrations
├── pkg/
│   └── errors/                        # Shared error types
├── postman/                           # API collections
├── resources/                         # Config examples
├── Makefile                           # Build and dev commands
├── docker-compose.yml                 # Local infrastructure
└── go.mod
```

### Architectural Patterns

#### Hexagonal Architecture

```
┌─────────────────────────────────────────────────┐
│                   Adapters                      │
│  ┌────────────┐               ┌─────────────┐  │
│  │  HTTP      │               │  PostgreSQL │  │
│  │  Handlers  │               │  Repository │  │
│  └─────┬──────┘               └──────┬──────┘  │
│        │                             │         │
│  ┌─────▼──────────────────────────▼────────┐  │
│  │         Ports (Interfaces)             │  │
│  └────────────────┬───────────────────────┘  │
│                   │                           │
│  ┌────────────────▼──────────────────────┐   │
│  │       Application Layer               │   │
│  │   (Use Cases, Commands, Queries)      │   │
│  └────────────────┬──────────────────────┘   │
│                   │                           │
│  ┌────────────────▼──────────────────────┐   │
│  │          Domain Layer                 │   │
│  │  (Entities, Value Objects, Events)    │   │
│  └───────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
```

#### CQRS Separation

- **Commands** - State-changing operations (Create, Update, Delete)
- **Queries** - Read operations (Get, List, Search)

Each feature has separate command and query services.

#### Event-Driven

All state changes emit domain events:
- `DocumentCreated`, `DocumentUpdated`
- `VersionCreated`, `VersionPublished`
- `DeprecationRequested`, `VersionDeprecated`
- `ApprovalCreated`, `ApprovalApproved`

---

## 🛠️ Development

### Available Make Commands

View all commands:
```bash
make help
```

#### Development Commands

```bash
# Run the application
make run

# Build binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Tidy dependencies
make tidy
```

#### Docker Commands

```bash
# Start all infrastructure
make docker-up

# Stop all infrastructure
make docker-down

# Restart infrastructure
make docker-restart

# Check running containers
make docker-ps

# View logs
make docker-postgres-logs
make docker-openfga-logs
make docker-localstack-logs
```

#### Database Commands

```bash
# Run migrations
make db-migrate

# Rollback last migration
make db-migrate-down

# Rollback all migrations
make db-migrate-down-all

# Check migration version
make db-migrate-version

# Reset database (WARNING: destructive)
make db-reset

# Connect to database shell
make db-shell

# Create new migration
make migrate-create NAME=add_new_table
```

#### OpenFGA Commands

```bash
# Create store
make openfga-setup

# List stores
make openfga-list-stores

# Check health
make openfga-health

# Open playground
make openfga-playground
```

#### S3 Commands

```bash
# Create bucket
make s3-create-bucket

# List buckets
make s3-list-buckets

# List objects
make s3-list-objects
```

#### Health Checks

```bash
# Check all services
make health

# Check port availability
make ports
```

### Adding a New Feature

Follow the `/go-feature-implementation` skill pattern:

1. **Domain Layer** - Create aggregates, entities, value objects
2. **Application Layer** - Define use cases, inputs, outputs
3. **Ports** - Define interfaces
4. **Adapters** - Implement repositories and handlers
5. **Events** - Define domain events
6. **Bootstrap** - Wire dependencies

See `ARCHITECTURE.md` for detailed guidelines.

---

## 🧪 Testing

### Run All Tests

```bash
make test
```

### Run with Coverage

```bash
make test-coverage
```

This generates `coverage.html` which you can open in a browser.

### Manual API Testing

Use the provided Postman collections:

```bash
# Import to Postman
# 1. Import postman/*.postman_collection.json
# 2. Import postman/local.postman_environment.json
# 3. Set Bearer token in environment
# 4. Start making requests
```

### Example cURL Requests

```bash
# Create a document
curl -X POST http://localhost:8080/documents \
  -H "Authorization: Bearer test" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Document",
    "type": "policy",
    "description": "A test document"
  }'

# List documents
curl http://localhost:8080/documents \
  -H "Authorization: Bearer test"

# Get document
curl http://localhost:8080/documents/<document-id> \
  -H "Authorization: Bearer test"
```

---

## 🚢 Deployment

### Building for Production

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/document-registry cmd/api/main.go
```

### Docker Build

```bash
# Build image
docker build -t document-registry:latest .

# Run container
docker run -p 8080:8080 \
  -e ENV=prod \
  -e POSTGRES_HOST=<db-host> \
  -e OPENFGA_URL=<openfga-url> \
  document-registry:latest
```

### Environment-Specific Config

Use different `.env` files for each environment:

```bash
# Development
ENV=dev go run cmd/api/main.go

# Staging
ENV=staging go run cmd/api/main.go

# Production
ENV=prod ./bin/document-registry
```

### Database Migrations

In production, run migrations before deploying:

```bash
migrate -path migrations \
  -database "postgresql://user:pass@host:5432/dbname?sslmode=require" \
  up
```

---

## 🐛 Troubleshooting

### Common Issues

#### "Port 8080 already in use"

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
HTTP_PORT=8081 make run
```

#### "Cannot connect to PostgreSQL"

```bash
# Check if container is running
make docker-ps

# Check PostgreSQL logs
make docker-postgres-logs

# Restart PostgreSQL
make docker-postgres-stop
make docker-postgres
```

#### "OpenFGA store not found"

```bash
# Create new store
make openfga-setup

# Copy STORE_ID to .env
# Update OPENFGA_STORE_ID in .env
```

#### "Database migration failed"

```bash
# Check current version
make db-migrate-version

# Force to specific version
make db-migrate-force VERSION=4

# Or reset and re-run
make db-reset
make db-migrate
```

#### "S3 bucket not found"

```bash
# Create bucket
make s3-create-bucket

# Verify
make s3-list-buckets
```

### Debugging

#### Enable Debug Logging

```bash
LOG_LEVEL=debug make run
```

#### Database Queries

```bash
# Connect to database
make db-shell

# Check tables
\dt

# View documents
SELECT * FROM documents;

# View versions
SELECT * FROM versions;

# Exit
\q
```

#### Check Service Health

```bash
# All services
make health

# Individual checks
curl http://localhost:8080/health
curl http://localhost:8081/healthz
docker exec postgres pg_isready
```

### Getting Help

- **Architecture Questions**: See `ARCHITECTURE.md`
- **Feature Documentation**: See `docs/` directory
- **Issues**: Open a GitHub issue
- **Discussions**: GitHub Discussions

---

## 📚 Documentation

- `ARCHITECTURE.md` - Architecture guidelines
- `IMPLEMENTATION.md` - Implementation details
- `postman/README.md` - API testing guide
- Feature-specific docs in `docs/`:
  - `DEPRECATION_FEATURE_COMPLETE.md`
  - `DEPRECATION_FLOWS_COMPLETE.md`
  - `AUTHORIZATION_IMPLEMENTATION.md`
  - `EVENT_HANDLERS_ANALYSIS.md`

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Follow the architecture guidelines in `ARCHITECTURE.md`
4. Write tests
5. Run `make lint` and `make fmt`
6. Submit a pull request

---

## 📄 License

[Your License Here]

---

## 👥 Authors

- Your Name / Team

---

## 🙏 Acknowledgments

- Built following Clean Architecture principles
- Uses err2 for error handling
- Implements CQRS and DDD patterns
- OpenFGA for authorization
- PostgreSQL for persistence
- S3-compatible storage

---

**Happy Coding!** 🚀