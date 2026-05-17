# Audience Export Lambda — Project Guide

> **For the learner:** This file is your roadmap. Hand it to Claude in VS Code and ask it to mentor you milestone by milestone. **Do not ask Claude to write the whole thing.** Work each milestone, share your code, get feedback, then move on.
>
> **Mentor instructions (for Claude in VS Code):**
> - Reveal ONE milestone at a time.
> - Never paste the full solution. Give hints, then partial hints, then a tiny worked example only if the learner is stuck after 2+ tries.
> - After the learner shares code, review: what works, what to improve, why it matters — then move on.
> - Explain new concepts (context, interfaces, dependency injection, `database/sql`, AWS SDK v2, etc.) **just in time**, not up front.

---

## 1. What you're building

A Go AWS Lambda that:

1. Reads **audience** rows from Table 1.
2. Joins to a **product** table for product name and price.
3. Joins to a **business** table for retailer code and retailer name.
4. Flattens each row into this CSV shape:

   ```
   project_id, retailer_code, retailer_name, product_flagging, product_1_name, product_1_price
   ```

5. Writes the CSV to **S3**, overwriting the previous file at the same key.
6. Does the heavy work **in the background** (caller gets `202 Accepted` and a job ID immediately).
7. Logs each run's status (`PENDING → RUNNING → SUCCESS | FAILED`) to a tracking table.

You will run this **locally** using AWS SAM CLI + LocalStack (S3 / Lambda) + Postgres in Docker. No real AWS account needed.

---

## 2. Architecture decisions (read before coding)

### 2.1 How "background" works on Lambda

Lambda has a 15-min hard timeout, and the caller is usually waiting. "Background" means **return fast, do slow work elsewhere.**

You'll use **async self-invoke**:

1. Sync invocation arrives. Handler validates input, generates a `job_id`, writes a `PENDING` row to the status table.
2. Handler calls `lambda.Invoke` on *itself* with `InvocationType=Event` (fire-and-forget), passing job_id and project_id.
3. Handler returns `202 Accepted` with `{ "job_id": "..." }`.
4. The async invocation hits the same handler — but the event shape is different. Your code routes it to the **worker** path: mark `RUNNING`, run the export, write CSV to S3, mark `SUCCESS` or `FAILED`.

This teaches: event-type dispatch (the **factory** pattern in your structure), clean separation of "trigger" vs "work", and status tracking — in one Lambda.

### 2.2 Status table

```sql
CREATE TABLE export_jobs (
    job_id          UUID PRIMARY KEY,
    project_id      VARCHAR NOT NULL,
    status          VARCHAR NOT NULL,        -- PENDING | RUNNING | SUCCESS | FAILED
    s3_key          VARCHAR,
    row_count       INT,
    error_message   TEXT,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);
```

Every status transition is a DB update. If the Lambda crashes mid-export the row stays in `RUNNING` — note that as a known gap; a sweeper job is a future enhancement.

### 2.3 The SQL join (same DB)

```sql
SELECT
    a.project_id,
    b.retailer_code,
    b.retailer_name,
    a.product_flagging,
    p.name  AS product_1_name,
    p.price AS product_1_price
FROM audience a
JOIN product  p ON p.id = a.product_id
JOIN business b ON b.id = a.business_id
WHERE a.project_id = $1;
```

Index `audience.project_id`, `product.id`, `business.id`.

### 2.4 S3 "replace"

S3 `PutObject` to the **same key** overwrites the previous object — that's the replace. Use a stable key like `exports/project-{project_id}/audience.csv`. If versioning is on, old versions are retained automatically.

### 2.5 CSV: in-memory vs streaming

Start simple: build the CSV in a `bytes.Buffer`, then upload. If volume grows later, swap to `io.Pipe` so rows stream SQL → CSV writer → S3 multipart upload with bounded memory. Do the simple version first.

---

## 3. Project structure

Don't pre-create empty folders — add each one when its milestone arrives. Final tree:

