## GO-NMOS Roadmap v2

This roadmap is a fresh starting point for evolving **GO-NMOS** from a useful controller into a more complete, production-ready NMOS control platform.

The previous phases (basic IS-04/05/08/09 integration, diagnostics panels, SDN ping, migration checklist, etc.) are treated as **v1 foundations**. Below we focus on what is still missing for a robust end‑to‑end NMOS system: security, completeness of specs, observability, automation, and operator UX.

Each item is intentionally high-level; we can break them down into implementation tasks as we go.

---

### 1. NMOS Spec Coverage & Compliance

- [ ] **IS-04 maturity & compatibility**
  - Expose more of the internal registry (tags, caps, version negotiation) to the UI.
  - Add basic conformance/self-checks against different IS-04 registry versions.

- [ ] **IS-05 advanced workflows**
  - Add support for staged/activation modes, scheduled activations and bulk patch (multiple receivers at once).
  - Provide a “Connection history” view per receiver (who connected what, when).

- [ ] **IS-08 audio mapping (first real implementation)**
  - Backend: minimal IS-08 controller operations for a subset of devices (read map, simple remap).
  - UI: per-flow/per-receiver audio channel map preview and a simple “swap / mute / mono-sum” editor.

- [ ] **IS-07 (Events & Tally) hooks**
  - Add a small IS-07 consumer capable of subscribing to event sources relevant to routing (tally, GPI, monitoring).
  - UI: lightweight “Events” pane with filters and correlation to flows/senders/receivers.

- [ ] **IS-09 refinements**
  - Allow editing key system parameters from the UI (with role-based protection).
  - Add validation to ensure consistency between configured expectations (IS-04/05 versions) and discovered nodes.

---

### 2. Security & Access Control (BCP‑003 Inspired)

- [ ] **API hardening & auth cleanup**
  - Review all endpoints and permissions; introduce clearer roles (viewer / operator / engineer / admin).
  - Add optional session timeout and refresh logic on the frontend.

- [ ] **TLS / HTTPS first-class support**
  - Make HTTPS configuration easier (certs, keys, auto-redirect from HTTP).
  - Provide a simple “Security status” card (HTTP vs HTTPS, weak defaults warnings).

- [ ] **BCP‑003‑style integration plan**
  - Design how GO-NMOS would talk to secured registries/nodes (tokens, CA trust, certificates) even if not fully implemented yet.
  - Add placeholders in config/settings for auth server URLs, JWKS, certificate store, etc.

---

### 3. Observability, Monitoring & Metrics

- [ ] **Structured logging & log search**
  - Standardise backend logs (JSON, correlation IDs, request IDs).
  - Frontend: add filters in Logs view (by component, severity, correlation ID).

- [ ] **Metrics & dashboards**
  - Expose Prometheus metrics (requests, error rates, flow operations, MQTT/WS stats).
  - Provide example Grafana dashboards and a short “Operations” section in the docs.

- [ ] **Alerting hooks**
  - Emit basic alerts (webhook / email / Slack-style placeholder) for key conditions:
    - Registry down / empty, MQTT disconnected, automation failures, repeated NMOS errors.

---

### 4. Topology, Routing & Automation

- [ ] **Topology view enhancements**
  - Improve the Topology view with grouping (by site, rack, device type).
  - Add simple path visualisation (sender → receiver → destination network segment).

- [ ] **Policy-based routing**
  - Allow definition of simple policies (preferred paths, forbidden paths, redundancy groups).
  - Integrate these checks into TAKE / Patch operations with clear warnings.

- [ ] **Deeper SDN / IS‑06 integration**
  - Move beyond `/sdn/ping` to at least one end-to-end example:
    - Query SDN topology, list paths, and show how a flow → SDN route mapping would look.

- [ ] **Automation playbooks**
  - Define reusable “playbooks” (e.g. failover to backup sender, swap studio layouts, maintenance reroute).
  - Attach them to buttons in the UI and cron-like schedules in automation settings.

---

### 5. Operator UX & Workflows

- [ ] **Flow edit UX refinements**
  - Group advanced ST 2110 / NMOS fields behind collapsible sections in the edit modal (basic vs advanced).
  - Provide presets (e.g. “2110-20 video HD”, “2110-30 audio 2.0”, “2110-30 audio 5.1”).

- [ ] **Contextual “How to fix this” hints**
  - Around NMOS / RDS / Port Explorer and diagnostics, add short hints based on common field issues (from the blog).
  - Link to the Migration and Diagnostics sections when relevant.

- [ ] **Multi-panel layouts & pinning**
  - Allow pinning a flow / sender / receiver to a side panel while navigating other views.
  - Make it easy to compare two flows or two receivers side by side.

---

### 6. Deployment, Scalability & Multi‑Tenancy

- [ ] **Profiles for small vs large systems**
  - Provide “single node lab”, “small facility”, and “large facility” example configs (DB sizing, MQTT, registry expectations).

- [ ] **High availability outline**
  - Document how to run GO-NMOS in an HA fashion (DB, MQTT, registry) even if HA management is external (Kubernetes / docker swarm).

- [ ] **Multi‑tenant / multi‑site concepts**
  - Design (even if not fully implement) how multiple logical sites or tenants could be represented (tags, namespaces, site IDs).

---

### 7. Testing, Tooling & Documentation

- [ ] **Deeper automated tests**
  - Add API-level tests for NMOS interactions, health checks, and critical routing workflows.
  - Add basic frontend component tests for the most important views (Flows, Dashboard, Migration).

- [ ] **Developer & operator documentation**
  - Extend README and add a dedicated `docs/` folder:
    - Quickstart, architecture overview, NMOS spec mapping, operations cookbook.

- [ ] **Interoperability test plan**
  - Define how to test GO-NMOS against popular reference registries/nodes.
  - Keep a short “interop matrix” documenting what has been tested.

