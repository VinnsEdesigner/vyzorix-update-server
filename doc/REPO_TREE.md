# Vyzorix Update Server — Repository Tree

Complete file structure of the vyzorix-update-server monorepo.

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
│
├── SETUP-GUIDE.md             # Manual setup for Google OAuth, Resend, Render
└── README.md                  # This repository's main documentation
```

---

## Backend — Go (`/cmd`, `/controllers`, `/services`, etc.)

```
vyzorix-update-server/
├── main.go                    # Server bootstrap: config → store → hub → gin
│
├── cmd/
│   └── mockserver/            # Phase 1 in-memory mock server
│       └── main.go            # Standalone testing server
│
├── config/
│   ├── config.go              # Environment variable parsing
│   └── config_test.go         # Config tests
│
├── storage/
│   ├── sqlite.go             # SQLite connection pool + all CRUD methods
│   ├── sqlite_test.go        # Storage tests
│
├── security/
│   ├── jwt.go                 # JWT signing and verification (session tokens)
│   ├── jwt_test.go            # JWT tests
│   ├── hmac.go                # HMAC-SHA256 command signing
│   ├── hmac_test.go           # HMAC tests
│   ├── google_token.go        # Google OAuth ID token verification (JWKS)
│   ├── google_token_test.go   # Google token tests
│   ├── password.go            # Password complexity validation
│   ├── password_test.go      # Password validation tests
│   ├── ratelimit.go           # In-memory rate limiting middleware
│   ├── ratelimit_test.go      # Rate limiter tests
│   └── secretstore/           # Key rotation and secret management
│       └── ...
│
├── hub/                       # WebSocket broker
│   ├── hub.go                 # Client registry + broadcast goroutine
│   ├── client.go              # readPump/writePump per connection
│   └── hub_test.go            # Hub tests
│
├── controllers/                # Gin HTTP handlers
│   ├── server.go              # Health, version, SPA serving
│   ├── auth.go                # Login, register, logout, Google OAuth
│   ├── auth_test.go           # Auth tests
│   ├── device.go              # Device registration, status, FCM
│   ├── device_test.go         # Device tests
│   ├── command.go             # Command dispatch (WS or FCM)
│   ├── command_test.go        # Command tests
│   ├── updater.go             # OTA version manifest endpoints
│   └── websocket_handler.go   # WebSocket upgrade handler
│
├── middleware/                 # Gin middleware
│   ├── auth.go               # Dashboard bearer token authentication
│   ├── auth_test.go          # Auth middleware tests
│   ├── cors.go               # CORS with configurable origins
│   ├── cors_test.go          # CORS tests
│   ├── logger.go             # Structured JSON request logging
│   ├── logger_test.go       # Logger tests
│   ├── rate_limiter.go       # Token-bucket rate limiting per IP
│   └── rate_limiter_test.go  # Rate limiter tests
│
├── models/                    # Shared types
│   ├── auth.go               # Operator, Session, login/register models
│   ├── device.go             # Device, registration, status
│   ├── command.go            # CommandFrame, CommandRequest
│   ├── telemetry.go          # TelemetryFrame (from device)
│   ├── updater.go            # VersionManifest, update state
│   ├── response.go           # APIError, APIResponse
│   └── models.go             # Re-exports
│
├── services/
│   ├── fcm/                  # Firebase Cloud Messaging
│   │   ├── fcm.go           # Firebase Admin SDK init
│   │   └── notifier.go      # Silent high-priority wake payloads
│   ├── email.go             # Resend email service (verification, reset)
│   ├── email_test.go        # Email service tests
│   ├── command_signer.go    # HMAC signing for device commands
│   └── command_signer_test.go
│
├── middleware/               # (listed above, duplicated for reference)
│
├── public/                   # Static assets
│   ├── index.html            # React SPA entry
│   ├── landing.html           # Native HTML landing page
│   ├── health.json            # Static health fallback
│   ├── favicon.ico
│   ├── manifest.json
│   └── style.css
│
├── data/                     # Runtime data (Render persistent disk)
│   ├── version.json          # OTA manifest (served to devices)
│   └── changelog.json         # Release notes
│
├── api/v1/                   # Source-of-truth version manifests
│   ├── version.json
│   └── changelog.json
│
├── bin/                      # APK binaries (populated by CI on release)
│
├── scripts/                  # Build and deployment scripts
│   ├── generate_version.sh  # Generate version.json
│   ├── compute_checksum.sh   # APK checksum generation
│   ├── validate_apk.sh       # APK validation
│   └── cleanup_old_apks.sh   # Remove outdated APKs
│
└── test/                     # Integration tests
    └── README.md
