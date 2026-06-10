// @vyzorix/config/presets/spa.ts - Pure SPA Preset
import { defineViteConfig } from "../vite";

export const spaPreset = defineViteConfig({
  proxy: {
    "/v1": "http://localhost:3000",
    "/api": "http://localhost:3000",
    "/health": "http://localhost:3000",
    "/healthz": "http://localhost:3000",
    "/bin": "http://localhost:3000",
  },
  vite: {
    build: {
      target: "esnext",
    },
  },
});

export default spaPreset;