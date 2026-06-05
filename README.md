# vyzorix-update-server — C2 Command & Update Server

> **Phase 1.5** — Production-ready Go server with SQLite persistence, WebSocket hub, Firebase FCM, and React dashboard (TanStack Start). Designed for deployment on Render with a persistent `/data` disk volume.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    VYZORIX CONTROL PLATFORM                      │
│                                                                  │
│  ┌────────────────┐   WebSocket   ┌──────────────────────────┐  │
│  │  React Dash    │◄─────────────►│   WebSocket Broker       │  │
│  │  (TanStack)    │  HTTP POST    │   hub.Hub (gorilla/ws)   │  │
│  └───────┬────────┘  /v1/command  └───────────┬──────────────┘  │
│          │                                    │                  │
│          │ HTTP REST                          │ Persistent TCP   │
│          ▼                                    ▼                  │
│  ┌────────────────┐                 ┌──────────────────────────┐│
│  │  Gin REST API   │                 │  Android Daemon (C22)    ││
│  │  /api/v1/*      │                 │  VyzorixAudioRouter      ││
│  │  /v1/device/*   │                 └──────────────────────────┘│
│  └───────┬────────┘                                    │         │
│          │ FCM Silent Push (high priority)             │         │
│          └──────────────────────► [ Google FCM ] ────►         │
└─────────────────────────────────────────────────────────────────┘
```

## Repository Layout

```
vyzorix-update-server/
├── main.go                    # Server entrypoint (config → store → hub → gin)
├── go.mod / go.sum           # Go module + locked dependencies
├── Dockerfile                # Multi-stage: Go build + Node build → alpine runtime
├── docker-compose.yml        # Local dev orchestration with volumes
├── Makefile                  # build, test, docker, dev commands
├── render.yaml               # Render deployment blueprint with /data disk
├── .env.example              # All environment variables (Go + Supabase + Firebase)
├── .dockerignore
│
├── config/config.go          # Env var parsing → typed Config struct
├── storage/sqlite.go         # SQLite store: devices, telemetry, commands
├── security/hmac.go          # HMAC-SHA256 verification + nonce replay cache
├── hub/                      # WebSocket broker
│   ├── hub.go               # Client registry + broadcast hub goroutine
│   └── client.go            # readPump / writePump for each device connection
├── controllers/              # Gin HTTP handlers (domain-per-file)
│   ├── server.go            # Health, version, changelog, APK serve, dashboard SPA
│   ├── device.go            # Register, status, fcm-token, delete
│   ├── command.go            # Send command → WS or FCM fallback
│   ├── updater.go           # CheckUpdate, DownloadProgress
│   └── websocket_handler.go  # WS upgrade + telemetry broadcast to dashboard
├── middleware/               # Gin middleware
│   ├── auth.go              # Dashboard bearer-token authentication
│   ├── cors.go              # CORS with configurable allowed origins
│   ├── logger.go            # Structured JSON request logging
│   └── rate_limiter.go       # Token-bucket rate limiting per IP
├── models/                   # Shared REST + WebSocket payload types
│   ├── device.go            # Device, RegisterRequest, RegisterResponse, DeviceStatus
│   ├── command.go           # CommandFrame, CommandRequest, CommandResponse
│   ├── telemetry.go         # TelemetryFrame
│   ├── updater.go           # VersionManifest
│   ├── response.go          # APIError, APIResponse envelopes
│   └── models.go            # Package doc (re-exports sub-package types)
├── services/fcm/            # Firebase Cloud Messaging
│   ├── fcm.go               # Firebase Admin SDK initialization
│   └── notifier.go          # Silent high-priority wake payloads
│
├── public/                  # Static SPA assets (built by TanStack Start)
│   └── health.json          # Static health fallback when Go server is offline
├── data/                    # Runtime data directory (mounted from Render disk)
│   ├── version.json         # OTA manifest (copy from api/v1/ at startup)
│   └── changelog.json       # Release notes (copy from api/v1/ at startup)
├── api/v1/                  # Source-of-truth for version manifests
│   ├── version.json
│   └── changelog.json
├── bin/                     # APK binaries (populated by CI on release)
├── src/                     # React dashboard (TanStack Start, Lovable-managed)
│   ├── lib/vyzorix-api.ts  # Browser API client for all Go server endpoints
│   ├── lib/vyzorix-config.tsx # LocalStorage-persisted app settings
│   ├── hooks/use-device-stream.ts  # WebSocket with exponential backoff reconnect
│   ├── hooks/use-server-health.ts  # Polling health check hook
│   └── routes/              # TanStack file-based routes
└── doc/                     # Architecture docs (DOC_8, SPEC, etc.)
```

## Quick Start

### Prerequisites
- Go 1.22+
- Node 22+ (for frontend dev)
- Supabase project (for auth — Lovable Cloud handles this automatically)
- Firebase project with FCM enabled (for push notifications)

### Local Development

```bash
# 1. Clone and install dependencies
git clone https://github.com/VinnsEdesigner/vyzorix-update-server.git
cd vyzorix-update-server

# 2. Copy env file and fill in secrets
cp .env.example .env

# 3. Initialize data directory (creates ./data/ with version.json, changelog.json)
make init-data

# 4. Install Go dependencies
make deps

# 5. Build the Go binary
make build

# 6. Run the server
make run
# Server listening on :3000
# API: http://localhost:3000
# Health: http://localhost:3000/healthz

# Frontend dev (separate terminal):
npm install
npm run dev
```

### Docker

```bash
# Build the image
make docker-build

# Run with docker-compose
make docker-up

# View logs
make docker-logs

# Shell into container
make docker-shell

# Tear down
make docker-down
```

## Environment Variables

### Go Server (`main.go`)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | HTTP listen port |
| `NODE_ENV` | `production` | Set to `development` to disable HMAC enforcement |
| `DATABASE_URL` | `./data/vyzorix.db` | SQLite file path |
| `VYZORIX_API_DIR` | `./data` | Where version.json and changelog.json are served from |
| `VYZORIX_BIN_DIR` | `./bin` | Where APK files are served from |
| `VYZORIX_PUBLIC_DIR` | `./public` | SPA static assets |
| `TOKEN_SECRET` | — | Bearer token for dashboard authentication |
| `JWT_SECRET` | — | JWT signing secret (future use) |
| `FIREBASE_CREDENTIALS` | — | Full JSON of Firebase service account |
| `ALLOWED_ORIGINS` | `*` | Comma-separated CORS origins |
| `ENFORCE_HMAC` | `true` in production | Require HMAC signature on device requests |
| `HMAC_WINDOW_SECONDS` | `300` | Time window for HMAC nonce replay prevention |

### Supabase Auth (Lovable Cloud — auto-injected)

| Variable | Description |
|----------|-------------|
| `SUPABASE_URL` | Supabase project URL |
| `SUPABASE_PUBLISHABLE_KEY` | Supabase anon/public key (VITE_) |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase service role key (server-only) |

Lovable Cloud injects these automatically when the Supabase integration is connected. For local development, copy from **Supabase → Settings → API**.

## API Endpoints

### Public (no auth required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health`, `/healthz` | Liveness probe — checks DB connectivity |
| `GET` | `/api/v1/version` | OTA version manifest |
| `GET` | `/api/v1/changelog` | Release notes |
| `GET` | `/api/v1/apk/:name` | APK file download (with Range support) |
| `GET` | `/bin/*` | Binary file download |

### Device (HMAC or dashboard token)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/device/register` | Register/update a device |
| `GET` | `/v1/device/:id/status` | Get device status and last-seen |
| `PATCH` | `/v1/device/:id/fcm-token` | Update FCM push token |
| `DELETE` | `/v1/device/:id` | Remove device registration |
| `POST` | `/v1/device/:id/command` | Send command (WS delivery or FCM fallback) |
| `GET` | `/v1/device/:id/stream` | WebSocket upgrade for bidirectional stream |

### Dashboard (dashboard token required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/dashboard/devices` | List all registered devices with online status |

### SPA Fallback

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/*path` | Serves `public/index.html` for all non-matched routes |

## Deployment

### Render (Recommended)

1. Connect the GitHub repository to [Render](https://render.com)
2. Select `render.yaml` as the blueprint
3. Set the required environment variables (secrets marked `sync: false`)
4. Deploy — the `/data` disk mounts automatically

**Important (Hobby plan):** Render sleeps after 15–30 minutes of inactivity. Set up a free [UptimeRobot](https://uptimerobot.com) monitor to `GET https://your-service.onrender.com/healthz` every 10 minutes.

### Docker

```bash
docker build -t vyzorix-update-server .
docker run -p 3000:3000 \
  -v ./data:/data \
  -e DATABASE_URL=/data/vyzorix.db \
  -e VYZORIX_API_DIR=/data \
  -e FIREBASE_CREDENTIALS="$(cat firebase-creds.json)" \
  vyzorix-update-server
```

## Supabase Setup (Required for Auth)

The dashboard uses Supabase for operator authentication (Google OAuth + email/password). The `app_admins` table controls access:

```sql
create table public.app_admins (
  created_at timestamptz default now() not null,
  email text not null,
  user_id text
);
alter table public.app_admins enable row level security;
create policy "Allow authenticated reads" on public.app_admins
  for select to authenticated using (true);
create policy "Allow first admin bootstrap" on public.app_admins
  for insert to authenticated with check (
    (select count(*) from public.app_admins) = 0
  );
```

The first user to sign in is automatically granted admin access and the table locks from further signups.

## Available Make Targets

```bash
make help          # Show all targets
make deps          # Download Go dependencies
make tidy          # Tidy go.mod
make build         # Build binary locally
make build-linux   # Cross-compile for Linux
make run           # go run .
make test          # Run tests with race detection
make test-coverage # Run tests + generate coverage report
make fmt           # gofmt -s -w
make lint          # go vet ./...
make check         # fmt + lint + test
make docker-build  # Build Docker image
make docker-up     # docker-compose up -d
make docker-down   # docker-compose down
make docker-logs   # docker-compose logs -f
make init-data     # Create ./data/ with version.json, changelog.json
make init-env      # Create .env from .env.example
```

## Phase Context

| Phase | Server | Description |
|-------|--------|-------------|
| **Phase 1** | `cmd/mockserver/` | In-memory mock server — 7-day continuous test target |
| **Phase 1.5** | Root `main.go` | SQLite + WebSocket + FCM — this repository |
| **Phase 2** | Same | Dashboard wires up to `/v1/dashboard/*` endpoints |
| **Phase 3** | Future | Key rotation, multi-device, audit logging, KMS secret store |
