# VirtualTest - Virtual NMOS Test Environment

This folder provides a virtual environment to test the NMOS system without physical devices or network infrastructure.

## 🎯 Purpose

- Test all NMOS features without physical devices
- Simulate IS-04, IS-05, IS-07, IS-08 APIs
- Test collision detection, flow management, routing
- Run automated test scenarios

## 📁 Folder Structure

```
virtualtest/
├── mock-devices/          # Mock NMOS device simulators
│   ├── mock-node-1/      # First virtual NMOS node
│   ├── mock-node-2/      # Second virtual NMOS node
│   └── mock-registry/    # Mock IS-04 Registry
├── test-scripts/          # Test automation scripts
│   ├── test_flows.py     # Flow test scenarios
│   ├── test_collisions.py # Collision test scenarios
│   └── test_routing.py   # Routing test scenarios
├── docs/                 # Documentation
│   └── test-scenarios.md # Test scenario descriptions
├── docker-compose.yml    # Start all services
└── README.md            # This file
```

## 🚀 Quick Start

### Start with Docker Compose

```bash
cd virtualtest

# Start all mock services (background)
docker-compose up -d

# Check service status
docker-compose ps

# Follow logs
docker-compose logs -f
```

**One command to start and check:**
```bash
cd virtualtest && docker-compose up -d && sleep 3 && docker-compose ps
```

### Stopping Services

```bash
# Stop all and remove containers
docker-compose down

# Stop only (keep containers)
docker-compose stop

# Restart
docker-compose restart
```

This starts:
- Mock NMOS Node 1 (IS-04 + IS-05) - Port 8080
- Mock NMOS Node 2 (IS-04 + IS-05) - Port 8081
- Mock IS-04 Registry (Query API) - Port 8082
- Mock IS-07 Event/Tally Service - Port 8083
- Mock IS-08 Audio Channel Mapping - Port 8084

### 2. Check Service Status

```bash
docker-compose ps
docker-compose logs -f
docker-compose logs -f mock-node-1
```

### 3. Integration with Main System

After mock services are up, you can discover them from the main system:

```bash
# Discover Mock Node 1
curl -X POST http://localhost:9090/api/nmos/discover \
  -H "Content-Type: application/json" \
  -d '{"base_url": "http://localhost:8080"}'

# Discover Mock Node 2
curl -X POST http://localhost:9090/api/nmos/discover \
  -H "Content-Type: application/json" \
  -d '{"base_url": "http://localhost:8081"}'

# Discover nodes from Mock Registry
curl -X POST http://localhost:9090/api/nmos/registry/discover-nodes \
  -H "Content-Type: application/json" \
  -d '{"query_url": "http://localhost:8082"}'
```

**Note:** The main system (`http://localhost:9090`) must be running.

### 4. Run Test Scenarios

```bash
cd test-scripts
pip install -r requirements.txt
python test_flows.py
python test_collisions.py
```

**Note:** Test scripts expect the main system to be running (`http://localhost:9090`).

## 🔧 Mock Services

### Mock NMOS Node

Each mock node exposes **only the IS-04 Node API** (per AMWA-NMOS spec; no Query API on the node):
- **IS-04 Node API**: `/x-nmos/node/v1.3/self`, `.../devices`, `.../flows`, `.../senders`, `.../receivers`
- **IS-05 Connection API**: `/x-nmos/connection/v1.0/`

### Mock IS-04 Registry

- **Query API**: `/x-nmos/query/v1.3/`
- Nodes, Devices, Flows, Senders, Receivers under `/x-nmos/query/v1.3/`

### Mock IS-07 Event/Tally

- **Events**: `/x-nmos/events/v1.0/events`
- **State**: `/x-nmos/events/v1.0/state`

### Mock IS-08 Audio Channel Mapping

- **Inputs/Outputs**: `/x-nmos/channelmapping/v1.0/`

## 📝 Test Scenarios

