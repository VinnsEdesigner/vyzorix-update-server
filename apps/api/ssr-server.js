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

let isReady = false;
let startupTime = 0;
let serverInstance = null;

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

function printWelcomeBanner(serverMode) {
  const lines = [
    pc.magenta(pc.bold("+-------------------------------------------------------------+")),
    pc.magenta(pc.bold("|   _   _           _        ____                           |")),
    pc.magenta(pc.bold("|  |_| |_|   ___   | |__    |  _|  ___  ___                 |")),
    pc.magenta(pc.bold("|  | | | |  / _ \\  | '_ \\  | |_  / _ \\/ __|                |")),
    pc.magenta(pc.bold("|  | |_| | | (_) | | |_) | |  _|  __/\\__ \\                |")),
    pc.magenta(pc.bold("|  |___|_|  \\___/  |_.__/   |_|   \\___||___/               |")),
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

function healthHandler(req, res) {
  if (req.url === "/health" || req.url === "/healthz") {
    const status = isReady ? 200 : 503;
    try {
      res.writeHead(status, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        status: isReady ? "ok" : "starting",
        uptime: startupTime > 0 ? Math.floor((Date.now() - startupTime) / 1000) : 0,
        mode: mode,
        timestamp: new Date().toISOString(),
      }));
    } catch (err) {
          }
    return true;
  }
  return false;
}

function createProxiedHandler(fetchHandler) {
  return (req, res) => {
        if (healthHandler(req, res)) return;

    const url = "http:    const headers = new Headers();
    for (const [key, value] of Object.entries(req.headers)) {
      if (typeof value === "string") headers.set(key, value);
      else if (Array.isArray(value)) headers.set(key, value.join(", "));
    }
    const bodyMethods = ["POST", "PUT", "PATCH", "DELETE"];
    const request = new Request(url, {
      method: req.method,
      headers,
      body: bodyMethods.includes(req.method) ? req : undefined,
    });

    let responseHandled = false;
    const handleError = (err) => {
      if (responseHandled) return;
      responseHandled = true;
      console.error("SSR handler error:", err);
      try {
        res.statusCode = 500;
        res.setHeader("Content-Type", "text/html");
        res.end("<html><body><h1>SSR Error</h1><p>Please try refreshing the page.</p></body></html>");
      } catch (e) {
              }
    };

    fetchHandler(request, process.env, {})
      .then((response) => {
        if (responseHandled) return;
        responseHandled = true;
        res.statusCode = response.status;
        response.headers.forEach((value, key) => {
          try {
            res.setHeader(key, value);
          } catch (e) {
                      }
        });
        response.text().then((body) => {
          try {
            res.end(body);
          } catch (e) {
                      }
        }).catch(handleError);
      })
      .catch(handleError);
  };
}

async function startDevServer() {
  printWelcomeBanner("development");
  log.info("Initializing", "Vite Dev Server with SSR");
  log.url("Server URL", "http:  log.kv("Health Check", "http:  log.divider();

  const vite = await createViteServer({
    root: WEB_APP_DIR,
    server: {
      port: PORT,
      proxy: {
        "/v1": "http:        "/api": "http:        "/health": "http:        "/healthz": "http:        "/bin": "http:      },
    },
    ssr: { resolve: { conditions: ["workerd", "worker", "browser"] } },
    logLevel: "info",
  });

  const server = createServer(vite.middlewares);

    serverInstance = server.listen(PORT, () => {
    startupTime = Date.now();
    isReady = true;
    log.divider();
    log.success("SSR Dev Server ready on http:    log.kv("Mode", "Development (Vite SSR + HMR)");
    log.kv("Status", "Healthy");
    console.log("");
    console.log("  " + pc.dim("Press") + " " + pc.bold("Ctrl+C") + " " + pc.dim("to stop"));
    console.log("");
  });

  return server;
}

async function startProdServer() {
  printWelcomeBanner("production");
  log.info("Initializing", "Pre-built SSR Handler");
  log.kv("Port", PORT.toString());
  log.divider();

  const distServerPath = path.join(WEB_APP_DIR, "dist/server/server.js");
  let fetchHandler;

  try {
    log.kv("SSR Entry", distServerPath);

        const module = await import(distServerPath);
    fetchHandler = module.default?.fetch || module.fetch;

    if (!fetchHandler) {
      throw new Error("No fetch handler found in SSR server entry");
    }

    log.success("SSR handler loaded");
  } catch (err) {
    console.log("");
    log.error("Failed to load SSR server entry:");
    log.error(err.message);
    log.warn("Run: cd ../web && pnpm run build");
    throw err;
  }

    const server = createServer(createProxiedHandler(fetchHandler));

    serverInstance = server.listen(PORT, () => {
    startupTime = Date.now();
    isReady = true;
    log.divider();
    log.success("SSR Server ready on http:    log.kv("Mode", "Production (H3/Nitro)");
    log.kv("Status", "Healthy");
    console.log("");
    console.log("  " + pc.dim("Press") + " " + pc.bold("Ctrl+C") + " " + pc.dim("to stop"));
    console.log("");
  });

    serverInstance.on("error", (err) => {
    log.error("Server error: " + err.message);
  });

  return serverInstance;
}

async function main() {
  try {
    console.log("");
    console.log("  " + pc.dim("Starting SSR server..."));
    console.log("");

    const server = mode === "production" ? await startProdServer() : await startDevServer();

    const shutdown = (signal) => {
      console.log("\n  " + pc.yellow("@") + " " + signal + " received, shutting down...");
      isReady = false;
      if (serverInstance) {
        serverInstance.close(() => {
          log.success("SSR server closed");
          process.exit(0);
        });
      } else {
        process.exit(0);
      }

            setTimeout(() => {
        log.warn("Forced exit after timeout");
        process.exit(1);
      }, 5000);
    };

    process.on("SIGINT", () => shutdown("SIGINT"));
    process.on("SIGTERM", () => shutdown("SIGTERM"));

        process.on("uncaughtException", (err) => {
      log.error("Uncaught exception: " + err.message);
      console.error(err.stack);
    });

        process.on("unhandledRejection", (reason, promise) => {
      log.error("Unhandled rejection at: " + promise + ", reason: " + reason);
    });

  } catch (err) {
    console.log("");
    log.error("Failed to start SSR server:");
    console.error("  " + pc.red(err.message));
    process.exit(1);
  }
}

main();
