# Server Backend - Dashboard Commands & Logs API Specification

> **Version:** 1.1
> **Status:** Draft
> **Created:** 2026-06-24
> **Target:** Production MVP

---

## Table of Contents

1. Overview
2. Current State Analysis
3. Required API Endpoints
4. Database Schema
5. Backend File Structure
6. Handler Specifications
7. Service Layer
8. GraphQL Schema Updates
9. Error Handling
10. Rate Limiting & Security
11. File Changes Summary
12. Implementation Order

---

## 1. Overview

### 1.1 Purpose

This document maps out the server-side requirements to support the Dashboard, Commands, and Logs front-end as specified in DASHBOARD_COMMANDS_LOGS.md.

### 1.2 Frontend Requirements Summary

| Feature | Description | Required Endpoints |
|---------|-------------|-------------------|
| Send Command | Dispatch commands to device | POST /v1/device/:imei/command |
| Command History | Paginated command history | GET /v1/device/:imei/commands |
| Cancel Command | Cancel pending command | DELETE /v1/device/:imei/command/:dispatchId |
| Device Logs | Fetch device logs | GET /v1/device/:imei/logs |
| Metrics Tab | Real-time charts with time ranges | GET /v1/device/:imei/metrics |
| Telemetry History | Historical telemetry data | GET /v1/device/:imei/telemetry |
| Metrics Export | Export metrics data | GET /v1/device/:imei/metrics/export |
| Dashboard Stats | Aggregated device stats | GET /v1/dashboard/stats |

---

## 2. Current State Analysis

### 2.1 Existing Command Endpoints

| Endpoint | Status | Handler |
|----------|--------|---------|
| POST /v1/device/:imei/command | EXISTS | ExecuteHandler.Handle |
| GET /v1/device/:imei/commands/pending | EXISTS | GetPending |
| GET /v1/command/:dispatchId/status | EXISTS | GetStatus |
| POST /v1/command/:dispatchId/retry | EXISTS | Retry |
| DELETE /v1/command/:dispatchId | EXISTS | Cancel |

### 2.2 Implemented Endpoints

| Endpoint | Status | Handler |
|----------|--------|---------|
| GET /v1/device/:imei/commands | EXISTS | command_history_handler.go |
| GET /v1/device/:imei/logs | EXISTS | device_logs_handler.go |
| GET /v1/device/:imei/metrics | EXISTS | device_metrics_handler.go |
| GET /v1/device/:imei/telemetry | EXISTS | device_telemetry_handler.go |
| GET /v1/device/:imei/metrics/export | EXISTS | device_metrics_handler.go ExportMetrics |
| GET /v1/dashboard/stats | EXISTS | dashboard_handler.go |

---

## 3. Required API Endpoints

### 3.1 GET /v1/device/:imei/commands

Get paginated command history for a device.

Query Parameters:
- status: Filter by status (pending, delivered, failed, completed)
- page: Page number (default: 1)
- limit: Items per page (default: 20, max: 100)
- startTime: Start timestamp ms (default: -30d)
- endTime: End timestamp ms (default: now)

Response:
{
  commands: [...],
  pagination: { page, limit, total, totalPages, hasMore }
}

### 3.2 GET /v1/device/:imei/logs

Get event logs for a device.

Query Parameters:
- type: Filter by type (connection, command, telemetry, error, warning)
- startTime: Start timestamp ms (default: -24h)
- endTime: End timestamp ms (default: now)
- limit: Max results (default: 100, max: 500)
- cursor: Pagination cursor

Response:
{
  events: [{ id, type, timestamp, data }],
  pagination: { limit, hasMore, nextCursor }
}

### 3.3 GET /v1/device/:imei/metrics

Get aggregated metrics for chart visualization with time range presets.

Query Parameters:
- range: Time range - "1h", "6h", "24h", "7d" (default: "6h")
- startTime: Start timestamp ms (overrides range)
- endTime: End timestamp ms
- resolution: Data resolution - "1m", "5m", "15m", "1h", "auto" (default: "auto")

