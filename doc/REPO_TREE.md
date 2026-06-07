# Vyzorix Update Server — Repository Tree

Complete file structure of the vyzorix-update-server monorepo. Updated to reflect current state including bug fixes #2-15 and frontend improvements.

---

## Root Level

```
vyzorix-update-server/
├── go.mod                     # Go module definition (1.22+)
├── go.sum                     # Dependency checksums (locked)
├── main.go                    # Server entrypoint
├── Makefile                   # Build, test, docker commands
├── Dockerfile                 # Multi-stage build (Go + React)
├── docker-compose.yml         # Local dev environment
├── render.yaml               # Render deployment blueprint
├── .env.example               # Environment variable template
├── .gitignore                 # Excludes: binaries, .db, node_modules
├── Todo.md                    # Task tracking for next features
│
├── .github/
│   └── workflows/
│       ├── ci.yml            # CI: tests, lint, security
│       ├── deploy.yml        # Deploy: staging, production
│       └── pr-labels.yml     # Auto-label PRs
│
├── SETUP-GUIDE.md             # Manual setup for Google OAuth, Resend, Render
└── README.md                  # Main documentation
```

---

## Backend — Go (`/cmd`, `/controllers`, `/services`, etc.)

```
vyzorix-update-server/
├── main.go                    # Server bootstrap: config → store → hub → gin
│
├── cmd/
│   └── mockserver/           # Phase 1 in-memory mock server
│       ├── main.go           # Standalone testing server
│       ├── server.go         # Mock server implementation
│       ├── device.go         # Mock device endpoints
│       ├── command.go        # Mock command handling
│       ├── update.go        # Mock OTA endpoints
│       ├── ws.go             # Mock WebSocket handler
│       ├── hmac.go           # HMAC verification for mock
│       ├── store.go         # In-memory device store
│       ├── server_test.go    # Mock server tests
│       └── hmac_test.go      # HMAC tests
│
├── config/
│   ├── config.go             # Environment variable parsing
│   └── config_test.go        # Config tests
│
├── storage/
│   ├── sqlite.go             # SQLite connection pool + all CRUD methods
│   └── sqlite_test.go        # Storage tests
│
├── security/
│   ├── jwt.go                # JWT signing and verification (session tokens)
│   ├── jwt_test.go           # JWT tests
│   ├── hmac.go               # HMAC-SHA256 command signing
│   ├── hmac_test.go          # HMAC tests
│   ├── google_token.go       # Google OAuth ID token verification (JWKS)
│   ├── google_token_test.go # Google token tests
│   ├── password.go           # Password complexity validation
│   │   ├── DefaultPasswordPolicy  # Strict (8+ chars, upper, lower, digit, special)
│   │   └── UserPasswordPolicy      # User-friendly (12+ chars, no special required)
│   ├── password_test.go     # Password validation tests
│   ├── ratelimit.go          # In-memory rate limiting middleware
│   ├── ratelimit_test.go     # Rate limiter tests
│   └── secretstore/          # Key rotation and secret management
│       └── secretstore.go    # Secret store implementation
│
├── hub/                       # WebSocket broker
│   ├── hub.go                # Client registry + broadcast goroutine
│   ├── client.go             # readPump/writePump per connection
│   └── hub_test.go           # Hub tests
│
├── controllers/               # Gin HTTP handlers
│   ├── server.go             # Health (with DB check), version, SPA serving
│   ├── auth.go               # Login, register, logout, Google OAuth, email verification
│   ├── auth_test.go          # Auth tests
│   ├── device.go             # Device registration, status, FCM token
│   ├── device_test.go        # Device tests
│   ├── command.go            # Command dispatch (WS or FCM)
│   ├── command_test.go       # Command tests
│   ├── updater.go            # OTA version manifest endpoints
│   ├── websocket_handler.go  # WebSocket upgrade handler with OriginValidator
│   └── websocket_handler_test.go # WebSocket tests
│
├── middleware/               # Gin middleware
│   ├── auth.go              # Dashboard bearer token authentication
│   ├── auth_test.go         # Auth middleware tests
│   ├── cors.go              # CORS with configurable origins, MaxAge
│   ├── cors_test.go         # CORS tests (strict origin validation)
│   ├── logger.go            # Structured JSON request logging
│   ├── logger_test.go        # Logger tests
│   ├── rate_limiter.go       # Token-bucket rate limiting per IP
│   └── rate_limiter_test.go # Rate limiter tests
│
├── models/                   # Shared types
│   ├── auth.go              # Operator, Session, login/register models
│   ├── device.go           # Device, registration, status
│   ├── command.go          # CommandFrame, CommandRequest
│   ├── telemetry.go         # TelemetryFrame (from device)
│   ├── updater.go           # VersionManifest, update state
│   ├── response.go          # APIError, APIResponse
│   └── models.go            # Re-exports
│
├── services/
│   ├── fcm/
│   │   ├── fcm.go          # Firebase Admin SDK init
│   │   ├── notifier.go     # SafeNotifier wrapper for graceful degradation
│   │   ├── notifier_test.go # FCM notifier tests
│   │   └── errors.go       # ErrUnavailable and custom errors
│   ├── email.go            # Resend email service (verification, reset)
│   ├── email_test.go       # Email service tests
│   ├── command_signer.go   # HMAC signing for device commands
│   └── command_signer_test.go
│
├── public/                   # Static assets
│   ├── index.html           # React SPA entry
│   ├── landing.html         # Native HTML landing page
│   ├── health.json          # Static health fallback
│   ├── favicon.ico
│   └── manifest.json
│
├── data/                     # Runtime data (Render persistent disk)
│   ├── version.json         # OTA manifest (served to devices)
│   └── changelog.json       # Release notes
│
├── api/v1/                   # Source-of-truth version manifests
│   ├── version.json
│   └── changelog.json
│
├── bin/                      # APK binaries (populated by CI on release)
│
└── scripts/                  # Build and deployment scripts
    ├── generate_version.sh  # Generate version.json
    ├── compute_checksum.sh  # APK checksum generation
    ├── validate_apk.sh      # APK validation
    └── cleanup_old_apks.sh  # Remove outdated APKs
```

