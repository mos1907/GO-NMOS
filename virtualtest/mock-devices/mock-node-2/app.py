#!/usr/bin/env python3
"""
Mock NMOS Node 2 - Decoder/Receiver
IS-04 Node API + IS-05 Connection API simülatörü
"""

import os
import json
import uuid
from datetime import datetime
from flask import Flask, jsonify, request
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

# Configuration
NODE_ID = os.getenv('NODE_ID', '550e8400-e29b-41d4-a716-446655440002')
NODE_LABEL = os.getenv('NODE_LABEL', 'Mock Decoder Node 2')
NODE_DESCRIPTION = os.getenv('NODE_DESCRIPTION', 'Virtual NMOS decoder for testing')
PORT = int(os.getenv('PORT', 8081))
IS04_VERSION = os.getenv('IS04_VERSION', 'v1.3')
IS05_VERSION = os.getenv('IS05_VERSION', 'v1.0')

# Mock data
DEVICE_ID = str(uuid.uuid4())
RECEIVER_ID_1 = str(uuid.uuid4())
RECEIVER_ID_2 = str(uuid.uuid4())
RECEIVER_ID_3 = str(uuid.uuid4())

# Mock receivers (decoder node has more receivers)
RECEIVERS = [
    {
        "id": RECEIVER_ID_1,
        "label": "Video Receiver 1",
        "description": "Mock video receiver",
        "format": "urn:x-nmos:format:video",
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "device_id": DEVICE_ID,
        "tags": {"site": ["CampusA"], "room": ["ControlRoom1"]},
        "version": "1523456789:123459",
        "subscription": {
            "sender_id": None,
            "active": False
        }
    },
    {
        "id": RECEIVER_ID_2,
        "label": "Video Receiver 2",
        "description": "Mock video receiver",
        "format": "urn:x-nmos:format:video",
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "device_id": DEVICE_ID,
        "tags": {"site": ["CampusA"], "room": ["ControlRoom1"]},
        "version": "1523456789:123460",
        "subscription": {
            "sender_id": None,
            "active": False
        }
    },
    {
        "id": RECEIVER_ID_3,
        "label": "Audio Receiver 1",
        "description": "Mock audio receiver",
        "format": "urn:x-nmos:format:audio",
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "device_id": DEVICE_ID,
        "tags": {"site": ["CampusA"], "room": ["ControlRoom1"]},
        "version": "1523456789:123461",
        "subscription": {
            "sender_id": None,
            "active": False
        }
    }
]

# Mock devices
DEVICES = [
    {
        "id": DEVICE_ID,
        "label": "Mock Decoder Device",
        "description": "Virtual decoder device",
        "node_id": NODE_ID,
        "type": "urn:x-nmos:device:generic",
        "tags": {"site": ["CampusA"], "room": ["ControlRoom1"]},
        "version": "1523456789:123456",
        "controls": {
            "href": f"http://localhost:{PORT}/x-nmos/connection/{IS05_VERSION}/"
        }
    }
]

# IS-05 Connection state
CONNECTION_STATE = {
    "receivers": {}
}

# IS-04 Node API Routes
# Root version discovery endpoint (go-nmos expects this)
@app.route('/x-nmos/node/', methods=['GET'])
def node_versions_root():
    """IS-04 Node API versions (root)"""
    return jsonify([IS04_VERSION])

@app.route(f'/x-nmos/node/{IS04_VERSION}/', methods=['GET'])
def node_versions():
    """IS-04 Node API versions (versioned)"""
    return jsonify([IS04_VERSION])

