# Cross-tenant isolation probes

Live, end-to-end checks that one tenant cannot reach another tenant's data. They
talk to a running server over HTTP; they are not unit tests.

## Why these exist

A batch of GraphQL resolvers read the `organizationId` **argument** and passed it
straight to the service. The GraphQL middleware only validates membership for the
organization in the URL (`/:org/graphql`), so a caller authenticated for org A
could pass `organizationId=org B` and read or mutate org B's alert rules, contact
points, service accounts and annotations. The same class of bug affected
`orgId` / `deviceId` / `dispatchId` on the subscription path.

These probes are the regression guard. They are worth running against any change
that touches GraphQL resolvers or the subscription client.

## Run

```bash
# 1. Build and start the server on a throwaway SQLite DB.
cd apps/api
CGO_ENABLED=1 go build -o /tmp/vyzorix-api ./cmd/api
ENVIRONMENT=development DATABASE_BACKEND=sqlite \
  DATABASE_URL="/tmp/vyzdata/vyzorix.db?mode=rwc" \
  API_KEY_primary="vxyz-bootstrap-key-1234567890" \
  JWT_SECRET="dev-jwt-secret-32-chars-minimum!!" \
  SESSION_SECRET="dev-session-secret-32-chars-min!!" \
  SERVER_API_TOKEN="dev-server-token" PORT=3000 \
  ENABLE_GRAPHQL=true SSR_ENABLED=false ENFORCE_HMAC=false \
  /tmp/vyzorix-api &

# 2. Seed two tenants into the same DB (needs: pip install argon2-cffi).
python3 e2e/security/seed-tenants.py /tmp/vyzdata/vyzorix.db

# 3. Probe.
python3 e2e/security/cross-tenant-probe.py
```

Exit code 0 means every probe behaved correctly.

`SSR_ENABLED=false` is required: with the SSR proxy enabled, `/:org/graphql` is
proxied to the (absent) SSR server and returns 307 before reaching the API.

## Reading the results

GraphQL returns **HTTP 200 even for authorization failures** — the denial is in
the JSON `errors` array with `data` nulled. A probe passes when the response is a
denial *and* carries no data. The controls (each tenant reading its own org) must
succeed, which is what stops a blanket "deny everything" from looking like a pass.

## Negative control

Verified by reverting the resolver fix (`git checkout <pre-fix> -- ...resolver/query_resolver.go ...resolver/mutation_resolver.go`)
and re-running: 4/8 probes fail, with `{"data":{"serviceAccounts":[]}}` and friends
returned to the wrong tenant. With the fix in place, 8/8 pass.
