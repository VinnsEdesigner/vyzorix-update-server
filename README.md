<div align="center">

# Vyzorix Update Server

Real-time device management and OTA updates for VyzorixAudioRouter.

Go · Gin · SQLite · GraphQL · WebSocket · React

</div>

---

The server manages Android device fleets: device registration, signed command
dispatch, telemetry, OTA updates, alerting, and a multi-tenant organization
model. It exposes a REST API, a GraphQL API, and a WebSocket stream.

## Quick start

```bash
cd apps/api
make all      # install deps, build the Go binary and the frontend
make run      # start the server
```

The server listens on `PORT` (default `3000`). See `apps/api/Makefile` for every
target (`deps`, `build`, `test`, `lint`, `wire`, `dev`, ...).

Required environment variables (see `.env.example`):

- `DATABASE_URL` — SQLite path or Turso URL
- `JWT_SECRET`, `SESSION_SECRET` — 32+ characters each
- `SERVER_API_TOKEN` — protects `/healthz`
- at least one `API_KEY_*` — bootstrap API key
- `ALLOWED_ORIGINS` — your dashboard origin (empty means fail-closed)

## Layout

```
apps/
  api/              Go server (REST + GraphQL + WebSocket)
  VyzoriX_web/      React dashboard
  VyzoriX_mobile/   Android client
packages/           Shared TypeScript packages
DOCUMENTATION/      Feature and operations documentation
```

## Documentation

Start at [DOCUMENTATION/README.md](./DOCUMENTATION/README.md) — it indexes
everything below.

| Area | Documents |
|:-----|:----------|
| Architecture | [ARCHITECTURE](./DOCUMENTATION/ARCHITECTURE.md), [GLOSSARY](./DOCUMENTATION/GLOSSARY.md) |
| API | [GRAPHQL_API](./DOCUMENTATION/GRAPHQL_API.md), [ERROR_CODES](./DOCUMENTATION/ERROR_CODES.md) |
| Auth & access | [AUTHENTICATION](./DOCUMENTATION/AUTHENTICATION.md), [API_KEYS](./DOCUMENTATION/API_KEYS.md), [ORGANIZATIONS](./DOCUMENTATION/ORGANIZATIONS.md), [SERVICE_ACCOUNTS](./DOCUMENTATION/SERVICE_ACCOUNTS.md) |
| Devices & commands | [DEVICE_MANAGEMENT](./DOCUMENTATION/DEVICE_MANAGEMENT.md), [COMMAND_EXECUTION](./DOCUMENTATION/COMMAND_EXECUTION.md), [RISK_AND_CONFIRMATION](./DOCUMENTATION/RISK_AND_CONFIRMATION.md) |
| Realtime | [REALTIME_COMMUNICATION](./DOCUMENTATION/REALTIME_COMMUNICATION.md), [WEBSOCKET_SYSTEM](./DOCUMENTATION/WEBSOCKET_SYSTEM.md), [LIVE_CHANNELS](./DOCUMENTATION/LIVE_CHANNELS.md) |
| Updates | [UPDATE_SYSTEM](./DOCUMENTATION/UPDATE_SYSTEM.md) |
| Operations | [DEPLOYMENT](./DOCUMENTATION/DEPLOYMENT.md), [SECURITY_HARDENING](./DOCUMENTATION/SECURITY_HARDENING.md), [METRICS](./DOCUMENTATION/METRICS.md), [BACKGROUND_WORKERS](./DOCUMENTATION/BACKGROUND_WORKERS.md) |
| Specs & setup | [SETUP-GUIDE](./SETUP-GUIDE.md), [Docker](./apps/api/README_DOCKER.md), [documents/](./documents/) |

## Contributing

`make lint` and `make test` (from `apps/api`) must pass before a change is
proposed. The CI gate runs `go build`, `go vet`, `golangci-lint`, `go test`, and
the OpenAPI/GraphQL drift checks.
