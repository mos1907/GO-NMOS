## GO-NMOS Roadmap / TODO List

This document tracks potential next steps to evolve **GO-NMOS** into a more complete ST 2110 / NMOS controller, using the AMWA NMOS article (`https://muratdemirci.com.tr/amwa-nmos/`) and other ecosystem patterns as inspiration.

Each item is a high-level feature; we can break them down further as we implement.

---

### 1. IS-04 / Registry & Query Enhancements

- [x] **Registry WebSocket feed (Query-style realtime)**  
  - Backend: WebSocket endpoint that streams internal NMOS registry changes (nodes/devices/flows/senders/receivers).  
  - Frontend: Small “Registry Events” panel on the dashboard showing latest registry changes.

- [x] **Registry health check**  
  - Backend: `/api/nmos/registry/health` that verifies connectivity to configured registry / nodes.  
  - Frontend: Compact status indicator (OK / WARN / DOWN) on dashboard or settings.

---

### 2. IS-05 Deep Integration (Connection + SDP)

- [x] **Enhanced IS-05 view in Flow Details**  
  - Show a dedicated “IS-05 / Transport” section:  
    - Path A/B: IP/port with simple “ready / missing” badges.  
    - Transport protocol and inferred ST 2110 format (e.g. 2110-20/30/40) based on SDP.

- [x] **IS-05 Receiver state check**  
  - Backend: Endpoint to query a receiver’s active/staged connection state and map it back to a flow.  
  - Frontend: Button in Flow Details to “Check Receiver State” and show whether the receiver matches the selected flow.

---

### 3. IS-08 / Audio Channel Mapping (Future Hooks)

- [ ] **Audio metadata fields on flows**  
  - Extend flow model with optional audio-related metadata (e.g. channel layout, program name) to prepare for IS-08.

- [ ] **Audio Mapping placeholder UI**  
  - Add a lightweight “Audio Channels (future)” block in Flow Details so we have a clear place to surface IS-08 integration later.

---

### 4. IS-09 / Timing & System Parameters

- [ ] **System parameters backend (IS-09-inspired)**  
  - Add a `system_parameters` concept (via settings or a small table) for:  
    - PTP domain & GM ID (manually entered for now).  
    - “Expected” NMOS versions (IS-04/05/08/09).

- [ ] **Timing / System card on dashboard**  
  - Add a “System / Timing” widget showing:  
    - PTP status (OK / WARN, even if initially manual).  
    - Configured NMOS base URLs and versions.

---

### 5. Diagnostics & Troubleshooting (Health Panel)

- [ ] **Detailed health endpoint**  
  - Backend: `/api/health/detail` summarising:  
    - DB, MQTT, registry, RDS, Port Explorer, NMOS node reachability.  
    - Basic status & error messages.

- [ ] **Diagnostics panel in UI**  
  - Frontend: “Diagnostics / Quick Check” section with buttons like:  
    - Check DB, Check MQTT, Check Registry, Check Node at URL.  
  - Display results as colored badges + timestamp, to mirror the troubleshooting patterns from the blog.

---

### 6. IS-06 / SDN Controller Hooks

- [ ] **SDN controller configuration**  
  - Backend: add `SDN_CONTROLLER_URL` config and a simple `/api/sdn/ping` endpoint to verify reachability.

- [ ] **Network Controller section in Settings**  
  - Frontend: small “Network Controller (IS-06)” box in Settings with URL field + “Ping” button and status indicator.

---

### 7. SDI → NMOS Migration Aids

- [ ] **Migration checklist UI (Ops-focused)**  
  - Frontend: a “Migration / Checklist” view that turns the SDI→NMOS steps from the blog into a simple interactive checklist.

- [ ] **Documentation cross-links**  
  - README / Help section: link clearly to the detailed AMWA NMOS article for deep-dive architecture and operations guidance.

---

### 8. Nice-to-Have UX / Polish Ideas

- [ ] **Flow edit UX refinements**  
  - Group advanced ST 2110 / NMOS fields behind collapsible sections in the edit modal (basic vs advanced).

- [ ] **Better empty/error states around NMOS / RDS / Port Explorer**  
  - Tailor error messages and “no data” screens to common real-world networking and NMOS misconfigurations described in the blog.

