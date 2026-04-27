---
name: go-feature-implementation
description: "Implement a new feature inside the existing architecture without breaking boundaries.  Before coding, map the feature into: - Domain changes - Application contract (inputs/outputs/usecase) - Outbound ports - Adapters (inbound + outbound) - Events - Errors - Bootstrap wiring  Follow ARCHITECTURE.md strictly: - Application layer is the entry point - No domain leakage into adapters - Explicit mapping (requests/responses/mapper) - Use err2 for error handling - Use TransactionManager for DB logic - Publish events only after persistence succeeds  Always output: 1. Feature mapping 2. Files to create/update 3. Implementation 4. Wiring updates"
---

# Skill: go-feature-implementation

## Purpose
Use this skill to implement a feature inside an existing Go service while preserving the architecture.

## Source of truth
Follow `ARCHITECTURE.md` strictly.

## Use when
- Implementing a new endpoint or use case
- Extending an aggregate
- Adding repository, storage, notification, or event behavior
- Adding feature logic to an existing bounded context

## Required architecture constraints
- Application use cases are the core inbound API
- All methods accept `context.Context`
- Inputs and outputs are explicit structs
- Domain entities are not returned from application
- Outbound dependencies are expressed as ports in `internal/ports/outbound/`
- Inbound adapters map transport and application models explicitly
- Handler shape remains `type Handler struct { service UseCase }`
- Internal events remain in `internal/events/`
- Shared errors remain in `pkg/errors/`
- Error handling uses `err2`
- Transactions use `TransactionManager`

## Implementation workflow
Before writing code, always map the feature into:
1. Domain changes
2. Application contract changes
3. Application input and output changes
4. Outbound port changes
5. Outbound adapter changes
6. Inbound HTTP or gRPC changes
7. Event changes
8. Error changes
9. Bootstrap wiring changes

Then implement in that order.

## Output expectations
When using this skill, produce:
1. Feature mapping summary
2. Files to create or update
3. The code changes by file
4. Event and error additions if needed
5. Bootstrap wiring updates
6. A short explanation of transaction and side-effect boundaries

## Do
- Keep handlers thin
- Keep mapping explicit
- Add events only where the domain or application needs them
- Publish events after persistence succeeds
- Keep transport concerns out of application

## Do not
- Put validation or business logic in handlers beyond transport-shape validation
- Return domain entities from application
- Put SDK or runtime code into domain or application
- Add side effects inside transactions unless explicitly designed for it