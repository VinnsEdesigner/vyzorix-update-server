# UPDATE_SERVER_ARCHITECTURE_SPEC.md — C2 & Update Server Architecture Specification

> **Updated for Phase 1.5** — This document reflects the current state of the Vyzorix Update Server with all bug fixes #2-15 and frontend improvements implemented.

## Document Purpose

This documents the system and file specification for the **Vyzorix Command & Control (C2) and Update Server** (`vyzorix-update-server`). 

It details:
1. The general operational mechanics of the server
2. How the full-duplex WebSocket Hub, the Firebase Push Notification engine, and the SQLite storage layer coordinate
3. The detailed technical specification for **every single backend file** in the repository

---

## 1. Core System Architecture & Operational Workflows

The server is engineered as a lightweight, static-binary Go web service with a React frontend. It coordinates three major runtime boundaries:

```
                           ╔═══════════════════════════════════════╗
                           ║      VYZORIX CONTROL PLATFORM       ║
                           ╚═══════════════════════════════════════╝

    ┌────────────────────────────────────────────────────────────────────────┐
    │                          Frontend (React)                             │
    │  ┌────────────────────────────────────────────────────────────────┐  │
    │  │                    TanStack Start Router                        │  │
    │  │  Dashboard  │  Device  │  Diagnostics  │  Alerts  │  Settings │  │
    │  └────────────────────────────────────────────────────────────────┘  │
    │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
    │  │ WS Stream    │  │ Config      │  │ Logs        │  │ Operator  │ │
    │  │ useDevice    │  │ useVyzorix   │  │ useLogs     │  │ Auth      │ │
    │  │ Stream       │  │ Config      │  │             │  │ useAuth    │ │
    │  └──────────────┘  └──────────────┘  └──────────────┘  └────────────┘ │
    └────────────────────────────────────────────────────────────────────────┘
                                      │
                              WebSocket + REST
                                      │
    ┌────────────────────────────────────────────────────────────────────────┐
    │                          Backend (Go)                                 │
    │  ┌──────────────────────────────────────────────────────────────────┐  │
    │  │                         Gin HTTP Router                         │  │
    │  │  /v1/auth/*  │  /v1/device/*  │  /api/v1/*  │  /healthz         │  │
    │  └──────────────────────────────────────────────────────────────────┘  │
    │  ┌────────────────────┐              ┌────────────────────────────┐   │
    │  │  WebSocket Hub      │              │       Services             │   │
    │  │  /v1/device/:id/    │              │  ┌─────────────────────┐  │   │
    │  │  stream              │              │  │  FCM Notifier       │  │   │
    │  └──────────┬───────────┘              │  │  (SafeNotifier)     │  │   │
    │             │                             │  └─────────────────────┘  │   │
    │  ┌──────────▼───────────────────────────┴──────────────────────────┐   │
    │  │                    Middleware Stack                             │   │
    │  │  CORS Handler  │  JWT Auth  │  Rate Limiter  │  Request Logger │   │
    │  └─────────────────────────────────────────────────────────────────┘   │
    └────────────────────────────────────────────────────────────────────────┘
                                      │
    ┌────────────────────────────────────────────────────────────────────────┐
    │                         Storage Layer                                  │
    │                      SQLite (WAL Mode)                                 │
    │   Operators  │  Sessions  │  Devices  │  Commands  │  Secrets     │
    └────────────────────────────────────────────────────────────────────────┘
```

### A. The REST API Layer
* **Purpose**: Manages public, low-overhead endpoints for the client updater (`/api/v1/version`, `/api/v1/changelog`) and handles signed APK package transfers (`/bin/*.apk`) using HTTP Range Support for resumable, chunked downloads. It also exposes private, authenticated routes for dashboard logins and manual command triggers.

