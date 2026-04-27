# 🚀 Quickstart Guide

Get the Document Registry running locally in **5 minutes**.

---

## ✅ Prerequisites Checklist

Before starting, ensure you have:

- [ ] **Go 1.21+** installed ([Download](https://golang.org/doc/install))
- [ ] **Docker** installed and running ([Download](https://docs.docker.com/get-docker/))
- [ ] **Make** available (pre-installed on macOS/Linux)
- [ ] **Git** for cloning the repository

**Verify installations**:
```bash
go version      # Should show 1.21 or higher
docker --version
make --version
```

---

## 🎯 5-Minute Setup

### Step 1: Clone and Enter Directory (30 seconds)

```bash
git clone https://github.com/bbridges_11/document-registry.git
cd document-registry
```

### Step 2: Start Infrastructure (1 minute)

```bash
make docker-up
```

This starts:
- **PostgreSQL 16** (localhost:5432)
- **LocalStack** (S3 + SNS on localhost:4566)

Wait for services to be ready (~30 seconds).

### Step 3: Create Configuration File (1 minute)

Create `.env` file in the project root:

```bash
cat > .env << 'EOF'
# Application
DOCUMENT_REGISTRY_PROFILE=local
DOCUMENT_REGISTRY_APP_NAME=document-registry

# Logging
DOCUMENT_REGISTRY_LOG_LEVEL=info
DOCUMENT_REGISTRY_LOG_FORMAT=console

# Server
DOCUMENT_REGISTRY_HTTP_PORT=8080

# Database (connects to docker-compose Postgres)
DOCUMENT_REGISTRY_POSTGRES_HOST=localhost
DOCUMENT_REGISTRY_POSTGRES_PORT=5432
DOCUMENT_REGISTRY_POSTGRES_DATABASE=document_registry
DOCUMENT_REGISTRY_POSTGRES_USER=postgres
DOCUMENT_REGISTRY_POSTGRES_PASSWORD=postgres
DOCUMENT_REGISTRY_POSTGRES_SSL_MODE=disable

# AWS Region (standard AWS SDK variable works)
AWS_REGION=us-east-1

# S3 (LocalStack)
DOCUMENT_REGISTRY_S3_BUCKET=document-registry
DOCUMENT_REGISTRY_S3_ENDPOINT=http://localhost:4566

# SNS (LocalStack - disabled by default)
DOCUMENT_REGISTRY_SNS_ENABLED=false
DOCUMENT_REGISTRY_SNS_ENDPOINT=http://localhost:4566
EOF
```

### Step 4: Create S3 Bucket (15 seconds)

```bash
make s3-create-bucket
```

### Step 5: Run Database Migrations (30 seconds)

```bash
make db-migrate
```

This creates all required tables using the consolidated migration.

### Step 6: Start the Application (15 seconds)

```bash
make run
```

You should see:
```
INFO  Starting Document Registry  {"profile": "local"}
INFO  Server listening             {"port": 8080}
```

### Step 7: Verify It's Working (15 seconds)

Open a new terminal and run:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "checks": {
    "postgres": "healthy",
    "s3": "healthy"
  },
  "timestamp": "2026-04-26T21:00:00Z"
}
```

---

## ✅ Success! What's Next?

### Try Creating a Document

```bash
# Create a pattern document with YAML content
CONTENT=$(echo 'kind: Pattern
version: v1
components:
  - name: api-gateway
    type: service' | base64)

curl -X POST http://localhost:8080/documents \
  -H "X-User-ID: user-123" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"API Gateway Pattern\",
    \"description\": \"Standard API gateway pattern\",
    \"document_type\": \"pattern\",
    \"tags\": [\"api\", \"gateway\"],
    \"version\": \"1.0.0\",
    \"content\": \"$CONTENT\"
  }"
```

### Check Health Endpoints

```bash
# ALB health check (Postgres + S3)
curl http://localhost:8080/health

# Liveness probe
curl http://localhost:8080/health/live

# Full readiness check (Postgres + S3 + SNS)
curl http://localhost:8080/health/ready

# Startup check
curl http://localhost:8080/health/startup
```

### List Documents

```bash
curl http://localhost:8080/documents \
  -H "X-User-ID: user-123"
```

### Search Documents

```bash
# Search by tag
curl "http://localhost:8080/documents/search?tags=api" \
  -H "X-User-ID: user-123"

# Search by name
curl "http://localhost:8080/documents/search?name=Gateway" \
  -H "X-User-ID: user-123"
```

---

## 🛑 Stopping the Application

### Stop the App

Press `Ctrl+C` in the terminal where `make run` is running.

### Stop Infrastructure

```bash
make docker-down
```

This stops and removes:
- PostgreSQL container
- LocalStack container
- Network

**Note:** Data is preserved in Docker volumes. Restart with `make docker-up` to restore.

---

## 🔄 Restarting Later

When you come back later:

```bash
# Start infrastructure
make docker-up

# Wait 10 seconds for services to be ready
sleep 10

# Start the app
make run
```

**Pro tip:** Use `make dev-full` to do everything except creating .env

---

## 🐛 Troubleshooting

### "Port 8080 already in use"

```bash
# Check what's using it
lsof -i :8080

# Use different port
echo "DOCUMENT_REGISTRY_HTTP_PORT=8081" >> .env
make run
```

### "Cannot connect to PostgreSQL"

```bash
# Check PostgreSQL status
docker ps | grep postgres

# View logs
make docker-postgres-logs

# Restart PostgreSQL
make docker-down
make docker-up
sleep 10
make run
```

### "S3 bucket not found"

```bash
# Recreate bucket
make s3-create-bucket

# Verify bucket exists
make s3-list-buckets
```

### "Migration failed"

```bash
# Check current migration version
make db-status

# Reset database (WARNING: deletes all data)
make db-reset
make db-migrate
```

### Health Check Fails

```bash
# Check individual services
curl http://localhost:8080/health/ready | jq .

# Example healthy response:
# {
#   "status": "ready",
#   "checks": {
#     "postgres": "healthy",
#     "s3": "healthy",
#     "sns": "disabled"
#   },
#   "timestamp": "2026-04-26T21:00:00Z"
# }

# If postgres is unhealthy:
make docker-postgres-logs

# If s3 is unhealthy:
make docker-localstack-logs
make s3-create-bucket
```

### Still Having Issues?

```bash
# Clean everything and start fresh
make docker-clean

# This removes:
# - All containers
# - All volumes (deletes data!)
# - All networks

# Start over from Step 2
make docker-up
```

---

## 📝 Quick Command Reference

| Command | Description |
|---------|-------------|
| `make docker-up` | Start all infrastructure |
| `make docker-down` | Stop all infrastructure |
| `make run` | Start the application |
| `make db-migrate` | Run database migrations |
| `make health` | Check all service health |
| `make s3-create-bucket` | Create S3 bucket in LocalStack |
| `make help` | Show all available commands |

---

## 🎓 Learning More

- **Full README**: See `README.md` for complete documentation
- **API Guide**: See `README.md#api-documentation` for all endpoints
- **Architecture**: See `ARCHITECTURE.md` for design patterns
- **Migrations**: See `MIGRATIONS.md` for database schema management
- **Environment Variables**: See `ENV_MIGRATION.md` for configuration guide

---

## 💡 Pro Tips

### Use the All-in-One Command

Instead of Steps 2-5, you can use:

```bash
make dev-full
```

This runs: `docker-up` → `db-migrate` → `s3-create-bucket` → `run`

You still need to create `.env` file first!

### Check Service Health

```bash
make health
```

Shows status of all services:
- Document Registry (all health endpoints)
- PostgreSQL
- LocalStack (S3)

### View Logs

```bash
# Application logs (if running in background)
tail -f logs/app.log

# PostgreSQL
make docker-postgres-logs

# LocalStack
make docker-localstack-logs
```

### Database Shell

```bash
make db-shell
```

Opens `psql` connected to the database. Try:
```sql
-- List all documents
SELECT id, name, document_type FROM documents;

-- List all versions
SELECT id, document_id, version, status FROM versions;
```

Type `\q` to exit.

### Working with S3

```bash
# List buckets
make s3-list-buckets

# List objects in bucket
aws --endpoint-url=http://localhost:4566 s3 ls s3://document-registry/

# Download a document version content
aws --endpoint-url=http://localhost:4566 s3 cp \
  s3://document-registry/pattern/{document-id}/{version}/content \
  ./downloaded-content.yaml
```

---

## 🎉 You're All Set!

The Document Registry is now running at:
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Liveness**: http://localhost:8080/health/live
- **Readiness**: http://localhost:8080/health/ready

**Happy building!** 🚀

---

## 📚 Next Steps

1. **Try the API** - Create documents, versions, and test workflows
2. **Read the docs** - Check out `README.md` for full details
3. **Explore features** - Test validation, approval, and deprecation
4. **Add document types** - See `MIGRATIONS.md` for extending the type system
5. **Deploy to ECS** - See `README.md#deployment` for production setup

Need help? Check the full `README.md` or open an issue on GitHub.
