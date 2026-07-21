# Docker Deployment Guide

This guide covers building and running the Vyzorix Update Server using Docker.

## Prerequisites

- Docker 20.10+ installed
- 2GB+ disk space for the image
- Network access to pull base images

## Quick Start

### Build the Image

```bash
# From repository root
docker build -t vyzorix-update-server:latest -f apps/api/Dockerfile .

# Or from apps/api directory
cd apps/api
docker build -t vyzorix-update-server:latest .
```

### Run the Container

```bash
# Create a directory for persistent data
mkdir -p /tmp/vyzorix-data

# Run with SSR enabled (recommended for production)
docker run -d --name vyzorix \
  -e TOKEN_SECRET=your-secure-token-secret-min-32-chars \
  -e JWT_SECRET=your-secure-jwt-secret-min-32-chars \
  -e SSR_ENABLED=true \
  -v /tmp/vyzorix-data:/data \
  -p 3000:3000 \
  vyzorix-update-server:latest
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TOKEN_SECRET` | Yes | - | Dashboard token secret (min 32 chars) |
| `JWT_SECRET` | Yes | - | JWT signing secret (min 32 chars) |
| `SSR_ENABLED` | No | `true` | Enable SSR mode |
| `SSR_AUTO_BUILD` | No | `false` | Auto-build web assets (not needed with pre-built) |
| `PORT` | No | `3000` | Server port |
| `DATABASE_URL` | No | `/data/vyzorix.db` | SQLite database path |
| `FIREBASE_CREDENTIALS` | No | - | Firebase service account JSON |
| `SMTP_HOST` | No | - | SMTP server host |
| `SMTP_PORT` | No | `587` | SMTP server port |
| `SMTP_USER` | No | - | SMTP username |
| `SMTP_PASS` | No | - | SMTP password |
| `SMTP_FROM` | No | - | From email address |

### Volume Mounts

| Path | Description |
|------|-------------|
| `/data` | Persistent storage for database and runtime data |

## Deployment Modes

### Mode 1: SSR (Server-Side Rendering)

Full SSR support with Node.js for dynamic page rendering.

```bash
docker run -d --name vyzorix \
  -e TOKEN_SECRET=your-secret \
  -e JWT_SECRET=your-jwt-secret \
  -e SSR_ENABLED=true \
  -v /path/to/data:/data \
  -p 3000:3000 \
  vyzorix-update-server:latest
```

### Mode 2: SPA (Static Fallback)

Lightweight mode serving pre-built static files.

```bash
docker run -d --name vyzorix \
  -e TOKEN_SECRET=your-secret \
  -e JWT_SECRET=your-jwt-secret \
  -e SSR_ENABLED=false \
  -v /path/to/data:/data \
  -p 3000:3000 \
  vyzorix-update-server:latest
```

## Docker Compose

### Example `docker-compose.yml`

```yaml
version: '3.8'

services:
  vyzorix:
    build:
      context: .
      dockerfile: apps/api/Dockerfile
    container_name: vyzorix-server
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - TOKEN_SECRET=${TOKEN_SECRET}
      - JWT_SECRET=${JWT_SECRET}
      - SSR_ENABLED=true
      - NODE_ENV=production
    volumes:
      - vyzorix-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  vyzorix-data:
```

### Run with Docker Compose

```bash
# Create .env file
echo "TOKEN_SECRET=your-secure-token-secret-min-32-chars" > .env
echo "JWT_SECRET=your-secure-jwt-secret-min-32-chars" >> .env

# Build and start
docker compose up -d

# View logs
docker compose logs -f

# Stop
docker compose down
```

## Health Checks

### Check Container Health

```bash
# Docker health status
docker ps | grep vyzorix

# HTTP health endpoint
curl http://localhost:3000/health
```

Expected response:
```json
{"status":"ok","uptime":123}
```

### SSR Health Check

```bash
curl http://localhost:3000/v1/ssr/health
```

## Production Hardening

For production deployments, add these security options:

```bash
docker run -d --name vyzorix \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m \
  --cap-drop=ALL \
  --security-opt=no-new-privileges:true \
  --memory=512m \
  --memory-swap=512m \
  --pids-limit=100 \
  -e TOKEN_SECRET=your-secret \
  -e JWT_SECRET=your-jwt-secret \
  -e SSR_ENABLED=true \
  -v /path/to/data:/data \
  -p 3000:3000 \
  vyzorix-update-server:latest
```

### Security Options Explained

| Option | Purpose |
|--------|---------|
| `--read-only` | Root filesystem is read-only |
| `--tmpfs /tmp` | Writable /tmp in memory only |
| `--cap-drop=ALL` | Remove all Linux capabilities |
| `--security-opt=no-new-privileges` | Prevent privilege escalation |
| `--memory=512m` | Limit memory usage |
| `--pids-limit=100` | Limit number of processes |

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs vyzorix

# Common issues:
# - Missing TOKEN_SECRET or JWT_SECRET
# - Permission issues with volume mount
# - Port 3000 already in use
```

### SSR Not Starting

```bash
# Enable SSR debug mode (check logs)
docker logs vyzorix | grep -i ssr

# SSR needs:
# - SSR_ENABLED=true
# - /web directory present
# - node_modules with vite installed
```

### Database Issues

```bash
# Check database file exists
docker exec vyzorix ls -la /data/

# Recreate database (WARNING: loses data)
docker exec vyzorix rm /data/vyzorix.db
docker restart vyzorix
```

### Network/Port Issues

```bash
# Check port is not in use
lsof -i :3000

# Check container port mapping
docker port vyzorix
```

## Image Details

### Image Size

- **Compressed**: ~292MB
- **Uncompressed**: ~1.34GB

The image includes:
- Ubuntu 24.04 base
- Go 1.26 runtime
- Node.js 18 runtime
- SQLite with WAL mode
- Full node_modules (for SSR)

### Multi-Platform Build

```bash
# Build for multiple architectures
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t vyzorix-update-server:latest \
  -f apps/api/Dockerfile \
  --push \
  .
```

## Cleanup

```bash
# Stop and remove container
docker stop vyzorix && docker rm vyzorix

# Remove image (if needed)
docker rmi vyzorix-update-server:latest

# Remove data volume
docker volume rm vyzorix_data
```

## Next Steps

1. Configure SSL/TLS termination (reverse proxy)
2. Set up environment-specific secrets
3. Configure monitoring and alerting
4. Review and apply security hardening options
5. Set up backup strategy for `/data` volume
6. Configure Firebase Cloud Messaging for push notifications
7. Set up email notifications via SMTP
