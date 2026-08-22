# API Client Codegen (swaggo + orval)

## Overview

Go handlers carry `swaggo` comments (`// @Summary`, `@Param`, `@Success`, `@Router`) as the single source of truth. From them:

1. `swag init` → `apps/api/swag/swagger.json` (Swagger 2.0)
2. `python3 tooling/codegen/build_openapi3.py` → `apps/api/swag/openapi3.json` (OpenAPI 3.0 normalized for orval)
3. `orval` → `packages/API_Client/src/generated/` (split per tag)

The generated TypeScript SDK replaces hand-rolled REST layers. Domain modules (`packages/API_Client/src/domain/<ctx>/`) become thin re-exports + mappers over generated types.

## Layout

```
packages/API_Client/src/generated/
├── {tag}.ts                   ← endpoints for one tag (alerts, service-accounts, ...)
├── {tag}/                     ← (internal to that tag)
├── rest-bridge.ts             ← customAxios mutator (shared auth/org scoping)
└── vyzorixUpdateServerAPI.schemas.ts   ← shared types + union constants

packages/API_Client/src/domain/<ctx>/
└── index.ts                   ← re-exports + snake→camel mappers
```

## Adding a new endpoint group

1. **Annotate the handler** in Go:

```go
// List handles GET /v1/contact-points.
// @Summary      List contact points
// @Tags         contact-points
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  object  "contact points"
// @Router       /contact-points [get]
func (h *Handler) List(c *gin.Context) {
```

2. **Regenerate the spec** (server only):
```bash
cd apps/api && swag init -g cmd/api/api_main.go --parseDependency --output ./swag
```

3. **Regenerate the SDK** (typescript):
```bash
node node_modules/.pnpm/orval*/node_modules/orval/dist/bin/orval.mjs --config orval.config.js
```

4. **Extend `openapi3.json`** (`tooling/codegen/build_openapi3.py`) with any new request/response schemas (`components.schemas`), then re-run step 3.

5. **Update the domain module** to re-export the new generated types + add mappers.

6. **Verify**:
```bash
pnpm --filter ./packages/API_Client exec tsc --noEmit
pnpm --filter @vyzorix/web run build
```

## Drift gate

`tooling/hooks/pre-commit` runs `swag init` to a temp dir and compares `swagger.json`; commit fails if the Go annotations moved and the spec is stale. Regenerate before commit.

## Notes

- Path parameters must be `required: true` in the spec (OpenAPI 3 hard requirement).
- Body params are consumed via `requestBody` in OpenAPI 3 — swag outputs Swagger 2.0 so the conversion script normalizes.
- Field names in the web app match Go `json:"..."` tags directly (snake_case) after codegen — no camelCase shim in the generated SDK.
- Tag names come from `// @Tags <name>`; each tag emits one file.
