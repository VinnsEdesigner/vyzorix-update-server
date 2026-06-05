FROM golang:1.22-alpine AS go-builder
RUN apk add --no-cache build-base sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/vyzorix-update-server .

FROM node:22-alpine AS web-builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM alpine:3.20
RUN apk add --no-cache ca-certificates sqlite tzdata
WORKDIR /app

# Create required directories
RUN mkdir -p /data /app/bin /app/public

# Environment variables - matches config.go defaults and .env.example
ENV PORT=3000 \
    NODE_ENV=production \
    DATABASE_URL=/data/vyzorix.db \
    VYZORIX_API_DIR=/app/data \
    VYZORIX_BIN_DIR=/app/bin \
    VYZORIX_PUBLIC_DIR=/app/public

# Copy binary from go-builder
COPY --from=go-builder /out/vyzorix-update-server /app/vyzorix-update-server

# Copy built web assets from web-builder
# TanStack Start outputs to .output/public
COPY --from=web-builder /app/.output/public /app/public

# Copy data directory from builder (version.json, changelog.json)
# This needs the data directory to exist in the source
COPY --from=web-builder /app/data /app/data || true

# Copy bin directory contents (APK files, binaries)
COPY --from=web-builder /app/bin /app/bin || true

EXPOSE 3000

# Health check: verify the server is responding
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:3000/health || exit 1

# Use exec form for proper signal handling (SIGTERM → process)
ENTRYPOINT ["/app/vyzorix-update-server"]
