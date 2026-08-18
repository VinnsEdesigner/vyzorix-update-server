# Vyzorix Deployment Guide

## Which deployment target are you using?

### Option A: Render (managed TLS — no Nginx needed)

Render terminates TLS at their load balancer. You do **not** need Nginx.

**Steps:**
1. Push this repo to GitHub.
2. Go to Render → New → Web Service → connect the repo.
3. Render auto-detects `render.yaml` — review the env vars.
4. Set these **required** secrets in the Render dashboard (mark as secret):
   - `JWT_SECRET` — `openssl rand -hex 32`
   - `SESSION_SECRET` — `openssl rand -hex 32`
   - `TOKEN_SECRET` — `openssl rand -hex 32`
   - `DEVICE_SECRET` — `openssl rand -hex 32`
   - `SERVER_API_TOKEN` — `openssl rand -hex 32`
   - `API_KEY_primary` — your primary API key
   - `TURSO_DB_URL` — your Turso database URL
   - `TURSO_AUTH_TOKEN` — your Turso auth token
5. Set `ALLOWED_ORIGINS` to your dashboard URL:
   - `https://your-app.onrender.com` (or your custom domain)
6. Deploy. Render handles TLS, HTTP/2, and health checks automatically.

**CORS on Render**: Set `ALLOWED_ORIGINS=https://your-dashboard.onrender.com` in the Render env vars. Never use `*`.

---

### Option B: VPS with Docker Compose + Nginx (self-managed TLS)

This is for DigitalOcean, Hetzner, AWS EC2, or any VPS where you run Docker.

**Architecture:**
```
Internet → [Nginx :443 TLS] → [Go API :3000 (internal)]
                ↓
           Let's Encrypt certs
```

**Steps:**

#### 1. Prepare the VPS
```bash
ssh root@your-vps
apt update && apt install -y docker.io docker-compose certbot
systemctl enable docker
```

#### 2. Clone the repo
```bash
git clone https://github.com/VinnsEdesigner/vyzorix.git
cd vyzorix
```

#### 3. Edit the Nginx config with your domain
```bash
# Replace api.vyzorix.com with your actual domain
sed -i 's/api.vyzorix.com/your-domain.com/g' tooling/nginx/vyzorix.conf
```

#### 4. Set production secrets in docker-compose.yml
```bash
# Edit the environment section — replace all "change-me" values
openssl rand -hex 32  # use this output for JWT_SECRET, SESSION_SECRET, etc.
```

#### 5. Get TLS certificates (first time only)
```bash
# Start just the API first (Nginx needs it healthy)
docker compose up -d vyzorix

# Get Let's Encrypt certs via certbot on the host
mkdir -p /var/www/html
certbot certonly --webroot -w /var/www/html -d your-domain.com

# Set up auto-renewal
echo "0 3 * * * certbot renew --quiet --post-hook 'docker compose restart nginx'" | crontab -
```

#### 6. Start everything
```bash
docker compose up -d
```

#### 7. Verify
```bash
curl -s https://your-domain.com/health
# Should return: {"status":"ok"}

# Check TLS
curl -sI https://your-domain.com/health | grep -i "Strict-Transport-Security"
```

#### 8. Updates / redeploy
```bash
git pull
docker compose up -d --build vyzorix  # rebuilds only the Go app, Nginx stays up
```

---

### Option C: Bare metal VPS (no Docker)

If you prefer to run the Go binary directly behind a system Nginx:

```bash
# 1. Install Nginx
apt install -y nginx

# 2. Copy the Nginx config
cp tooling/nginx/vyzorix.conf /etc/nginx/sites-available/vyzorix
ln -s /etc/nginx/sites-available/vyzorix /etc/nginx/sites-enabled/vyzorix
rm /etc/nginx/sites-enabled/default

# 3. Edit the config with your domain
nano /etc/nginx/sites-available/vyzorix  # replace api.vyzorix.com

# 4. Get TLS certs
certbot --nginx -d your-domain.com

# 5. Build the Go binary
cd apps/api && go build -o /usr/local/bin/vyzorix-api ./cmd/api

# 6. Create a systemd service
cat > /etc/systemd/system/vyzorix.service <<'EOF'
[Unit]
Description=Vyzorix API Server
After=network.target

[Service]
ExecStart=/usr/local/bin/vyzorix-api
EnvironmentFile=/etc/vyzorix/.env
Restart=always
User=vyzorix

[Install]
WantedBy=multi-user.target
EOF

# 7. Start
systemctl daemon-reload
systemctl enable vyzorix
systemctl start vyzorix
nginx -t && systemctl reload nginx
```

---

## What Nginx gives you that Go TLS doesn't

| Feature | Nginx | Go built-in TLS |
|---------|-------|----------------|
| TLS cert automation | certbot auto-renew | Manual or custom code |
| Rate limiting | `limit_req_zone` at edge | App-level only |
| HTTP/2 + HTTP/3 | Native, zero-config | HTTP/2 ok, HTTP/3 complex |
| Zero-downtime deploy | Nginx stays up, app restarts behind it | App restart = brief downtime |
| Static files | Serves SPA without Go overhead | Go handles everything |
| DDoS protection | `limit_conn` drops at edge | App must handle every connection |
| Battle-tested TLS | 20+ years of hardening | Go crypto is good but younger |

## On Render specifically

Render already gives you TLS, HTTP/2, health checks, and zero-downtime deploys. You don't need Nginx on Render. The only thing you must do:

1. Set `ALLOWED_ORIGINS` to your dashboard URL (not `*`).
2. Set all secrets (`JWT_SECRET`, `SESSION_SECRET`, etc.).
3. Ensure `ENFORCE_HMAC=true` in production.

The Nginx config and Docker Compose setup are for VPS deployments where you need self-managed TLS.
