#!/usr/bin/env python3
"""
Mock IS-04 Registry (Query API)
Tüm mock node'ları keşfeder ve Query API üzerinden sunar
"""

import os
import json
import requests
from datetime import datetime
from flask import Flask, jsonify, request
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

# Configuration
PORT = int(os.getenv('PORT', 8082))
QUERY_VERSION = os.getenv('QUERY_VERSION', 'v1.3')

# Mock node endpoints (from docker-compose)
# Internal Docker URLs for registry to discover resources
MOCK_NODES_INTERNAL = [
    {'url': 'http://mock-node-1:8080', 'name': 'Mock Node 1', 'external_url': 'http://localhost:8080'},
    {'url': 'http://mock-node-2:8081', 'name': 'Mock Node 2', 'external_url': 'http://localhost:8081'},
    {'url': 'http://mock-studio1:8090', 'name': 'Campus-A Studio-1', 'external_url': 'http://localhost:8090'},
    {'url': 'http://mock-tx:8091', 'name': 'Campus-A TX', 'external_url': 'http://localhost:8091'},
    {'url': 'http://mock-multiviewer:8092', 'name': 'Campus-A Multiviewer', 'external_url': 'http://localhost:8092'},
]
MOCK_NODES = MOCK_NODES_INTERNAL  # Keep for backward compatibility

# Cache for discovered resources
RESOURCE_CACHE = {
    'nodes': [],
    'devices': [],
    'flows': [],
    'senders': [],
    'receivers': []
}

def discover_node_resources(node_url, external_url=None):
    """Discover all resources from a mock node"""
    try:
        base_url = node_url.rstrip('/')
        # Use external_url for base_url if provided (for frontend access)
        display_base_url = external_url.rstrip('/') if external_url else base_url
        version = QUERY_VERSION
        
        # Get node self
        node_resp = requests.get(f'{base_url}/x-nmos/node/{version}/self', timeout=2)
        if node_resp.status_code == 200:
            node_data = node_resp.json()
            node_data['base_url'] = display_base_url  # Use external URL for frontend access
            # Add href field for IS-04 Query API compatibility (use external URL)
            node_data['href'] = f'{display_base_url}/x-nmos/node/{version}'
            RESOURCE_CACHE['nodes'].append(node_data)
        
        # Get devices
        devices_resp = requests.get(f'{base_url}/x-nmos/node/{version}/devices', timeout=2)
        if devices_resp.status_code == 200:
            devices = devices_resp.json()
            for device in devices:
                device['base_url'] = display_base_url  # Use external URL
                RESOURCE_CACHE['devices'].append(device)
        
        # Get flows
        flows_resp = requests.get(f'{base_url}/x-nmos/node/{version}/flows', timeout=2)
        if flows_resp.status_code == 200:
            flows = flows_resp.json()
            for flow in flows:
                flow['base_url'] = display_base_url  # Use external URL
                RESOURCE_CACHE['flows'].append(flow)
        
        # Get senders
        senders_resp = requests.get(f'{base_url}/x-nmos/node/{version}/senders', timeout=2)
        if senders_resp.status_code == 200:
            senders = senders_resp.json()
            for sender in senders:
                sender['base_url'] = display_base_url  # Use external URL
                RESOURCE_CACHE['senders'].append(sender)
        
        # Get receivers
        receivers_resp = requests.get(f'{base_url}/x-nmos/node/{version}/receivers', timeout=2)
        if receivers_resp.status_code == 200:
            receivers = receivers_resp.json()
            for receiver in receivers:
                receiver['base_url'] = display_base_url  # Use external URL
                RESOURCE_CACHE['receivers'].append(receiver)
        
        return True
    except Exception as e:
        print(f"Error discovering {node_url}: {e}")
        return False

def refresh_cache():
    """Refresh resource cache from all mock nodes"""
    RESOURCE_CACHE['nodes'] = []
    RESOURCE_CACHE['devices'] = []
    RESOURCE_CACHE['flows'] = []
    RESOURCE_CACHE['senders'] = []
    RESOURCE_CACHE['receivers'] = []
    
    for node in MOCK_NODES_INTERNAL:
        discover_node_resources(node['url'], node.get('external_url'))
    
    print(f"Cache refreshed: {len(RESOURCE_CACHE['nodes'])} nodes, "
          f"{len(RESOURCE_CACHE['devices'])} devices, "
          f"{len(RESOURCE_CACHE['flows'])} flows, "
          f"{len(RESOURCE_CACHE['senders'])} senders, "
          f"{len(RESOURCE_CACHE['receivers'])} receivers")

# IS-04 Query API Routes
@app.route('/x-nmos/query/', methods=['GET'])
@app.route('/x-nmos/query', methods=['GET'])  # Support both with and without trailing slash
def query_versions_root():
    """IS-04 Query API versions (root endpoint)"""
    return jsonify([QUERY_VERSION])

