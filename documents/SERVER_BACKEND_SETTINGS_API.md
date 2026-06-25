# Server Backend - Settings API Specification

> **Version:** 1.0
> **Status:** Draft
> **Created:** 2026-06-25
> **Target:** Production MVP
> **Go Package:** `github.com/VinnsEdesigner/vyzorix/apps/api`

---

## Table of Contents

1. Overview
2. Current State Analysis
3. Required API Endpoints
4. Database Schema
5. Backend File Structure
6. Handler Specifications
7. Service Layer
8. GraphQL Schema
9. Error Handling
10. Rate Limiting & Security
11. File Changes Summary
12. Implementation Order

---

## 1. Overview

### 1.1 Purpose

This document maps out the server-side requirements to support the Settings page as specified in SETTINGS_PAGE.md.

### 1.2 Frontend Requirements Summary

| Feature | Description | Required Endpoints |
|---------|-------------|-------------------|
| Connection Settings | Server URL, device ID, timeout | GET/PATCH /v1/auth/me/settings |
| Operator Settings | Display name, account info | GET/PATCH /v1/auth/me |
| Thresholds | Risk, thermal, buffer thresholds | GET/PATCH /v1/auth/me/thresholds |
| Notifications | Email, push, webhook settings | GET/PATCH /v1/auth/me/notifications |
| Advanced Settings | Buffer limits, danger zone | GET/PATCH /v1/auth/me/settings |
| Webhook Testing | Test webhook endpoint | POST /v1/auth/me/notifications/webhook/test |

---

## 2. Current State Analysis

### 2.1 Existing Related Endpoints

| Endpoint | Status | Handler | Notes |
|----------|--------|---------|-------|
| GET /v1/auth/me | EXISTS | AuthHandler.GetMe | Get operator info |
| PATCH /v1/auth/me | EXISTS | AuthHandler.UpdateMe | Update name |
| GET /v1/auth/me/settings | EXISTS | SettingsHandler.Get | Get settings |
| PATCH /v1/auth/me/settings | EXISTS | SettingsHandler.Patch | Update settings |

### 2.2 Missing Endpoints

| Endpoint | Status | Notes |
|----------|--------|-------|
| GET /v1/auth/me/thresholds | **MISSING** | Dedicated thresholds endpoint |
| PATCH /v1/auth/me/thresholds | **MISSING** | Update thresholds |
| GET /v1/auth/me/notifications | **MISSING** | Notification settings |
| PATCH /v1/auth/me/notifications | **MISSING** | Update notifications |
| POST /v1/auth/me/notifications/webhook/test | **MISSING** | Test webhook |

### 2.3 Existing Data Models

| Model | Location | Status |
|-------|----------|--------|
| operator.Operator | domain/operator/entity.go | EXISTS |
| operator.Settings | domain/operator/settings.go | EXISTS |

---

## 3. Required API Endpoints

### 3.1 GET /v1/auth/me/settings

Get current operator settings.

**Response (200 OK):**
```json
{
  "client": {
    "serverUrl": "https://api.example.com",
    "deviceId": "861234567890123",
    "requestTimeoutMs": 8000,
    "autoReconnect": true,
    "strictHmac": false,
    "logBufferLimit": 500,
    "signalHistoryLimit": 240
  },
  "thresholds": {
    "riskWarn": 70,
    "riskCrit": 85,
    "thermalWarn": 45,
    "thermalCrit": 50,
    "bufferWarn": 30,
    "bufferCrit": 15
  },
  "notifications": {
    "enabled": true,
    "channels": ["email", "push"],
    "email": {
      "thresholdBreach": true,
      "deviceOffline": true,
      "deviceOnline": true,
      "updateAvailable": false,
      "commandFailed": true,
      "registrationRequest": true
    },
    "push": {
      "thresholdBreach": true,
      "deviceOffline": true,
      "deviceOnline": false,
      "updateAvailable": false,
      "commandFailed": false,
      "registrationRequest": false
    },
    "webhook": {
      "enabled": false,
      "url": "",
      "types": []
    }
  }
}
```

---

### 3.2 PATCH /v1/auth/me/settings

Update operator settings.

**Request:**
```json
{
  "client": {
    "serverUrl": "https://api.example.com",
    "deviceId": "861234567890123",
    "requestTimeoutMs": 8000,
    "autoReconnect": true,
    "strictHmac": false,
    "logBufferLimit": 500,
    "signalHistoryLimit": 240
  }
}
```

**Response (200 OK):**
```json
{
  "client": {
    "serverUrl": "https://api.example.com",
    "deviceId": "861234567890123",
    "requestTimeoutMs": 8000,
    "autoReconnect": true,
    "strictHmac": false,
    "logBufferLimit": 500,
    "signalHistoryLimit": 240
  }
}
```

**Validation Rules:**
| Field | Rule |
|-------|------|
| serverUrl | Required, valid HTTP/HTTPS URL |
| deviceId | Optional, alphanumeric |
| requestTimeoutMs | 500-60000 |
| autoReconnect | Boolean |
| strictHmac | Boolean |
| logBufferLimit | 50-5000 |
| signalHistoryLimit | 30-2000 |

---

### 3.3 GET /v1/auth/me/thresholds

