# LTI Core API, Models, and Events

This document describes the current platform surface:
- HTTP endpoints
- request/response models
- AGS behavior
- enrollment synchronization event contracts (NRPS -> outbox -> Kafka -> consumer inbox)

## Base URL

- Local default: `http://localhost:8080`
- API prefix: `/lti`

---

## 1) HTTP Endpoints

## 1.1 Auth / Registration

- `GET /lti/auth/register`
  - Query params:
    - `openid_configuration` (required)
    - `registration_token` (required)
  - Behavior:
    - fetches OpenID config from platform
    - sends LTI Dynamic Registration
    - persists platform metadata
  - Response:
    - `200` HTML snippet with postMessage close instruction
    - `400` if required params missing
    - `500` on registration failure

- `POST /lti/auth/login`
  - Form/query model: `LoginRequest`
  - Required fields in practice:
    - `iss`
    - `client_id`
    - `login_hint`
    - `target_link_uri` (legacy alias `target_ling_uri` is accepted in adapter logic)
  - Behavior:
    - creates nonce/state session
    - redirects to platform authorization endpoint
  - Response:
    - `302` redirect
    - `400` for validation/platform lookup errors

---

## 1.2 LTI Launch

- `POST /lti/launch`
  - Form fields:
    - `id_token`
    - `state`
  - Behavior:
    - validates session and JWT
    - resolves platform and JWKS
    - stores deep linking session
    - starts/resumes enrollment sync run when NRPS claim exists
    - redirects to deep link selection page
  - Response:
    - `303` redirect to `deeplink/select?session_id=...`
    - `400` on launch validation failure

---

## 1.3 JWKS

- `GET /lti/.well-known/jwks.json`
  - Response:
    - `200` JWKS payload

---

## 1.4 Deep Linking

- `GET /lti/deeplink/select`
  - Query: `session_id`
  - Returns HTML selection page from template.

- `POST /lti/deeplink/return`
  - Reads `dl_session` cookie
  - Accepts selected items JSON
  - Builds and signs deep link JWT response
  - Returns JSON for AJAX clients or HTML auto-submit page

- `POST /lti/deeplink/api/return`
  - Body: `{ "jwt": "..." }`
  - Echo-style success API wrapper

- `GET /lti/deeplink/content`
  - Query:
    - `session_id` (required)
    - `types` (optional CSV filter)
  - Returns generated content catalog for selection UI.

- `POST /lti/deeplink/lineitem`
  - Creates line item from selection flow.

- `GET /lti/deeplink/cancel`
  - Cancels deep linking session.

---

## 1.5 AGS (Assignment and Grade Services)

- `GET /lti/ags/lineitems`
  - Query: `context_id` (required)
  - Response:
    - `200 { "lineitems": [LineItem...] }`
    - `400` on validation/repo error

- `POST /lti/ags/lineitems`
  - Body: `LineItem`
  - Validation:
    - `label` required
    - `context_id` required
    - `max_score > 0`
  - Persistence:
    - upsert into `ags_line_items`
    - if `id` omitted, UUID generated
  - Response:
    - `201` with stored line item model (includes generated `id`)
    - `400` on validation/error

- `GET /lti/ags/score`
  - Query:
    - `lineitem_id` (required, UUID)
    - `user_id` (required)
  - Response:
    - `200 Score`
    - `400` on validation/error

- `POST /lti/ags/score`
  - Body: `Score`
  - Validation:
    - `lineitem_id` required and UUID
    - `user_id` required
    - `score >= 0`
    - score must be within line item range: `0 <= score <= line_item.max_score`
  - Persistence:
    - upsert into `ags_scores` by `(lineitem_id, user_id)`
  - Response:
    - `201` saved score
    - `400` on validation/error

---

## 2) Domain/API Models

## 2.1 `LineItem`

```json
{
  "id": "uuid-string",
  "label": "Assignment 1",
  "max_score": 100,
  "context_id": "course-123",
  "start_date_time": "2026-05-01T00:00:00Z",
  "end_date_time": "2026-05-31T23:59:59Z"
}
```

## 2.2 `Score`

```json
{
  "lineitem_id": "uuid-string",
  "user_id": "user-123",
  "score": 87.5
}
```

## 2.3 `IncomingEvent` (event envelope)

```json
{
  "id": "uuid",
  "aggregate_type": "RosterSyncRun",
  "aggregate_id": "uuid",
  "event_type": "UserSynchronizationBatchV1",
  "payload": {},
  "created_at": "2026-05-07T10:00:00Z",
  "processed_at": null,
  "version": 42
}
```

## 2.4 `UserSynchronizationPayloadV1`

```json
{
  "lms": {
    "issuer": "https://canvas.example.com",
    "course_id": "course-123"
  },
  "batch": {
    "index": 1,
    "total": 0
  },
  "users": [
    {
      "sub": "1b2c3d",
      "email": "a@b.com",
      "display_name": "Ann",
      "avatar_url": ""
    }
  ]
}
```

---

## 3) AGS Storage Model

## 3.1 Table `ags_line_items`

- PK: `lineitem_id` (uuid)
- Indexed by: `(context_id, created_at desc)`
- Stores assignment metadata for a context.

## 3.2 Table `ags_scores`

- PK: `(lineitem_id, user_id)`
- FK: `lineitem_id -> ags_line_items.lineitem_id` (`ON DELETE CASCADE`)
- Stores latest score per learner per line item.

---

## 4) Enrollment Sync Event Pipeline

## 4.1 Producer Side

1. Launch receives NRPS claim.
2. Service starts/resumes `roster_sync_runs` aggregate.
3. Worker fetches NRPS pages (`GetMembersPage`).
4. Members are batched (`batch_size`, default 100).
5. For each batch:
   - deterministic event ID is generated from `(run_id, batch_index)`
   - event saved into `outbox_events`
   - run cursor/version advanced in same DB transaction

## 4.2 Kafka Publication

- Worker claims pending outbox rows (lease-based locking).
- Publishes to Kafka topic from config (`users.synchronization.v1` by default).
- Marks event processed on ack.
- Retries transient failures, with attempt counters.

## 4.3 Consumer Side (Exactly-Once Effect)

- Consumer reads Kafka message.
- Applies event in one DB transaction:
  - insert `inbox_events(event_id)` with unique key
  - if duplicate: no-op (already applied)
  - else upsert `lms_users` and `lms_course_enrollments`
- Commits Kafka offset only after successful DB transaction.

---

## 5) Consistency and Ordering Guarantees

- Producer write path (`roster_sync_runs` + `outbox_events`) is transactional.
- Outbox -> Kafka is at-least-once delivery.
- Consumer applies exactly-once effect via inbox dedup (`event_id` unique).
- Partition key is tied to sync aggregate ID to keep per-run batch ordering.

---

## 6) Config Flags

Key sections in `boot.yaml`:

- `kafka.*`
  - `enabled`
  - `brokers`
  - `topic`
  - `group_id`
- `enrollment_sync.*`
  - `batch_size`
  - `outbox_batch_size`
  - `outbox_poll_interval`
  - `outbox_max_attempts`
  - `consumer_enabled`

---

## 7) Operational Notes

- Apply schema changes before startup.
- AGS `lineitem_id` must be UUID in score operations.
- NRPS pagination support depends on LMS response (`next`, `next_page`, or `links.next.href`).
- `go mod tidy` and `go test ./...` should be run in an environment where Go toolchain is installed.