@app.route(f'/x-nmos/node/{IS04_VERSION}/self', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/self/', methods=['GET'])
def node_self():
    """IS-04 Node self resource"""
    return jsonify({
        "id": NODE_ID,
        "label": NODE_LABEL,
        "description": NODE_DESCRIPTION,
        "hostname": f"mock-node-2.local",
        "api": {
            "endpoints": [
                {
                    "host": "localhost",
                    "port": PORT,
                    "protocol": "http"
                }
            ],
            "versions": [IS04_VERSION]
        },
        "caps": {},
        "clocks": [
            {
                "name": "clk0",
                "ref_type": "internal"
            }
        ],
        "tags": {
            "site": ["CampusA"],
            "room": ["ControlRoom1"]
        },
        "version": "1523456789:123455"
    })

@app.route(f'/x-nmos/node/{IS04_VERSION}/devices', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/devices/', methods=['GET'])
def node_devices():
    """IS-04 Node devices"""
    return jsonify(DEVICES)

@app.route(f'/x-nmos/node/{IS04_VERSION}/flows', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/flows/', methods=['GET'])
def node_flows():
    """IS-04 Node flows (decoder node has no flows)"""
    return jsonify([])

@app.route(f'/x-nmos/node/{IS04_VERSION}/senders', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/senders/', methods=['GET'])
def node_senders():
    """IS-04 Node senders (decoder node has no senders)"""
    return jsonify([])

@app.route(f'/x-nmos/node/{IS04_VERSION}/receivers', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/receivers/', methods=['GET'])
def node_receivers():
    """IS-04 Node receivers"""
    return jsonify(RECEIVERS)

# IS-05 Connection API Routes
@app.route('/x-nmos/connection/', methods=['GET'])
@app.route('/x-nmos/connection', methods=['GET'])  # Support both with and without trailing slash
def connection_versions_root():
    """IS-05 Connection API versions (root endpoint)"""
    return jsonify([IS05_VERSION])

@app.route(f'/x-nmos/connection/{IS05_VERSION}/', methods=['GET'])
def connection_versions():
    """IS-05 Connection API versions"""
    return jsonify([IS05_VERSION])

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders/<sender_id>/active', methods=['GET'])
@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders/<sender_id>/active/', methods=['GET'])
def connection_sender_active(sender_id):
    """IS-05 stub: this node has no senders; return unconnected active for BCC"""
    return jsonify({
        "master_enable": False,
        "activation": {"mode": "activate_immediate", "requested_time": None},
        "transport_file": {"data": "", "type": "application/sdp"},
        "receiver_id": None
    })

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers', methods=['GET'])
def connection_receivers():
    """IS-05 Connection API receivers"""
    receivers = []
    for receiver in RECEIVERS:
        receiver_id = receiver['id']
        state = CONNECTION_STATE.get('receivers', {}).get(receiver_id, {})
        receivers.append({
            "id": receiver_id,
            "master_enable": state.get('master_enable', False),
            "activation": state.get('activation', {
                "mode": "activate_immediate",
                "requested_time": None
            }),
            "transport_file": state.get('transport_file', {
                "data": "",
                "type": "application/sdp"
            }),
            "transport_params": state.get('transport_params', []),
            "sender_id": state.get('sender_id')
        })
    return jsonify(receivers)

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers/<receiver_id>/active', methods=['GET'])
@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers/<receiver_id>/active/', methods=['GET'])
def connection_receiver_active(receiver_id):
    """IS-05 Connection API receiver active (read-only) - mock returns same as staged"""
    state = CONNECTION_STATE.get('receivers', {}).get(receiver_id, {})
    return jsonify({
        "master_enable": state.get('master_enable', False),
        "activation": state.get('activation', {"mode": "activate_immediate", "requested_time": None}),
        "transport_file": state.get('transport_file', {"data": "", "type": "application/sdp"}),
        "transport_params": state.get('transport_params', []),
        "sender_id": state.get('sender_id')
    })

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers/<receiver_id>/staged', methods=['GET', 'PATCH'])
@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/receivers/<receiver_id>/staged/', methods=['GET', 'PATCH'])
def connection_receiver_staged(receiver_id):
    """IS-05 Connection API receiver staged endpoint"""
    if request.method == 'GET':
        state = CONNECTION_STATE.get('receivers', {}).get(receiver_id, {})
        return jsonify({
            "master_enable": state.get('master_enable', False),
            "activation": state.get('activation', {
                "mode": "activate_immediate",
                "requested_time": None
            }),
            "transport_file": state.get('transport_file', {
                "data": "",
                "type": "application/sdp"
            }),
            "transport_params": state.get('transport_params', []),
            "sender_id": state.get('sender_id')
        })
    else:  # PATCH
        try:
            print(f"[PATCH] /receivers/{receiver_id}/staged - Headers: {dict(request.headers)}")
            print(f"[PATCH] /receivers/{receiver_id}/staged - Content-Type: {request.content_type}")
            data = request.get_json(force=True) or {}
            print(f"[PATCH] /receivers/{receiver_id}/staged - Received data: {json.dumps(data, indent=2)}")
            if receiver_id not in CONNECTION_STATE.get('receivers', {}):
                CONNECTION_STATE.setdefault('receivers', {})[receiver_id] = {}
            
            if 'master_enable' in data:
                CONNECTION_STATE['receivers'][receiver_id]['master_enable'] = data['master_enable']
            if 'activation' in data:
                CONNECTION_STATE['receivers'][receiver_id]['activation'] = data['activation']
            if 'transport_file' in data:
                CONNECTION_STATE['receivers'][receiver_id]['transport_file'] = data['transport_file']
            if 'transport_params' in data:
                CONNECTION_STATE['receivers'][receiver_id]['transport_params'] = data['transport_params']
            if 'sender_id' in data:
                CONNECTION_STATE['receivers'][receiver_id]['sender_id'] = data['sender_id']
            
            response = {
                "master_enable": CONNECTION_STATE['receivers'][receiver_id].get('master_enable', False),
                "activation": CONNECTION_STATE['receivers'][receiver_id].get('activation', {
                    "mode": "activate_immediate",
                    "requested_time": None
                }),
                "transport_file": CONNECTION_STATE['receivers'][receiver_id].get('transport_file', {
                    "data": "",
                    "type": "application/sdp"
                }),
                "transport_params": CONNECTION_STATE['receivers'][receiver_id].get('transport_params', []),
                "sender_id": CONNECTION_STATE['receivers'][receiver_id].get('sender_id')
            }
            print(f"[PATCH] /receivers/{receiver_id}/staged - Response: {json.dumps(response, indent=2)}")
            return jsonify(response), 200
        except Exception as e:
            import traceback
            print(f"[ERROR] PATCH /receivers/{receiver_id}/staged: {e}")
            print(traceback.format_exc())
            return jsonify({"error": str(e)}), 400

# Health check
@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "node_id": NODE_ID,
        "timestamp": datetime.utcnow().isoformat()
    })

if __name__ == '__main__':
    print(f"Starting Mock NMOS Node 2 on port {PORT}")
    print(f"Node ID: {NODE_ID}")
    print(f"IS-04 API: http://localhost:{PORT}/x-nmos/node/{IS04_VERSION}/")
    print(f"IS-05 API: http://localhost:{PORT}/x-nmos/connection/{IS05_VERSION}/")
    app.run(host='0.0.0.0', port=PORT, debug=True)
