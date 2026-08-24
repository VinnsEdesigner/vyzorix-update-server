# GraphQL API

The server has a GraphQL endpoint alongside the REST API. It's used by the dashboard for complex queries that would require multiple REST calls.

## Endpoint

`POST /:org/graphql` — the org ID is in the URL path, not the query. Requires operator auth (session cookie or API key).

`GET /:org/playground` — GraphQL Playground UI for development (disabled in production via `GIN_MODE=release`).

`GET /:org/graphql/ws` — WebSocket subscriptions for realtime data.

## Auth

GraphQL routes go through the same middleware chain as REST: cookie auth, API key auth, org context, membership check, request signing. The operator is extracted from the gin context and passed into the GraphQL resolver context via `gqlcontext`.

## Schema

The schema is code-first, built from Go types in `internal/api/graphql/schema/`. It includes:

- **Queries** — devices, device details, command history, telemetry, update versions, organization info
- **Mutations** — execute command, cancel command, push update, create API key, update settings
- **Subscriptions** — realtime telemetry updates, command status changes, device connection events

The schema types are defined in `internal/api/graphql/schema/objects.go` and `internal/api/graphql/schema/schema.go`.

### Schema generation + drift gate

The canonical SDL is generated from the code-first schema and checked in at
`apps/api/swag/graphql/schema.graphql`. Regenerate it after any change to
`internal/api/graphql/schema/`:

```
cd apps/api && go run ./cmd/graphql-schema -out swag/graphql
```

The generator (`apps/api/cmd/graphql-schema`) builds the schema with a nil
resolver (BuildSchema only reads field/type definitions, never invokes
resolvers) and renders SDL directly from the type system — graphql-go has no
SDL printer and its executor does not execute the standard introspection
`__schema` query. Output is deterministic (types, fields, args, and enum values
are sorted).

Drift is enforced two ways (same contract as the swagger drift gate):
- **Pre-commit** (`tooling/hooks/pre-commit`): regenerates to a temp dir and
  fails the commit if `schema.graphql` is stale.
- **Server Gate CI** (`.github/workflows/server-gate.yml`): the
  "GraphQL schema drift" step does the same on every push to main that touches
  `apps/api/`.

**Annotating the schema** = editing the `Description:` strings on the Go
objects/fields/enums in `internal/api/graphql/schema/*.go`. They flow into the
generated SDL as `"""..."""` doc comments. Every named type must have a
description (the generator output is checked for this in review).

## Resolvers

Resolvers are in `internal/api/graphql/resolver/`. They call the same application services as the REST handlers — no duplicated business logic. The resolver gets the operator from `gqlcontext.GetOperator(ctx)` and the org ID from `gqlcontext.GetOrganizationID(ctx)`.

### Subscription handler

WebSocket subscriptions are handled by `internal/api/graphql/subscription/subscription_handler.go`. It upgrades the connection and manages subscription lifecycle. Clients send `subscribe` messages with a query and receive `next` messages with data.

## Introspection

GraphQL introspection (`__schema`, `__type`) is blocked without authentication. The `POST /:org/graphql` endpoint requires a valid session cookie — unauthenticated requests get 401.

## Rate limiting

GraphQL routes are rate-limited via the same `authLimiter` middleware as REST routes. Complex queries (deeply nested, multiple selections) are bounded by the GraphQL depth limit configured in the schema.
