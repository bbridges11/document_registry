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
│   ├── Review Version.bru      # ⭐ Optional manual review
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
├── Publications/           # ⭐ Publication audit (Admin Only)
│   ├── Get Publication.bru
│   ├── List All Publications.bru
│   ├── List Publications by Version.bru
│   └── List Publications by Document.bru
├── Subscriptions/          # ⭐ SNS email subscriptions (Admin Only)
│   ├── Subscribe User.bru
│   ├── Unsubscribe User.bru
│   └── List Subscriptions.bru
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
- `user_id`: a123456 (7-character SID, default)
- `user_sid`: a123456 (alias for user_id)
- `document_id`: (auto-populated after creating document)
- `version_id`: (auto-populated after creating version)
- `stakeholder_id`: (auto-populated after adding stakeholder)
- `deprecation_id`: (auto-populated after requesting deprecation)
- `created_user_id`: (auto-populated after creating user - internal UUID)
- `created_user_sid`: (auto-populated after creating user - SID)

### 4. Start the Service

```bash
cd document-registry
make docker-up
make db-migrate
make s3-create-bucket
make run
```

## 🆔 User SIDs

**IMPORTANT**: The API now uses 7-character SIDs for user identification.

### SID Format
- **Length**: Exactly 7 characters
- **Characters**: Alphanumeric only (a-z, 0-9)
- **Case**: Case-insensitive (stored as lowercase)
- **Example**: `a123456`, `b789xyz`, `test001`

### Using SIDs

**In X-User-ID Header**:
```http
X-User-ID: a123456
```

**Creating a User**:
```json
POST /users
{
  "sid": "a123456",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "contributor"
}
```

**Response**:
```json
{
  "id": "550e8400-...",          // Internal UUID (ignore)
  "sid": "a123456",       // Use this in X-User-ID
  "email": "user@example.com",
  "name": "John Doe",
  "role": "contributor",
  "active": true,
  "created_at": "2026-04-30T...",
  "updated_at": "2026-04-30T..."
}
```

### Migration Note

- **Old**: X-User-ID used UUID format (550e8400-e29b-41d4-a716-446655440000)
- **New**: X-User-ID uses SID format (a123456)
- All existing requests have been updated to use `{{user_id}}` which defaults to `a123456`

## 📝 Testing Workflows

### Basic Flow

1. **Create User**
   ```
   POST /users
   → Auto-saves created_user_sid
   ```

2. **Create Document**
   ```
   POST /documents
   Headers: X-User-ID: {{user_id}}
   → Auto-saves document_id and version_id
   → Creator automatically added as owner stakeholder
   ```

3. **Submit for Review**
   ```
   POST /versions/{{version_id}}/submit
   Headers: X-User-ID: {{user_id}}
   → Triggers approval workflow
   ```

4. **Grant Approvals**
   ```
   POST /approvals
   Headers: X-User-ID: {{approver_user_id}}
   → Must have approval role
   ```

5. **Publish Version**
   ```
   POST /versions/{{version_id}}/publish
   Headers: X-User-ID: {{user_id}}
   → Version becomes live
   ```

### Manual Review Flow (Optional)

```
Submit → [Optional] Manual Review → Auto Approvals → Publish
```

- `POST /versions/{id}/review` - Trigger manual review
- Review can be requested anytime after submit
- Does not block automatic approvals
- Approval summaries show both manual + auto approvals

### Admin Operations

**Publications** (Admin Only):
- Track all version publications
- Audit who published what and when
- Query by document, version, or time range

**Subscriptions** (Admin Only):
- Subscribe users to SNS topics
- Receive email notifications for events
- Manage subscription lifecycle

## 🔐 Authentication

All endpoints (except health checks and user creation) require:

```http
X-User-ID: a123456
```

This identifies the user making the request using their 7-character SID.

## 📦 Variables Auto-Population

The collection uses post-response scripts to automatically populate variables:

- Creating a user → saves `created_user_sid`
- Creating a document → saves `document_id` and `version_id`
- Adding a stakeholder → saves `stakeholder_id`
- Requesting deprecation → saves `deprecation_id`

This allows seamless request chaining without manual ID copying.

## 🏗️ Document Types

### Pattern (Requires Approval)
```json
{
  "document_type": "pattern",
  "content": "<base64-encoded-yaml>",
  "validation": {
    "format": "yaml"
  }
}
```

**Approval Flow**:
1. Submit version
2. Automatic YAML validation
3. Approvals from required roles
4. Publish when all approved

## 🎯 Best Practices

1. **Always use variables** for IDs rather than hardcoding
2. **Check response status** in post-response scripts
3. **Validate base64 encoding** for document content
4. **Use correct user SIDs** in X-User-ID headers (7 chars)
5. **Test with admin users** for admin-only endpoints

## 🐛 Troubleshooting

### "X-User-ID header is required"
- Ensure you're setting the `X-User-ID` header
- Use SID format (e.g., `a123456`), not UUID

### "Invalid SID format"
- SID must be exactly 7 alphanumeric characters
- Case doesn't matter (converted to lowercase)
- No special characters allowed

### "User not found or inactive"
- User with that SID doesn't exist
- User may be deactivated
- Create user first with `POST /users`

### "Validation failed"
- Check document type requirements
- Ensure content is properly base64 encoded
- Verify YAML syntax for patterns

### "Insufficient approvals"
- Check approval summary: `GET /approvals/summary?version_id={id}`
- Ensure all required approval roles are granted
- Wait for automatic approvals if applicable

## 📚 Additional Resources

- API Documentation: (Link to Swagger/OpenAPI)
- Architecture Guide: See ARCHITECTURE.md
- Error Codes: See pkg/errors/codes.go