See `docs/test-scenarios.md` for detailed test scenarios.

## 🛠️ Development

Mock services are written in Python Flask. For development:

```bash
# Run mock node manually (without Docker)
cd mock-devices/mock-node-1
pip install -r requirements.txt
python app.py

# Run test script
cd test-scripts
pip install -r requirements.txt
python test_flows.py
```

## 🛑 Stopping Services

```bash
./stop-test-env.sh
# or
docker-compose down
```

## BCC / External Client Connection (AMWA-NMOS compliant)

Per IS-04, the **Query API** is only provided by the **Registry** (Query Service); nodes only expose the **Node API**. Any client using BCC or the Query API must connect to the **Registry address**.

- **In this test environment:** Query URL / Registry = **`http://<test-host>:8082`** or **`http://<test-host>:8082/x-nmos/query/v1.3`** (mock-registry).
- **If the main project backend is the registry:** **`http://<backend-host>:9090/x-nmos/query/v1.3`**.

Node URLs (`http://<host>:8080`, `http://<host>:8081`) are for Node API only; Query API paths are not on the node (use the Registry for getSenders etc.).

Example (localhost): Registry = `http://localhost:8082`, Node 1 = `http://localhost:8080`, Node 2 = `http://localhost:8081`.

## 🐛 Troubleshooting

### Services not starting
```bash
docker-compose logs
docker-compose down
docker-compose up -d
```

### Port conflict
If ports are in use, change port numbers in `docker-compose.yml`.

### Mock Registry not discovering nodes
The Mock Registry expects mock-node-1 and mock-node-2 to be healthy. Ensure they are up first:
```bash
docker-compose ps
```

## 📖 Integration Guide

To learn how to add mock nodes to the main system:
👉 See **[Integration Guide](docs/integration-guide.md)**.

**Quick reference:** See **[MOCK_NODES_INFO.md](MOCK_NODES_INFO.md)** for mock node details.

### Short summary – Adding mock nodes

**Method 1: From UI (easiest)**
1. Main system: `http://localhost:9090`
2. Dashboard → **"NMOS"** or **"NMOS Patch Panel"** tab
3. Base URL: `http://localhost:8080` (Mock Node 1) or `http://localhost:8081` (Mock Node 2)
4. Click **"Discover"**

**Method 2: Port Explorer (automatic)**
1. Dashboard → **"Port Explorer"**
2. Host: `localhost`, Port range: `8080-8084`
3. **"Explore Ports"** → **"Register Node"** for each found node

**Method 3: Registry (bulk)**
1. Dashboard → **"NMOS Patch Panel"** → **"Connect to RDS"**
2. Query URL: `http://localhost:8082`
3. Discover all nodes

## Validation with AMWA nmos-testing

You can validate the mock environment with [AMWA nmos-testing](https://github.com/AMWA-TV/nmos-testing) (official NMOS test tool):

- **IS-04-01 (Node API):** Use node URL `http://<host>:8080` or `http://<host>:8081` (mock-node-1 / mock-node-2).
- **IS-04-02 (Registry APIs):** Use Registry/Query API `http://<host>:8082` (mock-registry).
- **IS-05-01, IS-05-02:** Can be tested with node or registry.

Setup and usage: [nmos-testing README](https://github.com/AMWA-TV/nmos-testing) and [docs](https://github.com/AMWA-TV/nmos-testing/tree/master/docs). Ensure mock services are up with `docker-compose up -d` before running tests.

## 📚 Further Reading

- [NMOS Specifications](https://specs.amwa.tv/nmos/)
- [IS-04 Discovery & Registration](https://specs.amwa.tv/nmos/branches/main/docs/IS-04_v1.3.html)
- [IS-05 Connection Management](https://specs.amwa.tv/nmos/branches/main/docs/IS-05_v1.1.html)
- [AMWA NMOS Testing Tool](https://github.com/AMWA-TV/nmos-testing) — official conformance test tool
