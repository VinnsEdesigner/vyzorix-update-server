// @vyzorix/ssr-server - Node.js SSR Server for TanStack Start
import { createServer } from "node:http";
import { createServer as createViteServer } from "vite";
import path from "node:path";
import { fileURLToPath } from "node:url";
import pc from "picocolors";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const PORT = Number.parseInt(process.env.SSR_PORT || "3001", 10);
const isProduction = process.env.NODE_ENV === "production";
const mode = process.env.SSR_MODE || (isProduction ? "production" : "development");
const WEB_APP_DIR = path.join(__dirname, "../web");

const log = {
  banner: (lines) => {
    console.log("");
    lines.forEach((line) => console.log(pc.cyan(line)));
    console.log("");
  },
  info: (label, value) => {
    console.log("  " + pc.dim(">") + " " + pc.bold(label));
    console.log("    " + pc.green(value));
  },
  success: (message) => {
    console.log("  " + pc.green("*") + " " + message);
  },
  warn: (message) => {
    console.log("  " + pc.yellow("!") + " " + pc.yellow(message));
  },
  error: (message) => {
    console.log("  " + pc.red("x") + " " + pc.red(message));
  },
  divider: () => {
    console.log("  " + pc.gray("=".repeat(56)));
  },
  url: (label, url) => {
    console.log("  " + pc.green("->") + " " + pc.bold(label));
    console.log("    " + pc.cyan(url));
  },
  kv: (key, value) => {
    console.log("    " + pc.bold(pc.cyan(key + ":")) + " " + pc.white(value));
  },
};

// Pure ASCII VYZORIX banner
function printWelcomeBanner(serverMode) {
  const lines = [
    pc.magenta(pc.bold("+-------------------------------------------------------------+")),
    pc.magenta(pc.bold("|   _   _           _        ____                           |")),
    pc.magenta(pc.bold("|  |_| |_|   ___   | |__    |  _|  ___  ___                 |")),
    pc.magenta(pc.bold("|  | | | |  / _ \  | '_ \  | |_  / _ \/ __|                |")),
    pc.magenta(pc.bold("|  | |_| | | (_) | | |_) | |  _|  __/\__ \                |")),
    pc.magenta(pc.bold("|  |___|_|  \___/  |_.__/   |_|   \___||___/               |")),
    pc.magenta(pc.bold("|                                                              |")),
    pc.magenta(pc.bold("|                    SSR SERVER v1.0.0                         |")),
    pc.magenta(pc.bold("+-------------------------------------------------------------+")),
  ];
  log.banner(lines);
  const modeColor = serverMode === "production" ? pc.red : pc.yellow;
  const modeText = serverMode.toUpperCase();
  console.log("  " + pc.dim("Mode:") + " " + modeColor(pc.bold("[" + modeText + "]")));
  log.divider();
}

async function startDevServer() {
  printWelcomeBanner("development");
  log.info("Initializing", "Vite Dev Server with SSR");
  log.url("Server URL", "http://localhost:" + PORT);
  const vite = await createViteServer({
    root: WEB_APP_DIR,
    server: { port: PORT, proxy: { "/v1": "http://localhost:3000", "/api": "http://localhost:3000", "/health": "http://localhost:3000", "/healthz": "http://localhost:3000", "/bin": "http://localhost:3000" } },
    ssr: { resolve: { conditions: ["workerd", "worker", "browser"] } },
    logLevel: "info",
  });
  const server = createServer(vite.middlewares);
  server.listen(PORT, () => {
    log.divider();
    log.success("SSR Dev Server ready on http://localhost:" + PORT);
    log.kv("Mode", "Development (Vite SSR + HMR)");
    console.log("");
    console.log("  " + pc.dim("Press") + " " + pc.bold("Ctrl+C") + " " + pc.dim("to stop"));
    console.log("");
  });
  return server;
}

async function startProdServer() {
  printWelcomeBanner("production");
  log.info("Initializing", "Pre-built SSR Handler");
  const distServerPath = path.join(WEB_APP_DIR, "dist/server/server.js");
  let fetchHandler;
  try {
    log.kv("SSR Entry", distServerPath);
    log.divider();
    const module = await import(distServerPath);
    fetchHandler = module.default?.fetch || module.fetch;
    if (!fetchHandler) throw new Error("No fetch handler found");
  } catch (err) {
    console.log("");
    log.error("Failed to load SSR server entry");
    log.warn("Run: cd ../web && pnpm run build");
    throw err;
  }
  const server = createServer((req, res) => {
    const url = "http://localhost:" + PORT + req.url;
    const headers = new Headers();
    for (const [key, value] of Object.entries(req.headers)) {
      if (typeof value === "string") headers.set(key, value);
      else if (Array.isArray(value)) headers.set(key, value.join(", "));
    }
    const request = new Request(url, { method: req.method, headers, body: ["POST", "PUT", "PATCH"].includes(req.method) ? req : undefined });
    fetchHandler(request, process.env, {})
      .then((response) => {
        res.statusCode = response.status;
        response.headers.forEach((value, key) => res.setHeader(key, value));
        response.text().then((body) => res.end(body));
      })
      .catch((err) => { console.error("SSR handler error:", err); res.statusCode = 500; res.end("Internal Server Error"); });
  });
  server.listen(PORT, () => {
    log.divider();
    log.success("SSR Server ready on http://localhost:" + PORT);
    log.kv("Mode", "Production (H3/Nitro)");
    console.log("");
    console.log("  " + pc.dim("Press") + " " + pc.bold("Ctrl+C") + " " + pc.dim("to stop"));
    console.log("");
  });
  return server;
}

async function main() {
  try {
    const server = mode === "production" ? await startProdServer() : await startDevServer();
    const shutdown = (signal) => {
      console.log("\n  " + pc.yellow("@") + " " + signal + " received, shutting down...");
      server.close(() => { log.success("SSR server closed"); process.exit(0); });
    };
    process.on("SIGINT", () => shutdown("SIGINT"));
    process.on("SIGTERM", () => shutdown("SIGTERM"));
  } catch (err) {
    console.log("");
    log.error("Failed to start SSR server:");
    console.error("  " + pc.red(err.message));
    process.exit(1);
  }
}

main();
