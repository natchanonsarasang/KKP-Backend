# FLOW.md — Process Call Session: Concurrent Calling Architecture

> **Scope**: This document covers how the frontend (`voice-debt-mate`) initiates and manages concurrent outbound calls through the backend (`callecto-api`) and Supabase Edge Functions, focusing exclusively on the **`process-call-session`** workflow.

---

## Table of Contents

1. [High-Level Architecture](#1-high-level-architecture)
2. [Entity Relationship Model](#2-entity-relationship-model)
3. [End-to-End Flow Diagram](#3-end-to-end-flow-diagram)
4. [Step-by-Step Workflow](#4-step-by-step-workflow)
5. [Concurrency Model](#5-concurrency-model)
6. [API Contracts](#6-api-contracts)
7. [Frontend Session Management](#7-frontend-session-management)
8. [Backend Processing Logic](#8-backend-processing-logic)
9. [Webhook Callback Loop](#9-webhook-callback-loop)
10. [Session Lifecycle States](#10-session-lifecycle-states)
11. [Settings & Configuration](#11-settings--configuration)
12. [File Reference Map](#12-file-reference-map)
13. [Future Go API-Only Migration Plan](#13-future-go-api-only-migration-plan)
14. [Future Auto-Calling Workflow for Uncalled Debtors](#14-future-auto-calling-workflow-for-uncalled-debtors)

---

## 1. High-Level Architecture

```mermaid
flowchart LR
    FE["Frontend<br/>(voice-debt-mate)<br/>CallList.tsx"]
    GO["Go API<br/>(callecto-api)<br/>Fiber + MongoDB"]
    SB["Supabase Edge Function<br/>process-call-session"]
    BN["Botnoi Voicebot<br/>Outbound Call API"]
    WH["Webhook<br/>(voicebot-webhook)"]

    FE -- "1. POST /call-sessions<br/>(create session)" --> GO
    FE -- "2. POST /call-process<br/>{action: start}" --> SB
    SB -- "3. Fetch pending items<br/>from Supabase DB" --> SB
    SB -- "4. POST call_message_public<br/>(concurrent calls)" --> BN
    BN -- "5. Webhook callback<br/>on call completion" --> WH
    WH -- "6. Update records<br/>+ trigger continue" --> SB
    FE -- "7. Poll every 2s<br/>GET /call-sessions" --> GO
```

The system uses a **fire-and-forget** pattern where the frontend sends a single "start" command, and the backend processes calls autonomously in the background — even if the user closes the browser.

---

## 2. Entity Relationship Model

```mermaid
erDiagram
    CallSession ||--o{ CallListItem : "processes"
    CallListItem ||--|| Debtor : "calls"
    CallListItem ||--o{ CallAttempt : "has attempts"
    CallListItem ||--o| CallRecord : "links to"
    CallAttempt ||--o| CallRecord : "references"
    Workspace ||--o{ CallSession : "contains"
    Workspace ||--o{ Debtor : "manages"
    Workspace ||--o{ CallListItem : "owns"

    CallSession {
        string id PK
        string user_id
        string workspace_id
        string status "pending|running|paused|stopped|completed"
        int total_calls
        int completed_calls
        int failed_calls
        int confirmed_calls
        int token_used
        json settings "concurrentCalls, maxRetries, etc"
        timestamp started_at
        timestamp completed_at
    }

    CallListItem {
        string id PK
        string debtor_id FK
        string workspace_id FK
        string template_id
        string status "pending|calling|success|failed|pending_retry"
        string call_record_id FK
        string call_outcome
        boolean picked_up
        string ai_category
        timestamp next_retry_at
        timestamp called_at
    }

    CallRecord {
        string id PK
        string phone_number
        string botnoi_call_id "Outbound ID from Botnoi"
        string status "pending|confirmed|declined|no_answer|failed|..."
        json result_data "Full webhook payload"
        int call_duration
        string user_id
        string workspace_id
    }

    CallAttempt {
        string id PK
        string call_list_item_id FK
        string call_record_id FK
        int attempt_number
        string status "calling|success|failed"
        string call_outcome
        boolean picked_up
        string ai_category
        string conversation_log
        string audio_url
        int call_duration
    }

    Debtor {
        string id PK
        string phone_number
        string name
        float total_debt
        int contact_attempts
        int picked_up_count
        string call_outcome
        boolean is_blocked
        json variables "Template variables"
    }
```

### Row Cardinality & Data Creation Rules

Within a specific workspace, data is stored and created according to these cardinalities:

*   **`CallSession`** (Campaign run): **1 Row per Campaign Run**
    *   Created when the user starts a session from the frontend.
    *   Tracks overall settings (concurrency, retries, schedules) and aggregate metrics (total, completed, failed, confirmed calls) for that specific run.
*   **`CallListItem`** (Debtor target queue): **1 Row per Target Debtor**
    *   Represents the debtor queued to be called.
    *   **Reused on retry**: If a debtor call fails or has no answer and needs to be called again, the **same** `CallListItem` row is updated (status resets to `pending_retry` or `pending` with `next_retry_at` scheduled).
    *   Always maintains the latest call state, latest call outcome, and the reference to the latest `CallRecord` (`call_record_id`).
*   **`CallRecord`** (Attempt details): **1 Row per Physical Outbound Call Attempt (1 Call = 1 Row)**
    *   A new row is created every single time an API request is successfully dispatched to Botnoi Voicebot (generating a unique `botnoi_call_id`).
    *   If a debtor is called 3 times (the initial call + 2 retries), there will be **1** `CallListItem` row, but **3** separate `CallRecord` rows (and 3 corresponding `CallAttempt` rows documenting conversation details, durations, and audio URLs).

## 3. End-to-End Flow Diagram (Go-Only Architecture)

```mermaid
sequenceDiagram
    participant User as User (Browser)
    participant FE as Frontend<br/>CallList.tsx
    participant GoAPI as Go API<br/>(callecto-api)
    participant MongoDB as MongoDB
    participant Botnoi as Botnoi Voicebot API

    Note over User, FE: Phase 1: Queue Debtors
    User->>FE: Click "Queue All Debtors"
    FE->>GoAPI: POST /api/v1/call-list-items (for each debtor)
    GoAPI->>MongoDB: Insert CallListItem (status: "pending")

    Note over User, FE: Phase 2: Create Session & Process
    User->>FE: Click "Start Calling"
    FE->>GoAPI: POST /api/v1/call-sessions {status: "running", settings}
    GoAPI->>MongoDB: Insert CallSession (status: "running")
    FE->>GoAPI: POST /api/v1/call-process {session_id, action: "start"}
    
    Note over GoAPI, MongoDB: Phase 3: Slot Calculation & Allocation
    GoAPI->>GoAPI: Acquire Workspace Lock (Mutex)
    GoAPI->>MongoDB: Fetch running CallSession settings
    GoAPI->>MongoDB: Fetch items where status = "calling"
    GoAPI->>GoAPI: Reset stale calling items (>5m) to "failed"
    GoAPI->>GoAPI: availableSlots = maxConcurrent - activeCallingCount
    
    Note over GoAPI, Botnoi: Phase 4: Auto-Dialing / Dispatch
    GoAPI->>MongoDB: Fetch pending list items (LIMIT availableSlots)
    alt Queue is Dry & Auto-Calling Enabled
        GoAPI->>MongoDB: Find uncalled debtors (attempts=0, not in call_list_items)
        GoAPI->>MongoDB: Auto-insert new CallListItem (status: "calling")
    end
    
    par Concurrent Calls
        GoAPI->>MongoDB: Update CallListItem -> status="calling", called_at=now
        GoAPI->>MongoDB: Create CallRecord (status="pending")
        GoAPI->>MongoDB: Create CallAttempt (status="calling")
        GoAPI->>MongoDB: Update Debtor stats (contact_attempts++)
        GoAPI->>Botnoi: POST call_message_public (Outbound Call)
    end
    
    Note over Botnoi, GoAPI: Phase 5: Webhook Callback & Continuation
    Botnoi->>GoAPI: POST /api/v1/webhooks/botnoi {outbound_id, status, ...}
    GoAPI->>MongoDB: Update CallRecord status (final outcome)
    GoAPI->>MongoDB: Update CallListItem status (success/failed)
    GoAPI->>MongoDB: Update CallAttempt (transcript, duration)
    GoAPI->>MongoDB: Update Debtor counters (picked_up, response type)
    GoAPI->>MongoDB: Update CallSession progress counters
    GoAPI->>GoAPI: Trigger ProcessSession(sessionID) in goroutine
    GoAPI->>MongoDB: Recalculate slots & dispatch next concurrent batch
```

---

## 4. Step-by-Step Workflow & Database Mutations

Here is the exact step-by-step description of how data is created, read, and updated in MongoDB by the Go backend services.

### Phase 1: Database Operations Triggered by Frontend Actions

When the user queues debtors and starts the campaign, the frontend creates the baseline records:

1.  **Queue Debtor (Manual)**:
    *   **Action**: Clicking "Queue All Debtors" in the frontend.
    *   **Database Write**: Inserts a new document in **`call_list_items`** for each queued debtor:
        *   `id`: new UUID
        *   `workspace_id`: workspace ID
        *   `debtor_id`: debtor ID
        *   `status`: `"pending"`
        *   `retry_count`: `0`
2.  **Create Campaign Session**:
    *   **Action**: Clicking "Start Calling" in the frontend.
    *   **Database Write**: Inserts a new document in **`call_sessions`**:
        *   `id`: `sessionID` (UUID)
        *   `status`: `"running"`
        *   `total_calls`: number of items in the queue
        *   `settings`: concurrency limits, business hours, retries
        *   `started_at`: `time.Now().UTC()`

### Phase 2: Operations Triggered by the Call-Process Route (`POST /api/v1/call-process`)

Upon receiving a request with `action: "start"`, the HTTP Handler invokes `ICallProcessService.ProcessSession(...)`:

1.  **Calculate Slot Capacity**:
    *   **Read**: Fetches the active `CallSession` to retrieve configured `settings.concurrentCalls` (default: 5).
    *   **Read**: Counts the active calling items currently running in the workspace:
        `SELECT count FROM call_list_items WHERE status = 'calling' AND workspace_id = X`
    *   **Reset Stale Items**: If any item has been in `calling` status for longer than 5 minutes (stale threshold), the database is mutated:
        *   **Update** `call_list_items` setting `status = 'failed'` and `call_outcome = 'Call timed out'`.
        *   **Update** `call_attempts` setting `status = 'failed'` and details to `"Stale timeout"`.
    *   **Arithmetic**: Available slots are calculated as:
        $$\text{availableSlots} = \text{maxConcurrent} - \text{activeCallingCount}$$
        *(If $\text{availableSlots} \le 0$, the service exits immediately to wait for webhooks).*
2.  **Load Targets**:
    *   **Read**: Queries the database for pending queue items:
        `SELECT FROM call_list_items WHERE status = 'pending' LIMIT availableSlots`
3.  **Prepare Batch & Mutate Database**:
    For each target item selected in the batch:
    *   **Update** `call_list_items` to set:
        *   `status`: `"calling"`
        *   `called_at`: `time.Now().UTC()`
        *   `call_outcome`: `"Call initiated - awaiting response"`
    *   **Write** a new document in **`call_records`**:
        *   `id`: new UUID (`call_record_id`)
        *   `phone_number`: debtor phone number
        *   `botnoi_call_id`: `"outbound_" + item.ID`
        *   `status`: `"pending"`
    *   **Update** the `call_list_items` record:
        *   `call_record_id`: links the generated `call_record_id` back onto the item.
    *   **Write** a new document in **`call_attempts`**:
        *   `call_list_item_id`, `call_record_id`: relationships
        *   `attempt_number`: `item.retry_count + 1`
        *   `status`: `"calling"`
        *   `call_outcome`: `"Call initiated - awaiting response"`
    *   **Update** `debtors` stats:
        *   `contact_attempts`: `debtor.contact_attempts + 1`
        *   `last_contact_at`: `time.Now().UTC()`
4.  **Dispatch API Calls**:
    *   Invokes the Botnoi Outbound API concurrently for each debtor in the batch using goroutines.

### Phase 3: Operations Triggered by the Webhook Callback (`POST /api/v1/webhooks/botnoi`)

When Botnoi calls finish, the webhook handler updates the final transaction states:

1.  **Update Call Logs**:
    *   **Update** **`call_records`**: Sets `status` (e.g. `completed`, `no_answer`, `confirmed`, `declined`), duration, and appointment date/time.
    *   **Update** **`call_attempts`**: Sets `status` (`success` or `failed`), `call_outcome`, `conversation_log`, `audio_url`, and duration.
2.  **Update Queue List Item**:
    *   **Update** **`call_list_items`**: Sets final state (`success`/`failed`), `call_outcome`, `picked_up` flag, and raw conversation details inside `notes`.
3.  **Update Debtor Counters**:
    *   **Update** **`debtors`**: Increments `picked_up_count`/`not_picked_up_count` and response counts (`accept_count`/`reject_count` based on confirmed/declined outcomes).
4.  **Update Session Progress**:
    *   **Update** **`call_sessions`**: Increments `completed_calls` or `failed_calls`, and updates `confirmed_calls`.
5.  **Trigger Next Batch**:
    *   Directly calls `go s.CallProcessService.ProcessSession(sessionID)` to fill slots.

### Phase 4: Database Mutations for "AUTO" Calling Uncalled Debtors

If the workspace queue in `call_list_items` is dry, but auto-calling is enabled, the backend populates calling slots automatically:

1.  **Retrieve Uncontacted Debtors**:
    *   **Read**: Queries the `debtors` collection for records where:
        *   `contact_attempts == 0` (or `last_contact_at` is null)
        *   `id NOT IN` the existing `call_list_items` collection for this workspace.
2.  **Auto-Queue Ingestion**:
    *   **Write**: Automatically inserts a new record into **`call_list_items`** for each chosen debtor:
        *   `id`: new UUID
        *   `workspace_id`: workspace ID
        *   `debtor_id`: debtor ID
        *   `status`: `"calling"`
        *   `called_at`: `time.Now().UTC()`
        *   `retry_count`: `0`
3.  **Instantiate Call Logs & Dispatch**:
    *   Proceeds to write the `CallRecord` and `CallAttempt` (same as Phase 2, step 3).
    *   Dispatches calls to Botnoi concurrently.

---

## 5. Concurrency Model

```mermaid
flowchart TD
    A["ProcessSession starts"] --> B["Acquire Workspace Mutex"]
    B --> C["Count items WHERE status='calling'"]
    C --> D{"activeCount < maxConcurrent?"}
    D -- "Yes" --> E["availableSlots = max - active"]
    D -- "No" --> F["Release Mutex & wait for webhooks"]
    E --> G["Fetch LIMIT(availableSlots) pending items"]
    G --> H{"Found items?"}
    H -- "No" --> I["Are there uncalled debtors?<br/>(AUTO CALL ENABLED)"]
    H -- "Yes" --> K["Mark items as 'calling' (called_at = now)"]
    I -- "Yes" --> J["Auto-create CallListItems (status: calling)"]
    I -- "No" --> L["Release Mutex & complete/wait"]
    J --> K
    K --> M["Create CallRecord (pending)<br/>& CallAttempt (calling)"]
    M --> N["Release Mutex"]
    N --> O["Concurrently make API calls to Botnoi"]
    O --> P["Botnoi completes call"]
    P --> Q["Webhook callback received"]
    Q --> R["Update records & session counters"]
    R --> S["Trigger ProcessSession asynchronously"]
    S --> A
```

### Concurrency Rules

| Rule | Value | Source |
|------|-------|-------|
| Max concurrent calls | 1–10 (default: 5) | `settings.concurrentCalls` |
| Slot calculation | `maxConcurrent - activeCallingCount` | Queried from `call_list_items WHERE status='calling'` |
| Stale timeout | 5 minutes | Items in `"calling"` with `called_at` older than 5 min are reset to `"failed"` |
| Retry delay | 1 minute | Items with `status: "pending_retry"` wait for `next_retry_at` |
| Parallelism mechanism | `go func(...)` goroutines | Spawns separate threads with a `sync.WaitGroup` pool |

### How the "Conveyor Belt" Works

1. **Session starts**: The caller triggers the `POST /api/v1/call-process` handler. It loads `N` items (where `N = availableSlots`), updates their database status to `"calling"`, and spawns concurrent goroutines to fire calls to Botnoi simultaneously.
2. **Call completes**: Botnoi webhook calls the Go `/api/v1/webhooks/botnoi` callback endpoint. The webhook service processes conversation records and writes updates to `CallRecord`, `CallListItem`, `CallAttempt`, `Debtor`, and `CallSession` progress counters.
3. **Internal continue re-trigger**: As soon as the webhook updates are saved, the handler fires an asynchronous execution of `CallProcessService.ProcessSession(...)` in a background goroutine.
4. **Next batch fires**: The background thread recalculates the remaining available slots and instantly fills them with the next pending queue items or uncalled debtors.
5. **Repeat**: This loops automatically until all queued list items and uncontacted debtors are processed, at which point the session status is marked `"completed"`.

This creates a **self-sustaining loop** where the system maintains up to `maxConcurrent` active calls at all times in the background without needing any browser polling or frontend execution.

---

## 6. API Contracts

### Frontend → Go API (callecto-api)

#### `POST /api/v1/call-sessions` — Create Session

```json
// Request
{
  "id": "uuid-v4",
  "workspace_id": "workspace-uuid",
  "status": "running",
  "total_calls": 42,
  "settings": {
    "maxRetries": 2,
    "delayBetweenCalls": 5,
    "concurrentCalls": 5,
    "businessHoursOnly": true,
    "businessHoursStart": "09:00",
    "businessHoursEnd": "18:00",
    "businessDays": [1, 2, 3, 4, 5],
    "testMode": false,
    "timezoneOffset": 420,
    "interruptible": false
  }
}

// Response 200
{ "message": "success" }
```

#### `GET /api/v1/call-sessions?workspace_id=X&user_id=Y` — Poll Sessions

```json
// Response 200
{
  "message": "success",
  "data": [
    {
      "id": "session-uuid",
      "status": "running",
      "total_calls": 42,
      "completed_calls": 15,
      "failed_calls": 3,
      "confirmed_calls": 8,
      "token_used": 18,
      "started_at": "2026-06-22T10:00:00Z",
      "completed_at": null
    }
  ]
}
```

#### `PUT /api/v1/call-sessions/:id` — Update Session

Used by the frontend to set `status: "running"` when resuming a paused session.

### Frontend → Supabase Edge Function (process-call-session)

#### `POST /functions/v1/process-call-session`

```json
// Start processing
{ "session_id": "uuid", "action": "start" }

// Pause (from frontend)
{ "session_id": "uuid", "action": "pause" }

// Stop completely (from frontend)
{ "session_id": "uuid", "action": "stop" }

// Continue (from webhook, triggers next batch)
{ "session_id": "uuid", "action": "continue" }
```

| Action | Behavior |
|--------|----------|
| `start` | Begin or resume processing with `EdgeRuntime.waitUntil()` for background execution |
| `continue` | Same as `start` — triggered by webhook after a call completes |
| `pause` | Set session `status: "paused"`, processing stops at next check |
| `stop` | Set session `status: "stopped"` + `completed_at`, terminate immediately |

### Go Backend → Supabase Edge Function (via webhook)

The Go webhook service triggers re-processing after updating session counters:

```go
// webhook.go — triggerSessionProcessor
func (s *webhookService) triggerSessionProcessor(sessionID string) {
    url := fmt.Sprintf("%s/functions/v1/process-call-session", supabaseURL)
    payload := map[string]string{"session_id": sessionID, "action": "continue"}
    // POST with SUPABASE_SERVICE_ROLE_KEY
}
```

### Edge Function → Botnoi Voicebot API

```json
// POST https://bn-voicebot-system-9ehp.onrender.com/api/voicebot/custom/call_message_public
{
  "outbound_id": "outbound_<call_list_item_id>",
  "event_id": "event_<session_id>_<item_id>",
  "tel_number": "0812345678",
  "phonenumber": "0812345678",
  "variables": {
    "name": "สมชาย",
    "outstanding_amount": "หนึ่งหมื่นห้าพัน",
    "due_date": "วันจันทร์ ที่ 22 มิถุนายน 2569",
    "policy_no": "หนึ่ง สอง สาม สี่ ห้า"
  },
  "bot_id": "6a06964fb875327d960f05f0",
  "bot_type": "Confirm1",
  "speaker": "212",
  "language": "th",
  "tts": "voicebot-premium",
  "asr_provider": "botnoi-aws-th-noise-classifier-v17c",
  "interruptible": "True"
}
```

---

## 7. Frontend Session Management

### Session Polling

The frontend polls for active session status every **2 seconds**:

```typescript
// CallList.tsx — lines 453-469
const { data: activeSession, refetch: refetchSession } = useQuery({
  queryKey: ["active-call-session", effectiveUserId, currentWorkspace?.id],
  queryFn: async () => {
    const sessions = await listCallSessions({
      workspace_id: currentWorkspace.id,
      user_id: effectiveUserId,
    });
    const active = sessions
      .filter((s) => ["running", "stopping", "paused"].includes(s.status))
      .sort((a, b) => (b.created_at || "").localeCompare(a.created_at || ""));
    return active[0] ?? null;
  },
  refetchInterval: 2000,
});
```

Call list items are polled every **10 seconds** for individual status updates.

> **No WebSocket/realtime** — the system relies entirely on polling.

### User Actions

| Action | Frontend Function | API Call |
|--------|-------------------|----------|
| Start Calling | `startCallingSession()` | `createCallSession()` + `processCallSession({action: "start"})` |
| Pause | `pauseCallingSession()` | `processCallSession({action: "pause"})` |
| Resume | `resumeCallingSession()` | `updateCallSession({status: "running"})` + `processCallSession({action: "start"})` |
| Stop | `stopCallingSession()` | `processCallSession({action: "stop"})` |

### Progress Display

The UI shows a live progress banner with:

- **Progress bar**: `(completed_calls + failed_calls) / total_calls`
- **Active calls**: `callingCount / concurrentCalls max`
- **Completed count**: `session.completed_calls`
- **Confirmed count**: `session.confirmed_calls`
- **Failed count**: `session.failed_calls`
- **Tokens used**: `session.token_used`

---

## 8. Backend Processing Logic

### Two Backend Systems

| System | Where | Handles |
|--------|-------|---------|
| **Go API** (`callecto-api`) | `src/services/call_sessions.go` | CRUD for sessions, records, items, attempts. Webhook processing. Session counter updates. |
| **Supabase Edge Function** | `supabase/functions/process-call-session/` | Orchestration logic: pick pending items, manage concurrency slots, fire calls to Botnoi, loop control. |

### Go API Call Session Service

The Go service provides CRUD with ownership checks:

```go
// services/call_sessions.go
type ICallSessionsService interface {
    CreateCallSessionByUser(callerUserID string, data CallSessionDataModel) error
    GetCallSessionsByUser(callerUserID string, filter CallSessionFilter) (*[]CallSessionDataModel, error)
    UpdateCallSessionByUser(callerUserID string, id string, data CallSessionDataModel) error
    DeleteCallSessionByUser(callerUserID string, id string) error
}
```

All "ByUser" methods enforce ownership — `callerUserID` must match the session's `UserID`.

### Go Webhook Processing (`webhook.go`)

After receiving a webhook from Botnoi:

1. **Resolve call identity**: Find `CallRecord` by `botnoi_call_id`, fallback to `Debtor` by phone number
2. **Map status**: Botnoi raw status → internal status enum (confirmed/declined/no_answer/etc.)
3. **AI categorize**: Send conversation log to Gemini for classification into 16 categories
4. **Cascade updates**: CallRecord → CallListItem → CallAttempt → Debtor → CallSession
5. **Re-trigger**: Call `triggerSessionProcessor(sessionID)` to fire the next batch

---

## 9. Webhook Callback Loop

```mermaid
flowchart TD
    subgraph "Botnoi Voicebot"
        A["Call completes"]
    end

    subgraph "Webhook Handler"
        B["Receive POST /webhooks/botnoi"]
        C["Parse payload<br/>(outbound_id, status, conversation_log)"]
        D["Map status to internal enum"]
        E["AI classify conversation<br/>(Gemini 2.5 Flash)"]
        F["Update CallRecord"]
        G["Update CallListItem"]
        H["Update CallAttempt"]
        I["Update Debtor stats"]
        J["Update CallSession counters"]
        K["triggerSessionProcessor(sessionId)"]
    end

    subgraph "process-call-session"
        L["Receive {action: continue}"]
        M["Recalculate available slots"]
        N["Fetch & fire next batch"]
    end

    A --> B --> C --> D --> E
    E --> F --> G --> H --> I --> J --> K
    K --> L --> M --> N
```

### Status Mapping (Webhook)

| Botnoi Status / Action | Mapped Status | Final Status |
|------------------------|---------------|-------------|
| action=confirm/yes | `confirmed` | `success` |
| action=decline/no | `declined` | `success` |
| action=unknown | `no_response` | `success` |
| status=completed (user spoke) | `completed` | `success` |
| status=completed (no user speech) | `no_answer` | `failed` |
| status=hanged_up/hangup | `hanged_up` | `failed` |
| status=no answer | `no_answer` | `failed` |
| status=busy | `busy` | `failed` |
| status=failed/error | `failed` | `failed` |
| status=rejected | `rejected` | `failed` |
| status=voicemail | `voicemail` | `failed` |

---

## 10. Session Lifecycle States

```mermaid
stateDiagram-v2
    [*] --> pending: NewCallSession()
    pending --> running: createCallSession(status: "running")
    running --> paused: processCallSession(action: "pause")
    running --> stopped: processCallSession(action: "stop")
    running --> completed: All items processed
    running --> paused: Outside business hours (auto)
    paused --> running: processCallSession(action: "start")
    paused --> stopped: processCallSession(action: "stop")
    stopped --> [*]
    completed --> [*]
```

| Status | Meaning | Triggered By |
|--------|---------|-------------|
| `pending` | Session created, not yet started | Default on creation |
| `running` | Actively processing calls | User clicks "Start" or "Resume" |
| `paused` | Temporarily halted | User clicks "Pause" or auto (outside business hours) |
| `stopped` | Terminated by user | User clicks "Stop" |
| `completed` | All items processed | Automatic when no pending items and no active calls remain |

---

## 11. Settings & Configuration

### AutoDialSettings (passed in session creation)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `concurrentCalls` | `number` | `5` | Max simultaneous outbound calls (1–10) |
| `maxRetries` | `number` | `2` | Max retry attempts per debtor |
| `delayBetweenCalls` | `number` | `5` | Seconds between batches |
| `businessHoursOnly` | `boolean` | `true` | Restrict to business hours |
| `businessHoursStart` | `string` | `"09:00"` | Start of calling window |
| `businessHoursEnd` | `string` | `"18:00"` | End of calling window |
| `businessDays` | `number[]` | `[1,2,3,4,5]` | Mon-Fri (0=Sun, 6=Sat) |
| `testMode` | `boolean` | `false` | Simulate calls without hitting Botnoi API |
| `timezoneOffset` | `number` | Auto-detect | UTC offset in minutes (e.g., +7h = 420) |
| `interruptible` | `boolean` | `false` | Whether bot speech can be interrupted |

### Environment Variables

| Variable | Used By | Purpose |
|----------|---------|---------|
| `VITE_CALLECTO_API_URL` | Frontend | Go API base URL (e.g., `http://localhost:1818/api/v1`) |
| `SUPABASE_URL` | Go API, Edge Functions | Supabase project URL |
| `SUPABASE_SERVICE_ROLE_KEY` | Go API, Edge Functions | Service role key for server-to-server calls |
| `LOVABLE_API_KEY` | Go webhook | API key for AI classification |

---

## 12. File Reference Map

### callecto-api (Go Backend)

| File | Purpose |
|------|---------|
| [call_sessions.go](file:///home/cellul4r/Documents/botnoi/callecto-api/domain/entities/call_sessions.go) | `CallSessionDataModel` entity + filter |
| [call_records.go](file:///home/cellul4r/Documents/botnoi/callecto-api/domain/entities/call_records.go) | `CallRecordDataModel` entity + status enum |
| [call_attempts.go](file:///home/cellul4r/Documents/botnoi/callecto-api/domain/entities/call_attempts.go) | `CallAttemptModel` entity |
| [call_list_items.go](file:///home/cellul4r/Documents/botnoi/callecto-api/domain/entities/call_list_items.go) | `CallListItemModel` entity |
| [call_sessions.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/services/call_sessions.go) | Session service (CRUD + ownership) |
| [webhook.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/services/webhook.go) | Webhook processing, AI classify, trigger loop |
| [call_sessions.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/gateways/call_sessions.go) | HTTP handlers for `/call-sessions` |
| [webhook.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/gateways/webhook.go) | HTTP handler for `/webhooks/botnoi` |
| [route.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/gateways/route.go) | All route definitions |
| [http.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/gateways/http.go) | Gateway struct + service wiring |

### voice-debt-mate (Frontend + Supabase)

| File | Purpose |
|------|---------|
| [CallList.tsx](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/src/test/CallList.tsx) | Main component: session mgmt, queue, polling, UI |
| [voicebot.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/src/test/api/voicebot.ts) | `processCallSession()` + `makeCall()` API wrappers |
| [callSessions.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/src/test/api/callSessions.ts) | Session CRUD API client |
| [callRecords.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/src/test/api/callRecords.ts) | Record CRUD API client |
| [callListItems.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/src/test/api/callListItems.ts) | Call list item CRUD API client |
| [callAttempts.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/src/test/api/callAttempts.ts) | Attempt CRUD API client |
| [client.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/src/test/api/client.ts) | HTTP client with Supabase JWT auth |
| [types.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/src/test/api/types.ts) | TypeScript interfaces for all entities |
| [process-call-session/index.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/supabase/functions/process-call-session/index.ts) | Edge Function: orchestration loop, concurrent calls |
| [voicebot-webhook/index.ts](file:///home/cellul4r/Documents/botnoi/voice-debt-mate/supabase/functions/voicebot-webhook/index.ts) | Webhook handler: status mapping, AI classify, re-trigger |

---

> **Note**: The `voicebot-make-call` function is intentionally excluded from this document. It operates as a standalone single-call endpoint and is not part of the session-based concurrent calling workflow.

---

## 13. Future Go API-Only Migration Plan

In the future, the dual-system architecture (Go backend + Supabase DB & Edge Functions) will be consolidated into a unified backend using only **Go API** and **MongoDB**, eliminating Supabase Edge Functions and Supabase DB entirely.

```mermaid
flowchart TD
    FE["Frontend<br/>(voice-debt-mate)<br/>CallList.tsx"]
    GO["Go API<br/>(callecto-api)<br/>Fiber + MongoDB"]
    BN["Botnoi Voicebot<br/>Outbound Call API"]
    WH["Webhook Handler<br/>(callecto-api)"]

    FE -- "1. POST /call-sessions (start)" --> GO
    GO -- "2. Background worker: Fetch pending<br/>& dispatch concurrent calls" --> BN
    BN -- "3. Call completes webhook" --> WH
    WH -- "4. Update MongoDB<br/>& notify internal worker" --> GO
    GO -- "5. Process next batch (continue)" --> GO
    FE -- "6. Poll progress GET /call-sessions" --> GO
```

### Migration Roadmap & Architectural Changes

1.  **Unified Database Strategy (MongoDB)**
    *   Migrate all relational tables (`call_sessions`, `call_list_items`, `call_records`, `call_attempts`, `debtors`, `call_tokens`, `call_templates`) from Supabase PostgreSQL to MongoDB collections in `callecto-api`.
    *   The Go API will become the single source of truth for all transactional calling data.
2.  **Porting Concurrency and Loop Logic**
    *   Rewrite the Deno TypeScript logic (`process-call-session`) into a Go background daemon/worker service in `callecto-api`.
    *   Utilize Go-native concurrency mechanisms (goroutines, channels, or an asynchronous task queue like Asynq/Machinery) to manage the available slots:
        $$\text{availableSlots} = \text{maxConcurrent} - \text{activeCallingCount}$$
    *   The background worker will fetch pending `CallListItem` documents from MongoDB, mark them as `calling`, and call the Botnoi API concurrently using HTTP client connection pooling.
3.  **Direct Service-to-Service Loop Re-trigger**
    *   In the future architecture, instead of using the network webhook re-trigger, the webhook handler service layer directly invokes the session service layer's batch processing method.
    *   This keeps the loop in-process inside the Fiber backend, eliminating external HTTP overhead, API gateways, and authorization token handshakes.
4.  **State Management & Auth**
    *   Move authentication validation in `callecto-api` middleware from checking Supabase JWT tokens to standard stateless JWTs generated directly by `callecto-api` or a custom auth provider.

### Service Layer Integration & Code Reuse

To achieve clean separation of concerns, satisfy the **Single Responsibility Principle (SRP)**, and prevent **circular dependencies**, we utilize the actual **`ICallProcessService`** (located in [process_call_session.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/services/process_call_session.go)):

```
  [Webhook Callback Route]               [Start Session Route]
             │                                     │
             ▼                                     ▼
     [WebhookService]                      [Session Controller]
             │                                (HTTP Gateway)
             │                                     │
             │      ┌──────────────────────────────┤
             │      │                              ▼
             ▼      ▼                     [CallSessionsService]
        [ICallProcessService]                  (CRUD only)
                 │
  ┌──────────────┼──────────────┐
  ▼              ▼              ▼
[MongoDB] [OutboundClient] [Workspace Mutex]
```

1.  **Dependency Injection Flow**:
    *   **`CallSessionsService`** (in [call_sessions.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/services/call_sessions.go)) does **not** depend on or inject `ICallProcessService`. It is kept completely focused on simple database CRUD.
    *   The **Session Controller/Gateway Handler** (in [process_call_session.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/gateways/process_call_session.go)) maps the route `POST /api/v1/call-process` and injects **`ICallProcessService`** (which dispatches calls).
    *   **`WebhookService`** (in [webhook.go](file:///home/cellul4r/Documents/botnoi/callecto-api/src/services/webhook.go)) injects `ICallProcessService` to trigger continuation batches.
2.  **Unification of Calling Logic**:
    *   Both the **initial campaign start trigger** (invoked via the existing HTTP route `POST /api/v1/call-process` with `{ "session_id": "uuid", "action": "start" }`) and the **webhook continuation trigger** (invoked from `ProcessWebhook` when a call completes) reuse the exact same `ProcessSession(sessionID)` method inside `ICallProcessService`.
    *   This guarantees that slot-capacity checking, uncalled-debtor selection, state transition, and API dispatch are encapsulated in a single, reusable worker class.
    *   **API Compatibility**: The existing Go handler for `POST /api/v1/call-process` accepts the exact same JSON payload structure as the legacy Supabase Edge function, meaning the frontend only needs to update its target URL and keep the request payload identical.

---

## 14. Future Auto-Calling Workflow for Uncalled Debtors

To improve agent utilization and efficiency, an "AUTO" calling mode will be implemented. After receiving and processing a webhook callback, the Go backend will automatically find and call debtors in the same workspace who have never been called.

```mermaid
flowchart TD
    A["Webhook Call Completes"] --> B["Go Webhook Service: Update database stats"]
    B --> C{"Check Session Running?"}
    C -- "No" --> D["End execution"]
    C -- "Yes" --> E{"Are there free slots<br/>(active < maxConcurrent)?"}
    E -- "No" --> F["Wait for other webhooks"]
    E -- "Yes" --> G["Query Workspace Debtors<br/>WHERE contact_attempts = 0<br/>AND ID NOT IN call_list_items"]
    G --> H{"Found Debtors?"}
    H -- "No" --> I["Fallback: Fetch remaining pending items"]
    H -- "Yes" --> J["Auto-create CallListItem (status: calling)"]
    J --> K["Create CallRecord + CallAttempt"]
    K --> L["Call Botnoi Voicebot API (AUTO CALL)"]
    L --> F
```

### Specifications for Auto-Calling

*   **Target Selection Rule**:
    When a call slot becomes available, the Go backend will filter the workspace debtors collection using:
    1.  `contact_attempts == 0` (or `last_contact_at` is null), ensuring they have never received an outbound call.
    2.  `id NOT IN` the existing `call_list_items` collection for this workspace (to prevent double-queuing or duplicates).
*   **Automatic Ingestion**:
    *   The backend will dynamically create a new `CallListItem` for each selected debtor, bypass the manual frontend queuing interface, and mark the status directly as `calling`.
    *   It will create the corresponding `CallRecord` and `CallAttempt`.
*   **Execution Flow**:
    *   After parsing the completed call webhook and updating the counters, the backend immediately checks if the workspace session is still in `running` status.
    *   It calculates the remaining capacity. If capacity is available, it pulls uncalled debtors up to the available capacity, saves them, and dispatches the call immediately to Botnoi.
    *   This ensures the queue never runs dry as long as there are uncontacted debtors in the workspace.

---

## 15. Concurrency Race Conditions & Prevention

In a high-throughput, concurrent calling system, race conditions are a critical risk during automated debtor selection.

### The Race Condition Scenario
If two active calls finish at the exact same moment:
1.  **Webhook A** and **Webhook B** hit the Go API `/webhooks/botnoi` concurrently.
2.  Both handlers update their records and simultaneously calculate:
    $$\text{activeCalls} = 3, \quad \text{maxConcurrent} = 5 \implies \text{slotsAvailable} = 2$$
3.  Both handlers independently query the database to pull uncalled debtors.
4.  If they both fetch the same debtor (e.g., *Debtor X*), they will both attempt to trigger a call for *Debtor X*.
5.  **Result**: The debtor receives two phone calls at the same time (double dialing).

### Prevention Strategies in Go + MongoDB

To eliminate this race condition in the Go API-only architecture, the following strategies will be implemented:

1.  **Workspace-Level Serialization Lock (Recommended)**
    *   Implement an in-memory Mutex or lock registry mapping active workspaces.
    *   When the webhook handler finishes updating status and decides to check/fill slots, it must acquire the lock for that workspace.
    *   This ensures that even if 10 webhooks finish simultaneously, only one goroutine at a time can query the database for uncalled debtors and mark them as `calling`.
2.  **Atomic Database Reservation**
    *   Create a compound unique index on `call_list_items` for `(workspace_id, debtor_id)`.
    *   Use MongoDB's atomic query `FindAndModify` (or `FindOneAndUpdate`) to select the next uncalled debtor and write a `CallListItem` in a single transaction.
    *   If a parallel thread attempts to queue the same debtor, the unique constraint will trigger a duplicate key error, allowing the thread to safely catch the error and pick the next debtor.

---

## 16. Webhook Auto-Calling Example (Go Code)

Here is a conceptual implementation demonstrating the Clean Architecture structure: the **Webhook Service** handles webhook record updates and triggers the **Call Process Service**, which owns the concurrent dialing logic.

```go
package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
)

// =========================================================================
// 1. WEBHOOK SERVICE LAYER (webhook.go)
// =========================================================================

type webhookService struct {
	CallRecordsService  ICallRecordsService
	DebtorService       IDebtorsService
	CallListItemService ICallListItemsService
	CallAttemptService  ICallAttemptsService
	CallProcessService  ICallProcessService // Injected Call Process Service
}

func NewWebhookService(
	callRecords ICallRecordsService,
	debtors IDebtorsService,
	items ICallListItemsService,
	attempts ICallAttemptsService,
	process ICallProcessService,
) IWebhookService {
	return &webhookService{
		CallRecordsService:  callRecords,
		DebtorService:       debtors,
		CallListItemService: items,
		CallAttemptService:  attempts,
		CallProcessService:  process,
	}
}

// ProcessWebhook handles callback, updates outcome stats, and calls CallProcessService
func (s *webhookService) ProcessWebhook(payload WebhookPayload) error {
	// A. Update call details and debtor stats in MongoDB
	err := s.updateDatabaseRecords(payload)
	if err != nil {
		return fmt.Errorf("failed to save webhook result: %w", err)
	}

	// B. Reuse ProcessSession: Call the callProcessService asynchronously.
	// Running inside a goroutine returns an immediate 200 OK response to Botnoi.
	go func() {
		_ = s.CallProcessService.ProcessSession(payload.SessionID)
	}()

	return nil
}

func (s *webhookService) updateDatabaseRecords(payload WebhookPayload) error {
	// DB modifications here
	return nil
}

// =========================================================================
// 2. PROCESS CALL SESSION SERVICE LAYER (process_call_session.go)
// =========================================================================

type ICallProcessService interface {
	ProcessSession(sessionID string) error
	PauseSession(sessionID string) error
	StopSession(sessionID string) error
}

type callProcessService struct {
	CallSessionsRepository  repositories.ICallSessionsRepository
	CallListItemsRepository repositories.ICallListItemsRepository
	DebtorsRepository       repositories.IDebtorsRepository
	CallRecordsRepository   repositories.ICallRecordsRepository
	CallAttemptsRepository  repositories.ICallAttemptsRepository
	OutboundClient          client.IOutboundBotnoiClient
	locks                   map[string]*sync.Mutex
	locksMu                 sync.Mutex
}

func NewCallProcessService(
	sessions repositories.ICallSessionsRepository,
	items repositories.ICallListItemsRepository,
	debtors repositories.IDebtorsRepository,
	records repositories.ICallRecordsRepository,
	attempts repositories.ICallAttemptsRepository,
	client client.IOutboundBotnoiClient,
) ICallProcessService {
	return &callProcessService{
		CallSessionsRepository:  sessions,
		CallListItemsRepository: items,
		DebtorsRepository:       debtors,
		CallRecordsRepository:   records,
		CallAttemptsRepository:  attempts,
		OutboundClient:          client,
		locks:                   make(map[string]*sync.Mutex),
	}
}

// GetLock retrieves or creates a mutex for a specific workspace to serialize checks
func (sv *callProcessService) GetLock(workspaceID string) *sync.Mutex {
	sv.locksMu.Lock()
	defer sv.locksMu.Unlock()

	if _, exists := sv.locks[workspaceID]; !exists {
		sv.locks[workspaceID] = &sync.Mutex{}
	}
	return sv.locks[workspaceID]
}

// ProcessSession computes slot capacity, handles retries, and triggers concurrent dialing
func (sv *callProcessService) ProcessSession(sessionID string) error {
	session, err := sv.CallSessionsRepository.FindByID(sessionID)
	if err != nil || session == nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "running" {
		return nil // Session has been stopped or paused
	}

	// 4. Update with actual Botnoi call identifier
	_ = sv.RecordsRepository.LinkCallID(recordID, botnoiCallID)
}
```




