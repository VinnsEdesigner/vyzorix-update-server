# GraphQL API Documentation

This directory contains documentation for the Vyzorix GraphQL API.

## Files

- [API.md](./API.md) - Complete GraphQL API reference
- [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md) - REST to GraphQL migration guide

## Quick Links

- [API Reference](./API.md)
- [Migration Guide](./MIGRATION_GUIDE.md)

## Architecture

The GraphQL API is built using:

- **Schema**: `apps/api/internal/api/graphql/schema/`
- **Resolvers**: `apps/api/internal/api/graphql/resolver/`
- **Handlers**: `apps/api/internal/api/graphql/handler/`
- **Subscriptions**: `apps/api/internal/api/graphql/subscription/`
- **Middleware**: `apps/api/internal/api/graphql/middleware/`
- **Errors**: `apps/api/internal/api/graphql/errors/`

## Frontend Integration

Frontend GraphQL hooks are in:

- **Client**: `apps/web/src/lib/api/graphql/client.ts`
- **Queries**: `apps/web/src/lib/api/graphql/queries.ts`
- **Mutations**: `apps/web/src/lib/api/graphql/mutations.ts`
- **Hooks**: `apps/web/src/lib/api/graphql/hooks.ts`
- **Components**: `apps/web/src/components/api/`

## Endpoints

| Endpoint | Purpose |
|----------|---------|
| `POST /graphql` | GraphQL queries and mutations |
| `GET /graphql` | GraphQL queries |
| `GET /playground` | Interactive GraphQL playground |
| `GET /graphql/ws` | GraphQL subscriptions (WebSocket) |
