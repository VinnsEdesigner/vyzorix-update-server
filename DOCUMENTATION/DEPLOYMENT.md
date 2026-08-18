# Deployment

## Render (managed — no Nginx needed)

Render handles TLS, HTTP/2, and health checks. You don't need Nginx.

1. Push the repo to GitHub
2. Go to Render → New → Web Service → connect the repo
3. Render reads `render.yaml` — review the env vars
4. Set these secrets in the Render dashboard:
   - `JWT_SECRET` — `openssl rand -hex 32`
   - `SESSION_SECRET` — `openssl rand -hex 32`
   - `TOKEN_SECRET` — `openssl rand -hex 32`
   - `DEVICE_SECRET` — `openssl rand -hex 32`
   - `SERVER_API_TOKEN` — `openssl rand -hex 32`
   - `API_KEY_primary` — your primary API key
   - `TURSO_DB_URL` — your Turso database URL
   - `TURSO_AUTH_TOKEN` — your Turso auth token
5. Set `ALLOWED_ORIGINS` to your dashboard URL (never `*`)
6. Deploy

## VPS with Docker Compose + Nginx

```
Internet → [Nginx :443 TLS] → [Go API :3000 (internal)]
```

### Setup

```bash
ssh root@your-vps
apt update && apt install -y docker.io docker-compose certbot
git clone https://github.com/VinnsEdesigner/vyzorix-update-server.git
cd vyzorix-update-server
```

### Configure

Edit `tooling/nginx/vyzorix.conf` — replace `api.vyzorix.com` with your domain.

Edit `docker-compose.yml` — replace all `change-me` values with real secrets:
```bash
openssl rand -hex 32  # use for JWT_SECRET, SESSION_SECRET, etc.
```

### TLS (first time)

```bash
docker compose up -d vyzorix          # start API first
mkdir -p /var/www/html
certbot certonly --webroot -w /var/www/html -d your-domain.com
echo "0 3 * * * certbot renew --quiet --post-hook 'docker compose restart nginx'" | crontab -
```

### Start

```bash
docker compose up -d
curl https://your-domain.com/health  # should return {"status":"ok"}
```

### Redeploy

```bash
git pull
docker compose up -d --build vyzorix  # Nginx stays up, only Go restarts
```

## VPS bare metal (no Docker)

```bash
apt install -y nginx
cp tooling/nginx/vyzorix.conf /etc/nginx/sites-available/vyzorix
ln -s /etc/nginx/sites-available/vyzorix /etc/nginx/sites-enabled/vyzorix
certbot --nginx -d your-domain.com
cd apps/api && go build -o /usr/local/bin/vyzorix-api ./cmd/api
# create /etc/vyzorix/.env with your secrets
# create systemd service (see DEPLOYMENT_GUIDE.md)
systemctl start vyzorix && systemctl reload nginx
```

## Nginx config

The config at `tooling/nginx/vyzorix.conf` provides:

- TLS 1.2/1.3 with Mozilla-recommended ciphers
- OCSP stapling
- Rate limiting: 10 req/sec per IP, burst 20
- Connection limiting: max 10 per IP
- HTTP/2
- Body size limit: 10MB
- WebSocket proxy (for `/v1/device/:imei/stream`)
- HSTS, security headers
- Health check bypass (no rate limit on `/health`)
- Hidden file blocking (`.env`, `.git`)

## Environment variables

### Required

| Var | Purpose | Example |
|-----|---------|---------|
| `NODE_ENV` | Environment mode | `production` |
| `PORT` | Listen port | `3000` (Render) or `12000` |
| `DATABASE_URL` | SQLite path or Turso URL | `file:/data/vyzorix.db` |
| `DATABASE_BACKEND` | `sqlite` or `turso` | `sqlite` |
| `JWT_SECRET` | JWT signing key (32+ chars) | `openssl rand -hex 32` |
| `SESSION_SECRET` | Session encryption key (32+ chars) | `openssl rand -hex 32` |
| `DEVICE_SECRET` | Device command signing (32+ chars) | `openssl rand -hex 32` |
| `API_KEY_primary` | Primary API key | `vxyz-prod-...` |
| `SERVER_API_TOKEN` | Server API token | `openssl rand -hex 32` |
| `ALLOWED_ORIGINS` | CORS origins (never `*`) | `https://your-dashboard.com` |

### Optional

| Var | Default | Purpose |
|-----|---------|---------|
| `ENFORCE_HMAC` | `true` in production | Require HMAC on device routes |
| `SSR_ENABLE` | `true` | Server-side rendering |
| `JWT_DURATION_HOURS` | `24` | Session/token lifetime |
| `AUTH_RATE_LIMIT_MIN` | `10` | Login attempts per minute |
| `RATE_LIMIT_PER_MIN` | `60` | General rate limit |
| `TURSO_DB_URL` | — | Turso database URL |
| `TURSO_AUTH_TOKEN` | — | Turso auth token |
| `RESEND_API_KEY` | — | Email service key |
| `FIREBASE_CREDENTIALS` | — | FCM service account JSON |
| `GOOGLE_OAUTH_CLIENT_ID` | — | Google OAuth |
| `GOOGLE_OAUTH_CLIENT_SECRET` | — | Google OAuth |
| `GITHUB_OAUTH_CLIENT_ID` | — | GitHub OAuth |
| `GITHUB_OAUTH_CLIENT_SECRET` | — | GitHub OAuth |
| `GITHUB_WEBHOOK_SECRET` | — | GitHub webhook verification |
| `VYZORIX_DOCS_BASE_URL` | `https://docs.vyzorix.com/errors` | Base URL for error docs links |

## SQLite vs Turso

**SQLite** (local/VPS): `DATABASE_URL=file:/data/vyzorix.db`, `DATABASE_BACKEND=sqlite`. The DB file lives on the mounted disk. Simple, no network, single instance.

**Turso** (Render/production): `DATABASE_URL=<turso-url>`, `DATABASE_BACKEND=turso`, plus `TURSO_DB_URL` and `TURSO_AUTH_TOKEN`. Turso is libSQL over HTTP — distributed SQLite with a primary/replica model. Use this for multi-region or when you need the DB to survive Render redeployments without a persistent disk.

Migrations run automatically on both backends. The `runMigrations` function works on any `*sql.DB` connection.
