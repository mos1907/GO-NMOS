#!/usr/bin/env python3
"""
Test script for NMOS flow discovery and management
"""

import requests
import json
import sys
import time

BASE_URL = "http://localhost:9090"
MOCK_NODE_1 = "http://localhost:8080"
MOCK_NODE_2 = "http://localhost:8081"
MOCK_REGISTRY = "http://localhost:8082"

def test_discover_node():
    """Test discovering a mock node"""
    print("\n=== Test: Discover Mock Node 1 ===")
    try:
        response = requests.post(
            f"{BASE_URL}/api/nmos/discover",
            json={"base_url": MOCK_NODE_1},
            headers={"Content-Type": "application/json"}
        )
        if response.status_code == 200:
            data = response.json()
            print(f"✓ Node discovered successfully")
            print(f"  Senders: {len(data.get('senders', []))}")
            print(f"  Receivers: {len(data.get('receivers', []))}")
            print(f"  Flows: {len(data.get('flows', []))}")
            return True
        else:
            print(f"✗ Failed: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"✗ Error: {e}")
        return False

def test_discover_registry():
    """Test discovering from mock registry"""
    print("\n=== Test: Discover from Mock Registry ===")
    try:
        response = requests.post(
            f"{BASE_URL}/api/nmos/registry/discover-nodes",
            json={"query_url": MOCK_REGISTRY},
            headers={"Content-Type": "application/json"}
        )
        if response.status_code == 200:
            data = response.json()
            print(f"✓ Registry discovery successful")
            print(f"  Nodes discovered: {len(data.get('nodes', []))}")
            return True
        else:
            print(f"✗ Failed: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"✗ Error: {e}")
        return False

def test_list_flows():
    """Test listing flows"""
    print("\n=== Test: List Flows ===")
    try:
        response = requests.get(
            f"{BASE_URL}/api/flows?limit=10",
            headers={"Content-Type": "application/json"}
        )
        if response.status_code == 200:
            data = response.json()
            flows = data.get('items', [])
            print(f"✓ Flows listed successfully")
            print(f"  Total flows: {len(flows)}")
            for flow in flows[:3]:  # Show first 3
                print(f"  - {flow.get('display_name', 'N/A')} ({flow.get('multicast_ip', 'N/A')}:{flow.get('port', 'N/A')})")
            return True
        else:
            print(f"✗ Failed: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"✗ Error: {e}")
        return False

def main():
    print("=" * 60)
    print("NMOS Flow Test Suite")
    print("=" * 60)
    
    # Wait for services to be ready
    print("\nWaiting for services to be ready...")
    time.sleep(2)
    
    results = []
    
    # Run tests
    results.append(("Discover Node", test_discover_node()))
    results.append(("Discover Registry", test_discover_registry()))
    results.append(("List Flows", test_list_flows()))
    
    # Summary
    print("\n" + "=" * 60)
    print("Test Summary")
    print("=" * 60)
    passed = sum(1 for _, result in results if result)
    total = len(results)
    for name, result in results:
        status = "✓ PASS" if result else "✗ FAIL"
        print(f"{status}: {name}")
    print(f"\nTotal: {passed}/{total} tests passed")
    
    return 0 if passed == total else 1

if __name__ == "__main__":
    sys.exit(main())
