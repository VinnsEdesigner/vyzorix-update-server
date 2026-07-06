# Frontend Structure Blueprint

This directory contains the complete directory structure for the frontend application, as defined in `documents/FRONTEND_ARCHITECTURE.md`.

## Purpose

This is a **blueprint only** - no implementation files exist here. Use this structure as a reference when implementing features.

## Directory Overview

```
src/
├── domain/           # Pure business logic (types, mappers, validators)
├── lib/api/         # Data layer (GraphQL, REST, WebSocket clients)
├── hooks/           # Presentation layer (React hooks)
└── components/      # UI layer (React components)
```

## Feature Directories

| Feature | Domain | GraphQL | REST | WebSocket |
|---------|--------|---------|------|-----------|
| API Keys | `domain/apikey/` | `lib/api/graphql/apikey/` | `lib/api/rest/apikey/` | - |
| Settings | `domain/settings/` | `lib/api/graphql/settings/` | `lib/api/rest/settings/` | - |
| Updates | `domain/updates/` | `lib/api/graphql/updates/` | `lib/api/rest/updates/` | - |
| Diagnostics | `domain/diagnostics/` | `lib/api/graphql/diagnostics/` | `lib/api/rest/diagnostics/` | - |
| Commands | `domain/commands/` | `lib/api/graphql/commands/` | `lib/api/rest/commands/` | - |
| Logs | `domain/logs/` | `lib/api/graphql/logs/` | `lib/api/rest/logs/` | - |
| Registration | `domain/registration/` | `lib/api/graphql/registration/` | `lib/api/rest/registration/` | - |
| Realtime | `domain/realtime/` | `lib/api/graphql/realtime/` | `lib/api/rest/realtime/` | `lib/api/websocket/realtime/` |
| Devices | `domain/devices/` | `lib/api/graphql/devices/` | `lib/api/rest/devices/` | - |
| Telemetry | `domain/telemetry/` | `lib/api/graphql/telemetry/` | `lib/api/rest/telemetry/` | - |
| Alerts | `domain/alerts/` | `lib/api/graphql/alerts/` | `lib/api/rest/alerts/` | - |
| Export | `domain/export/` | - | - | - |

## Shared Directories

| Directory | Purpose |
|----------|---------|
| `domain/_shared/` | Cross-feature types (pagination, errors) |
| `lib/api/graphql/_shared/` | GraphQL client setup |
| `lib/api/rest/_shared/` | REST client setup |
| `hooks/_shared/` | Cross-feature hooks |
| `components/ui/` | Base UI components |
| `components/layout/` | Layout components |

## Filename Conventions

### Domain Layer
- `{feature}-entity.ts` - Types, interfaces
- `{feature}-mappers.ts` - Transformations
- `{feature}-validators.ts` - Validation
- `{feature}-constants.ts` - Constants

### Data Layer
- `graphql-{feature}-queries.ts` - GraphQL queries
- `graphql-{feature}-mutations.ts` - GraphQL mutations
- `graphql-{feature}-types.ts` - Raw response types
- `rest-{feature}-endpoints.ts` - REST endpoints

### Hooks
- `use-{feature}.ts` - Single resource
- `use-{feature}-list.ts` - List with pagination
- `use-{feature}-create.ts` - Create mutation
- `use-{feature}-update.ts` - Update mutation
- `use-{feature}-delete.ts` - Delete mutation

## Reference

See `documents/CROSS_SPEC_FILE_MAPPING.md` for complete file listings per spec.
