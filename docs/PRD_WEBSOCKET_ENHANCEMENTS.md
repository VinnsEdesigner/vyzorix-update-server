# PRD: WebSocket & Data Fetching Enhancement System

> **Feature Name:** Advanced WebSocket Communication & Real-Time Data Platform  
> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-15  
> **Target Release:** Production MVP  

---

## 1. Introduction

### Problem Statement

The current WebSocket implementation provides basic real-time communication but lacks critical features for production use:

1. **No Message Persistence**: Messages lost if client disconnected or buffer full
2. **Limited Error Recovery**: No automatic reconnection or message retry
3. **Basic Telemetry**: Simple broadcast with no filtering or history
4. **No Scalability Features**: No compression, rate limiting, or load balancing
5. **Basic UI Integration**: Minimal frontend support for real-time data

### Solution

Implement an **advanced WebSocket communication platform** with:
- Message queuing and persistence
- Automatic reconnection and error recovery
- Enhanced telemetry with filtering and history
- Performance optimizations (compression, rate limiting)
- Rich frontend components for real-time data visualization

---

## 2. Goals

### Primary Goals

- **G1:** Ensure 100% message delivery with queuing and persistence
- **G2:** Achieve 99.9% WebSocket uptime with automatic reconnection
- **G3:** Support 10,000+ concurrent WebSocket connections
- **G4:** Reduce bandwidth usage by 60% with compression
- **G5:** Provide real-time telemetry with filtering and history
- **G6:** Deliver sub-500ms end-to-end latency for commands
- **G7:** Maintain backward compatibility with existing clients

### Secondary Goals

- **G8:** Add WebSocket message encryption for sensitive data
- **G9:** Implement session resumption for seamless reconnects
- **G10:** Add advanced telemetry aggregation and analytics
- **G11:** Support multi-device subscription management

---

## 3. User Stories

### US-001: Message Queue for Offline Devices
**Description:** As a device, I want my commands to be queued when I'm offline so I receive them when I reconnect.

**Acceptance Criteria:**
- [ ] Commands sent to offline devices are stored in queue
- [ ] Queue persists across server restarts
- [ ] Commands delivered in order when device reconnects
- [ ] Queue size limited to prevent memory exhaustion
- [ ] Queue metrics tracked (size, age, delivery time)

---

### US-002: Automatic WebSocket Reconnection
**Description:** As a frontend client, I want automatic reconnection if my WebSocket disconnects so I don't lose real-time updates.

**Acceptance Criteria:**
- [ ] Automatic reconnection with exponential backoff
- [ ] Max 5 reconnection attempts
- [ ] Connection status displayed in UI
- [ ] Queued messages sent after reconnect
- [ ] Reconnection metrics logged

---

### US-003: Message Compression
**Description:** As a mobile user, I want WebSocket messages compressed to reduce data usage.

**Acceptance Criteria:**
- [ ] GZIP compression for large messages (>1KB)
- [ ] Configurable compression threshold
- [ ] Fallback to uncompressed if compression fails
- [ ] Bandwidth savings >40%
- [ ] Compression metrics tracked

---

### US-004: Telemetry Filtering
**Description:** As a dashboard user, I want to filter telemetry by device so I only see relevant data.

**Acceptance Criteria:**
- [ ] Subscribe to specific devices
- [ ] Unsubscribe from devices
- [ ] Filter by device ID
- [ ] Filter by telemetry type
- [ ] UI controls for filtering

---

### US-005: Telemetry History
**Description:** As a dashboard user, I want to view historical telemetry so I can analyze past device behavior.

**Acceptance Criteria:**
- [ ] Query telemetry by time range
- [ ] Limit results (max 1000 items)
- [ ] Cache recent telemetry client-side
- [ ] Display in chronological order
- [ ] Export to CSV/JSON

---

### US-006: Rate Limiting
**Description:** As a system administrator, I want rate limiting to prevent WebSocket abuse.

**Acceptance Criteria:**
- [ ] Max 100 messages/sec per client
- [ ] Burst limit of 200 messages
- [ ] Returns 429 when limit exceeded
- [ ] Rate limit metrics tracked
- [ ] Configurable limits

---

### US-007: Connection Status UI
**Description:** As a user, I want to see WebSocket connection status so I know if real-time updates are working.

**Acceptance Criteria:**
- [ ] Visual connection indicator (green/red)
- [ ] Connection timestamp
- [ ] Reconnection attempt counter
- [ ] Last message received time
- [ ] Error message display

---

### US-008: Enhanced FCM Integration
**Description:** As a device, I want reliable command delivery via FCM when WebSocket is unavailable.

**Acceptance Criteria:**
- [ ] Automatic FCM fallback when WebSocket fails
- [ ] FCM retry logic (3 attempts)
- [ ] FCM success/failure metrics
- [ ] FCM token validation
- [ ] Topic messaging support

---

## 4. Functional Requirements

