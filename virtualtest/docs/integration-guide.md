# Mock Nodes Integration Guide

This guide explains how to add mock nodes from the VirtualTest environment to the main go-nmos system.

## 📋 Mock Node Info

### Mock Node 1 (Encoder)
- **Base URL**: `http://localhost:8080`
- **IS-04 API**: `http://localhost:8080/x-nmos/node/v1.3/`
- **IS-05 API**: `http://localhost:8080/x-nmos/connection/v1.0/`
- **Features**:
  - 2 Video/Audio Senders
  - 1 Receiver
  - 2 Flows (Video + Audio)
  - Site: CampusA, Room: Studio1

### Mock Node 2 (Decoder)
- **Base URL**: `http://localhost:8081`
- **IS-04 API**: `http://localhost:8081/x-nmos/node/v1.3/`
- **IS-05 API**: `http://localhost:8081/x-nmos/connection/v1.0/`
- **Features**:
  - 3 Receivers (2 Video + 1 Audio)
  - Site: CampusA, Room: ControlRoom1

### Mock Registry (IS-04 Query API)
- **Query URL**: `http://localhost:8082`
- **Query API**: `http://localhost:8082/x-nmos/query/v1.3/`
- **Features**: Automatically discovers all mock nodes

## 🚀 Integration Methods

### Method 1: Manual discovery from UI (recommended)

1. **Open the main system**: `http://localhost:9090`

2. **Go to NMOS Discovery**:
   - From the dashboard, open the "NMOS Discovery" or "NMOS Patch Panel" tab

3. **Discover Mock Node 1**:
   - Click "Discover Node" or similar
   - Enter Base URL: `http://localhost:8080`
   - Click "Discover"
   - The system will discover the node and add flows

4. **Discover Mock Node 2**:
   - Repeat with Base URL: `http://localhost:8081`

5. **Discover from Registry** (alternative):
   - Use "Discover from Registry"
   - Query URL: `http://localhost:8082`
   - The system will discover all nodes

### Method 2: Discovery via API

#### Discover Mock Node 1

```bash
curl -X POST http://localhost:9090/api/nmos/discover \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"base_url": "http://localhost:8080"}'
```

#### Discover Mock Node 2

```bash
curl -X POST http://localhost:9090/api/nmos/discover \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"base_url": "http://localhost:8081"}'
```

#### Discover all nodes from Registry

```bash
curl -X POST http://localhost:9090/api/nmos/registry/discover-nodes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"query_url": "http://localhost:8082"}'
```

### Method 3: Automatic discovery with Port Explorer

1. **Open Port Explorer**: From the dashboard, go to the "Port Explorer" tab

2. **Start port scan**:
   - Host: `localhost` or `127.0.0.1`
   - Port range: `8080-8084` (covers all mock services)
   - Click "Explore Ports"

3. **Identify NMOS services**: The system will scan; NMOS APIs will be marked as "IS-04 NMOS"

4. **Register nodes**: For each discovered NMOS node, click "Register Node"; the system will add node info and flows

## 📊 Post-discovery checks

### Check flows

```bash
curl http://localhost:9090/api/flows?limit=50 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Flows from mock nodes should include:
- **Video Flow 1** (Mock Node 1)
- **Audio Flow 1** (Mock Node 1)

### Check nodes

```bash
curl http://localhost:9090/api/nmos/registry/nodes \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Check senders/receivers

```bash
curl http://localhost:9090/api/nmos/registry/senders \
  -H "Authorization: Bearer YOUR_TOKEN"

curl http://localhost:9090/api/nmos/registry/receivers \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🔧 Mock node details

### Mock Node 1 - Flow details

**Video Flow 1:**
- Multicast IP: `239.0.0.1` (parsed from SDP)
- Port: `5004`
- Format: `urn:x-nmos:format:video`
- Resolution: 1920x1080
- Frame rate: 25fps

**Audio Flow 1:**
- Multicast IP: `239.0.0.2` (parsed from SDP)
- Port: `5004`
- Format: `urn:x-nmos:format:audio`
- Sample rate: 48kHz
- Channels: 2

### Mock Node 2 - Receiver details

- **Video Receiver 1**: Accepts video format
- **Video Receiver 2**: Accepts video format
- **Audio Receiver 1**: Accepts audio format

## 🧪 Test scenarios

### Scenario 1: Sender–receiver connection

1. Select a sender from Mock Node 1
2. Select a compatible receiver from Mock Node 2
3. Establish connection via IS-05 API
4. Check connection status

### Scenario 2: Collision detection

1. Check flows from mock nodes
2. Run Collision Check
3. View collisions
4. Check alternative suggestions

## ⚠️ Notes

1. **Port conflict**: If the main system also uses ports 8080–8084, change mock service ports in `docker-compose.yml` (e.g. map `9080:8080`).

2. **Network access**: If mock services run in a Docker network, use the container IP instead of `localhost` when accessing from the main system.

3. **CORS**: Mock services have CORS enabled; if the main system runs on a different port, this should still work.

4. **SDP**: Mock nodes serve SDP files from the `/sdp/` endpoint. The main system parses these to get multicast IP and port.

## 🐛 Troubleshooting

### Node not discovered

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8080/x-nmos/node/v1.3/self
```

### Flows not showing

- Flows should be added automatically after discovery. If not, trigger SDP fetch manually via `POST /api/flows/{flow_id}/fetch-sdp` with `manifest_url`.

### Registry discovery failing

- Ensure the mock registry can reach mock nodes. On a Docker network, use container names (`mock-node-1`, `mock-node-2`).
