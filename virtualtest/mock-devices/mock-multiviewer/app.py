#!/usr/bin/env python3
"""
Mock Campus-A Multiviewer Node – 16 inputs (receivers), 4 outputs (senders).
ST 2110 SDPs for each output.
"""

import os
import uuid
from datetime import datetime
from flask import Flask, jsonify, request
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

PORT = int(os.getenv('PORT', 8092))
IS04_VERSION = os.getenv('IS04_VERSION', 'v1.3')
IS05_VERSION = os.getenv('IS05_VERSION', 'v1.0')
NODE_ID = os.getenv('NODE_ID', '550e8400-e29b-41d4-a716-446655440030')
NODE_LABEL = os.getenv('NODE_LABEL', 'Campus-A Multiviewer')
SOURCE_IP = os.getenv('SOURCE_IP', '192.168.1.30')

DEVICE_ID = str(uuid.uuid4())

def _uuid(name):
    return str(uuid.uuid5(uuid.NAMESPACE_DNS, f"mv.{name}"))

# 4 output senders (multiviewer walls); IS-04 flow_video_raw: media_type "video/raw"
SENDER_SPECS = [
    ("MV Output 1", "urn:x-nmos:format:video", "video/raw", "239.3.1.1", 5004, "mv_out1"),
    ("MV Output 2", "urn:x-nmos:format:video", "video/raw", "239.3.1.2", 5004, "mv_out2"),
    ("MV Output 3", "urn:x-nmos:format:video", "video/raw", "239.3.1.3", 5004, "mv_out3"),
    ("MV Output 4", "urn:x-nmos:format:video", "video/raw", "239.3.1.4", 5004, "mv_out4"),
]

FLOWS = []
SENDERS = []
for label, format_urn, media_type, multicast, port, sdp_name in SENDER_SPECS:
    fid = _uuid(f"flow.{sdp_name}")
    sid = _uuid(f"sender.{sdp_name}")
    FLOWS.append({
        "id": fid, "source_id": sid, "device_id": DEVICE_ID, "parents": [],
        "format": format_urn, "media_type": media_type, "label": label + " Flow",
        "description": f"ST 2110 {media_type}", "tags": {"site": ["Campus-A"], "room": ["Multiviewer"]},
        "version": "1523456789:0",
        "bit_rate": 1000000000, "frame_width": 1920, "frame_height": 1080,
    })
    SENDERS.append({
        "id": sid, "label": label, "description": f"Multiviewer {label}", "flow_id": fid, "device_id": DEVICE_ID,
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "manifest_href": f"http://localhost:{PORT}/sdp/{sdp_name}.sdp",
        "tags": {"site": ["Campus-A"], "room": ["Multiviewer"]}, "version": "1523456789:0",
        "subscription": {"receiver_id": None, "active": False},
        "_sdp_name": sdp_name, "_media_type": media_type, "_multicast": multicast, "_port": port,
    })

RECEIVERS = [
    {
        "id": _uuid(f"receiver.mv_in{i}"),
        "label": f"MV Input {i}",
        "description": f"Multiviewer input {i}",
        "device_id": DEVICE_ID,
        "format": "urn:x-nmos:format:video",
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "tags": {"site": ["Campus-A"], "room": ["Multiviewer"]},
        "version": "1523456789:0",
        "subscription": {"sender_id": None, "active": False},
    }
    for i in range(1, 17)
]

DEVICES = [{
    "id": DEVICE_ID,
    "label": "Multiviewer Device",
    "description": "Campus-A Multiviewer 16in/4out",
    "node_id": NODE_ID,
    "type": "urn:x-nmos:device:generic",
    "tags": {"site": ["Campus-A"], "room": ["Multiviewer"]},
    "version": "1523456789:0",
    "controls": {"href": f"http://localhost:{PORT}/x-nmos/connection/{IS05_VERSION}/"},
}]

CONNECTION_STATE = {"senders": {}, "receivers": {}}

def make_sdp_video(multicast_ip, port, source_ip, label="Video"):
    return (
        "v=0\r\n"
        f"o=- 0 0 IN IP4 {source_ip}\r\n"
        f"s={label}\r\n"
        "t=0 0\r\n"
        f"m=video {port} RTP/AVP 96\r\n"
        f"c=IN IP4 {multicast_ip}/32\r\n"
        f"a=source-filter: incl IN IP4 {multicast_ip} {source_ip}\r\n"
        "a=rtpmap:96 smpte291/90000\r\n"
        f"a=rtcp:{port + 1}\r\n"
        "a=sendrecv\r\n"
        "a=ts-refclk:ptp=IEEE1588-2008:00-1B-63-FF-FE-FF-FF-FF\r\n"
    )

SDP_BY_NAME = {s["_sdp_name"]: make_sdp_video(s["_multicast"], s["_port"], SOURCE_IP, s["label"]) for s in SENDERS}

def _strip_private(obj):
    if isinstance(obj, dict):
        return {k: _strip_private(v) for k, v in obj.items() if not k.startswith("_")}
    if isinstance(obj, list):
        return [_strip_private(x) for x in obj]
    return obj

# --- IS-04 ---
@app.route('/x-nmos/node/', methods=['GET'])
def node_versions_root():
    return jsonify([IS04_VERSION])

@app.route(f'/x-nmos/node/{IS04_VERSION}/', methods=['GET'])
def node_versions():
    return jsonify([IS04_VERSION])

