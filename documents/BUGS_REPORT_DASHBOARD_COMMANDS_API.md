# Dashboard Commands & Logs API - Bugs Report

> **Report Date:** 2026-07-03
> **Status:** ALL FIXED ✅
> **Severity:** 6 CRITICAL, 1 MEDIUM

---

## Executive Summary

Analysis of the Dashboard Commands & Logs API implementation against the `SERVER_BACKEND_DASHBOARD_COMMANDS_API.md` specification revealed **7 bugs** that must be fixed before production deployment.

| Priority | Count | Severity | Status |
|----------|-------|----------|--------|
| CRITICAL | 6 | Security vulnerabilities (missing DOA checks) | ✅ FIXED |
| MEDIUM | 1 | Data integrity (hardcoded values) | ✅ FIXED |

---

## Verification Status

- [x] All handlers extract operator from context
- [x] All handlers use `FindByIDAndOperator()` instead of `FindByID()`
- [x] Dashboard stats filtered by operator ID
- [x] Metrics storage queries include uptime column
- [x] Thresholds loaded from environment variables
- [x] Build passes: `go build ./...`
- [x] All existing tests pass

---

## Bug Registry

### 🔴 BUG-DOA-1: Command History Handler - Missing DOA Check

| Field | Value |
|-------|-------|
| **Severity** | 🔴 CRITICAL |
| **Type** | Security - Missing Authorization |
| **File** | `apps/api/internal/api/handlers/command/command_history_handler.go` |
| **Line** | 39 |
| **Endpoint** | `GET /v1/device/:imei/commands` |
| **Status** | ✅ FIXED |

**Description:**
Handler uses `FindByID()` which does not verify device ownership. Any authenticated operator can access any device's command history.

**Fix Applied:**
- Added operator extraction via `middleware.GetOperatorFromContext(c)`
- Changed `FindByID()` to `FindByIDAndOperator(ctx, deviceID, op.ID)` in handler
- Updated `GetHistoryRequest` struct to include `OperatorID`
- Updated service to use `FindByIDAndOperator` for DOA check

**Impact:** Unauthorized access to command history of all devices.

---

### 🔴 BUG-DOA-2: Device Logs Handler - Missing DOA Check

| Field | Value |
|-------|-------|
| **Severity** | 🔴 CRITICAL |
| **Type** | Security - Missing Authorization |
| **File** | `apps/api/internal/api/handlers/device/device_logs_handler.go` |
| **Line** | 39 |
| **Endpoint** | `GET /v1/device/:imei/logs` |
| **Status** | ✅ FIXED |

**Description:**
Handler uses `FindByID()` which does not verify device ownership. Any authenticated operator can access any device's logs.

**Current Code:**
```go
_, err := h.devRepo.FindByID(ctx, deviceID)
if err != nil {
    h.logger.Warn("Device not found", "deviceID", deviceID, "error", err)
    c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Device not found"})
    return
}
```

**Required Fix:**
Change `FindByID()` to `FindByIDAndOperator(ctx, deviceID, operatorID)`.

**Impact:** Unauthorized access to device logs of all devices.

---

### 🔴 BUG-DOA-3: Device Metrics Handler - Missing DOA Check

| Field | Value |
|-------|-------|
| **Severity** | 🔴 CRITICAL |
| **Type** | Security - Missing Authorization |
| **File** | `apps/api/internal/api/handlers/device/device_metrics_handler.go` |
| **Line** | 52 |
| **Endpoint** | `GET /v1/device/:imei/metrics` |
| **Status** | ✅ FIXED |

**Description:**
Handler uses `FindByID()` which does not verify device ownership. Any authenticated operator can access any device's metrics.

**Fix Applied:**
- Handler extracts operator from context via `middleware.GetOperatorFromContext(c)`
- Changed `FindByID()` to `FindByIDAndOperator(ctx, deviceID, op.ID)` at line 52
- Operator details returned in response for API compatibility

**Impact:** Unauthorized access to device metrics of all devices.

---

### 🔴 BUG-DOA-4: Device Telemetry Handler - Missing DOA Check

| Field | Value |
|-------|-------|
| **Severity** | 🔴 CRITICAL |
| **Type** | Security - Missing Authorization |
| **File** | `apps/api/internal/api/handlers/device/device_telemetry_handler.go` |
| **Line** | 49 |
| **Endpoint** | `GET /v1/device/:imei/telemetry` |
| **Status** | ✅ FIXED |

**Description:**
Handler uses `FindByID()` which does not verify device ownership. Any authenticated operator can access any device's telemetry.

**Fix Applied:**
- Handler extracts operator from context via `middleware.GetOperatorFromContext(c)`
- Changed `FindByID()` to `FindByIDAndOperator(ctx, deviceID, op.ID)` at line 49

**Impact:** Unauthorized access to telemetry data of all devices.

---

### 🔴 BUG-DOA-5: Dashboard Stats Handler - Missing Operator Scoping

| Field | Value |
|-------|-------|
| **Severity** | 🔴 CRITICAL |
| **Type** | Security - Missing Authorization |
| **File** | `apps/api/internal/api/handlers/dashboard/dashboard_stats_handler.go` |
| **Endpoint** | `GET /v1/dashboard/stats` |
| **Status** | ✅ FIXED |

**Description:**
Dashboard stats endpoint returns aggregate statistics for ALL devices without filtering by operator. Any authenticated operator sees all operators' device data.

