---
name: go-architecture-refactor
description: "Refactor existing Go code to conform to the project architecture without changing behavior.  Focus on: - Moving business logic into application layer - Enforcing application contracts (inputs/outputs/usecase) - Separating ports, adapters, and platform concerns - Cleaning up handlers (thin + mapping only) - Normalizing bootstrap into modular wiring - Enforcing err2 error handling and transaction manager usage  Follow ARCHITECTURE.md strictly.  Always output: 1. Violations found 2. Refactor plan 3. Updated structure 4. Updated code 5. Explanation of changes"
---

# Skill: go-architecture-refactor

## Purpose
Use this skill to refactor existing Go code so it conforms to the project architecture without changing intended behavior.

## Source of truth
Follow `ARCHITECTURE.md` strictly.

## Use when
- A feature was added without respecting the architecture
- Handlers contain business logic
- Ports and adapters are mixed together
- Domain types leak into transport or application outputs
- Bootstrap is monolithic or unclear
- Error handling and transaction boundaries are inconsistent

## Required architecture constraints
- `main.go` must remain minimal
- Bootstrap must be modular:
  - `app.go`
  - `wire_infra.go`
  - `wire_events.go`
  - `wire_services.go`
  - `wire_http.go`
- Application layer is the primary inbound boundary
- Use case files:
  - `contract.go`
  - `inputs.go`
  - `outputs.go`
  - `service.go`
- Only outbound ports live in `internal/ports/outbound/`
- HTTP handlers must use:
  - `requests.go`
  - `responses.go`
  - `mapper.go`
- Internal events live in `internal/events/`
- Shared errors live in `pkg/errors/`
- `err2` is the standard error-handling approach
- Transactions must use `TransactionManager`

## Refactor procedure
1. Audit the current code against the architecture
2. List every violation by layer:
   - domain
   - application
   - adapters
   - ports
   - platform
   - bootstrap
3. Propose the target file structure
4. Move business logic into application services
5. Introduce or clean up outbound ports
6. Split handler request and response mapping into dedicated files
7. Move runtime concerns into `internal/platform/`
8. Normalize event handling into `internal/events/`
9. Normalize error handling with `err2`
10. Normalize transaction handling with the shared transaction manager
11. Update bootstrap wiring
12. Preserve behavior while improving boundaries

## Output expectations
When using this skill, produce:
1. Architecture violations found
2. Refactor plan
3. New target file tree
4. Updated code by file
5. Notes on behavior-preserving decisions
6. Any migration notes for renamed or moved packages

## Do
- Preserve external behavior unless explicitly told otherwise
- Refactor incrementally and clearly
- Explain architectural reasoning
- Keep explicit mappers and thin handlers

## Do not
- Sneak in unrelated redesigns
- Change public API shape unless requested
- Leave half-moved boundaries
- Keep duplicate abstractions once the refactor is complete
