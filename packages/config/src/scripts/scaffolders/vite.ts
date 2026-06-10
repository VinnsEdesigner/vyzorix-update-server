// @vyzorix/config/scripts/scaffolders/vite.ts - Vite scaffolding
import { writeFile } from "fs/promises";
import { join } from "path";

export async function scaffoldVite(target: string): Promise<void> {
  const viteConfig = `import { defineViteConfig } from "@vyzorix/config/vite";

export default defineViteConfig({
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
      target: "${target === "cloudflare" ? "cloudflare" : "esnext"}",
    },
  },
});
`;

  await writeFile(join(process.cwd(), "vite.config.ts"), viteConfig);
}