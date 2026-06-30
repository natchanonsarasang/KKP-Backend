# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Callecto API — a Go (Fiber + MongoDB) backend for an automated outbound debt-collection
voicebot. It drives "call sessions" that dial debtors through the Botnoi outbound voicebot
service, then ingests results via webhooks and classifies the conversation outcome.

The Go service is a port of the Supabase edge function still present at the repo root as
`index.ts` (Deno/TypeScript) — keep `index.ts` as the behavioral reference when changing the
call-processing or webhook-classification logic. The module is still named `go-fiber-template`
(it was scaffolded from a template); imports use that prefix.

## Commands

```sh
go mod tidy            # install/sync dependencies
go run .               # run the API (also: go run main.go)
air                    # hot-reload dev server (needs: go install github.com/cosmtrek/air@latest)
go build -o main .     # build binary
go test ./...          # run all tests
go test ./src/services/ -run TestXxx   # run a single test by name
go run ./cmd/seed      # seed one debtor + pending call_list_item + running call_session
```

Docker: `docker build -t callecto-api .` then run exposing port 8080. The Dockerfile pins
`golang:1.22.1` but `go.mod` declares `go 1.23.0` — bump the image if a build fails on version.

## Environment

`.env` is loaded at startup via godotenv (falls back to system env if absent). Required keys:
`MONGODB_URI`, `MONGODB_NAME` (and legacy `DATABASE_NAME`), `PORT` (defaults to 8080).
Auth uses Supabase JWKS — set `JWK_SET_URL` or `SUPABASE_URL` (falls back to `VITE_SUPABASE_URL`);
if neither is set, it validates HS256 against `JWT_SECRET_KEY`. Outbound calls need
`OUTBOUND_URL` and `OUTBOUND_ACCESS_TOKEN`.

Note: `.env` is **not** in `.gitignore` and currently holds live MongoDB credentials and an
outbound access token. Do not add new secrets to it expecting them to be ignored; treat the
committed values as compromised.

## Architecture

Clean/layered architecture. A request flows **gateway → service → repository → MongoDB**, with
`entities` as shared models across all layers. Each layer is wired by hand in `main.go` — there is
no DI container, so adding a feature means constructing its repo, service, and gateway there and
passing them through.

- **`domain/entities/`** — data models (`*Model` structs) with bson/json tags. Shared by every layer.
- **`domain/repositories/`** — one file per collection. Each defines an `I<Name>Repository`
  interface + a private struct holding a `*mongo.Collection`. Collections are resolved from
  `MONGODB_NAME` at construction. Repos own all bson queries; errors are logged via `fiberlog`.
- **`domain/datasources/mongodb.go`** — single Mongo client (pool size 10) with Elastic APM
  command monitoring. `repo.New*(mongodb)` pulls `.Database(MONGODB_NAME).Collection("...")`.
- **`src/services/`** — business logic against repository interfaces, exposed as `I<Name>Service`.
- **`src/gateways/`** — Fiber HTTP handlers (methods on `HTTPGateway`). `route.go` registers all
  route groups; `http.go` defines `HTTPGateway` and `NewHTTPGateway` (the central wiring point).
- **`src/client/`** — outbound HTTP clients (resty), e.g. the Botnoi outbound voicebot.
- **`src/middlewares/`** — Fiber logger + JWT. `SetJWtHeaderHandler()` guards route groups;
  handlers call `DecodeJWTToken(ctx)` to read `user_id`/`uid` claims.
- **`cmd/seed/`** — standalone seeding program with fixed UUIDs (upsert-safe, re-runnable).

### Conventions to follow when extending

- **IDs are app-generated UUID strings** stored in a string `id` field (not Mongo `_id`).
  Repositories filter on `bson.M{"id": id}`. Services assign `uuid.NewString()` and set
  `CreatedAt`/`UpdatedAt` on create.
- **Multi-tenancy / ownership.** Most data is scoped by `workspace_id` and `user_id`. `*ByUser`
  service/repo methods enforce ownership; gateways take `user_id` from the JWT and `workspace_id`
  from the path or `?workspace_id=` query param, returning 403 on mismatch.
- **Responses** use `entities.ResponseModel{Message, Data}` for success and
  `entities.ResponseMessage{Message}` for errors. Handlers map service errors to HTTP status
  themselves.
- **New routes**: add a `Gateway<Name>` group in `route.go`, a handler method in a `src/gateways`
  file, then thread the service through `http.go` and `main.go`.

### Call-processing flow (the core feature)

`src/services/process_call_session.go` is the heart of the system. `POST /api/v1/call-process`
with `{session_id, action}`:

- `start`/`continue` → runs `ProcessSession` in a **background goroutine** (returns 200
  immediately, mirroring `EdgeRuntime.waitUntil` in `index.ts`). `pause`/`stop` update status.
- `ProcessSession` enforces business hours, resets stale `calling` items (>5 min) to failed,
  computes available concurrency slots (`Settings.ConcurrentCalls`, default 5), pulls pending
  `call_list_items`, and places calls **concurrently** (goroutines + WaitGroup). It **recurses**
  to keep filling slots, and marks the session `completed` only when nothing is calling/waiting.
- `Settings.TestMode` short-circuits the real outbound call with a randomized mock outcome and
  fabricated records/stats — use it to exercise the pipeline without dialing.
- Real calls go through `src/client/outbound_botnoi.go`; the call is correlated by
  `outbound_<itemID>`, which the webhook later echoes back.

Per call, state is written across four collections: `call_list_items` (status `calling`),
`call_records` (status `pending`), `call_attempts`, and debtor `stats`. The webhook closes the loop.

### Webhook + classification

`POST /api/v1/webhooks/botnoi` (no JWT) is handled in `src/services/webhook.go`. It receives the
Botnoi call result, then updates the matching `call_record`/`call_list_item`/`call_attempt` and
the debtor's aggregate stats, advancing the session. Conversation outcomes are mapped to the fixed
`CONVERSATION_CATEGORIES` taxonomy (Thai/English status names, main vs sub groups). Thai-language
helpers in `process_call_session.go` (`toThaiDigitSpeech`, `formatThaiDate`, Buddhist-era dates)
prepare TTS variables — preserve their Thai output when editing.

### Testing

Tests live next to the code as `*_test.go` (testify). Repository/service tests rely on interface
mocks rather than a live Mongo. When adding a service method, add it to the interface and update
any mock implementing it or the package won't compile.

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