Get alert thresholds.

**Response (200 OK):**
```json
{
  "riskWarn": 70,
  "riskCrit": 85,
  "thermalWarn": 45,
  "thermalCrit": 50,
  "bufferWarn": 30,
  "bufferCrit": 15
}
```

---

### 3.4 PATCH /v1/auth/me/thresholds

Update alert thresholds.

**Request:**
```json
{
  "riskWarn": 75,
  "riskCrit": 90
}
```

**Response (200 OK):**
```json
{
  "riskWarn": 75,
  "riskCrit": 90,
  "thermalWarn": 45,
  "thermalCrit": 50,
  "bufferWarn": 30,
  "bufferCrit": 15
}
```

**Validation Rules:**
| Field | Range | Rule |
|-------|-------|------|
| riskWarn | 0-100 | Must be < riskCrit |
| riskCrit | 0-100 | Must be > riskWarn |
| thermalWarn | 0-100 | Must be < thermalCrit |
| thermalCrit | 0-100 | Must be > thermalWarn |
| bufferWarn | 0-100 | Must be > bufferCrit (inverted) |
| bufferCrit | 0-100 | Must be < bufferWarn (inverted) |

---

### 3.5 GET /v1/auth/me/notifications

Get notification settings.

**Response (200 OK):**
```json
{
  "enabled": true,
  "channels": ["email", "push"],
  "email": {
    "thresholdBreach": true,
    "deviceOffline": true,
    "deviceOnline": true,
    "updateAvailable": false,
    "commandFailed": true,
    "registrationRequest": true
  },
  "push": {
    "thresholdBreach": true,
    "deviceOffline": true,
    "deviceOnline": false,
    "updateAvailable": false,
    "commandFailed": false,
    "registrationRequest": false
  },
  "webhook": {
    "enabled": false,
    "url": "",
    "secret": "",
    "types": []
  }
}
```

---

### 3.6 PATCH /v1/auth/me/notifications

Update notification settings.

**Request:**
```json
{
  "enabled": true,
  "email": {
    "thresholdBreach": true,
    "deviceOffline": true,
    "deviceOnline": false
  },
  "webhook": {
    "enabled": true,
    "url": "https://hooks.example.com/vyzorix",
    "types": ["threshold_breach", "device_offline"]
  }
}
```

**Response (200 OK):**
```json
{
  "enabled": true,
  "channels": ["email", "webhook"],
  "email": {
    "thresholdBreach": true,
    "deviceOffline": true,
    "deviceOnline": false,
    "updateAvailable": false,
    "commandFailed": true,
    "registrationRequest": true
  },
  "push": {
    "thresholdBreach": true,
    "deviceOffline": true,
    "deviceOnline": false,
    "updateAvailable": false,
    "commandFailed": false,
    "registrationRequest": false
  },
  "webhook": {
    "enabled": true,
    "url": "https://hooks.example.com/vyzorix",
    "secret": "generated_or_updated_secret",
    "types": ["threshold_breach", "device_offline"]
  }
}
```

---

### 3.7 POST /v1/auth/me/notifications/webhook/test

Test webhook configuration.

**Request:**
```json
{
  "url": "https://hooks.example.com/vyzorix"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "statusCode": 200,
  "responseTime": 145
}
```

**Response (400 Bad Request):**
```json
{
  "success": false,
  "error": "webhook_timeout",
  "message": "Webhook did not respond within 10 seconds"
}
```

---

### 3.8 POST /v1/auth/me/notifications/webhook/rotate

Rotate webhook secret.

**Response (200 OK):**
```json
{
  "secret": "new_webhook_secret_here"
}
```

---

### 3.9 GET /v1/auth/me

Get operator profile (enhanced).

**Response (200 OK):**
```json
{
  "id": "op_abc123",
  "email": "operator@example.com",
  "name": "John Doe",
  "role": "admin",
  "permissions": [
    "register_devices",
    "deregister_devices",
    "send_commands",
    "view_telemetry",
    "push_updates",
    "manage_settings"
  ],
  "notifications": {
    "enabled": true,
    "channels": ["email", "push"]
  },
  "createdAt": 1705334400000,
  "lastLoginAt": 1718900000000
}
```

---

### 3.10 PATCH /v1/auth/me

Update operator profile.

**Request:**
```json
{
  "name": "John Doe"
}
```

**Response (200 OK):**
```json
{
  "id": "op_abc123",
  "email": "operator@example.com",
  "name": "John Doe",
  "role": "admin"
}
```

---

### 3.11 POST /v1/auth/me/settings/reset

Reset settings to defaults (super_admin only).

**Response (200 OK):**
```json
{
  "client": {
    "serverUrl": "",
    "deviceId": "",
    "requestTimeoutMs": 8000,
    "autoReconnect": true,
    "strictHmac": false,
    "logBufferLimit": 500,
    "signalHistoryLimit": 240
  },
  "thresholds": {
    "riskWarn": 70,
    "riskCrit": 85,
    "thermalWarn": 45,
    "thermalCrit": 50,
    "bufferWarn": 30,
    "bufferCrit": 15
  },
  "notifications": {
    "enabled": true,
    "channels": ["email"],
    "email": {
      "thresholdBreach": true,
      "deviceOffline": true,
      "deviceOnline": true,
      "updateAvailable": false,
      "commandFailed": true,
      "registrationRequest": true
    },
    "push": {
      "thresholdBreach": false,
      "deviceOffline": false,
      "deviceOnline": false,
      "updateAvailable": false,
      "commandFailed": false,
      "registrationRequest": false
    },
    "webhook": {
      "enabled": false,
      "url": "",
      "secret": "",
      "types": []
    }
  }
}
```