### B. The Bidirectional WebSocket Hub
* **Purpose**: Maintains persistent full-duplex TCP connections with active client daemons. Telemetry data (active CPU usage, memory thresholds, active routes, and risk scores) is parsed and broadcast directly to connected React web control panels in real-time. Incoming dashboard command frames are instantly routed to target device sockets with sub-millisecond dispatch times.

### C. The Firebase Messaging (FCM) Signaling Engine
* **Purpose**: Translates out-of-band commands into high-priority silent push intents. If a target client goes offline or is put into deep Doze sleep, the server uses the FCM SDK to bypass system-level background execution limits and awaken the daemon.

---

## 2. Root Files & Deployment Specifications

```
vyzorix-update-server/
├── go.mod / go.sum         # Go module definition (1.22+)
├── main.go                 # Server entrypoint
├── Makefile                # Build, test, docker commands
├── Dockerfile              # Multi-stage build (Go + React)
├── docker-compose.yml      # Local dev environment
├── render.yaml             # Render deployment blueprint
├── .env.example            # Environment variable template
├── Todo.md                 # Task tracking
├── README.md               # Main documentation
│
├── .github/
│   └── workflows/
│       ├── ci.yml          # CI: tests, lint, security
│       ├── deploy.yml      # Deploy: staging, production
│       └── pr-labels.yml   # Auto-label PRs
│
├── SETUP-GUIDE.md          # Manual setup for Google OAuth, Resend, Render
└── LICENSE                 # MIT License
```

### 2.1 `go.mod` & `go.sum`
* **Location**: `/go.mod`, `/go.sum`
* **Architectural Role**: Binds the compiler version and defines type-safe external package dependencies.
* **Key Packages**:
  * `github.com/gin-gonic/gin` — Main HTTP web framework
  * `github.com/gorilla/websocket` — High-performance WebSocket protocol engine
  * `firebase.google.com/go/v4` — Google Admin SDK for push notifications
  * `github.com/mattn/go-sqlite3` — CGO-based SQLite driver
  * `github.com/golang-jwt/jwt/v5` — JWT signing and verification

### 2.2 `main.go`
* **Location**: `/main.go`
* **Architectural Role**: The master system initialization entrypoint.
* **Execution Sequence**:
  1. Loads configurations from `.env` or system variables via `config.Load()`
  2. Instantiates the secure SQLite connection pool (`storage.InitDB()`) and executes migrations
  3. Initializes the Go Firebase Admin Client (`services/fcm.InitFCM()`)
  4. Spawns the concurrent WebSocket signaling hub goroutine (`go hub.ActiveHub.Run()`)
  5. Configures CORS, middleware interceptors, and registers GIN REST route controllers
  6. Binds and listens on the specified port

---

## 3. Storage & Configuration Modules

```
vyzorix-update-server/
├── config/
│   ├── config.go          # Environment variable parsing
│   └── config_test.go     # Config tests
└── storage/
    ├── sqlite.go          # SQLite connection pool + all CRUD methods
    └── sqlite_test.go     # Storage tests
```

### 3.1 `config/config.go`
* **Path**: `config/config.go`
* **Architectural Role**: Parses environmental variables into a strictly typed `Config` struct.
* **Key Fields**:
  * `Port` — Server listen port
  * `Env` — "production" or "development"
  * `DatabaseURL` — SQLite file path
  * `JWT_SECRET` — JWT signing secret
  * `TOKEN_SECRET` — Dashboard token secret
  * `ENFORCE_HMAC` — Require HMAC on device commands
  * `ALLOWED_ORIGINS` — CORS origins (comma-separated)
  * Google OAuth, Email (Resend), Firebase credentials

### 3.2 `storage/sqlite.go`
* **Path**: `storage/sqlite.go`
* **Architectural Role**: Manages the local SQLite3 connection pool with WAL mode for concurrent access.
* **Key Features**: Write-Ahead Logging (`PRAGMA journal_mode=WAL`), foreign key enforcement
* **Tables**: operators, sessions, devices, commands, settings

---

## 4. Security Modules (`security/`)

