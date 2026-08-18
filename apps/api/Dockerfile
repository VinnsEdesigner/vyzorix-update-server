# =============================================================================
# Vyzorix Update Server
# syntax=docker/dockerfile:1.7-labs

# =============================================================================
# Stage 1: Go Builder - Build the Go server binary
# =============================================================================
FROM golang:1.26 AS go-builder

WORKDIR /app

# Copy dependency files first for better layer caching
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

# Copy source and build
COPY apps/api ./

# Build the Go server (from apps/api directory)
RUN CGO_ENABLED=1 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/vyzorix-update-server ./cmd/api

# =============================================================================
# Stage 2: Web Builder - Build the frontend with Vite and pnpm
# =============================================================================
FROM node:22-bookworm AS web-builder

# Disable telemetry
ENV PNPM_TELEMETRY=0 \
    COREPACK_ENABLE_DOWNLOAD_PROMPT=0 \
    COREPACK_ENABLE_AUTOINSTALL=0

RUN npm install -g pnpm@9

WORKDIR /app

# Copy workspace files
COPY . .

# Install all dependencies
RUN pnpm install --no-frozen-lockfile

# Build config package first (outputs to dist/)
RUN pnpm --filter @vyzorix/config run build

# Then build web (which depends on config)
RUN pnpm --filter @vyzorix/web run build

# =============================================================================
# Stage 3: Production - Ubuntu with Node.js for SSR
# =============================================================================
FROM ubuntu:24.04 AS production

# Build arguments for version info
ARG VERSION=unknown
ARG BUILD_DATE=unknown

# Update and install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        sqlite3 \
        tzdata \
        nodejs \
        npm \
        curl \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean \
    && npm install -g pnpm

# Create app user
RUN useradd -m -s /bin/bash appuser

# Labels
LABEL org.opencontainers.image.title="Vyzorix Update Server" \
      org.opencontainers.image.description="Production update server" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.vendor="Vyzorix"

WORKDIR /app

# Copy package files (needed for SSR server)
COPY --from=web-builder /app/package.json /app/pnpm-lock.yaml* ./

# Ensure ES modules are enabled for SSR server (add to existing package.json)
RUN node -e "const pkg = require('./package.json'); pkg.type = 'module'; require('fs').writeFileSync('./package.json', JSON.stringify(pkg, null, 2));"

# Install all dependencies (vite is in devDependencies but needed for SSR)
RUN npm install --ignore-scripts --legacy-peer-deps

# Environment variables
ENV PORT=3000 \
    NODE_ENV=production \
    NODE_PATH=/app \
    DATABASE_URL=/data/vyzorix.db \
    VYZORIX_API_DIR=/data \
    VYZORIX_BIN_DIR=/app/bin \
    VYZORIX_PUBLIC_DIR=/app/public \
    PATH=/app/bin:/usr/local/bin:$PATH \
    # Disable SSR auto-build since we use pre-built assets
    SSR_AUTO_BUILD=false

# Create directories
RUN mkdir -p /data /app/bin /app/public /app/logs && \
    chown -R appuser:appuser /app /data

# Copy binary
COPY --from=go-builder --chown=appuser:appuser /out/vyzorix-update-server /app/vyzorix-update-server

# Copy .env.example for default configuration
COPY .env.example /app/.env.example

# Copy web assets
COPY --from=web-builder --chown=appuser:appuser /app/apps/web/dist/client /app/public
COPY --from=web-builder --chown=appuser:appuser /app/apps/web/dist/server /app/data/server

# Copy SSR server entry point from source
COPY apps/api/ssr-server.js /app/ssr-server.js

# Create web directory for SSR server
COPY --from=web-builder --chown=appuser:appuser /app/apps/web /web

# Make binary executable
RUN chmod +x /app/vyzorix-update-server /app/ssr-server.js

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Switch to non-root user
USER appuser

# Use exec form
ENTRYPOINT ["/app/vyzorix-update-server"]

# =============================================================================

