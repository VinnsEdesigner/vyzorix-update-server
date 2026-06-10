// @vyzorix/config/nitro/targets/static.ts - Static Hosting Preset
// Static hosting deployment configuration

export interface StaticConfig {
  routeRules?: Record<string, Record<string, unknown>>;
  prerender?: Record<string, unknown>;
}

export const staticPreset: StaticConfig = {
  routeRules: {
    // All routes are pre-rendered for static hosting
    "/**": {
      prerender: true,
      cors: true,
      headers: {
        "Cache-Control": "public, max-age=31536000, immutable",
        "X-Content-Type-Options": "nosniff",
        "X-Frame-Options": "DENY",
        "X-XSS-Protection": "1; mode=block",
      },
    },
    // API routes should not be prerendered
    "/v1/**": {
      prerender: false,
      headers: {
        "Cache-Control": "no-store",
      },
    },
    "/api/**": {
      prerender: false,
      headers: {
        "Cache-Control": "no-store",
      },
    },
  },
  // Static hosting specific
  prerender: {
    crawlLinks: true,
    routes: ["/"],
  },
};

export default staticPreset;