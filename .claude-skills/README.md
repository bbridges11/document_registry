# Claude Skills for Document Registry

This directory contains AI assistant skills specifically designed for working with the Document Registry codebase. These skills help maintain architectural consistency and accelerate development.

## 📁 Available Skills

### 🏗️ Architecture & Development

#### **go-project-bootstrap.md**
Bootstrap a new Go service following the project architecture.

**Use when:**
- Starting a new microservice
- Creating a service similar to document-registry
- Need complete project structure from scratch

**Creates:**
- Full repository structure
- Minimal main.go
- Modular bootstrap
- Platform setup (config, logger, DB, events, server)
- Transaction manager
- Event bus
- One example feature end-to-end

**Example prompt:**
```
Using the go-project-bootstrap skill, create a new service called "user-profile-service" 
that manages user profiles with CRUD operations.
```

---

#### **go-feature-implementation.md**
Implement a new feature inside the existing architecture without breaking boundaries.

**Use when:**
- Adding a new feature to existing service
- Need to follow architecture strictly
- Want feature mapped before coding

**Process:**
1. Maps feature into layers (domain, application, adapters, ports)
2. Identifies required files (create/update)
3. Implements with full code
4. Updates bootstrap wiring

**Example prompt:**
```
Using the go-feature-implementation skill, add a "comment" feature that allows 
users to add comments to document versions. Comments should have: text, author, 
timestamp, and support replies.
```

---

#### **go-aggregate-scaffold.md**
Scaffold a new aggregate/feature end-to-end following the project architecture.

**Use when:**
- Adding a complete new aggregate (like adding "Template" alongside "Document")
- Need domain entity + full application layer + adapters

**Creates:**
- Domain entity and value objects
- Application layer (contract, inputs, outputs, service)
- Outbound ports (if needed)
- Inbound HTTP adapter (handler, requests, responses, mapper)
- Event definitions (if aggregate emits events)
- Bootstrap wiring

**Example prompt:**
```
Using the go-aggregate-scaffold skill, create a "Template" aggregate that stores 
reusable document templates. Templates should have: name, category, content, 
and can be published/unpublished.
```

---

#### **go-architecture-refactor.md**
Refactor existing Go code to conform to the project architecture without changing behavior.

**Use when:**
- Code violates architecture boundaries
- Business logic in handlers
- Missing application contracts
- Need to clean up violations

**Process:**
1. Identifies violations
2. Creates refactor plan
3. Shows updated structure
4. Provides updated code
5. Explains changes

**Example prompt:**
```
Using the go-architecture-refactor skill, refactor the user handler that 
currently has business logic inline. Move validation and business rules 
to the application layer.
```

---

### 🔧 Utilities

#### **generate-postman-collection.md**
Generate a Postman collection from Echo routes registered in `RegisterRoutes` functions.

**Use when:**
- Need API documentation
- Want to generate collection from code
- Routes changed and collection needs update

**Scans:**
- `internal/adapters/**/handler.go`
- Looks for `RegisterRoutes(e *echo.Echo)` functions
- Extracts all route registrations

**Example prompt:**
```
Using the generate-postman-collection skill, create a Postman collection 
for all endpoints in the document-registry service.
```

---

## 🎯 How to Use These Skills

### With Claude.ai

1. Upload the relevant skill file to your conversation
2. Reference it in your prompt:
   ```
   Using the go-feature-implementation skill, add a rating feature...
   ```

### With GitHub Copilot Chat

1. Include the skill content in your prompt:
   ```
   @workspace Using the following skill guide, implement a new feature:
   
   [paste skill content]
   
   Feature: Add tagging system to documents...
   ```

### With Cursor

1. Add skills to your `.cursorrules` or reference them:
   ```
   @docs go-feature-implementation.md
   
   Add a notification feature that sends emails when versions are approved.
   ```

### With Other AI Assistants

1. Copy the relevant skill content
2. Paste it into your conversation context
3. Reference it in your implementation request

---

## 📋 Skill Selection Guide

**Choose the right skill for your task:**

| Task | Skill |
|------|-------|
| Create new service from scratch | `go-project-bootstrap.md` |
| Add feature to existing service | `go-feature-implementation.md` |
| Add new aggregate/entity | `go-aggregate-scaffold.md` |
| Fix architecture violations | `go-architecture-refactor.md` |
| Generate API collection | `generate-postman-collection.md` |

---

## 🏛️ Architecture Principles

All skills enforce these principles from [ARCHITECTURE.md](../ARCHITECTURE.md):

### Layer Structure
- `cmd/` - Application entrypoints
- `internal/bootstrap/` - Dependency wiring
- `internal/application/` - Use cases and contracts
- `internal/domain/` - Entities and business logic
- `internal/ports/outbound/` - Dependency interfaces
- `internal/adapters/inbound/` - HTTP/gRPC handlers
- `internal/adapters/outbound/` - Repository implementations
- `internal/platform/` - Technical infrastructure
- `internal/events/` - Event contracts
- `pkg/errors/` - Shared errors

### Dependency Direction
```
inbound adapter → application → outbound port → outbound adapter → platform
```

