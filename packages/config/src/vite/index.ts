// @vyzorix/config/vite - Vite configuration utilities for TanStack Start + React
import { defineConfig, type UserConfig as ViteUserConfig } from "vite";

/**
 * Vyzorix Vite Configuration Options
 */
export interface VyzorixViteConfig {
  /**
   * TanStack Start configuration
   */
  tanstackStart?: {
    server?: {
      entry?: string;
    };
  };

  /**
   * Proxy configuration for API endpoints
   */
  proxy?: Record<string, string | { target: string; changeOrigin?: boolean }>;

  /**
   * Additional Vite configuration
   */
  vite?: Omit<ViteUserConfig, "plugins">;

  /**
   * Plugins array - should be provided by the consuming app
   * tanstackStart() must come first, then react(), then tsconfigPaths()
   */
  plugins?: ViteUserConfig["plugins"];
}

/**
 * Define Vyzorix Vite configuration with TanStack Start SSR support
 * Replaces @lovable.dev/vite-tanstack-config
 * 
 * @example
 * import { defineViteConfig } from "@vyzorix/config/vite";
 * import { tanstackStart } from "@tanstack/react-start/plugin/vite";
 * import react from "@vitejs/plugin-react";
 * import tsconfigPaths from "vite-tsconfig-paths";
 * import tailwindcss from "@tailwindcss/vite";
 * 
 * export default defineViteConfig({
 *   plugins: [
 *     tanstackStart(),
 *     tailwindcss(),
 *     react(),
 *     tsconfigPaths(),
 *   ],
 *   tanstackStart: { server: { entry: "src/server.ts" } },
 *   proxy: { "/api": "http://localhost:3000" },
 * });
 */
export function defineViteConfig(config: VyzorixViteConfig = {}): ViteUserConfig {
  const { tanstackStart: _tsConfig = {}, proxy = {}, vite: viteConfig = {}, plugins } = config;

  // Build proxy configuration
  const proxyConfig: Record<string, { target: string; changeOrigin: boolean }> = {};
  for (const [path, target] of Object.entries(proxy)) {
    if (typeof target === "string") {
      proxyConfig[path] = { target, changeOrigin: true };
    } else {
      proxyConfig[path] = { ...target, changeOrigin: target.changeOrigin ?? true };
    }
  }

  return defineConfig({
    ...viteConfig,
    plugins,
    server: {
      proxy: proxyConfig,
      ...viteConfig.server,
    },
    build: {
      sourcemap: true,
      ...viteConfig.build,
    },
  });
}

export type { UserConfig as ViteUserConfig } from "vite";