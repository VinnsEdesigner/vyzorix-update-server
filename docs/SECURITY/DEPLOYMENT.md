# Deployment Guide
# Implements PRD Section 4.16 - Documentation

## Prerequisites

- Docker 20.10+
- Kubernetes 1.28+ (for K8s deployment)
- PostgreSQL 15+ (if not using SQLite)
- Nginx (for reverse proxy)

## Environment Variables

### Required

```bash
# Server Configuration
PORT=3000
NODE_ENV=production
DATABASE_URL=./data/vyzorix.db

# Security (generate with: openssl rand -hex 32)
TOKEN_SECRET=your-dashboard-token-secret-min-32-chars
JWT_SECRET=your-jwt-secret-min-32-chars

# URLs
BASE_URL=https://api.vyzorix.example.com
FRONTEND_URL=https://vyzorix.example.com
```

### Optional

```bash
# CORS
ALLOWED_ORIGINS=https://vyzorix.example.com

# Request Signing
REQUEST_SIGNING_ENABLED=true
SIGNING_TIMESTAMP_WINDOW=300

# Turnstile
TURNSTILE_ENABLED=true
TURNSTILE_SECRET=your-turnstile-secret

# CSRF
CSRF_ENABLED=true
CSRF_SECRET=your-csrf-secret

# Account Lockout
ACCOUNT_LOCKOUT_ENABLED=true
ACCOUNT_LOCKOUT_ATTEMPTS=5
ACCOUNT_LOCKOUT_DURATION=3600

# Audit Logging
AUDIT_LOGGING_ENABLED=true
AUDIT_LOG_RETENTION_DAYS=90

# MFA
MFA_ENABLED=true
MFA_ISSUER=Vyzorix

# Logging
LOG_LEVEL=info
LOG_REDACT_PII=true
```

## Docker Deployment

### Building the Image

```bash
# Build the Go API server
cd apps/api
docker build -t vyzorix-api:latest .

# Build the web app
cd ../web
pnpm install
pnpm run build
```

### Running with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f vyzorix-api

# Stop services
docker-compose down
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  api:
    build:
      context: ./apps/api
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - DATABASE_URL=/data/vyzorix.db
      - TOKEN_SECRET=${TOKEN_SECRET}
      - JWT_SECRET=${JWT_SECRET}
      - BASE_URL=https://api.vyzorix.example.com
    volumes:
      - ./data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health/live"]
      interval: 30s
      timeout: 10s
      retries: 3
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp:rw,noexec,nosuid,size=64m

  web:
    build:
      context: ./apps/web
      dockerfile: Dockerfile
    ports:
      - "3001:3001"
    environment:
      - NODE_ENV=production
      - API_URL=http://api:3000
    depends_on:
      - api
    restart: unless-stopped
```

## Kubernetes Deployment

### Prerequisites

```bash
# Create namespace
kubectl create namespace vyzorix

# Create secrets
kubectl create secret generic vyzorix-secrets \
  --from-literal=token-secret=$(openssl rand -hex 32) \
  --from-literal=jwt-secret=$(openssl rand -hex 32) \
  --namespace=vyzorix
```

### Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vyzorix-api
  namespace: vyzorix
spec:
  replicas: 3
  selector:
    matchLabels:
      app: vyzorix-api
  template:
    metadata:
      labels:
        app: vyzorix-api
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
      containers:
        - name: api
          image: ghcr.io/vinnsedigner/vyzorix:latest
          ports:
            - containerPort: 3000
          envFrom:
            - secretRef:
                name: vyzorix-secrets
            - configMapRef:
                name: vyzorix-config
          env:
            - name: NODE_ENV
              value: production
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
          livenessProbe:
            httpGet:
              path: /health/live
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /health/ready
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 10
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir:
            medium: Memory
            sizeLimit: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: vyzorix-api
  namespace: vyzorix
spec:
  selector:
    app: vyzorix-api
  ports:
    - port: 80
      targetPort: 3000
  type: ClusterIP
```

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: vyzorix-api-hpa
  namespace: vyzorix
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: vyzorix-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

## Nginx Configuration

```nginx
upstream vyzorix_api {
    server localhost:3000;
    keepalive 32;
}

upstream vyzorix_web {
    server localhost:3001;
    keepalive 16;
}

server {
    listen 443 ssl http2;
    server_name api.vyzorix.example.com;

    ssl_certificate /etc/ssl/certs/api.crt;
    ssl_certificate_key /etc/ssl/private/api.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Rate limiting zone
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

    location / {
        limit_req zone=api_limit burst=20 nodelay;

        proxy_pass http://vyzorix_api;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://vyzorix_api;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        health_check;
    }
}
```

## Monitoring

### Health Endpoints

- `/health/live` - Liveness probe (always returns 200)
- `/health/ready` - Readiness probe (checks DB)
- `/health/secure` - Security health check

### Metrics

Enable Prometheus metrics at `/metrics` with:

```bash
METRICS_ENABLED=true
METRICS_PATH=/metrics
```

## Troubleshooting

### Container Won't Start

1. Check logs: `docker-compose logs vyzorix-api`
2. Verify environment variables are set
3. Check database permissions

### 502 Bad Gateway

1. Check if API container is running
2. Verify health check endpoint
3. Check nginx upstream configuration

### High Memory Usage

1. Review resource limits
2. Enable query logging
3. Check for memory leaks in application

## Security Checklist

- [ ] All secrets stored in Kubernetes secrets or vault
- [ ] TLS 1.2+ enforced
- [ ] Security headers enabled
- [ ] Rate limiting configured
- [ ] Non-root user in container
- [ ] Read-only root filesystem
- [ ] Resource limits set
- [ ] Network policies configured
- [ ] Audit logging enabled
- [ ] Image scanning in CI/CD