### Key Patterns
- **CQRS** - Commands and queries separated
- **Hexagonal** - Ports and adapters
- **Application Contracts** - Explicit inputs/outputs
- **Transaction Safety** - err2-based error handling
- **Event-Driven** - Internal domain events
- **Minimal main.go** - Bootstrap only

---

## 🎓 Workflow Examples

### Example 1: Add New Feature

```
1. Review ARCHITECTURE.md
2. Choose: go-feature-implementation.md
3. Prompt:
   "Using go-feature-implementation skill, add a 'bookmark' feature 
    that lets users bookmark documents for quick access."
4. AI will:
   - Map to architecture layers
   - List files to create/update
   - Implement full feature
   - Update bootstrap wiring
```

### Example 2: Refactor Violations

```
1. Identify code with violations
2. Choose: go-architecture-refactor.md
3. Prompt:
   "Using go-architecture-refactor skill, refactor the approval handler 
    which has validation logic in the HTTP layer."
4. AI will:
   - Identify violations
   - Create refactor plan
   - Move logic to application layer
   - Maintain behavior
```

### Example 3: Bootstrap New Service

```
1. Plan your service
2. Choose: go-project-bootstrap.md
3. Prompt:
   "Using go-project-bootstrap skill, create 'notification-service' 
    with email and SMS notification features."
4. AI will:
   - Create complete structure
   - Add example feature
   - Wire dependencies
   - Setup platform
```

---

## 📖 Reading the Skills

Each skill follows this structure:

1. **Description** - What the skill does
2. **When to use** - Triggers and use cases
3. **Process** - Step-by-step methodology
4. **Output** - What you'll get
5. **Examples** - Sample prompts and outputs

**Pro tip:** Read the skill file before using it to understand what it will do.

---

## 🔄 Keeping Skills Updated

These skills are snapshots from the project's current architecture. If the architecture changes significantly:

1. Update `ARCHITECTURE.md`
2. Regenerate skills or update manually
3. Test with sample prompts
4. Commit updated skills

---

## 🤝 Contributing New Skills

To add a new skill:

1. Create skill file: `.claude-skills/my-new-skill.md`
2. Follow existing skill format
3. Include clear examples
4. Test with AI assistant
5. Update this README
6. Commit to repository

---

## ⚠️ Important Notes

### Architecture Compliance

These skills enforce the architecture in `ARCHITECTURE.md`. If you need to deviate:
1. Update `ARCHITECTURE.md` first
2. Then update relevant skills
3. Document the change

### Code Quality

Skills generate code that:
- ✅ Follows Go best practices
- ✅ Uses err2 for error handling
- ✅ Includes transaction management
- ✅ Respects layer boundaries
- ✅ Has explicit mapping
- ✅ Implements proper events

### Not a Replacement

Skills help with:
- Structure and scaffolding
- Architectural consistency
- Boilerplate reduction

You still need to:
- Review generated code
- Add business logic specifics
- Write tests
- Handle edge cases
- Add domain validations

---

## 🔗 Related Documentation

- [ARCHITECTURE.md](../ARCHITECTURE.md) - Complete architecture guide
- [README.md](../README.md) - Project overview
- [QUICKSTART.md](../QUICKSTART.md) - Getting started
- [MIGRATIONS.md](../MIGRATIONS.md) - Database migrations

---

## 💡 Tips

1. **Start with architecture** - Read `ARCHITECTURE.md` before using skills
2. **Be specific** - Clear prompts get better results
3. **Review output** - Always review and understand generated code
4. **Combine skills** - Use bootstrap → feature → refactor as needed
5. **Iterate** - Refine prompts based on output
6. **Test thoroughly** - Run tests after using skills

---

## 📚 Examples

### Quick Feature Add

```bash
# Prompt
"Using go-feature-implementation skill, add version comparison feature 
that shows diff between two versions of a document"

# AI generates:
- Domain: VersionComparison value object
- Application: CompareVersions use case
- Handler: POST /versions/compare endpoint
- Tests: Unit and integration tests
- Wiring: Bootstrap updates
```

### Full Aggregate

```bash
# Prompt
"Using go-aggregate-scaffold skill, create Notification aggregate with:
- Types: email, sms, push
- Status: pending, sent, failed
- Retry logic for failures"

# AI generates:
- Domain: Notification entity + NotificationType VO
- Application: Send, Retry, ListNotifications
- Ports: NotificationSender interface
- Adapters: Email, SMS, Push adapters
- Events: NotificationSent, NotificationFailed
- Full wiring
```

### Architecture Cleanup

```bash
# Prompt
"Using go-architecture-refactor skill, fix these violations:
- Document handler has validation logic
- Repository returns domain entities to HTTP layer
- No transaction management in create flow"

# AI generates:
- Refactor plan
- Moved validation to application
- Added input/output models
- Wrapped in transaction
- Updated all layers
```

---

## 🎉 Success Stories

These skills help maintain:
- ✅ **Consistent architecture** - Every feature follows the same pattern
- ✅ **Faster development** - Scaffolding saves hours
- ✅ **Fewer bugs** - Architecture prevents common mistakes
- ✅ **Easier onboarding** - New developers follow established patterns
- ✅ **Better reviews** - Clear structure makes PRs easier to review

---

Happy building! 🚀
