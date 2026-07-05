# WebSocket/Real-Time Architecture Mismatch Report

**Date:** 2026-07-05
**Document:** REALTIME_WEBSOCKET_ARCHITECTURE.md vs Implementation

---

## 1. Endpoint Path Mismatch

| Document Says | Implementation Uses | Status |
|--------------|---------------------|--------|
| `/v1/device/:id/stream` | `/v1/device/:imei/stream` | ⚠️ MISMATCH - Uses `:imei` not `:id` |

**Fix Needed:** Update doc to use `:imei` for consistency with rest of API.

---

## 2. Missing GraphQL Subscriptions Documentation

### Implemented Subscriptions (from schema/subscription.go):
- `deviceUpdated` - Subscribe to device update events
- `telemetryReceived` - Subscribe to real-time telemetry  
- `commandStatusChanged` - Subscribe to command status changes

### Document Status:
The REALTIME_WEBSOCKET_ARCHITECTURE.md does NOT document these GraphQL subscriptions.

**Fix Needed:** Add section documenting GraphQL subscription types and their usage.

---

## 3. Implementation Status: ✅ EXISTS

| Component | Handler File | Status |
|-----------|--------------|--------|
| WebSocket Stream | `websocket_handler.go` | ✅ IMPLEMENTED |
| Message Router | `stream_message.go` | ✅ IMPLEMENTED |
| Telemetry Handling | `MessageRouter.handleTelemetry()` | ✅ IMPLEMENTED |
| Command Dispatch | `SendToClient()` | ✅ IMPLEMENTED |
| GraphQL Subscriptions | `schema/subscription.go` | ✅ IMPLEMENTED |

---

## 4. Message Types

### Device → Server (Implemented):
| Type | Handler | Status |
|------|---------|--------|
| `TELEMETRY` | `handleTelemetry()` | ✅ |
| `PONG` | `handlePong()` | ✅ |
| `STATUS` | `handleStatus()` | ✅ |

### Document Says vs Implementation:
| Document Says | Implementation | Status |
|---------------|----------------|--------|
| `AUTH` | Not in MessageRouter | ⚠️ Uses HMAC middleware instead |
| `CMD_ACK` | Not in MessageRouter | ⚠️ Handled via REST API |
| `PONG` | `handlePong()` | ✅ |

---

## 5. Fix Plan

1. **Update endpoint path** in REALTIME_WEBSOCKET_ARCHITECTURE.md:
   - Change all `/v1/device/:id/stream` → `/v1/device/:imei/stream`

2. **Add GraphQL subscription documentation**:
   - Document `deviceUpdated(deviceId: ID): Device`
   - Document `telemetryReceived(deviceId: ID): TelemetryEntry`
   - Document `commandStatusChanged(dispatchId: ID): Command`

3. **Clarify AUTH flow**:
   - Document that authentication uses HMAC middleware, not a message type

4. **Clarify CMD_ACK flow**:
   - Document that command acknowledgments go through REST API, not WebSocket
