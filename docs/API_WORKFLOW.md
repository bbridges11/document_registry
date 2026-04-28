# Version Approval Workflow API Documentation

## Overview

The version approval workflow follows a state machine pattern with automatic transitions and role-based authorization.

---

## State Machine

### Pattern Document Type

```
DRAFT → SUBMITTED → IN_REVIEW → APPROVED → PUBLISHED
                        ↓
                    REJECTED
```

### State Descriptions

- **DRAFT**: Initial state, version can be edited
- **SUBMITTED**: Version submitted for approval, waiting for review
- **IN_REVIEW**: Version is being reviewed/approved, approvals being collected
- **APPROVED**: All required approvals received, ready to publish
- **PUBLISHED**: Version is published and active
- **REJECTED**: Version rejected, terminal state (create new version to retry)

---

## Role-Based Authorization

### User Roles → Approval Roles

| User Role | Can Approve | Approval Role Used | Can Reject | Can Review | Can Publish |
|-----------|-------------|-------------------|-----------|-----------|-------------|
| **admin** | ✅ | technical (first) | ✅ | ✅ | ❌ |
| **architect** | ✅ | architect | ✅ | ✅ | ❌ |
| **engineer** | ✅ | technical | ✅ | ✅ | ❌ |
| **product** | ❌ | N/A | ❌ | ❌ | ✅ |
| **viewer** | ❌ | N/A | ❌ | ❌ | ❌ |

### Approval Requirements (Pattern Documents)

- **1 technical approval**: From engineer or admin
- **1 architect approval**: From architect or admin

**Important**: Same user cannot provide multiple approval roles

---

## Endpoints

### 1. Submit Version

**Endpoint**: `POST /versions/{id}/submit`

**Authorization**: Any authenticated user (creator typically)

**State Transition**: `DRAFT → SUBMITTED`

**Request**:
```json
{
  "comment": "Ready for review"
}
```

**Response**: `204 No Content`

**Next Steps**: Version is ready for approval

---

### 2. Review Version (Optional)

**Endpoint**: `POST /versions/{id}/review`

**Authorization**: Users with approval roles (admin, architect, engineer)

**State Transition**: `SUBMITTED → IN_REVIEW`

**Request**: None (empty body)

**Response**: `204 No Content`

**Use Cases**:
- Team lead signals "ready for approval" without approving themselves
- Explicitly start review phase for audit trail
- Trigger notifications to approvers

**Note**: This is **OPTIONAL**
- First approval will auto-transition `SUBMITTED → IN_REVIEW`
- Use this only if you want explicit control

**Errors**:
- `412 Precondition Failed`: Version not in SUBMITTED state
- `403 Forbidden`: User has no approval roles

---

### 3. Approve Version

**Endpoint**: `POST /versions/{id}/approve`

**Authorization**: Users with approval roles (admin, architect, engineer)

**State Transitions**:
- If `SUBMITTED`: Auto-transitions to `IN_REVIEW`, then creates approval
- If `IN_REVIEW`: Creates approval
- When policy satisfied: Transitions to `APPROVED`

**Request**:
```json
{
  "comment": "Approved - looks good"
}
```

**Response**: `204 No Content`

**Behavior**:
1. First approval from `SUBMITTED` auto-transitions to `IN_REVIEW`
2. Approval role automatically determined from user's role
3. Approval record created
4. If all required approvals collected, version transitions to `APPROVED`

**Events Published**:
- `VersionInReview` (if auto-transition occurs)
- `VersionApproved` (individual approval)
- `VersionFullyApproved` (when policy satisfied)

**Errors**:
- `412 Precondition Failed`: Version not in SUBMITTED or IN_REVIEW
- `403 Forbidden`: User has no approval roles
- `409 Conflict`: User already approved with a different role
- `403 Forbidden`: Author cannot approve their own version

**Example Flow**:

