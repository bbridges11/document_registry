# Document Registry API - Bruno Collection

Complete API collection for the Document Registry service.

## 📁 Collection Structure

```
bruno/
├── bruno.json              # Collection configuration
├── files/                  # Example files for testing
│   └── pattern-example.yaml
├── Health/                 # Health check endpoints (ECS)
│   ├── Health Check (ALB).bru
│   ├── Liveness Probe.bru
│   ├── Readiness Check.bru
│   └── Startup Check.bru
├── Documents/              # Document CRUD operations
│   ├── Create Document.bru
│   ├── Get Document.bru
│   ├── List Documents.bru
│   ├── Search Documents.bru
│   └── Update Document.bru
├── Versions/               # Version management
│   ├── Create Version.bru
│   ├── Get Version.bru
│   ├── List Versions.bru
│   ├── Submit Version.bru
│   ├── Review Version.bru      # ⭐ NEW: Optional manual review
│   ├── Approve Version.bru
│   ├── Reject Version.bru
│   ├── Publish Version.bru
│   └── Get Version Status.bru
├── Approvals/              # Approval management
│   ├── Grant Approval.bru
│   ├── Revoke Approval.bru
│   ├── Get Approval.bru
│   ├── List Approvals.bru
│   └── Get Approval Summary.bru
├── Validation/             # Content validation
│   └── Validate Content.bru
├── Deprecation/            # Deprecation workflow
│   ├── Request Deprecation.bru
│   └── Approve Deprecation.bru
├── Stakeholders/           # Stakeholder management
│   ├── Add Stakeholder.bru
│   └── List Stakeholders.bru
└── Users/                  # User management
    ├── Create User.bru
    ├── Get User.bru
    ├── List Users.bru
    ├── Update User.bru
    ├── Deactivate User.bru
    ├── Activate User.bru
    └── Delete User.bru
```

## 🚀 Quick Start

### 1. Install Bruno