```
.
├── cmd/
│   └── lambda/
│       └── main.go                    # entry point, wires deps, calls lambda.Start
├── config/
│   └── config.go                      # env vars → typed struct, validated at startup
├── core/
│   ├── aws/
│   │   ├── s3.go                      # S3 client wrapper (uploader)
│   │   └── lambda.go                  # Lambda Invoke client (for self-invoke)
│   ├── context/
│   │   └── context.go                 # request-scoped helpers (job_id in ctx)
│   ├── factory/
│   │   └── factory.go                 # builds all deps once at cold start
│   ├── logger/
│   │   └── logger.go                  # structured logger (slog), job_id-aware
│   └── vars/
│       └── vars.go                    # shared constants (status values, defaults)
├── errors/
│   └── errors.go                      # typed errors (ErrInvalidInput, etc.)
├── handler/
│   ├── handler.go                     # the Lambda handler, dispatches by event type
│   ├── sync.go                        # sync path: validate, enqueue, return 202
│   └── async.go                       # async path: run the export
├── modules/
│   ├── dto/
│   │   ├── request.go                 # incoming sync request shape
│   │   ├── async_event.go             # internal async event payload
│   │   └── response.go                # outgoing response shape
│   ├── entity/
│   │   ├── audience_row.go            # the joined SQL row (one struct)
│   │   └── export_job.go              # the status table entity
│   ├── repository/
│   │   ├── audience_repo.go           # SELECT joined rows for a project_id
│   │   └── job_repo.go                # INSERT / UPDATE export_jobs
│   └── usecase/
│       ├── enqueue_export.go          # validate → write PENDING → self-invoke → 202
│       └── run_export.go              # mark RUNNING → query → CSV → S3 → mark SUCCESS
├── deployments/
│   ├── template.yaml                  # AWS SAM template (for `sam local invoke`)
│   ├── docker-compose.yml             # postgres + localstack
│   └── init.sql                       # schema + sample seed data
├── events/
│   ├── sync.json                      # sample sync event
│   └── async.json                     # sample async event
├── go.mod
├── go.sum
└── README.md
```

### Why each layer exists (one line each)

- **config** — fail at startup if env is wrong; never read env from deep inside code.
- **core/aws** — wrap the SDK so you can mock it in tests and swap clients.
- **core/context** — carry job_id and logger through every layer without polluting signatures.
- **core/factory** — build everything ONCE at cold start; reuse across invocations.
- **core/logger** — structured (JSON) logs with job_id on every line.
- **core/vars** — magic strings (status values, env keys) live here.
- **errors** — typed errors so callers can branch with `errors.Is`.
- **handler** — thin: parse event, call usecase, format response. No business logic.
- **modules/dto** — wire format (JSON in/out). Never leaks to the DB layer.
- **modules/entity** — domain shapes. Don't tie to JSON tags or DB tags.
- **modules/repository** — SQL lives here, nowhere else.
- **modules/usecase** — the *business* steps. Orchestrates repos + AWS. The interesting part.

---

## 4. Local toolchain

- **Go 1.22+**
- **Docker + docker-compose** — Postgres + LocalStack.
- **AWS SAM CLI** — `sam local invoke` runs your Lambda in a container that mimics the real runtime.
- **LocalStack** — emulates S3 and Lambda Invoke locally; SDK points at `http://localhost:4566`.
- **awslocal** (optional, `pip install awscli-local`) — `aws` CLI pointed at LocalStack.

Install each when its milestone needs it. Don't front-load.

---

## 5. Milestones

10 milestones. Don't peek ahead. After each, share your code with the mentor for review.

### Milestone 1 — Project skeleton & config

**Goal:** `go mod init`, create the folder skeleton (empty `.go` files are fine), implement `config/config.go` that loads env vars into a struct.

**You'll learn:** Go modules, `os.Getenv`, why config validation belongs at startup, struct tags.

**Env vars:**
- `DB_DSN` — Postgres connection string
- `AWS_REGION`
- `AWS_ENDPOINT_URL` — LocalStack endpoint (`http://localhost:4566`)
- `S3_BUCKET`
- `LAMBDA_FUNCTION_NAME` — for self-invoke
- `LOG_LEVEL`

Loader returns an error if any required var is missing.

---

### Milestone 2 — Logger, errors, vars, context helpers

**Goal:** Implement `core/logger` (`log/slog`), `core/vars` (status constants), `errors/errors.go` (typed errors like `ErrInvalidInput`, `ErrJobNotFound`), and `core/context` (helpers to put/get `job_id` from `context.Context`).

**You'll learn:** `context.Context`, `log/slog`, custom error types with `errors.Is`, why constants live separately.

---

### Milestone 3 — docker-compose + schema + seed data

**Goal:** Write `deployments/docker-compose.yml` to run Postgres and LocalStack. Write `deployments/init.sql` with three source tables (audience, product, business), the `export_jobs` table, and ~10 sample rows that join cleanly.

**You'll learn:** Docker networking basics, seeding a DB on container start, how LocalStack exposes AWS services.

After this, `docker compose up -d` gives you a working local Postgres and a LocalStack endpoint. Create the S3 bucket with `awslocal s3 mb s3://your-bucket`.

---

### Milestone 4 — Entities & DTOs

**Goal:** Define the structs:
- `entity.AudienceRow` — one joined row (the 6 fields).
- `entity.ExportJob` — matches the `export_jobs` table.
- `dto.EnqueueRequest` — incoming sync request (`project_id`).
- `dto.AsyncEvent` — internal payload to self (`job_id`, `project_id`, plus a discriminator like `"type": "async_export"`).
- `dto.EnqueueResponse` — `{ job_id, status }`.

**You'll learn:** entity ≠ DTO, JSON struct tags, "DTOs at the edges, entities in the core".

---

### Milestone 5 — Repository layer