```
vyzorix-update-server/security/
├── jwt.go                   # JWT signing and verification
├── jwt_test.go              # JWT tests
├── hmac.go                  # HMAC-SHA256 command signing
├── hmac_test.go             # HMAC tests
├── google_token.go          # Google OAuth ID token verification (JWKS)
├── google_token_test.go     # Google token tests
├── password.go              # Password complexity validation
│   ├── DefaultPasswordPolicy  // Strict: 8+ chars, upper, lower, digit, special
│   └── UserPasswordPolicy    // User-friendly: 12+ chars, no special required
├── password_test.go         # Password validation tests
├── ratelimit.go             # In-memory rate limiting middleware
├── ratelimit_test.go        # Rate limiter tests
├── origin.go                 # WebSocket origin validation
├── origin_test.go           # Origin validation tests
└── secretstore/
    └── secretstore.go       # Key rotation and secret management
```

---

## 5. Data Models (`models/`)

```
vyzorix-update-server/models/
├── auth.go              # Operator, Session, login/register models
├── device.go           # Device, registration, status
├── command.go          # CommandFrame, CommandRequest
├── telemetry.go        # TelemetryFrame (from device)
├── updater.go           # VersionManifest, update state
├── response.go          # APIError, APIResponse
└── models.go            # Re-exports
```

### 5.1 `models/auth.go`
* **Key Structs**: `Operator` (id, email, name, role, createdAt), `Session` (id, operatorId, token, expiresAt, createdAt)

### 5.2 `models/device.go`
* **Key Structs**: `Device` (id, fcmToken, appVersion, deviceClass, commandSecret, secretHash, online, lastSeen, createdAt)

### 5.3 `models/command.go`
* **Key Structs**: `Command` (id, deviceId, command, args, status, dispatchId, deliveredAt, result)

### 5.4 `models/telemetry.go`
* **Key Structs**: `TelemetryFrame` (type, deviceId, uptime, riskScore, audioMode, speakerOn, activeDevice, bufferLevel, thermalTemp, timestamp)

---

## 6. Real-Time WebSocket Broker (`hub/`)

```
vyzorix-update-server/hub/
├── hub.go              # Client registry + broadcast goroutine
├── client.go           # readPump/writePump per connection
└── hub_test.go         # Hub tests
```

### 6.1 `hub/hub.go`
* **Architectural Role**: Binds client WebSocket connections
* **Hub Thread Loop**: Manages three active channels: `register`, `unregister`, `broadcast`
* **Thread Safety**: Uses mutex locks (`sync.RWMutex`) around the connection map

### 6.2 `hub/client.go`
* **Architectural Role**: Binds network sockets, wraps Gorilla's `*websocket.Conn`
* **Goroutines**: `readPump` (incoming telemetry), `writePump` (outgoing commands)
* **Heartbeats**: Ping/pong frames every 15 seconds

---

## 7. REST Controllers (`controllers/`)

```
vyzorix-update-server/controllers/
├── server.go                 # Health, version, SPA serving
├── auth.go                   # Login, register, logout, Google OAuth, email verification
├── auth_test.go              # Auth tests
├── device.go                 # Device registration, status, FCM
├── device_test.go            # Device tests
├── command.go                # Command dispatch (WS or FCM)
├── command_test.go           # Command tests
├── updater.go                # OTA version manifest endpoints
├── websocket_handler.go      # WebSocket upgrade handler with OriginValidator
└── websocket_handler_test.go # WebSocket tests
```

### 7.1 `controllers/auth.go` — Authentication Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/auth/login` | Operator login |
| `POST` | `/v1/auth/register` | Operator registration |
| `POST` | `/v1/auth/logout` | Logout (JWT required) |
| `GET` | `/v1/auth/me` | Get current operator (JWT required) |
| `PATCH` | `/v1/auth/me` | Update operator name (JWT required) |
| `GET` | `/v1/auth/google` | Initiate Google OAuth |
| `GET` | `/v1/auth/google/callback` | OAuth callback |
| `POST` | `/v1/auth/verify-email` | Email verification |
| `POST` | `/v1/auth/resend-verification` | Resend verification email |
| `POST` | `/v1/auth/forgot-password` | Request password reset |
| `POST` | `/v1/auth/reset-password` | Reset password with token |

