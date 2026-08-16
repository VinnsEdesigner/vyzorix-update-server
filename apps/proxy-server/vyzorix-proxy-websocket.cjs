/**
 * vyzorix-proxy-websocket.cjs
 *
 * WebSocket proxy for GraphQL subscriptions — implemented with raw Node.js
 * TCP sockets (no external `ws` dependency).
 *
 * Browsers cannot set custom headers on WebSocket upgrade requests, so the
 * Go API's signing middleware is bypassed for WS connections (by design —
 * see SIGNING_ARCHITECTURE.md). Instead, WS is authenticated via the session
 * cookie.
 *
 * How it works:
 *   1. Browser sends WS upgrade request to proxy
 *   2. Proxy authenticates via _vyz_proxy_sid cookie → resolves session
 *   3. Proxy opens a raw TCP connection to the Go API
 *   4. Proxy rewrites the upgrade request with the Go session cookie
 *   5. Proxy pipes raw bytes bidirectionally (browser ↔ Go API)
 *
 * Raw TCP piping handles all WS framing transparently — the proxy doesn't
 * need to understand WebSocket frames.
 */

'use strict';

const net = require('node:net');
const { URL } = require('node:url');
var cfg = require('./vyzorix-proxy-config.cjs');
var logger = require('./vyzorix-proxy-logger.cjs');
var sessionStore = require('./vyzorix-proxy-session-store.cjs');

var config = cfg.config;
var debug = logger.debug;
var error = logger.error;
var info = logger.info;
var getSession = sessionStore.getSession;

/**
 * Attach a WebSocket upgrade handler to an HTTP server.
 * @param {http.Server} server
 */
function attachWebSocketProxy(server) {
  server.on('upgrade', function (req, socket, head) {
    handleUpgrade(req, socket, head).catch(function (err) {
      error('WS upgrade error', { error: err.message });
      try {
        socket.write('HTTP/1.1 500 Internal Server Error\r\n\r\n');
        socket.destroy();
      } catch { /* socket already dead */ }
    });
  });

  info('WebSocket proxy ready for GraphQL subscriptions');
}

async function handleUpgrade(req, socket, head) {
  var rawPathname = new URL(req.url || '/', 'http://localhost').pathname;

  // Strip the /api prefix (same as HTTP request handling) so the Go API
  // receives /:org/graphql/ws or /v1/device/:imei/stream without the prefix.
  var pathname = rawPathname;
  if (pathname.startsWith('/api/') || pathname === '/api') {
    pathname = pathname.slice(4);
  }

  // Determine the type of WebSocket upgrade:
  //   - GraphQL subscriptions:  /:org/graphql/ws
  //   - Device streaming:       /v1/device/:imei/stream
  var isGraphQLWs = pathname.endsWith('/graphql/ws');
  var isDeviceStream = /^\/v1\/device\/[^/]+\/stream$/.test(pathname);

  if (!isGraphQLWs && !isDeviceStream) {
    // Not our route — close the socket.
    socket.write('HTTP/1.1 404 Not Found\r\n\r\n');
    socket.destroy();
    return;
  }

  var cookieHeader = req.headers.cookie;
  var sess = getSession(cookieHeader);

  // GraphQL subscriptions authenticate via session cookie. Device streams
  // authenticate via HMAC query params (hmac_timestamp, hmac_nonce,
  // hmac_signature, device_id) — no session required.
  if (isGraphQLWs && (!sess || !sess.apiSessionCookie)) {
    debug('WS upgrade rejected: no session for GraphQL subscription', { path: pathname });
    socket.write('HTTP/1.1 401 Unauthorized\r\n\r\n');
    socket.destroy();
    return;
  }

  // Parse the target.
  var apiTargetUrl = new URL(config.apiTarget);
  var apiPort = parseInt(apiTargetUrl.port, 10) || (apiTargetUrl.protocol === 'https:' ? 443 : 80);
  var apiHost = apiTargetUrl.hostname;

  // Build the rewritten upgrade request to send to the Go API.
  // Use the prefix-stripped path so the Go API sees the correct route.
  var search = new URL(req.url || '/', 'http://localhost').search;
  var forwardUrl = search ? pathname + search : pathname;
  var reqLine = 'GET ' + forwardUrl + ' HTTP/1.1\r\n';
  var headerLines = [];

  // Copy essential WS handshake headers from browser request.
  var headersToCopy = [
    'host', 'upgrade', 'connection', 'sec-websocket-key',
    'sec-websocket-version', 'sec-websocket-protocol',
    'sec-websocket-extensions', 'origin',
  ];

  headersToCopy.forEach(function (h) {
    var val = req.headers[h];
    if (val !== undefined) {
      // Replace host with the API server's host.
      if (h === 'host') {
        headerLines.push('Host: ' + apiHost + ':' + apiPort);
      } else {
        headerLines.push(h + ': ' + val);
      }
    }
  });

  // Inject the Go session cookie for GraphQL subscriptions. Device streams
  // carry their HMAC credentials in query params and need no session cookie.
  if (isGraphQLWs && sess && sess.apiSessionCookie) {
    headerLines.push('Cookie: vyz_session=' + sess.apiSessionCookie);
  }

  // Override origin to match the API target (some servers check origin).
  if (!req.headers.origin) {
    headerLines.push('Origin: ' + config.apiTarget);
  }

  var rawRequest = reqLine + headerLines.join('\r\n') + '\r\n\r\n';

  debug('WS upgrade forwarding', { path: pathname, target: apiHost + ':' + apiPort });

  // Open a raw TCP connection to the Go API.
  var upstream = net.connect(apiPort, apiHost, function () {
    // Send the rewritten upgrade request.
    upstream.write(rawRequest);

    // If there's buffered data from the browser (head), forward it.
    if (head && head.length > 0) {
      upstream.write(head);
    }

    debug('WS upstream connected', { path: pathname });

    // Pipe bidirectionally: browser ↔ Go API.
    // Raw TCP piping handles all WS framing transparently.
    socket.pipe(upstream);
    upstream.pipe(socket);
  });

  // Handle errors on both sockets.
  var cleaned = false;
  function cleanup(reason) {
    if (cleaned) return;
    cleaned = true;
    debug('WS connection closed', { path: pathname, reason: reason });
    try { socket.destroy(); } catch { /* already destroyed */ }
    try { upstream.destroy(); } catch { /* already destroyed */ }
  }

  socket.on('error', function (err) {
    error('WS browser socket error', { error: err.message, path: pathname });
    cleanup('browser_error');
  });

  upstream.on('error', function (err) {
    error('WS upstream socket error', { error: err.message, path: pathname });
    cleanup('upstream_error');
  });

  socket.on('close', function () { cleanup('browser_close'); });
  upstream.on('close', function () { cleanup('upstream_close'); });
}

module.exports = { attachWebSocketProxy };