**Scenario 1: Engineer then Architect (auto-transition)**
```
1. Version is SUBMITTED
2. Engineer approves:
   → Auto-transitions SUBMITTED → IN_REVIEW
   → Creates approval (technical)
   → Policy NOT satisfied (need architect)
   → Version is IN_REVIEW

3. Architect approves:
   → Creates approval (architect)
   → Policy satisfied (has technical + architect)
   → Transitions IN_REVIEW → APPROVED
   → Version is APPROVED
```

**Scenario 2: Manual review first**
```
1. Version is SUBMITTED
2. Tech lead reviews:
   → Transitions SUBMITTED → IN_REVIEW
   → Version is IN_REVIEW

3. Engineer approves:
   → Creates approval (technical)
   → Policy NOT satisfied
   → Version stays IN_REVIEW

4. Architect approves:
   → Creates approval (architect)
   → Policy satisfied
   → Transitions IN_REVIEW → APPROVED
   → Version is APPROVED
```

---

### 4. Reject Version

**Endpoint**: `POST /versions/{id}/reject`

**Authorization**: Users with approval roles (admin, architect, engineer)

**State Transitions**:
- If `SUBMITTED`: Auto-transitions to `IN_REVIEW`, then to `REJECTED`
- If `IN_REVIEW`: Transitions to `REJECTED`

**Request**:
```json
{
  "reason": "Needs more detail in the authentication section"
}
```

**Response**: `204 No Content`

**Behavior**:
1. If version in `SUBMITTED`, auto-transitions to `IN_REVIEW` first
2. Then transitions to `REJECTED` (terminal state)
3. New version must be created to resubmit content

**Events Published**:
- `VersionInReview` (if auto-transition occurs)
- `VersionRejected`

**Errors**:
- `412 Precondition Failed`: Version not in SUBMITTED or IN_REVIEW
- `403 Forbidden`: User has no approval roles

---

### 5. Publish Version

**Endpoint**: `POST /versions/{id}/publish`

**Authorization**: Product role only

**State Transition**: `APPROVED → PUBLISHED`

**Request**:
```json
{
  "destination": "production"
}
```

**Response**: `204 No Content`

**Requirements**:
- Version must be in `APPROVED` state
- User must have product role

---

## Workflow Examples

### Complete Happy Path

```
1. Engineer creates version → DRAFT
2. Engineer submits version → SUBMITTED
3. Engineer approves:
   → Auto-transition → IN_REVIEW
   → Approval created (technical)
   → Waiting for architect
4. Architect approves:
   → Approval created (architect)
   → Policy satisfied
   → Transition → APPROVED
5. Product manager publishes:
   → Transition → PUBLISHED
```

### With Manual Review

```
1. Engineer creates version → DRAFT
2. Engineer submits version → SUBMITTED
3. Tech lead reviews → IN_REVIEW
4. Engineer approves:
   → Approval created (technical)
   → Waiting for architect
5. Architect approves:
   → Approval created (architect)
   → Policy satisfied
   → Transition → APPROVED
6. Product manager publishes → PUBLISHED
```

### Rejection Flow

```
1. Engineer creates version → DRAFT
2. Engineer submits version → SUBMITTED
3. Architect rejects:
   → Auto-transition → IN_REVIEW
   → Transition → REJECTED
4. Create new version to retry
```

---

## Concurrent Approvals

The system uses **pessimistic locking** to handle concurrent approvals safely.

**Example**: Engineer and Architect approve simultaneously

```
Thread 1 (Engineer):
  🔒 Acquires lock on version
  ✅ Reads version (SUBMITTED)
  ✅ Auto-transitions to IN_REVIEW
  ➕ Creates approval (technical)
  ❌ Policy not satisfied (need architect)
  🔓 Releases lock

Thread 2 (Architect):
  ⏳ Waits for lock
  🔒 Acquires lock
  ✅ Reads version (IN_REVIEW) - fresh data!
  ➕ Creates approval (architect)
  ✅ Policy satisfied (has both)
  ✅ Transitions to APPROVED
  🔓 Releases lock

Result: ✅ Version is APPROVED with both approvals
```

