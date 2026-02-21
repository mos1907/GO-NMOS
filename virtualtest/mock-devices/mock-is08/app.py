#!/usr/bin/env python3
"""
Mock IS-08 Audio Channel Mapping Service
"""

import os
import json
import uuid
from datetime import datetime
from flask import Flask, jsonify, request
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

PORT = int(os.getenv('PORT', 8084))
CHANNELMAPPING_VERSION = os.getenv('CHANNELMAPPING_VERSION', 'v1.0')

# Mock channel mapping state (IS-08 /io view: inputs & outputs with channels)
MOCK_INPUT_ID = "input1"
MOCK_OUTPUT_ID = "output1"
IO_RESPONSE = {
    "inputs": {
        MOCK_INPUT_ID: {
            "parent": {"id": None, "type": None},
            "channels": [{"label": "L"}, {"label": "R"}, {"label": "C"}, {"label": "LFE"}, {"label": "LS"}, {"label": "RS"}],
            "caps": {"reordering": True, "block_size": 1},
            "properties": {"name": "Mock 5.1 Input", "description": "Mock stereo/5.1 input"},
        }
    },
    "outputs": {
        MOCK_OUTPUT_ID: {
            "source_id": str(uuid.uuid4()),
            "channels": [{"label": "L"}, {"label": "R"}],
            "caps": {"routable_inputs": [MOCK_INPUT_ID]},
            "properties": {"name": "Mock Stereo Output", "description": "Mock stereo output"},
        }
    },
}

# Active map: output_id -> list of { "input": id, "channel_index": n } or { "mute": True }
ACTIVE_MAP = {
    MOCK_OUTPUT_ID: [
        {"input": MOCK_INPUT_ID, "channel_index": 0},
        {"input": MOCK_INPUT_ID, "channel_index": 1},
    ]
}

# IS-08 API Routes
@app.route(f'/x-nmos/channelmapping/{CHANNELMAPPING_VERSION}/', methods=['GET'])
def channelmapping_versions():
    """IS-08 Channel Mapping API versions"""
    return jsonify([CHANNELMAPPING_VERSION])

@app.route(f'/x-nmos/channelmapping/{CHANNELMAPPING_VERSION}/inputs', methods=['GET'])
def get_inputs():
    """IS-08 Inputs API"""
    return jsonify(list(IO_RESPONSE["inputs"].keys()))

@app.route(f'/x-nmos/channelmapping/{CHANNELMAPPING_VERSION}/outputs', methods=['GET'])
def get_outputs():
    """IS-08 Outputs API"""
    return jsonify(list(IO_RESPONSE["outputs"].keys()))

@app.route(f'/x-nmos/channelmapping/{CHANNELMAPPING_VERSION}/io', methods=['GET'])
def get_io():
    """IS-08 Inputs/Outputs view (single /io endpoint for controller)"""
    return jsonify(IO_RESPONSE)

@app.route(f'/x-nmos/channelmapping/{CHANNELMAPPING_VERSION}/map/active', methods=['GET'])
def get_map_active():
    """IS-08 active channel map"""
    return jsonify(ACTIVE_MAP)

@app.route(f'/x-nmos/channelmapping/{CHANNELMAPPING_VERSION}/map/activations', methods=['POST'])
def post_map_activations():
    """IS-08 create activation (apply channel map). Mock: update ACTIVE_MAP from requested."""
    global ACTIVE_MAP
    data = request.get_json(force=True) or {}
    requested = data.get("requested", {})
    if requested:
        ACTIVE_MAP = requested
    return jsonify({"activation_id": str(uuid.uuid4()), "mode": data.get("mode", "activate_immediate")}), 201

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat()
    })

if __name__ == '__main__':
    print(f"Starting Mock IS-08 Audio Channel Mapping Service on port {PORT}")
    app.run(host='0.0.0.0', port=PORT, debug=True)
