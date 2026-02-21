# NMOS Interoperability Matrix

This document tracks tested combinations of GO-NMOS with various NMOS devices, registries, and vendor implementations.

## Test Status Legend

- ✅ **Pass** - Fully tested and working
- ⚠️ **Warning** - Works with known limitations or warnings
- ❌ **Fail** - Does not work or has critical issues
- ⏭️ **Skip** - Not applicable or not tested
- 🔄 **In Progress** - Currently being tested

## Test Categories

### IS-04 (Discovery & Registration)

| Component | Test | Status | Notes |
|-----------|------|--------|-------|
| Node API Discovery | Version enumeration | ✅ | Standard `/x-nmos/node/` endpoint |
| Resource Enumeration | Devices, Flows, Senders, Receivers | ✅ | All resource types supported |
| Query API Discovery | Registry version discovery | ✅ | Standard `/x-nmos/query/` endpoint |
| Query Resources | Nodes, Devices, Flows queries | ✅ | Full Query API support |

### IS-05 (Device Connection Management)

| Component | Test | Status | Notes |
|-----------|------|--------|-------|
| Connection API Discovery | Version enumeration | ✅ | Standard `/x-nmos/connection/` endpoint |
| Receivers Endpoint | List receivers | ✅ | `/single/receivers` endpoint |
| Receiver Staging | Staged state access | ✅ | Read-only test (no PATCH) |
| Receiver Activation | PATCH operations | ⚠️ | Requires device-specific testing |

### IS-08 (Audio Channel Mapping)

| Component | Test | Status | Notes |
|-----------|------|--------|-------|
| Channel Mapping API | Version discovery | ✅ | Optional API, detected when available |
| IO Endpoint | Input/Output enumeration | ⏭️ | Requires device-specific testing |
| Active Map | Active mapping retrieval | ⏭️ | Requires device-specific testing |

### IS-07 (Events & Tally)

| Component | Test | Status | Notes |
|-----------|------|--------|-------|
| Events API Discovery | Version enumeration | ✅ | Optional API, detected when available |
| Event Sources | Source enumeration | ⏭️ | Requires device-specific testing |
| Event Subscription | WebSocket/SSE subscriptions | ⏭️ | Requires device-specific testing |

### IS-09 (System Parameters)

| Component | Test | Status | Notes |
|-----------|------|--------|-------|
| System Parameters | PTP domain, IS-04/05 versions | ✅ | Validated via System Parameters Validation |
| PTP Domain Mismatch | Detection and alerting | ✅ | Alerting system detects mismatches |

## Tested Devices & Registries

### Reference Implementations

| Vendor/Product | Type | IS-04 | IS-05 | IS-08 | IS-07 | IS-09 | Status | Notes |
|----------------|------|-------|-------|-------|-------|-------|--------|-------|
| AMWA NMOS Test Suite | Registry | ✅ | ⏭️ | ⏭️ | ⏭️ | ⏭️ | ✅ | Reference registry for testing |
| AMWA NMOS Test Suite | Node | ✅ | ✅ | ⏭️ | ⏭️ | ⏭️ | ✅ | Reference node implementation |

### Vendor Implementations

*Note: Add tested vendor devices here as they are validated*

| Vendor/Product | Type | IS-04 | IS-05 | IS-08 | IS-07 | IS-09 | Status | Notes |
|----------------|------|-------|-------|-------|-------|-------|--------|-------|
| *To be populated* | - | - | - | - | - | - | - | - |

## Common Vendor Profiles

### Broadcast Equipment Vendors

#### Grass Valley
- **Profile**: Typically implements IS-04 v1.3, IS-05 v1.1+
- **Known Issues**: None reported
- **Test Status**: ⏭️ Not yet tested

#### Sony
- **Profile**: Typically implements IS-04 v1.3, IS-05 v1.1+
- **Known Issues**: None reported
- **Test Status**: ⏭️ Not yet tested