Download from [https://www.usebruno.com/](https://www.usebruno.com/)

### 2. Open Collection

1. Open Bruno
2. Click "Open Collection"
3. Navigate to `document-registry/bruno`
4. Select the `bruno` folder

### 3. Configure Variables

The collection uses these variables (configured in `bruno.json`):

- `base_url`: http://localhost:8080 (default)
- `user_id`: user-123 (default)
- `document_id`: (auto-populated after creating document)
- `version_id`: (auto-populated after creating version)
- `stakeholder_id`: (auto-populated after adding stakeholder)
- `deprecation_id`: (auto-populated after requesting deprecation)
- `created_user_id`: (auto-populated after creating user)

### 4. Start the Service

```bash
cd document-registry
make docker-up
make db-migrate
make s3-create-bucket
make run
```

## 📝 Usage Guide

### Basic Workflow

1. **Check Health**
   - Run `Health/Health Check (ALB)` to verify service is running
   - Should return `{"status": "healthy", ...}`

2. **Create a Document**
   - Run `Documents/Create Document`
   - This automatically:
     - Encodes `files/pattern-example.yaml` as base64
     - Creates document with first version
     - Saves `document_id` and `version_id` to variables

3. **Submit for Review**
   - Run `Versions/Submit Version`
   - Moves version from DRAFT → SUBMITTED

4. **Review (Optional)**
   - Run `Versions/Review Version` to manually transition to IN_REVIEW
   - OR skip this step - first approval will auto-transition to IN_REVIEW

5. **Approve Version**
   - Run `Versions/Approve Version` (as engineer)
   - Run `Versions/Approve Version` again (as architect)
   - Approval role is automatically determined from user's role
   - Version moves to APPROVED when all required approvals received

6. **Publish Version**
   - Run `Versions/Publish Version`
   - Makes version live (APPROVED → PUBLISHED)

### Working with Files

All requests that need file content use **pre-request scripts** to handle base64 encoding:

```javascript
const fs = require("fs");
const filePath = bru.getVar("file_path") || "./files/pattern-example.yaml";
const fileBuffer = fs.readFileSync(filePath);
const base64String = fileBuffer.toString("base64");
bru.setVar("file_base64", base64String);
```

**To use a different file:**

1. Add your file to `bruno/files/`
2. Set the `file_path` variable before running the request:
   - Right-click request → "Settings" → "Script" → "Pre Request"
   - Or set in collection variables

**Example files provided:**
- `files/pattern-example.yaml` - Example pattern document

## 🔑 Environment Variables

The collection references these variables from your `.env`:

```bash
DOCUMENT_REGISTRY_HTTP_PORT=8080  # Used in base_url
```

If running on a different port, update `base_url` in collection variables.

## 📋 Request Categories

### Health Checks (ECS-optimized)

- **Health Check (ALB)** - For ALB target groups (Postgres + S3)
- **Liveness Probe** - For container health checks (always 200)
- **Readiness Check** - For monitoring (Postgres + S3 + SNS)
- **Startup Check** - For container startup (10s timeout)

### Documents

- **Create Document** - Creates document with first version (base64 content)
- **Get Document** - Retrieve document details
- **List Documents** - Paginated list
- **Search Documents** - Filter by name, tags, type, etc.
- **Update Document** - Update metadata (name, description, tags)

### Versions

- **Create Version** - Add new version to document (base64 content)
- **Get Version** - Retrieve version details
- **List Versions** - All versions of a document
- **Submit Version** - Submit for review (DRAFT → SUBMITTED)
- **Review Version** - ⭐ Optional manual transition to IN_REVIEW state
- **Approve Version** - Approve version (auto-transitions to IN_REVIEW if needed, role determined by server)
- **Reject Version** - Reject version (auto-transitions to IN_REVIEW first, then REJECTED)
- **Publish Version** - Publish approved version (APPROVED → PUBLISHED)
- **Get Version Status** - Detailed workflow status with approval progress

### Validation

- **Validate Content** - Validate YAML/JSON before creating (base64 content)

### Approvals

- **Grant Approval** - Grant approval for a version as stakeholder
- **Revoke Approval** - Revoke a previously granted approval
- **Get Approval** - Retrieve approval details by ID
- **List Approvals** - View all approvals for a version
- **Get Approval Summary** - Check approval progress (required/received/remaining)

### Deprecation

- **Request Deprecation** - Request to deprecate a version
- **Approve Deprecation** - Approve deprecation request

### Stakeholders

- **Add Stakeholder** - Add user to document
- **List Stakeholders** - View all stakeholders

### Users

- **Create User** - Register new user in system
- **Get User** - Retrieve user details
- **List Users** - View all users
- **Update User** - Update user name/role (admin only)
- **Deactivate User** - Soft delete user (admin only)
- **Activate User** - Reactivate user (admin only)
- **Delete User** - Permanently delete user (admin only)

## 🎯 Testing Workflows

### Complete Approval Workflow

**Option 1: Auto-Transition (Recommended)**
1. Create Document → `document_id` saved
2. Submit Version → DRAFT → SUBMITTED
3. Approve Version (engineer) → Auto-transitions to IN_REVIEW, creates approval (technical)
4. Approve Version (architect) → Creates approval (architect), transitions to APPROVED
5. Publish Version → APPROVED → PUBLISHED

**Option 2: With Manual Review**
1. Create Document → `document_id` saved
2. Submit Version → DRAFT → SUBMITTED
3. Review Version → SUBMITTED → IN_REVIEW (explicit)
4. Approve Version (engineer) → Creates approval (technical)
5. Approve Version (architect) → Creates approval (architect), transitions to APPROVED
6. Publish Version → APPROVED → PUBLISHED

**Versioning Flow**
1. Create Version (v2.0.0)
2. Submit, approve, publish v2.0.0
3. Previous version auto-deprecates

### Validation Workflow

1. Validate Content → Check YAML is valid
2. Create Document → If validation passes
3. Create Version → Each version validated

### Deprecation Workflow

1. Request Deprecation → Creates deprecation request
2. Approve Deprecation → Version moves to DEPRECATED

### User Management Workflow

1. Create User → `created_user_id` saved
2. List Users → View all users in system
3. Update User → Change role (admin only)
4. Deactivate User → Soft delete (admin only)
5. Activate User → Restore access (admin only)

## 🔧 Customization

### Adding Custom Files

1. Create your file in `bruno/files/`
2. Update the `file_path` variable in request
3. Pre-request script will auto-encode to base64

### Changing User ID

Update `user_id` in collection variables to test different users.

### Testing Different Document Types

Currently supports:
- `pattern` - YAML validation, approval workflow

To test, change `document_type` in request body.

## 📚 API Reference

### ⭐ Important: Approval Role Changes

**The `role` field has been REMOVED from approve/reject requests.**

**Before** (old behavior):
```json
POST /versions/{id}/approve
{
  "role": "engineer",      // ❌ No longer needed
  "comment": "LGTM"
}
```

**Now** (current behavior):
```json
POST /versions/{id}/approve
{
  "comment": "LGTM"        // ✅ Only comment needed
}
```

**How approval roles work now:**

| User Role | Approves As | Notes |
|-----------|-------------|-------|
| admin | technical | First available role |
| architect | architect | - |
| engineer | technical | - |
| product | ❌ Cannot approve | Can only publish |
| viewer | ❌ Cannot approve | Read-only |

**Pattern document requirements:**
- 1 technical approval (engineer or admin)
- 1 architect approval (architect or admin)
- Same user cannot provide multiple roles

**Auto-Transition Behavior:**

When you approve/reject a version in SUBMITTED state:
1. Version automatically transitions to IN_REVIEW
2. Then approval/rejection is processed
3. No need to call `/review` separately (but you can if you want explicit control)

See full documentation: [docs/API_WORKFLOW.md](../docs/API_WORKFLOW.md)

See [README.md](../README.md#api-documentation) for complete API documentation.

## 🐛 Troubleshooting

### "Cannot read file"

**Error:** Pre-request script can't find file

**Solution:** Ensure file exists in `bruno/files/` or update `file_path` variable

### "Invalid base64"

**Error:** Content not properly encoded

**Solution:** Check pre-request script ran (look for `file_base64` variable)

### "Document type required"

**Error:** Validation endpoint requires document_type

**Solution:** Ensure `document_type` field is in request body

### "Document ID not set"

**Error:** `document_id` variable is empty

**Solution:** Run "Create Document" first (post-response script saves ID)

## 💡 Tips

1. **Run in order** - Requests are numbered (seq) for suggested order
2. **Check responses** - Post-response scripts auto-save IDs
3. **Use search** - Filter documents by tags for easier testing
4. **Validate first** - Run validation before creating documents
5. **Check health** - Always verify service is healthy first

## 🔗 Related Documentation

- [README.md](../README.md) - Full service documentation
- [QUICKSTART.md](../QUICKSTART.md) - 5-minute setup guide
- [ARCHITECTURE.md](../ARCHITECTURE.md) - Architecture details
- [MIGRATIONS.md](../MIGRATIONS.md) - Database schema info