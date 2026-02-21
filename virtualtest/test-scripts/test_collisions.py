#!/usr/bin/env python3
"""
Test script for collision detection
"""

import requests
import json
import sys
import time

BASE_URL = "http://localhost:9090"

def test_collision_check():
    """Test collision detection"""
    print("\n=== Test: Collision Detection ===")
    try:
        # First, create some flows with potential collisions
        flows = [
            {
                "display_name": "Test Flow 1",
                "multicast_ip": "239.0.0.1",
                "port": 5004,
                "flow_status": "active"
            },
            {
                "display_name": "Test Flow 2",
                "multicast_ip": "239.0.0.1",  # Same IP
                "port": 5004,  # Same port - COLLISION!
                "flow_status": "active"
            }
        ]
        
        # Create flows
        flow_ids = []
        for flow in flows:
            response = requests.post(
                f"{BASE_URL}/api/flows",
                json=flow,
                headers={"Content-Type": "application/json"}
            )
            if response.status_code == 200:
                data = response.json()
                flow_ids.append(data.get('id'))
                print(f"✓ Created flow: {flow['display_name']}")
        
        # Run collision check
        print("\nRunning collision check...")
        response = requests.post(
            f"{BASE_URL}/api/checker/run",
            headers={"Content-Type": "application/json"}
        )
        
        if response.status_code == 200:
            data = response.json()
            collisions = data.get('result', {}).get('items', [])
            total = data.get('result', {}).get('total_collisions', 0)
            
            print(f"✓ Collision check completed")
            print(f"  Total collisions: {total}")
            if collisions:
                for collision in collisions:
                    print(f"  - {collision.get('multicast_ip')}:{collision.get('port')} "
                          f"({collision.get('count')} flows)")
            
            # Cleanup
            for flow_id in flow_ids:
                requests.delete(f"{BASE_URL}/api/flows/{flow_id}")
            
            return total > 0  # Test passes if collisions detected
        else:
            print(f"✗ Failed: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"✗ Error: {e}")
        return False

def test_alternative_suggestions():
    """Test alternative IP/Port suggestions"""
    print("\n=== Test: Alternative Suggestions ===")
    try:
        response = requests.get(
            f"{BASE_URL}/api/checker/check?multicast_ip=239.0.0.1&port=5004",
            headers={"Content-Type": "application/json"}
        )
        if response.status_code == 200:
            data = response.json()
            alternatives = data.get('alternatives', [])
            print(f"✓ Alternative suggestions retrieved")
            print(f"  Alternatives found: {len(alternatives)}")
            for alt in alternatives[:3]:  # Show first 3
                print(f"  - {alt.get('multicast_ip')}:{alt.get('port')} "
                      f"({alt.get('reason', 'N/A')})")
            return True
        else:
            print(f"✗ Failed: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"✗ Error: {e}")
        return False

def main():
    print("=" * 60)
    print("NMOS Collision Detection Test Suite")
    print("=" * 60)
    
    results = []
    
    # Run tests
    results.append(("Collision Detection", test_collision_check()))
    results.append(("Alternative Suggestions", test_alternative_suggestions()))
    
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
