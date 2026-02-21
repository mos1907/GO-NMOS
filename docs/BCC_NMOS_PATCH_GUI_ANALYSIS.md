# BCC (nmos-patch-gui) GUI Analysis

What the [taqq505/nmos-patch-gui](https://github.com/taqq505/nmos-patch-gui) interface does and features that can be added to our NMOS Patch.

---

## 1. Active routing indicator after TAKE

**How BCC does it:**
- For each receiver in the receiver list, **GET** `/x-nmos/connection/{ver}/single/receivers/{id}/active` is called.
- The returned `sender_id` indicates which sender is being received; `master_enable` gives the enable/disable state.
- **On the receiver card:** "Receiving: **{sender label}** ({node name})" is shown; if the sender is not in the registry, "Unknown (uuid...)".

**Code (app.js):**
- `refreshReceiverConnections()` → `getReceiverActiveConnection(receiverId)` for all receivers.
- `updateReceiverConnectionDisplay(receiverId, senderId, senderInfo, masterEnable)` → card shows "Receiving: ..." and enable/disable toggle.

**We can add:** In RegistryPatchView, show active connection on the receiver row/cell: "← Sender X (Node A)" and colour (active/inactive).

---

## 2. Enable / Disable (Un-TAKE style)

**How BCC does it:**
- **Receiver:** PATCH `/single/receivers/{id}/staged` body: `{ master_enable: false, activation: { mode: "activate_immediate" } }` → disables the receiver (stops receiving).
- **Sender:** Similarly PATCH senders/{id}/staged with `master_enable: false` → turns off sender output.
- On first enable/disable a **warning modal** (Warning: Receiver Enable/Disable / Sender Enable/Disable) is shown; after confirmation the PATCH is sent.
- On the card a **toggle** (switch) and "Enabled" / "Disabled" label; click calls `toggleReceiverEnable(id, !current)` / `toggleSenderEnable(id, !current)`.

**We can add:**
- "Disable" (un-patch) for receiver: close current connection with master_enable: false + activate_immediate.
- Read and show master_enable on sender/receiver card and add a toggle (optional warning).

---

## 3. Active state on sender side

**How BCC does it:**
- When a sender node is selected, **GET** `/single/senders/{id}/active` is called for each sender.
- `master_enable` is shown on the sender card as "Enabled" / "Disabled" with a toggle.
- Toggle → PATCH senders/{id}/staged to update `master_enable`.

**We can add:** Enable/disable indicator (and optional toggle) on the sender list.

---

## 4. Resource details modal (SDP + Active)

**How BCC does it:**
- Clicking a Sender/Receiver opens a **Resource Details** modal.
- **Active Connection (IS-05):** GET .../active result shown as formatted JSON; Copy button.
- **SDP (Transport File):** For sender, SDP is fetched from manifest_href and shown as text; Copy button.

**We can add:** In the patch panel, when clicking a resource (or "Details"), show SDP and active connection JSON in a modal.

---

## 5. Auto-refresh after patch

**BCC:** On successful TAKE, `refreshReceiverConnections()` is called; receiver list and "Receiving: ..." / enable state are updated.

**We can add:** After TAKE, re-fetch receiver (and sender if needed) connection state and refresh the UI.

---

## 6. IS-05 endpoints used (summary)

| Purpose | Method | Path |
|--------|--------|------|
| Receiver active connection | GET | `.../single/receivers/{id}/active` |
| Receiver staged (read/update) | GET/PATCH | `.../single/receivers/{id}/staged` |
| Receiver disable (un-TAKE) | PATCH | staged + `master_enable: false` + `activation: { mode: "activate_immediate" }` |
| Sender active state | GET | `.../single/senders/{id}/active` |
| Sender enable/disable | PATCH | `.../single/senders/{id}/staged` + master_enable |

---

## 7. Suggested implementation order (our project)

1. **Receiver active connection:** After TAKE (and on page load) GET active for receivers; show "← {sender label}" or similar in the UI.
2. **Receiver Disable (Un-TAKE):** "Disable" button on selected receiver → PATCH staged with `master_enable: false` + activate_immediate.
3. **Sender/Receiver enable state:** Read `master_enable` from GET active and show Enabled/Disabled on card with optional toggle.
4. **Details modal:** Show SDP and Active JSON with Copy.

This document is based on the [nmos-patch-gui](https://github.com/taqq505/nmos-patch-gui) source (app.js, nmos-api.js).
