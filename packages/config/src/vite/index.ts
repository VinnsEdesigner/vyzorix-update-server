// @vyzorix/config/vite - Vite configuration utilities
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
 * Define Vyzorix Vite configuration
 * Replaces @lovable.dev/vite-tanstack-config
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