Response:
{
  device: { imei, deviceName },
  timeRange: { start, end, range, resolution },
  metrics: {
    riskScore: { current, avg, min, max, unit, chart: [...], threshold: { warning, critical } },
    thermalTemp: { current, avg, min, max, unit, chart: [...], threshold: { warning, critical } },
    bufferLevel: { current, avg, min, max, unit, chart: [...], threshold: { warning, critical } },
    uptime: { current, unit }
  },
  events: [{ timestamp, type, metric, value, threshold }]
}

Time Range Resolution Mapping:
| Range | Resolution | Max Points |
|-------|------------|------------|
| 1h | 1m | 60 |
| 6h | 5m | 72 |
| 24h | 15m | 96 |
| 7d | 1h | 168 |

### 3.4 GET /v1/device/:imei/metrics/export

Export metrics data in various formats.

Query Parameters:
- format: "json" or "csv" (default: "json")
- range: Time range (default: "24h")
- metrics: Comma-separated list (default: all)

Response: File download (JSON or CSV)

### 3.5 GET /v1/device/:imei/telemetry

Get historical raw telemetry frames.

Query Parameters:
- startTime: Start timestamp ms (default: -6h)
- endTime: End timestamp ms (default: now)
- limit: Max results (default: 500, max: 10000)

Response:
{
  frames: [{ timestamp, riskScore, thermalTemp, bufferLevel, uptime }],
  stats: { riskScore: { current, avg, min, max }, ... }
}

### 3.6 GET /v1/dashboard/stats

Get aggregated dashboard statistics.

Response:
{
  devices: { total, online, offline },
  commands: { totalToday, pending, failed },
  activity: { last24h: { commands, registrations, deregistrations } }
}

---

## 4. Database Schema

### 4.1 Logs Table (NEW)

CREATE TABLE device_logs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    device_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    data JSONB,
    CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_device_logs ON device_logs(device_id, timestamp DESC);
CREATE INDEX idx_device_logs_cursor ON device_logs(device_id, timestamp DESC, id);
CREATE INDEX idx_event_type ON device_logs(event_type);

---

## 5. Backend File Structure

```
apps/api/internal/
├── api/
│   ├── server_routes.go          # Route registration
│   └── handlers/
│       ├── command/
│       │   ├── command_history_handler.go  # CommandHistoryHandler
│       │   └── command_history_routes.go  # (included)
│       ├── device/
│       │   ├── device_logs_handler.go     # LogsHandler
│       │   ├── device_logs_routes.go     # (included)
│       │   ├── device_metrics_handler.go   # MetricsHandler
│       │   ├── device_metrics_routes.go   # (included)
│       │   ├── device_telemetry_handler.go # TelemetryHandler
│       │   └── device_telemetry_routes.go # (included)
│       └── dashboard/
│           ├── dashboard_stats_handler.go # DashboardStatsHandler
│           └── dashboard_stats_routes.go   # (included)
├── application/
│   ├── command/
│   │   └── command_service.go     # CommandService
│   └── dashboard/
│       └── dashboard_service.go   # DashboardService
├── domain/
│   ├── command/
│   │   └── command_entity.go      # Command entity
│   └── device/
│       └── device_entity.go       # Device entity
└── infrastructure/
    └── storage/
        └── device_storage.go      # Storage (commands, logs, metrics)
```

---

## 6. Handler Specifications

### 6.1 Metrics Handler

api/handlers/device/device_metrics_handler.go

GetMetrics: GET /v1/device/:imei/metrics
- Parse range param (1h, 6h, 24h, 7d)
- Determine resolution based on range
- Call metricsService.GetDeviceMetrics()
- Return aggregated chart data

ExportMetrics: GET /v1/device/:imei/metrics/export
- Parse format param (json, csv)
- Call metricsService.ExportMetrics()
- Return file download

### 6.2 Logs Handler

api/handlers/device/device_logs_handler.go

GetLogs: GET /v1/device/:imei/logs
- Parse query params (type, limit, cursor, startTime, endTime)
- Default to -24h if no time range
- Call logsService.GetDeviceLogs()
- Return cursor-paginated events

