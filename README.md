# AeroXe ApprovalFlow

> Phase 1 | AeroXe Ecosystem

---

## Table of Contents

- [What Is This Project?](#what-is-this-project)
- [Why Was It Built?](#why-was-it-built)
- [When Should You Use It?](#when-should-you-use-it)
- [Where Does It Run?](#where-does-it-run)
- [How Does It Work?](#how-does-it-work)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Service Modules](#service-modules)
- [Saga Orchestrator](#saga-orchestrator)
- [WebSocket Events](#websocket-events)
- [Database Schema](#database-schema)
- [Setup & Installation](#setup--installation)
- [Environment Variables](#environment-variables)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)

---

## What Is This Project?

AeroXe ApprovalFlow is a digital approval and workflow automation platform that manages application processing, multi-level approvals, escalation workflows, and decision notification for organizational processes.

---

## Why Was It Built?

Manual approval processes are slow, lack transparency, and create bottlenecks. ApprovalFlow digitizes approvals with automated routing, escalation rules, and complete audit trails, reducing approval cycle time by 70%.

---

## When Should You Use It?

Use ApprovalFlow for any organization with structured approval processes — government permits, corporate expense approvals, HR onboarding workflows, or procurement approvals.

---

## Where Does It Run?

Backend runs Go/Hertz with PostgreSQL for workflow and approval data, Redis for approval state caching, NATS for approval notification events. React web portal for approvers. WebSocket pushes real-time approval requests and decision notifications.

---

## How Does It Work?

ApprovalFlow implements ApprovalService, ApplicationService, WorkflowService, NotificationService, TemplateService, EscalationService, AnalyticsService, and ReportService. ApplicationSubmission saga handles: submission → validation → routing → approval/rejection → notification. EscalationProcess saga monitors SLA timers and escalates overdue approvals.

---

## Architecture

### High-Level Architecture

```
+---------------------------------------------------------------------+
|                          CLIENT LAYER                               |
|  +----------+  +------------------+  +-----------------------+     |
|  | React    |  | Android          |  | iOS                   |     |
|  | (Web)    |  | (Kotlin+Compose) |  | (SwiftUI)             |     |
|  +----+-----+  +--------+---------+  +----------+------------+     |
|       |                 |                        |                  |
+-------+-----------------+------------------------+------------------+
        |                 |                        |
        v                 v                        v
+---------------------------------------------------------------------+
|                       API GATEWAY (Hertz)                           |
|  +-------------+  +--------------+  +------------------------+     |
|  | HTTP REST   |  | gRPC Proxy   |  | WebSocket Hub          |     |
|  | Routes      |  | (grpc-gw)    |  | (coder/websocket)      |     |
|  +------+------+  +------+-------+  +----------+------------+     |
|         |                |                       |                  |
|  +------v----------------v-----------------------v----------+      |
|  |  Auth | Rate Limit | Circuit Breaker | Logging           |      |
|  +----------------------------------------------------------+      |
+-------------+-------------------+------------------+---------------+
              |                   |                  |
     +--------v--------+  +------v------+  +--------v--------+
     |  gRPC (sync)    |  |  NATS (async)|  |  WebSocket      |
     |  point-to-point |  |  pub/sub     |  |  real-time      |
     +--------+--------+  +------+-------+  +--------+--------+
              |                   |                  |
+-------------v-------------------v------------------v---------------+
|                  MODULAR MONOLITH BACKEND                          |
|  +------------+ +------------+ +------------+ +------------+     |
|  | Module A   | | Module B   | | Module C   | | Module D   |     |
|  +-----+------+ +-----+------+ +-----+------+ +-----+------+     |
|        |              |              |              |              |
|  +-----v--------------v--------------v--------------v------+      |
|  |  +----------+ +-------+ +--------+ +--------------+    |      |
|  |  | Postgres | | Redis | |  NATS  | | Saga Engine  |    |      |
|  |  +----------+ +-------+ +--------+ +--------------+    |      |
|  +---------------------------------------------------------+      |
+--------------------------------------------------------------------+
```

### Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **HTTP Framework** | [Hertz](https://github.com/cloudwego/hertz) | High-performance HTTP server |
| **Messaging** | [NATS](https://nats.io/) JetStream | Durable async event-driven messaging |
| **Database** | [PostgreSQL](https://www.postgresql.org/) 15+ | Primary data store |
| **Cache** | [Redis](https://redis.io/) 7+ | Caching, sessions, saga state |
| **WebSocket** | In-repo RFC 6455 implementation | Real-time communication (ping/pong keepalive, close handshake) |
| **Frontend** | React 18 + TypeScript + Tailwind + Vite | Web application (served by nginx in production) |

> **Roadmap (not in this phase):** gRPC service-to-service transport, and native
> Android (Kotlin + Jetpack Compose) / iOS (SwiftUI) clients.

---

## Service Modules

| Module | Description | Protocol |
|--------|-------------|----------|
| ApprovalService | Core approval operations | gRPC + NATS |
| ApplicationService | Core application operations | gRPC + NATS |
| WorkflowService | Core workflow operations | gRPC + NATS |
| NotificationService | Core notification operations | gRPC + NATS |
| TemplateService | Core template operations | gRPC + NATS |
| EscalationService | Core escalation operations | gRPC + NATS |
| AnalyticsService | Core analytics operations | gRPC + NATS |
| ReportService | Core report operations | gRPC + NATS |

---

## Saga Orchestrator

| Saga | Pattern |
|------|---------|
| ApplicationSubmission | Orchestrated via NATS + Redis state |
| ApprovalRouting | Orchestrated via NATS + Redis state |
| EscalationProcess | Orchestrated via NATS + Redis state |
| DecisionNotification | Orchestrated via NATS + Redis state |

---

## WebSocket Events

| Event | Description |
|-------|-------------|
| `application_submitted` | Real-time updates for application submitted |
| `approval_needed` | Real-time updates for approval needed |
| `decision_made` | Real-time updates for decision made |
| `escalation_trigger` | Real-time updates for escalation trigger |

> **WebSocket authentication:** the `/ws` endpoint requires an
> `Authorization: Bearer <token>` header and derives the user identity from the
> signed JWT — client-supplied user IDs are ignored. Connections from an
> `Origin` outside `CORS_ALLOWED_ORIGINS` are rejected. Clients must send the
> token header on the WebSocket upgrade request.

---

## Database Schema

| Table | Description | Key Fields |
|-------|-------------|------------|
| Approval | `approvals` table | UUID, timestamps, soft delete |
| Application | `applications` table | UUID, timestamps, soft delete |
| Workflow | `workflows` table | UUID, timestamps, soft delete |
| Notification | `notifications` table | UUID, timestamps, soft delete |
| Template | `templates` table | UUID, timestamps, soft delete |
| Escalation | `escalations` table | UUID, timestamps, soft delete |
| Status | `statuss` table | UUID, timestamps, soft delete |
| Comment | `comments` table | UUID, timestamps, soft delete |
| Document | `documents` table | UUID, timestamps, soft delete |
| AuditLog | `auditlogs` table | UUID, timestamps, soft delete |

### Redis Usage

| Key Pattern | Purpose | TTL |
|------------|---------|-----|
| `session:<user_id>` | User session | 24h |
| `cache:<slug>:<id>` | Entity cache | 15m |
| `saga:<saga_id>` | Saga state | Until completion |
| `ratelimit:<ip>` | Rate limiting | 1m |

---

## Setup & Installation

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Redis 7+
- NATS Server 2.10+ (with JetStream)
- Node.js 18+ (for React)
- Docker & Docker Compose (optional)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/aeroxe/approval-flow.git
cd approval-flow

# Create your environment file (set a strong JWT_SECRET and ADMIN_PASSWORD)
cp backend/.env.example backend/.env

# Start infrastructure services + backend + web UI (nginx-served SPA)
cd backend
docker compose up -d postgres redis nats approval-flow web
#   → API at http://localhost:8080, SPA at http://localhost:3000

# Or run the backend directly (migrations run automatically at startup):
make run

# In another terminal - start the React frontend dev server
cd ../frontend
npm install
npm run dev   # → http://localhost:5173 (proxies /api and /ws to :8080)
```

> **Security note:** The initial administrator account is no longer seeded with
> a default password. It is created on startup from the `ADMIN_EMAIL` and
> `ADMIN_PASSWORD` environment variables. In production (`ENV=production`)
> both are required, `JWT_SECRET` must be at least 32 characters, and the
> default `DATABASE_URL` is rejected.
>
> **Migrations:** the server applies `MIGRATIONS_PATH` SQL files via
> golang-migrate and **fails to start if they error** — it never silently falls
> back to AutoMigrate when migration files are present (AutoMigrate is used
> only when no SQL files exist, i.e. local development).

### Docker Compose

The full stack (PostgreSQL, Redis, NATS with JetStream, the backend, and the
nginx-served frontend) is defined in [`backend/docker-compose.yml`](backend/docker-compose.yml):

```bash
cd backend
cp .env.example .env     # set a real JWT_SECRET, ADMIN_EMAIL, ADMIN_PASSWORD
# Infra + backend only:
docker compose up -d postgres redis nats approval-flow
# Full stack incl. the web UI:
docker compose up -d
```

> NATS must run with JetStream enabled (`-js`), which the compose file does.
> The backend fails to start if the durable event stream cannot be created.

---

## Environment Variables

```bash
# Server
SERVER_PORT=8080
ENV=development
LOG_LEVEL=debug

# PostgreSQL
DATABASE_URL=postgres://aeroxe:secret@localhost:5432/approval-flow?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# NATS (JetStream required)
NATS_URL=nats://localhost:4222

# JWT
JWT_SECRET=your-secret-key          # ≥32 chars in production
JWT_EXPIRY=24h                      # access token lifetime (refresh: 7 days)

# WebSocket
WS_MAX_CONNECTIONS=1000
WS_PING_INTERVAL=30

# Initial admin account (created at startup if it does not exist)
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=change-me

# Rate limiting (per-IP fixed window)
RATE_LIMIT_RPS=100                  # general API allowance per window
RATE_LIMIT_BURST=20                 # auth endpoints (login/register/refresh)
RATE_LIMIT_WINDOW=60

# Trusted proxies: CIDRs whose X-Forwarded-For is honored for ClientIP().
# Empty = trust none (spoof-proof, but behind a proxy you must set this).
TRUSTED_PROXIES=

# Swagger UI (/docs): default true in dev, false in production.
SWAGGER_ENABLED=

# Bearer token for /metrics; empty + production = metrics disabled.
METRICS_TOKEN=

# CORS allow-list (comma separated; empty = * for development only)
CORS_ALLOWED_ORIGINS=

# golang-migrate directory (copied into the Docker image)
MIGRATIONS_PATH=./migrations
```

---

## Development

### Makefile Targets

```makefile
run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./... -v -cover

migrate-up:
	go run cmd/migrate/main.go up

proto-gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

lint:
	golangci-lint run
```

### React Development

```bash
cd clients/web
npm install
npm run dev
npm run build
npm test
```

---

## Testing

```bash
# Unit tests
go test ./internal/modules/... -v

# Integration tests
go test ./internal/modules/... -tags=integration -v

# Coverage
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Deployment

### Kubernetes

Production manifests live in [`backend/deploy/k8s`](backend/deploy/k8s): namespace,
ConfigMap, Secret (placeholders — replace or use ExternalSecrets), a hardened
backend Deployment (non-root UID 10001, read-only rootfs, dropped
capabilities, liveness/readiness/startup probes, resource limits), a frontend
Deployment + Service (nginx-served SPA, same hardening), two ClusterIP
Services, and an nginx Ingress with TLS for `api.example.com` (backend) and
`app.example.com` (SPA).

```bash
# 1. Replace the placeholders in secret.yaml (or use a secret store).
#    Production startup rejects placeholder secrets (CHANGE_ME, etc.) and
#    short JWT secrets.
# 2. Provision PostgreSQL, Redis and NATS out of band (managed services or
#    separate manifests) and point DATABASE_URL / REDIS_URL / NATS_URL at them.
# 3. Set TRUSTED_PROXIES in configmap.yaml to the ingress controller's CIDRs
#    (otherwise the rate limiter sees one shared IP for all users).
# 4. Create the ingress TLS secret (or enable the cert-manager annotation).
# 5. Apply everything
kubectl apply -k backend/deploy/k8s
```

CI builds and publishes both images to GHCR on `main`:
`ghcr.io/<owner>/approval-flow` (backend) and `ghcr.io/<owner>/approval-flow-web`
(frontend), tagged with the commit SHA plus `latest`. Pin the exact sha tag in
`deployment.yaml` / `frontend.yaml` for immutable deployments.
Replicas run golang-migrate concurrently; the postgres driver serializes
migrations with an advisory lock, so startup is safe with multiple replicas.

> The manifests deploy only the application — PostgreSQL, Redis and NATS are
> intentionally not included (teams typically use managed services). The app
> will not become ready until those dependencies are reachable at the
> configured URLs.

> The WebSocket endpoint (`/ws`) requires an `Authorization: Bearer` token, so
> clients connecting from outside the cluster must route through the same
> Ingress (which terminates TLS). The frontend nginx proxies `/ws` and `/api`
> to the backend in-cluster, so browsers are same-origin and need no CORS.

---

## License

Copyright (c) 2026 AeroXe Enterprises Private Limited. All rights reserved.

---

*Built with love by the AeroXe Team*