### FR-1: Message Queue System
- FR-1.1: Queue implementation using channel + database backup
- FR-1.2: Max queue size: 1000 messages per device
- FR-1.3: Message TTL: 7 days
- FR-1.4: FIFO delivery order
- FR-1.5: Queue metrics: size, age, delivery rate

### FR-2: Automatic Reconnection
- FR-2.1: Exponential backoff: 1s, 2s, 4s, 8s, 16s
- FR-2.2: Max attempts: 5
- FR-2.3: Connection timeout: 10s
- FR-2.4: Reconnection jitter: ±500ms
- FR-2.5: Reconnection metrics: attempts, success rate

### FR-3: Message Compression
- FR-3.1: GZIP compression for messages >1KB
- FR-3.2: Compression level: default
- FR-3.3: Fallback to uncompressed on error
- FR-3.4: Compression ratio tracking
- FR-3.5: Configurable threshold

### FR-4: Telemetry Filtering
- FR-4.1: Subscribe to devices via WebSocket message
- FR-4.2: Unsubscribe via WebSocket message
- FR-4.3: Server-side filter implementation
- FR-4.4: Max subscriptions: 50 devices per client
- FR-4.5: Subscription metrics

### FR-5: Telemetry History
- FR-5.1: SQLite storage with time index
- FR-5.2: Query by device ID + time range
- FR-5.3: Max results: 1000
- FR-5.4: Cache TTL: 1 hour
- FR-5.5: Export formats: CSV, JSON

### FR-6: Rate Limiting
- FR-6.1: Token bucket algorithm
- FR-6.2: 100 messages/sec limit
- FR-6.3: 200 message burst
- FR-6.4: Per-client tracking
- FR-6.5: Metrics: requests, rejects, rate

### FR-7: Connection Status
- FR-7.1: WebSocket state tracking
- FR-7.2: Last message timestamp
- FR-7.3: Reconnection attempt counter
- FR-7.4: Error message display
- FR-7.5: Visual indicators

### FR-8: Enhanced FCM
- FR-8.1: Automatic fallback detection
- FR-8.2: 3 retry attempts
- FR-8.3: Exponential backoff between retries
- FR-8.4: Success/failure metrics
- FR-8.5: Token validation

---

## 5. File Structure & Implementation

### New Files Required (12 files)

```
apps/api/internal/ws/
├── message_queue.go           # Message queue implementation
├── message_queue_test.go
├── compression.go             # Message compression
├── compression_test.go
├── rate_limiter.go            # WebSocket rate limiting
└── rate_limiter_test.go

apps/api/pkg/storage/
├── message_queue.go           # Database-backed queue
└── message_queue_test.go

apps/api/internal/api/handlers/
├── telemetry_history.go       # Historical data endpoints
└── telemetry_history_test.go

apps/web/src/
├── websocket/
│   ├── enhanced-client.ts     # Enhanced WebSocket client
│   ├── telemetry-cache.ts     # Client-side cache
│   └── connection-monitor.ts # Connection status
└── components/
    └── DeviceTelemetry.tsx     # Enhanced telemetry component
```

### Files to Update (8 files)

```
apps/api/internal/ws/
├── hub.go                     # Add queue support, filtering
├── client.go                  # Add compression, reconnection
└── websocket_handler.go       # Add rate limiting

apps/api/internal/fcm/
├── notifier.go                # Add retry logic
└── fcm.go                     # Add metrics

apps/api/pkg/models/
├── updater.go                 # Add new message types
└── command_frame.go           # Add queue metadata

apps/web/src/
├── api-client.ts              # Update WebSocket usage
└── App.tsx                    # Add connection status
```

---

## 6. Database Schema Changes

```sql
-- Message queue table
CREATE TABLE message_queue (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    message_type TEXT NOT NULL,
    message_payload TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'queued',
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_message_queue_device ON message_queue(device_id);
CREATE INDEX idx_message_queue_status ON message_queue(status);
CREATE INDEX idx_message_queue_expires ON message_queue(expires_at);

-- Telemetry history table (enhanced)
CREATE TABLE telemetry_history (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    received_at INTEGER NOT NULL,
    payload TEXT NOT NULL,
    risk_score INTEGER,
    buffer_level INTEGER,
    thermal_temp REAL,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_telemetry_device_time ON telemetry_history(device_id, received_at DESC);
CREATE INDEX idx_telemetry_risk ON telemetry_history(device_id, risk_score);
```

---

## 7. Detailed Implementation

### 7.1 Message Queue Implementation

