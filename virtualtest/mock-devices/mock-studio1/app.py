#!/usr/bin/env python3
"""
Mock Campus-A Studio-1 Node – full broadcast scenario.
4 cameras, 4 playout video, 2 CG (fill+key), mic-1 stereo, 4 playout audio,
Vision Mixer 16 inputs, PGM + Clean outputs.
Audio Mixer (ST 2110-30): 16 stereo inputs (receivers), 2 stereo outputs (senders).
ST 2110: video = one flow per stream; audio = one flow per RTP stream, L24/48000/N (N=channels).
Stereo = one flow with 2 channels in a single RTP stream (not one flow per channel).
ST 2110 SDPs with multicast_ip, source_ip, port for each flow.
"""

import os
import json
import uuid
from datetime import datetime
from flask import Flask, jsonify, request
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

PORT = int(os.getenv('PORT', 8090))
IS04_VERSION = os.getenv('IS04_VERSION', 'v1.3')
IS05_VERSION = os.getenv('IS05_VERSION', 'v1.0')
NODE_ID = os.getenv('NODE_ID', '550e8400-e29b-41d4-a716-446655440010')
NODE_LABEL = os.getenv('NODE_LABEL', 'Campus-A Studio-1')
SOURCE_IP = os.getenv('SOURCE_IP', '192.168.1.10')  # SDP source address

DEVICE_ID = str(uuid.uuid4())
# Audio Mixer: second device on same node (16 stereo inputs, 2 stereo outputs)
AUDIO_MIXER_DEVICE_ID = str(uuid.uuid5(uuid.NAMESPACE_DNS, "studio1.audiomixer.device"))

# --- Sender definitions: (label, format_urn, media_type, multicast, port, sdp_name) ---
# IS-04 flow_video_raw: media_type "video/raw"; flow_audio_raw: "audio/L24" (NMOS parameter register)
SENDER_SPECS = [
    # 4 cameras
    *[("Cam " + str(i), "urn:x-nmos:format:video", "video/raw", f"239.1.1.{i}", 5004, f"cam{i}") for i in range(1, 5)],
    # 4 playout video
    *[("Playout " + str(i), "urn:x-nmos:format:video", "video/raw", f"239.1.1.{4+i}", 5004, f"playout_v{i}") for i in range(1, 5)],
    # 2 CG fill + key
    ("CG Fill 1", "urn:x-nmos:format:video", "video/raw", "239.1.1.9", 5004, "cg_fill_1"),
    ("CG Key 1", "urn:x-nmos:format:video", "video/raw", "239.1.1.10", 5004, "cg_key_1"),
    ("CG Fill 2", "urn:x-nmos:format:video", "video/raw", "239.1.1.11", 5004, "cg_fill_2"),
    ("CG Key 2", "urn:x-nmos:format:video", "video/raw", "239.1.1.12", 5004, "cg_key_2"),
    # Mic 1 stereo
    ("Mic 1 L-R", "urn:x-nmos:format:audio", "audio/L24", "239.1.2.1", 5004, "mic1"),
    # 4 playout audio (stereo each)
    *[("Playout Audio " + str(i), "urn:x-nmos:format:audio", "audio/L24", f"239.1.2.{1+i}", 5004, f"playout_a{i}") for i in range(1, 5)],
    # Vision mixer outputs
    ("VM PGM", "urn:x-nmos:format:video", "video/raw", "239.1.3.1", 5004, "vm_pgm"),
    ("VM Clean", "urn:x-nmos:format:video", "video/raw", "239.1.3.2", 5004, "vm_clean"),
]

# Audio Mixer outputs: 2 stereo (ST 2110-30 = one flow per RTP stream, 2 channels in one stream)
AUDIO_MIXER_SENDER_SPECS = [
    ("AM Mix Main L-R", "urn:x-nmos:format:audio", "audio/L24", "239.1.4.1", 5004, "am_mix_main"),
    ("AM Mix Aux L-R", "urn:x-nmos:format:audio", "audio/L24", "239.1.4.2", 5004, "am_mix_aux"),
]

