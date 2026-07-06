# Vyzorix Update Server - Agent Context

## Project Overview

Vyzorix Update Server is a Go-based update management system with React frontend. It provides device management, OTA updates, and real-time communication via WebSocket.

**Repository:** `VinnsEdesigner/vyzorix-update-server`  
**Main Technologies:** Go 1.24, React 19, TanStack Start, SQLite, Gin

---

## Architecture

```
apps/
 api/           # Go backend server
    cmd/server/          # Entry point
    internal/
       api/            # HTTP handlers and middleware
          handlers/   # Route handlers (auth, device, command, websocket)
          middleware/ # Security, CORS, rate limiting
       application/     # Business logic services
       domain/         # Domain entities and repository interfaces
       infrastructure/ # External integrations
          crypto/     # Cryptographic utilities
          external/   # Email, SMS services
          auth/       # Auth implementations
          ws/          # WebSocket hub
       fcm/           # Firebase Cloud Messaging
       ssr/           # Server-side rendering
    pkg/               # Shared packages (config, models, storage, logging, crypto)
    main.go            # Server entry point

 web/           # React frontend
    src/       # React components, routes, hooks

 ... (other apps)

packages/
 config/        # ESLint, TypeScript configs
 types/        # Shared TypeScript types
 ui/           # Shared React components

docs/
 REPOSITORY/   # Repository structure migration docs
 SECURITY/     # Security architecture
 ...           # Feature specs and ADRs
```

---

## Key Go Packages

| Package | Purpose |
|---------|---------|
| `internal/api/handlers` | HTTP request handlers |
| `internal/application` | Business logic layer |
| `internal/domain` | Domain entities and repository interfaces |
| `internal/infrastructure` | External service integrations |
| `pkg/models` | Shared data models |
| `pkg/storage` | SQLite database operations |
| `pkg/crypto` | Cryptographic utilities |

---

## SiliconFlow AI Agent

### Setup

```bash
# Set API key
export SILICONFLOW_API_KEY="sk-your-key"

# Run agent
cd AGENTS
python3 siliconflow_agent.py
```

### Available Models

| Key | Model | Best For |
|-----|-------|----------|
| `deepseek-v4-pro` | DeepSeek-V4-Pro | Fast reasoning, general tasks  DEFAULT |
| `kimi-k2.7-code` | Kimi-K2.7-Code | Coding, debugging  SECONDARY |
| `kimi-k2.6` | Kimi-K2.6 | General purpose |
| `glm-5.2` | GLM-5.2 | Latest model (June 16) |

### Quick Usage

```python
from siliconflow_agent import quick_ask, coding_agent, use_model

# One-shot question
quick_ask("How do I add a new endpoint?")

# Interactive coding session
agent = coding_agent()
agent("Refactor the auth handlers")

# Use specific model
kimi = use_model('kimi-k2.7-code')
kimi("Explain the WebSocket hub implementation")
```

---

## Common Tasks

### Run the Server
```bash
cd apps/api
go run main.go
```

### Run Tests
```bash
cd apps/api
go test ./...
```

### Run Linter (golangci-lint v2.12.2)
```bash
cd apps/api
make lint-go
```

### Build Frontend
```bash
pnpm --filter @vyzorix/web build
```

---

## Migration Status

The codebase is undergoing restructuring from flat structure to enterprise monorepo.

**Completed:**
-  `apps/api/internal/api/handlers/` - Split auth handlers
-  `apps/api/internal/domain/` - Domain entities
-  `apps/api/internal/application/` - Application services
-  `apps/api/internal/infrastructure/crypto/` - Cryptographic utilities
-  `apps/api/internal/infrastructure/external/` - Email service

**In Progress:**
-  Verifying all old handlers vs new handlers
-  Removing duplicate files after verification

**See:** `docs/REPOSITORY/REPO_STRUCTURE.md` and `docs/FILE_MAPPING.md`

---

## Important Files

| File | Description |
|------|-------------|
| `apps/api/main.go` | Server entry point |
| `apps/api/internal/api/server.go` | Gin router setup |
| `apps/api/internal/api/handlers/` | All HTTP handlers |
| `apps/api/pkg/storage/store.go` | Database initialization |
| `docs/FILE_MAPPING.md` | Migration file mapping |
| `AGENTS/SETUP_GUIDE.md` | Development setup guide |
