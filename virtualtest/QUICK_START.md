# 🚀 Quick Start - CLI Commands

This file lists the commands needed to start the test environment from the CLI.

## 📋 Prerequisites

- Docker and Docker Compose installed
- Main system (go-nmos) running (`http://localhost:9090`)

## 🎯 Step-by-step start

### 1. Go to test environment folder

```bash
cd virtualtest
```

### 2. Start mock services

```bash
# Start all mock services (background)
docker-compose up -d

# Or with logs (foreground)
docker-compose up
```

### 3. Check service status

```bash
docker-compose ps

# Expected: mock-node-1, mock-node-2, mock-registry, mock-is07, mock-is08 Up (healthy)
```

### 4. Health checks

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
```

### 5. Follow logs (optional)

```bash
docker-compose logs -f
docker-compose logs -f mock-node-1
```

## 🔍 Test mock nodes

Use `curl` on `/x-nmos/node/v1.3/self`, `.../senders`, `.../receivers`, `.../flows`, and IS-05 connection URLs for ports 8080, 8081, 8082 as in the integration guide.

## 🔌 Integration with main system

```bash
API_URL="http://localhost:9090"
TOKEN="your_auth_token_here"

# Discover Mock Node 1
curl -X POST ${API_URL}/api/nmos/discover -H "Content-Type: application/json" -H "Authorization: Bearer ${TOKEN}" -d '{"base_url": "http://localhost:8080"}'

# Discover Mock Node 2
curl -X POST ${API_URL}/api/nmos/discover -H "Content-Type: application/json" -H "Authorization: Bearer ${TOKEN}" -d '{"base_url": "http://localhost:8081"}'

# Discover from Registry
curl -X POST ${API_URL}/api/nmos/registry/discover-nodes -H "Content-Type: application/json" -H "Authorization: Bearer ${TOKEN}" -d '{"query_url": "http://localhost:8082"}'
```

**Get token:** `POST ${API_URL}/api/login` with `username` and `password`; use the returned token in the commands above.

## 🧪 Run test scripts

```bash
cd test-scripts
pip install -r requirements.txt
python test_flows.py
python test_collisions.py
```

## 🛑 Stop services

```bash
docker-compose down
docker-compose stop
docker-compose restart
```

## 🐛 Troubleshooting

- **Services not starting:** `docker-compose logs`, then `docker-compose build` and `docker-compose up -d`
- **Port conflict:** Check with `lsof -i :8080` etc.; change ports in `docker-compose.yml` if needed
- **Unhealthy containers:** `docker-compose logs mock-node-1` and run health check inside container

## ✅ Success check

All health endpoints should return OK; IS-04 self and Registry nodes should return JSON. See integration guide for full checks.