---

## 4. Database Schema

### 4.1 Operator Settings Table (NEW)

```sql
CREATE TABLE operator_settings (
    operator_id TEXT PRIMARY KEY REFERENCES operators(id) ON DELETE CASCADE,
    
    -- Connection settings
    server_url TEXT,
    device_id TEXT,
    request_timeout_ms INT DEFAULT 8000,
    auto_reconnect BOOLEAN DEFAULT TRUE,
    strict_hmac BOOLEAN DEFAULT FALSE,
    
    -- Buffer settings
    log_buffer_limit INT DEFAULT 500,
    signal_history_limit INT DEFAULT 240,
    
    -- Thresholds
    risk_warn INT DEFAULT 70,
    risk_crit INT DEFAULT 85,
    thermal_warn INT DEFAULT 45,
    thermal_crit INT DEFAULT 50,
    buffer_warn INT DEFAULT 30,
    buffer_crit INT DEFAULT 15,
    
    -- Notification settings
    notifications_enabled BOOLEAN DEFAULT TRUE,
    notify_email BOOLEAN DEFAULT TRUE,
    notify_push BOOLEAN DEFAULT FALSE,
    notify_webhook BOOLEAN DEFAULT FALSE,
    webhook_url TEXT,
    webhook_secret TEXT,
    webhook_types TEXT[], -- Array of notification types
    
    -- Feature toggles per notification type
    notify_threshold_breach BOOLEAN DEFAULT TRUE,
    notify_device_offline BOOLEAN DEFAULT TRUE,
    notify_device_online BOOLEAN DEFAULT TRUE,
    notify_update_available BOOLEAN DEFAULT FALSE,
    notify_command_failed BOOLEAN DEFAULT TRUE,
    notify_registration_request BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_operator_settings_id ON operator_settings(operator_id);
```

### 4.2 Operator Notifications Audit Log (NEW)

```sql
CREATE TABLE notification_audit_log (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    operator_id TEXT NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    channel TEXT NOT NULL,
    payload JSONB,
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT idx_operator_notifications FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

CREATE INDEX idx_notification_audit_operator ON notification_audit_log(operator_id, sent_at DESC);
CREATE INDEX idx_notification_audit_type ON notification_audit_log(event_type, sent_at DESC);
```

---

## 5. Backend File Structure

```
apps/api/internal/
├── api/
│   ├── handlers/
│   │   ├── auth/
│   │   │   ├── auth_handler.go       # EXISTS
│   │   │   └── auth_settings_handler.go   # MODIFIED - settings handlers
│   │   ├── operator/
│   │   │   ├── operator_handler.go   # EXISTS
│   │   │   ├── operator_thresholds_handler.go  # NEW - threshold handlers
│   │   │   └── operator_notifications_handler.go  # NEW - notification handlers
│   │   └── router.go               # MODIFIED
│   └── middleware/
│       └── ...
├── application/
│   ├── auth/
│   │   ├── auth_service.go         # EXISTS
│   │   └── auth_settings_service.go # NEW - settings service
│   ├── operator/
│   │   ├── operator_service.go      # EXISTS
│   │   └── operator_settings_service.go  # NEW - threshold/notification service
│   └── dto/
│       └── settings_dto.go          # NEW - settings DTOs
├── domain/
│   ├── auth/
│   │   └── session.go              # EXISTS
│   ├── operator/
│   │   ├── operator_entity.go       # EXISTS
│   │   ├── operator_settings.go     # NEW - settings entity
│   │   └── operator_repository.go   # EXISTS - add settings methods
│   └── errors.go                   # EXISTS
├── infrastructure/
│   ├── storage/
│   │   ├── operator_storage.go     # EXISTS - add settings
│   │   ├── notification_storage.go # NEW - notification audit
│   │   └── migrations/
│   │       └── 001_operator_settings.sql
│   └── webhook/
│       └── webhook_client.go        # NEW - webhook testing
```

---

## 6. Handler Specifications

### 6.1 Settings Handler

**File:** `api/handlers/auth/auth_settings_handler.go`

