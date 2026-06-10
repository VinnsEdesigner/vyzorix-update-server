// @vyzorix/config - Vite configuration with TanStack Start, React, Tailwind
import { defineViteConfig } from "@vyzorix/config/vite";

export default defineViteConfig({
  tanstackStart: {
    server: { entry: "src/server.ts" },
  },
  proxy: {
    "/v1": {
      target: "http://localhost:3000",
      changeOrigin: true,
    },
    "/api": {
      target: "http://localhost:3000",
      changeOrigin: true,
    },
    "/health": {
      target: "http://localhost:3000",
      changeOrigin: true,
    },
    "/healthz": {
      target: "http://localhost:3000",
      changeOrigin: true,
    },
    "/bin": {
      target: "http://localhost:3000",
      changeOrigin: true,
    },
  },
});
