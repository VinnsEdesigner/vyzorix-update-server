// @vyzorix/ssr-server - Node.js SSR Server for TanStack Start
// Uses H3/Nitro (already in TanStack Start dependencies) for better SSR integration

import { createServer } from "node:http";
import { createServer as createViteServer } from "vite";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const PORT = Number.parseInt(process.env.SSR_PORT || "3001", 10);
const isProduction = process.env.NODE_ENV === "production";
const WEB_APP_DIR = path.join(__dirname, "../web");

// ============================================
// DEVELOPMENT MODE - Vite with SSR
// ============================================
async function startDevServer() {
  console.log("🚀 Starting SSR server in DEVELOPMENT mode...");

  const vite = await createViteServer({
    root: WEB_APP_DIR,
    server: {
      port: PORT,
      proxy: {
        "/v1": "http://localhost:3000",
        "/api": "http://localhost:3000",
        "/health": "http://localhost:3000",
        "/healthz": "http://localhost:3000",
        "/bin": "http://localhost:3000",
      },
    },
    ssr: {
      resolve: {
        conditions: ["workerd", "worker", "browser"],
      },
    },
    logLevel: "info",
  });

  // Use Vite's connect middleware for development
  const server = createServer(vite.middlewares);

  server.listen(PORT, () => {
    console.log(`✅ SSR Dev Server ready on http://localhost:${PORT}`);
    console.log(`📦 Mode: development (Vite SSR)`);
    console.log(`🔄 Proxying /v1/*, /api/* to Go server at localhost:3000`);
  });

  return server;
}

// ============================================
// PRODUCTION MODE - Standalone H3/Nitro Server
// ============================================
async function startProdServer() {
  console.log("🚀 Starting SSR server in PRODUCTION mode...");

  // Load the pre-built SSR entry from Vite build output
  const distServerPath = path.join(WEB_APP_DIR, "dist/server/server.js");

  let fetchHandler;
  try {
    const module = await import(distServerPath);
    // TanStack Start generates a server with a fetch method
    fetchHandler = module.default?.fetch || module.fetch;
    
    if (!fetchHandler) {
      throw new Error("No fetch handler found in server entry");
    }
  } catch (err) {
    console.error("❌ Failed to load SSR server entry:");
    console.error("   Make sure the web app is built: cd ../web && pnpm run build");
    console.error("   Expected path:", distServerPath);
    throw err;
  }

  // Create HTTP server that uses the pre-compiled SSR handler
  const server = createServer((req, res) => {
    const url = `http://localhost:${PORT}${req.url}`;
    
    const headers = new Headers();
    for (const [key, value] of Object.entries(req.headers)) {
      if (typeof value === 'string') {
        headers.set(key, value);
      } else if (Array.isArray(value)) {
        headers.set(key, value.join(', '));
      }
    }

    const request = new Request(url, {
      method: req.method,
      headers,
      body: ['POST', 'PUT', 'PATCH'].includes(req.method) ? req : undefined,
    });

    fetchHandler(request, process.env, {})
      .then((response) => {
        res.statusCode = response.status;
        response.headers.forEach((value, key) => {
          res.setHeader(key, value);
        });
        
        response.text().then((body) => {
          res.end(body);
        });
      })
      .catch((err) => {
        console.error("SSR handler error:", err);
        res.statusCode = 500;
        res.end('Internal Server Error');
      });
  });

  server.listen(PORT, () => {
    console.log(`✅ SSR Server ready on http://localhost:${PORT}`);
    console.log(`📦 Mode: production (H3/Nitro SSR)`);
    console.log(`🔄 Go server proxies HTML requests here`);
  });

  return server;
}

// ============================================
// STARTUP
// ============================================
async function main() {
  try {
    const server = isProduction ? await startProdServer() : await startDevServer();

    const shutdown = (signal) => {
      console.log(`\n${signal} received, shutting down...`);
      server.close(() => {
        console.log("SSR server closed");
        process.exit(0);
      });
    };

    process.on("SIGINT", () => shutdown("SIGINT"));
    process.on("SIGTERM", () => shutdown("SIGTERM"));
  } catch (err) {
    console.error("❌ Failed to start SSR server:", err);
    process.exit(1);
  }
}

main();