**File: `apps/api/internal/ws/message_queue.go`**
```go
package ws

import (
    "context"
    "encoding/json"
    "sync"
    "time"

    "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
    "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

type MessageQueue struct {
    mu          sync.RWMutex
    memoryQueue map[string][]models.CommandFrame // deviceID -> []messages
    store       *storage.Store
    maxSize     int
}

func NewMessageQueue(store *storage.Store, maxSize int) *MessageQueue {
    return &MessageQueue{
        memoryQueue: make(map[string][]models.CommandFrame),
        store:       store,
        maxSize:     maxSize,
    }
}

func (q *MessageQueue) Enqueue(deviceID string, frame models.CommandFrame) error {
    q.mu.Lock()
    defer q.mu.Unlock()

    // Add to memory queue
    q.memoryQueue[deviceID] = append(q.memoryQueue[deviceID], frame)
    
    // Persist to database
    data, _ := json.Marshal(frame)
    return q.store.EnqueueMessage(context.Background(), deviceID, data)
}

func (q *MessageQueue) Dequeue(deviceID string) (models.CommandFrame, bool) {
    q.mu.Lock()
    defer q.mu.Unlock()

    if len(q.memoryQueue[deviceID]) == 0 {
        return models.CommandFrame{}, false
    }

    // Get oldest message
    frame := q.memoryQueue[deviceID][0]
    q.memoryQueue[deviceID] = q.memoryQueue[deviceID][1:]
    
    // Remove from database
    q.store.DequeueMessage(context.Background(), deviceID, frame.DispatchID)
    
    return frame, true
}

func (q *MessageQueue) ProcessQueue(deviceID string, sendFunc func(models.CommandFrame) bool) {
    q.mu.Lock()
    defer q.mu.Unlock()

    for len(q.memoryQueue[deviceID]) > 0 {
        frame := q.memoryQueue[deviceID][0]
        if sendFunc(frame) {
            // Successfully sent, remove from queue
            q.memoryQueue[deviceID] = q.memoryQueue[deviceID][1:]
            q.store.DequeueMessage(context.Background(), deviceID, frame.DispatchID)
        } else {
            // Send buffer full, stop processing
            break
        }
    }
}

func (q *MessageQueue) Cleanup(expiresBefore time.Time) error {
    q.mu.Lock()
    defer q.mu.Unlock()

    return q.store.CleanupQueue(context.Background(), expiresBefore.UnixMilli())
}
```

---

### 7.2 Enhanced Client with Reconnection

**File: `apps/api/internal/ws/client.go` (updates)**
```go
// Add to Client struct
type Client struct {
    // ... existing fields
    reconnectAttempts int
    lastMessageTime  time.Time
    subscriptions    map[string]bool // deviceID -> subscribed
}

// Add reconnection logic to ReadPump
func (c *Client) ReadPump() {
    defer func() {
        c.Hub.Unregister(c)
        closeConn(c.Conn, c.log, "readPump")
        if c.reconnectAttempts < 5 {
            go c.attemptReconnect()
        }
    }()
    // ... existing code
}

func (c *Client) attemptReconnect() {
    time.Sleep(time.Duration(c.reconnectAttempts) * time.Second)
    c.reconnectAttempts++
    // In production, this would trigger a reconnection from client side
    // Server-side reconnection not typically implemented
}

// Add subscription management
func (c *Client) Subscribe(deviceID string) {
    c.subscriptions[deviceID] = true
}

func (c *Client) Unsubscribe(deviceID string) {
    delete(c.subscriptions, deviceID)
}

func (c *Client) IsSubscribed(deviceID string) bool {
    return c.subscriptions[deviceID]
}
```

---

### 7.3 Message Compression

**File: `apps/api/internal/ws/compression.go`**
```go
package ws

import (
    "bytes"
    "compress/gzip"
    "encoding/json"
    "io"

    "github.com/gorilla/websocket"
    "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

const compressionThreshold = 1024 // 1KB

func CompressMessage(frame models.CommandFrame) ([]byte, error) {
    data, err := json.Marshal(frame)
    if err != nil {
        return nil, err
    }

    if len(data) < compressionThreshold {
        return data, nil
    }

    var buf bytes.Buffer
    writer := gzip.NewWriter(&buf)
    if _, err := writer.Write(data); err != nil {
        return nil, err
    }
    if err := writer.Close(); err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}

func DecompressMessage(data []byte) (models.CommandFrame, error) {
    reader, err := gzip.NewReader(bytes.NewBuffer(data))
    if err != nil {
        return models.CommandFrame{}, err
    }
    defer reader.Close()

    decompressed, err := io.ReadAll(reader)
    if err != nil {
        return models.CommandFrame{}, err
    }

    var frame models.CommandFrame
    if err := json.Unmarshal(decompressed, &frame); err != nil {
        return models.CommandFrame{}, err
    }

    return frame, nil
}

func (c *Client) WriteCompressed(frame models.CommandFrame) error {
    data, err := CompressMessage(frame)
    if err != nil {
        return c.Conn.WriteJSON(frame) // Fallback to uncompressed
    }

    return c.Conn.WriteMessage(websocket.BinaryMessage, data)
}
```

---

### 7.4 Rate Limiter