### 7.2 `controllers/device.go` — Device Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/device/register` | Register device |
| `GET` | `/v1/device/:id/status` | Get device status |
| `PATCH` | `/v1/device/:id/fcm-token` | Update FCM token |
| `DELETE` | `/v1/device/:id` | Deregister device |
| `GET` | `/v1/dashboard/devices` | List all devices (JWT required) |

### 7.3 `controllers/command.go` — Command Dispatch
* **Operational Flow**:
  ```
  POST /v1/device/:id/command ──► Parse JSON ──► Check if target online?
                                                │
                ┌──────────────────────────────┴────────────────────────────┐
                ▼ (YES: Direct WS Route)              ▼ (NO: FCM Signaling)
     hub.ActiveHub.Send()                  services.fcm.SendSilentPush()
  ```

---

## 8. Firebase Push Messaging (`services/fcm/`)

```
vyzorix-update-server/services/fcm/
├── fcm.go               # Firebase Admin SDK init
├── notifier.go          # SafeNotifier wrapper for graceful degradation
├── notifier_test.go     # FCM notifier tests
└── errors.go            # ErrUnavailable and custom errors
```

### 8.1 `services/fcm/notifier.go` (SafeNotifier)
* **Architectural Role**: Silent push notifier with graceful degradation
* **Features**:
  * `SafeNotifier` wrapper catches all FCM errors
  * Logs warnings instead of propagating errors
  * Service continues operating if FCM is unavailable
* **Payload Schema**:
  ```go
  message := &messaging.Message{
      Token: targetToken,
      Android: &messaging.AndroidConfig{
          Priority: "high", // Guarantees wake-up bypassing Doze mode
      },
      Data: map[string]string{
          "action": "WAKE_DAEMON",
          "command": "FORCE_SPEAKER",
      },
  }
  ```

---

## 9. Middleware Interceptors (`middleware/`)

```
vyzorix-update-server/middleware/
├── auth.go               # Dashboard bearer token authentication
├── auth_test.go         # Auth middleware tests
├── cors.go              # CORS with configurable origins, MaxAge
├── cors_test.go         # CORS tests (strict origin validation)
├── logger.go            # Structured JSON request logging
├── logger_test.go       # Logger tests
├── rate_limiter.go      # Token-bucket rate limiting per IP
└── rate_limiter_test.go # Rate limiter tests
```

### 9.1 `middleware/cors.go`
* **Features**:
  * Access-Control-Allow-Credentials header
  * Access-Control-Max-Age header (1 hour preflight cache)
  * X-Request-ID to allowed headers
  * Rejects requests without Origin header for security
  * Never uses wildcard with credentials

---

## 10. Frontend Architecture (`src/`)

