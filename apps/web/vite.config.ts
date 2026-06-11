// @vyzorix/config - Vite configuration with TanStack Start, React, Tailwind
import { defineViteConfig } from "@vyzorix/config/vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";
import tailwindcss from "@tailwindcss/vite";

export default defineViteConfig({
  plugins: [
    // TanStack Start MUST come first - it generates routes and SSR entry points
    tanstackStart(),
    // Tailwind CSS v4 plugin - handles @import "tailwindcss" properly
    tailwindcss(),
    // React plugin comes after Tailwind
    react(),
    // TypeScript paths resolution
    tsconfigPaths(),
  ],
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
