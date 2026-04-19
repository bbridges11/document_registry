# 🚀 Quickstart Guide

Get the Document Registry running locally in **5 minutes**.

---

## ✅ Prerequisites Checklist

Before starting, ensure you have:

- [ ] **Go 1.25+** installed ([Download](https://golang.org/doc/install))
- [ ] **Docker** installed and running ([Download](https://docs.docker.com/get-docker/))
- [ ] **Make** available (pre-installed on macOS/Linux)
- [ ] **Git** for cloning the repository

**Verify installations**:
```bash
go version     # Should show 1.25 or higher
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
- **PostgreSQL** (localhost:5432)
- **OpenFGA** (localhost:8081)
- **LocalStack S3** (localhost:4566)

Wait for services to be ready (~30 seconds).

### Step 3: Setup OpenFGA Store (30 seconds)

```bash
make openfga-setup
```

**⚠️ IMPORTANT**: Copy the `OPENFGA_STORE_ID` from the output. You'll need it in Step 4.

Example output:
```
✓ Store created successfully

Add this to your .env file:
OPENFGA_STORE_ID=01HXXXXXXXXXXXXXXXXXXX
```

### Step 4: Create Configuration File (1 minute)

Create `.env` file in the project root:

```bash
cat > .env << 'EOF'
# Server
HTTP_PORT=8080
ENV=local

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=document_registry

# OpenFGA (REPLACE WITH YOUR STORE ID FROM STEP 3)
OPENFGA_ENABLED=false
OPENFGA_URL=http://localhost:8081
OPENFGA_STORE_ID=<PASTE-YOUR-STORE-ID-HERE>

# AWS/S3 (LocalStack)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_S3_ENDPOINT=http://localhost:4566
AWS_S3_BUCKET=document-registry
AWS_S3_FORCE_PATH_STYLE=true

# Logging
LOG_LEVEL=info
LOG_FORMAT=console
EOF
```

**Replace `<PASTE-YOUR-STORE-ID-HERE>` with your actual store ID from Step 3!**

### Step 5: Create S3 Bucket (15 seconds)

```bash
make s3-create-bucket
```

### Step 6: Run Database Migrations (30 seconds)

```bash
make db-migrate
```

This creates all required tables.

### Step 7: Start the Application (15 seconds)

```bash
make run
```

You should see:
```
INFO  Starting Document Registry
INFO  Server listening on :8080
```

### Step 8: Verify It's Working (15 seconds)

Open a new terminal and run:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy"}
```

---

## ✅ Success! What's Next?

### Try the API

```bash
# Create a document
curl -X POST http://localhost:8080/documents \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Document",
    "type": "policy",
    "description": "Testing the API"
  }'
```

### Use Postman

1. Import `postman/document-registry.postman_collection.json`
2. Import `postman/local.postman_environment.json`
3. Set token to `test-token` in environment
4. Start making requests!

### Explore the Features

- **Documents**: Create, search, update documents
- **Versions**: Version control with approval workflow
- **Deprecation**: Mark old versions as deprecated
- **Search**: Full-text search with filters
- **Stakeholders**: Manage document collaborators

---

## 🛑 Stopping the Application

### Stop the App

Press `Ctrl+C` in the terminal where `make run` is running.

### Stop Infrastructure

```bash
make docker-down
```

---

## 🔄 Restarting Later

When you come back later:

```bash
# Start infrastructure
make docker-up

# Wait 5 seconds for services to be ready
sleep 5

# Start the app
make run
```

No need to run migrations or setup OpenFGA again!

---

## 🐛 Troubleshooting

### "Port 8080 already in use"

```bash
# Check what's using it
lsof -i :8080

# Use different port
HTTP_PORT=8081 make run
```

### "Cannot connect to PostgreSQL"

```bash
# Restart PostgreSQL
make docker-postgres-stop
make docker-postgres

# Wait 5 seconds
sleep 5

# Try again
make run
```

### "Migration failed"

```bash
# Reset database
make db-reset

# Wait 5 seconds
sleep 5

# Run migrations again
make db-migrate
```

### "Store not found"

```bash
# Create new store
make openfga-setup

# Copy the new STORE_ID to .env
```

### Still Having Issues?

```bash
# Clean everything and start fresh
make docker-clean

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
| `make help` | Show all available commands |

---

## 🎓 Learning More

- **Full README**: See `README.md` for complete documentation
- **API Guide**: See `postman/README.md` for API testing
- **Architecture**: See `ARCHITECTURE.md` for design patterns
- **Deprecation Feature**: See `docs/DEPRECATION_FEATURE_COMPLETE.md`

---

## 💡 Pro Tips

### Use the All-in-One Command

Instead of Steps 2-5, you can use:

```bash
make dev-full
```

This does everything except creating the `.env` file.

### Check Service Health

```bash
make health
```

Shows status of all services:
- PostgreSQL
- OpenFGA
- Document Registry

### View Logs

```bash
# PostgreSQL
make docker-postgres-logs

# OpenFGA
make docker-openfga-logs

# LocalStack
make docker-localstack-logs
```

### Database Shell

```bash
make db-shell
```

Connects you to the PostgreSQL database.

---

## 🎉 You're All Set!

The Document Registry is now running at:
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **OpenFGA Playground**: http://localhost:3000

**Happy building!** 🚀

---

## 📚 Next Steps

1. **Explore the API** - Use the Postman collection
2. **Read the docs** - Check out `README.md` for full details
3. **Try the features** - Create documents, versions, and test deprecation
4. **Build something** - Integrate with your application!

Need help? Check the full `README.md` or open an issue on GitHub.