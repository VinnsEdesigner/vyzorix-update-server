// @vyzorix/config/presets/ssr.ts - SSR Preset (Nitro + TanStack Start)
import { defineViteConfig } from "../vite";

export const ssrPreset = defineViteConfig({
  tanstackStart: {
    server: { entry: "src/server.ts" },
  },
  proxy: {
    "/v1": "http://localhost:3000",
    "/api": "http://localhost:3000",
    "/health": "http://localhost:3000",
    "/healthz": "http://localhost:3000",
    "/bin": "http://localhost:3000",
  },
  vite: {
    build: {
      target: "node",
      ssr: true,
    },
    ssr: {
      noExternal: ["@vyzorix/config"],
    },
  },
});

export default ssrPreset;