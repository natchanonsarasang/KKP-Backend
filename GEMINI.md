# GEMINI.md - Project Instructions for Callecto API

This file guides Gemini/Antigravity and other agents when working within this codebase. It documents the core architecture, tooling, coding standards, and system behaviors.

---

## 🚀 Quick Reference Commands

Always run these commands from the project root:

```powershell
# Install/sync dependencies
go mod tidy

# Run the API locally
go run main.go

# Hot-reload development server (requires Air)
air

# Run all unit tests
go test ./...

# Run a single test
go test ./src/services/ -run TestProcessSession_Success

# Seed a mock debtor, call list item, and call session
go run ./cmd/seed
```

---

## 🛠️ Technology Stack & Environment

- **Backend Framework**: Go Fiber v2
- **Database**: MongoDB (using Go Driver)
- **Authentication**: Supabase JWT (keys validated via `JWK_SET_URL` or fallback `JWT_SECRET_KEY`)
- **Integration**: Botnoi Outbound Voicebot API
- **Tooling**: Air (Hot reload), Docker/Podman

### Required `.env` Configuration
The following variables are loaded via `godotenv` from the project root:
- `MONGODB_URI` - MongoDB Atlas connection string.
- `MONGODB_NAME` / `DATABASE_NAME` - Database name.
- `PORT` - Service port (default `8080`).
- `JWK_SET_URL` - Supabase JWKS endpoint for decoding signatures.
- `OUTBOUND_URL` - Botnoi Voicebot dialer endpoint.
- `OUTBOUND_ACCESS_TOKEN` - Secret token to authenticate with Botnoi.

> [!WARNING]
> `.env` is currently committed to git containing active/fallback credentials. Never commit new production secrets to it. Treat existing credentials as compromised.

---

## 🏛️ Code Architecture & Directory Structure

The project follows a hand-wired **Clean Architecture** (Request flow: `Gateways` ➔ `Services` ➔ `Repositories` ➔ `MongoDB`).

```
.
├── cmd/                  # Seeding scripts & standalone commands
├── configuration/        # Bootstrapping (Fiber config, clients)
├── domain/             
│   ├── datasources/      # MongoDB connection pooling & APM integration
│   ├── entities/         # Shared data models (*Model / BSON tags)
│   └── repositories/     # Database queries / mutations (I*Repository)
├── src/                
│   ├── gateways/         # HTTP route handlers (HTTPGateway / REST routes)
│   ├── middlewares/      # Fiber logging & JWT validation
│   └── services/         # Core business logic (campaigns, webhooks, I*Service)
└── docs/                 # Swagger documents, manuals, and schemas
```

### Dependency Injection
There is **no DI container** (like FX or Wire). All components are manually instantiated and threaded together:
1. Repositories are initialized in `main.go`.
2. Services are initialized in `main.go`, taking repositories as constructor arguments.
3. The `HTTPGateway` is initialized in `src/gateways/http.go` (the central wiring point).
4. Routes are registered on Fiber router groups in `src/gateways/route.go`.

---

## 📐 Coding Conventions & Guidelines

### 1. Identity & Database Identifiers
- Do not use MongoDB's native `_id` (`ObjectID`) for application-level entity queries.
- All primary keys must be **app-generated UUID strings** in an `id` field.
- Repositories should query using `bson.M{"id": id}`.
- Services generate `uuid.NewString()` and assign `CreatedAt`/`UpdatedAt` when creating new entities.

### 2. Multi-Tenancy & Security Boundaries
- Data is strictly partitioned by `workspace_id` and `user_id`.
- Handlers extract the `user_id` from JWT context and `workspace_id` from parameters/queries.
- All service methods checking/altering data must enforce ownership filters (e.g. `*ByUser` or `*ByWorkspaceByUser` methods).

### 3. Response Formats
- Successful responses must wrap data in `entities.ResponseModel{Message, Data}`.
- Error responses must return `entities.ResponseMessage{Message}` with the appropriate HTTP status code.

### 4. Background Concurrency (`ProcessSession`)
- **Asynchronous Campaign Loops**: Triggering a session (`start` or `continue`) initiates a background Goroutine:
  ```go
  go func() {
      _ = h.CallProcessService.ProcessSession(body.SessionID)
  }()
  ```
  This returns `200 OK` immediately to the API client, avoiding gateway timeouts.
- **Concurrency Control**: Loops fetch pending `call_list_items`, verify business hours, check limits (`Settings.ConcurrentCalls`), and trigger outbound dials concurrently using WaitGroups.

### 5. Webhook Ingestion & Thai Language Helpers
- Inbound Botnoi callbacks go to `POST /api/v1/webhooks/botnoi` (JWT authentication is bypassed).
- Webhook parses call duration, AMD status (`HUMAN` vs `MACHINE`), and maps conversation outcomes.
- Keep Thai text generation helpers (Buddhist calendar conversion, numbers to text) intact. Do not change Thai prompts without explicit instruction.

### 6. Writing Tests
- Unit tests live alongside production code as `*_test.go` and use `testify` assertions.
- Interfaces are mocked. If you modify or add methods to a repository/service interface, **you must update the corresponding mocks** or the test suite will fail compilation.

---

## 🧠 Active Agent Skills & Workflows

### 1. Codebase Design (`codebase-design`)
- **When**: Designing/improving module interfaces, separating concerns, deciding on seams, or refactoring.
- **Rules**:
  - Focus on **Deep Modules** (small interface hiding high implementation complexity). Avoid pass-through or shallow modules.
  - Establish clear **Seams** (locations where interfaces live) to decouple modules. Do not add a seam unless there are at least two distinct implementations/adapters (avoid hypothetical abstraction).
  - Test at the interface level (external seam). Do not write tests that bypass the interface.
  - Accept dependencies (DI) instead of instantiating them inside the module. Return results rather than producing side-effects.

### 2. Diagnosing Bugs (`diagnosing-bugs`)
- **When**: Debugging errors, crashes, failing tests, or performance regressions.
- **Rules**:
  - **Phase 1: Build a Feedback Loop**: Do NOT read code or hypothesize before establishing a fast (<2s), deterministic, agent-runnable command (test, curl, script) that fails on this specific bug. If impossible, halt and request info.
  - **Phase 2: Minimize**: Cut down inputs/config until only the absolute load-bearing elements of the failure remain.
  - **Phase 3: Hypothesize**: Rank 3–5 falsifiable hypotheses (`"If X is the cause, then changing Y makes it disappear"`). Share with the user.
  - **Phase 4: Instrument**: Target logs with prefix `[DEBUG-xxxx]` (remove them during final cleanup).
  - **Phase 5: Fix & Regression Test**: Write a regression test at the correct seam *before* fixing, then verify both test and loop pass.
  - **Phase 6: Cleanup & Post-Mortem**: Clean up all instrumentation, delete throwaways, and record the solution.

### 3. Domain Modeling (`domain-modeling`)
- **When**: Aligning on terms/glossary, resolving business domain logic, or writing ADRs.
- **Rules**:
  - Challenge ambiguous terms (e.g. "Customer" vs "User") and align with the `CONTEXT.md` glossary.
  - Update `CONTEXT.md` immediately when new terms are resolved. Keep implementation details out of `CONTEXT.md`.
  - Create ADRs (Architecture Decision Records) sparingly. Only write them if the decision is: (1) Hard to reverse, (2) Surprising without context, and (3) A trade-off with clear alternatives.



