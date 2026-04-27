# Quick Reference - Claude Skills

## 🎯 Skill Selection (One-Liner)

```
New service?          → go-project-bootstrap.md
New feature?          → go-feature-implementation.md
New aggregate?        → go-aggregate-scaffold.md
Fix violations?       → go-architecture-refactor.md
Generate API docs?    → generate-postman-collection.md
```

## ⚡ Quick Prompts

### Bootstrap New Service
```
Using go-project-bootstrap skill, create "[SERVICE_NAME]" with [FEATURES].
```

### Add Feature
```
Using go-feature-implementation skill, add [FEATURE] that does [WHAT].
```

### Scaffold Aggregate
```
Using go-aggregate-scaffold skill, create "[AGGREGATE]" aggregate with [PROPERTIES].
```

### Refactor Code
```
Using go-architecture-refactor skill, fix [FILE/COMPONENT] which has [VIOLATIONS].
```

### Generate Collection
```
Using generate-postman-collection skill, create collection for [SERVICE].
```

## 📐 Architecture Layers (Reference)

```
cmd/                    → main.go only
internal/
  bootstrap/            → Dependency wiring
  application/          → Use cases (commands/queries)
  domain/              → Entities & business logic
  ports/outbound/      → Dependency interfaces
  adapters/
    inbound/           → HTTP handlers
    outbound/          → Repositories, storage
  platform/            → Config, DB, AWS clients
  events/              → Event contracts
pkg/errors/            → Shared errors
```

## 🔄 Request Flow

```
HTTP → Handler → Mapper → Application Input → Use Case → Domain + Repos
                                                            ↓
HTTP ← Response ← Mapper ← Application Output ← Result ← Domain + Repos
```

## ✅ File Naming Conventions

```
Application:
  contract.go          → UseCase interface
  inputs.go            → Command/query inputs
  outputs.go           → Result models
  service.go           → Implementation

HTTP Adapter:
  handler.go           → HTTP handlers
  requests.go          → Request DTOs
  responses.go         → Response DTOs
  mapper.go            → Input/output mapping
```

## 🎨 Example Prompts

**Add Comments:**
```
Using go-feature-implementation skill, add comments to versions. 
Comments have: text, author, created_at, parent_comment_id (for replies).
```

**Add Templates:**
```
Using go-aggregate-scaffold skill, create Template aggregate.
Templates have: name, category, content, is_public, created_by.
```

**Fix Handler:**
```
Using go-architecture-refactor skill, refactor DocumentHandler which 
has validation and business logic in HTTP layer.
```

**New Service:**
```
Using go-project-bootstrap skill, create "analytics-service" that 
tracks document views and generates usage reports.
```

## 🚫 Common Mistakes

❌ **Don't:**
- Put business logic in handlers
- Return domain entities from application
- Skip transaction management
- Forget to wire in bootstrap
- Mix concerns across layers

✅ **Do:**
- Use application inputs/outputs
- Wrap writes in transactions
- Emit events after persistence
- Map explicitly (no inline mapping)
- Follow dependency direction

## 📊 Skill Outputs

### go-project-bootstrap
- Complete repo structure
- Working example feature
- All platform setup
- Ready to run

### go-feature-implementation
- Feature mapping
- Files to create/update
- Full implementation
- Bootstrap wiring

### go-aggregate-scaffold
- Domain entities
- Application layer
- HTTP handlers
- Event definitions
- Full wiring

### go-architecture-refactor
- Violation analysis
- Refactor plan
- Updated code
- Migration guide

### generate-postman-collection
- Complete Postman JSON
- All endpoints
- Request examples
- Environment setup

## 💾 Save for Quick Access

Bookmark this file or keep open when working with AI assistants!