```

---

## Frontend — React (TanStack Start, Vite)

```
vyzorix-update-server/
├── package.json             # Node dependencies + scripts
├── bun.lock                  # Bun lockfile
├── bunfig.toml              # Bun configuration
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
│   ├── routeTree.gen.ts    # Generated route tree
│   │
│   ├── styles.css          # Global Tailwind styles
│   │
│   ├── lib/                # Utilities and API client
│   │   ├── vyzorix-api.ts  # Browser API client (fetch wrapper)
│   │   └── vyzorix-config.tsx # LocalStorage settings
│   │
│   ├── hooks/               # Custom React hooks
│   │   ├── use-logs.ts      # Log fetching and display
│   │   ├── use-device-stream.ts  # WebSocket with reconnect
│   │   └── use-server-health.ts  # Health polling
│   │
│   ├── integrations/        # Third-party integrations
│   │
│   ├── components/          # React components
│   │   ├── layout/         # Layout components (Sidebar, Navbar, Footer)
│   │   ├── ui/             # Base UI components (Button, Card, Badge, etc.)
│   │   ├── dashboard/      # Dashboard-specific components
│   │   ├── device/        # Device management components
│   │   ├── logs/          # Log display components
│   │   │   ├── LogConsole.tsx
│   │   │   └── LogDock.tsx
│   │   └── charts/        # Telemetry charts
│   │
│   └── routes/             # TanStack file-based routes
│       ├── index.tsx       # Root route
│       ├── auth.tsx        # Auth pages
│       ├── login.tsx       # Login page
│       ├── dashboard.tsx   # Dashboard layout
│       ├── devices.tsx     # Devices page
│       └── settings.tsx    # Settings page
│
├── dist/                    # Built frontend assets
│   ├── client/             # Static assets
│   └── server/             # SSR bundle
│
└── node_modules/            # npm packages
```

---

## Documentation (`/doc`)

```
vyzorix-update-server/doc/
├── README.md               # Doc index
├── SYSTEM_MAP.md           # System overview
├── UPDATE_SERVER.md        # Server endpoints reference
├── UPDATE_SERVER_ARCHITECTURE_SPEC.md  # Deep-dive architecture
├── DOC_1_BOOTSTRAP_AND_ORCHESTRATION.md
├── DOC_2_ACCESSIBILITY_AND_AUTOMATION_GOVERNANCE.md
├── DOC_3_AUDIO_PIPELINE_AND_VOIP_EXEMPTIONS.md
├── DOC_4_RESILIENCE_FALLBACKS_AND_RECOVERY.md
├── DOC_5_DIAGNOSTICS_CRASH_FORENSICS_AND_STORAGE.md
├── DOC_6_MEMORY_PERFORMANCE_AND_HARDWARE_MONITORING.md
├── DOC_7_DATA_SECURITY_AND_PERSISTENCE.md
├── DOC_8_REALTIME_C2_COMMUNICATION_AND_UPDATES.md
├── DOC_8_REALTIME_C2_COMMUNICATION_AND_UPDATES_UPDATED.md
├── DEVICE_REGISTRATION.md   # Device registration flow
├── COMMAND_SECURITY.md      # HMAC command signing
├── FEATURES.md             # Feature list
├── FEATURES_UPDATED.md
├── BUILD_ORDER.md          # Build sequence
├── CI_CD_WORKFLOWS.md       # CI/CD documentation
├── NAMING_RENAMES.md        # Naming conventions
├── GLOSSARY.md             # Terminology
├── ADR/                    # Architecture Decision Records
└── [various project docs]
```

---

## Configuration Files

```
vyzorix-update-server/
├── .env.example           # Environment variable template
├── Dockerfile              # Multi-stage Docker build
├── docker-compose.yml      # Local dev stack
├── render.yaml             # Render deployment
├── Makefile                # Build commands
├── go.mod / go.sum         # Go dependencies
├── package.json            # Node dependencies
├── tsconfig.json           # TypeScript config
├── vite.config.ts          # Vite bundler
├── components.json         # Radix UI setup
├── eslint.config.js        # Linting
└── supabase/               # (deprecated - no longer used)
    ├── config.toml
    └── migrations/
```

---

## Data Directories

```
vyzorix-update-server/
├── data/                   # Runtime data (persistent on Render)
│   ├── version.json         # OTA manifest
│   ├── changelog.json       # Release notes
│   └── vyzorix.db           # SQLite database (created at runtime)
│
├── api/v1/                 # Version manifest source
│   ├── version.json
│   └── changelog.json
│
├── bin/                    # APK storage (CI populates)
│   └── *.apk
│
├── public/                 # Static frontend assets
│   ├── index.html
│   ├── landing.html
│   └── ...
│
└── dist/                   # Built frontend
    ├── client/
    └── server/
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
| **Styling** | Tailwind CSS, Radix UI |
| **Deployment** | Render (with persistent disk) |

---

## Package Summary

| Directory | Purpose |
|-----------|---------|
| `cmd/` | Entry points (mockserver, main) |
| `config/` | Environment configuration |
| `controllers/` | HTTP request handlers |
| `hub/` | WebSocket broker |
| `middleware/` | HTTP middleware (auth, CORS, logging, rate limit) |
| `models/` | Type definitions |
| `security/` | JWT, HMAC, Google OAuth, password validation |
| `services/` | FCM, email, command signing |
| `storage/` | SQLite database operations |
| `src/` | React frontend |
| `doc/` | Architecture documentation |
| `public/` | Static HTML/CSS |
| `scripts/` | Build automation |