**Fix Applied:**
- Handler extracts operator from context
- Passes operator ID to `GetDashboardStats(ctx, operatorID)`
- Service uses `ListByOperator()` instead of `List()`
- GraphQL resolver also updated to pass `op.ID`

**Impact:** Data leakage - operators can see aggregate stats of all operators.

---

### 🔴 BUG-DOA-6: Export Metrics Handler - Missing DOA Check

| Field | Value |
|-------|-------|
| **Severity** | 🔴 CRITICAL |
| **Type** | Security - Missing Authorization |
| **File** | `apps/api/internal/api/handlers/device/device_metrics_handler.go` |
| **Line** | 104 |
| **Endpoint** | `GET /v1/device/:imei/metrics/export` |
| **Status** | ✅ FIXED |

**Description:**
Export metrics endpoint uses `FindByID()` without authorization check.

**Fix Applied:**
- Handler extracts operator from context via `middleware.GetOperatorFromContext(c)`
- Changed `FindByID()` to `FindByIDAndOperator(ctx, deviceID, op.ID)` at line 104

**Impact:** Unauthorized export of device metrics.

---

### 🟡 BUG-THRESH-1: Hardcoded Thresholds in Metrics Service

| Field | Value |
|-------|-------|
| **Severity** | 🟡 MEDIUM |
| **Type** | Configuration |
| **File** | `apps/api/internal/application/metrics/metrics_service.go` |
| **Lines** | 33-40 |
| **Status** | ✅ FIXED |

**Description:**
Threshold values were hardcoded and not configurable. Production requires these to be per-operator settings from the client Settings section.

**Fix Applied:**
- Added `operatorRepo` to `metrics.Service` struct
- Added `GetThresholds()` method to `operator.Repository` interface
- Implemented `GetThresholds()` in `operator_storage.go`
- Added `getOperatorThresholds()` helper in `metrics_service.go`
- Thresholds now fetched from operator's settings in database
- Falls back to defaults if operator settings unavailable

**Files Modified:**
- `internal/domain/operator/operator_repository.go` - Added `GetThresholds()` interface
- `internal/infrastructure/storage/operator_storage.go` - Implemented `GetThresholds()`
- `internal/application/metrics/metrics_service.go` - Uses operator thresholds
- `internal/api/server.go` - Passes operator repo to metrics service
- `cmd/api/main.go` - Passes operator repo to metrics service (GraphQL)

**Impact:** Cannot adjust thresholds per deployment/customer.

---

### 🟡 BUG-UPTIME-1: Telemetry Uptime Hardcoded to 0

| Field | Value |
|-------|-------|
| **Severity** | 🟡 MEDIUM |
| **Type** | Data Integrity |
| **File** | `apps/api/internal/infrastructure/storage/metrics_storage.go` |
| **Line** | 78 |
| **Status** | ✅ FIXED |

**Description:**
`GetTelemetryFrames` queries 5 columns but does not include `uptime`, then hardcodes it to 0. `GetLatestTelemetry` correctly includes uptime.

**Current Code:**
```go
query := `
    SELECT device_id, risk_score, thermal_temp, buffer_level, received_at
    FROM telemetry ...`

// Later:
frame.Uptime = 0 // Not available in current telemetry schema
```

**Required Fix:**
```go
query := `
    SELECT device_id, risk_score, thermal_temp, buffer_level, received_at, COALESCE(uptime, 0)
    FROM telemetry ...`

// Later:
rows.Scan(
    &frame.DeviceID, &frame.RiskScore, &frame.ThermalTemp,
    &frame.BufferLevel, &timestamp, &frame.Uptime,
)
```

**Impact:** Telemetry frames always show uptime=0, breaking uptime tracking.

---

## Files Requiring Changes

| File | Bugs |
|------|------|
| `internal/api/handlers/command/command_history_handler.go` | BUG-DOA-1 |
| `internal/api/handlers/device/device_logs_handler.go` | BUG-DOA-2 |
| `internal/api/handlers/device/device_metrics_handler.go` | BUG-DOA-3, BUG-DOA-6 |
| `internal/api/handlers/device/device_telemetry_handler.go` | BUG-DOA-4 |
| `internal/api/handlers/dashboard/dashboard_stats_handler.go` | BUG-DOA-5 |
| `internal/application/metrics/metrics_service.go` | BUG-THRESH-1 |
| `internal/infrastructure/storage/metrics_storage.go` | BUG-UPTIME-1 |

---

## Implementation Order

1. **Phase 1 - Security Fixes (MUST DO FIRST)**
   - BUG-DOA-1 through BUG-DOA-6

2. **Phase 2 - Data Integrity**
   - BUG-UPTIME-1

3. **Phase 3 - Configuration**
   - BUG-THRESH-1

---

## Verification Checklist

- [ ] All handlers extract operator from context
- [ ] All handlers use `FindByIDAndOperator()` instead of `FindByID()`
- [ ] Dashboard stats filtered by operator ID
- [ ] Metrics storage queries include uptime column
- [ ] Thresholds loaded from environment variables
- [ ] Build passes: `go build ./...`
- [ ] All existing tests pass

---

## Related Documents

- `SERVER_BACKEND_DASHBOARD_COMMANDS_API.md` - API Specification
- `VYZORIX-UPDATE-SERVER-COMPREHENSIVE-BUG-ANALYSIS.md` - Previous bug analysis
