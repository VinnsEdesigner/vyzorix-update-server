import fs from "node:fs";
import http from "node:http";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { createServer as createViteServer } from "vite";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PORT = 3000;
const GO_BACKEND = "http://127.0.0.1:8080";

async function startServer() {
  const isProd = process.env.NODE_ENV === "production";
  let vite: any;

  if (!isProd) {
    vite = await createViteServer({
      server: { middlewareMode: true },
      appType: "custom",
    });
  }

  const getIndexHtml = async (reqUrl: string, reqCookies: string): Promise<string> => {
    const htmlPath = isProd
      ? path.join(__dirname, "dist", "index.html")
      : path.join(__dirname, "index.html");

    let html = fs.readFileSync(htmlPath, "utf-8");

    if (!isProd && vite) {
      html = await vite.transformIndexHtml(reqUrl, html);
    }

    const prefetchedState: Record<string, any> = {
      view: "signup",
      profileData: null,
      successReport: null,
      verificationToken: null,
    };

    if (reqCookies.includes("vyzorix_session=")) {
      try {
        const meResponse = await fetch(`${GO_BACKEND}/api/auth/me`, {
          headers: { Cookie: reqCookies },
        });
        if (meResponse.ok) {
          const report = await meResponse.json();
          prefetchedState.view = "success";
          prefetchedState.successReport = report;
          prefetchedState.profileData = {
            fullName: report.fullName,
            email: report.email,
            username: report.username,
          };
        }
      } catch (_e) {}
    } else if (reqCookies.includes("vyzorix_pending_auth=")) {
      try {
        const match = reqCookies.match(/vyzorix_pending_auth=([^;]+)/);
        if (match && match[1]) {
          const cookieVal = decodeURIComponent(match[1]);
          const parts = cookieVal.split("|");
          if (parts.length >= 4) {
            const [token, fullName, email, username] = parts;
            prefetchedState.view = "waiting_verification";
            prefetchedState.verificationToken = token;
            prefetchedState.profileData = { fullName, email, username };
          }
        }
      } catch (_e) {}
    }

    const stateScript = `<script>window.__VYZORIX_PREFETCHED_STATE__ = ${JSON.stringify(prefetchedState)};</script>`;
    return html.replace('<div id="root">', `${stateScript}\n<div id="root">`);
  };

  const server = http.createServer(async (req, res) => {
    if (req.url?.startsWith("/api/auth/")) {
      const targetUrl = `${GO_BACKEND}${req.url}`;
      const headers = new Headers();

      for (let i = 0; i < req.rawHeaders.length; i += 2) {
        const key = req.rawHeaders[i];
        const val = req.rawHeaders[i + 1];
        if (key && val && !["host", "connection"].includes(key.toLowerCase())) {
          headers.append(key, val);
        }
      }

      try {
        const bodyChunks: Buffer[] = [];
        req.on("data", (chunk) => bodyChunks.push(chunk));

        req.on("end", async () => {
          const bodyBuffer = Buffer.concat(bodyChunks);
          const options: RequestInit = {
            method: req.method,
            headers: headers,
          };
          if (req.method !== "GET" && req.method !== "HEAD" && bodyBuffer.length > 0) {
            options.body = bodyBuffer;
          }

          const response = await fetch(targetUrl, options);
          res.statusCode = response.status;

          response.headers.forEach((value, key) => {
            if (key.toLowerCase() !== "transfer-encoding") {
              res.setHeader(key, value);
            }
          });

          const buffer = await response.arrayBuffer();
          res.end(Buffer.from(buffer));
        });
      } catch (error) {
        res.statusCode = 502;
        res.end(JSON.stringify({ message: "Bad Gateway - Go backend connectivity failure." }));
      }
      return;
    }

    if (vite) {
      vite.middlewares(req, res, async () => {
        if (!req.url?.includes(".") || req.url.endsWith(".html")) {
          try {
            const html = await getIndexHtml(req.url || "/", req.headers.cookie || "");
            res.setHeader("Content-Type", "text/html");
            res.end(html);
          } catch (e: any) {
            if (vite) vite.ssrFixStacktrace(e);
            res.statusCode = 500;
            res.end(e.message);
          }
        }
      });
    } else {
      if (!req.url?.includes(".") || req.url?.endsWith(".html")) {
        try {
          const html = await getIndexHtml(req.url || "/", req.headers.cookie || "");
          res.setHeader("Content-Type", "text/html");
          res.end(html);
        } catch (e: any) {
          res.statusCode = 500;
          res.end(e.message);
        }
      } else {
        const filePath = path.join(__dirname, "dist", req.url!);
        if (fs.existsSync(filePath)) {
          fs.createReadStream(filePath).pipe(res);
        } else {
          res.statusCode = 404;
          res.end();
        }
      }
    }
  });

  server.listen(PORT, "0.0.0.0", () => {
    console.log(
      `[Native Node Server]: Multi-tier hydration listener active on http://0.0.0.0:${PORT}`,
    );
  });
}

startServer();
