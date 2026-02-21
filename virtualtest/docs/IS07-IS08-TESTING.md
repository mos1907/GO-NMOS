# IS-07 and IS-08 Testing Guide

This document describes how to run **IS-07 (Event & Tally)** and **IS-08 (Audio Channel Mapping)** tests step by step.

## Prerequisite: Mock services running

```bash
cd virtualtest
docker compose up -d
```

Ensure these containers are **Up (healthy)**:

- **mock-is07** → `http://localhost:8083`
- **mock-is08** → `http://localhost:8084`

Check: `docker compose ps`

---

## 1. IS-07 (Event & Tally) tests

### 1.1 Discovery test (Checker / Interop)

1. Go to the **Checker** tab in the dashboard.
2. Enter the IS-07 device base URL in **Test at URL**:
   - For mock: `http://localhost:8083`
3. Click **Run tests**.
4. **IS-07 Events API Discovery** should be "pass" or "skip"; if the mock is running you should see "Discovered IS-07 API version: v1.0".

### 1.2 Events tab (in-app)

1. Click the **Events** tab in the dashboard.
2. In the **Events (IS-07 / Tally)** panel:
   - Use filters: Source, Severity, Since, Limit.
   - **Load events** fetches the event list from the backend (may be empty at first).
3. To add events, use the backend API:
   - `POST /api/events` (body: `source_url`, `source_id`, `severity`, `message`, optional `payload`, `flow_id`, `sender_id`, `receiver_id`, `job_id`).
   - Example:
     ```json
     {
       "source_url": "http://localhost:8083/x-nmos/events/v1.0",
       "source_id": "mock-source-1",
       "severity": "info",
       "message": "Test event from IS-07 mock"
     }
     ```
   - Then refresh the list with **Load events** in the Events tab.

### 1.3 IS-07 sources (device source list)

The backend proxies listing of event sources for an IS-07 device:

- **GET** `/api/events/is07/sources?base_url=http://localhost:8083/x-nmos/events/v1.0`
- The mock IS-07 supports `/x-nmos/events/v1.0/sources`; this request returns the source list from the mock.

You can call this from Postman/curl or your own frontend/script (auth token may be required).

---

## 2. IS-08 (Audio Channel Mapping) tests

### 2.1 Discovery test (Checker / Interop)

1. Go to the **Checker** tab.
2. Enter the IS-08 device base URL in **Test at URL**:
   - For mock (node base): `http://localhost:8084`
   - Note: The test harness tries the `.../x-nmos/channelmapping/` path for IS-08; if the full base URL does not work, use only `http://localhost:8084`.
3. Click **Run tests**.
4. **IS-08 Audio Channel Mapping Discovery** should be "pass" or "skip"; if the mock is running you should see "Discovered IS-08 API version: v1.0".

### 2.2 Audio (IS-08) tab (in-app)

1. Click the **Audio (IS-08)** tab.
2. Enter **Channel Mapping API base URL**:
   - For mock: `http://localhost:8084/x-nmos/channelmapping/v1.0`
   - Note: The mock uses the path `channelmapping` (one word); the spec sometimes uses `channel_mapping`; our mock uses `channelmapping`.
3. Click **Load**.
   - The backend proxies `/io` and `/map/active` to this base URL.
   - The mock supports these endpoints, so Inputs/Outputs and the active map are shown.
4. **Presets**: You can apply example maps (**Pass-through**, **Stereo**, **5.1**). **Apply** sends `POST .../map/activations` to the mock; then use **Load** again to refresh.

### 2.3 Quick IS-08 check with curl

```bash
# Version list
curl -s http://localhost:8084/x-nmos/channelmapping/v1.0/

# IO (input/output list and channel info)
curl -s http://localhost:8084/x-nmos/channelmapping/v1.0/io

# Active map
curl -s http://localhost:8084/x-nmos/channelmapping/v1.0/map/active
```

---

## 3. Summary table

| What is tested?       | Where?                | URL / Step |
|-----------------------|------------------------|------------|
| IS-07 API discovery   | Checker → Test at URL  | `http://localhost:8083` |
| IS-07 event list      | Events tab             | Load events (DB) |
| IS-07 add event       | API                    | `POST /api/events` |
| IS-07 source list     | API                    | `GET /api/events/is07/sources?base_url=...` |
| IS-08 API discovery   | Checker → Test at URL  | `http://localhost:8084` |
| IS-08 IO and map      | Audio (IS-08) tab      | Base: `http://localhost:8084/x-nmos/channelmapping/v1.0` → Load |
| IS-08 apply preset    | Audio (IS-08) tab      | Pass-through / Stereo / 5.1 → Apply |

---

## 4. Troubleshooting

- **IS-07/IS-08 discovery "skip"**: Ensure mocks are running (`docker compose ps`) and ports 8083/8084 are open. If the backend runs on another machine, use an address the backend can reach for "Test at URL" (e.g. host IP).
- **Audio (IS-08) Load error**: Do not add a trailing `/` to the base URL; use `http://localhost:8084/x-nmos/channelmapping/v1.0`.
- **Events empty**: Add a few events with `POST /api/events` first, then use **Load events** in the Events tab.

With these steps you can run IS-07 and IS-08 tests in-app and against the mock services.
