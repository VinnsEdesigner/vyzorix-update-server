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
ENV PORT=3000 NODE_ENV=production DATABASE_URL=/data/vyzorix.db VYZORIX_API_DIR=/app/api/v1 VYZORIX_BIN_DIR=/app/bin VYZORIX_PUBLIC_DIR=/app/public
COPY --from=go-builder /out/vyzorix-update-server /app/vyzorix-update-server
COPY --from=web-builder /app/.output/public /app/public
COPY api /app/api
COPY bin /app/bin
EXPOSE 3000
CMD ["/app/vyzorix-update-server"]