---

## Frontend — React (TanStack Start, Vite)

```
vyzorix-update-server/
├── package.json             # Node dependencies + scripts
├── package-lock.json        # npm lockfile
├── vite.config.ts           # Vite bundler config
├── tsconfig.json            # TypeScript config
├── components.json          # Radix UI / Tailwind setup
├── eslint.config.js         # Linting rules
│
├── src/
│   ├── main.tsx             # React entrypoint
│   ├── start.ts             # TanStack Start adapter
│   ├── server.ts            # SSR server
│   ├── router.tsx           # React Router setup
│   ├── routeTree.gen.ts     # Generated route tree
│   │
│   ├── styles.css           # Global Tailwind styles
│   │
│   ├── lib/                 # Utilities and API client
│   │   ├── vyzorix-api.ts  # Browser API client (fetch wrapper)
│   │   │   ├── getHealth()           # GET /healthz
│   │   │   ├── getVersion()          # GET /api/v1/version
│   │   │   ├── headApk()             # HEAD /api/v1/apk/:filename
│   │   │   ├── getDeviceStatus()     # GET /v1/device/:id/status
│   │   │   ├── getDashboardDevices() # GET /v1/dashboard/devices
│   │   │   ├── registerDevice()      # POST /v1/device/register
│   │   │   └── dispatchCommand()     # POST /v1/device/:id/command
│   │   ├── vyzorix-auth.ts  # Auth client (login, register, OAuth)
│   │   │   ├── login()              # POST /v1/auth/login
│   │   │   ├── register()           # POST /v1/auth/register
│   │   │   ├── logout()             # POST /v1/auth/logout
│   │   │   ├── me()                 # GET /v1/auth/me
│   │   │   ├── updateName()         # PATCH /v1/auth/me
│   │   │   └── redirectToGoogleOAuth() # GET /v1/auth/google
│   │   ├── vyzorix-config.tsx # LocalStorage settings + VyzorixConfigProvider
│   │   ├── logger.ts         # App-wide log bus with persistence
│   │   ├── format.ts         # Utility functions (formatUptime, formatBytes, etc.)
│   │   ├── format.test.ts    # Format utility tests
│   │   ├── utils.ts          # General utilities
│   │   ├── utils.test.ts    # Utility tests
│   │   ├── settings.test.ts  # Settings validation tests
│   │   ├── error-page.ts     # Error boundary component
│   │   ├── error-capture.ts  # Error capture utilities
│   │   ├── config.server.ts  # Server-side config
│   │   ├── device-stream-context.tsx # WebSocket context provider
│   │   └── admin.functions.ts # Admin function utilities
│   │
│   ├── hooks/               # Custom React hooks
│   │   ├── use-logs.ts      # Log fetching and display (useSyncExternalStore)
│   │   ├── use-device-stream.ts # WebSocket with reconnect (autoReconnect)
│   │   ├── use-server-health.ts # Health polling (useQuery)
│   │   └── use-mobile.tsx   # Mobile detection hook
│   │
│   ├── components/
│   │   ├── ui/              # Base UI components (shadcn/ui)
│   │   │   ├── button.tsx, card.tsx, badge.tsx, ...
│   │   │   └── [all shadcn components]
│   │   ├── layout/
│   │   │   └── footer.tsx  # App footer
│   │   ├── loading/
│   │   │   └── page-skeleton.tsx # Loading skeleton
│   │   ├── logs/           # Log display components
│   │   │   ├── log-dock.tsx  # Docked log panel (collapsible)
│   │   │   └── log-console.tsx # Full log viewer with filtering
│   │   ├── app-sidebar.tsx # Main app sidebar
│   │   ├── connection-badge.tsx # WebSocket connection status
│   │   └── status-badge.tsx # Device health status badge
│   │
│   └── routes/              # TanStack file-based routes
│       ├── __root.tsx      # Root layout
│       ├── _app.tsx        # App layout (sidebar, header, log dock)
│       ├── _app.index.tsx   # Root redirect → /dashboard
│       ├── _app.dashboard.tsx # Main dashboard with device info
│       ├── _app.device.tsx  # Device page with registration
│       ├── _app.diagnostics.tsx # Command panel, charts
│       ├── _app.alerts.tsx   # System alerts derivation
│       ├── _app.updates.tsx  # OTA version info
│       ├── _app.logs.tsx     # Full page log viewer
│       ├── _app.settings.tsx # Settings layout
│       ├── _app.settings.index.tsx # Settings overview
│       ├── _app.settings.connection.tsx # Server URL, device ID, transport
│       ├── _app.settings.operator.tsx # Operator profile, notifications
│       ├── _app.settings.thresholds.tsx # Alert thresholds
│       ├── _app.settings.notifications.tsx # Toast, browser notifications
│       ├── _app.settings.appearance.tsx # Theme (system/light/dark)
│       ├── _app.settings.advanced.tsx # Buffer limits, reset
│       ├── login.tsx        # Login/register page
│       └── auth.callback.tsx # OAuth callback handler
│
├── dist/                    # Built frontend assets
│   ├── client/             # Static assets
│   └── server/             # SSR bundle
│
└── node_modules/            # npm packages
```

