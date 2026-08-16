/**
 * vyzorix-proxy-server.cjs
 *
 * Vyzorix BFF Proxy Server — main entry point.
 *
 * Sits between the browser and the Go API:
 *
 *   Browser ──(httpOnly cookie)──> Proxy (:3001) ──(signed)──> Go API (:3000)
 *
 * Responsibilities:
 *   1. Serves the web app static files (no vite — pure Node http + fs)
 *   2. Manages server-side auth sessions (signing key, tokens, CSRF)
 *   3. Signs all API requests with X-Vyzorix-* HMAC headers
 *   4. Auto-fetches and injects CSRF tokens
 *   5. Intercepts login/logout to capture/destroy credentials
 *   6. Proxies WebSocket connections for GraphQL subscriptions
 *   7. Logs every request to reveal bugs in the pipeline
 *
 * Usage:
 *   node vyzorix-proxy-server.cjs                    # dev defaults
 *   NODE_ENV=production node vyzorix-proxy-server.cjs # prod
 *   PROXY_PORT=3001 API_TARGET=http://localhost:3000 node vyzorix-proxy-server.cjs
 *
 * Build the web app first:
 *   cd apps/VyzoriX_web && pnpm build
 */

'use strict';

const http = require('node:http');
const { URL } = require('node:url');
var cfg = require('./vyzorix-proxy-config.cjs');
var logger = require('./vyzorix-proxy-logger.cjs');
var core = require('./vyzorix-proxy-core.cjs');
var webServer = require('./vyzorix-proxy-web-server.cjs');
var wsProxy = require('./vyzorix-proxy-websocket.cjs');
var sessionStore = require('./vyzorix-proxy-session-store.cjs');

var config = cfg.config;
var isApiPath = cfg.isApiPath;
var banner = logger.banner;
var divider = logger.divider;
var kv = logger.kv;
var success = logger.success;
var info = logger.info;
var warn = logger.warn;
var error = logger.error;
var pc = logger.pc;
var handleApiRequest = core.handleApiRequest;
var bufferBody = core.bufferBody;
var serveStaticFile = webServer.serveStaticFile;
var attachWebSocketProxy = wsProxy.attachWebSocketProxy;
var cleanupExpiredSessions = sessionStore.cleanupExpiredSessions;
var activeSessionCount = sessionStore.activeSessionCount;

function printBanner() {
  banner([
    pc.magenta(pc.bold('╔══════════════════════════════════════════════════════════════╗')),
    pc.magenta(pc.bold('║  Vyzorix Proxy Server (BFF)                          v1.0.0   ║')),
    pc.magenta(pc.bold('║  Serves Web · Signs Requests · Manages Sessions              ║')),
    pc.magenta(pc.bold('╚══════════════════════════════════════════════════════════════╝')),
  ]);
  var modeColor = config.mode === 'production' ? pc.red : pc.yellow;
  kv('Mode', '[' + modeColor(config.mode.toUpperCase()) + ']');
  divider();
  kv('Proxy URL', 'http://localhost:' + config.port);
  kv('API Target', config.apiTarget);
  kv('Web Static', config.webStaticDir);
  kv('Sign Requests', config.signRequests ? pc.green('enabled') : pc.red('disabled'));
  kv('Strip Signing Key', config.stripSigningKey ? pc.green('yes') : pc.red('no'));
  kv('Verbose Logging', config.verbose ? pc.green('on') : pc.red('off'));
  divider();
}

async function main() {
  printBanner();

  var server = http.createServer(function (req, res) {
    handleRequest(req, res).catch(function (err) {
      error('Unhandled request error', { error: String(err) });
      if (!res.headersSent) {
        res.writeHead(500, { 'content-type': 'application/json' });
        res.end(JSON.stringify({ error: 'proxy_error', message: 'Internal proxy error' }));
      }
    });
  });

  // Attach WebSocket proxy for GraphQL subscriptions.
  attachWebSocketProxy(server);

  // Start listening.
  server.listen(config.port, function () {
    success('Proxy server ready on http://localhost:' + config.port);
    kv('Status', pc.green('Healthy'));
    console.log('');
    console.log('  ' + pc.dim('Press') + ' ' + pc.bold('Ctrl+C') + ' ' + pc.dim('to stop'));
    console.log('');
  });

  // Periodic session cleanup (every hour).
  setInterval(function () {
    cleanupExpiredSessions();
  }, 60 * 60 * 1000);

  // Health stats interval (every 5 min, verbose only).
  if (config.verbose) {
    setInterval(function () {
      info('Session stats', { active: activeSessionCount() });
    }, 5 * 60 * 1000);
  }

  // Graceful shutdown.
  function shutdown(signal) {
    console.log('\n  ' + pc.yellow('@') + ' ' + signal + ' received, shutting down...');
    server.close(function () {
      success('Proxy server closed');
      process.exit(0);
    });
    setTimeout(function () {
      warn('Forced exit after timeout');
      process.exit(1);
    }, 5000);
  }

  process.on('SIGINT', function () { shutdown('SIGINT'); });
  process.on('SIGTERM', function () { shutdown('SIGTERM'); });

  process.on('uncaughtException', function (err) {
    error('Uncaught exception', { error: err.message });
    console.error(err.stack);
  });

  process.on('unhandledRejection', function (reason) {
    error('Unhandled rejection', { reason: String(reason) });
  });
}

/**
 * Route an incoming request to the API proxy or static file server.
 */
async function handleRequest(req, res) {
  var method = req.method || 'GET';
  var url = new URL(req.url || '/', 'http://localhost');
  var pathname = url.pathname;
  var search = url.search;

  // Proxy health check.
  if (pathname === '/proxy-health' || pathname === '/proxy/health') {
    res.writeHead(200, { 'content-type': 'application/json' });
    res.end(JSON.stringify({
      status: 'ok',
      mode: config.mode,
      activeSessions: activeSessionCount(),
      signRequests: config.signRequests,
      apiTarget: config.apiTarget,
      timestamp: new Date().toISOString(),
    }));
    return;
  }

  // Route to API proxy or static file server.
  if (isApiPath(pathname)) {
    var body = await bufferBody(req);

    var result = await handleApiRequest(
      method,
      pathname,
      search,
      req.headers,
      body,
      req.headers.cookie
    );

    if (result.setCookie) {
      res.setHeader('set-cookie', result.setCookie);
    }

    res.writeHead(result.status, result.headers);

    if (result.body && result.body.length > 0) {
      res.end(result.body);
    } else {
      res.end();
    }
  } else {
    // Serve static web app file (no vite — pure fs).
    await serveStaticFile(pathname, res);
  }
}

main().catch(function (err) {
  error('Failed to start proxy server', { error: err.message });
  console.error(err.stack);
  process.exit(1);
});