**File: `apps/api/internal/ws/rate_limiter.go`**
```go
package ws

import (
    "sync"
    "time"

    "golang.org/x/time/rate"
)

type RateLimiter struct {
    mu      sync.Mutex
    limits  map[string]*rate.Limiter // clientID -> limiter
    rate    rate.Limit
    burst   int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limits: make(map[string]*rate.Limiter),
        rate:   r,
        burst:  b,
    }
}

func (rl *RateLimiter) GetLimiter(clientID string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.limits[clientID]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limits[clientID] = limiter
    }

    return limiter
}

func (rl *RateLimiter) Allow(clientID string) bool {
    return rl.GetLimiter(clientID).Allow()
}

func (rl *RateLimiter) Remove(clientID string) {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    delete(rl.limits, clientID)
}
```

---

### 7.5 Telemetry History Endpoint

**File: `apps/api/internal/api/handlers/telemetry_history.go`**
```go
package handlers

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

type TelemetryController struct {
    store *storage.Store
}

func NewTelemetryController(store *storage.Store) *TelemetryController {
    return &TelemetryController{store: store}
}

// GetTelemetryHistory retrieves historical telemetry for a device.
// GET /v1/device/:id/telemetry
func (t *TelemetryController) GetTelemetryHistory(c *gin.Context) {
    deviceID := c.Param("id")
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, map[string]string{
            "error": "bad_request",
            "message": "device id required",
        })
        return
    }

    // Parse query parameters
    limit := 100
    if l := c.Query("limit"); l != "" {
        if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
            limit = parsed
        }
    }

    sinceStr := c.Query("since")
    var since time.Time
    var err error
    if sinceStr != "" {
        since, err = time.Parse(time.RFC3339, sinceStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, map[string]string{
                "error": "bad_request",
                "message": "invalid since parameter",
            })
            return
        }
    }

    // Query telemetry
    telemetry, err := t.store.GetTelemetryHistory(c.Request.Context(), deviceID, since, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "query_failed",
            "message": err.Error(),
        })
        return
        
    }

    c.JSON(http.StatusOK, map[string]interface{}{
        "deviceId": deviceID,
        "telemetry": telemetry,
        "count": len(telemetry),
    })
}

// ExportTelemetry exports telemetry to CSV/JSON.
// GET /v1/device/:id/telemetry/export
func (t *TelemetryController) ExportTelemetry(c *gin.Context) {
    // Implementation for CSV/JSON export
}
```

---

### 7.6 Enhanced FCM Notifier

**File: `apps/api/internal/fcm/notifier.go` (updates)**
```go
// Add retry logic
func (c *Client) SendSilentWake(ctx context.Context, wake SilentWake) error {
    // ... existing validation code

    var lastErr error
    for attempt := 1; attempt <= 3; attempt++ {
        result, err := client.Send(ctx, msg)
        if err == nil {
            c.log.Info("fcm silent wake sent", 
                "deviceId", wake.DeviceID,
                "dispatchId", wake.DispatchID,
                "messageId", result,
                "attempt", attempt)
            return nil
        }

        lastErr = err
        c.log.Warn("fcm send failed", 
            "deviceId", wake.DeviceID,
            "dispatchId", wake.DispatchID,
            "attempt", attempt,
            "err", err)

        // Exponential backoff
        time.Sleep(time.Duration(attempt*attempt) * time.Second)
    }

    return fmt.Errorf("fcm send failed after 3 attempts: %w", lastErr)
}
```

---

### 7.7 Frontend WebSocket Client

