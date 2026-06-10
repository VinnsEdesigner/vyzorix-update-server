// @vyzorix/config/scripts/scaffolders/docker.ts - Docker scaffolding
import { writeFile } from "fs/promises";
import { join } from "path";

export async function scaffoldDocker(_target: string): Promise<void> {
  const dockerCompose = `version: "3.9"

services:
  api:
    build:
      context: ./apps/api
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
      - PORT=3000
      - DATABASE_URL=/data/vyzorix.db
      - JWT_SECRET=dev-jwt-secret-min-32-chars-change-in-production
      - TOKEN_SECRET=dev-token-secret-change-in-production
      - ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
    volumes:
      - api-data:/data
    command: go run .
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3

  web:
    build:
      context: ./apps/web
      dockerfile: Dockerfile
    ports:
      - "5173:5173"
    environment:
      - VITE_API_BASE_URL=http://api:3000
    volumes:
      - ./apps/web:/app
      - /app/node_modules
    command: pnpm dev --host
    depends_on:
      - api

volumes:
  api-data:
`;

  await writeFile(join(process.cwd(), "docker-compose.yml"), dockerCompose);
}