# AMWA NMOS Compliance Note

This document summarises how G0-NMOS-PRO and the virtualtest mock environment align with **AMWA NMOS** (IS-04, IS-05) and **ST 2110 / SDP** standards. Use it as a reference when integrating with real devices and controllers.

---

## 1. IS-04 Discovery and Registration (v1.3)

**Source:** [specs.amwa.tv/is-04](https://specs.amwa.tv/is-04)

### 1.1 Node API

- **Base path:** `/x-nmos/node/{version}/` (e.g. `v1.3`)
- **Self:** `GET /x-nmos/node/v1.3/self` — `id`, `label`, `api.endpoints`, `clocks`, etc.
- **Resources:** `devices`, `flows`, `senders`, `receivers` (and optionally `sources`) are exposed under the same version.

In this project, mock nodes use these paths; the registry and backend use the same version (v1.3).

### 1.2 Format URNs (Parameter Register)

- Video: `urn:x-nmos:format:video`
- Audio: `urn:x-nmos:format:audio`

Mocks use these format URNs.

### 1.3 Flow resource

- **Required / common fields:** `id`, `source_id`, `device_id`, `parents`, `format`, `label`, `description`, `version`, `media_type`
- **Video (raw):** Per IS-04 examples and flow_video schema, `media_type`: **`video/raw`** (uncompressed video). Frame info: `frame_width`, `frame_height`, `interlace_mode`, `grain_rate`, optional `colorspace`, `components`.
- **Audio (raw):** `media_type`: **`audio/L24`**, `sample_rate`, `channels`.

Mock flows use `media_type: "video/raw"` for video and `"audio/L24"` for audio, matching IS-04 flow_video_raw / flow_audio_raw.

### 1.4 Sender resource

- **Required:** `id`, `label`, `flow_id`, `device_id`, `transport`, **`manifest_href`** (SDP or transport file URL), `version`
- **Transport:** For RTP multicast, `urn:x-nmos:transport:rtp.mcast`

Mocks provide an SDP URL via `manifest_href` for each sender; transport is `urn:x-nmos:transport:rtp.mcast`.

### 1.5 Receiver resource

- **Required:** `id`, `label`, `device_id`, `format`, `transport`, `version`

Mocks define receivers with these fields.

---

## 2. IS-05 Device Connection Management (v1.1)

**Source:** [specs.amwa.tv/is-05](https://specs.amwa.tv/is-05)

### 2.1 Connection API

- **Base:** `/x-nmos/connection/{version}/`
- **Single-legged:** `/single/senders`, `/single/receivers`, `/single/senders/{id}/staged|active`, `/single/receivers/{id}/staged|active`
- **Transport file:** The sender’s SDP is obtained via `manifest_href` or the IS-05 transportfile endpoint; Content-Type **`application/sdp`** is recommended.

The backend and mocks align with this behaviour using staged/active and transport_file (SDP).

### 2.2 RTP transport parameters (IS-05 Behaviour – RTP)

- **Sender core:** `source_ip`, `destination_ip`, `source_port`, `destination_port`
- **Receiver core:** `source_ip`, `destination_port`, `interface_ip`; for multicast, `multicast_ip` (destination)

The backend snapshot and sync use these parameter names (source_ip, destination_ip, destination_port, etc.).

### 2.3 SDP rules (RFC 4566 + RFC 4570)

- **RFC 4566:** `v=`, `o=`, `s=`, `t=`, `m=`, `c=`, `a=` lines.
- **RFC 4570 (Source Filters):** For multicast, **`a=source-filter: incl IN IP4 <multicast> <source_ip>`** must be used.

In mock SDPs:

- `v=0`, `o=- 0 0 IN IP4 <source_ip>`, `s=`, `t=0 0`
- `m=video|audio <port> RTP/AVP 96`
- `c=IN IP4 <multicast>/32`
- **`a=source-filter: incl IN IP4 <multicast> <source_ip>`**
- `a=rtpmap:96 ...` (video: smpte291/90000, audio: L24/48000/2)
- `a=rtcp:...`, `a=sendrecv`, `a=ts-refclk:ptp=...` (for ST 2110 alignment)

This structure is consistent with IS-05 RTP behaviour and ST 2110 practice.

---

## 3. ST 2110 and SDP

- **Video (ST 2110-20):** SDP uses `a=rtpmap` with `smpte291/90000`; at flow level, IS-04 `media_type` is **`video/raw`**.
- **Audio (ST 2110-30):** PCM audio is carried in a single RTP stream; **not one flow per channel** — one flow = one RTP stream = N channels. In SDP: **`a=rtpmap:96 L24/48000/<N>`** (N = channel count; stereo = 2). In the flow: **`audio/L24`**, `sample_rate` (48 kHz), `channels` (e.g. 2). Example: 16 stereo inputs = 16 receivers (each subscribing to one stereo flow), 2 stereo outputs = 2 senders + 2 flows (each flow 2 channels).
- **PTP:** `a=ts-refclk:ptp=IEEE1588-2008:...` is consistent with ST 2110-10.

The backend SDP parser reads `c=`, `m=`, `a=source-filter`, and `a=rtpmap`; multicast, source IP, and port are correctly applied to the flow record.

---

## 4. Summary table

| Component            | Standard / Source | Compliance |
|----------------------|------------------|------------|
| Node API path        | IS-04 v1.3       | `/x-nmos/node/v1.3/` |
| Format URN           | NMOS Formats     | `urn:x-nmos:format:video|audio` |
| Flow media_type      | flow_video_raw / flow_audio_raw | `video/raw`, `audio/L24` |
| Sender manifest_href | IS-04            | SDP URL provided |
| Transport            | Transports register | `urn:x-nmos:transport:rtp.mcast` |
| SDP source-filter    | RFC 4570 / IS-05 | `a=source-filter: incl IN IP4 ...` |
| IS-05 transport params | IS-05 RTP     | source_ip, destination_ip, ports |

---

## 5. Testing and validation

- **NMOS Testing Tool:** [specs.amwa.tv/nmos-testing](https://specs.amwa.tv/nmos-testing) — Node and Connection API tests.
- **SDPoker:** AMWA SDP validator; use it to check SDPs against ST 2110 / RFC expectations.

This document is maintained against AMWA specifications and parameter registers. Always refer to [specs.amwa.tv](https://specs.amwa.tv) for the latest authoritative text.

---

## 6. Project-wide compliance audit (summary)

### 6.1 Fixes applied

- **mock-node-1:** Flow `media_type` changed from `"video/smpte291"` to **`"video/raw"`** (IS-04 flow_video_raw).
- **mock-studio1, mock-tx, mock-multiviewer:** Already use `video/raw` and `audio/L24`.

### 6.2 Compliant components

| Component | Description |
|-----------|-------------|
| **Node API** | All mocks expose self, devices, flows, senders, receivers under `/x-nmos/node/v1.3/`; format URN, transport URN, manifest_href, device_id are correct. |
| **Flow media_type** | Video: `video/raw`, audio: `audio/L24` (NMOS flow_video_raw / flow_audio_raw). |
| **SDP** | RFC 4566 + RFC 4570 source-filter; `c=`, `m=`, `a=rtpmap` (smpte291/90000, L24/48000/2); PTP ts-refclk; Content-Type: `application/sdp`. |
| **IS-05** | Staged/active, transport_file (type: application/sdp); backend snapshot uses source_ip, destination_ip, destination_port (IS-05 RTP parameter names). |
| **Query API** | Backend serves nodes, devices, flows, senders, receivers under `/x-nmos/query/v1.3/`; responses include id, label, format, device_id, manifest_href, etc. |
| **Receiver device_id** | All mock receivers have `device_id`; backend applies a single-device fallback when missing. |

### 6.3 Optional / known limitations

- **Query API flow response:** Currently returns only `id`, `label`, `description`, `format`, `source_id`, `tags`, `meta`. The IS-04 flow schema also includes `version` and `parents`; these are not stored in internal registry tables. Many controllers work with this subset; full schema alignment could add version/parents (and optionally device_id) to `nmos_flows` later.
- **Source resource:** In IS-04, the flow’s `source_id` references a Source resource; this project does not expose a separate Source API. Nodes expose only devices, flows, senders, and receivers, which is sufficient for many scenarios.
- **IS-05 version:** Mocks use Connection API `v1.0` or `v1.1`; the backend tries paths compatible with both versions.

This section summarises the project-wide checks for AMWA-NMOS and ST 2110 compliance; use the NMOS Testing Tool and SDPoker for additional testing.
