# virtualtest_go – Studio B

NMOS test environment written in Go: **RDS** (Query API, port 6062) and **Studio B** – mock node with 3 camera outputs. Uses port 8180 (avoids clash with virtualtest Python on 8080).

## Components

| Component   | Port | Description |
|------------|------|-------------|
| **RDS**    | 6062 | IS-04 Query API – aggregates resources from node(s) |
| **Studio B (camera-node)** | 8180 | 1 node, 1 device, 3 video senders (Camera 1/2/3), IS-04 + IS-05 + SDP |

## Local run

```bash
# 1) Start Studio B (3-camera) node
go run ./cmd/camera-node -port 8180 -label "Studio B"

# 2) In another terminal, start RDS (default node URL: http://localhost:8180)
go run ./cmd/rds -port 6062
```

To add more nodes to RDS:

```bash
go run ./cmd/rds -port 6062 -nodes "http://localhost:8180,http://localhost:8181"
# or
RDS_NODE_URLS="http://localhost:8180" go run ./cmd/rds -port 6062
```

## Docker

```bash
docker compose up -d
```

- RDS: http://localhost:6062  
- Studio B (IS-04/IS-05): http://localhost:8180  

**Registry (RDS) URL in G0-NMOS-PRO:**  
- App running on **host**: **http://localhost:6062**  
- App running **inside Docker** (Mac/Windows/Linux): **http://host.docker.internal:6062**  

RDS uses node URL `http://host.docker.internal:8180` so the main app backend (in a different compose) can reach the node during sync. Add this registry URL and use "Discover" or "Reload" to see the 3 senders. (8180 = Studio B node port, RDS port is **6062**.)

## Endpoints

**RDS (6062)**  
- `GET /x-nmos/query/v1.3/` – version list  
- `GET /x-nmos/query/v1.3/nodes` – all nodes  
- `GET /x-nmos/query/v1.3/devices` – devices  
- `GET /x-nmos/query/v1.3/flows` – flows  
- `GET /x-nmos/query/v1.3/senders` – senders  
- `GET /x-nmos/query/v1.3/receivers` – receivers  
- `GET /health` – health  

**Studio B – camera-node (8180)**  
- IS-04: `/x-nmos/node/v1.3/self`, `/devices`, `/flows`, `/senders`, `/receivers`  
- IS-05: `/x-nmos/connection/v1.0/single/senders`, `.../senders/<id>/staged` (GET/PATCH)  
- SDP: `/sdp/cam1.sdp`, `/sdp/cam2.sdp`, `/sdp/cam3.sdp`  
- `GET /health`  

## SDP (3 cameras)

| File     | Multicast | Port |
|----------|-----------|------|
| cam1.sdp | 239.0.0.1 | 5004 |
| cam2.sdp | 239.0.0.2 | 5006 |
| cam3.sdp | 239.0.0.3 | 5008 |

Example source IP: 192.168.1.101.
