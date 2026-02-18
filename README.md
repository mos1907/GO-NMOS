# go-NMOS

Production-oriented rewrite baseline of NMOS management stack using **Go + Svelte**.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?logo=svelte)](https://svelte.dev/)

## Stack

- Backend: Go 1.22, chi router, JWT auth, PostgreSQL (pgx)
- Frontend: Svelte 5 + Vite + Tailwind CSS
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
│   │   │       ├── nmos.go         # NMOS discovery, apply & port scanning
│   │   │       ├── nmos_registry.go # NMOS Registry (RDS) discovery
│   │   │       ├── checker.go     # Collision detection
│   │   │       ├── automation.go   # Automation jobs
│   │   │       ├── planner.go      # Address planner
│   │   │       ├── users.go        # User management
│   │   │       ├── settings.go     # Settings management
│   │   │       ├── logs.go         # Log viewing/download
│   │   │       ├── handler.go      # Router & middleware setup
│   │   │       └── ...
│   │   ├── models/
│   │   │   ├── models.go           # Data models
│   │   │   └── nmos.go             # NMOS-specific models
│   │   ├── repository/
│   │   │   ├── repository.go       # Repository interface
│   │   │   └── postgres_repo.go    # PostgreSQL implementation
│   │   ├── mqtt/
│   │   │   └── client.go           # MQTT event publishing
│   │   └── service/
│   │       └── automation_runner.go # Scheduled job runner
│   ├── migrations/                  # Database migrations
│   │   ├── 0001_init.sql
│   │   ├── 0002_seed_admin.sql
│   │   ├── 0003_checker_automation.sql
│   │   ├── 0004_address_buckets.sql
│   │   ├── 0005_nmos_registry.sql
│   │   └── 0006_add_flow_alias_user_fields.sql
│   ├── Dockerfile
│   ├── go.mod
│   └── env.example                  # Environment variables template
│
├── frontend/                        # Svelte frontend application
│   ├── src/
│   │   ├── pages/
│   │   │   ├── LoginPage.svelte   # Login UI
│   │   │   └── DashboardPage.svelte # Main dashboard orchestrator
│   │   ├── components/             # Feature-specific components
│   │   │   ├── DashboardHomeView.svelte    # Dashboard overview
│   │   │   ├── FlowsView.svelte            # Flow list & management
│   │   │   ├── SearchView.svelte           # Flow search
│   │   │   ├── NewFlowView.svelte          # Create new flow (modal)
│   │   │   ├── NMOSView.svelte             # NMOS discovery & apply
│   │   │   ├── NMOSPatchView.svelte        # NMOS Patch Panel (IS-04/IS-05)
│   │   │   ├── TopologyView.svelte         # NMOS Topology view
│   │   │   ├── CheckerView.svelte          # Collision checker
│   │   │   ├── AutomationJobsView.svelte   # Automation jobs
│   │   │   ├── PlannerView.svelte         # Address planner
│   │   │   ├── AddressMapView.svelte      # Address map
│   │   │   ├── PortExplorerView.svelte    # Port scanning & discovery
│   │   │   ├── UsersView.svelte           # User management
│   │   │   ├── SettingsView.svelte        # Settings
│   │   │   ├── LogsView.svelte            # Log viewer
│   │   │   ├── EmptyState.svelte          # Empty state component
│   │   │   └── SkeletonLoader.svelte      # Loading skeleton component
│   │   ├── lib/
│   │   │   ├── api.js              # API client utilities
│   │   │   ├── mqtt.js            # MQTT WebSocket client
│   │   │   └── icons.js           # Icon library (SVG icons)
│   │   ├── stores/
│   │   │   └── auth.js            # Authentication store
│   │   ├── App.svelte
│   │   └── main.js
│   ├── Dockerfile
│   ├── package.json
│   ├── tailwind.config.cjs        # Tailwind CSS configuration
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

## Features

### Core Features

- **JWT Authentication** - Secure login with role-based access control
- **Flow Management** - CRUD operations for multicast flows
- **NMOS Integration** - IS-04 discovery and IS-05 connection management
- **NMOS Patch Panel** - Visual router-style interface for sender/receiver patching
- **NMOS Registry (RDS) Support** - Connect to IS-04 Query API registries
- **Collision Detection** - Automatic detection of IP/port conflicts
- **Automation Jobs** - Scheduled tasks for flow management
- **Address Planning** - Hierarchical address bucket management
- **Real-time Updates** - MQTT-based live UI updates

### API Endpoints

#### Authentication
- `POST /api/login` - User login
- `GET /api/me` - Current user info

#### Flows
- `GET /api/flows?limit=50&offset=0&sort_by=updated_at&sort_order=desc`
- `GET /api/flows/summary`
- `GET /api/flows/search?q=...&limit=50&offset=0`
- `GET /api/flows/export`
- `POST /api/flows`
- `POST /api/flows/import`
- `PATCH /api/flows/{id}`
- `POST /api/flows/{id}/lock` (lock/unlock flow)
- `DELETE /api/flows/{id}` (admin)
- `DELETE /api/flows/{id}/hard` (admin, requires `hard_delete_enabled=true` setting)

#### NMOS Discovery & Apply
- `GET /api/nmos/discover?base_url=http://<host>:<port>`
- `POST /api/nmos/discover` body: `{ "base_url": "http://<host>:<port>" }`
- `POST /api/nmos/detect-is05` body: `{ "base_url": "http://<host>:<port>" }` - Auto-detect IS-05 endpoint
- `POST /api/nmos/detect-is04-from-rds` body: `{ "rds_query_url": "http://<registry>:<port>", "node_id": "optional" }` - Auto-detect IS-04 from RDS
- `POST /api/nmos/explore-ports` body: `{ "host": "192.168.1.100", "ports": [8080, 8081], "port_range": "8080-8090", "concurrency": 10, "timeout": 3 }` - Port scanning (Admin only)
- `POST /api/flows/{id}/nmos/apply` body: `{ "connection_url": "http://.../staged", "sender_id": "optional" }`

#### NMOS Registry (RDS)
- `POST /api/nmos/registry/discover-nodes` body: `{ "query_url": "http://<registry>:<port>" }`
  - Discovers nodes from an IS-04 Query API registry
  - Supports both root URLs and versioned Query API URLs

#### Checker
- `GET /api/checker/collisions`
- `GET /api/checker/nmos?timeout=5` - NMOS difference detection
- `GET /api/checker/latest?kind=collisions|nmos`

#### Automation
- `GET /api/automation/jobs`
- `GET /api/automation/jobs/{job_id}`
- `PUT /api/automation/jobs/{job_id}`
- `POST /api/automation/jobs/{job_id}/enable`
- `POST /api/automation/jobs/{job_id}/disable`
- `GET /api/automation/summary`

#### Address Planning
- `GET /api/address-map`
- `GET /api/address/buckets/privileged`
- `GET /api/address/buckets/{id}/children`
- `POST /api/address/buckets/parent`
- `POST /api/address/buckets/child`
- `PATCH /api/address/buckets/{id}`
- `DELETE /api/address/buckets/{id}`
- `GET /api/address/buckets/export`
- `POST /api/address/buckets/import`

#### Logs
- `GET /api/logs?kind=api|audit&lines=200`
- `GET /api/logs/download?kind=api|audit`

#### Users & Settings
- `GET /api/users`
- `POST /api/users`
- `GET /api/settings`
- `PATCH /api/settings/{key}`

#### Health
- `GET /api/health` - Health check with DB connectivity status

## Usage Guide

### Getting Started

1. **Start the application**:
   ```bash
   cp backend/env.example backend/.env
   make up
   ```

2. **Access the UI**:
   - Frontend: `http://localhost:4173` (or your configured port)
   - Default credentials:
     - Username: `admin`
     - Password: `change-this-password`

### Finding and Adding NMOS Nodes

#### Method 1: Manual Node Addition

1. Navigate to **NMOS Patch** tab in the dashboard
2. Click **"Add Node"** button
3. Enter:
   - **Node Name**: A descriptive name (e.g., "Camera Router")
   - **IS-04 URL**: The base URL of your NMOS node (e.g., `http://192.168.1.100:8080`)
4. Click **"Add Node"**
5. The system will automatically derive the IS-05 connection endpoint

#### Method 2: Connect to NMOS Registry (RDS)

The **Connect RDS** feature allows you to discover multiple nodes from a central NMOS Registry (IS-04 Query API).

1. Navigate to **NMOS Patch** tab
2. Click **"Connect RDS"** button
3. Enter the **Registry Query API URL**:
   - Root URL: `http://registry-host:port`
   - Or versioned URL: `http://registry-host:port/x-nmos/query/v1.3`
4. Click **"Discover Nodes"**
5. The system will query the registry and list all available nodes
6. Select the nodes you want to add (use **"Select All"** for convenience)
7. Click **"Add Selected Nodes"**

**Example Registry URLs**:
- `http://192.168.1.50:8080`
- `http://registry.example.com:8080/x-nmos/query/v1.3`

### Using the NMOS Patch Panel

The **NMOS Patch Panel** provides a visual router-style interface for connecting IS-04 senders to receivers:

1. **Select Source Node**: Choose a node from the "Sources" dropdown
2. **Select Destination Node**: Choose a node from the "Destinations" dropdown
3. **Filter Sources**: Use the search box or format filter to find specific senders
4. **Filter Destinations**: Use the search box or format filter to find specific receivers
5. **Select Sender**: Click on a sender from the Sources list
6. **Select Receiver**: Click on a receiver from the Destinations list
7. **Apply Patch**: Click the green **"TAKE"** button in the center to apply the connection

**Status Indicators**:
- Green dot: Ready to patch (both sender and receiver selected, IS-05 URL configured)
- Gray dot: Waiting for selection or missing configuration

### NMOS Discovery (Legacy View)

The **NMOS** tab provides a simpler discovery interface:

1. Enter the **IS-04 Base URL** of your NMOS node
2. Click **"Discover"** to fetch receivers
3. Select a **Source Flow** from your local flows
4. Select a **Destination Receiver** from discovered receivers
5. Enter the **IS-05 Base URL** (typically: `base_url + /x-nmos/connection/<version>`)
6. Click **"TAKE"** to apply the connection

### Flow Management

- **View Flows**: Navigate to **Flows** tab to see all flows
- **Search**: Use **Search** tab for quick flow lookup by name, IP, or flow ID
- **Create Flow**: Click **"Add Flow"** button in Flows view (requires edit permissions) - opens a modal dialog
- **Lock/Unlock**: Click lock/unlock button in Flows view to prevent modifications
- **Export/Import**: Use Settings tab to export flows as JSON or import from file

### Collision Detection

1. Navigate to **Checker** tab
2. Click **"Run Collision Check"**
3. Review detected conflicts (IP/port collisions)
4. Resolve conflicts by modifying flows or address assignments

### Automation Jobs

1. Navigate to **Automation** tab (admin/editor only)
2. View configured automation jobs
3. Enable/disable jobs as needed
4. Jobs run automatically based on their schedule

### Port Explorer

The **Port Explorer** (admin only) allows you to scan IP addresses and port ranges to discover NMOS nodes:

1. Navigate to **Port Explorer** tab
2. Enter the **Host IP** address to scan
3. Specify ports:
   - **Port List**: Comma-separated ports (e.g., `8080,8081,8082`)
   - **Port Range**: Range format (e.g., `8080-8090`)
4. Adjust **Concurrency** (default: 10) and **Timeout** (default: 3 seconds)
5. Click **"Scan Ports"** to start scanning
6. Review results showing discovered NMOS API endpoints
7. Use **"Copy"** buttons to copy endpoint URLs to clipboard

**Note**: Scanning local networks (192.168.x.x, 10.x.x.x, 172.16-31.x.x) will show a warning. Be cautious when scanning external networks.

## MQTT Realtime Updates

MQTT is **enabled by default** for realtime event notifications:
- Flow create/update/delete events are published to MQTT topics:
  - `go-nmos/flows/events/all` - all flow events
  - `go-nmos/flows/events/flow/{flow_id}` - events for specific flow
- Frontend automatically subscribes via WebSocket (`ws://host:9001`) for realtime UI updates
- To disable MQTT, set `MQTT_ENABLED=false` in backend `.env`

## Quick Start

### Prerequisites

- Docker Desktop / OrbStack (or docker engine + docker compose)
- Go 1.22+ (for local development)
- Node.js 18+ (for local frontend development)

### Docker Compose Setup

```bash
# Clone the repository
git clone <repository-url>
cd GO-NMOS

# Configure backend environment
cp backend/env.example backend/.env
# Edit backend/.env and set:
# - DATABASE_URL
# - JWT_SECRET (use a strong secret!)
# - CORS_ORIGIN (frontend URL)

# Start all services
make up

# Or start only database (for local development)
docker compose up -d postgres mosquitto
```

After startup:

- API: `http://localhost:9090/api/health`
- UI: `http://localhost:4173`
- MQTT: `ws://localhost:9001`

### Local Development

**Backend**:

```bash
cd backend
go mod tidy
go run ./cmd/api
```

**Frontend**:

```bash
cd frontend
npm install
npm run dev
```

**Hybrid Setup** (Database in Docker, Backend/Frontend locally):

1. Start only database services:
   ```bash
   docker compose up -d postgres mosquitto
   ```

2. Configure `backend/.env`:
   ```env
   DATABASE_URL=postgres://user:pass@localhost:5432/postgres
   CORS_ORIGIN=http://localhost:4173
   ```

3. Run backend locally:
   ```bash
   cd backend
   go run ./cmd/api
   ```

4. Run frontend locally:
   ```bash
   cd frontend
   npm run dev
   ```

## Configuration

### Backend Environment Variables

See `backend/env.example` for all available options. Key variables:

- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT token signing (use a strong value!)
- `CORS_ORIGIN` - Allowed frontend origin (e.g., `http://localhost:4173`)
- `MQTT_ENABLED` - Enable/disable MQTT (`true`/`false`)
- `RATE_LIMIT_RPM` - Requests per minute limit

### Frontend Configuration

Frontend uses Vite and reads API URL from environment. Default API endpoint: `http://localhost:9090`

## Logging

- Backend writes API/Audit logs to `LOG_DIR` (default `/tmp/go-nmos-logs`)
- In Docker compose, logs are mounted to host `./logs`
- View logs in UI via **Logs** tab (admin only)

## Security Notes for Production

- Put backend behind reverse proxy with TLS (or use built-in HTTPS support)
- Use strong `JWT_SECRET` (generate with `openssl rand -base64 32`)

### HTTPS Support

The backend supports HTTPS with TLS certificates. To enable HTTPS:

1. **Generate a self-signed certificate** (for development):
   ```bash
   ./scripts/generate-self-signed-cert.sh [CN] [SANs]
   # Example:
   ./scripts/generate-self-signed-cert.sh localhost "DNS:localhost,DNS:*.local,IP:127.0.0.1"
   ```

2. **Configure environment variables** in `backend/.env`:
   ```bash
   HTTPS_ENABLED=true
   HTTPS_PORT=8443
   CERT_FILE=/certs/server.crt
   KEY_FILE=/certs/server.key
   ```

3. **Mount certificates** in `docker-compose.yml` (already configured):
   ```yaml
   volumes:
     - ./certs:/certs:ro
   ```

4. **Access HTTPS endpoint**:
   - HTTPS: `https://localhost:8443`
   - HTTP: `http://localhost:9090` (still available)

**Note**: Self-signed certificates will trigger browser security warnings. For production, use certificates from a trusted CA (Let's Encrypt, corporate CA, etc.).

**Production Recommendations**:
- Use Let's Encrypt certificates via certbot
- Or use a reverse proxy (nginx, Traefik) with TLS termination
- Keep HTTP server enabled for internal health checks
- Replace seeded admin credentials immediately
- Restrict CORS origin to your frontend domain
- Add rate-limit, audit trail, and secret management (Vault/KMS)
- Add CI checks (test, lint, SAST, dependency scanning)
- Review and restrict user permissions

## CI

- GitHub Actions workflow is included at `.github/workflows/ci.yml`
- Backend job: `go mod tidy` consistency + `go test ./...`
- Frontend job: `npm install` + `npm run build`

## Troubleshooting

### Database Connection Issues

- Verify `DATABASE_URL` in `backend/.env` is correct
- Check PostgreSQL is running: `docker compose ps`
- Check logs: `docker compose logs postgres`

### CORS Errors

- Ensure `CORS_ORIGIN` in `backend/.env` matches your frontend URL exactly
- Restart backend after changing `.env`

### NMOS Discovery Fails

- Verify the NMOS node is accessible from the backend server
- Check IS-04 endpoint is correct (typically `/x-nmos/query/v1.3`)
- Check network connectivity and firewall rules

### RDS Connection Issues

- Verify Registry Query API URL is correct
- Ensure registry is accessible from backend
- Check registry supports IS-04 Query API (not just Node API)

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) first.

### Quick Contribution Steps

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Reporting Issues

Found a bug or have a feature request? Please [open an issue](https://github.com/mos1907/GO-NMOS/issues).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

MIT License allows you to:
- Use the software commercially
- Modify the software
- Distribute the software
- Sublicense
- Use privately

The only requirement is to include the original copyright and license notice.