```go
package auth

import (
    "net/http"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
    "github.com/gin-gonic/gin"
)

type SettingsHandler struct {
    settingsService *auth.SettingsService
}

func NewSettingsHandler(svc *auth.SettingsService) *SettingsHandler {
    return &SettingsHandler{settingsService: svc}
}

// GetSettings handles GET /v1/auth/me/settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
    op := middleware.GetOperatorFromContext(c)
    if op == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    settings, err := h.settingsService.GetSettings(c.Request.Context(), op.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get settings"})
        return
    }

    c.JSON(http.StatusOK, settings)
}

// PatchSettings handles PATCH /v1/auth/me/settings
func (h *SettingsHandler) PatchSettings(c *gin.Context) {
    op := middleware.GetOperatorFromContext(c)
    if op == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var req struct {
        Client struct {
            ServerURL         string `json:"serverUrl"`
            DeviceID         string `json:"deviceId"`
            RequestTimeoutMs *int   `json:"requestTimeoutMs"`
            AutoReconnect    *bool  `json:"autoReconnect"`
            StrictHmac      *bool  `json:"strictHmac"`
            LogBufferLimit   *int   `json:"logBufferLimit"`
            SignalHistoryLimit *int `json:"signalHistoryLimit"`
        } `json:"client"`
        Reset *bool `json:"reset"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    if req.Reset != nil && *req.Reset {
        // Reset to defaults (super_admin only)
        if op.Role != "super_admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            return
        }
        settings, err := h.settingsService.ResetSettings(c.Request.Context(), op.ID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset settings"})
            return
        }
        c.JSON(http.StatusOK, settings)
        return
    }

    settings, err := h.settingsService.UpdateClientSettings(c.Request.Context(), op.ID, &req.Client)
    if err != nil {
        if err == auth.ErrValidationError {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
        return
    }

    c.JSON(http.StatusOK, settings)
}
```

---

### 6.2 Threshold Handler

**File:** `api/handlers/operator/thresholds.go`

```go
package operator

import (
    "net/http"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
    "github.com/gin-gonic/gin"
)

type ThresholdHandler struct {
    service *operator.SettingsService
}

func NewThresholdHandler(svc *operator.SettingsService) *ThresholdHandler {
    return &ThresholdHandler{service: svc}
}

// GetThresholds handles GET /v1/auth/me/thresholds
func (h *ThresholdHandler) GetThresholds(c *gin.Context) {
    op := middleware.GetOperatorFromContext(c)
    if op == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    thresholds, err := h.service.GetThresholds(c.Request.Context(), op.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get thresholds"})
        return
    }

    c.JSON(http.StatusOK, thresholds)
}

// PatchThresholds handles PATCH /v1/auth/me/thresholds
func (h *ThresholdHandler) PatchThresholds(c *gin.Context) {
    op := middleware.GetOperatorFromContext(c)
    if op == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var req operator.ThresholdsInput
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    thresholds, err := h.service.UpdateThresholds(c.Request.Context(), op.ID, &req)
    if err != nil {
        if err == operator.ErrValidation {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update thresholds"})
        return
    }

    c.JSON(http.StatusOK, thresholds)
}
```

---

### 6.3 Notification Handler

**File:** `api/handlers/operator/notifications.go`

```go
package operator

import (
    "net/http"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
    "github.com/gin-gonic/gin"
)

type NotificationHandler struct {
    service       *operator.SettingsService
    webhookClient *webhook.Client
}

func NewNotificationHandler(svc *operator.SettingsService, wh *webhook.Client) *NotificationHandler {
    return &NotificationHandler{
        service:       svc,
        webhookClient: wh,
    }
}

// GetNotifications handles GET /v1/auth/me/notifications
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
    op := middleware.GetOperatorFromContext(c)
    if op == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    notifications, err := h.service.GetNotifications(c.Request.Context(), op.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notifications"})
        return
    }

    c.JSON(http.StatusOK, notifications)
}

// PatchNotifications handles PATCH /v1/auth/me/notifications
func (h *NotificationHandler) PatchNotifications(c *gin.Context) {
    op := middleware.GetOperatorFromContext(c)
    if op == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var req operator.NotificationInput
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    notifications, err := h.service.UpdateNotifications(c.Request.Context(), op.ID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notifications"})
        return
    }

    c.JSON(http.StatusOK, notifications)
}

