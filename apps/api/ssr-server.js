// @vyzorix/ssr-server - Node.js SSR Server for TanStack Start
// This server handles SSR rendering for the React app
// Go server can proxy requests to this server

import express from "express";
import { createServer } from "http";
import path from "path";
import { fileURLToPath } from "url";
import { createRequestHandler } from "@tanstack/react-start/server";
import { createStart } from "@tanstack/react-start";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const PORT = process.env.SSR_PORT || 3001;
const isProduction = process.env.NODE_ENV === "production";

// Create TanStack Start instance
const start = createStart({
  relativePath: path.join(__dirname, "../web/src"),
  configPath: path.join(__dirname, "../web/vite.config.ts"),
});

// Create Express app
const app = express();

// Health check endpoint
app.get("/health", (req, res) => {
  res.json({
    ok: true,
    ssr: true,
    mode: isProduction ? "production" : "development",
  });
});

// In production, use the built SSR handler
// In development, use Vite's dev server directly
if (isProduction) {
  // Production: load the pre-built SSR entry
  const { createServer } = await import(
    path.join(__dirname, "../web/dist/server/server-entry.js")
  );

  const handler = createServer({
    start,
    mode: "production",
  });

  app.use(handler);
} else {
  // Development: use Vite's SSR middleware
  const vite = await import("vite");

  const viteServer = await vite.createServer({
    root: path.join(__dirname, "../web"),
    server: {
      port: Number.parseInt(PORT),
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
  });

  app.use(viteServer.middlewares);

  // Handle SSR requests through Vite
  app.use("*", async (req, res, next) => {
    try {
      const url = req.originalUrl;

      const template = await viteServer.transformIndexHtml(
        url,
        `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Vyzorix Update Server</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/entry-client.tsx"></script>
  </body>
</html>`
      );

      const { render } = await viteServer.ssrLoadModule(
        path.join(__dirname, "../web/src/server.ts")
      );

      const response = await render({
        request: new Request(url, { method: "GET" }),
        context: {},
        mode: "development",
      });

      if (response) {
        const html = await response.text();
        res.status(response.status).send(
          template.replace("<!--app-html-->", html).replace("<!--head-->", "").replace("<!--scripts-->", `<script type="module" src="/src/entry-client.tsx"></script>`)
        );
      } else {
        next();
      }
    } catch (e) {
      viteServer.ssrFixStacktrace(e);
      next(e);
    }
  });
}

// Create HTTP server
const server = createServer(app);

server.listen(PORT, () => {
  console.log(`✅ SSR Server ready on http://localhost:${PORT}`);
  console.log(`📦 Mode: ${isProduction ? "production" : "development"}`);
  console.log(`🔄 Proxy from Go server to this SSR server`);
});

// Export for testing
export { app, server };