**File: `apps/web/src/websocket/enhanced-client.ts`**
```typescript
import { TelemetryCache } from './telemetry-cache';

export class EnhancedWebSocketClient {
    private socket: WebSocket | null = null;
    private url: string;
    private deviceId: string;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000; // ms
    private messageQueue: any[] = [];
    private telemetryCache: TelemetryCache;
    private onMessage: (data: any) => void;
    private onStatusChange: (status: string) => void;
    private subscriptions: Set<string>;

    constructor(url: string, deviceId: string, cache: TelemetryCache) {
        this.url = url;
        this.deviceId = deviceId;
        this.telemetryCache = cache;
        this.subscriptions = new Set();
    }

    setOnMessage(callback: (data: any) => void) {
        this.onMessage = callback;
    }

    setOnStatusChange(callback: (status: string) => void) {
        this.onStatusChange = callback;
    }

    connect() {
        this.socket = new WebSocket(this.url);
        this.updateStatus('connecting');

        this.socket.onopen = () => {
            this.reconnectAttempts = 0;
            this.updateStatus('connected');
            this.authenticate();
            this.processQueue();
        };

        this.socket.onclose = (event) => {
            this.updateStatus('disconnected');
            this.attemptReconnect();
        };

        this.socket.onerror = (error) => {
            this.updateStatus('error');
            console.error('WebSocket error', error);
        };

        this.socket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (err) {
                console.error('Message parsing error', err);
            }
        };
    }

    private updateStatus(status: string) {
        if (this.onStatusChange) {
            this.onStatusChange(status);
        }
    }

    private authenticate() {
        if (this.socket?.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify({
                type: 'auth',
                deviceId: this.deviceId
            }));
        }
    }

    private attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = this.reconnectDelay * this.reconnectAttempts;
            console.log(`Attempting to reconnect in ${delay}ms...`);
            setTimeout(() => this.connect(), delay);
        } else {
            console.error('Max reconnection attempts reached');
            this.updateStatus('failed');
        }
    }

    private processQueue() {
        while (this.messageQueue.length > 0 && this.socket?.readyState === WebSocket.OPEN) {
            const message = this.messageQueue.shift();
            this.socket.send(JSON.stringify(message));
        }
    }

    send(message: any) {
        if (this.socket?.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(message));
        } else {
            console.warn('WebSocket not connected, queuing message');
            this.messageQueue.push(message);
        }
    }

    subscribe(deviceId: string) {
        this.subscriptions.add(deviceId);
        this.send({
            type: 'subscribe',
            deviceId: deviceId
        });
    }

    unsubscribe(deviceId: string) {
        this.subscriptions.delete(deviceId);
        this.send({
            type: 'unsubscribe',
            deviceId: deviceId
        });
    }

    isSubscribed(deviceId: string): boolean {
        return this.subscriptions.has(deviceId);
    }

    private handleMessage(message: any) {
        switch (message.type) {
            case 'telemetry':
                this.telemetryCache.add(this.deviceId, message);
                if (this.onMessage) {
                    this.onMessage(message);
                }
                break;
            case 'command':
                if (this.onMessage) {
                    this.onMessage(message);
                }
                break;
            case 'status':
                // Handle status updates
                break;
            case 'pong':
                // Handle heartbeat
                break;
            default:
                console.warn('Unknown message type', message.type);
        }
    }

    disconnect() {
        if (this.socket) {
            this.socket.close();
            this.socket = null;
        }
    }

    getStatus(): string {
        if (!this.socket) return 'disconnected';
        switch (this.socket.readyState) {
            case WebSocket.CONNECTING: return 'connecting';
            case WebSocket.OPEN: return 'connected';
            case WebSocket.CLOSING: return 'closing';
            case WebSocket.CLOSED: return 'disconnected';
            default: return 'unknown';
        }
    }
}
```

---

### 7.8 Telemetry Cache

**File: `apps/web/src/websocket/telemetry-cache.ts`**
```typescript
export class TelemetryCache {
    private cache: Map<string, any[]>;
    private maxItems: number;

    constructor(maxItems = 1000) {
        this.cache = new Map();
        this.maxItems = maxItems;
    }

    add(deviceId: string, data: any) {
        if (!this.cache.has(deviceId)) {
            this.cache.set(deviceId, []);
        }
        const deviceCache = this.cache.get(deviceId)!;
        deviceCache.push(data);
        if (deviceCache.length > this.maxItems) {
            deviceCache.shift(); // Remove oldest
        }
    }

    get(deviceId: string, limit = 100): any[] {
        const data = this.cache.get(deviceId) || [];
        return data.slice(-limit); // Return most recent
    }

    clear(deviceId: string) {
        this.cache.delete(deviceId);
    }

    getAll(): Map<string, any[]> {
        return new Map(this.cache); // Return copy
    }

    exportToCSV(deviceId: string): string {
        const data = this.get(deviceId);
        if (data.length === 0) return '';

        // Get headers from first item
        const headers = Object.keys(data[0]);
        const csvRows = [];

        // Header row
        csvRows.push(headers.join(','));

        // Data rows
        for (const row of data) {
            const values = headers.map(header => {
                const value = row[header];
                // Escape quotes
                return `"${String(value).replace(/"/g, '""')}"`;
            });
            csvRows.push(values.join(','));
        }

        return csvRows.join('\n');
    }

    exportToJSON(deviceId: string): string {
        return JSON.stringify(this.get(deviceId), null, 2);
    }
}
```

---

### 7.9 Connection Monitor

**File: `apps/web/src/websocket/connection-monitor.ts`**
```typescript
export class ConnectionMonitor {
    private status: string;
    private lastMessageTime: number;
    private lastConnectTime: number;
    private reconnectCount: number;
    private errorMessage: string;

    constructor() {
        this.status = 'disconnected';
        this.lastMessageTime = 0;
        this.lastConnectTime = 0;
        this.reconnectCount = 0;
        this.errorMessage = '';
    }

    updateStatus(status: string) {
        this.status = status;
        if (status === 'connected') {
            this.lastConnectTime = Date.now();
            this.reconnectCount = 0;
            this.errorMessage = '';
        } else if (status === 'disconnected' || status === 'error') {
            this.reconnectCount++;
        }
    }

    updateLastMessage() {
        this.lastMessageTime = Date.now();
    }

    setErrorMessage(message: string) {
        this.errorMessage = message;
    }

