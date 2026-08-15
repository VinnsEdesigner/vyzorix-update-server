import { fixApolloSsrPlugin } from './fix-apollo-ssr-plugin';
import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import react from "@vitejs/plugin-react";
import { defineViteConfig } from "@vyzorix/config/vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineViteConfig({
  plugins: [
    fixApolloSsrPlugin(),
    tanstackStart({ server: { entry: "src/server.ts" } }),
    tailwindcss(),
    react(),
    tsconfigPaths(),
  ] as never,
  proxy: {
    "/v1": { target: "http://localhost:3000", changeOrigin: true },
    "/api": { target: "http://localhost:3000", changeOrigin: true },
    "/health": { target: "http://localhost:3000", changeOrigin: true },
    "/healthz": { target: "http://localhost:3000", changeOrigin: true },
    "/bin": { target: "http://localhost:3000", changeOrigin: true },
  },
});