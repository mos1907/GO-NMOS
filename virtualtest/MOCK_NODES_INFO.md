# Mock Node Info - Quick Reference

This file contains all mock node information. Use it when adding them to the main system.

## 🎯 Quick Access

### Mock Node 1 (Encoder)
```
Base URL: http://localhost:8080
IS-04 API: http://localhost:8080/x-nmos/node/v1.3/
IS-05 API: http://localhost:8080/x-nmos/connection/v1.0/
```

### Mock Node 2 (Decoder)
```
Base URL: http://localhost:8081
IS-04 API: http://localhost:8081/x-nmos/node/v1.3/
IS-05 API: http://localhost:8081/x-nmos/connection/v1.0/
```

### Mock Registry
```
Query URL: http://localhost:8082
Query API: http://localhost:8082/x-nmos/query/v1.3/
```

## 📋 Detailed Info

### Mock Node 1 - Encoder Node

**Node info:**
- **Node ID**: `550e8400-e29b-41d4-a716-446655440001`
- **Label**: `Mock Encoder Node 1`
- **Description**: `Virtual NMOS encoder for testing`
- **Site**: `CampusA`, **Room**: `Studio1`
- **Hostname**: `mock-node-1.local`

**Senders:**
1. **Video Sender 1** – Sender ID (UUID, changes on each start), Flow: Video Flow 1, SDP: `http://localhost:8080/sdp/video1.sdp`, Multicast IP: `239.0.0.1`, Port: `5004`
2. **Audio Sender 1** – Sender ID (UUID), Flow: Audio Flow 1, SDP: `http://localhost:8080/sdp/audio1.sdp`, Multicast IP: `239.0.0.2`, Port: `5004`

**Receivers:** 1 Video Receiver 1 (UUID, format: video)

**Flows:** Video Flow 1 (video/smpte291, 1920x1080, 25fps), Audio Flow 1 (audio/L24, 48kHz, 2 ch)

### Mock Node 2 - Decoder Node

**Node info:** Label `Mock Decoder Node 2`, Site `CampusA`, Room `ControlRoom1`.

**Receivers:** 2 Video, 1 Audio (RTP multicast). **Senders/Flows:** None (decoder node).

## 🔌 Adding to the Main System

### Method 1: From UI (easiest)
1. Open main system: `http://localhost:9090`
2. Go to **"NMOS"** or **"NMOS Patch Panel"**
3. Enter **"NMOS Node Base URL"**: `http://localhost:8080` (Node 1) or `http://localhost:8081` (Node 2)
4. Click **"Discover"**

### Method 2: Port Explorer
1. Dashboard → **"Port Explorer"**, Host: `localhost`, Port range: `8080-8084`
2. **"Explore Ports"** → **"Register Node"** for each NMOS service found

### Method 3: Registry (bulk)
1. **"NMOS Patch Panel"** → **"Connect to RDS"**, Query URL: `http://localhost:8082`
2. **"Discover"** to add all mock nodes

### Method 4: API (for developers)
Use `POST /api/nmos/discover` with `base_url` or `POST /api/nmos/registry/discover-nodes` with `query_url`. See integration guide for examples.

## ✅ After Discovery

- **Flows:** "Video Flow 1" and "Audio Flow 1" should appear in Flows tab
- **Nodes:** Mock Node 1 and 2 should appear in NMOS Patch Panel
- **SDP:** Multicast IP and port should be parsed in flow details

## 📝 Notes

- Mock node UUIDs change on each start; this is expected
- SDP files are served from the `/sdp/` endpoint on mock nodes
- The main system parses SDPs to extract multicast IP and port
- From Docker, use container IP or `host.docker.internal` instead of `localhost` if needed
