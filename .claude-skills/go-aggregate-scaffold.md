---
name: go-aggregate-scaffold
description: "Scaffold a new aggregate/feature end-to-end following the project architecture.  This includes: - Domain entity and value objects - Application layer (contract, inputs, outputs, service) - Outbound ports (if needed) - Inbound HTTP adapter (handler, requests, responses, mapper) - Event definitions (if aggregate emits events) - Bootstrap wiring  Follow ARCHITECTURE.md strictly: - Application is the primary inbound boundary - Use explicit input/output models - Keep handlers thin and mapping explicit - Do not return domain entities from use cases - Use EventEmitter for aggregates if emitting events - Use err2 for error handling - Use TransactionManager for DB operations  Always output: 1. Feature mapping 2. File structure 3. Full implementation by file 4. Wiring updates"
---

# Skill: go-aggregate-scaffold

## Purpose
Use this skill to scaffold a new aggregate or feature in a Go service that follows the project architecture.

## Source of truth
Follow `ARCHITECTURE.md` strictly.

## Use when
- Adding a new aggregate such as `customer`, `order`, `document`, or `attachment`
- Creating the first end-to-end slice for a new feature
- Scaffolding missing domain, application, and adapter layers for an entity

## Required architecture constraints
- Application is the primary inbound boundary
- Use cases live in `internal/application/<entity>/`
- Feature structure uses:
  - `contract.go`
  - `inputs.go`
  - `outputs.go`
  - `service.go`
- Outbound ports live in `internal/ports/outbound/`
- Inbound HTTP adapter lives in `internal/adapters/inbound/http/<entity>/`
- Transport mapping must be explicit in:
  - `requests.go`
  - `responses.go`
  - `mapper.go`
- Handler shape:
  - `type Handler struct { service UseCase }`
  - `func (h *Handler) RegisterRoutes(e *echo.Echo)`
- Domain entities must not depend on transport or platform
- If the aggregate emits events, it must implement:
  - `Events() []events.Event`
  - `ClearEvents()`
- Events are internal and live in `internal/events/`
- Shared typed errors live in `pkg/errors/`
- Errors must follow `github.com/lainio/err2`
- Transactions go through the shared `TransactionManager`

## Workflow
1. Identify the aggregate boundary and domain responsibilities
2. Create or update:
   - `internal/domain/<entity>/`
   - `internal/application/<entity>/`
   - `internal/adapters/inbound/http/<entity>/`
   - outbound adapters if needed
   - `internal/events/` types if the aggregate emits events
3. Define the application contract in `contract.go`
4. Define command and query inputs in `inputs.go`
5. Define outputs and views in `outputs.go`
6. Implement application logic in `service.go`
7. Add outbound port interfaces if new dependencies are required
8. Add HTTP request and response models plus explicit mappers
9. Add route registration in the entity handler
10. Wire the feature through bootstrap

## Output expectations
When using this skill, produce:
1. A feature mapping summary
2. The exact file tree for the new aggregate
3. The code for each new file
4. Any required bootstrap wiring updates
5. Any new event definitions
6. Any new tests or test stubs

## Do
- Keep handlers thin
- Keep mapping explicit
- Return application outputs, not domain entities
- Keep transaction boundaries in application
- Publish events only after successful persistence or commit

## Do not
- Put business logic in handlers
- Return domain entities from use cases
- Use transport DTOs in application
- Put ports in adapters or domain
- Publish side effects inside transactions
