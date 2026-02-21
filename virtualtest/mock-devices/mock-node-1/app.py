#!/usr/bin/env python3
"""
Mock NMOS Node 1 - Encoder/Transmitter
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
NODE_ID = os.getenv('NODE_ID', '550e8400-e29b-41d4-a716-446655440001')
NODE_LABEL = os.getenv('NODE_LABEL', 'Mock Encoder Node 1')
NODE_DESCRIPTION = os.getenv('NODE_DESCRIPTION', 'Virtual NMOS encoder for testing')
PORT = int(os.getenv('PORT', 8080))
IS04_VERSION = os.getenv('IS04_VERSION', 'v1.3')
IS05_VERSION = os.getenv('IS05_VERSION', 'v1.0')

# Mock data
DEVICE_ID = str(uuid.uuid4())
FLOW_ID_1 = str(uuid.uuid4())
FLOW_ID_2 = str(uuid.uuid4())
SENDER_ID_1 = str(uuid.uuid4())
SENDER_ID_2 = str(uuid.uuid4())
RECEIVER_ID_1 = str(uuid.uuid4())

# Mock flows
FLOWS = [
    {
        "id": FLOW_ID_1,
        "source_id": SENDER_ID_1,
        "device_id": DEVICE_ID,
        "parents": [],
        "format": "urn:x-nmos:format:video",
        "media_type": "video/raw",
        "label": "Video Flow 1",
        "description": "Mock video flow",
        "tags": {"site": ["CampusA"], "room": ["Studio1"]},
        "version": "1523456789:123456",
        "bit_rate": 1000000000,
        "frame_width": 1920,
        "frame_height": 1080,
        "interlace_mode": "progressive",
        "colorspace": "BT709",
        "transfer_characteristic": "SDR",
        "components": [
            {
                "name": "Y",
                "width": 1920,
                "height": 1080,
                "bit_depth": 10
            }
        ],
        "grain_rate": {
            "numerator": 25,
            "denominator": 1
        }
    },
    {
        "id": FLOW_ID_2,
        "source_id": SENDER_ID_2,
        "device_id": DEVICE_ID,
        "parents": [],
        "format": "urn:x-nmos:format:audio",
        "media_type": "audio/L24",
        "label": "Audio Flow 1",
        "description": "Mock audio flow",
        "tags": {"site": ["CampusA"], "room": ["Studio1"]},
        "version": "1523456789:123457",
        "sample_rate": {
            "numerator": 48000,
            "denominator": 1
        },
        "bit_depth": 24,
        "channels": 2
    }
]

# Mock senders
SENDERS = [
    {
        "id": SENDER_ID_1,
        "label": "Video Sender 1",
        "description": "Mock video sender",
        "flow_id": FLOW_ID_1,
        "device_id": DEVICE_ID,
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "manifest_href": f"http://localhost:{PORT}/sdp/video1.sdp",
        "tags": {"site": ["CampusA"], "room": ["Studio1"]},
        "version": "1523456789:123456",
        "subscription": {
            "receiver_id": None,
            "active": False
        }
    },
    {
        "id": SENDER_ID_2,
        "label": "Audio Sender 1",
        "description": "Mock audio sender",
        "flow_id": FLOW_ID_2,
        "device_id": DEVICE_ID,
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "manifest_href": f"http://localhost:{PORT}/sdp/audio1.sdp",
        "tags": {"site": ["CampusA"], "room": ["Studio1"]},
        "version": "1523456789:123457",
        "subscription": {
            "receiver_id": None,
            "active": False
        }
    }
]

# Mock receivers (IS-04 requires device_id for each receiver)
RECEIVERS = [
    {
        "id": RECEIVER_ID_1,
        "label": "Video Receiver 1",
        "description": "Mock video receiver",
        "device_id": DEVICE_ID,
        "format": "urn:x-nmos:format:video",
        "transport": "urn:x-nmos:transport:rtp.mcast",
        "tags": {"site": ["CampusA"], "room": ["Studio1"]},
        "version": "1523456789:123458",
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
        "label": "Mock Encoder Device",
        "description": "Virtual encoder device",
        "node_id": NODE_ID,
        "type": "urn:x-nmos:device:generic",
        "tags": {"site": ["CampusA"], "room": ["Studio1"]},
        "version": "1523456789:123455",
        "controls": {
            "href": f"http://localhost:{PORT}/x-nmos/connection/{IS05_VERSION}/"
        }
    }
]

# IS-05 Connection state
CONNECTION_STATE = {
    "senders": {},
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
        "hostname": f"mock-node-1.local",
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
            "room": ["Studio1"]
        },
        "version": "1523456789:123454"
    })

@app.route(f'/x-nmos/node/{IS04_VERSION}/devices', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/devices/', methods=['GET'])
def node_devices():
    """IS-04 Node devices"""
    return jsonify(DEVICES)

@app.route(f'/x-nmos/node/{IS04_VERSION}/flows', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/flows/', methods=['GET'])
def node_flows():
    """IS-04 Node flows"""
    return jsonify(FLOWS)

@app.route(f'/x-nmos/node/{IS04_VERSION}/senders', methods=['GET'])
@app.route(f'/x-nmos/node/{IS04_VERSION}/senders/', methods=['GET'])
def node_senders():
    """IS-04 Node senders"""
    return jsonify(SENDERS)

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

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders', methods=['GET'])
def connection_senders():
    """IS-05 Connection API senders"""
    senders = []
    for sender in SENDERS:
        sender_id = sender['id']
        state = CONNECTION_STATE.get('senders', {}).get(sender_id, {})
        senders.append({
            "id": sender_id,
            "master_enable": state.get('master_enable', False),
            "activation": state.get('activation', {
                "mode": "activate_immediate",
                "requested_time": None
            }),
            "transport_file": state.get('transport_file', {
                "data": "",
                "type": "application/sdp"
            }),
            "receiver_id": state.get('receiver_id')
        })
    return jsonify(senders)

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders/<sender_id>', methods=['GET'])
def connection_sender(sender_id):
    """IS-05 Connection API single sender"""
    sender = next((s for s in SENDERS if s['id'] == sender_id), None)
    if not sender:
        return jsonify({"error": "Sender not found"}), 404
    
    state = CONNECTION_STATE.get('senders', {}).get(sender_id, {})
    return jsonify({
        "id": sender_id,
        "master_enable": state.get('master_enable', False),
        "activation": state.get('activation', {
            "mode": "activate_immediate",
            "requested_time": None
        }),
        "transport_file": state.get('transport_file', {
            "data": "",
            "type": "application/sdp"
        }),
        "receiver_id": state.get('receiver_id')
    })

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders/<sender_id>/active', methods=['GET'])
@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders/<sender_id>/active/', methods=['GET'])
def connection_sender_active(sender_id):
    """IS-05 Connection API sender active (read-only) - mock returns same as staged"""
    state = CONNECTION_STATE.get('senders', {}).get(sender_id, {})
    return jsonify({
        "master_enable": state.get('master_enable', False),
        "activation": state.get('activation', {"mode": "activate_immediate", "requested_time": None}),
        "transport_file": state.get('transport_file', {"data": "", "type": "application/sdp"}),
        "receiver_id": state.get('receiver_id')
    })

@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders/<sender_id>/staged', methods=['GET', 'PATCH'])
@app.route(f'/x-nmos/connection/{IS05_VERSION}/single/senders/<sender_id>/staged/', methods=['GET', 'PATCH'])
def connection_sender_staged(sender_id):
    """IS-05 Connection API sender staged endpoint"""
    if request.method == 'GET':
        state = CONNECTION_STATE.get('senders', {}).get(sender_id, {})
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
            "receiver_id": state.get('receiver_id')
        })
    else:  # PATCH
        data = request.get_json()
        if sender_id not in CONNECTION_STATE.get('senders', {}):
            CONNECTION_STATE.setdefault('senders', {})[sender_id] = {}
        
        if 'master_enable' in data:
            CONNECTION_STATE['senders'][sender_id]['master_enable'] = data['master_enable']
        if 'activation' in data:
            CONNECTION_STATE['senders'][sender_id]['activation'] = data['activation']
        if 'transport_file' in data:
            CONNECTION_STATE['senders'][sender_id]['transport_file'] = data['transport_file']
        if 'receiver_id' in data:
            CONNECTION_STATE['senders'][sender_id]['receiver_id'] = data['receiver_id']
        
        return jsonify({
            "master_enable": CONNECTION_STATE['senders'][sender_id].get('master_enable', False),
            "activation": CONNECTION_STATE['senders'][sender_id].get('activation', {
                "mode": "activate_immediate",
                "requested_time": None
            }),
            "transport_file": CONNECTION_STATE['senders'][sender_id].get('transport_file', {
                "data": "",
                "type": "application/sdp"
            }),
            "receiver_id": CONNECTION_STATE['senders'][sender_id].get('receiver_id')
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

# Mock SDP endpoint - format compatible with BCC/IS-05 transport_params extraction
# (source_ip from o= and source-filter; multicast_ip from c=; destination_port from m=; rtcp from a=rtcp:)
def _sdp_video():
    # Order and format per IS-05 / BCC: o= source, c= multicast, m= port(s), a=source-filter + rtcp
    return (
        "v=0\r\n"
        "o=- 0 0 IN IP4 192.168.1.100\r\n"
        "s=Mock Video Stream\r\n"
        "t=0 0\r\n"
        "m=video 5004 RTP/AVP 96\r\n"
        "c=IN IP4 239.0.0.1/32\r\n"
        "a=source-filter: incl IN IP4 239.0.0.1 192.168.1.100\r\n"
        "a=rtpmap:96 raw/90000\r\n"
        "a=rtcp:5005\r\n"
        "a=sendrecv\r\n"
        "a=ts-refclk:ptp=IEEE1588-2008:00-1B-63-FF-FE-FF-FF-FF\r\n"
    )

def _sdp_audio():
    return (
        "v=0\r\n"
        "o=- 0 0 IN IP4 192.168.1.100\r\n"
        "s=Mock Audio Stream\r\n"
        "t=0 0\r\n"
        "m=audio 5006 RTP/AVP 96\r\n"
        "c=IN IP4 239.0.0.2/32\r\n"
        "a=source-filter: incl IN IP4 239.0.0.2 192.168.1.100\r\n"
        "a=rtpmap:96 L24/48000/2\r\n"
        "a=rtcp:5007\r\n"
        "a=sendrecv\r\n"
        "a=ts-refclk:ptp=IEEE1588-2008:00-1B-63-FF-FE-FF-FF-FF\r\n"
    )

@app.route('/sdp/<filename>', methods=['GET'])
def sdp_file(filename):
    """Mock SDP file endpoint - parsers (e.g. BCC) extract transport_params from this"""
    if filename == 'video1.sdp':
        return _sdp_video(), 200, {'Content-Type': 'application/sdp'}
    elif filename == 'audio1.sdp':
        return _sdp_audio(), 200, {'Content-Type': 'application/sdp'}
    else:
        return jsonify({"error": "SDP file not found"}), 404

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
    print(f"Starting Mock NMOS Node 1 on port {PORT}")
    print(f"Node ID: {NODE_ID}")
    print(f"IS-04 API: http://localhost:{PORT}/x-nmos/node/{IS04_VERSION}/")
    print(f"IS-05 API: http://localhost:{PORT}/x-nmos/connection/{IS05_VERSION}/")
    app.run(host='0.0.0.0', port=PORT, debug=True)
