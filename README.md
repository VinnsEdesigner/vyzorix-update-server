<!--

                                                                          
                                
                            
                                       
                                       
                               
                                  
                                                                          
   UPDATE SERVER — Real-time Device Management & OTA Updates             
                                                                          

-->
<div align="center">

<!-- Badges Row -->
<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=black" alt="React">
  <img src="https://img.shields.io/badge/TypeScript-5-blue?style=for-the-badge&logo=typescript&logoColor=white" alt="TypeScript">
  <img src="https://img.shields.io/badge/SQLite-WAL%20Mode-003B57?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<!-- Title -->
# Vyzorix Update Server

### Real-time Device Management & OTA Updates for VyzorixAudioRouter

**A production-ready C2 server for managing Android device fleets with WebSocket telemetry, FCM push notifications, and secure command dispatch.**

---

**[Quick Start](#-quick-start)** •
**[Architecture](#-architecture)** •
**[API Reference](#-api-reference)** •
**[Testing](#-testing)** •
**[Deployment](#-deployment)** •
**[Documentation](#-documentation)**

</div>

---

## Features

### Core Capabilities

| Feature | Description |
|:--------|:-------------|
| **WebSocket Streaming** | Real-time telemetry from devices with auto-reconnect |
| **OTA Updates** | Version manifest and APK distribution |
| **HMAC Commands** | SHA256-signed commands per device |
| **FCM Notifications** | Firebase Cloud Messaging integration |
| **Multi-Tenant Organizations** | Team-based access with role-based permissions |
| **Operator Auth** | JWT + Google OAuth + MFA authentication |
| **GraphQL API** | Comprehensive GraphQL API for all operations |
| **REST API** | RESTful endpoints for device management |
| **CORS Security** | Configurable origin validation |
| **Rate Limiting** | Token-bucket per IP |

### Dashboard Pages

| Page | Description |
|:-----|:-------------|
| **Dashboard** | Real-time device status, risk score, thermal metrics |
| **Diagnostics** | Command panel with 8 device commands |
| **Alerts** | Threshold-based alert derivation |
| **Updates** | Version manifest and APK download |
| **Settings** | Connection, operator, thresholds, notifications |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         VYZORIX UPDATE SERVER                               │
└─────────────────────────────────────────────────────────────────────────────┘

                              ┌─────────────┐
                              │   Clients   │
                              │ (Android)   │
                              └──────┬──────┘
                                     │
                                     │ HTTP/HTTPS + WebSocket
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           GIN HTTP SERVER (Go)                              │
│                              Port 3000                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│  ROUTES                                                                   │
│  ┌─────────────┬──────────────┬────────────┬──────────┬────────────────┐  │
│  │ /v1/auth/* │ /v1/*        │ /graphql/* │ /api/*   │ /healthz       │  │
│  │ Login       │ Devices      │ Queries    │ Updates  │ Server Health  │  │
│  │ Register    │ Dashboard    │ Mutations  │ APK      │                │  │
│  │ MFA         │ Commands     │ Subscript. │ Version  │                │  │
│  │ OAuth       │ Telemetry    │            │          │                │  │
│  │ Sessions    │ Metrics      │            │          │                │  │
│  │ API Keys    │ Logs         │            │          │                │  │
│  │ Orgs        │ Diagnostics  │            │          │                │  │
│  │ Members     │ Events       │            │          │                │  │
│  └─────────────┴──────────────┴────────────┴──────────┴────────────────┘  │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        MIDDLEWARE STACK                               │   │
│  │  Request ID │ Logger │ CORS │ Security │ Body Limit │ Disable Trace │   │
│  │  Rate Limiter │ HMAC Signing │ JWT Auth │ Org Context │ API Key Auth │   │
│  │  Organization Membership │ SSR Proxy │ Turnstile │ CSRF │ Lockout    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
         ┌───────────────────────────┼───────────────────────────┐
         │                           │                           │
         ▼                           ▼                           ▼
┌─────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│  WebSocket Hub  │     │   REST Handlers      │     │   GraphQL Server    │
│                 │     │                     │     │                     │
│ • Device Stream │     │ • Auth Handlers     │     │ • Query Resolvers   │
│ • Telemetry     │     │ • Device Handlers   │     │ • Mutation Resolver  │
│ • Subscriptions │     │ • Command Handlers  │     │ • Subscription Res. │
│ • Message Queue │     │ • Org Handlers     │     │ • Schema Types      │
│ • Rate Limiting │     │ • Dashboard API    │     │                     │
└────────┬────────┘     │ • Updates Handler │     │                     │
         │               │ • Diagnostics API  │     │                     │
         │               │ • Metrics/Logs API │     │                     │
         │               └─────────┬─────────┘     └─────────────────────┘
         │                         │
         └─────────────────────────┼─────────────────────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         APPLICATION LAYER                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  Services                                                                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐  │
│  │ AuthService  │ │DeviceService │ │OrgService    │ │ CommandService   │  │
│  │              │ │              │ │              │ │                  │  │
│  │ • Login      │ │ • Register   │ │ • Create    │ │ • Dispatch      │  │
│  │ • Register   │ │ • Status     │ │ • Members   │ │ • Status        │  │
│  │ • MFA        │ │ • Settings   │ │ • Invite    │ │ • Retry/Cancel  │  │
│  │ • OAuth      │ │ • Transfer   │ │ • Settings  │ │ • History       │  │
│  │ • Sessions   │ │ • Deregister │ │              │ │                  │  │
│  │ • API Keys   │ │ • Events     │ │              │ │                  │  │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────────┘  │
│                                                                             │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐  │
│  │MetricsService│ │UpdatesService│ │DiagService  │ │ EventProcessor   │  │
│  │              │ │              │ │              │ │                  │  │
│  │ • Aggregate  │ │ • Versions   │ │ • Inspect   │ │ • Broadcast     │  │
│  │ • Timeline   │ │ • Changelog │ │ • Timeline  │ │ • Telemetry     │  │
│  │ • Stats      │ │ • APK Dist   │ │ • Commands  │ │ • Notifications │  │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DOMAIN LAYER                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│  Entities                        Repositories                              │
│  ┌─────────────┐ ┌─────────────┐  ┌─────────────────────────────────────┐│
│  │ Operator    │ │ Organization│  │ Device │ Command │ Operator │ Session ││
│  │ Membership  │ │ Invitation  │  │ APIKey │ Client  │ Inbox   │ Logs   ││
│  │ Device      │ │ APIKey      │  │ Metrics│ Updates │ Telemetry│ Org   ││
│  │ Command     │ │ ClientCreds │  └─────────────────────────────────────┘│
│  │ DeviceSettings│ │ OrgSettings│                                        │
│  │ InboxEntry  │ │ Threshold   │                                          │
│  └─────────────┘ └─────────────┘                                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      INFRASTRUCTURE LAYER                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────────┐    │
│  │ Storage  │ │  Config  │ │ Logging  │ │  Email   │ │      FCM       │    │
│  │ SQLite   │ │  Env     │ │ slog     │ │ Resend   │ │ Firebase      │    │
│  │ WAL Mode │ │  .env    │ │          │ │ SMTP     │ │ Cloud Message  │    │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └────────────────┘    │
│                                                                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐                      │
│  │  Crypto  │ │  OAuth   │ │ Webhook  │ │   SSR    │                      │
│  │ HMAC/SHA │ │ GitHub   │ │ HTTP     │ │ TanStack │                      │
│  │ JWT/Keys │ │ Google   │ │          │ │ Start    │                      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘                      │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Tech Stack

| Layer | Technology |
|:------|:------------|
| **Backend** | Go 1.26, Gin web framework, Wire DI |
| **Database** | SQLite (WAL mode) with auto-migrations |
| **WebSocket** | gorilla/websocket |
| **GraphQL** | graphql-go library |
| **Push** | Firebase Admin SDK |
| **Auth** | JWT, Google OAuth (JWKS), TOTP MFA |
| **Email** | SMTP support via Resend API |
| **Frontend** | React 18, TanStack Start, Vite |
| **Styling** | Tailwind CSS, shadcn/ui |
| **Testing** | Vitest (frontend), Go testing (backend) |
| **Linting** | ESLint, golangci-lint v2 |
| **Container** | Docker, Docker Compose |
| **Deployment** | Render (persistent disk) |

---

## Quick Start

### Prerequisites

- Go 1.26+
- Node.js 20+
- pnpm (for frontend)
- Docker (optional)

### 1. Clone the Repository

```bash
git clone https://github.com/VinnsEdesigner/vyzorix-update-server.git
cd vyzorix-update-server
```

### 2. Setup and Run

```bash
cd apps/api

# Install dependencies and build
make all

# Or step by step:
make deps
make build

# Run the server
make run
```

### 3. Access Dashboard

Open [http://localhost:3000](http://localhost:3000) in your browser.

---

## Makefile Commands

The API server includes a comprehensive Makefile for development:

| Command | Description |
|:--------|:-------------|
| `make all` | Install deps, tidy, build, and frontend |
| `make deps` | Download Go dependencies |
| `make tidy` | Clean up go.mod |
| `make build` | Build the Go binary |
| `make run` | Run the server |
| `make test` | Run tests with race detector |
| `make test-quick` | Fast tests without race detector |
| `make lint` | Run all linters |
| `make wire` | Generate wire dependency injection |
| `make dev` | Start development server |
| `make docker-build` | Build Docker image |
| `make docker-up` | Start with Docker Compose |

---

## Multi-Tenant Organization Model

The server supports multi-tenant organizations with role-based access control:

### Organization Roles

| Role | Permissions |
|:-----|:------------|
| `super_admin` | Full access, can manage members, transfer ownership |
| `admin` | Can manage devices, send commands, manage settings |
| `operator` | Can view devices and telemetry |
| `viewer` | Read-only access to assigned devices |

### Key Features

- **Organization-scoped access**: All API operations are scoped to organizations
- **Invitation system**: Invite members via email
- **API Keys**: Per-operator API keys for programmatic access
- **Audit logging**: Track all operations with audit logs

---

## API Reference

### REST API

#### Authentication Endpoints

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `POST` | `/v1/auth/login` | Operator login |
| `POST` | `/v1/auth/register` | Operator registration |
| `POST` | `/v1/auth/logout` | Operator logout |
| `GET` | `/v1/auth/me` | Get current operator |
| `PATCH` | `/v1/auth/me` | Update operator profile |
| `POST` | `/v1/auth/mfa/enroll` | Enroll MFA |
| `POST` | `/v1/auth/mfa/verify` | Verify MFA code |
| `GET` | `/v1/auth/google` | Initiate Google OAuth |
| `GET` | `/v1/auth/google/callback` | OAuth callback |

#### Device Endpoints

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `POST` | `/v1/device/register` | Register device |
| `GET` | `/v1/device/:id/status` | Get device status |
| `PATCH` | `/v1/device/:id/fcm-token` | Update FCM token |
| `POST` | `/v1/device/:id/command` | Dispatch command |
| `GET` | `/v1/device/:id/stream` | WebSocket stream |
| `DELETE` | `/v1/device/:id` | Deregister device |

#### Organization Endpoints

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `POST` | `/v1/organizations` | Create organization |
| `GET` | `/v1/organizations` | List my organizations |
| `GET` | `/v1/organizations/:id` | Get organization |
| `POST` | `/v1/organizations/:id/invite` | Invite member |
| `POST` | `/v1/organizations/:id/join` | Accept invitation |

#### API Key Endpoints

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `POST` | `/v1/api-keys` | Create API key |
| `GET` | `/v1/api-keys` | List API keys |
| `GET` | `/v1/api-keys/:id` | Get API key |
| `PATCH` | `/v1/api-keys/:id` | Update API key |
| `DELETE` | `/v1/api-keys/:id` | Revoke API key |
| `POST` | `/v1/api-keys/:id/rotate` | Rotate API key |

### GraphQL API

The server provides a comprehensive GraphQL API at `/graphql`:

```bash
# Query devices
curl -X POST http://localhost:3000/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: vyz_session=<session>" \
  -d '{"query": "query { devices(organizationId: \"org-123\") { id name online } }"}'
```

See [GraphQL README](./apps/api/internal/api/graphql/README.md) for full documentation.

### Health Check

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `GET` | `/healthz` | Server health check (with DB verification) |

---

## Testing

### Run All Tests

```bash
cd apps/api

# Go backend tests (with race detector)
make test

# Fast tests without race detector
make test-quick

# With coverage
make test-coverage
```

---

## Deployment

### Docker

```bash
# Build and run
docker build -t vyzorix-update-server:latest -f apps/api/Dockerfile .
docker run -p 3000:3000 \
  -e JWT_SECRET=your-secret \
  -e TOKEN_SECRET=your-secret \
  -v ./data:/data \
  vyzorix-update-server:latest

# Or with Docker Compose
docker compose up -d
```

See [Docker README](./apps/api/README_DOCKER.md) for full deployment guide.

### Render (Recommended)

1. Connect your GitHub repo to [Render](https://render.com)
2. Use `render.yaml` as the blueprint
3. Set environment variables in Render dashboard
4. The `/data` disk persists SQLite across redeploys

---

## Documentation

| Document | Description |
|:---------|:------------|
| [Architecture Spec](./doc/UPDATE_SERVER_ARCHITECTURE_SPEC.md) | Deep-dive into server internals |
| [Repo Tree](./doc/REPO_TREE.md) | Complete file structure |
| [API Reference](./doc/UPDATE_SERVER.md) | All REST endpoints |
| [Device Registration](./doc/DEVICE_REGISTRATION.md) | Device lifecycle |
| [Command Security](./doc/COMMAND_SECURITY.md) | HMAC signing details |
| [Setup Guide](./SETUP-GUIDE.md) | Google OAuth, Resend, Render setup |
| [Docker Guide](./apps/api/README_DOCKER.md) | Docker deployment |
| [GraphQL API](./apps/api/internal/api/graphql/README.md) | GraphQL API reference |

---

## Project Structure

```
vyzorix-update-server/
├── apps/
│   ├── api/                    # Go backend server
│   │   ├── cmd/api/           # Main entry point
│   │   ├── internal/
│   │   │   ├── api/           # HTTP handlers, middleware, GraphQL
│   │   │   ├── application/   # Business logic services
│   │   │   ├── audit/         # Audit logging
│   │   │   ├── domain/        # Domain models and interfaces
│   │   │   ├── infrastructure/# DB, email, Firebase, etc.
│   │   │   └── ws/            # WebSocket hub
│   │   ├── Makefile
│   │   └── README_DOCKER.md
│   ├── web/                    # React frontend
│   └── VyzoriX_mobile/        # Android client
├── packages/                   # Shared packages
│   ├── types/                 # TypeScript types
│   ├── config/                # Shared config
│   └── ui/                    # Shared UI components
├── tooling/                   # Build and deployment tools
└── doc/                       # Documentation
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [TanStack](https://tanstack.com/)
- [shadcn/ui](https://ui.shadcn.com/)
- [Render](https://render.com/)
- [Firebase](https://firebase.google.com/)

---

<div align="center">

**[↑ Back to top](#vyzorix-update-server)**

development by [VinnsEdesigner](https://github.com/VinnsEdesigner)