### 6.3 Command History Handler

api/handlers/command/command_history_handler.go

GetHistory: GET /v1/device/:imei/commands
- Parse query params (status, page, limit, startTime, endTime)
- Default to -30d if no time range
- Call commandService.GetCommandHistory()
- Return page-paginated commands

---

## 7. Service Layer

### 7.1 Metrics Service

application/metrics/metrics_service.go

GetDeviceMetrics(query):
- Parse time range (1h, 6h, 24h, 7d)
- Determine resolution (1m, 5m, 15m, 1h)
- Fetch raw frames from repository
- Aggregate into chart data points
- Calculate stats (current, avg, min, max)
- Fetch threshold breach events
- Return structured metrics result

ExportMetrics(query):
- Fetch frames for time range
- Format as JSON or CSV
- Return file data with filename

---

## 8. GraphQL Schema Updates

New Types:
- MetricChartPoint { timestamp, value }
- MetricData { current, avg, min, max, unit, chart, threshold }
- DeviceMetrics { device, timeRange, metrics, events }
- MetricsCollection { riskScore, thermalTemp, bufferLevel, uptime }

New Queries:
- deviceMetrics(imei, range) -> DeviceMetrics
- deviceLogs(imei, type, limit, cursor) -> LogConnection
- dashboardStats -> DashboardStats

---

## 9. Error Handling

Error Response Format:
{ "error": "code", "message": "Human readable", "details": {} }

Error Codes:
- bad_request (400)
- unauthorized (401)
- forbidden (403)
- not_found (404)
- rate_limited (429)
- internal_error (500)

---

## 10. Rate Limiting & Security

Rate Limits (per endpoint):
- GET /v1/device/:imei/commands: 60/min
- GET /v1/device/:imei/logs: 60/min
- GET /v1/device/:imei/metrics: 30/min
- GET /v1/device/:imei/metrics/export: 10/min
- POST /v1/device/:imei/command: 10/min

---

## 11. File Changes Summary

| Category | New | Modified | Total |
|----------|-----|----------|-------|
| Domain Layer | 5 | 1 | 6 |
| Application Layer | 5 | 1 | 6 |
| Handler Layer | 5 | 2 | 7 |
| Infrastructure | 3 | 1 | 4 |
| GraphQL | 2 | 2 | 4 |
| **TOTAL** | **20** | **7** | **27** |

---

## 12. Implementation Order

Phase 1: Database (Day 1)
- Create device_logs table
- Add indexes

Phase 2: Domain Layer (Day 1)
- Create `domain/logs/logs_entity.go`
- Create `domain/logs/logs_repository.go`
- Create `domain/logs/logs_errors.go`
- Create `domain/metrics/metrics_entity.go`
- Create `domain/metrics/metrics_repository.go`
- Update `domain/command/command_repository.go`

Phase 3: Infrastructure (Day 1-2)
- Create `infrastructure/storage/logs_storage.go`
- Create `infrastructure/storage/metrics_storage.go`

Phase 4: Application Layer (Day 2)
- Create `application/logs/logs_service.go`
- Create `application/logs/logs_dto.go`
- Create `application/metrics/metrics_service.go`
- Create `application/metrics/metrics_dto.go`
- Create `application/dashboard/dashboard_service.go`
- Create `application/dashboard/dashboard_dto.go`

Phase 5: Handlers (Day 2-3)
- Create `handlers/device/device_metrics_handler.go`
- Create `handlers/device/device_logs_handler.go`
- Create `handlers/device/device_telemetry_handler.go`
- Create `handlers/dashboard/dashboard_stats_handler.go`
- Create `handlers/command/command_history_handler.go`
- Update router.go

Phase 6: GraphQL (Day 3)
- Add metrics schema
- Add logs schema
- Add resolvers

Phase 7: Testing (Day 3-4)
- Unit tests
- Integration tests
- E2E tests

---

Document Version: 1.1
Status: Ready for Implementation