#### Blackmagic Design
- **Profile**: Typically implements IS-04 v1.3, IS-05 v1.1+
- **Known Issues**: None reported
- **Test Status**: ⏭️ Not yet tested

#### AJA
- **Profile**: Typically implements IS-04 v1.3, IS-05 v1.1+
- **Known Issues**: None reported
- **Test Status**: ⏭️ Not yet tested

#### NewTek (now Vizrt)
- **Profile**: Typically implements IS-04 v1.3, IS-05 v1.1+
- **Known Issues**: None reported
- **Test Status**: ⏭️ Not yet tested

## Test Results by Version

### IS-04 Versions

| Version | Tested | Status | Notes |
|---------|--------|--------|-------|
| v1.0 | ⏭️ | ⏭️ | Legacy, not commonly used |
| v1.1 | ⏭️ | ⏭️ | Legacy, not commonly used |
| v1.2 | ⏭️ | ⏭️ | Legacy, not commonly used |
| v1.3 | ✅ | ✅ | Most common version, fully supported |
| v1.4 | ⚠️ | ⚠️ | Supported, less common |

### IS-05 Versions

| Version | Tested | Status | Notes |
|---------|--------|--------|-------|
| v1.0 | ⏭️ | ⏭️ | Legacy, not commonly used |
| v1.1 | ✅ | ✅ | Most common version, fully supported |
| v1.2 | ⚠️ | ⚠️ | Supported, less common |

## Known Limitations

1. **IS-05 PATCH Operations**: 
   - Full PATCH testing requires actual device connections
   - Some devices may have device-specific requirements

2. **IS-08 Audio Channel Mapping**:
   - Requires devices with audio capabilities
   - Testing limited to API discovery without physical audio hardware

3. **IS-07 Events**:
   - WebSocket/SSE subscription testing requires long-running connections
   - Event source enumeration depends on device capabilities

4. **Multi-Registry**:
   - Multiple registry support tested with internal registries
   - External registry HA scenarios require production testing

## Testing Procedures

### Running Interoperability Tests

1. **Via API:**
```bash
# Test a node
curl -X POST http://localhost:9090/api/interop/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target": {
      "name": "Test Node",
      "type": "node",
      "base_url": "http://node-host:8080",
      "vendor": "Vendor Name"
    }
  }'

# Test a registry
curl -X POST http://localhost:9090/api/interop/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target": {
      "name": "Test Registry",
      "type": "registry",
      "base_url": "http://registry-host:8080",
      "vendor": "Vendor Name"
    }
  }'
```

2. **List Reference Targets:**
```bash
curl http://localhost:9090/api/interop/targets \
  -H "Authorization: Bearer $TOKEN"
```

### Test Coverage

Current test suite covers:
- ✅ IS-04 Node API discovery and resource enumeration
- ✅ IS-04 Query API discovery and resource queries
- ✅ IS-05 Connection API discovery
- ✅ IS-05 Receiver endpoint accessibility
- ✅ IS-08 Audio Channel Mapping API discovery (when available)
- ✅ IS-07 Events API discovery (when available)
- ✅ Registry health endpoint (when available)

Future enhancements:
- ⏭️ IS-05 PATCH operation testing (requires device connection)
- ⏭️ IS-08 IO and mapping operations
- ⏭️ IS-07 event subscription testing
- ⏭️ Conformance testing against AMWA test suite

## Contributing Test Results

To add test results for a new device or registry:

1. Run interoperability tests via API
2. Document results in this matrix
3. Include:
   - Vendor and product name
   - NMOS versions supported
   - Any known issues or limitations
   - Test date and environment details

## References

- [AMWA NMOS Specifications](https://specs.amwa.tv/nmos/)
- [AMWA NMOS Test Suite](https://github.com/AMWA-TV/nmos-testing)
- [NMOS Implementation Guide](https://github.com/AMWA-TV/nmos)

## Last Updated

- **Date**: 2026-02-19
- **GO-NMOS Version**: v0.2.0
- **Test Suite Version**: Initial release
