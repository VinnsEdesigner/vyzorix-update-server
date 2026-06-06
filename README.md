<!--
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                          ║
║   ██████╗ ██╗      ██████╗ ██████╗ ███████╗████████╗                    ║
║   ██╔══██╗██║     ██╔═══██╗██╔══██╗██╔════╝╚══██╔══╝                    ║
║   ██████╔╝██║     ██║   ██║██████╔╝█████╗     ██║                       ║
║   ██╔══██╗██║     ██║   ██║██╔══██╗██╔══╝     ██║                       ║
║   ██████╔╝███████╗╚██████╔╝██║  ██║███████╗   ██║                       ║
║   ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝   ╚═╝                       ║
║                                                                          ║
║   UPDATE SERVER — Real-time Device Management & OTA Updates             ║
║                                                                          ║
╚══════════════════════════════════════════════════════════════════════════════╝
-->
<div align="center">

<!-- Badges Row -->
<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=black" alt="React">
  <img src="https://img.shields.io/badge/TypeScript-5-blue?style=for-the-badge&logo=typescript&logoColor=white" alt="TypeScript">
  <img src="https://img.shields.io/badge/SQLite-WAL%20Mode-003B57?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<!-- Title -->
# Vyzorix Update Server

### 🔌 Real-time Device Management & OTA Updates for VyzorixAudioRouter

**A production-ready C2 server for managing Android device fleets with WebSocket telemetry, FCM push notifications, and secure command dispatch.**

---

