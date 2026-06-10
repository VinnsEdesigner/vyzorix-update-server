// @vyzorix/config/nitro/targets/node.ts - Node.js Server Target
// Node.js server deployment configuration

export interface NodeConfig {
  routeRules?: Record<string, any>;
  node?: Record<string, any>;
}

export const nodePreset: NodeConfig = {
  routeRules: {
    "/**": {
      cors: true,
      headers: {
        "X-Content-Type-Options": "nosniff",
        "X-Frame-Options": "DENY",
        "X-XSS-Protection": "1; mode=block",
        "Referrer-Policy": "strict-origin-when-cross-origin",
      },
    },
    "/v1/auth/**": {
      headers: {
        "Cache-Control": "no-store",
        "X-Robots-Tag": "noindex, nofollow",
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
        "Content-Type": "application/vnd.android.dex",
      },
    },
  },
  node: {},
};

export default nodePreset;