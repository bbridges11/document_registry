# Document Registry API - HTTP Endpoints Reference for UI Development

This guide provides complete input/output specifications for all HTTP endpoints in the Document Registry API.

**Base URL**: `http://localhost:8080`

**Authentication**: All endpoints (except health checks) require `X-User-ID` header with a 7-character SID (e.g., "a123456")

---

## Table of Contents

1. [Documents](#documents)
2. [Versions](#versions)
3. [Stakeholders](#stakeholders)
4. [Approvals](#approvals)
5. [Deprecations](#deprecations)
6. [Validation](#validation)

---

## Documents

### 1. Create Document
**POST** `/documents`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "name": "API Authentication Pattern",  // Required
  "description": "OAuth 2.0 implementation guide",
  "document_type": "pattern",  // Required, one of: pattern
  "tags": ["authentication", "oauth", "security"]  // Optional
}
```

**Success Response** (201 Created):
```json
{
  "id": "doc-uuid-here",
  "name": "API Authentication Pattern",
  "description": "OAuth 2.0 implementation guide",
  "document_type": "pattern",
  "tags": ["authentication", "oauth", "security"],
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-04-30T12:00:00Z",
  "updated_at": "2026-04-30T12:00:00Z",
  "version": {
    "id": "version-uuid-here",
    "version": "1.0.0",
    "status": "draft"
  }
}
```

**Note**: Creating a document automatically creates version 1.0.0 in DRAFT status

---

### 2. Get Document
**GET** `/documents/:id`

**Headers**:
```
X-User-ID: a123456
```

**URL Parameters**:
- `id` (UUID): Document ID

**Success Response** (200 OK):
```json
{
  "id": "doc-uuid-here",
  "name": "API Authentication Pattern",
  "description": "OAuth 2.0 implementation guide",
  "document_type": "pattern",
  "tags": ["authentication", "oauth", "security"],
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-04-30T12:00:00Z",
  "updated_at": "2026-04-30T12:00:00Z"
}
```

---

### 3. List Documents
**GET** `/documents`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "documents": [
    {
      "id": "doc-uuid-here",
      "name": "API Authentication Pattern",
      "description": "OAuth 2.0 implementation guide",
      "document_type": "pattern",
      "tags": ["authentication", "oauth"],
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2026-04-30T12:00:00Z",
      "updated_at": "2026-04-30T12:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 4. Update Document
**PUT** `/documents/:id`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "name": "Updated Pattern Name",
  "description": "Updated description",
  "tags": ["new", "tags"]
}
```

**Success Response** (200 OK): Same structure as Get Document

---

### 5. Search Documents
**GET** `/documents/search?q={query}`

**Headers**:
```
X-User-ID: a123456
```

**Query Parameters**:
- `q` (string): Search query

**Success Response** (200 OK): Same structure as List Documents

---

## Versions

### 1. Create Version
**POST** `/documents/:documentId/versions`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**URL Parameters**:
- `documentId` (UUID): Document ID

**Request Body**:
```json
{
  "content": "base64-encoded-yaml-or-json-content-here",  // Required, base64 encoded file content
  "description": "Version description"  // Optional
}
```

**Success Response** (201 Created):
```json
{
  "id": "version-uuid-here",
  "document_id": "doc-uuid-here",
  "version": "1.1.0",
  "status": "draft",
  "description": "Version description",
  "content_ref": "s3://bucket/path/to/content",
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-04-30T12:00:00Z",
  "updated_at": "2026-04-30T12:00:00Z",
  "metadata": {},
  "validation_result": {
    "valid": true,
    "errors": [],
    "warnings": []
  }
}
```

**Note**: Content should be base64-encoded YAML or JSON file

---

### 2. Get Version
**GET** `/versions/:id`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "id": "version-uuid-here",
  "document_id": "doc-uuid-here",
  "version": "1.1.0",
  "status": "draft",
  "description": "Version description",
  "content_ref": "s3://bucket/path",
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-04-30T12:00:00Z",
  "updated_at": "2026-04-30T12:00:00Z",
  "metadata": {}
}
```

---

### 3. List Versions by Document
**GET** `/documents/:documentId/versions`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "versions": [
    {
      "id": "version-uuid-here",
      "document_id": "doc-uuid-here",
      "version": "1.1.0",
      "status": "published",
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2026-04-30T12:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 4. Update Version
**PUT** `/versions/:id`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "description": "Updated description"
}
```

**Success Response** (200 OK): Same structure as Get Version

---

### 5. Submit Version
**POST** `/versions/:id/submit`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (204 No Content)

**Note**: Transitions version from DRAFT to SUBMITTED, triggers approval workflow

---

### 6. Review Version
**POST** `/versions/:id/review`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (204 No Content)

**Note**: Transitions version from SUBMITTED to IN_REVIEW (optional manual step)

---

### 7. Approve Version
**POST** `/versions/:id/approve`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "comment": "Looks good, approved"  // Optional
}
```

**Success Response** (204 No Content)

**Note**: Adds approval, transitions to APPROVED when all required approvals received

---

### 8. Reject Version
**POST** `/versions/:id/reject`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "reason": "Needs more documentation"  // Required
}
```

**Success Response** (204 No Content)

**Note**: Transitions version to REJECTED

---

### 9. Publish Version
**POST** `/versions/:id/publish`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "destination": "dev-portal",  // Required: dev-portal, s3, sns, mock
  "environment": "staging"      // Required: dev, staging, prod
}
```

**Success Response** (200 OK):
```json
{
  "publication_id": "pub-uuid-here",
  "version_id": "version-uuid-here",
  "destination": "dev-portal",
  "environment": "staging",
  "status": "success",
  "published_at": "2026-04-30T12:00:00Z"
}
```

**Note**: 
- Transitions version to PUBLISHED
- Auto-deprecates previously published version
- Can only publish APPROVED versions

---

### 10. Get Version Status
**GET** `/versions/:id/status`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "version_id": "version-uuid-here",
  "status": "in_review",
  "workflow_type": "pattern_approval",
  "can_submit": false,
  "can_approve": true,
  "can_reject": true,
  "can_publish": false,
  "required_approvals": 2,
  "received_approvals": 1,
  "remaining_roles": ["architect"]
}
```

---

### 11. Get Version Approvals
**GET** `/versions/:id/approvals`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "approvals": [
    {
      "id": "approval-uuid-here",
      "version_id": "version-uuid-here",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "technical",
      "comment": "Approved",
      "approved_at": "2026-04-30T12:00:00Z"
    }
  ],
  "count": 1
}
```

---

## Stakeholders

### 1. Add Stakeholder
**POST** `/documents/:documentId/stakeholders`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",  // Required, internal UUID
  "role": "contributor"  // Required, one of: owner, contributor, viewer
}
```

**Success Response** (201 Created):
```json
{
  "id": "stakeholder-uuid-here",
  "document_id": "doc-uuid-here",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "role": "contributor",
  "created_at": "2026-04-30T12:00:00Z"
}
```

**Note**: Document creator is automatically added as owner stakeholder

---

### 2. List Stakeholders
**GET** `/documents/:documentId/stakeholders`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "stakeholders": [
    {
      "id": "stakeholder-uuid-here",
      "document_id": "doc-uuid-here",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "owner",
      "created_at": "2026-04-30T12:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 3. Remove Stakeholder
**DELETE** `/documents/:documentId/stakeholders/:userId`

**Headers**:
```
X-User-ID: a123456
```

**URL Parameters**:
- `documentId` (UUID): Document ID
- `userId` (string): User internal UUID

**Success Response** (204 No Content)

---

## Approvals

### 1. Get Approval
**GET** `/approvals/:id`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "id": "approval-uuid-here",
  "version_id": "version-uuid-here",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "role": "technical",
  "comment": "Approved",
  "approved_at": "2026-04-30T12:00:00Z"
}
```

---

### 2. List Approvals by Version
**GET** `/versions/:versionId/approvals`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "approvals": [
    {
      "id": "approval-uuid-here",
      "version_id": "version-uuid-here",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "technical",
      "comment": "Approved",
      "approved_at": "2026-04-30T12:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 3. Get Approval Summary
**GET** `/approvals/summary?version_id={versionId}`

**Headers**:
```
X-User-ID: a123456
```

**Query Parameters**:
- `version_id` (UUID): Version ID

**Success Response** (200 OK):
```json
{
  "version_id": "version-uuid-here",
  "required_count": 2,
  "received_count": 1,
  "required_roles": ["technical", "architect"],
  "received_roles": ["technical"],
  "remaining_roles": ["architect"],
  "is_approved": false,
  "approvals": [
    {
      "id": "approval-uuid-here",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "technical",
      "comment": "Approved",
      "approved_at": "2026-04-30T12:00:00Z"
    }
  ]
}
```

---

## Deprecations

### 1. Request Deprecation
**POST** `/versions/:id/deprecate`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "reason": "Superseded by version 2.0.0",  // Required
  "deprecation_note": "Migrate by end of Q2"  // Optional
}
```

**Success Response** (201 Created):
```json
{
  "id": "deprecation-uuid-here",
  "version_id": "version-uuid-here",
  "document_id": "doc-uuid-here",
  "reason": "Superseded by version 2.0.0",
  "deprecation_note": "Migrate by end of Q2",
  "requested_by": "550e8400-e29b-41d4-a716-446655440000",
  "requested_at": "2026-04-30T12:00:00Z",
  "status": "pending"
}
```

---

### 2. Approve Deprecation
**POST** `/versions/:id/deprecation/approve`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "role": "admin",     // Required: admin, architect, technical
  "comment": "Approved"  // Optional
}
```

**Success Response** (204 No Content)

---

### 3. Reject Deprecation
**POST** `/versions/:id/deprecation/reject`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "reason": "Migration path not clear"  // Required
}
```

**Success Response** (204 No Content)

---

### 4. Cancel Deprecation
**POST** `/versions/:id/deprecation/cancel`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (204 No Content)

---

### 5. Get Deprecation
**GET** `/versions/:id/deprecation`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "id": "deprecation-uuid-here",
  "version_id": "version-uuid-here",
  "document_id": "doc-uuid-here",
  "reason": "Superseded by version 2.0.0",
  "deprecation_note": "Migrate by end of Q2",
  "requested_by": "550e8400-e29b-41d4-a716-446655440000",
  "requested_at": "2026-04-30T12:00:00Z",
  "status": "pending"
}
```

---

### 6. Get Deprecation Status
**GET** `/versions/:id/deprecation/status`

**Headers**:
```
X-User-ID: a123456
```

**Success Response** (200 OK):
```json
{
  "deprecation_id": "deprecation-uuid-here",
  "version_id": "version-uuid-here",
  "status": "pending",
  "required_approvals": 2,
  "received_approvals": 1,
  "approvals": [
    {
      "approver_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "admin",
      "comment": "Approved",
      "approved_at": "2026-04-30T12:00:00Z"
    }
  ],
  "remaining_roles": ["architect"]
}
```

---

## Validation

### 1. Validate Content
**POST** `/validate`

**Headers**:
```
X-User-ID: a123456
Content-Type: application/json
```

**Request Body**:
```json
{
  "content": "base64-encoded-yaml-or-json-content-here",  // Required, base64 encoded file content
  "document_type": "pattern"  // Required
}
```

**Success Response** (200 OK):
```json
{
  "valid": true,
  "errors": [],
  "warnings": [
    "Consider adding more examples"
  ]
}
```

**Validation Failure Response** (200 OK):
```json
{
  "valid": false,
  "errors": [
    "Missing required field: openapi.info.title",
    "Invalid version format"
  ],
  "warnings": []
}
```

**Note**: Content should be base64-encoded YAML or JSON file

---

## Common Error Responses

All endpoints may return these error responses:

### 400 Bad Request
```json
{
  "error": "invalid request body"
}
```

### 401 Unauthorized
```json
{
  "error": "X-User-ID header is required"
}
```

### 403 Forbidden
```json
{
  "error": "admin privileges required"
}
```

### 404 Not Found
```json
{
  "error": "resource not found"
}
```

### 409 Conflict
```json
{
  "error": "resource already exists"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Status Enums

### Version Status
- `draft` - Initial state, can be edited
- `submitted` - Submitted for review
- `in_review` - Under manual review (optional)
- `approved` - All approvals received
- `rejected` - Rejected by reviewer
- `published` - Live and accessible
- `deprecated` - Marked for removal

### Approval Roles
- `technical` - Technical review
- `architect` - Architecture review
- `admin` - Administrative approval

### User Roles
- `admin` - Full access
- `contributor` - Can create/edit documents
- `viewer` - Read-only access

### Stakeholder Roles
- `owner` - Document owner (full control)
- `contributor` - Can edit
- `viewer` - Can view

### Publication Destinations
- `dev-portal` - Developer portal (HTTP)
- `s3` - S3 bucket storage
- `sns` - SNS notification
- `mock` - Mock publisher (testing)

### Environments
- `dev` - Development
- `staging` - Staging
- `prod` - Production

---

## Workflow Summary

### Document Creation Flow
1. POST `/documents` - Create document (auto-creates v1.0.0 DRAFT)
2. Current user is auto-added as owner stakeholder

### Version Publication Flow
1. POST `/documents/:id/versions` - Upload new version content (base64 encoded)
2. POST `/versions/:id/submit` - Submit for approval
3. (Optional) POST `/versions/:id/review` - Manual review
4. POST `/versions/:id/approve` - Approve (repeat for all required roles)
5. POST `/versions/:id/publish` - Publish to destination

### Deprecation Flow
1. POST `/versions/:id/deprecate` - Request deprecation
2. POST `/versions/:id/deprecation/approve` - Approve (repeat for all required roles)
3. Version moves to DEPRECATED status

---

## Notes for UI Development

1. **SID vs UUID**: 
   - SID (7 chars) used in X-User-ID header
   - UUID used in all request/response bodies

2. **Auto-created Resources**:
   - Creating document auto-creates v1.0.0
   - Creating document auto-adds creator as owner stakeholder
   - Publishing new version auto-deprecates previous published version

3. **File Uploads**:
   - All file content sent as base64-encoded strings in JSON
   - Use `content` field with base64-encoded data
   - Content-Type is always `application/json`

4. **Pagination**: Currently not implemented (returns all results)

5. **Filtering**: Limited to search on documents endpoint

6. **Real-time Updates**: Not available (poll for status changes)

---

This guide covers all HTTP endpoints with complete input/output specifications for UI development.