@app.route(f'/x-nmos/node/{IS04_VERSION}/self', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/self/', methods=['GET'])
def node_self():
    return jsonify({
        "id": NODE_ID, "label": NODE_LABEL,
        "description": "Campus-A Multiviewer 16 inputs, 4 outputs",
        "hostname": "multiviewer.local",
        "api": {"endpoints": [{"host": "localhost", "port": PORT, "protocol": "http"}], "versions": [IS04_VERSION]},
        "caps": {}, "clocks": [{"name": "clk0", "ref_type": "internal"}],
        "tags": {"site": ["Campus-A"], "room": ["Multiviewer"]}, "version": "1523456789:0",
    })

@app.route(f'/x-nmos/node/{IS04_VERSION}/devices', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/devices/', methods=['GET'])
def node_devices():
    return jsonify(DEVICES)

@app.route(f'/x-nmos/node/{IS04_VERSION}/flows', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/flows/', methods=['GET'])
def node_flows():
    return jsonify(_strip_private(FLOWS))

@app.route(f'/x-nmos/node/{IS04_VERSION}/senders', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/senders/', methods=['GET'])
def node_senders():
    return jsonify(_strip_private(SENDERS))

@app.route(f'/x-nmos/node/{IS04_VERSION}/receivers', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/receivers/', methods=['GET'])
def node_receivers():
    return jsonify(RECEIVERS)

# --- IS-05 ---
@app.route('/x-nmos/connection/', methods=['GET'])
@app.route('/x-nmos/connection', methods=['GET'])
def connection_versions_root():
    return jsonify([IS05_VERSION])

@app.route(f'/x-nmos/connection/{IS05_VERSION}/', methods=['GET'])
def connection_versions():
    return jsonify([IS05_VERSION])

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders', methods=['GET'])
def connection_senders():
    out = []
    for s in SENDERS:
        sid = s["id"]
        state = CONNECTION_STATE.get("senders", {}).get(sid, {})
        out.append({
            "id": sid,
            "master_enable": state.get("master_enable", False),
            "activation": state.get("activation", {"mode": "activate_immediate", "requested_time": None}),
            "transport_file": state.get("transport_file", {"data": "", "type": "application/sdp"}),
            "receiver_id": state.get("receiver_id"),
        })
    return jsonify(out)

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers', methods=['GET'])
def connection_receivers():
    out = []
    for r in RECEIVERS:
        rid = r["id"]
        state = CONNECTION_STATE.get("receivers", {}).get(rid, {})
        out.append({
            "id": rid,
            "master_enable": state.get("master_enable", False),
            "activation": state.get("activation", {"mode": "activate_immediate", "requested_time": None}),
            "transport_file": state.get("transport_file", {"data": "", "type": "application/sdp"}),
            "transport_params": state.get("transport_params", []),
            "sender_id": state.get("sender_id"),
        })
    return jsonify(out)

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers/<receiver_id>/active', methods=['GET'])
@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers/<receiver_id>/active/', methods=['GET'])
def receiver_active(receiver_id):
    state = CONNECTION_STATE.get("receivers", {}).get(receiver_id, {})
    return jsonify({
        "master_enable": state.get("master_enable", False),
        "activation": state.get("activation", {"mode": "activate_immediate", "requested_time": None}),
        "transport_file": state.get("transport_file", {"data": "", "type": "application/sdp"}),
        "transport_params": state.get("transport_params", []),
        "sender_id": state.get("sender_id"),
    })

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers/<receiver_id>/staged', methods=['GET', 'PATCH'])
@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers/<receiver_id>/staged/', methods=['GET', 'PATCH'])
def receiver_staged(receiver_id):
    state = CONNECTION_STATE.setdefault("receivers", {}).setdefault(receiver_id, {})
    if request.method == 'GET':
        return jsonify({
            "master_enable": state.get("master_enable", False),
            "activation": state.get("activation", {"mode": "activate_immediate", "requested_time": None}),
            "transport_file": state.get("transport_file", {"data": "", "type": "application/sdp"}),
            "transport_params": state.get("transport_params", []),
            "sender_id": state.get("sender_id"),
        })
    data = request.get_json(force=True) or {}
    for key in ("master_enable", "activation", "transport_file", "transport_params", "sender_id"):
        if key in data:
            state[key] = data[key]
    return jsonify({
        "master_enable": state.get("master_enable", False),
        "activation": state.get("activation", {"mode": "activate_immediate", "requested_time": None}),
        "transport_file": state.get("transport_file", {"data": "", "type": "application/sdp"}),
        "transport_params": state.get("transport_params", []),
        "sender_id": state.get("sender_id"),
    }), 200

@app.route('/sdp/<filename>', methods=['GET'])
def sdp_file(filename):
    name = filename.replace(".sdp", "") if filename.endswith(".sdp") else filename
    if name in SDP_BY_NAME:
        return SDP_BY_NAME[name], 200, {"Content-Type": "application/sdp"}
    return jsonify({"error": "SDP not found"}), 404

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "healthy", "node_id": NODE_ID, "timestamp": datetime.utcnow().isoformat()})

if __name__ == '__main__':
    print(f"Starting {NODE_LABEL} on port {PORT}")
    app.run(host='0.0.0.0', port=PORT, debug=True)
