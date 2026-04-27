---
name: go-project-bootstrap
description: "Bootstrap a new Go service that fully follows the project architecture.  This includes: - Full repository structure - Minimal main.go - Modular bootstrap (app + wire files) - Platform setup (config, logger, DB, events, server) - Transaction manager - Event bus - One example feature implemented end-to-end  Follow ARCHITECTURE.md strictly: - Application is the primary boundary - Platform handles runtime concerns - Ports define dependencies - Adapters implement ports - Events are internal - Errors use err2 - DB access is transaction-safe  Always output: 1. Repo structure 2. Bootstrap code 3. Platform setup 4. Example feature 5. Config examples"
---

# Skill: go-project-bootstrap

## Purpose
Use this skill to bootstrap a new Go service that follows the project architecture from day one.

## Source of truth
Follow `ARCHITECTURE.md` strictly.

## Use when
- Creating a brand new Go service
- Rebuilding a starter template
- Creating a clean architecture baseline before feature work starts

## Required architecture constraints
- Minimal `cmd/.../main.go`
- Modular bootstrap in:
  - `internal/bootstrap/app.go`
  - `internal/bootstrap/wire_infra.go`
  - `internal/bootstrap/wire_events.go`
  - `internal/bootstrap/wire_services.go`
  - `internal/bootstrap/wire_http.go`
- Application layer structure:
  - `contract.go`
  - `inputs.go`
  - `outputs.go`
  - `service.go`
- Outbound ports in `internal/ports/outbound/`
- Inbound adapters in `internal/adapters/inbound/...`
- Outbound adapters in `internal/adapters/outbound/...`
- Runtime and platform in `internal/platform/...`
- Internal events in `internal/events/...`
- Shared errors in `pkg/errors/...`
- Config must support:
  - profiles (`local`, `dev`, `prod` at minimum)
  - driver-based selection (`http` or `grpc`, `postgres` or `dynamo`)
- Errors use `err2`
- Postgres access uses a shared transaction-safe pool wrapper

## Bootstrap workflow
1. Generate the repository structure
2. Create minimal `main.go`
3. Create platform config and profile loading
4. Create logger setup
5. Create DB and runtime client setup
6. Create transaction manager
7. Create internal event bus runtime
8. Create bootstrap wiring modules
9. Add one example feature end to end
10. Add HTTP server and route registration
11. Add starter tests or test stubs

## Output expectations
When using this skill, produce:
1. Full repository tree
2. Starter code for bootstrap and platform setup
3. Starter code for one example feature
4. Example config and env shape
5. Example event and error setup
6. Example transaction-aware repository wiring

## Do
- Start with the architecture, not the endpoints
- Make the first aggregate a full vertical slice
- Keep boundaries visible in the file tree
- Use explicit mapping and explicit application contracts

## Do not
- Hide all startup logic in one file
- Start by writing handlers only
- Put config, logging, or DB setup in `main.go`
- Skip platform abstractions for transactions and events