# Build flows and senders with stable UUIDs (seed by label)
def _uuid(name):
    return str(uuid.uuid5(uuid.NAMESPACE_DNS, f"studio1.{name}"))

FLOWS = []
SENDERS = []
for label, format_urn, media_type, multicast, port, sdp_name in SENDER_SPECS:
    fid = _uuid(f"flow.{sdp_name}")
    sid = _uuid(f"sender.{sdp_name}")
    FLOWS.append({
        "id": fid,
        "source_id": sid,
        "device_id": DEVICE_ID,
        "parents": [],
        "format": format_urn,
        "media_type": media_type,
        "label": label + " Flow",
        "description": f"ST 2110 {media_type}",
        "tags": {"site": ["Campus-A"], "room": ["Studio-1"]},
        "version": "1523456789:0",
        "bit_rate": 1000000000 if "video" in format_urn else None,
        "frame_width": 1920 if "video" in format_urn else None,
        "frame_height": 1080 if "video" in format_urn else None,
        "interlace_mode": "progressive" if "video" in format_urn else None,
        "grain_rate": {"numerator": 25, "denominator": 1} if "video" in format_urn else None,
        "sample_rate": {"numerator": 48000, "denominator": 1} if "audio" in format_urn else None,
        "channels": 2 if "audio" in format_urn else None,
    })
    SENDERS.append({
        "id": sid,
        "label": label,
        "description": f"Studio-1 {label}",
        "flow_id": fid,
        "device_id": DEVICE_ID,
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "manifest_href": f"http://localhost:{PORT}/sdp/{sdp_name}.sdp",
        "tags": {"site": ["Campus-A"], "room": ["Studio-1"]},
        "version": "1523456789:0",
        "subscription": {"receiver_id": None, "active": False},
        "_sdp_name": sdp_name,
        "_media_type": media_type,
        "_multicast": multicast,
        "_port": port,
    })

# Audio Mixer: 2 flows + 2 senders (device_id = AUDIO_MIXER_DEVICE_ID)
for label, format_urn, media_type, multicast, port, sdp_name in AUDIO_MIXER_SENDER_SPECS:
    fid = _uuid(f"flow.{sdp_name}")
    sid = _uuid(f"sender.{sdp_name}")
    FLOWS.append({
        "id": fid,
        "source_id": sid,
        "device_id": AUDIO_MIXER_DEVICE_ID,
        "parents": [],
        "format": format_urn,
        "media_type": media_type,
        "label": label + " Flow",
        "description": f"ST 2110-30 {media_type} stereo",
        "tags": {"site": ["Campus-A"], "room": ["Studio-1"], "role": ["audio-mixer"]},
        "version": "1523456789:0",
        "sample_rate": {"numerator": 48000, "denominator": 1},
        "channels": 2,
    })
    SENDERS.append({
        "id": sid,
        "label": label,
        "description": f"Studio-1 Audio Mixer {label}",
        "flow_id": fid,
        "device_id": AUDIO_MIXER_DEVICE_ID,
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "manifest_href": f"http://localhost:{PORT}/sdp/{sdp_name}.sdp",
        "tags": {"site": ["Campus-A"], "room": ["Studio-1"], "role": ["audio-mixer"]},
        "version": "1523456789:0",
        "subscription": {"receiver_id": None, "active": False},
        "_sdp_name": sdp_name,
        "_media_type": media_type,
        "_multicast": multicast,
        "_port": port,
    })

# 16 Vision Mixer inputs (receivers)
RECEIVERS = [
    {
        "id": _uuid(f"receiver.vm_in{i}"),
        "label": f"VM Input {i}",
        "description": f"Vision Mixer input {i}",
        "device_id": DEVICE_ID,
        "format": "urn:x-nmos:format:video",
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "tags": {"site": ["Campus-A"], "room": ["Studio-1"]},
        "version": "1523456789:0",
        "subscription": {"sender_id": None, "active": False},
    }
    for i in range(1, 17)
]

