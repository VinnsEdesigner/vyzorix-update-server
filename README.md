# Vyzorix Update Server

A production-ready command-and-control (C2) server for managing Android device fleets. Built with Go, SQLite, WebSockets, and Firebase Cloud Messaging.

---

## Overview

Vyzorix Update Server provides:
- **Device Registration & Management** — Register, track, and manage Android devices
- **Over-The-Air (OTA) Updates** — Serve APK updates with version manifest support
- **Real-Time Telemetry** — WebSocket-based live device monitoring
- **Command Dispatch** — Send commands via WebSocket or FCM push notifications
- **Secure Authentication** — JWT sessions + Google OAuth + email/password
- **Email Verification** — Account verification and password reset via Resend

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    VYZORIX CONTROL PLATFORM                      │
│                                                                  │
│  ┌────────────────┐   WebSocket   ┌──────────────────────────┐  │
│  │  React Dash    │◄─────────────►│   WebSocket Broker       │  │
│  │  (TanStack)    │  HTTP POST    │   hub.Hub (gorilla/ws)   │  │
│  │  + Auth        │  /v1/command  └───────────┬──────────────┘  │
│  └───────┬────────┘                            │                  │
│          │ HTTP REST                         │ Persistent TCP    │
│          ▼                                   ▼                  │
│  ┌────────────────┐                 ┌──────────────────────────┐ │
│  │  Gin REST API   │                 │  Android Daemon (C22)    │ │
│  │  /api/v1/*      │                 │  VyzorixAudioRouter      │ │
│  │  /v1/device/*   │                 └──────────────────────────┘ │
│  │  /v1/auth/*     │                                    │         │
│  └───────┬────────┘                                    │         │
│          │ FCM Silent Push (high priority)             │         │
│          └──────────────────────► [ Google FCM ] ────►         │
└─────────────────────────────────────────────────────────────────┘
```

### Tech Stack

| Component | Technology |
|-----------|------------|
| **Backend** | Go 1.22+, Gin web framework |
| **Database** | SQLite (WAL mode, persistent on Render) |
| **WebSocket** | gorilla/websocket |
| **Push Notifications** | Firebase Cloud Messaging (FCM) |
| **Authentication** | JWT + Google OAuth + email/password |
| **Email** | Resend API |
| **Frontend** | React 18, TanStack Start, Vite |
| **Styling** | Tailwind CSS, Radix UI |
| **Deployment** | Docker, Render |

---

## Repository Structure

```
vyzorix-update-server/
├── cmd/                    # Entry points
│   └── mockserver/         # Phase 1 mock server
├── config/                  # Environment configuration
├── controllers/             # HTTP request handlers
│   ├── auth.go             # Login, register, Google OAuth, password reset
│   ├── device.go           # Device registration, status, FCM
│   ├── command.go          # Command dispatch
│   └── server.go           # Health, version, SPA serving
├── hub/                    # WebSocket broker
├── middleware/              # Auth, CORS, logging, rate limiting
├── models/                  # Type definitions
├── security/               # JWT, HMAC, Google OAuth, password validation
│   ├── jwt.go              # JWT signing and verification
│   ├── hmac.go             # HMAC-SHA256 command signing
│   ├── google_token.go     # Google OAuth ID token verification
│   ├── password.go         # Password complexity validation
│   └── ratelimit.go        # In-memory rate limiting
├── services/               # Business logic
│   ├── fcm/                # Firebase Cloud Messaging
│   ├── email.go            # Resend email service
│   └── command_signer.go   # Command signing
├── storage/                 # SQLite database operations
├── src/                     # React frontend
│   ├── components/         # UI components
│   ├── hooks/              # Custom React hooks
│   ├── lib/                # API client, utilities
│   └── routes/             # TanStack file-based routes
├── public/                  # Static assets (HTML, CSS)
├── data/                    # Runtime data (persistent disk)
├── api/v1/                  # Version manifest source
├── bin/                     # APK storage (CI populated)
├── doc/                     # Architecture documentation
├── scripts/                 # Build automation scripts
├── SETUP-GUIDE.md          # Third-party setup (Google OAuth, Resend, Render)
├── prd-security-email-password.md  # Security feature implementation plan
├── doc/REPO_TREE.md        # Full repository structure
└── README.md              # This file
```

---

## Quick Start

### Prerequisites

- Go 1.22+
- Node 22+ (for frontend development)
- Firebase project with FCM enabled
- Google Cloud project for OAuth (optional)
- Resend account for email (optional)

### Local Development

```bash
# Clone the repository
git clone https://github.com/VinnsEdesigner/vyzorix-update-server.git
cd vyzorix-update-server

# Copy environment template
cp .env.example .env

# Initialize data directory
make init-data

# Install Go dependencies
make deps

# Install Node dependencies
npm install

# Run Go server (port 3000)
make run

# In another terminal, run frontend dev server (port 5173)
npm run dev
```

### Docker

```bash
# Build Docker image
make docker-build

# Run with docker-compose
make docker-up

# View logs
make docker-logs

# Stop
make docker-down
```

---

## Authentication

The server supports three authentication methods:

### 1. Email & Password
- Registration with email verification
- Password reset via email link
- Password complexity requirements (8+ chars, upper, lower, number, special)

### 2. Google OAuth
- One-click sign-in with Google
- Cryptographic JWT verification via Google's JWKS endpoint

### 3. JWT Sessions
- All authenticated requests use Bearer tokens
- Sessions can be revoked

---

## API Endpoints

### Public (No Auth)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health`, `/healthz` | Liveness probe |
| `GET` | `/api/v1/version` | OTA version manifest |
| `GET` | `/api/v1/changelog` | Release notes |
| `GET` | `/api/v1/apk/:name` | APK download (with Range support) |
| `GET` | `/bin/*` | Binary file download |

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/auth/register` | Register new operator |
| `POST` | `/v1/auth/login` | Email/password login |
| `GET` | `/v1/auth/google` | Google OAuth redirect |
| `GET` | `/v1/auth/google/callback` | Google OAuth callback |
| `POST` | `/v1/auth/verify-email` | Verify email address |
| `POST` | `/v1/auth/resend-verification` | Resend verification email |
| `POST` | `/v1/auth/forgot-password` | Request password reset |
| `POST` | `/v1/auth/reset-password` | Reset password with token |
| `GET` | `/v1/auth/me` | Get current operator |
| `PATCH` | `/v1/auth/me` | Update name |
| `POST` | `/v1/auth/logout` | Logout |

### Device (HMAC or Auth Token)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/device/register` | Register device |
| `GET` | `/v1/device/:id/status` | Get device status |
| `PATCH` | `/v1/device/:id/fcm-token` | Update FCM token |
| `DELETE` | `/v1/device/:id` | Remove device |
| `POST` | `/v1/device/:id/command` | Send command |
| `GET` | `/v1/device/:id/stream` | WebSocket stream |

### Dashboard (Auth Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/dashboard/devices` | List all devices |

---

## Environment Variables

### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | HTTP listen port |
| `NODE_ENV` | `production` | Set to `development` to disable HMAC |
| `DATABASE_URL` | `./data/vyzorix.db` | SQLite file path |
| `VYZORIX_API_DIR` | `./data` | Version manifest directory |
| `VYZORIX_BIN_DIR` | `./bin` | APK storage directory |
| `VYZORIX_PUBLIC_DIR` | `./public` | Static assets directory |
| `JWT_SECRET` | — | JWT signing secret (generate with `openssl rand -hex 32`) |
| `JWT_DURATION_HOURS` | `168` | JWT expiry (7 days) |
| `ALLOWED_ORIGINS` | `*` | CORS origins (comma-separated) |

### Google OAuth

| Variable | Description |
|----------|-------------|
| `GOOGLE_OAUTH_CLIENT_ID` | From Google Cloud Console |
| `GOOGLE_OAUTH_CLIENT_SECRET` | From Google Cloud Console |
| `BASE_URL` | Your deployment URL |

### Email (Resend)

| Variable | Default | Description |
|----------|---------|-------------|
| `RESEND_API_KEY` | — | From Resend dashboard |
| `EMAIL_FROM` | `noreply@vyzorix.app` | Sender email |
| `EMAIL_FROM_NAME` | `Vyzorix` | Sender name |
| `EMAIL_VERIFY_TOKEN_EXPIRY_HOURS` | `24` | Verification link expiry |
| `PASSWORD_RESET_TOKEN_EXPIRY_MINUTES` | `60` | Reset link expiry |

### Firebase

| Variable | Description |
|----------|-------------|
| `FIREBASE_CREDENTIALS` | Full JSON of Firebase service account |

---

## Deployment

### Render (Recommended)

1. Connect your GitHub repository to [Render](https://render.com)
2. Use `render.yaml` as the blueprint
3. Add required environment variables
4. The `/data` disk persists SQLite across redeploys

> **Note:** Render hobby plan sleeps after 15-30 minutes. Use UptimeRobot to ping `/healthz` every 10 minutes.

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

---

## Manual Setup Guide

For step-by-step instructions on setting up Google OAuth, Resend, and Render, see [SETUP-GUIDE.md](./SETUP-GUIDE.md).

---

## Make Targets

```bash
make help          # Show all targets
make deps          # Download Go dependencies
make tidy          # Tidy go.mod
make build         # Build Go binary
make build-linux   # Cross-compile for Linux
make run           # Run server
make test          # Run tests with race detection
make test-coverage # Generate coverage report
make fmt           # Format code
make lint          # Lint code
make check         # fmt + lint + test
make docker-build  # Build Docker image
make docker-up     # Start Docker Compose
make docker-down   # Stop Docker Compose
make docker-logs   # View Docker logs
make init-data     # Create ./data/ with manifests
```

---

## Project Phases

| Phase | Server | Description |
|-------|--------|-------------|
| **Phase 1** | `cmd/mockserver/` | In-memory mock server |
| **Phase 1.5** | Root `main.go` | SQLite + WebSocket + FCM (current) |
| **Phase 2** | Same | Full dashboard integration |
| **Phase 3** | Future | Key rotation, multi-device, audit logging |

---

## Documentation

- [SETUP-GUIDE.md](./SETUP-GUIDE.md) — Third-party service setup
- [doc/REPO_TREE.md](./doc/REPO_TREE.md) — Full repository structure
- [doc/UPDATE_SERVER_ARCHITECTURE_SPEC.md](./doc/UPDATE_SERVER_ARCHITECTURE_SPEC.md) — Deep-dive architecture
- [doc/SYSTEM_MAP.md](./doc/SYSTEM_MAP.md) — System overview

---

## License

Private project. All rights reserved.