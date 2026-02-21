# NMOS Orchestration – Design Checklist

This document summarises what is needed to design and operate GO-NMOS-PRO as an **NMOS orchestration** platform (central controller for an All-IP / ST 2110 campus).

---

## 1. What “NMOS orchestration” means here

- **Single control plane:** One system (GO-NMOS) that discovers NMOS resources, decides connections, and executes IS-05 on behalf of operators.
- **Multi-registry:** Can talk to several IS-04 Registries (prod, lab, remote sites) and present a unified view.
- **Policies & automation:** Routing rules (allowed/forbidden pairs, path requirements), scheduled actions, playbooks, and maintenance windows.
- **Observability:** Health, metrics, logs, alerts so that operations and SRE can run the plant.

So “orchestration” = **discover → decide (policies) → execute (IS-05 / playbooks) → observe**.

---

## 2. What you already have (orchestration pillars)

| Pillar | What you have |
|--------|----------------|
| **Discovery** | Multi-registry config (DB + `/api/registry/config`), registry discovery, internal snapshot of nodes/devices/senders/receivers, IS-04 compatibility matrix, site/room/domain on resources. |
| **Topology & view** | Topology view (group by site/room), flow list with site/room, multi-site filters, cross-site routing view. |
| **Connection execution** | IS-05 TAKE/PATCH (single + bulk), receiver active/disable (BCC-style), staged activations, scheduled activations runner. |
| **Policies** | Routing policies (allowed_pair, forbidden_pair, path_requirement, constraint), policy check before TAKE, override + audit, maintenance windows with policy binding. |
| **Automation** | Automation jobs (collision check, NMOS check), scheduled playbooks, playbook definitions (failover, studio swap, etc.), role-aware execution. |
| **System parameters** | IS-09-style system parameters (PTP, expected IS-04/05 versions, domain), system validation (D.2), diagnostics. |
| **SDN / IS-06** | SDN controller config, topology/paths (stub/demo), flow ↔ path association, SDN ping. |
| **Audio / events** | IS-08 proxy (channel mapping, presets), audio chain view, IS-07 events + sources, Events panel. |
| **Observability** | Structured logs, Logs UI, Prometheus metrics, Grafana example, alert hooks (webhook/Slack), health detail, routing/registry metrics. |
| **Deployment** | Env profiles (lab, small facility, large campus), HA guide, backup/restore (config, policies, registry). |

So the **design** for NMOS orchestration is already in place: central controller, multi-registry, policies, automation, observability.

---

## 3. What you may still need (to strengthen orchestration)

- **Operational clarity**
  - **Single “orchestrator” identity:** Document clearly that GO-NMOS is the **only** system that performs IS-05 TAKE/PATCH for the campus (or define which actions are allowed from other tools). Prevents conflicting controllers.
  - **Registry roles:** Use “primary / secondary / lab” (or similar) in registry config and in UI so that operators know which registry is authoritative for which site.
  - **Read-only vs control:** Ensure “viewer” cannot run TAKE/playbooks; only editor/admin (or operator role) can. Already partially there; verify and document.

- **Policy and safety**
  - **Default-deny or default-allow:** Decide whether a connection is allowed when no policy matches (e.g. default-deny in critical areas, default-allow in lab). Implement and document.
  - **Policy scope:** Optionally bind policies to site/room or registry so that “Studio A” rules do not apply to “Lab.” You have constraints; adding scope (site/room/registry) would make orchestration rules clearer.
  - **Audit and compliance:** You have routing policy audit; ensure all TAKE/bulk/scheduled actions are logged with user, time, override flag, and policy result. Use for compliance and “who did what.”

- **Lifecycle and resilience**
  - **Registry failover:** If primary registry is down, can the orchestrator switch to a secondary (or read-only cache) and still show topology? Document or implement a simple failover/fallback strategy.
  - **Connection recovery:** After registry or node restart, do you want the orchestrator to “re-apply” last known good connections (from your connection state/history)? If yes, define a small “recovery” playbook or background job.
  - **Maintenance windows:** You have maintenance windows and routing policy binding; ensure operators can see “active maintenance window” and which policy is in effect in the UI.

- **Integration and APIs**
  - **External orchestration:** If another system (e.g. broadcast automation) must trigger a connection or playbook, expose stable REST APIs (e.g. “run playbook X with params”) and document them as the “orchestration API.”
  - **IS-07 / tally:** Use IS-07 events (tally, GPI) to drive or validate routing (e.g. “do not route to output if tally shows on air”). You have events; tie them more explicitly to policy or UI warnings.

- **Documentation and training**
  - **Runbooks:** One-page runbooks for “daily ops,” “failover,” “add new registry,” “change routing policy,” “investigate connection failure.” Links from Diagnostics or Settings.
  - **Orchestration diagram:** Keep the diagram in `AI_HANDOFF.md` (and optionally in docs) as the single picture of “GO-NMOS as orchestrator” and update it when you add registries or SDN.

---

## 4. Minimal checklist to “design as NMOS orchestration”

- [ ] **Define orchestrator role:** Document that GO-NMOS is the central IS-05 controller; no other system does TAKE for the same receivers (or document exceptions).
- [ ] **Registry strategy:** At least one registry (or registry pair) per logical site; config and roles in DB and UI.
- [ ] **Policy model:** Default behaviour (allow/deny when no rule matches); policies cover critical paths (e.g. TX outputs, MCR).
- [ ] **Audit:** Every connection change (and override) logged with user, time, policy result.
- [ ] **Automation:** At least one playbook (e.g. failover) and one scheduled job (e.g. collision check) so that “orchestration” is not only manual TAKE.
- [ ] **Observability:** Health of registries and key nodes visible; alerts on registry down / repeated connection failures; logs searchable.
- [ ] **Recovery/failover:** Document (or implement) what happens when a registry or the controller restarts (reconnect, re-sync, optional re-apply of connections).

---

## 5. Summary

- You **already have** the core of NMOS orchestration: multi-registry, IS-05 execution, routing policies, playbooks, scheduling, maintenance windows, multi-site view, observability.
- To **design and run** the system explicitly as **NMOS orchestration**, focus on: clear orchestrator identity, policy defaults and scope, audit, registry failover/recovery, and runbooks so that both humans and external systems know how the plant is controlled.

For the full roadmap (A–G), see `TODO.md`. For architecture and handoff context, see `AI_HANDOFF.md`.