**[🚀 Quick Start](#-quick-start)** •
**[🏗️ Architecture](#-architecture)** •
**[📡 API Reference](#-api-reference)** •
**[🧪 Testing](#-testing)** •
**[📦 Deployment](#-deployment)** •
**[📚 Documentation](#-documentation)**

</div>

---

## ✨ Features

### Core Capabilities

| Feature | Description |
|:--------|:-------------|
| 📡 **WebSocket Streaming** | Real-time telemetry from devices with auto-reconnect |
| 📱 **OTA Updates** | Version manifest and APK distribution |
| 🔒 **HMAC Commands** | SHA256-signed commands per device |
| 🔔 **FCM Notifications** | Firebase Cloud Messaging integration |
| 👥 **Operator Auth** | JWT + Google OAuth authentication |
| 📊 **Dashboard API** | Device status, registration, management |
| 🌐 **CORS Security** | Configurable origin validation |
| 📈 **Rate Limiting** | Token-bucket per IP |

### Dashboard Pages

| Page | Description |
|:-----|:-------------|
| 📊 **Dashboard** | Real-time device status, risk score, thermal metrics |
| 📡 **Diagnostics** | Command panel with 8 device commands |
| 🔔 **Alerts** | Threshold-based alert derivation |
| 📦 **Updates** | Version manifest and APK download |
| ⚙️ **Settings** | Connection, operator, thresholds, notifications |

---

## 🏗️ Architecture

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

### Tech Stack

| Layer | Technology |
|:------|:------------|
| **Backend** | Go 1.22+, Gin web framework |
| **Database** | SQLite (WAL mode) |
| **WebSocket** | gorilla/websocket |
| **Push** | Firebase Admin SDK |
| **Auth** | JWT, Google OAuth (JWKS) |
| **Email** | Resend API |
| **Frontend** | React 18, TanStack Start, Vite |
| **Styling** | Tailwind CSS, shadcn/ui |
| **Testing** | Vitest (frontend), Go testing (backend) |
| **Linting** | ESLint, golangci-lint |
| **Security** | Gosec, Dependency Review |
| **Deployment** | Render (persistent disk) |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- SQLite

### 1. Clone the Repository

```bash
git clone https://github.com/VinnsEdesigner/vyzorix-update-server.git
cd vyzorix-update-server
```

### 2. Environment Setup

```bash
# Copy environment template
cp .env.example .env

# Edit with your configuration
nano .env
```

**Required variables:**

```env
PORT=3000
DATABASE_URL=./data/vyzorix.db
JWT_SECRET=your-jwt-secret-min-32-chars
TOKEN_SECRET=your-dashboard-token-secret
ENFORCE_HMAC=false
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

**Optional (for full features):**

```env
# Google OAuth
GOOGLE_OAUTH_CLIENT_ID=your-client-id
GOOGLE_OAUTH_CLIENT_SECRET=your-client-secret

# Email (Resend)
RESEND_API_KEY=your-resend-api-key
EMAIL_FROM=your@email.com
```

### 3. Run the Server

```bash
# Backend only
go run .

# With frontend dev server (another terminal)
cd src && npm install && npm run dev
```

### 4. Access Dashboard

Open [http://localhost:3000](http://localhost:3000) in your browser.

---

## 📡 API Reference

### Authentication Endpoints

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `POST` | `/v1/auth/login` | Operator login |
| `POST` | `/v1/auth/register` | Operator registration |
| `POST` | `/v1/auth/logout` | Operator logout |
| `GET` | `/v1/auth/me` | Get current operator |
| `PATCH` | `/v1/auth/me` | Update operator profile |
| `GET` | `/v1/auth/google` | Initiate Google OAuth |
| `GET` | `/v1/auth/google/callback` | OAuth callback |

### Device Endpoints

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `POST` | `/v1/device/register` | Register device |
| `GET` | `/v1/device/:id/status` | Get device status |
| `PATCH` | `/v1/device/:id/fcm-token` | Update FCM token |
| `POST` | `/v1/device/:id/command` | Dispatch command |
| `GET` | `/v1/device/:id/stream` | WebSocket stream |
| `DELETE` | `/v1/device/:id` | Deregister device |

### Dashboard Endpoints

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `GET` | `/v1/dashboard/devices` | List all devices (JWT required) |

### Update Endpoints

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `GET` | `/api/v1/version` | Get version manifest |
| `HEAD` | `/api/v1/apk/:filename` | Check APK size |

### Health

| Method | Endpoint | Description |
|:-------|:---------|:------------|
| `GET` | `/healthz` | Server health check (with DB verification) |

---

## 🧪 Testing

### Run All Tests

```bash
# Go backend tests
go test ./...

# Go with coverage
go test -coverprofile=coverage.out ./...

# Frontend tests
cd src && npx vitest run

# Frontend with coverage
cd src && npx vitest run --coverage
```

### Test Results

| Suite | Status |
|:------|:-------|
| Go Tests (12 packages) | ✅ All passing |
| Vitest Tests (79 tests) | ✅ All passing |
| Build | ✅ Successful |

---

## 📦 Deployment

### Render (Recommended)

1. Connect your GitHub repo to [Render](https://render.com)
2. Use `render.yaml` as the blueprint
3. Set environment variables in Render dashboard
4. The `/data` disk persists SQLite across redeploys

### Docker

```bash
# Build image
docker build -t vyzorix-update-server .

# Run container
docker run -p 3000:3000 \
  -e DATABASE_URL=/data/vyzorix.db \
  -e JWT_SECRET=your-secret \
  -v ./data:/data \
  vyzorix-update-server
```

---

## 📚 Documentation

| Document | Description |
|:---------|:------------|
| [Architecture Spec](./doc/UPDATE_SERVER_ARCHITECTURE_SPEC.md) | Deep-dive into server internals |
| [Repo Tree](./doc/REPO_TREE.md) | Complete file structure |
| [API Reference](./doc/UPDATE_SERVER.md) | All REST endpoints |
| [Device Registration](./doc/DEVICE_REGISTRATION.md) | Device lifecycle |
| [Command Security](./doc/COMMAND_SECURITY.md) | HMAC signing details |
| [Frontend Bug Fixes](./doc/FRONTEND_BUG_FIXES.md) | Frontend fixes (FE-1 to FE-4) |
| [Backend Bug Fixes](./doc/BACKEND_BUG_FIXES.md) | Backend fixes (#2-15) |
| [Setup Guide](./SETUP-GUIDE.md) | Google OAuth, Resend, Render setup |

---

## 🛠️ Development

### Project Structure

```
vyzorix-update-server/
├── cmd/
│   └── mockserver/           # Phase 1 mock server
├── controllers/               # HTTP handlers
├── hub/                      # WebSocket broker
├── middleware/               # Auth, CORS, rate limit
├── models/                   # Type definitions
├── security/                # JWT, HMAC, password
├── services/                # FCM, email
├── storage/                  # SQLite operations
├── src/                      # React frontend
│   ├── lib/                  # API clients, config
│   ├── hooks/                # Custom React hooks
│   ├── components/          # UI components
│   └── routes/               # Page routes
└── doc/                      # Documentation
```

### Available Scripts

```bash
# Backend
go build -o bin/server .       # Build server
go run .                       # Run server
go test ./...                  # Run tests
golangci-lint run             # Lint

# Frontend
cd src && npm install          # Install deps
npm run dev                    # Dev server
npm run build                  # Production build
npm run lint                   # Lint
npx vitest run                 # Run tests
```

---

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [TanStack](https://tanstack.com/)
- [shadcn/ui](https://ui.shadcn.com/)
- [Render](https://render.com/)
- [Firebase](https://firebase.google.com/)

---

<div align="center">

**[↑ Back to top](#vyzorix-update-server)**

development by [VinnsEdesigner](https://github.com/VinnsEdesigner)

