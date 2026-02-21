# What is a Flow? How Is It Used with Node / Sender / Receiver?

This document summarises the **Flow** concept, its relationship to **Node / Sender / Receiver**, and how the system routes “source → destination”.

---

## 1. Core concepts (NMOS / ST 2110)

| Concept | Meaning |
|--------|---------|
| **Node** | Physical or virtual device (encoder, decoder, router). Can contain both **senders** and **receivers**. |
| **Device** | A unit inside a node (e.g. a single encoder card). |
| **Sender** | An **output** (source). Publishes a specific **flow** onto the network (multicast IP/port). |
| **Receiver** | An **input** (destination). Receives and consumes a stream from the network. |
| **Flow** | **Definition of the media stream**: format (video/audio), multicast IP, port, SDP, etc. It describes *what* the content is, not *where* it goes. |

In short: **Flow = description of the stream**. **Source = Sender**, **Destination = Receiver**. We do not “route the flow as source or destination”; we choose **which Sender (source) to connect to which Receiver (destination)**; the Sender’s flow is used for that connection.

---

## 2. Relationship chain

```
Node
 └── Device
      ├── Sender  →  flow_id  →  Flow (video/audio stream definition)
      └── Receiver (format: video/audio; “receives” the flow)
```

- **Sender** is always tied to **exactly one Flow** (`flow_id`). That flow’s address (multicast IP, port) is where the sender is publishing.
- **Receiver** accepts a **format** (video/audio); **which flow it receives** is set via **IS-05 TAKE/PATCH**: “Connect this receiver to that sender (and thus to its flow).”

So:
- **Flow** = what is being published (content + address)
- **Sender** = output that publishes this flow (source)
- **Receiver** = input that will receive this flow (destination)
- **Routing** = “Connect this Sender (source) to this Receiver (destination)” → the flow is the content carried on that connection.

---

## 3. Who creates the flow? Is it from another system or do we add it?

- **Flows do exist and are created first on the device/registry side.** In NMOS/ST 2110, when an encoder (node) publishes, it lists “my flows and senders” in its Node API. If it registers with a registry, the flow list also lives in the **registry**. So the **existence and definition** of the flow (id, label, format) are created elsewhere (device + registry).
- **What do we do?** When we discover the node and sync the registry, we see the **existing** flow list. We do not “invent” the flow; we **keep a copy in our own database** (internal flow). Reasons: (1) keep address info (multicast/port) in one place, (2) show it in the Flows tab, (3) send that address in TAKE via IS-05. So **the flow is created elsewhere; we add it as a record in our system** (e.g. “create flow from registry” or fill address from SDP).
- **We can also create flows ourselves:** You can add a flow **manually** from the Flows tab (e.g. for multicast planning when no device exists yet). Then the flow exists “only in our system” and can later be matched to a sender that uses that flow_id.

Summary: **The device/registry side creates the flow first; we add the existing flow as a record (and optionally fill address from SDP/manual). You can also create flows only on our side.**

---

## 4. Two kinds of “flow” in the system

### 4.1 NMOS (IS-04) flow – definition from the registry

- **Flow** record from the Node API / registry: id, label, format, `source_id` (which sender produces this flow).
- It is only a **definition**; address info (multicast IP/port) usually comes from the **Sender’s manifest (SDP)** or IS-05.

### 4.2 Internal flow – record in GO-NMOS-PRO database

- The record you see in the **Flows** tab: `display_name`, `multicast_ip`, `source_ip`, `port`, `transport_protocol`, etc.
- **TAKE** sends **transport_params** (multicast_ip, source_ip, destination_port) from this internal flow to IS-05.
- A **sender** from the registry is matched by its `flow_id` to an internal flow; if none exists, an internal flow is created from registry flow data (create flow from registry).

Summary: **Address info we use for routing = internal flow.** NMOS flow = “what is this stream”; internal flow = “where does it go (IP/port).”

---

### 4.3 Where do multicast IP and port come from? (Who do we give them to?)

- **We don’t “give” them to someone else.** These addresses are **where the device (encoder/sender) is publishing**. The encoder is already sending a stream to a multicast group (e.g. 239.0.0.1) and port (e.g. 5004).
- **How does the system get this info?**
  1. **SDP (manifest):** In NMOS, each Sender has a `manifest_href` (URL of the SDP). GO-NMOS-PRO fetches and parses that SDP and extracts multicast_addr_a, source_addr_a, group_port_a into the internal flow’s multicast_ip, source_ip, port. So **source of address = device’s SDP**.
  2. **Creating flow from registry:** When creating a flow from the registry, the sender’s manifest_href is stored as sdp_url on the flow; you can then “Fetch SDP” to fill multicast/port.
  3. **Manual flow:** If you create a flow from the Flows tab, you enter multicast IP and port yourself (planned address).