**No race conditions**: Approvals are processed sequentially

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "invalid request"
}
```

### 403 Forbidden
```json
{
  "error": "user has no approval roles"
}
```

```json
{
  "error": "author cannot approve own version"
}
```

```json
{
  "error": "user has no approval roles and cannot review versions"
}
```

### 404 Not Found
```json
{
  "error": "version not found"
}
```

### 409 Conflict
```json
{
  "error": "user has already approved with a different role"
}
```

### 412 Precondition Failed
```json
{
  "error": "version must be SUBMITTED or IN_REVIEW to approve"
}
```

```json
{
  "error": "version must be SUBMITTED to manually transition to IN_REVIEW"
}
```

---

## Best Practices

### When to Use Manual Review

**Use `/review` when**:
- You want explicit audit trail of who initiated review
- You want to trigger notifications before approvals start
- Team lead wants to signal "ready" without approving

**Skip `/review` when**:
- First approver can just start approving
- Auto-transition is fine for your workflow
- Simpler is better

### Approval Strategy

**Sequential approvals** (recommended):
```
1. Submit
2. Technical reviewer approves
3. Architect approves
4. Publish
```

**With manual review**:
```
1. Submit
2. Team lead reviews
3. Approvers approve in any order
4. Publish
```

### Error Handling

Always check for:
- `412 Precondition Failed`: Wrong state
- `403 Forbidden`: Wrong role
- `409 Conflict`: Already approved

---

## Testing Checklist

### Manual Testing

- [ ] Submit version (DRAFT → SUBMITTED)
- [ ] Approve from SUBMITTED (auto-transitions to IN_REVIEW)
- [ ] Approve from IN_REVIEW (adds approval)
- [ ] Two approvals trigger APPROVED transition
- [ ] Reject from SUBMITTED (auto-transitions through IN_REVIEW to REJECTED)
- [ ] Reject from IN_REVIEW (transitions to REJECTED)
- [ ] Manual review (SUBMITTED → IN_REVIEW)
- [ ] Publish from APPROVED
- [ ] Error: Approve from DRAFT (should fail)
- [ ] Error: Product user tries to approve (should fail)
- [ ] Error: Author approves own version (should fail)
- [ ] Error: Same user approves twice (should fail)

### Concurrent Testing

- [ ] Two users approve simultaneously (both should succeed, version APPROVED)
- [ ] User approves while another reviews (should handle gracefully)

---

## Migration Notes

### Changes from Previous Version

**Removed**:
- ❌ `role` field from approve request body
- ❌ `role` field from reject request body

**Added**:
- ✅ Auto-transition to IN_REVIEW on first approval/rejection
- ✅ `/review` endpoint for manual state transition
- ✅ Pessimistic locking for concurrent safety
- ✅ State validation (must be SUBMITTED or IN_REVIEW)

**Behavioral Changes**:
- Approval role now determined by server (from user's role)
- First approval automatically moves version to IN_REVIEW
- Rejections also auto-transition to IN_REVIEW first

### API Compatibility

- ✅ Endpoint URLs unchanged
- ✅ HTTP methods unchanged
- ⚠️ Request body changed (removed `role` field)
- ✅ Response codes unchanged
- ✅ Status transitions more automated

---

## Quick Reference

| State | Allowed Actions |
|-------|----------------|
| DRAFT | Submit |
| SUBMITTED | Review (optional), Approve, Reject |
| IN_REVIEW | Approve, Reject |
| APPROVED | Publish |
| REJECTED | None (terminal) |
| PUBLISHED | None (terminal) |

| Role | Submit | Review | Approve | Reject | Publish |
|------|--------|--------|---------|--------|---------|
| admin | ✅ | ✅ | ✅ | ✅ | ❌ |
| architect | ✅ | ✅ | ✅ | ✅ | ❌ |
| engineer | ✅ | ✅ | ✅ | ✅ | ❌ |
| product | ✅ | ❌ | ❌ | ❌ | ✅ |
| viewer | ❌ | ❌ | ❌ | ❌ | ❌ |