```
vyzorix-update-server/src/
├── main.tsx             # React entrypoint
├── start.ts             # TanStack Start adapter
├── server.ts            # SSR server
├── router.tsx           # React Router setup
├── routeTree.gen.ts     # Generated route tree
│
├── lib/                 # Utilities and API client
│   ├── vyzorix-api.ts  # Browser API client (fetch wrapper)
│   ├── vyzorix-auth.ts  # Auth client (login, register, OAuth)
│   ├── vyzorix-config.tsx # LocalStorage settings + VyzorixConfigProvider
│   ├── logger.ts        # App-wide log bus with persistence
│   ├── format.ts        # Utility functions
│   ├── utils.ts         # General utilities
│   └── settings.test.ts # Settings validation tests
│
├── hooks/               # Custom React hooks
│   ├── use-logs.ts      # Log fetching and display
│   ├── use-device-stream.ts # WebSocket with reconnect
│   ├── use-server-health.ts # Health polling
│   └── use-mobile.tsx   # Mobile detection
│
├── components/
│   ├── ui/              # Base UI components (shadcn/ui)
│   ├── layout/          # Layout components
│   ├── logs/            # LogDock, LogConsole components
│   └── status-badge.tsx  # Device health status
│
└── routes/              # TanStack file-based routes
    ├── __root.tsx      # Root layout
    ├── _app.tsx        # App layout (sidebar, header, log dock)
    ├── _app.dashboard.tsx # Main dashboard
    ├── _app.device.tsx  # Device page
    ├── _app.diagnostics.tsx # Command panel, charts
    ├── _app.alerts.tsx   # System alerts
    ├── _app.updates.tsx  # OTA version info
    ├── _app.logs.tsx     # Full page log viewer
    ├── _app.settings.*.tsx # All settings pages
    ├── login.tsx        # Login/register page
    └── auth.callback.tsx # OAuth callback
```

### Key Frontend Features
* **URL Validation**: Validates http:// or https:// protocol
* **Race Condition Fix**: Stable closure in operator name auto-save
* **Dynamic Device Names**: Uses formatDeviceClass() from API
* **Log Components**: LogDock and LogConsole for debugging

---

## 11. GitHub Workflows (`.github/workflows/`)

```
.github/workflows/
├── ci.yml           # Continuous Integration
│   ├── backend-test      # Go tests with coverage
│   ├── frontend-test     # Vitest tests with coverage
│   ├── backend-lint      # golangci-lint
│   ├── frontend-lint     # ESLint + TypeScript
│   ├── build            # Build Go + frontend
│   ├── security         # Gosec scanner
│   └── dependency-review # Dependency audit
│
├── deploy.yml        # Deployment
│   ├── build          # Build & package
│   ├── deploy-staging    # Deploy to staging (on push to main)
│   ├── deploy-production # Deploy to production (on tag v*)
│   └── notify-failure   # Notify on deployment failure
│
└── pr-labels.yml     # PR automation
    ├── label-pr          # Auto-label PRs (size, type)
    └── auto-assign       # Request reviews for fixes
```

---

## 12. Recent Changes (Phase 1.5)

### Bug Fixes Implemented

| # | Component | Description |
|---|-----------|-------------|
| #2 | Backend | Device online status integration via hub |
| #3 | Backend | Enforce HMAC dashboard setting |
| #4 | Backend | Command status implementation |
| #5 | Backend | command_secrets_hash column with bcrypt |
| #6 | Backend | HMAC window configuration (30 seconds) |
| #7 | Backend | WebSocket handler factory |
| #8 | Backend | Duplicate device registration cleanup |
| #9 | Backend | Rate limiting on public endpoints |
| #10 | Backend | Dashboard API pagination |
| #11 | Backend | Request ID in logs |
| #12 | Backend | Device list online filter |
| #13 | Backend | Health check with DB verification |
| #14 | Backend | FCM graceful degradation (SafeNotifier) |
| #15 | Backend | CORS security hardening |

### Frontend Bug Fixes

| # | Description |
|---|-------------|
| FE-1 | URL validation in connection settings |
| FE-2 | Race condition in operator auto-save |
| FE-3 | Dynamic device name in dashboard |
| FE-4 | Dynamic device name in device page |

---

## 13. Deployment Notes

### Render (Hobby Plan)
* Server sleeps after 15-30 minutes of inactivity
* Use UptimeRobot to ping `/healthz` every 10 minutes for keepalive
* `/data` persistent disk preserves SQLite across redeploys

### Environment Variables Required
```env
PORT=3000
DATABASE_URL=./data/vyzorix.db
JWT_SECRET=your-jwt-secret-min-32-chars
TOKEN_SECRET=your-dashboard-token-secret
ENFORCE_HMAC=false
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

---

*Last Updated: June 2026*  
*Phase: 1.5 (Production Ready)*