- **What happens on TAKE?** The internal flow’s multicast_ip, source_ip, port are sent to the **receiver** via IS-05 PATCH: “Receive from this address (multicast:port).” So this info is **given to the receiver**; the receiver joins that multicast and listens.

In short: **Multicast/port = encoder’s publish address.** It is written to the flow from SDP (or manual); on TAKE it is sent to the receiver as “receive from this address”.

### 4.4 One flow per node? How many flows should there be?

- **No.** The number of flows depends on the **number of streams**, not the number of nodes.
- **Rule:** Each **Sender** has exactly **one** flow_id. So **1 Sender = 1 Flow** (the single stream that sender publishes).
- A **node** can have multiple **Senders** (e.g. 1 video + 1 audio); that node then has **multiple flows**. Example: mock-node-1 has “Video Sender 1” (Flow 1) and “Audio Sender 1” (Flow 2).
- In practice: When you add a node, registry sync brings that node’s **senders**; each sender already has an **NMOS flow** (in the registry). The internal flow holds the **address info** (multicast, port) needed for TAKE/IS-05; it is created from SDP or “create flow from registry”. So **one internal flow per sender** is the goal; there is no “one flow per node” requirement.

---

## 5. Registry & patch flow (source → destination)

1. Select a **node** from the Sender Nodes list → that node’s **senders** are listed.
2. Select a **sender** → the **internal flow** matching this sender’s **flow_id** is found (or created from the registry).
3. Select a **node** from the Receiver Nodes list → that node’s **receivers** are listed.
4. Select a **receiver**.
5. Click **TAKE PATCH**:
   - The backend sends a **PATCH** to IS-05 with the selected **internal flow’s** multicast_ip, source_ip, port: “Connect this receiver to this sender; use this transport.”
   - Result: **Source (Sender) → destination (Receiver)** is connected; the content carried is that flow.

So we do **not** “route the flow as source/destination”; we choose **source = Sender** and **destination = Receiver**, and the **Sender’s flow** is used on that connection.

---

## 6. Summary table

| Question | Answer |
|----------|--------|
| What is the flow for? | It describes **what** the media stream is (format, address: multicast IP/port). On TAKE, “receive from this address” is sent from here to the receiver. |
| Do we route the flow as source/destination? | No. **Source = Sender**, **Destination = Receiver**. The flow is the **content (and address)** carried between them. |
| How is it used with the node? | Node → Device → Device has Sender(s) and/or Receiver(s). Sender is tied to a flow; we connect the receiver to that sender (and thus to that flow) via TAKE. |
| Why is there an internal flow? | To hold address (multicast_ip, port), display_name, etc., and to use them in the IS-05 PATCH. The registry flow is often definition only; address lives in SDP/internal flow. |

---

## 7. Flow info and “stream active” display after TAKE

- The **TAKE PATCH** response now includes **flow** info: `flow.display_name`, `flow.multicast_ip`, `flow.port`. So “which flow was connected” and address info can be shown in the UI.
- In the **receiver list** (Registry & Patch), for connected receivers both **Receiving: &lt;Sender&gt;** and **Flow: &lt;display_name&gt; (multicast_ip:port)** are shown. This comes from the `flow` field in the `GET /api/nmos/receivers-active` response (backend fills it from receiver_connections + flow record).
- So “flow is active” / “this receiver is receiving this flow” is visible both in the TAKE response and on the receiver card.

---

## 8. Step-by-step usage in practice

1. **Add nodes:** Use Port Explorer or RDS to discover encoder/decoder nodes; add them with “Register”. Each node’s sender/receiver list is synced.
2. **Filling in flows:** Each sender has a flow_id. On first TAKE, if there is no internal flow for that flow_id, the system **creates a flow from the registry** (display_name, nmos_flow_id). Multicast/port may be empty; then open that flow in the **Flows** tab and use **Fetch SDP** (via manifest_href) → SDP is parsed and multicast_ip and port are written to the flow. Or enter multicast/port manually.
3. **TAKE (routing):** In Registry & Patch, choose **Sender Node** → a **Sender**, **Receiver Node** → a **Receiver**, then click **TAKE PATCH**. The system finds the selected sender’s flow (internal flow), and sends multicast_ip, source_ip, port to the receiver via IS-05. The receiver receives from that address.
4. **Summary:** Add node → senders (and flow_ids) appear → optionally fill flow with SDP or manual multicast/port → use TAKE to connect sender to receiver. Each **sender** has its own **flow** (and optionally its own internal flow record); the number of streams matters, not the number of nodes.

---

## 9. In one sentence

- **Flow** = Definition of the stream (content + address).
- **Source** = **Sender** (output that publishes the flow).
- **Destination** = **Receiver** (input that receives the flow).
- **Routing** = “Connect this Sender to this Receiver” (TAKE); the flow is the stream used on that connection.

We do not select the flow separately as “source or destination”; we select Sender (source) and Receiver (destination), and the flow is determined by the Sender.
