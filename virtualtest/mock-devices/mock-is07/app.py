#!/usr/bin/env python3
"""
Mock IS-07 Event & Tally Service
"""

import os
import json
import uuid
from datetime import datetime
from flask import Flask, jsonify, request
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

PORT = int(os.getenv('PORT', 8083))
EVENTS_VERSION = os.getenv('EVENTS_VERSION', 'v1.0')

# Mock event state
EVENTS = {
    "source_id": str(uuid.uuid4()),
    "events": []
}

# IS-07 API Routes
@app.route(f'/x-nmos/events/{EVENTS_VERSION}/', methods=['GET'])
def events_versions():
    """IS-07 Events API versions"""
    return jsonify([EVENTS_VERSION])

@app.route(f'/x-nmos/events/{EVENTS_VERSION}/events', methods=['GET'])
def get_events():
    """IS-07 Events API"""
    return jsonify(EVENTS['events'])

@app.route(f'/x-nmos/events/{EVENTS_VERSION}/state', methods=['GET'])
def get_state():
    """IS-07 State API"""
    return jsonify({
        "source_id": EVENTS['source_id'],
        "events": EVENTS['events']
    })

@app.route(f'/x-nmos/events/{EVENTS_VERSION}/sources', methods=['GET'])
def get_sources():
    """IS-07 Event sources (list of sources for controllers)"""
    return jsonify([
        {"id": EVENTS["source_id"], "description": "Mock IS-07 event source"}
    ])

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat()
    })

if __name__ == '__main__':
    print(f"Starting Mock IS-07 Event & Tally Service on port {PORT}")
    app.run(host='0.0.0.0', port=PORT, debug=True)
