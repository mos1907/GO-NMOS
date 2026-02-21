# Test Scenarios

This document describes the test scenarios that can be run in the VirtualTest environment.

## 1. Basic Flow Discovery Tests

### Scenario 1.1: Mock Node Discovery
- **Goal**: Discover a mock NMOS node
- **Steps**:
  1. Discover Mock Node 1 (`http://localhost:8080`)
  2. Verify node info
  3. Check sender, receiver, and flow counts
- **Expected**: Node discovered successfully; 2 senders, 1 receiver, 2 flows

### Scenario 1.2: Registry Discovery
- **Goal**: Discover all nodes via the Mock Registry
- **Steps**:
  1. Discover Mock Registry (`http://localhost:8082`)
  2. List all nodes
  3. Check resources for each node
- **Expected**: 2 nodes discovered; all resources visible

## 2. Collision Detection Tests

### Scenario 2.1: IP/Port Collision Detection
- **Goal**: Detect flows using the same IP and port
- **Steps**:
  1. Create two flows with the same IP and port
  2. Run collision check
  3. Check reported collisions
- **Expected**: Collision detected

### Scenario 2.2: Alternative Suggestions
- **Goal**: Get alternative IP/port suggestions when there is a collision
- **Steps**:
  1. Request alternatives for a conflicting IP/port
  2. Check suggestions
- **Expected**: At least 3–5 alternatives suggested

## 3. IS-05 Connection Tests

### Scenario 3.1: Sender–Receiver Connection
- **Goal**: Establish a sender–receiver connection on mock nodes
- **Steps**:
  1. Select a sender from Mock Node 1
  2. Select a receiver from Mock Node 2
  3. Establish connection via IS-05 API
  4. Check connection status
- **Expected**: Connection established successfully

## 4. Address Planning Tests

### Scenario 4.1: Address Bucket Creation
- **Goal**: Create address buckets and assign flows
- **Steps**:
  1. Create root bucket
  2. Create child bucket
  3. Assign flow to bucket
  4. Check bucket usage stats
- **Expected**: Buckets created and flows assigned

## 5. Port Explorer Tests

### Scenario 5.1: Port Scan
- **Goal**: Port-scan mock nodes
- **Steps**:
  1. Start port scan (range 8080–8084)
  2. Identify NMOS services
  3. Register nodes automatically
- **Expected**: Mock nodes detected and registered

## 6. IS-07 Event & Tally Tests

### Scenario 6.1: IS-07 API discovery
- **Goal**: Discover the Mock IS-07 Events API
- **Steps**:
  1. Start mock-is07 with `docker compose up -d` (port 8083)
  2. Dashboard → Checker → Test at URL: `http://localhost:8083` → Run tests
  3. Check "IS-07 Events API Discovery" result
- **Expected**: "Discovered IS-07 API version: v1.0" (or skip)

### Scenario 6.2: Events tab and IS-07 sources
- **Goal**: Event list and device source list
- **Steps**:
  1. In Events tab, Load events (events from DB)
  2. Optional: add event with `POST /api/events`, then Load again
  3. `GET /api/events/is07/sources?base_url=http://localhost:8083/x-nmos/events/v1.0` for source list from mock
- **Details**: See `virtualtest/docs/IS07-IS08-TESTING.md`

## 7. IS-08 Audio Channel Mapping Tests

### Scenario 7.1: IS-08 API discovery
- **Goal**: Discover the Mock IS-08 Channel Mapping API
- **Steps**:
  1. Start mock-is08 with `docker compose up -d` (port 8084)
  2. Dashboard → Checker → Test at URL: `http://localhost:8084` → Run tests
  3. Check "IS-08 Audio Channel Mapping Discovery" result
- **Expected**: "Discovered IS-08 API version: v1.0" (or skip)

### Scenario 7.2: Audio (IS-08) tab – IO and map
- **Goal**: View channel mapping and apply presets
- **Steps**:
  1. Dashboard → Audio (IS-08) tab
  2. Base URL: `http://localhost:8084/x-nmos/channelmapping/v1.0` → Load
  3. Inputs/Outputs and active map should appear
  4. Choose Pass-through / Stereo / 5.1 preset → Apply
- **Details**: See `virtualtest/docs/IS07-IS08-TESTING.md`

## Running Tests

To run all tests:

```bash
cd virtualtest/test-scripts
python test_flows.py
python test_collisions.py
```

To run a single scenario, edit the relevant script and run only that test.
