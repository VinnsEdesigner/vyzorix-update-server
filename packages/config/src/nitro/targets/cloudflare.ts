// @vyzorix/config/nitro/targets/cloudflare.ts - Cloudflare Workers Target
// Cloudflare Workers deployment configuration

export interface CloudflareConfig {
  routeRules?: Record<string, any>;
  wasm?: Record<string, any>;
  externals?: Record<string, any>;
  cfProperties?: Record<string, Function>;
}

export const cloudflarePreset: CloudflareConfig = {
  routeRules: {
    "/**": {
      cors: true,
      headers: {
        "Cache-Control": "public, max-age=0, must-revalidate",
        "X-Content-Type-Options": "nosniff",
        "X-Frame-Options": "DENY",
        "X-XSS-Protection": "1; mode=block",
      },
    },
    "/v1/auth/**": {
      headers: {
        "Cache-Control": "no-store",
      },
    },
    "/api/v1/version": {
      headers: {
        "Cache-Control": "public, max-age=3600",
      },
    },
    "/bin/**": {
      headers: {
        "Cache-Control": "public, max-age=86400",
      },
    },
  },
  wasm: {
    wasmLazyDirs: [],
  },
  externals: {
    external: ["node:async_hooks"],
  },
  cfProperties: {
    asyncHeaders(age: number) {
      return {
        "Cache-Ttl": age,
        "CDN-Cache-Control": `public, max-age=${age}, s-maxage=${age}`,
      };
    },
  },
};

export default cloudflarePreset;