// TestWebhook handles POST /v1/auth/me/notifications/webhook/test
func (h *NotificationHandler) TestWebhook(c *gin.Context) {
    var req struct {
        URL string `json:"url" binding:"required,url"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    result, err := h.webhookClient.TestWebhook(c.Request.Context(), req.URL)
    if err != nil {
        if err == webhook.ErrTimeout {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "error": "webhook_timeout",
                "message": "Webhook did not respond within 10 seconds",
            })
            return
        }
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error": "webhook_error",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, result)
}

// RotateWebhookSecret handles POST /v1/auth/me/notifications/webhook/rotate
func (h *NotificationHandler) RotateWebhookSecret(c *gin.Context) {
    op := middleware.GetOperatorFromContext(c)
    if op == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    secret, err := h.service.RotateWebhookSecret(c.Request.Context(), op.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate secret"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"secret": secret})
}
```

---

## 7. Service Layer

### 7.1 Settings Service

**File:** `application/auth/settings.go`

```go
package auth

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "net/url"
    "time"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

var (
    ErrValidationError = errors.New("validation error")
    ErrNotFound       = errors.New("settings not found")
)

type SettingsService struct {
    repo *operator.Repository
}

func NewSettingsService(repo *operator.Repository) *SettingsService {
    return &SettingsService{repo: repo}
}

type ClientSettingsInput struct {
    ServerURL           *string `json:"serverUrl,omitempty"`
    DeviceID            *string `json:"deviceId,omitempty"`
    RequestTimeoutMs   *int    `json:"requestTimeoutMs,omitempty"`
    AutoReconnect      *bool   `json:"autoReconnect,omitempty"`
    StrictHmac         *bool   `json:"strictHmac,omitempty"`
    LogBufferLimit     *int    `json:"logBufferLimit,omitempty"`
    SignalHistoryLimit  *int    `json:"signalHistoryLimit,omitempty"`
}

type ClientSettings struct {
    ServerURL           string `json:"serverUrl"`
    DeviceID            string `json:"deviceId"`
    RequestTimeoutMs   int    `json:"requestTimeoutMs"`
    AutoReconnect      bool   `json:"autoReconnect"`
    StrictHmac         bool   `json:"strictHmac"`
    LogBufferLimit     int    `json:"logBufferLimit"`
    SignalHistoryLimit int    `json:"signalHistoryLimit"`
}

type FullSettings struct {
    Client       *ClientSettings       `json:"client,omitempty"`
    Thresholds   *operator.Thresholds `json:"thresholds,omitempty"`
    Notifications *operator.NotificationSettings `json:"notifications,omitempty"`
}

func (s *SettingsService) GetSettings(ctx context.Context, operatorID string) (*FullSettings, error) {
    settings, err := s.repo.GetSettings(ctx, operatorID)
    if err != nil {
        if err == operator.ErrNotFound {
            // Create default settings
            settings = &operator.Settings{
                OperatorID:           operatorID,
                RequestTimeoutMs:     8000,
                AutoReconnect:       true,
                LogBufferLimit:      500,
                SignalHistoryLimit:   240,
                NotificationsEnabled: true,
                NotifyEmail:          true,
            }
            if err := s.repo.CreateSettings(ctx, settings); err != nil {
                return nil, err
            }
        } else {
            return nil, err
        }
    }

    return &FullSettings{
        Client: &ClientSettings{
            ServerURL:           settings.ServerURL,
            DeviceID:            settings.DeviceID,
            RequestTimeoutMs:   settings.RequestTimeoutMs,
            AutoReconnect:      settings.AutoReconnect,
            StrictHmac:         settings.StrictHmac,
            LogBufferLimit:     settings.LogBufferLimit,
            SignalHistoryLimit: settings.SignalHistoryLimit,
        },
        Thresholds: &operator.Thresholds{
            RiskWarn:    settings.RiskWarn,
            RiskCrit:    settings.RiskCrit,
            ThermalWarn: settings.ThermalWarn,
            ThermalCrit: settings.ThermalCrit,
            BufferWarn:  settings.BufferWarn,
            BufferCrit:  settings.BufferCrit,
        },
        Notifications: s.buildNotificationSettings(settings),
    }, nil
}

func (s *SettingsService) UpdateClientSettings(ctx context.Context, operatorID string, input *ClientSettingsInput) (*FullSettings, error) {
    settings, err := s.repo.GetSettings(ctx, operatorID)
    if err != nil {
        return nil, err
    }

    // Apply updates
    if input.ServerURL != nil {
        if err := validateServerURL(*input.ServerURL); err != nil {
            return nil, ErrValidationError
        }
        settings.ServerURL = *input.ServerURL
    }
    if input.DeviceID != nil {
        settings.DeviceID = *input.DeviceID
    }
    if input.RequestTimeoutMs != nil {
        if *input.RequestTimeoutMs < 500 || *input.RequestTimeoutMs > 60000 {
            return nil, ErrValidationError
        }
        settings.RequestTimeoutMs = *input.RequestTimeoutMs
    }
    if input.AutoReconnect != nil {
        settings.AutoReconnect = *input.AutoReconnect
    }
    if input.StrictHmac != nil {
        settings.StrictHmac = *input.StrictHmac
    }
    if input.LogBufferLimit != nil {
        if *input.LogBufferLimit < 50 || *input.LogBufferLimit > 5000 {
            return nil, ErrValidationError
        }
        settings.LogBufferLimit = *input.LogBufferLimit
    }
    if input.SignalHistoryLimit != nil {
        if *input.SignalHistoryLimit < 30 || *input.SignalHistoryLimit > 2000 {
            return nil, ErrValidationError
        }
        settings.SignalHistoryLimit = *input.SignalHistoryLimit
    }

    settings.UpdatedAt = time.Now()
    if err := s.repo.UpdateSettings(ctx, settings); err != nil {
        return nil, err
    }

    return s.GetSettings(ctx, operatorID)
}

func (s *SettingsService) ResetSettings(ctx context.Context, operatorID string) (*FullSettings, error) {
    settings := operator.DefaultSettings(operatorID)
    if err := s.repo.ResetSettings(ctx, operatorID, settings); err != nil {
        return nil, err
    }
    return s.GetSettings(ctx, operatorID)
}

func validateServerURL(urlStr string) error {
    if urlStr == "" {
        return nil
    }
    u, err := url.Parse(urlStr)
    if err != nil {
        return errors.New("invalid URL")
    }
    if u.Scheme != "http" && u.Scheme != "https" {
        return errors.New("URL must use http or https")
    }
    return nil
}
```

---

### 7.2 Threshold Service

**File:** `application/operator/thresholds.go`

```go
package operator

import (
    "context"
    "errors"
    "time"
)

var ErrValidation = errors.New("validation error")

type ThresholdService struct {
    repo *Repository
}

func NewThresholdService(repo *Repository) *ThresholdService {
    return &ThresholdService{repo: repo}
}

type ThresholdsInput struct {
    RiskWarn    *int `json:"riskWarn,omitempty"`
    RiskCrit   *int `json:"riskCrit,omitempty"`
    ThermalWarn *int `json:"thermalWarn,omitempty"`
    ThermalCrit *int `json:"thermalCrit,omitempty"`
    BufferWarn *int `json:"bufferWarn,omitempty"`
    BufferCrit *int `json:"bufferCrit,omitempty"`
}

func (s *ThresholdService) GetThresholds(ctx context.Context, operatorID string) (*Thresholds, error) {
    settings, err := s.repo.GetSettings(ctx, operatorID)
    if err != nil {
        return nil, err
    }

    return &Thresholds{
        RiskWarn:    settings.RiskWarn,
        RiskCrit:    settings.RiskCrit,
        ThermalWarn: settings.ThermalWarn,
        ThermalCrit: settings.ThermalCrit,
        BufferWarn:  settings.BufferWarn,
        BufferCrit:  settings.BufferCrit,
    }, nil
}

func (s *ThresholdService) UpdateThresholds(ctx context.Context, operatorID string, input *ThresholdsInput) (*Thresholds, error) {
    settings, err := s.repo.GetSettings(ctx, operatorID)
    if err != nil {
        return nil, err
    }

    if input.RiskWarn != nil {
        if *input.RiskWarn < 0 || *input.RiskWarn > 100 {
            return nil, ErrValidation
        }
        settings.RiskWarn = *input.RiskWarn
    }
    if input.RiskCrit != nil {
        if *input.RiskCrit < 0 || *input.RiskCrit > 100 {
            return nil, ErrValidation
        }
        settings.RiskCrit = *input.RiskCrit
    }
    // Validate: riskWarn must be < riskCrit
    if settings.RiskWarn >= settings.RiskCrit {
        return nil, errors.New("riskWarn must be less than riskCrit")
    }

    if input.ThermalWarn != nil {
        settings.ThermalWarn = *input.ThermalWarn
    }
    if input.ThermalCrit != nil {
        settings.ThermalCrit = *input.ThermalCrit
    }
    if settings.ThermalWarn >= settings.ThermalCrit {
        return nil, errors.New("thermalWarn must be less than thermalCrit")
    }

    if input.BufferWarn != nil {
        settings.BufferWarn = *input.BufferWarn
    }
    if input.BufferCrit != nil {
        settings.BufferCrit = *input.BufferCrit
    }
    // Note: Buffer is inverted (lower is worse)
    if settings.BufferCrit >= settings.BufferWarn {
        return nil, errors.New("bufferCrit must be less than bufferWarn")
    }

    settings.UpdatedAt = time.Now()
    if err := s.repo.UpdateSettings(ctx, settings); err != nil {
        return nil, err
    }

    return s.GetThresholds(ctx, operatorID)
}
```

---

### 7.3 Notification Service

**File:** `application/operator/notifications.go`

```go
package operator

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "time"
)

type NotificationService struct {
    repo *Repository
}

func NewNotificationService(repo *Repository) *NotificationService {
    return &NotificationService{repo: repo}
}

type NotificationInput struct {
    Enabled *bool             `json:"enabled,omitempty"`
    Email   *EmailSettings   `json:"email,omitempty"`
    Push    *PushSettings    `json:"push,omitempty"`
    Webhook *WebhookSettings `json:"webhook,omitempty"`
}

type EmailSettings struct {
    ThresholdBreach      *bool `json:"thresholdBreach,omitempty"`
    DeviceOffline        *bool `json:"deviceOffline,omitempty"`
    DeviceOnline         *bool `json:"deviceOnline,omitempty"`
    UpdateAvailable      *bool `json:"updateAvailable,omitempty"`
    CommandFailed        *bool `json:"commandFailed,omitempty"`
    RegistrationRequest  *bool `json:"registrationRequest,omitempty"`
}

type WebhookSettings struct {
    Enabled *bool    `json:"enabled,omitempty"`
    URL     *string `json:"url,omitempty"`
    Types   []string `json:"types,omitempty"`
}

func (s *NotificationService) GetNotifications(ctx context.Context, operatorID string) (*NotificationSettings, error) {
    settings, err := s.repo.GetSettings(ctx, operatorID)
    if err != nil {
        return nil, err
    }

    return s.buildNotificationSettings(settings), nil
}

func (s *NotificationService) UpdateNotifications(ctx context.Context, operatorID string, input *NotificationInput) (*NotificationSettings, error) {
    settings, err := s.repo.GetSettings(ctx, operatorID)
    if err != nil {
        return nil, err
    }

    if input.Enabled != nil {
        settings.NotificationsEnabled = *input.Enabled
    }

    if input.Email != nil {
        if input.Email.ThresholdBreach != nil {
            settings.NotifyThresholdBreach = *input.Email.ThresholdBreach
        }
        if input.Email.DeviceOffline != nil {
            settings.NotifyDeviceOffline = *input.Email.DeviceOffline
        }
        if input.Email.DeviceOnline != nil {
            settings.NotifyDeviceOnline = *input.Email.DeviceOnline
        }
        if input.Email.UpdateAvailable != nil {
            settings.NotifyUpdateAvailable = *input.Email.UpdateAvailable
        }
        if input.Email.CommandFailed != nil {
            settings.NotifyCommandFailed = *input.Email.CommandFailed
        }
        if input.Email.RegistrationRequest != nil {
            settings.NotifyRegistrationRequest = *input.Email.RegistrationRequest
        }
    }

    if input.Push != nil {
        settings.NotifyPush = true // Enable push channel if push settings provided
    }

    if input.Webhook != nil {
        if input.Webhook.Enabled != nil {
            settings.NotifyWebhook = *input.Webhook.Enabled
        }
        if input.Webhook.URL != nil {
            settings.WebhookURL = *input.Webhook.URL
            // Generate secret if new URL provided
            if settings.WebhookSecret == "" {
                secret := generateSecret(32)
                settings.WebhookSecret = secret
            }
        }
        if input.Webhook.Types != nil {
            settings.WebhookTypes = input.Webhook.Types
        }
    }

    settings.UpdatedAt = time.Now()
    if err := s.repo.UpdateSettings(ctx, settings); err != nil {
        return nil, err
    }

    return s.buildNotificationSettings(settings), nil
}

func (s *NotificationService) RotateWebhookSecret(ctx context.Context, operatorID string) (string, error) {
    settings, err := s.repo.GetSettings(ctx, operatorID)
    if err != nil {
        return "", err
    }

    secret := generateSecret(32)
    settings.WebhookSecret = secret
    settings.UpdatedAt = time.Now()

    if err := s.repo.UpdateSettings(ctx, settings); err != nil {
        return "", err
    }

    return secret, nil
}

func generateSecret(length int) string {
    bytes := make([]byte, length)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
```

---

## 8. GraphQL Schema

### 8.1 Types

```graphql
type ClientSettings {
  serverUrl: String
  deviceId: String
  requestTimeoutMs: Int!
  autoReconnect: Boolean!
  strictHmac: Boolean!
  logBufferLimit: Int!
  signalHistoryLimit: Int!
}

type Thresholds {
  riskWarn: Int!
  riskCrit: Int!
  thermalWarn: Int!
  thermalCrit: Int!
  bufferWarn: Int!
  bufferCrit: Int!
}

type NotificationChannels {
  email: Boolean!
  push: Boolean!
  webhook: Boolean!
}

type NotificationTypes {
  thresholdBreach: Boolean!
  deviceOffline: Boolean!
  deviceOnline: Boolean!
  updateAvailable: Boolean!
  commandFailed: Boolean!
  registrationRequest: Boolean!
}

type WebhookSettings {
  enabled: Boolean!
  url: String
  secret: String
  types: [String!]!
}

type NotificationSettings {
  enabled: Boolean!
  channels: NotificationChannels!
  email: NotificationTypes!
  push: NotificationTypes!
  webhook: WebhookSettings!
}

type OperatorSettings {
  client: ClientSettings!
  thresholds: Thresholds!
  notifications: NotificationSettings!
}

type ThresholdUpdateResult {
  riskWarn: Int!
  riskCrit: Int!
  thermalWarn: Int!
  thermalCrit: Int!
  bufferWarn: Int!
  bufferCrit: Int!
}

type WebhookTestResult {
  success: Boolean!
  statusCode: Int
  responseTime: Int
  error: String
}
```

### 8.2 Queries

```graphql
type Query {
  mySettings: OperatorSettings!
  myThresholds: Thresholds!
  myNotifications: NotificationSettings!
}
```

### 8.3 Mutations

```graphql
type Mutation {
  updateMySettings(input: ClientSettingsInput!): OperatorSettings!
  resetMySettings: OperatorSettings!
  updateMyThresholds(input: ThresholdsInput!): ThresholdUpdateResult!
  updateMyNotifications(input: NotificationInput!): NotificationSettings!
  testWebhook(url: String!): WebhookTestResult!
  rotateWebhookSecret: String!
}

input ClientSettingsInput {
  serverUrl: String
  deviceId: String
  requestTimeoutMs: Int
  autoReconnect: Boolean
  strictHmac: Boolean
  logBufferLimit: Int
  signalHistoryLimit: Int
}

input ThresholdsInput {
  riskWarn: Int
  riskCrit: Int
  thermalWarn: Int
  thermalCrit: Int
  bufferWarn: Int
  bufferCrit: Int
}

input NotificationInput {
  enabled: Boolean
  email: EmailNotificationInput
  push: PushNotificationInput
  webhook: WebhookNotificationInput
}

input EmailNotificationInput {
  thresholdBreach: Boolean
  deviceOffline: Boolean
  deviceOnline: Boolean
  updateAvailable: Boolean
  commandFailed: Boolean
  registrationRequest: Boolean
}

input PushNotificationInput {
  thresholdBreach: Boolean
  deviceOffline: Boolean
  deviceOnline: Boolean
  updateAvailable: Boolean
  commandFailed: Boolean
  registrationRequest: Boolean
}

input WebhookNotificationInput {
  enabled: Boolean
  url: String
  types: [String!]
}
```

---

## 9. Error Handling

### 9.1 Error Response Format

```json
{
  "error": "error_code",
  "message": "Human readable message",
  "details": {}
}
```

### 9.2 Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| unauthorized | 401 | Authentication required |
| forbidden | 403 | Insufficient permissions |
| validation_error | 400 | Invalid input |
| not_found | 404 | Resource not found |
| internal_error | 500 | Server error |

### 9.3 Validation Errors

| Field | Validation |
|-------|-----------|
| serverUrl | Must be valid HTTP/HTTPS URL |
| requestTimeoutMs | 500-60000 |
| logBufferLimit | 50-5000 |
| signalHistoryLimit | 30-2000 |
| riskWarn < riskCrit | Warning must be < Critical |
| thermalWarn < thermalCrit | Warning must be < Critical |
| bufferCrit < bufferWarn | Critical < Warning (inverted) |
| webhook.url | Must be valid URL if provided |

---

## 10. Rate Limiting & Security

### 10.1 Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| GET /v1/auth/me/settings | 60 | 1 minute |
| PATCH /v1/auth/me/settings | 30 | 1 minute |
| GET /v1/auth/me/thresholds | 60 | 1 minute |
| PATCH /v1/auth/me/thresholds | 30 | 1 minute |
| POST /v1/auth/me/settings/reset | 5 | 1 hour |
| POST /v1/auth/me/notifications/webhook/test | 10 | 1 hour |

### 10.2 Security Requirements

1. **Authentication** - All endpoints require authenticated operator
2. **Authorization** - Reset requires super_admin role
3. **Webhook Secrets** - Stored hashed, rotated via dedicated endpoint
4. **Input Validation** - All inputs validated server-side
5. **Audit Logging** - Log all settings changes

---

## 11. File Changes Summary

### 11.1 Total File Count

| Category | New | Modified | Total |
|----------|-----|----------|-------|
| Domain Layer | 2 | 1 | 3 |
| Application Layer | 3 | 0 | 3 |
| Handler Layer | 3 | 1 | 4 |
| Infrastructure | 3 | 1 | 4 |
| GraphQL | 2 | 2 | 4 |
| Router | 0 | 1 | 1 |
| **TOTAL** | **13** | **6** | **19** |

### 11.2 All Files Listed

#### Domain Layer (2 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| domain/operator/operator_settings.go | NEW | Settings entity types |
| domain/operator/operator_repository.go | NEW | Repository interface |
| domain/operator/operator_entity.go | MODIFIED | Add settings fields |

#### Application Layer (3 NEW)

| File | Status | Purpose |
|------|--------|---------|
| application/auth/auth_settings_service.go | NEW | Settings service |
| application/operator/operator_thresholds_service.go | NEW | Threshold service |
| application/operator/operator_notifications_service.go | NEW | Notification service |

#### Handler Layer (3 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| api/handlers/auth/auth_settings_handler.go | MODIFIED | Add new settings methods |
| api/handlers/operator/operator_thresholds_handler.go | NEW | Threshold handlers |
| api/handlers/operator/operator_notifications_handler.go | NEW | Notification handlers |
| api/handlers/router.go | MODIFIED | Add routes |

#### Infrastructure (3 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| infrastructure/storage/operator_storage.go | MODIFIED | Add settings queries |
| infrastructure/storage/migrations/ | NEW | SQL migrations |
| infrastructure/webhook/webhook_client.go | NEW | Webhook testing |
| infrastructure/notification/notification_audit.go | NEW | Audit logging |

#### GraphQL (2 NEW, 2 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| api/graphql/schema/objects.go | MODIFIED | Add settings types |
| api/graphql/schema/resolver.go | MODIFIED | Add resolvers |
| api/graphql/schema/schema.go | MODIFIED | Add mutations |

---

## 12. Implementation Order

### Phase 1: Database (Day 1)
1. Create operator_settings table
2. Create notification_audit_log table
3. Add columns to operators table if needed
4. Test migrations

### Phase 2: Domain Layer (Day 1)
1. Create `domain/operator/operator_settings.go`
2. Create `domain/operator/operator_repository.go`
3. Update `domain/operator/operator_entity.go`

### Phase 3: Infrastructure (Day 1-2)
1. Update `infrastructure/storage/operator_storage.go`
2. Create `infrastructure/webhook/webhook_client.go`
3. Create `infrastructure/notification/notification_audit.go`

### Phase 4: Application Layer (Day 2)
1. Create `application/auth/auth_settings_service.go`
2. Create `application/operator/operator_thresholds_service.go`
3. Create `application/operator/operator_notifications_service.go`

### Phase 5: Handlers (Day 2)
1. Create `handlers/auth/auth_settings_handler.go`
2. Create `handlers/operator/operator_thresholds_handler.go`
3. Create `handlers/operator/operator_notifications_handler.go`
4. Wire routes

### Phase 6: GraphQL (Day 2-3)
1. Add schema types
2. Add resolvers
3. Test queries

### Phase 7: Integration (Day 3)
1. Test webhook delivery
2. Test notification routing
3. Test audit logging

---

*Document Version: 1.0*
*Status: Ready for Implementation*