@app.route(f'/x-nmos/query/{QUERY_VERSION}/', methods=['GET'])
def query_versions():
    """IS-04 Query API versions"""
    return jsonify([QUERY_VERSION])

@app.route(f'/x-nmos/query/{QUERY_VERSION}/nodes', methods=['GET'])
@app.route(f'/x-nmos/query/{QUERY_VERSION}/nodes/', methods=['GET'])
def query_nodes():
    """IS-04 Query API nodes (BCC uses trailing slash)"""
    refresh_cache()
    return jsonify(RESOURCE_CACHE['nodes'])

@app.route(f'/x-nmos/query/{QUERY_VERSION}/devices', methods=['GET'])
@app.route(f'/x-nmos/query/{QUERY_VERSION}/devices/', methods=['GET'])
def query_devices():
    """IS-04 Query API devices"""
    refresh_cache()
    return jsonify(RESOURCE_CACHE['devices'])

@app.route(f'/x-nmos/query/{QUERY_VERSION}/flows', methods=['GET'])
@app.route(f'/x-nmos/query/{QUERY_VERSION}/flows/', methods=['GET'])
def query_flows():
    """IS-04 Query API flows"""
    refresh_cache()
    return jsonify(RESOURCE_CACHE['flows'])

@app.route(f'/x-nmos/query/{QUERY_VERSION}/senders', methods=['GET'])
@app.route(f'/x-nmos/query/{QUERY_VERSION}/senders/', methods=['GET'])
def query_senders():
    """IS-04 Query API senders"""
    refresh_cache()
    return jsonify(RESOURCE_CACHE['senders'])

@app.route(f'/x-nmos/query/{QUERY_VERSION}/receivers', methods=['GET'])
@app.route(f'/x-nmos/query/{QUERY_VERSION}/receivers/', methods=['GET'])
def query_receivers():
    """IS-04 Query API receivers"""
    refresh_cache()
    return jsonify(RESOURCE_CACHE['receivers'])

# Query API with filters
@app.route(f'/x-nmos/query/{QUERY_VERSION}/nodes/<node_id>', methods=['GET'])
def query_node(node_id):
    """IS-04 Query API single node"""
    refresh_cache()
    node = next((n for n in RESOURCE_CACHE['nodes'] if n.get('id') == node_id), None)
    if node:
        return jsonify(node)
    return jsonify({"error": "Node not found"}), 404

@app.route(f'/x-nmos/query/{QUERY_VERSION}/devices/<device_id>', methods=['GET'])
def query_device(device_id):
    """IS-04 Query API single device"""
    refresh_cache()
    device = next((d for d in RESOURCE_CACHE['devices'] if d.get('id') == device_id), None)
    if device:
        return jsonify(device)
    return jsonify({"error": "Device not found"}), 404

@app.route(f'/x-nmos/query/{QUERY_VERSION}/flows/<flow_id>', methods=['GET'])
def query_flow(flow_id):
    """IS-04 Query API single flow"""
    refresh_cache()
    flow = next((f for f in RESOURCE_CACHE['flows'] if f.get('id') == flow_id), None)
    if flow:
        return jsonify(flow)
    return jsonify({"error": "Flow not found"}), 404

@app.route(f'/x-nmos/query/{QUERY_VERSION}/senders/<sender_id>', methods=['GET'])
def query_sender(sender_id):
    """IS-04 Query API single sender"""
    refresh_cache()
    sender = next((s for s in RESOURCE_CACHE['senders'] if s.get('id') == sender_id), None)
    if sender:
        return jsonify(sender)
    return jsonify({"error": "Sender not found"}), 404

@app.route(f'/x-nmos/query/{QUERY_VERSION}/receivers/<receiver_id>', methods=['GET'])
def query_receiver(receiver_id):
    """IS-04 Query API single receiver"""
    refresh_cache()
    receiver = next((r for r in RESOURCE_CACHE['receivers'] if r.get('id') == receiver_id), None)
    if receiver:
        return jsonify(receiver)
    return jsonify({"error": "Receiver not found"}), 404

# Health check
@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    refresh_cache()
    return jsonify({
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
        "resources": {
            "nodes": len(RESOURCE_CACHE['nodes']),
            "devices": len(RESOURCE_CACHE['devices']),
            "flows": len(RESOURCE_CACHE['flows']),
            "senders": len(RESOURCE_CACHE['senders']),
            "receivers": len(RESOURCE_CACHE['receivers'])
        }
    })

if __name__ == '__main__':
    print(f"Starting Mock IS-04 Registry on port {PORT}")
    print(f"Query API: http://localhost:{PORT}/x-nmos/query/{QUERY_VERSION}/")
    
    # Initial cache refresh
    refresh_cache()
    
    app.run(host='0.0.0.0', port=PORT, debug=True)