# 16 Audio Mixer inputs (stereo each; ST 2110-30: one receiver subscribes to one stereo flow)
RECEIVERS += [
    {
        "id": _uuid(f"receiver.am_in{i}"),
        "label": f"AM Input {i}",
        "description": f"Audio Mixer input {i} (stereo)",
        "device_id": AUDIO_MIXER_DEVICE_ID,
        "format": "urn:x-nmos:format:audio",
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "tags": {"site": ["Campus-A"], "room": ["Studio-1"], "role": ["audio-mixer"]},
        "version": "1523456789:0",
        "subscription": {"sender_id": None, "active": False},
    }
    for i in range(1, 17)
]

DEVICES = [{
    "id": DEVICE_ID,
    "label": "Studio-1 Device",
    "description": "Campus-A Studio-1 sources and vision mixer",
    "node_id": NODE_ID,
    "type": "urn:x-nmos:device:generic",
    "tags": {"site": ["Campus-A"], "room": ["Studio-1"]},
    "version": "1523456789:0",
    "controls": {"href": f"http://localhost:{PORT}/x-nmos/connection/{IS05_VERSION}/"},
}, {
    "id": AUDIO_MIXER_DEVICE_ID,
    "label": "Studio-1 Audio Mixer",
    "description": "Campus-A Studio-1 Audio Mixer: 16 stereo inputs, 2 stereo outputs (ST 2110-30)",
    "node_id": NODE_ID,
    "type": "urn:x-nmos:device:generic",
    "tags": {"site": ["Campus-A"], "room": ["Studio-1"], "role": ["audio-mixer"]},
    "version": "1523456789:0",
    "controls": {"href": f"http://localhost:{PORT}/x-nmos/connection/{IS05_VERSION}/"},
}]

CONNECTION_STATE = {"senders": {}, "receivers": {}}

# --- ST 2110 SDP generation (backend parser expects c=, m=, a=source-filter) ---
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

def make_sdp_audio(multicast_ip, port, source_ip, label="Audio", channels=2):
    return (
        "v=0\r\n"
        f"o=- 0 0 IN IP4 {source_ip}\r\n"
        f"s={label}\r\n"
        "t=0 0\r\n"
        f"m=audio {port} RTP/AVP 96\r\n"
        f"c=IN IP4 {multicast_ip}/32\r\n"
        f"a=source-filter: incl IN IP4 {multicast_ip} {source_ip}\r\n"
        f"a=rtpmap:96 L24/48000/{channels}\r\n"
        f"a=rtcp:{port + 1}\r\n"
        "a=sendrecv\r\n"
        "a=ts-refclk:ptp=IEEE1588-2008:00-1B-63-FF-FE-FF-FF-FF\r\n"
    )

SDP_BY_NAME = {}
for s in SENDERS:
    name = s["_sdp_name"]
    mcast = s["_multicast"]
    port = s["_port"]
    if "video" in s["_media_type"]:
        SDP_BY_NAME[name] = make_sdp_video(mcast, port, SOURCE_IP, s["label"])
    else:
        SDP_BY_NAME[name] = make_sdp_audio(mcast, port, SOURCE_IP, s["label"])

# --- IS-04 Node API ---
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
        "id": NODE_ID,
        "label": NODE_LABEL,
        "description": "Campus-A Studio-1: cameras, playouts, CG, mic, vision mixer, audio mixer (16 in / 2 out)",
        "hostname": "studio1.local",
        "api": {"endpoints": [{"host": "localhost", "port": PORT, "protocol": "http"}], "versions": [IS04_VERSION]},
        "caps": {},
        "clocks": [{"name": "clk0", "ref_type": "internal"}],
        "tags": {"site": ["Campus-A"], "room": ["Studio-1"]},
        "version": "1523456789:0",
    })

def _strip_private(obj):
    if isinstance(obj, dict):
        return {k: _strip_private(v) for k, v in obj.items() if not k.startswith("_")}
    if isinstance(obj, list):
        return [_strip_private(x) for x in obj]
    return obj

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

# --- IS-05 Connection API (same pattern as mock-node-1) ---
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

# --- SDP files (ST 2110, parsable by backend) ---
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
    print(f"Devices: {len(DEVICES)}, Senders: {len(SENDERS)}, Receivers: {len(RECEIVERS)}, Flows: {len(FLOWS)}")
    app.run(host='0.0.0.0', port=PORT, debug=True)
