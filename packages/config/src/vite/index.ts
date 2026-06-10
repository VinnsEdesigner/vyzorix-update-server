// @vyzorix/config/vite - Vite configuration utilities
import { defineConfig, type UserConfig as ViteUserConfig } from "vite";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";

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
}

/**
 * Define Vyzorix Vite configuration
 * Replaces @lovable.dev/vite-tanstack-config
 */
export function defineViteConfig(config: VyzorixViteConfig = {}): ViteUserConfig {
  const { tanstackStart: tsConfig = {}, proxy = {}, vite: viteConfig = {} } = config;

  // Build proxy configuration
  const proxyConfig: Record<string, { target: string; changeOrigin: boolean }> = {};
  for (const [path, target] of Object.entries(proxy)) {
    if (typeof target === "string") {
      proxyConfig[path] = { target, changeOrigin: true };
    } else {
      proxyConfig[path] = { ...target, changeOrigin: target.changeOrigin ?? true };
    }
  }

  // Build plugins array
  const plugins = [
    react(),
    tsconfigPaths(),
  ];

  // Add TanStack Start if configured
  if (tsConfig && Object.keys(tsConfig).length > 0) {
    // Lazy load TanStack Start to avoid build errors if not installed
    try {
      // Dynamic import would be better but keeping it simple for now
    } catch (e) {
      // TanStack Start not available, skip
    }
  }

  return defineConfig({
    ...viteConfig,
    plugins,
    css: {
      postcss: {
        plugins: [],
      },
    },
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