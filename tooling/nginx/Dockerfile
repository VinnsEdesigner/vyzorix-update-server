# Vyzorix Nginx Container
# TLS-terminating reverse proxy + static file server for the Go API.
#
# Build: docker build -t vyzorix-nginx -f tooling/nginx/Dockerfile .
# Run:   docker compose up (uses docker-compose.yml with nginx + vyzorix services)

FROM nginx:1.27-alpine

# Remove default config.
RUN rm /etc/nginx/conf.d/default.conf

# Copy our production config.
COPY tooling/nginx/vyzorix.conf /etc/nginx/conf.d/vyzorix.conf

# Health check.
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -q --spider http://localhost/health || exit 1

EXPOSE 80 443