---

## GitHub Workflows (`.github/workflows/`)

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

## Documentation (`/doc`)

```
vyzorix-update-server/doc/
├── README.md                        # Doc index
├── REPO_TREE.md                     # This file
├── SYSTEM_MAP.md                    # System overview
├── FRONTEND_BUG_FIXES.md            # Frontend bug fixes (FE-1 to FE-4)
├── BACKEND_BUG_FIXES.md             # Backend bug fixes (#2-15)
├── UPDATE_SERVER.md                 # Server endpoints reference
├── UPDATE_SERVER_ARCHITECTURE_SPEC.md  # Deep-dive architecture
├── DEVICE_REGISTRATION.md           # Device registration flow
├── COMMAND_SECURITY.md              # HMAC command signing
├── FEATURES.md                     # Feature list
├── BUILD_ORDER.md                  # Build sequence
├── CI_CD_WORKFLOWS.md              # CI/CD documentation
├── NAMING_RENAMES.md               # Naming conventions
├── GLOSSARY.md                     # Terminology
├── ADR/                            # Architecture Decision Records
│   └── README.md
└── [additional project docs]
```

---

## Key Technologies

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.22+, Gin web framework |
| **Database** | SQLite (WAL mode, persistent on Render) |
| **WebSocket** | gorilla/websocket |
| **Push Notifications** | Firebase Cloud Messaging (FCM) |
| **Auth** | JWT + Google OAuth |
| **Email** | Resend API |
| **Frontend** | React 18, TanStack Start, Vite |
| **Styling** | Tailwind CSS, Radix UI/shadcn-ui |
| **Testing** | Vitest (frontend), Go testing (backend) |
| **Linting** | ESLint, golangci-lint |
| **Security** | Gosec, Dependency Review |
| **Deployment** | Render (with persistent disk) |

---

## Package Summary

| Directory | Purpose |
|-----------|---------|
| `cmd/mockserver/` | Mock server for Phase 1 testing |
| `config/` | Environment configuration |
| `controllers/` | HTTP request handlers |
| `hub/` | WebSocket broker |
| `middleware/` | HTTP middleware (auth, CORS, logging, rate limit) |
| `models/` | Type definitions |
| `security/` | JWT, HMAC, Google OAuth, password validation |
| `services/fcm/` | Firebase Cloud Messaging with SafeNotifier |
| `services/` | Email, command signing |
| `storage/` | SQLite database operations |
| `src/` | React frontend |
| `src/lib/` | API clients, config, logger |
| `src/hooks/` | Custom React hooks |
| `src/routes/` | TanStack file-based routes |
| `src/components/` | React components |
| `doc/` | Architecture documentation |
| `scripts/` | Build automation |

---

## Recent Changes (Session Summary)

### Bug Fixes Implemented

| # | Component | Description |
|---|-----------|-------------|
| #2-5 | Backend | Initial bug fixes |
| #7-12 | Backend | Additional fixes |
| #13 | Backend | Enhanced health check with DB verification |
| #14 | Backend | FCM graceful degradation (SafeNotifier) |
| #15 | Backend | CORS security hardening |
| FE-1 | Frontend | URL validation in connection settings |
| FE-2 | Frontend | Race condition fix in operator auto-save |
| FE-3 | Frontend | Dynamic device name in dashboard |
| FE-4 | Frontend | Dynamic device name in device page |

### Tests Added

| Test Suite | Count | Status |
|------------|-------|--------|
| Go Tests | 12 packages | [OK] All passing |
| Vitest Tests | 79 tests | [OK] All passing |
| Build | - | [OK] Successful |