    getStatus(): string {
        return this.status;
    }

    getUptime(): number {
        if (this.status !== 'connected') return 0;
        return Date.now() - this.lastConnectTime;
    }

    getLastMessageAge(): number {
        if (this.lastMessageTime === 0) return -1;
        return Date.now() - this.lastMessageTime;
    }

    getReconnectCount(): number {
        return this.reconnectCount;
    }

    getErrorMessage(): string {
        return this.errorMessage;
    }

    isConnected(): boolean {
        return this.status === 'connected';
    }

    isHealthy(): boolean {
        return this.isConnected() && this.getLastMessageAge() < 60000; // < 60s
    }
}
```

---

### 7.10 Enhanced Telemetry Component

**File: `apps/web/src/components/DeviceTelemetry.tsx`**
```tsx
import React, { useState, useEffect, useRef } from 'react';
import { EnhancedWebSocketClient } from '../websocket/enhanced-client';
import { TelemetryCache } from '../websocket/telemetry-cache';
import { ConnectionMonitor } from '../websocket/connection-monitor';

interface TelemetryProps {
    deviceId: string;
    apiBaseUrl: string;
}

export const DeviceTelemetry: React.FC<TelemetryProps> = ({ deviceId, apiBaseUrl }) => {
    const [telemetry, setTelemetry] = useState<any[]>([]);
    const [status, setStatus] = useState('disconnected');
    const [uptime, setUptime] = useState(0);
    const [lastMessage, setLastMessage] = useState('Never');
    const [reconnects, setReconnects] = useState(0);
    const wsClientRef = useRef<EnhancedWebSocketClient | null>(null);
    const cacheRef = useRef(new TelemetryCache());
    const monitorRef = useRef(new ConnectionMonitor());

    useEffect(() => {
        // Initialize WebSocket client
        const wsUrl = `${apiBaseUrl.replace(/^http/, 'ws')}/v1/device/${deviceId}/stream`;
        const client = new EnhancedWebSocketClient(wsUrl, deviceId, cacheRef.current);
        wsClientRef.current = client;

        client.setOnMessage((message) => {
            setTelemetry(prev => [...prev.slice(-99), message]);
            monitorRef.current.updateLastMessage();
            setLastMessage(new Date().toLocaleTimeString());
        });

        client.setOnStatusChange((status) => {
            setStatus(status);
            monitorRef.current.updateStatus(status);
            setReconnects(monitorRef.current.getReconnectCount());
        });

        client.connect();

        // Start uptime timer
        const timer = setInterval(() => {
            setUptime(monitorRef.current.getUptime());
        }, 1000);

        return () => {
            clearInterval(timer);
            client.disconnect();
        };
    }, [deviceId, apiBaseUrl]);

    const formatUptime = (ms: number): string => {
        const seconds = Math.floor(ms / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
    };

    const getStatusColor = () => {
        switch (status) {
            case 'connected': return 'bg-green-500';
            case 'connecting': return 'bg-yellow-500';
            case 'disconnected': return 'bg-red-500';
            case 'error': return 'bg-red-700';
            default: return 'bg-gray-500';
        }
    };

    return (
        <div className="telemetry-container p-4 bg-gray-800 rounded-lg">
            {/* Connection Status */}
            <div className="connection-status mb-4 p-3 bg-gray-700 rounded">
                <div className="flex items-center justify-between">
                    <div className="flex items-center">
                        <div className={`w-3 h-3 rounded-full ${getStatusColor()} mr-2`}></div>
                        <span className="font-semibold">Connection: {status}</span>
                    </div>
                    <div className="text-sm text-gray-400">
                        Uptime: {formatUptime(uptime)}, 
                        Reconnects: {reconnects}, 
                        Last Message: {lastMessage}
                    </div>
                </div>
            </div>

            {/* Telemetry Data */}
            <div className="telemetry-data overflow-x-auto">
                <table className="w-full text-sm">
                    <thead className="bg-gray-700">
                        <tr>
                            <th className="p-2 text-left">Timestamp</th>
                            <th className="p-2 text-left">Type</th>
                            <th className="p-2 text-left">Battery</th>
                            <th className="p-2 text-left">CPU</th>
                            <th className="p-2 text-left">Memory</th>
                            <th className="p-2 text-left">Network</th>
                            <th className="p-2 text-left">Risk</th>
                        </tr>
                    </thead>
                    <tbody>
                        {telemetry.map((item, index) => (
                            <tr key={index} className="border-b border-gray-600 hover:bg-gray-700">
                                <td className="p-2">{new Date(item.timestamp).toLocaleString()}</td>
                                <td className="p-2">{item.type || 'telemetry'}</td>
                                <td className="p-2">{item.battery}%</td>
                                <td className="p-2">{item.cpu}%</td>
                                <td className="p-2">{item.memory}MB</td>
                                <td className="p-2">{item.network}Mbps</td>
                                <td className="p-2">
                                    <span className={`px-2 py-1 rounded ${item.riskScore > 75 ? 'bg-red-500' : item.riskScore > 50 ? 'bg-yellow-500' : 'bg-green-500'}`}>
                                        {item.riskScore}
                                    </span>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {/* Export Controls */}
            <div className="export-controls mt-4 flex justify-end">
                <button
                    onClick={() => {
                        const csv = cacheRef.current.exportToCSV(deviceId);
                        const blob = new Blob([csv], { type: 'text/csv' });
                        const url = URL.createObjectURL(blob);
                        const a = document.createElement('a');
                        a.href = url;
                        a.download = `telemetry-${deviceId}-${new Date().toISOString()}.csv`;
                        a.click();
                    }}
                    className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                >
                    Export CSV
                </button>
            </div>
        </div>
    );
};
```

---

## 8. Configuration

### Environment Variables
```bash
# WebSocket
WS_MAX_MESSAGE_SIZE=1048576       # 1MB
WS_READ_TIMEOUT=60               # 60 seconds
WS_WRITE_TIMEOUT=10             # 10 seconds
WS_PING_PERIOD=30               # 30 seconds
WS_PONG_WAIT=60                 # 60 seconds
WS_MAX_QUEUE_SIZE=1000          # Messages per device
WS_QUEUE_TTL=604800             # 7 days in seconds
WS_RATE_LIMIT=100               # Messages per second
WS_RATE_BURST=200               # Burst limit

# Telemetry
TELEMETRY_HISTORY_LIMIT=1000     # Max items per query
TELEMETRY_CACHE_TTL=3600        # 1 hour in seconds
```

---

## 9. Testing Strategy

### Unit Tests
- Message queue enqueue/dequeue operations
- Compression/decompression
- Rate limiting logic
- Telemetry filtering
- Subscription management

### Integration Tests
- WebSocket connection lifecycle
- Message delivery with reconnection
- FCM fallback scenarios
- Telemetry history queries
- Rate limiting enforcement

### Performance Tests
- 10,000 concurrent WebSocket connections
- Message throughput (target: 10,000 msg/sec)
- Latency measurements (target: <500ms)
- Memory usage under load

### End-to-End Tests
- Device connects → sends telemetry → dashboard receives
- Command sent → device offline → FCM fallback → device receives on reconnect
- Reconnection scenarios
- Rate limiting enforcement

---

## 10. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| WebSocket uptime | 99.9% | Percentage of time connected |
| Message delivery rate | 100% | Messages successfully delivered |
| FCM fallback success | 95% | FCM sends successful |
| Reconnection success | 99% | Successful reconnections |
| Bandwidth reduction | 40% | Compression savings |
| Latency (p99) | <500ms | Command delivery time |
| Concurrent connections | 10,000 | Max supported clients |
| Rate limit compliance | 100% | Excess requests blocked |
| Telemetry query time | <100ms | History query performance |

---

## 11. Open Questions

1. **Should we implement WebSocket message encryption?** For sensitive commands.

2. **Should we add session resumption?** To maintain state across reconnects.

3. **Should we implement message acknowledgments?** For guaranteed delivery.

4. **Should we add WebSocket load balancing?** For horizontal scaling.

5. **Should we implement client-side message batching?** To reduce small messages.

---

## 13. FCM Enhancement Details

### FCM-001: Retry Logic Implementation
**Description:** Add automatic retry for failed FCM sends with exponential backoff.

**Implementation:**
```go
// apps/api/internal/fcm/notifier.go
func (c *Client) SendSilentWake(ctx context.Context, wake SilentWake) error {
    // Existing validation code...

    var lastErr error
    for attempt := 1; attempt <= 3; attempt++ {
        result, err := client.Send(ctx, msg)
        if err == nil {
            c.log.Info("fcm silent wake sent", 
                "deviceId", wake.DeviceID,
                "dispatchId", wake.DispatchID,
                "messageId", result,
                "attempt", attempt)
            return nil
        }

        lastErr = err
        c.log.Warn("fcm send failed", 
            "deviceId", wake.DeviceID,
            "dispatchId", wake.DispatchID,
            "attempt", attempt,
            "err", err)

        // Exponential backoff: 1s, 4s, 9s
        time.Sleep(time.Duration(attempt*attempt) * time.Second)
    }

    return fmt.Errorf("fcm send failed after 3 attempts: %w", lastErr)
}
```

**Files to Update:**
- `apps/api/internal/fcm/notifier.go`

**Testing:**
- Unit tests for retry logic
- Integration tests with FCM mock
- Metrics verification

---

### FCM-002: Success/Failure Metrics
**Description:** Track FCM delivery metrics for monitoring and alerting.

**Implementation:**
```go
// apps/api/internal/fcm/notifier.go
type FCMMetrics struct {
    SuccessCount int
    FailureCount int
    LastSuccess   time.Time
    LastFailure   time.Time
}

func (c *Client) incrementSuccess() {
    // Increment counters, log to monitoring system
}

func (c *Client) incrementFailure() {
    // Increment counters, log to monitoring system
}
```

**Files to Update:**
- `apps/api/internal/fcm/notifier.go`
- `apps/api/internal/fcm/fcm.go` (add metrics struct)

**Testing:**
- Metrics collection verification
- Dashboard integration tests

---

### FCM-003: Token Validation
**Description:** Validate FCM tokens before sending to reduce failures.

**Implementation:**
```go
// apps/api/internal/fcm/notifier.go
func validateFCMToken(token string) bool {
    // Check token format
    if len(token) == 0 {
        return false
    }
    
    // Check token length (FCM tokens are typically 150-200 chars)
    if len(token) < 50 || len(token) > 500 {
        return false
    }
    
    // Check token characters (base64-like)
    const validChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_="
    for _, char := range token {
        if !strings.Contains(validChars, string(char)) {
            return false
        }
    }
    
    return true
}
```

**Files to Update:**
- `apps/api/internal/fcm/notifier.go`

**Testing:**
- Unit tests with valid/invalid tokens
- Integration with device registration

---

### FCM-004: Topic Messaging Support
**Description:** Support FCM topic messaging for device groups.

**Implementation:**
```go
// apps/api/internal/fcm/notifier.go
type TopicMessage struct {
    Topic      string
    Command    string
    DispatchID string
}

func (c *Client) SendToTopic(ctx context.Context, msg TopicMessage) error {
    if c == nil || !c.enabled {
        return ErrDisabled
    }

    client := c.Messaging()
    if client == nil {
        return ErrUnavailable
    }

    fcmMsg := &messaging.Message{
        Topic: msg.Topic,
        Android: &messaging.AndroidConfig{
            Priority: "high",
            TTL:      ptr24Hours(),
            Data: map[string]string{
                "action":      "WAKE_DAEMON",
                "command":     msg.Command,
                "dispatch_id": msg.DispatchID,
            },
        },
        Data: map[string]string{
            "action":      "WAKE_DAEMON",
            "command":     msg.Command,
            "dispatch_id": msg.DispatchID,
        },
    }

    result, err := client.Send(ctx, fcmMsg)
    if err != nil {
        c.log.Warn("fcm topic send failed",
            "topic", msg.Topic,
            "dispatchId", msg.DispatchID,
            "err", err)
        return fmt.Errorf("fcm topic send: %w", err)
    }

    c.log.Info("fcm topic message sent",
        "topic", msg.Topic,
        "dispatchId", msg.DispatchID,
        "messageId", result)

    return nil
}
```

**Files to Update:**
- `apps/api/internal/fcm/notifier.go`
- `apps/api/internal/fcm/fcm.go` (add topic support)

**Testing:**
- Unit tests for topic messaging
- Integration with device groups

---

### FCM-005: Message Prioritization
**Description:** Add priority levels for FCM messages.

**Implementation:**
```go
// apps/api/pkg/models/fcm.go
type FCMPriority string

const (
    PriorityHigh   FCMPriority = "high"
    PriorityNormal FCMPriority = "normal"
)

type SilentWake struct {
    Token      string
    Command    string
    DispatchID string
    DeviceID   string
    Priority   FCMPriority
}

// apps/api/internal/fcm/notifier.go
func (c *Client) SendSilentWake(ctx context.Context, wake SilentWake) error {
    // ... existing code

    priority := "high" // default
    if wake.Priority == PriorityNormal {
        priority = "normal"
    }

    msg := &messaging.Message{
        Token: wake.Token,
        Android: &messaging.AndroidConfig{
            Priority: priority, // Use configured priority
            // ... rest of config
        },
        // ... rest of message
    }

    // ... rest of implementation
}
```

**Files to Update:**
- `apps/api/pkg/models/fcm.go` (new file)
- `apps/api/internal/fcm/notifier.go`
- `apps/api/internal/api/handlers/command.go` (add priority parameter)

**Testing:**
- Unit tests for priority handling
- Integration tests with different priorities

---

## 14. Implementation Order

```
Phase 1: Core Infrastructure (P0)
├── Message queue implementation
├── WebSocket reconnection logic
├── FCM retry enhancements      [FCM-001]
├── FCM metrics tracking        [FCM-002]
└── Rate limiting

Phase 2: Performance & Features (P1)
├── Message compression
├── Telemetry filtering
├── Telemetry history
├── FCM token validation        [FCM-003]
└── Connection status UI

Phase 3: Enhancements (P2)
├── WebSocket encryption
├── Session resumption
├── FCM topic messaging         [FCM-004]
├── FCM priority levels         [FCM-005]
├── Advanced aggregation
└── Offline sync
```

---

*Document Version: 1.1*  
*Status: Ready for Review*  
*Next Steps: Review with team, then implementation*