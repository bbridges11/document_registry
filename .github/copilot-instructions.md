# GitHub Copilot Instructions - Document Registry

## Project Overview

Document Registry is a production-ready document management and version control system following clean architecture principles with hexagonal architecture, CQRS, and DDD-inspired boundaries.

## Core Architecture Principles

### 1. Layer Boundaries (Strict)

```
cmd/                    - Entrypoint only
internal/bootstrap/     - Dependency wiring
internal/application/   - Use cases (pure business logic)
internal/domain/        - Entities and business rules
internal/ports/         - Dependency interfaces
internal/adapters/      - Implementations (HTTP, repos, storage)
internal/platform/      - Infrastructure runtime
internal/events/        - Event contracts
pkg/errors/            - Shared errors
```

### 2. Dependency Direction

```
Handler → Application → Domain + Ports → Adapters → Platform
```

**Never reverse this flow.**

### 3. Application Layer Rules

- Primary inbound boundary
- Defines `UseCase` interfaces
- Uses explicit input/output models
- NO domain entities in/out
- NO infrastructure concerns

**File structure:**
```
internal/application/<entity>/
  contract.go   - UseCase interface
  inputs.go     - Command/query inputs
  outputs.go    - Result models
  service.go    - Implementation
```

### 4. Handler Rules

- Parse request → validate shape
- Map to application input
- Call use case
- Map output to response
- NO business logic here

**File structure:**
```
internal/adapters/inbound/http/<entity>/
  handler.go    - HTTP handlers
  requests.go   - Request DTOs
  responses.go  - Response DTOs
  mapper.go     - Mapping functions
```

### 5. Transaction Management

```go
// Application layer decides when to transact
func (s *Service) Create(ctx context.Context, input Input) (Output, error) {
    defer err2.Handle(&err)
    
    var entity *domain.Entity
    
    // Wrap in transaction
    try.To(s.txMgr.WithinTransaction(ctx, func(ctx context.Context) error {
        entity = domain.New(...)
        return s.repo.Save(ctx, entity)
    }))
    
    // Publish events AFTER transaction succeeds
    for _, e := range entity.Events() {
        s.eventBus.Publish(ctx, e)
    }
    entity.ClearEvents()
    
    return Output{ID: entity.ID()}, nil
}
```

### 6. Error Handling (err2)

```go
func DoSomething(ctx context.Context) (result Result, err error) {
    defer err2.Handle(&err)  // ALWAYS at function start
    
    data := try.To1(FetchData(ctx))
    try.To(SaveData(ctx, data))
    
    return Result{Data: data}, nil
}
```

**Never use:**
- `if err != nil { return ..., err }`
- Mixed error styles

**Always use:**
- `defer err2.Handle(&err)` at start
- `try.To()` and `try.To1()` for calls

## Code Generation Guidelines

### When suggesting new features:

1. **Map to layers first**
   ```
   Domain: What entities/VOs needed?
   Application: What use cases?
   Ports: What dependencies?
   Adapters: What implementations?
   Events: What should be published?
   ```

2. **Generate in order**
   - Domain entities
   - Application contract
   - Application service
   - Ports (if new dependencies)
   - Adapters (HTTP, repository)
   - Bootstrap wiring

3. **Use established patterns**
   - Check existing features for reference
   - Follow naming conventions
   - Use same error handling style
   - Match testing patterns

### Naming Conventions

**Application:**
- Interface: `UseCase`, `CommandService`, `QueryService`
- Methods: `Create`, `Get`, `GetByID`, `List`, `Update`, `Delete`
- Inputs: `CreateInput`, `UpdateInput`, `GetQuery`
- Outputs: `CreateOutput`, `View`, `ListView`

**HTTP:**
- Handler: `Handler` struct with `RegisterRoutes(e *echo.Echo)`
- Requests: `CreateRequest`, `UpdateRequest`
- Responses: `Response`, `ListView`
- Mapper: `toInput()`, `toOutput()`, `toResponse()`

**Domain:**
- Entities: `Document`, `Version`, `Approval`
- Value Objects: `Status`, `DocumentType`, `Actor`
- Methods: Business operation verbs

### Common Patterns

**Handler pattern:**
```go
func (h *Handler) Create(c echo.Context) error {
    var req CreateRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, ErrorResponse{Error: "invalid"})
    }
    
    input := toCreateInput(req, getUserID(c))
    output, err := h.service.Create(c.Request().Context(), input)
    if err != nil {
        return handleError(c, err)
    }
    
    return c.JSON(201, toCreateResponse(output))
}
```

**Service pattern:**
```go
func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
    defer err2.Handle(&err)
    
    entity := domain.NewEntity(input.Name)
    try.To(s.repo.Save(ctx, entity))
    
    return CreateOutput{ID: entity.ID()}, nil
}
```

**Repository pattern:**
```go
type Repository interface {
    Save(ctx context.Context, entity *domain.Entity) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.Entity, error)
    // Use domain types, not DTOs
}
```

## Environment Variables

All use `DOCUMENT_REGISTRY_` prefix:
```go
DOCUMENT_REGISTRY_HTTP_PORT=8080
DOCUMENT_REGISTRY_POSTGRES_HOST=localhost
DOCUMENT_REGISTRY_S3_BUCKET=my-bucket
```

Exception: `AWS_REGION` (standard AWS SDK variable has fallback)

## Testing Patterns

```go
// Table-driven tests
func TestService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateInput
        wantErr bool
    }{
        {
            name: "success",
            input: CreateInput{Name: "test"},
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## Bootstrap Wiring

Always update bootstrap after adding features:

```go
// internal/bootstrap/wire_services.go
func wireServices(infra *infrastructure) *services {
    newService := newfeature.NewService(
        infra.NewRepo,
        infra.TransactionManager,
        infra.EventBus,
    )
    
    return &services{
        // ... existing
        NewFeature: newService,
    }
}

// internal/bootstrap/wire_http.go
func wireHTTP(services *services) *httpHandlers {
    return &httpHandlers{
        // ... existing
        NewFeature: newfeature.NewHandler(services.NewFeature),
    }
}
```

## Documentation

- Add comments to public types and interfaces
- Document business rules in domain layer
- Keep README updated with new features
- Update API docs when adding endpoints

## What NOT to Suggest

❌ Business logic in handlers
❌ Direct database access from handlers
❌ Domain entities as API responses
❌ Skipping transaction wrappers for writes
❌ Inline mapping in handlers
❌ Publishing events inside transactions
❌ Mixed error handling styles
❌ Forgetting bootstrap wiring

## What TO Suggest

✅ Clean separation of concerns
✅ Explicit input/output models
✅ Transaction-wrapped writes
✅ Event emission after persistence
✅ Consistent error handling with err2
✅ Proper layer boundaries
✅ Bootstrap wiring updates
✅ Tests for new features

## Quick References

- **Architecture:** See `ARCHITECTURE.md`
- **Skills:** See `.claude-skills/` directory
- **Examples:** Look at existing features in codebase
- **API:** See `bruno/` collection

## For Complex Features

Reference the skills in `.claude-skills/`:
- `go-feature-implementation.md` - Full feature workflow
- `go-aggregate-scaffold.md` - New aggregate scaffolding
- `go-architecture-refactor.md` - Refactoring guidance

These provide detailed step-by-step processes for maintaining architectural consistency.
