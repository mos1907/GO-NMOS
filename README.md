# go-NMOS

Production-oriented rewrite baseline of NMOS management stack using **Go + Svelte**.

## Stack

- Backend: Go 1.22, chi router, JWT auth, PostgreSQL (pgx)
- Frontend: Svelte 5 + Vite
- Infra: Docker Compose (PostgreSQL + Mosquitto + backend + frontend)

## Project Structure

```
go-NMOS/
├── backend/                          # Go backend service
│   ├── cmd/
│   │   └── api/
│   │       └── main.go              # Application entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go           # Configuration management
│   │   ├── db/
│   │   │   └── postgres.go         # Database connection & migrations
│   │   ├── http/
│   │   │   └── handlers/           # HTTP request handlers
│   │   │       ├── auth.go         # Authentication endpoints
│   │   │       ├── flows.go        # Flow CRUD operations
│   │   │       ├── nmos.go         # NMOS discovery & apply
│   │   │       ├── checker.go     # Collision detection
│   │   │       ├── automation.go  # Automation jobs
│   │   │       ├── planner.go     # Address planner
│   │   │       ├── users.go       # User management
│   │   │       ├── settings.go    # Settings management
│   │   │       ├── logs.go        # Log viewing/download
│   │   │       ├── handler.go     # Router & middleware setup
│   │   │       └── ...
│   │   ├── models/
│   │   │   └── models.go          # Data models
│   │   ├── repository/
│   │   │   ├── repository.go      # Repository interface
│   │   │   └── postgres_repo.go   # PostgreSQL implementation
│   │   ├── mqtt/
│   │   │   └── client.go          # MQTT event publishing
│   │   └── service/
│   │       └── automation_runner.go # Scheduled job runner
│   ├── migrations/                  # Database migrations
│   │   ├── 0001_init.sql
│   │   ├── 0002_seed_admin.sql
│   │   ├── 0003_checker_automation.sql
│   │   └── 0004_address_buckets.sql
│   ├── Dockerfile
│   ├── go.mod
│   └── env.example                  # Environment variables template
│
├── frontend/                        # Svelte frontend application
│   ├── src/
│   │   ├── pages/
│   │   │   ├── LoginPage.svelte   # Login UI
│   │   │   └── DashboardPage.svelte # Main dashboard
│   │   ├── lib/
│   │   │   ├── api.js             # API client utilities
│   │   │   └── mqtt.js            # MQTT WebSocket client
│   │   ├── stores/
│   │   │   └── auth.js           # Authentication store
│   │   ├── App.svelte
│   │   └── main.js
│   ├── Dockerfile
│   ├── package.json
│   └── vite.config.js
│
├── deploy/
│   └── mosquitto.conf              # MQTT broker configuration
│
├── docker-compose.yml              # Docker orchestration
├── Makefile                        # Build & deployment commands
├── README.md                       # This file
├── PROJECT_STATUS.md              # Project status (Turkish)
├── PROJECT_STATUS_EN.md           # Project status (English)
├── MQTT_EXPLANATION.md            # MQTT guide (Turkish)
└── MQTT_EXPLANATION_EN.md         # MQTT guide (English)
```

## Features included now

- JWT login and `/api/me`
- Flow CRUD baseline
  - `GET /api/flows?limit=50&offset=0&sort_by=updated_at&sort_order=desc`
  - `GET /api/flows/summary`
  - `GET /api/flows/search?q=...&limit=50&offset=0&sort_by=updated_at&sort_order=desc`
  - `GET /api/flows/export`
  - `POST /api/flows`
  - `POST /api/flows/import`
  - `PATCH /api/flows/{id}`
  - `POST /api/flows/{id}/lock` (lock/unlock flow)
  - `DELETE /api/flows/{id}` (admin)
- NMOS discover
  - `GET /api/nmos/discover?base_url=http://<host>:<port>`
  - `POST /api/nmos/discover` body: `{ "base_url": "http://<host>:<port>" }`
  - `GET /api/flows/{id}/nmos/check?base_url=http://<host>:<port>`
  - `POST /api/flows/{id}/nmos/apply` body: `{ "connection_url": "http://.../staged", "sender_id": "optional" }`
- Checker
  - `GET /api/checker/collisions`
  - `GET /api/checker/latest?kind=collisions`
- Automation jobs
  - `GET /api/automation/jobs`
  - `GET /api/automation/jobs/{job_id}`
  - `PUT /api/automation/jobs/{job_id}`
  - `POST /api/automation/jobs/{job_id}/enable`
  - `POST /api/automation/jobs/{job_id}/disable`
  - `GET /api/automation/summary`
- Address map
  - `GET /api/address-map`
- Planner buckets
  - `GET /api/address/buckets/privileged`
  - `GET /api/address/buckets/{id}/children`
  - `POST /api/address/buckets/parent`
  - `POST /api/address/buckets/child`
  - `PATCH /api/address/buckets/{id}`
  - `DELETE /api/address/buckets/{id}`
  - `GET /api/address/buckets/export`
  - `POST /api/address/buckets/import`
- Logs
  - `GET /api/logs?kind=api|audit&lines=200`
  - `GET /api/logs/download?kind=api|audit`
- User management baseline
  - `GET /api/users`
  - `POST /api/users`
- Settings baseline
  - `GET /api/settings`
  - `PATCH /api/settings/{key}`
- Health check: `GET /api/health`
  - includes DB connectivity status
- Migration runner with schema bootstrap
- Web UI tabs: Dashboard / Flows / Search / New Flow / Users / NMOS / Checker / Automation / Planner / Address Map / Logs / Settings

## MQTT Realtime Updates (Enabled by Default)

MQTT is **enabled by default** for realtime event notifications:
- Flow create/update/delete events are published to MQTT topics:
  - `go-nmos/flows/events/all` - all flow events
  - `go-nmos/flows/events/flow/{flow_id}` - events for specific flow
- Frontend automatically subscribes via WebSocket (`ws://host:9001`) for realtime UI updates
- To disable MQTT, set `MQTT_ENABLED=false` in backend `.env`

## Quick Start

Prerequisites:

- Docker Desktop / OrbStack (or docker engine + docker compose)

```bash
cd /Users/muratdemirci/nmosProject/go-NMOS
# backend bootstrap credentials and secret
cp backend/env.example backend/.env
make up
```

After startup:

- API: `http://192.168.248.133:9090/api/health`
- UI: `http://192.168.248.133:4173`
- MQTT: `ws://192.168.248.133:9001`

Default bootstrap admin from `backend/env.example`:

- username: `admin`
- password: `change-this-password`

You must set a strong `JWT_SECRET` and change admin password before production rollout.

Logging:

- Backend writes API/Audit logs to `LOG_DIR` (default `/tmp/go-nmos-logs`)
- In Docker compose, logs are mounted to host `./logs`

Rate limit:

- Global request limiter is enabled by default.
- Configure requests per minute with `RATE_LIMIT_RPM` in backend env.

## Local Development

Backend:

```bash
cd backend
go mod tidy
go run ./cmd/api
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

## Security notes for production

- Put backend behind reverse proxy with TLS
- Use strong `JWT_SECRET`
- Replace seeded admin credentials immediately
- Restrict CORS origin
- Add rate-limit, audit trail, and secret management (Vault/KMS)
- Add CI checks (test, lint, SAST, dependency scanning)

## CI

- GitHub Actions workflow is included at `.github/workflows/ci.yml`
- Backend job: `go mod tidy` consistency + `go test ./...`
- Frontend job: `npm install` + `npm run build`
