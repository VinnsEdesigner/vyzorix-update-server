# Vyzorix Monorepo Migration Plan

> **Document Version:** 1.0  
> **Status:** Proposed  
> **Last Updated:** 2026-06-08  
> **Depends On:** [REPO_STRUCTURE.md](./REPO_STRUCTURE.md)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Prerequisites](#2-prerequisites)
3. [Migration Strategy](#3-migration-strategy)
4. [Phase 1: Infrastructure Setup](#phase-1-infrastructure-setup)
5. [Phase 2: Create Directory Structure](#phase-2-create-directory-structure)
6. [Phase 3: Migrate Go Backend](#phase-3-migrate-go-backend)
7. [Phase 4: Migrate React Frontend](#phase-4-migrate-react-frontend)
8. [Phase 5: Create Shared Packages](#phase-5-create-shared-packages)
9. [Phase 6: Configure Build System](#phase-6-configure-build-system)
10. [Phase 7: Verify & Test](#phase-7-verify--test)
11. [Rollback Plan](#rollback-plan)

---

## 1. Overview

### 1.1 Purpose

This document provides **step-by-step instructions** for migrating the current `vyzorix-update-server` repository to the enterprise monorepo structure defined in [REPO_STRUCTURE.md](./REPO_STRUCTURE.md).

### 1.2 Migration Goals

1. ✅ Restructure into `apps/` and `packages/` layout
2. ✅ Implement pnpm workspaces
3. ✅ Configure Turborepo build pipeline
4. ✅ Maintain all existing functionality
5. ✅ Zero downtime migration

### 1.3 Time Estimate

| Phase | Description | Duration |
|-------|-------------|----------|
| Phase 1 | Infrastructure Setup | 30 min |
| Phase 2 | Create Directory Structure | 20 min |
| Phase 3 | Migrate Go Backend | 2 hours |
| Phase 4 | Migrate React Frontend | 1.5 hours |
| Phase 5 | Create Shared Packages | 1 hour |
| Phase 6 | Configure Build System | 30 min |
| Phase 7 | Verify & Test | 1 hour |
| **Total** | | **~6.5 hours** |

### 1.4 Important Notes

> ⚠️ **CRITICAL:** This migration is performed in a **feature branch**. The `main` branch remains untouched until verification is complete.

> 📝 All file movements must be tracked in git to enable rollback if needed.

---

## 2. Prerequisites

### 2.1 Required Tools

```bash
# Check installed versions
node --version    # Should be >= 20.0.0
pnpm --version    # Should be >= 9.0.0
go version        # Should be >= 1.22.0
git --version     # Should be >= 2.40.0
```

### 2.2 Pre-Migration Checklist

- [ ] Create backup of repository
- [ ] Commit all pending changes to `main`
- [ ] Verify all tests pass on current branch
- [ ] Notify stakeholders of migration window
- [ ] Disable auto-deployments temporarily

### 2.3 Backup Commands

```bash
# Create timestamped backup
cd /workspace/project
cp -r vyzorix-update-server vyzorix-update-server.backup.$(date +%Y%m%d_%H%M%S)

# Verify backup
ls -la vyzorix-update-server.backup.*/
```

---

## 3. Migration Strategy

### 3.1 Branch Strategy

```
main (protected - never touch directly)
└── feature/monorepo-migration (working branch)
    └── Will merge to main after successful testing
```

### 3.2 Migration Approach: "Lift and Shift"

```
Step 1: Create new structure (empty folders)
Step 2: Move files (no modifications)
Step 3: Update imports (automated + manual)
Step 4: Configure build system
Step 5: Verify everything works
```

### 3.3 Commit Strategy

Each phase should be a separate commit with clear message:

```
feat(monorepo): create directory structure
feat(monorepo): migrate Go backend to apps/api
feat(monorepo): migrate React frontend to apps/web
feat(monorepo): add shared packages
feat(monorepo): configure pnpm workspaces
feat(monorepo): add Turborepo pipeline
fix(monorepo): resolve import path issues
feat(monorepo): run full verification
```

---

## Phase 1: Infrastructure Setup

### Step 1.1: Create Feature Branch

```bash
cd /workspace/project/vyzorix-update-server

# Ensure main is up to date
git checkout main
git pull origin main

# Create migration branch
git checkout -b feature/monorepo-migration

# Verify we're on the new branch
git branch --show-current
```

### Step 1.2: Install pnpm (if not present)

```bash
# Check if pnpm is installed
which pnpm || npm install -g pnpm@latest

# Verify installation
pnpm --version
```

### Step 1.3: Create pnpm-workspace.yaml

Create file: `pnpm-workspace.yaml`

```yaml
packages:
  - 'apps/*'
  - 'packages/*'
```

### Step 1.4: Update Root package.json

Replace contents of `package.json` with:

```json
{
  "name": "vyzorix",
  "version": "0.0.0",
  "private": true,
  "workspaces": [
    "apps/*",
    "packages/*"
  ],
  "scripts": {
    "dev": "turbo run dev",
    "build": "turbo run build",
    "test": "turbo run test",
    "lint": "turbo run lint",
    "typecheck": "turbo run typecheck",
    "clean": "turbo run clean"
  },
  "devDependencies": {
    "turbo": "^2.0.0"
  },
  "packageManager": "pnpm@9.0.0",
  "engines": {
    "node": ">=20.0.0",
    "pnpm": ">=9.0.0"
  }
}
```

### Step 1.5: Create turbo.json

Create file: `turbo.json`

```json
{
  "$schema": "https://turbo.build/schema.json",
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**", ".next/**", "!.next/cache/**"],
      "cache": true
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "test": {
      "dependsOn": ["build"],
      "outputs": ["coverage/**", "*.out"],
      "cache": true
    },
    "lint": {
      "cache": true
    },
    "typecheck": {
      "dependsOn": ["^build"],
      "cache": true
    },
    "clean": {
      "cache": false
    }
  }
}
```

### Step 1.6: Commit Phase 1

```bash
git add pnpm-workspace.yaml package.json turbo.json
git commit -m "feat(monorepo): setup pnpm workspaces and turbo pipeline"
```

---

## Phase 2: Create Directory Structure

### Step 2.1: Create All Directories

```bash
# Create apps directory structure
mkdir -p apps/web/src/{routes,components,hooks,lib,integrations}
mkdir -p apps/web/public/assets
mkdir -p apps/web/tests/e2e

# Create packages directory structure
mkdir -p packages/ui/src/{components/ui,lib}
mkdir -p packages/types/src
mkdir -p packages/config/{eslint-config-web,tsconfig-base,tailwind-config}

# Create Go api structure
mkdir -p apps/api/{cmd/server,internal/{api/{handlers,middleware},auth,device,ws,fcm},pkg/{models,crypto,storage/migrations,config}}
mkdir -p apps/api/scripts

# Create tooling structure
mkdir -p tooling/{scripts,docker}
```

### Step 2.2: Create .gitkeep Files

```bash
# Ensure empty directories are tracked
touch apps/web/public/assets/.gitkeep
touch apps/web/tests/e2e/.gitkeep
touch apps/api/pkg/storage/migrations/.gitkeep
```

### Step 2.3: Create apps/api/go.mod

Create file: `apps/api/go.mod`

```go
module github.com/VinnsEdesigner/vyzorix/apps/api

go 1.22

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/gorilla/websocket v1.5.1
	github.com/mattn/go-sqlite3 v1.14.22
	golang.org/x/crypto v0.21.0
	golang.org/x/time v0.5.0
)

require (
	github.com/bytedance/sonic v1.11.3 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.19.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.7.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

### Step 2.4: Create packages/ui/package.json

Create file: `packages/ui/package.json`

```json
{
  "name": "@vyzorix/ui",
  "version": "0.0.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "exports": {
    ".": "./src/index.ts",
    "./utils": "./src/lib/utils.ts"
  },
  "scripts": {
    "lint": "eslint src",
    "typecheck": "tsc --noEmit"
  },
  "devDependencies": {
    "@vyzorix/tsconfig": "workspace:*",
    "typescript": "^5.0.0"
  },
  "peerDependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  }
}
```

### Step 2.5: Create packages/types/package.json

Create file: `packages/types/package.json`

```json
{
  "name": "@vyzorix/types",
  "version": "0.0.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "scripts": {
    "lint": "eslint src",
    "typecheck": "tsc --noEmit"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}
```

### Step 2.6: Commit Phase 2

```bash
git add -A
git commit -m "feat(monorepo): create directory structure and package manifests"
```

---

## Phase 3: Migrate Go Backend

### Step 3.1: Move Main Entry Point

```bash
# Move main.go
mv main.go apps/api/main.go

# Move go.mod content (merge into existing apps/api/go.mod)
# This is manual - copy requires sections from root go.mod
```

### Step 3.2: Move Controllers → Handlers

```bash
# Move and rename controller files
mv controllers/auth.go apps/api/internal/api/handlers/auth.go
mv controllers/auth_test.go apps/api/internal/api/handlers/auth_test.go
mv controllers/device.go apps/api/internal/api/handlers/device.go
mv controllers/device_test.go apps/api/internal/api/handlers/device_test.go
mv controllers/command.go apps/api/internal/api/handlers/command.go
mv controllers/command_test.go apps/api/internal/api/handlers/command_test.go
mv controllers/updater.go apps/api/internal/api/handlers/updater.go

# Remove server.go (functionality moved to router.go)
# Note: We will recreate this functionality
```

### Step 3.3: Move Middleware

```bash
# Move middleware files
mv middleware/auth.go apps/api/internal/api/middleware/auth.go
mv middleware/auth_test.go apps/api/internal/api/middleware/auth_test.go
mv middleware/cors.go apps/api/internal/api/middleware/cors.go
mv middleware/cors_test.go apps/api/internal/api/middleware/cors_test.go
mv middleware/ratelimit.go apps/api/internal/api/middleware/ratelimit.go
mv middleware/ratelimit_test.go apps/api/internal/api/middleware/ratelimit_test.go
mv middleware/body_size.go apps/api/internal/api/middleware/body_size.go
mv middleware/body_size_test.go apps/api/internal/api/middleware/body_size_test.go
mv middleware/logger.go apps/api/internal/api/middleware/logger.go
mv middleware/logger_test.go apps/api/internal/api/middleware/logger_test.go
mv middleware/request_id.go apps/api/internal/api/middleware/request_id.go
mv middleware/request_id_test.go apps/api/internal/api/middleware/request_id_test.go
```

### Step 3.4: Move Models

```bash
# Move to pkg/models
mv models/auth.go apps/api/pkg/models/auth.go
mv models/command.go apps/api/pkg/models/command.go
mv models/device.go apps/api/pkg/models/device.go
mv models/models.go apps/api/pkg/models/models.go
mv models/models_test.go apps/api/pkg/models/models_test.go
mv models/response.go apps/api/pkg/models/response.go
mv models/telemetry.go apps/api/pkg/models/telemetry.go
mv models/updater.go apps/api/pkg/models/updater.go
```

### Step 3.5: Move Security → Auth + Crypto

```bash
# Move security → internal/auth (authentication logic)
mv security/jwt.go apps/api/internal/auth/jwt.go
mv security/jwt_test.go apps/api/internal/auth/jwt_test.go
mv security/password.go apps/api/internal/auth/password.go
mv security/password_test.go apps/api/internal/auth/password_test.go
mv security/ratelimit.go apps/api/internal/auth/ratelimit.go
mv security/ratelimit_test.go apps/api/internal/auth/ratelimit_test.go
mv security/google_token.go apps/api/internal/auth/google_token.go
mv security/google_token_test.go apps/api/internal/auth/google_token_test.go
mv security/validate.go apps/api/internal/auth/validate.go
mv security/validate_test.go apps/api/internal/auth/validate_test.go

# Move crypto to pkg/crypto (cryptographic utilities)
mv security/hmac.go apps/api/pkg/crypto/hmac.go
mv security/hmac_test.go apps/api/pkg/crypto/hmac_test.go

# Move origin to internal/api/middleware (already done above)
# Note: origin.go is part of CORS handling
```

### Step 3.6: Move Hub → WebSocket

```bash
# Move WebSocket hub
mv hub/hub.go apps/api/internal/ws/hub.go
mv hub/hub_test.go apps/api/internal/ws/hub_test.go
mv hub/client.go apps/api/internal/ws/client.go

# Move WebSocket handler
mv controllers/websocket_handler.go apps/api/internal/ws/handler.go
mv controllers/websocket_handler_test.go apps/api/internal/ws/handler_test.go
```

### Step 3.7: Move Storage

```bash
# Move to pkg/storage
mv storage/sqlite.go apps/api/pkg/storage/sqlite.go
mv storage/sqlite_test.go apps/api/pkg/storage/sqlite_test.go
mv storage/client_settings_test.go apps/api/pkg/storage/client_settings_test.go
mv storage/pagination_test.go apps/api/pkg/storage/pagination_test.go
mv storage/secret_hash_test.go apps/api/pkg/storage/secret_hash_test.go
```

### Step 3.8: Move Services

```bash
# Move to internal/ (restructure services)
mv services/command_signer.go apps/api/internal/command_signer.go
mv services/command_signer_test.go apps/api/internal/command_signer_test.go
mv services/email.go apps/api/internal/email.go
mv services/email_test.go apps/api/internal/email_test.go

# Move FCM
mv services/fcm/notifier.go apps/api/internal/fcm/notifier.go
# Create internal/fcm directory if needed
```

### Step 3.9: Move Config

```bash
# Move to pkg/config
mv config/config.go apps/api/pkg/config/config.go
mv config/config_test.go apps/api/pkg/config/config_test.go
```

### Step 3.10: Move Scripts

```bash
# Move to tooling/scripts
mv scripts/migrate.sh tooling/scripts/migrate.sh
mv scripts/seed.sh tooling/scripts/seed.sh
mv scripts/compute_checksum.sh tooling/scripts/compute_checksum.sh
mv scripts/generate_version.sh tooling/scripts/generate_version.sh
mv scripts/validate_apk.sh tooling/scripts/validate_apk.sh
mv scripts/cleanup_old_apks.sh tooling/scripts/cleanup_old_apks.sh
```

### Step 3.11: Create apps/api/internal/api/router.go

Create new file: `apps/api/internal/api/router.go`

```go
package api

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Router     *gin.Engine
	Handlers   *handlers.Handlers
	Middleware *middleware.Middleware
	Auth       *auth.Auth
	WSHub      *ws.Hub
	Config     *config.Config
	Store      *storage.Store
}

func NewServer(cfg *config.Config, store *storage.Store, hub *ws.Hub) *Server {
	s := &Server{
		Config: cfg,
		Store:  store,
		WSHub:  hub,
	}

	// Initialize auth
	s.Auth = auth.NewAuth(cfg, store)

	// Initialize handlers
	s.Handlers = handlers.NewHandlers(s.Auth, store, hub)

	// Initialize middleware
	s.Middleware = middleware.NewMiddleware(cfg, store)

	// Setup router
	s.setupRouter()

	return s
}

func (s *Server) setupRouter() {
	if s.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(s.Config.AllowedOrigins))
	r.Use(middleware.BodySizeLimit())

	// Public routes
	public := r.Group("")
	public.Use(s.Middleware.RateLimit())
	{
		public.GET("/health", s.Handlers.Health)
		public.GET("/healthz", s.Handlers.Health)
		public.GET("/api/v1/version", s.Handlers.Version)
		public.GET("/api/v1/changelog", s.Handlers.Changelog)
	}

	// Auth routes
	auth := r.Group("/v1/auth")
	{
		auth.POST("/login", s.Middleware.AuthLimiter(), s.Handlers.Login)
		auth.POST("/register", s.Middleware.AuthLimiter(), s.Handlers.Register)
		// ... other auth routes
	}

	// Protected routes
	protected := r.Group("")
	protected.Use(s.Middleware.JWTAuth())
	{
		protected.GET("/v1/dashboard/devices", s.Handlers.DashboardDevices)
		// ... other protected routes
	}

	s.Router = r
}

func (s *Server) Engine() *gin.Engine {
	return s.Router
}
```

### Step 3.12: Create apps/api/internal/api/middleware/security.go

Create new file: `apps/api/internal/api/middleware/security.go`

```go
package middleware

import (
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	"github.com/gin-gonic/gin"
)

type SecurityMiddleware struct {
	Config *config.Config
}

func NewSecurityMiddleware(cfg *config.Config) *SecurityMiddleware {
	return &SecurityMiddleware{Config: cfg}
}

// SecurityHeaders adds security-related HTTP headers
func (sm *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")
		c.Next()
	}
}

// PanicRecovery catches panics and returns a safe error response
func (sm *SecurityMiddleware) PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_error",
					"message": "An unexpected error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

### Step 3.13: Clean Up Empty Directories

```bash
# Remove old directories (should be empty now)
rmdir controllers middleware models security hub services config cmd/mockserver 2>/dev/null || true

# Remove legacy files
rm -f coverage.out
rm -f server.log vite.log 2>/dev/null || true
```

### Step 3.14: Commit Phase 3

```bash
git add -A
git commit -m "feat(monorepo): migrate Go backend to apps/api structure"
```

---

## Phase 4: Migrate React Frontend

### Step 4.1: Move src/ Files

```bash
# Move React source files
mv src/main.tsx apps/web/src/main.tsx
mv src/start.ts apps/web/src/start.ts
mv src/server.ts apps/web/src/server.ts
mv src/router.tsx apps/web/src/router.tsx
mv src/routeTree.gen.ts apps/web/src/routeTree.gen.ts
mv src/styles.css apps/web/src/styles.css

# Move routes
mv src/routes/* apps/web/src/routes/

# Move components
mv src/components/* apps/web/src/components/

# Move hooks
mv src/hooks/* apps/web/src/hooks/

# Move lib
mv src/lib/* apps/web/src/lib/
```

### Step 4.2: Move public/ Files

```bash
# Move public assets
mv public/landing.html apps/web/public/landing.html
mv public/index.html apps/web/public/index.html
mv public/favicon.ico apps/web/public/favicon.ico
mv public/health.json apps/web/public/health.json
mv public/manifest.json apps/web/public/manifest.json
mv public/style.css apps/web/public/style.css
```

### Step 4.3: Move Frontend Config Files

```bash
# Move frontend configuration
mv package.json apps/web/package.json
mv package-lock.json apps/web/package-lock.json
mv tsconfig.json apps/web/tsconfig.json
mv vite.config.ts apps/web/vite.config.ts
mv eslint.config.js apps/web/eslint.config.js
mv tailwind.config.ts apps/web/tailwind.config.ts
mv components.json apps/web/components.json
```

### Step 4.4: Update apps/web/package.json

Update the package.json to reflect new structure:

```json
{
  "name": "@vyzorix/web",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite dev",
    "build": "vite build",
    "build:dev": "vite build --mode development",
    "preview": "vite preview",
    "lint": "eslint src",
    "typecheck": "tsc --noEmit",
    "test": "vitest run"
  },
  "dependencies": {
    "@tanstack/react-router": "^1.168.25",
    "@tanstack/react-start": "^1.167.50",
    "react": "^19.2.0",
    "react-dom": "^19.2.0"
    // ... other deps
  }
}
```

### Step 4.5: Update vite.config.ts Proxy

Update the proxy configuration:

```typescript
import { defineConfig } from "@lovable.dev/vite-tanstack-config";

export default defineConfig({
  tanstackStart: {
    server: { entry: "server" },
  },
  vite: {
    server: {
      proxy: {
        '/v1': {
          target: process.env.API_URL || 'http://localhost:3000',
          changeOrigin: true,
        },
        '/api': {
          target: process.env.API_URL || 'http://localhost:3000',
          changeOrigin: true,
        },
        '/health': {
          target: process.env.API_URL || 'http://localhost:3000',
          changeOrigin: true,
        },
      },
    },
  },
});
```

### Step 4.6: Clean Up Frontend Directories

```bash
# Remove empty source directories
rmdir src/routes src/components src/hooks src/lib src/integrations 2>/dev/null || true

# Remove build artifacts
rm -rf dist node_modules 2>/dev/null || true

# Remove legacy files
rm -f bun.lock bunfig.toml 2>/dev/null || true
```

### Step 4.7: Commit Phase 4

```bash
git add -A
git commit -m "feat(monorepo): migrate React frontend to apps/web structure"
```

---

## Phase 5: Create Shared Packages

### Step 5.1: Create packages/ui Components

```bash
# Create ui package structure
cat > packages/ui/src/index.ts << 'EOF'
export * from "./components/ui/button"
export * from "./components/ui/card"
export * from "./components/ui/input"
export * from "./lib/utils"
EOF

# Copy existing shadcn components (if any were customized)
# Note: Standard shadcn components will be added via CLI
```

### Step 5.2: Create packages/types

```bash
# Create type definitions
cat > packages/types/src/index.ts << 'EOF'
export * from "./device"
export * from "./command"
export * from "./api"
export * from "./auth"
export * from "./telemetry"
export * from "./updater"
EOF

cat > packages/types/src/device.ts << 'EOF'
export interface Device {
  id: string;
  deviceId: string;
  appVersion: string;
  deviceClass: string;
  firebaseInstallId?: string;
  fcmToken?: string;
  online: boolean;
  lastSeen: number;
}

export interface DeviceStatus {
  deviceId: string;
  online: boolean;
  lastSeen: number;
  appVersion: string;
  deviceClass: string;
}
EOF
```

### Step 5.3: Create packages/config

```bash
# Create eslint config
cat > packages/config/eslint-config-web/index.js << 'EOF'
import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

export const eslintConfig = [
  ...compat.extends("next/core-web-vitals"),
  {
    rules: {
      "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
    },
  },
];
EOF

# Create tsconfig base
cat > packages/config/tsconfig-base/base.json << 'EOF'
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx"
  },
  "exclude": ["node_modules"]
}
EOF
```

### Step 5.4: Commit Phase 5

```bash
git add -A
git commit -m "feat(monorepo): create shared packages (ui, types, config)"
```

---

## Phase 6: Configure Build System

### Step 6.1: Update Root go.mod

Update root `go.mod` to only contain tooling dependencies:

```go
module github.com/VinnsEdesigner/vyzorix

go 1.22

require (
	golang.org/x/vuln v1.0.0
)
```

### Step 6.2: Create GitHub Actions Workflows

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup pnpm
        uses: pnpm/action-setup@v3
        with:
          version: 9
          
      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
          
      - name: Install dependencies
        run: pnpm install --frozen-lockfile
        
      - name: Lint
        run: pnpm lint
        
      - name: Typecheck
        run: pnpm typecheck

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup pnpm
        uses: pnpm/action-setup@v3
        with:
          version: 9
          
      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'
          
      - name: Install dependencies
        run: pnpm install --frozen-lockfile
        
      - name: Test
        run: pnpm test
        
  go-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v5
        with:
          version: latest
          working-directory: apps/api
          
      - name: Run tests
        run: go test ./...
        working-directory: apps/api
```

### Step 6.3: Create Deploy Workflows

Create `.github/workflows/deploy-api.yml` and `.github/workflows/deploy-web.yml` with Render deployment steps.

### Step 6.4: Move Docker Files

```bash
# Move Dockerfile to apps/api
mv Dockerfile apps/api/Dockerfile

# Move docker-compose to tooling/docker
mv docker-compose.yml tooling/docker/docker-compose.yml

# Move render.yaml to root (already there)
```

### Step 6.5: Create Makefile in apps/api

Create `apps/api/Makefile`:

```makefile
.PHONY: build run test lint clean migrate

build:
	go build -o bin/server .

run:
	go run .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
	rm -f coverage.out

migrate:
	./tooling/scripts/migrate.sh
```

### Step 6.6: Commit Phase 6

```bash
git add -A
git commit -m "feat(monorepo): configure build system and CI/CD"
```

---

## Phase 7: Verify & Test

### Step 7.1: Install Dependencies

```bash
# Clean install all dependencies
pnpm install
```

### Step 7.2: Build All Packages

```bash
# Build web
pnpm --filter @vyzorix/web build

# Build Go API
cd apps/api && go mod tidy && go build -o bin/server .
```

### Step 7.3: Run Tests

```bash
# Run all tests
pnpm test

# Run Go tests
cd apps/api && go test ./...
```

### Step 7.4: Verify Dev Server

```bash
# Terminal 1: Start API
cd apps/api && go run .

# Terminal 2: Start Web
pnpm --filter @vyzorix/web dev
```

### Step 7.5: Manual Testing Checklist

- [ ] Landing page loads at `/`
- [ ] Login page accessible at `/login`
- [ ] User can register
- [ ] User can login
- [ ] Dashboard shows devices
- [ ] Device commands work
- [ ] WebSocket connection established
- [ ] Health endpoint responds

### Step 7.6: Commit Final

```bash
git add -A
git commit -m "feat(monorepo): complete migration verification"
```

---

## Rollback Plan

### If Migration Fails

```bash
# 1. Checkout main branch
git checkout main

# 2. Delete migration branch
git branch -D feature/monorepo-migration

# 3. Restore from backup
rm -rf vyzorix-update-server
cp -r vyzorix-update-server.backup.*/vyzorix-update-server .

# 4. Verify restoration
git status
```

### Partial Rollback

If only certain phases failed:

```bash
# Revert to specific commit
git revert <commit-hash>

# Or reset to last working state
git reset --hard <last-working-commit>
```

---

## Post-Migration Tasks

### 1. Update Documentation

- [ ] Update README.md with new structure
- [ ] Update SETUP-GUIDE.md
- [ ] Update SECURITY docs path references

### 2. Update External Integrations

- [ ] Update Render deployment configuration
- [ ] Update Cloudflare settings if needed
- [ ] Notify any API consumers of path changes

### 3. Cleanup

- [ ] Remove backup directory
- [ ] Delete unused branches
- [ ] Update git remote if needed

---

## Quick Reference

### Common Commands

```bash
# Start migration
git checkout -b feature/monorepo-migration

# Install deps
pnpm install

# Build all
pnpm build

# Test all
pnpm test

# Dev mode
pnpm dev

# Commit phase
git add -A && git commit -m "feat(monorepo): <description>"
```

### File Movement Summary

| From | To |
|------|-----|
| `main.go` | `apps/api/main.go` |
| `controllers/` | `apps/api/internal/api/handlers/` |
| `middleware/` | `apps/api/internal/api/middleware/` |
| `models/` | `apps/api/pkg/models/` |
| `security/` | `apps/api/internal/auth/` + `apps/api/pkg/crypto/` |
| `hub/` | `apps/api/internal/ws/` |
| `storage/` | `apps/api/pkg/storage/` |
| `services/` | `apps/api/internal/` |
| `config/` | `apps/api/pkg/config/` |
| `src/` | `apps/web/src/` |
| `public/` | `apps/web/public/` |
| `scripts/` | `tooling/scripts/` |

---

**End of Document**

*Ready to execute migration. Follow phases in order.*