**Goal:** Implement `repository.AudienceRepo` with `FindByProjectID(ctx, projectID) ([]entity.AudienceRow, error)`, and `repository.JobRepo` with `Create`, `MarkRunning`, `MarkSuccess`, `MarkFailed`.

**You'll learn:** `database/sql`, `*sql.DB` is a pool not a connection, parameterized queries (NEVER string-concat), `rows.Scan`, `defer rows.Close()`.

Driver: `github.com/jackc/pgx/v5/stdlib` — modern Postgres driver that plugs into `database/sql`.

---

### Milestone 6 — AWS clients (S3 + Lambda)

**Goal:** `core/aws/s3.go` exposes `Upload(ctx, key string, body io.Reader) error`. `core/aws/lambda.go` exposes `InvokeAsync(ctx, functionName string, payload []byte) error`.

**You'll learn:** AWS SDK v2 (`github.com/aws/aws-sdk-go-v2`), pointing the SDK at LocalStack via a custom endpoint resolver, the difference between `InvocationType=Event` (async) and `RequestResponse` (sync).

---

### Milestone 7 — Factory & wiring

**Goal:** `core/factory/factory.go` builds and holds: config, logger, `*sql.DB`, S3 client, Lambda client, both repos, both usecases. Called once from `main.go` before `lambda.Start`.

**You'll learn:** manual dependency injection (no framework needed), why cold-start init belongs in `main`, struct-held deps vs. package-level singletons.

---

### Milestone 8 — Usecase 1: EnqueueExport (sync path)

**Goal:** Implement `usecase.EnqueueExport`. Steps:
1. Validate `project_id` (non-empty).
2. Generate `job_id` (UUID — `github.com/google/uuid`).
3. `JobRepo.Create(PENDING)`.
4. Marshal an `AsyncEvent`, call `LambdaClient.InvokeAsync`.
5. Return `{ job_id, status: PENDING }`.

**You'll learn:** how a usecase reads — it orchestrates, it doesn't do SQL or HTTP itself.

---

### Milestone 9 — Usecase 2: RunExport (async path)

**Goal:** Implement `usecase.RunExport`. Steps:
1. `JobRepo.MarkRunning(job_id)`.
2. `AudienceRepo.FindByProjectID(project_id)`.
3. Build CSV in a `bytes.Buffer` using `encoding/csv` — header row + data rows.
4. `S3Client.Upload(key="exports/project-{project_id}/audience.csv", body)`.
5. `JobRepo.MarkSuccess(job_id, s3_key, row_count)`.
6. On any error: `JobRepo.MarkFailed(job_id, err.Error())` and return the error.

**You'll learn:** `encoding/csv`, `bytes.Buffer` as `io.Reader`, error handling that *also* records state (defer + named return is a clean pattern here).

---

### Milestone 10 — Handler & main.go

**Goal:** Handler receives the event and dispatches:
- If event has `type: "async_export"` → route to `RunExport`.
- Otherwise → treat as `EnqueueRequest` → route to `EnqueueExport`.

`cmd/lambda/main.go`:
1. Build factory.
2. `lambda.Start(handler.Handle)`.

Write `deployments/template.yaml` (SAM template) so you can run:

```
sam local invoke -e events/sync.json
```

Provide `events/sync.json` (`{"project_id": "p-123"}`) and `events/async.json` (`{"type":"async_export","job_id":"...","project_id":"p-123"}`). Test both paths.

**You'll learn:** `github.com/aws/aws-lambda-go/lambda`, how SAM local runs your binary in a container, event routing patterns.

---

## 6. Done state — how you know it works

1. `docker compose up -d` brings up Postgres + LocalStack.
2. `awslocal s3 ls` shows your bucket exists.
3. `sam local invoke -e events/sync.json` returns a `job_id` in <1s.
4. A few seconds later: `awslocal s3 cp s3://your-bucket/exports/project-p-123/audience.csv -` prints the CSV.
5. `psql ... -c "SELECT * FROM export_jobs"` shows the row went `PENDING → RUNNING → SUCCESS` with `row_count` populated.
6. Re-running the sync invoke replaces the S3 file in place (same key, new content).
7. Break the SQL on purpose → job ends as `FAILED` with the error message stored.

---

## 7. After you finish — stretch goals

- Switch CSV build to `io.Pipe` streaming.
- Add a "stale job sweeper" Lambda that flips `RUNNING` jobs older than 20 minutes to `FAILED`.
- Replace async self-invoke with SQS + DLQ.
- Add unit tests for usecases using mock repos (good interface-design practice).
- Add structured logging context propagation (every log line gets `job_id`).

---

## 8. Rules for the mentor session

When you give this file to Claude in VS Code:

1. Say: *"Be my mentor for this project. Go one milestone at a time. Don't write the code for me — guide with hints. Wait for me to share my code after each milestone before continuing."*
2. After each milestone, paste your code and ask for review.
3. If stuck, ask for "a stronger hint" before asking for the answer.
4. Don't skip the review — that's where the learning compounds.

Good luck. Have fun.
