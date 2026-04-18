[Русская версия](./README.md)

# Room Booking Service

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![chi](https://img.shields.io/badge/chi-v5-5C2D91)](https://github.com/go-chi/chi)
[![JWT](https://img.shields.io/badge/JWT-v5-F4A100?logo=jsonwebtokens&logoColor=white)](https://github.com/golang-jwt/jwt)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![pgx](https://img.shields.io/badge/pgx-v5-4169E1)](https://github.com/jackc/pgx)
[![golang-migrate](https://img.shields.io/badge/golang--migrate-v4-0078D4)](https://github.com/golang-migrate/migrate)
[![testify](https://img.shields.io/badge/testify-v1-4CAF50)](https://github.com/stretchr/testify)
[![mockery](https://img.shields.io/badge/mockery-v2-9E9E9E)](https://github.com/vektra/mockery)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)](./docker-compose.yaml)
[![Swagger](https://img.shields.io/badge/API-Swagger-85EA2D?logo=swagger&logoColor=black)](http://localhost:8080/swagger/)

An HTTP service for booking conference rooms inside a company. Admins create rooms and configure their availability schedules; the system automatically generates 30-minute slots from the schedule. Employees browse available slots, create and cancel bookings.

---

## Quick Start

```bash
docker-compose up -d --build
```

Service: `http://localhost:8080` | Swagger UI: `http://localhost:8080/swagger/`

---

## Usage

```bash
make up               # start service with all dependencies
make down             # stop and remove containers
make seed             # seed the database with test data
make test             # all tests
make test-unit        # unit tests only (no DB required)
make test-e2e         # E2E tests via docker-compose
make coverage         # unit test coverage (82%)
make swagger          # regenerate Swagger docs
make mock             # regenerate mocks
make lint             # run linters
```

---

## API

Spec: [api.yaml](./api.yaml) | Docs: [Swagger UI](http://localhost:8080/swagger/)

| Method | Path                                   | Access      | Description                          |
|--------|----------------------------------------|-------------|--------------------------------------|
| GET    | `/_info`                               | public      | Health check                         |
| POST   | `/dummyLogin`                          | public      | Get test JWT by role                 |
| POST   | `/register`                            | public      | Register with email/password         |
| POST   | `/login`                               | public      | Sign in with email/password          |
| GET    | `/rooms/list`                          | admin, user | List rooms                           |
| POST   | `/rooms/create`                        | admin       | Create a room                        |
| POST   | `/rooms/{roomId}/schedule/create`      | admin       | Create a schedule for a room         |
| GET    | `/rooms/{roomId}/slots/list`           | admin, user | Available slots by date              |
| POST   | `/bookings/create`                     | user        | Create a booking                     |
| GET    | `/bookings/my`                         | user        | My upcoming bookings                 |
| POST   | `/bookings/{bookingId}/cancel`         | user        | Cancel a booking                     |
| GET    | `/bookings/list`                       | admin       | All bookings with pagination         |

---

## Target Scale

The service is designed for a mid-size company with the following load profile:

| Parameter                             | Value      |
|---------------------------------------|------------|
| Conference rooms                      | up to 50   |
| Employees                             | up to 10,000 |
| Slots per day                         | up to 1,000 |
| Active bookings                       | up to 100,000 |
| RPS                                   | 100        |
| Success SLI                           | 99.9%      |
| Response time p99 SLI (slots endpoint)| 200 ms     |
| Slot requests within next 7 days      | ~99.9%     |

The hottest endpoint is `GET /rooms/{roomId}/slots/list` — storage layout and indexes are optimised for it first. This load profile shaped the hybrid slot generation strategy, the lazy generation horizon, and the connection pool configuration.

---

## Design Decisions

### Slot Generation

A **hybrid approach** (Eager + Lazy generation) was chosen:

- **Eager generation:** when a schedule is created, the system immediately generates all slots for the **next 7 days** and persists them to the DB. Slot queries within this range are pure `SELECT` operations with no computation.
- **Lazy generation:** when slots are requested for a date **beyond 7 days**, they are computed and saved on-demand at query time. Lazy generation is capped at a **90-day** horizon from the current date.

> **Rationale:**
>
> **99.9%** of requests fall within the nearest 7 days — slots are already in the DB, so `GET /rooms/{roomId}/slots/list` is a pure indexed `SELECT` on `(room_id, start_time)`.
> Eager generation for only 7 days minimises DB load when a schedule is created. Lazy generation for distant dates ensures correctness without a background job.
>
> The 90-day cap on lazy generation is a deliberate architectural decision.
> Without it, a request for an arbitrary future date (e.g. year 2099) would generate and persist slots, causing unbounded DB growth.
> 90 days was chosen as a reasonable business planning horizon: booking a meeting room more than 3 months in advance is not a typical corporate use case.

### Conference Link on External Service Failure

When `createConferenceLink: true` is passed, the system calls the Conference Service before saving the booking to the DB. If a link is returned, it is included in the single `INSERT` together with the booking.

**Failure scenarios modelled:**

- **Scenario 1: Conference Service unavailable (network error, timeout)**
  The service returns an error; the booking is created without a link (`conferenceLink = null`); the error is logged via `slog.Warn`. Booking is not blocked — an auxiliary service being unavailable must not block a critical operation.

- **Scenario 2: Conference Service returned a link, then the booking `INSERT` failed**
  An **idempotency key** approach is implemented: the stub stores a `bookingID → link` mapping, so a retry with the same `bookingID` returns the same link. Retrying the `INSERT` after a failure is safe.

### Booking Cancellation Idempotency

A repeated `POST /bookings/{bookingId}/cancel` on an already-cancelled booking returns `200 OK` with the current state (`status: cancelled`).

### Environment Variables

| Variable                        | Default | Description                    |
|---------------------------------|---------|--------------------------------|
| `DATABASE_URL`                  | —       | PostgreSQL DSN                 |
| `JWT_SECRET`                    | —       | JWT signing secret             |
| `JWT_EXPIRATION_HOURS`          | `24`    | Token lifetime (hours)         |
| `PORT`                          | `8080`  | Service port                   |
| `DB_MAX_CONNS`                  | `20`    | Max DB connections             |
| `DB_MIN_CONNS`                  | `5`     | Min DB connections             |
| `DB_MAX_CONN_LIFETIME_MINUTES`  | `60`    | Max connection lifetime        |
| `DB_MAX_CONN_IDLE_TIME_MINUTES` | `10`    | Idle connection timeout        |

Copy `.env.template` to `.env` and fill in the values before running.

---

## Load Testing

Tool: Apache Bench (`ab`), 30 seconds per endpoint.
Full report: [LOAD_TEST_REPORT.md](LOAD_TEST_REPORT.md)

DB state: 10,002 users · 50 rooms · 206,000 slots (~1,000/day) · 101,581 bookings.

| Endpoint                   | Concurrency |        RPS |    p50 |    p99 |
|----------------------------|------------:|-----------:|-------:|-------:|
| POST /dummyLogin           |         200 | **13,136** |  15 ms |  32 ms |
| POST /login                |          50 |    **151** | 325 ms | 542 ms |
| POST /register             |          50 |    **150** | 324 ms | 551 ms |
| GET /_info                 |         200 | **13,502** |  15 ms |  34 ms |
| GET /rooms/list            |         150 |  **7,627** |  19 ms |  35 ms |
| GET /slots/list            |         100 |  **7,010** |  14 ms |  36 ms |
| GET /bookings/my           |         150 |  **7,323** |  20 ms |  38 ms |
| GET /bookings/list         |          50 |  **1,672** |  29 ms |  52 ms |
| POST /bookings/create      |         100 |  **9,082** |  10 ms |  28 ms |
| POST /bookings/{id}/cancel |         100 | **11,716** |   8 ms |  17 ms |

`/login` and `/register` are CPU-bound: `bcrypt.DefaultCost`.
