# Vyzorix Enterprise Monorepo Structure

> **Document Version:** 2.0  
> **Status:** Proposed - Pending Implementation  
> **Last Updated:** 2026-06-08  
> **Author:** VinnsEdesigner + OpenHands

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Directory Structure](#3-directory-structure)
4. [Complete File Inventory](#4-complete-file-inventory)
5. [Current → Target Mapping](#5-current--target-mapping)
6. [Package Responsibilities](#6-package-responsibilities)
7. [Build System](#7-build-system)
8. [Dependency Graph](#8-dependency-graph)
9. [Naming Conventions](#9-naming-conventions)
10. [Git Strategy](#10-git-strategy)

---

## 1. Overview

### 1.1 Purpose

This document defines the **target enterprise monorepo structure** for the Vyzorix Update Server project. It serves as the **source of truth** before migration begins.

### 1.2 Tech Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Frontend** | React 19, TanStack Start, Vite | SSR web dashboard |
| **Backend** | Go 1.22+, Gin | REST API, WebSocket hub |
| **Database** | SQLite (WAL Mode) | Persistent storage |
| **Styling** | Tailwind CSS, shadcn/ui | Component library |
| **Build** | Turborepo, pnpm | Monorepo tooling |
| **E2E Testing** | Playwright | End-to-end tests |
| **CI/CD** | GitHub Actions | Deployment automation |
| **Security** | Cloudflare (Free) | WAF, DDoS, Turnstile |

### 1.3 Design Principles

1. **Defense in Depth** - Multiple security layers
2. **Zero Trust** - Every request authenticated and authorized
3. **Fail Securely** - Panic recovery, no stack trace leaks
4. **Observable** - Structured logging, audit trails
5. **Single Responsibility** - Each package has one job

---

## 2. Architecture

### 2.1 High-Level System Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INTERNET                                        │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CLOUDFLARE (Free Tier)                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐   │
│  │     WAF     │  │  DDoS       │  │  Turnstile  │  │  SSL/TLS        │   │
│  │  (OWASP)    │  │  Protection │  │  Bot Block   │  │  (Auto-Renew)   │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              VYZORIX MONOREPO                                 │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         apps/web (React SSR)                         │   │
│  │   TanStack Start + Vite + Tailwind + shadcn/ui                       │   │
│  │   Serves dashboard at /*, landing at /                             │   │
│  └─────────────────────────────────────┬───────────────────────────────┘   │
│                                        │ Proxy /v1/*, /api/* to apps/api   │
│  ┌─────────────────────────────────────▼───────────────────────────────┐   │
│  │                         apps/api (Go Backend)                       │   │
│  │   Gin HTTP Router + WebSocket Hub + SQLite                           │   │
│  │   ┌──────────────┐ ┌──────────────┐ ┌──────────────┐               │   │
│  │   │ REST API     │ │ WebSocket    │ │ FCM          │               │   │
│  │   │ /v1/auth/*   │ │ /v1/device/* │ │ Notifier     │               │   │
│  │   │ /v1/device/* │ │ /stream      │ │              │               │   │
│  │   └──────────────┘ └──────────────┘ └──────────────┘               │   │
│  │   ┌──────────────────────────────────────────────────────┐          │   │
│  │   │              Security Middleware Stack              │          │   │
│  │   │  RateLimit → CORS → SecurityHeaders → PanicRecover │          │   │
│  │   │  CSRF → JWT Auth → HMAC Verify → DOA Check         │          │   │
│  │   └──────────────────────────────────────────────────────┘          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         packages/ (Shared Libraries)                  │   │
│  │   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────────────┐   │   │
│  │   │  types  │  │   ui    │  │  config │  │    api-client       │   │   │
│  │   │(TS def) │  │(components)│ │(ESLint)│  │    (Go SDK)        │   │   │
│  │   └─────────┘  └─────────┘  └─────────┘  └─────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              PERSISTENT DISK                                 │
│  ┌─────────────────────────┐  ┌─────────────────────────────────────────┐ │
│  │  SQLite Database         │  │  APK/Binary Storage                     │ │
│  │  (WAL Mode, Encrypted)   │  │  (version.json, *.apk, *.bin)          │ │
│  │  ./data/vyzorix.db       │  │  ./bin/                                  │ │
│  └─────────────────────────┘  └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Security Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SECURITY LAYERS (Defense in Depth)                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Layer 1: Network Perimeter                                                 │
│  └─ Cloudflare WAF + DDoS + Turnstile                                       │
│                                                                              │
│  Layer 2: Transport Security                                                 │
│  └─ HTTPS Only (TLS 1.3), HSTS Header                                       │
│                                                                              │
│  Layer 3: Application Firewall                                              │
│  └─ Rate Limiting (IP + Session based)                                      │
│  └─ CORS Origin Whitelist                                                   │
│  └─ Security Headers (X-Frame-Options, CSP, etc.)                           │
│                                                                              │
│  Layer 4: Request Validation                                                 │
│  └─ MaxBytesReader (1MB limit)                                               │
│  └─ Input Sanitization                                                       │
│  └─ HMAC Signature Verification                                              │
│                                                                              │
│  Layer 5: Authentication & Authorization                                     │
│  └─ JWT Validation (HttpOnly cookies)                                        │
│  └─ CSRF Token (Synchronizer pattern)                                        │
│  └─ Deep Object Authorization (DOA)                                          │
│                                                                              │
│  Layer 6: Cryptographic Operations                                           │
│  └─ Argon2id Password Hashing (64MB, 1 iteration)                            │
│  └─ Ed25519 Command Signing                                                  │
│  └─ Token Revocation List                                                     │
│                                                                              │
│  Layer 7: Audit & Monitoring                                                 │
│  └─ Structured JSON Logging                                                  │
│  └─ Security Event Tracking                                                  │
│  └─ Error Masking (no stack traces)                                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Directory Structure

```
vyzorix/                                # Root: Monorepo workspace root
│
├── .github/                            # GitHub configuration
│   ├── workflows/                      # CI/CD pipelines
│   │   ├── ci.yml                     # Lint → Test → Type-check
│   │   ├── deploy-api.yml            # Build + Deploy Go backend
│   │   └── deploy-web.yml            # Build + Deploy React app
│   ├── CODEOWNERS                    # Auto-assign PR reviewers
│   └── PULL_REQUEST_TEMPLATE.md      # PR description template
│
├── .husky/                            # Git hooks
│   ├── pre-commit                     # Runs lint-staged on staged files
│   └── commit-msg                     # Validates conventional commit format
│
├── .vscode/                           # VS Code workspace settings
│   ├── settings.json                  # Editor config
│   └── extensions.json                # Recommended extensions
│
├── apps/                              # Deployable applications
│   ├── web/                          # React frontend (TanStack Start + Vite)
│   └── api/                          # Go backend (Gin HTTP server)
│
├── packages/                          # Shared libraries (versioned independently)
│   ├── ui/                          # shadcn/ui component library
│   ├── types/                       # TypeScript type definitions
│   ├── config/                      # Shared ESLint, TS, Tailwind configs
│   └── api-client/                  # Go HTTP client (for external integrations)
│
├── tooling/                          # Build scripts and utilities
│   ├── scripts/                     # DevOps automation scripts
│   └── docker/                      # Container definitions
│
├── docs/                             # Project documentation
│   ├── SECURITY/                    # Security documentation
│   │   ├── IMPLEMENTATION_PLAN.md  # Security implementation roadmap
│   │   ├── THREAT_MODEL.md         # Risk assessment
│   │   ├── AUTH.md                 # Authentication pipeline
│   │   ├── DEFENSE.md              # Defense matrix
│   │   ├── FUNNEL.md               # Request funnel
│   │   └── MATRIX.md               # API protection
│   ├── ARCHITECTURE.md             # System design
│   └── DEPLOYMENT.md               # Deployment guide
│
├── turbo.json                        # Turborepo pipeline configuration
├── pnpm-workspace.yaml               # pnpm workspaces manifest
├── package.json                      # Root workspace manifest
├── go.mod                           # Root Go module (for tools)
├── render.yaml                      # Render deployment blueprint
├── .env.example                     # All environment variables
└── README.md                        # Project overview
```

---

## 4. Complete File Inventory

### 4.1 Root Level Files

| File | Purpose | Migration Note |
|------|---------|----------------|
| `.github/workflows/ci.yml` | CI Pipeline | New structure |
| `.github/workflows/deploy-api.yml` | API Deploy | New structure |
| `.github/workflows/deploy-web.yml` | Web Deploy | New structure |
| `.github/CODEOWNERS` | Code ownership | NEW |
| `.github/PULL_REQUEST_TEMPLATE.md` | PR template | NEW |
| `.husky/pre-commit` | Pre-commit hook | NEW |
| `.husky/commit-msg` | Commit validator | NEW |
| `.vscode/settings.json` | Editor config | NEW |
| `.vscode/extensions.json` | VS Code recs | NEW |
| `turbo.json` | Build pipeline | NEW |
| `pnpm-workspace.yaml` | Workspace manifest | NEW (replaces root package.json workspaces) |
| `package.json` | Root workspace | UPDATE (remove app-specific scripts) |
| `go.mod` | Root Go module | UPDATE (add tooling deps) |
| `render.yaml` | Render blueprint | UPDATE |
| `.env.example` | Env vars | UPDATE |
| `.gitignore` | Git exclusions | UPDATE |
| `LICENSE` | MIT License | EXISTING |
| `README.md` | Project overview | UPDATE |

### 4.2 apps/web/ (Frontend Application)

```
apps/web/                         # React SSR Dashboard
│                                 # Responsible for: UI, routing, API calls, state
│
├── package.json                  # @vyzorix/web - workspace dependency
│
├── tsconfig.json                # TypeScript (extends @vyzorix/tsconfig)
│
├── vite.config.ts               # Vite + TanStack Start + proxy to apps/api
│
├── eslint.config.js             # ESLint (extends @vyzorix/eslint-web)
│
├── tailwind.config.ts           # Tailwind CSS config
│
├── components.json              # shadcn/ui CLI config
│
├── src/
│   ├── main.tsx               # Client hydration entry point
│   ├── start.ts               # TanStack Start SSR entry
│   ├── server.ts              # SSR error wrapper
│   ├── router.tsx             # TanStack Router config
│   ├── routeTree.gen.ts       # Auto-generated (DO NOT EDIT)
│   │
│   ├── routes/                # File-based routing
│   │   ├── __root.tsx        # Root layout
│   │   ├── _app.tsx          # App shell (sidebar + header)
│   │   ├── _app.index.tsx    # Redirect to /dashboard
│   │   ├── _app.dashboard.tsx
│   │   ├── _app.device.tsx
│   │   ├── _app.diagnostics.tsx
│   │   ├── _app.alerts.tsx
│   │   ├── _app.updates.tsx
│   │   ├── _app.logs.tsx
│   │   ├── _app.settings.tsx
│   │   ├── _app.settings.index.tsx
│   │   ├── _app.settings.connection.tsx
│   │   ├── _app.settings.operator.tsx
│   │   ├── _app.settings.thresholds.tsx
│   │   ├── _app.settings.notifications.tsx
│   │   ├── _app.settings.appearance.tsx
│   │   ├── _app.settings.advanced.tsx
│   │   ├── login.tsx
│   │   ├── auth.callback.tsx
│   │   ├── forgot-password.tsx
│   │   ├── reset-password.tsx
│   │   └── verify-email.tsx
│   │
│   ├── components/           # Page-specific components
│   │   ├── dashboard/
│   │   │   ├── device-list.tsx
│   │   │   ├── device-card.tsx
│   │   │   ├── stats-panel.tsx
│   │   │   └── quick-actions.tsx
│   │   ├── diagnostics/
│   │   │   ├── command-panel.tsx
│   │   │   └── command-history.tsx
│   │   └── settings/
│   │       ├── threshold-slider.tsx
│   │       └── connection-form.tsx
│   │
│   ├── hooks/               # Custom React hooks
│   │   ├── use-auth.ts
│   │   ├── use-device.ts
│   │   ├── use-websocket.ts
│   │   ├── use-device-stream.ts
│   │   └── use-vyzorix-config.ts
│   │
│   ├── lib/                 # Utility modules
│   │   ├── api/
│   │   │   ├── index.ts     # API client functions
│   │   │   ├── device.ts    # Device API calls
│   │   │   └── auth.ts      # Auth API calls
│   │   ├── config.tsx       # Vyzorix config context + provider
│   │   ├── utils.ts         # cn() helper, formatters
│   │   ├── logger.ts        # Structured logging client
│   │   ├── error-page.tsx   # 500 error page component
│   │   ├── error-capture.ts  # SSR error capture utility
│   │   └── integrations/
│   │       └── supabase/
│   │           └── auth-attacher.ts  # Supabase auth middleware
│   │
│   └── styles.css           # Global styles + Tailwind
│
├── public/
│   ├── landing.html         # Static landing page (served by Go backend)
│   ├── index.html          # Static fallback (served by Go backend)
│   ├── favicon.ico
│   ├── manifest.json       # PWA manifest
│   └── assets/
│       └── .gitkeep
│
└── tests/
    └── e2e/                 # Playwright E2E tests
        ├── playwright.config.ts  # Playwright config
        ├── login.spec.ts         # Login flow tests
        ├── dashboard.spec.ts     # Dashboard tests
        └── device.spec.ts       # Device management tests
```

### 4.3 apps/api/ (Go Backend)

```
apps/api/                       # Go Backend
│                               # Responsible for: HTTP, auth, device management, WebSocket
│
├── go.mod                     # module github.com/VinnsEdesigner/vyzorix/apps/api
│
├── main.go                    # Bootstrap: SQLite init, FCM init, server start
│
├── cmd/
│   └── server/
│       └── main.go           # Alternative entry point (for docker)
│
├── internal/                  # Private application code (not importable)
│   │
│   ├── api/                  # HTTP layer
│   │   ├── router.go        # Gin router setup + middleware registration
│   │   │
│   │   ├── handlers/        # HTTP request handlers
│   │   │   ├── auth.go      # Login, register, logout, me, OAuth
│   │   │   ├── auth_test.go
│   │   │   ├── device.go    # Register, status, fcm-token, delete
│   │   │   ├── device_test.go
│   │   │   ├── command.go   # Command dispatch
│   │   │   ├── command_test.go
│   │   │   ├── health.go    # Health check
│   │   │   ├── health_test.go
│   │   │   └── rate_limit_test.go
│   │   │
│   │   └── middleware/      # HTTP middleware
│   │       ├── security.go  # CORS, security headers, panic recovery
│   │       ├── ratelimit.go # IP + session rate limiting
│   │       ├── auth.go     # JWT validation
│   │       ├── csrf.go     # CSRF token validation
│   │       ├── body_size.go # MaxBytesReader wrapper
│   │       ├── logger.go   # Request logging
│   │       └── request_id.go # X-Request-ID injection
│   │
│   ├── auth/                 # Authentication logic
│   │   ├── jwt.go           # JWT generation + validation
│   │   ├── jwt_test.go
│   │   ├── password.go     # Argon2id hashing
│   │   ├── password_test.go
│   │   ├── session.go      # Session management
│   │   ├── revocation.go    # Token blacklist (logout)
│   │   └── csrf.go         # CSRF token generation
│   │
│   ├── device/              # Device management
│   │   ├── service.go       # Device business logic
│   │   ├── repository.go    # SQLite operations
│   │   └── models.go        # Device types
│   │
│   ├── ws/                  # WebSocket hub
│   │   ├── hub.go          # WebSocket connection manager
│   │   ├── hub_test.go
│   │   └── client.go      # Individual WebSocket client
│   │
│   └── fcm/                 # Firebase Cloud Messaging
│       └── notifier.go     # FCM notification sender
│
├── pkg/                      # Public libraries (can be imported)
│   │
│   ├── models/              # Shared data models
│   │   ├── device.go       # Device struct + JSON tags
│   │   ├── command.go      # Command structs
│   │   ├── auth.go         # Auth request/response types
│   │   ├── response.go     # Standard API response wrapper
│   │   ├── telemetry.go    # Telemetry frame types
│   │   └── updater.go      # OTA update types
│   │
│   ├── crypto/             # Cryptographic utilities
│   │   ├── hmac.go        # HMAC-SHA256 verification
│   │   ├── hmac_test.go
│   │   └── signature.go   # Ed25519 command signing
│   │
│   ├── storage/            # Database layer
│   │   ├── sqlite.go      # SQLite connection + WAL config
│   │   ├── sqlite_test.go
│   │   └── migrations/     # SQL migration files
│   │       ├── 001_initial.sql
│   │       └── 002_uuidv7.sql
│   │
│   └── config/             # Configuration
│       ├── config.go       # Config struct + Load()
│       └── config_test.go
│
├── scripts/                 # Build/maintenance scripts
│   ├── migrate.sh          # Run SQL migrations
│   ├── seed.sh             # Seed database with test data
│   └── cleanup_old_apks.sh # Remove stale APK files
│
├── Dockerfile               # Multi-stage Docker build
│
└── docker-compose.yml      # Local development services
```

### 4.4 packages/ui/ (Shared React Components)

```
packages/ui/                    # shadcn/ui component library
│                              # Responsible for: Reusable UI primitives
│
├── package.json              # @vyzorix/ui - published to npm (internal only)
│
├── tsconfig.json            # TypeScript (extends @vyzorix/tsconfig)
│
├── src/
│   ├── index.ts            # Exports all components
│   │
│   ├── components/
│   │   └── ui/             # shadcn/ui base components
│   │       ├── button.tsx
│   │       ├── card.tsx
│   │       ├── input.tsx
│   │       ├── label.tsx
│   │       ├── select.tsx
│   │       ├── switch.tsx
│   │       ├── tabs.tsx
│   │       ├── dialog.tsx
│   │       ├── dropdown-menu.tsx
│   │       ├── separator.tsx
│   │       ├── progress.tsx
│   │       ├── slider.tsx
│   │       ├── tooltip.tsx
│   │       ├── toast.tsx
│   │       ├── sonner.tsx
│   │       └── ... (other shadcn components)
│   │
│   └── lib/
│       └── utils.ts        # cn() classname utility
│
└── README.md               # Usage instructions
```

### 4.5 packages/types/ (TypeScript Type Definitions)

```
packages/types/                # Shared TypeScript types
│                              # Responsible for: Type definitions for API contracts
│
├── package.json              # @vyzorix/types
│
├── tsconfig.json            # TypeScript (extends @vyzorix/tsconfig)
│
└── src/
    ├── index.ts            # Re-exports all types
    ├── device.ts           # Device types
    ├── command.ts          # Command types
    ├── api.ts              # API response types
    ├── auth.ts             # Auth types
    ├── telemetry.ts        # Telemetry types
    └── updater.ts          # OTA update types
```

### 4.6 packages/config/ (Shared Configurations)

```
packages/config/               # Shared configurations
│                              # Responsible for: ESLint, TypeScript, Tailwind base configs
│
├── eslint-config-web/        # ESLint config for web apps
│   ├── index.js             # .eslintrc replacement (flat config)
│   └── README.md
│
├── tsconfig-base/            # Base TypeScript config
│   ├── base.json            # Compiler options shared by all apps
│   └── README.md
│
└── tailwind-config/          # Base Tailwind config
    ├── base.ts              # Theme extension base
    └── README.md
```

### 4.7 packages/api-client/ (Go API Client)

```
packages/api-client/          # Go HTTP client library
│                              # Responsible for: Type-safe API client for external integrations
│
├── go.mod                   # module github.com/VinnsEdesigner/vyzorix/packages/api-client
│
├── client.go               # HTTP client with auth
├── types.go                # Shared Go types
├── mock.go                 # Mock server for testing
└── README.md               # Usage documentation
```

### 4.8 tooling/ (Build Scripts)

```
tooling/
│
├── scripts/                 # DevOps automation
│   ├── bootstrap.sh        # Initial setup: install deps, run migrations
│   ├── build.sh            # Full build: web + api
│   ├── test-all.sh         # Run all tests: unit + e2e
│   ├── release.sh          # Version bump + git tag + changelog
│   ├── lint.sh             # Run all linters
│   └── setup-db.sh         # Initialize SQLite database
│
└── docker/
    ├── Dockerfile.web       # Multi-stage build for React
    ├── Dockerfile.api      # Multi-stage build for Go
    └── docker-compose.yml   # Local development stack
```

### 4.9 docs/ (Documentation)

```
docs/
│
├── README.md               # Documentation index
│
├── ARCHITECTURE.md         # System design document
│
├── DEPLOYMENT.md           # Deployment guide (Render + Cloudflare)
│
├── API.md                  # API reference
│
├── REPOSITORY/             # Repository structure docs
│   ├── REPO_STRUCTURE.md  # This file
│   ├── CURRENT_STATE.md   # Current repo analysis
│   └── MIGRATION_PLAN.md  # Migration steps
│
└── SECURITY/              # Security documentation
    ├── README.md          # Security index
    ├── IMPLEMENTATION_PLAN.md  # Security implementation roadmap
    ├── THREAT_MODEL.md    # Risk assessment
    ├── AUTH.md            # Authentication pipeline
    ├── DEFENSE.md         # Defense matrix
    ├── FUNNEL.md          # Request funnel
    └── MATRIX.md          # API protection
```

---

## 5. Current → Target Mapping

| Current Path | Target Path | Notes |
|--------------|-------------|-------|
| `src/` | `apps/web/src/` | Move as-is |
| `public/` | `apps/web/public/` | Move, merge with Go's public |
| `src/routes/` | `apps/web/src/routes/` | Move as-is |
| `src/components/` | `apps/web/src/components/` | Move as-is |
| `src/hooks/` | `apps/web/src/hooks/` | Move as-is |
| `src/lib/` | `apps/web/src/lib/` | Move, some to packages/ |
| `controllers/` | `apps/api/internal/api/handlers/` | Rename |
| `middleware/` | `apps/api/internal/api/middleware/` | Rename |
| `models/` | `apps/api/pkg/models/` | Rename |
| `security/` | `apps/api/internal/auth/` | Rename, keep some in pkg/crypto |
| `hub/` | `apps/api/internal/ws/` | Rename |
| `storage/` | `apps/api/pkg/storage/` | Rename |
| `services/` | `apps/api/internal/` | Restructure |
| `config/` | `apps/api/pkg/config/` | Rename |
| `main.go` | `apps/api/main.go` | Move |
| `go.mod` | `apps/api/go.mod` | Split |
| `package.json` | `apps/web/package.json` | Split |
| `doc/` | `docs/` | Move, restructure |
| `SECURITY/` | `docs/SECURITY/` | Move |
| `scripts/` | `tooling/scripts/` | Move |
| `Dockerfile` | `apps/api/Dockerfile` | Move |
| `docker-compose.yml` | `tooling/docker/` | Move |
| `render.yaml` | root | Move to root |
| `.env.example` | root | Move to root |

---

## 6. Package Responsibilities

### 6.1 apps/web

**Purpose:** React SSR web dashboard for device management.

**Dependencies:**
- `packages/ui` - UI components
- `packages/types` - TypeScript types
- `apps/api` (via proxy) - API calls

**Exports:** None (deployable app)

**Key Files:**
- `src/main.tsx` - Client entry
- `src/start.ts` - SSR entry
- `src/router.tsx` - Route config
- `src/routes/` - Page components

### 6.2 apps/api

**Purpose:** Go backend for REST API, WebSocket, device management.

**Dependencies:**
- `packages/api-client` (optional) - For external integrations
- `pkg/*` - Internal packages

**Exports:** None (deployable app)

**Key Files:**
- `main.go` - Entry point
- `internal/api/router.go` - HTTP routing
- `internal/ws/hub.go` - WebSocket hub
- `pkg/storage/sqlite.go` - Database

### 6.3 packages/ui

**Purpose:** Reusable React UI components (shadcn/ui).

**Consumers:** `apps/web`

**Exports:** React components

**Key Files:**
- `src/components/ui/*` - UI primitives
- `src/lib/utils.ts` - cn() helper

### 6.4 packages/types

**Purpose:** Shared TypeScript type definitions.

**Consumers:** `apps/web`, future external clients

**Exports:** TypeScript interfaces

**Key Files:**
- `src/device.ts` - Device types
- `src/api.ts` - API types

### 6.5 packages/config

**Purpose:** Shared ESLint, TypeScript, Tailwind configurations.

**Consumers:** `apps/web`, `apps/api` (for Go linting)

**Exports:** Configuration objects

**Key Files:**
- `eslint-config-web/index.js` - ESLint config
- `tsconfig-base/base.json` - TS base config

### 6.6 packages/api-client

**Purpose:** Go HTTP client library for external integrations.

**Consumers:** Future external tools, scripts

**Exports:** Go package

**Key Files:**
- `client.go` - HTTP client

---

## 7. Build System

### 7.1 pnpm Workspaces

```yaml
# pnpm-workspace.yaml
packages:
  - 'apps/*'
  - 'packages/*'
```

### 7.2 Turborepo Pipeline

```json
// turbo.json
{
  "$schema": "https://turbo.build/schema.json",
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**", "!.next/**"],
      "cache": true
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "test": {
      "dependsOn": ["build"],
      "outputs": ["coverage/**"],
      "cache": true
    },
    "lint": {
      "cache": true
    },
    "typecheck": {
      "dependsOn": ["^build"],
      "cache": true
    }
  }
}
```

### 7.3 Build Commands

```bash
# Install all dependencies
pnpm install

# Development (all apps)
pnpm dev

# Development (single app)
pnpm --filter @vyzorix/web dev
pnpm --filter @vyzorix/api dev

# Build all
pnpm build

# Test all
pnpm test

# Lint all
pnpm lint

# Typecheck all
pnpm typecheck
```

---

## 8. Dependency Graph

```
                    ┌─────────────────────┐
                    │   Root package.json │
                    │   (workspaces only) │
                    └──────────┬──────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
          ▼                    ▼                    ▼
    ┌───────────┐        ┌───────────┐        ┌───────────┐
    │  tooling/ │        │  packages │        │   apps/   │
    │  (deps)   │        │  (deps)   │        │  (deps)   │
    └─────┬─────┘        └─────┬─────┘        └─────┬─────┘
          │                    │                    │
          │              ┌─────┴─────┐              │
          │              │           │              │
          ▼              ▼           ▼              ▼
    ┌───────────┐  ┌────────┐ ┌────────┐  ┌────────────┐
    │  scripts  │  │   ui    │ │ types  │  │    web     │
    │  (bash)  │  │(React)  │ │  (TS)  │  │  (React)   │
    └───────────┘  └────┬────┘ └────────┘  └──────┬─────┘
                        │                        │
                        │                        │ proxy /v1/* → api
                        ▼                        ▼
                   ┌────────────┐         ┌────────────┐
                   │ @vyzorix/ui│         │    api     │
                   │  (deps)    │         │   (Go)     │
                   └────────────┘         └─────┬──────┘
                                                │
                   ┌────────────────────────────┼────────────────────────────┐
                   │                            │                            │
                   ▼                            ▼                            ▼
            ┌────────────┐              ┌────────────┐              ┌────────────┐
            │   pkg/     │              │ internal/  │              │   pkg/     │
            │  storage   │              │    ws/     │              │   crypto   │
            └────────────┘              └────────────┘              └────────────┘
```

---

## 9. Naming Conventions

### 9.1 Package Names

| Package | npm/GitHub Name | Go Module Path |
|---------|-----------------|----------------|
| `apps/web` | `@vyzorix/web` | N/A (deployable) |
| `apps/api` | N/A | `github.com/VinnsEdesigner/vyzorix/apps/api` |
| `packages/ui` | `@vyzorix/ui` | N/A (TS only) |
| `packages/types` | `@vyzorix/types` | N/A (TS only) |
| `packages/config` | `@vyzorix/config` | N/A (TS only) |
| `packages/api-client` | N/A | `github.com/VinnsEdesigner/vyzorix/packages/api-client` |

### 9.2 File Naming

| Type | Convention | Example |
|------|------------|---------|
| Go source | `snake_case.go` | `auth_handler.go` |
| Go test | `*_test.go` | `auth_handler_test.go` |
| TypeScript source | `camelCase.ts` | `useAuth.ts` |
| TypeScript component | `PascalCase.tsx` | `DeviceCard.tsx` |
| React route | `kebab-case.tsx` | `login.tsx` |
| SQL migration | `NNN_name.sql` | `001_initial.sql` |

### 9.3 Directory Naming

| Type | Convention | Example |
|------|------------|---------|
| Go internal | `snake_case/` | `api/handlers/` |
| Go pkg | `snake_case/` | `pkg/models/` |
| TypeScript | `kebab-case/` | `hooks/` |
| React components | `PascalCase/` | `components/Dashboard/` |

---

## 10. Git Strategy

### 10.1 Branching Model

```
main                    # Production-ready code
├── develop             # Integration branch
│   ├── feature/*       # Feature branches
│   ├── fix/*           # Bug fix branches
│   └── refactor/*      # Refactoring branches
└── release/*           # Release preparation
```

### 10.2 Commit Convention

```
<type>(<scope>): <description>

Types:
  feat     - New feature
  fix      - Bug fix
  docs     - Documentation
  style    - Formatting (no code change)
  refactor - Code refactoring
  test     - Adding tests
  chore    - Maintenance tasks

Examples:
  feat(auth): add CSRF token validation
  fix(device): handle nil FCM token gracefully
  docs(api): update endpoint documentation
```

### 10.3 PR Workflow

1. Create feature branch from `develop`
2. Make changes + add tests
3. Run `pnpm lint && pnpm test`
4. Open PR to `develop`
5. Review + merge
6. Periodically merge `develop` → `main` for releases

---

## Appendix: Current Repository State

### A.1 Current Root Files

```
vyzorix-update-server/
├── .env                      # Environment (NOT in git)
├── .env.example              # Template
├── .gitignore
├── .golangci.yml
├── .prettierignore
├── .prettierrc
├── .github/
│   └── workflows/           # Existing CI (needs update)
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── render.yaml
├── go.mod
├── go.sum
├── package.json
├── package-lock.json
├── bun.lock                  # Legacy (will convert to pnpm)
├── bunfig.toml               # Legacy
├── tsconfig.json
├── vite.config.ts
├── eslint.config.js
├── components.json
└── coverage.out
```

### A.2 Current Source Structure

```
src/                    # React frontend
├── main.tsx
├── start.ts
├── server.ts
├── router.tsx
├── routeTree.gen.ts
├── routes/
├── components/
├── hooks/
├── lib/
└── styles.css

public/                 # Static files
├── index.html         # Placeholder (Go serves React)
├── landing.html        # Static landing page
├── favicon.ico
├── health.json
├── manifest.json
└── style.css

controllers/            # Go HTTP handlers
middleware/            # Go middleware
models/                # Go data models
security/              # Go security utils
hub/                   # WebSocket hub
storage/               # SQLite operations
services/              # Business services
config/                # Configuration
```

### A.3 Migration Complexity

| Component | Complexity | Notes |
|-----------|-------------|-------|
| `src/` → `apps/web/src/` | Low | Direct move |
| `public/` → `apps/web/public/` + Go public | Medium | Merge with Go's public |
| `controllers/` → `internal/api/handlers/` | Medium | Rename files |
| `middleware/` → `internal/api/middleware/` | Medium | Rename files |
| `models/` → `pkg/models/` | Medium | Move + rename |
| `security/` → `internal/auth/` | Medium | Restructure |
| `hub/` → `internal/ws/` | Low | Rename directory |
| `storage/` → `pkg/storage/` | Medium | Move + integrate |
| `services/` → `internal/` | High | Restructure + merge |
| `config/` → `pkg/config/` | Low | Rename directory |
| Package setup | Medium | pnpm workspaces + turbo |

---

**End of Document**

*Next Step: See [MIGRATION_PLAN.md](./MIGRATION_PLAN.md) for implementation steps.*