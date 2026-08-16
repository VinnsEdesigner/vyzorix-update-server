/**
 * vyzorix-proxy-core.cjs
 *
 * Core HTTP proxy: receives browser requests, injects auth/signing headers,
 * forwards to the Go API, intercepts auth responses, and returns to browser.
 *
 * Request flow:
 *   1. Resolve proxy session from httpOnly cookie
 *   2. Buffer request body
 *   3. If signing is enabled and endpoint requires signing:
 *      a. Ensure CSRF token is available
 *      b. Compute HMAC signature over (method, path, nonce, timestamp, body)
 *      c. Inject X-Vyzorix-* headers + X-CSRF-Token + Cookie
 *   4. Forward to Go API via node:http
 *   5. If response is from an auth endpoint, intercept and capture credentials
 *   6. Return response (possibly modified) to browser with session cookie
 */

'use strict';

const http = require('node:http');
const { URL } = require('node:url');
var cfg = require('./vyzorix-proxy-config.cjs');
var logger = require('./vyzorix-proxy-logger.cjs');
var sessionStore = require('./vyzorix-proxy-session-store.cjs');
var signing = require('./vyzorix-proxy-signing.cjs');
var authInterceptor = require('./vyzorix-proxy-auth-interceptor.cjs');

var config = cfg.config;
var isAuthEndpoint = cfg.isAuthEndpoint;
var isAuthInterceptEndpoint = cfg.isAuthInterceptEndpoint;
var isStateChanging = cfg.isStateChanging;
var logRequest = logger.logRequest;
var info = logger.info;
var warn = logger.warn;
var error = logger.error;
var debug = logger.debug;
var getOrCreateSession = sessionStore.getOrCreateSession;
var getSession = sessionStore.getSession;
var sessionCookieHeader = sessionStore.sessionCookieHeader;
var clearSessionCookieHeader = sessionStore.clearSessionCookieHeader;
var extractCookie = sessionStore.extractCookie;
var signRequest = signing.signRequest;
var deriveAPIKeySigningSecret = signing.deriveAPIKeySigningSecret;
var interceptAuthResponse = authInterceptor.interceptAuthResponse;
var ensureCsrfToken = authInterceptor.ensureCsrfToken;

// Hop-by-hop headers that should NOT be forwarded.
var HOP_BY_HOP = new Set([
  'connection', 'keep-alive', 'transfer-encoding', 'te',
  'trailer', 'upgrade', 'proxy-authorization', 'proxy-authenticate',
]);

// Headers to forward from browser to Go API.
var FORWARD_HEADERS = [
  'content-type', 'accept', 'accept-encoding', 'accept-language',
  'user-agent', 'x-api-key', 'x-organization-id', 'x-idempotency-key',
  'x-requested-with',
];

/**
 * Handle an incoming API request: sign, forward, intercept, return.
 * @param {string} method
 * @param {string} pathname
 * @param {string} search - query string including '?'
 * @param {object} reqHeaders - incoming headers (lowercased keys)
 * @param {Buffer} reqBody - buffered request body
 * @param {string|undefined} cookieHeader - browser Cookie header
 * @returns {Promise<{status, headers, body, setCookie}>}
 */
async function handleApiRequest(method, pathname, search, reqHeaders, reqBody, cookieHeader) {
  var startTime = Date.now();
  var sess = getOrCreateSession(cookieHeader);

  // Strip the /api prefix before forwarding to the Go API.
  // The REST client uses baseURL "/api" (VITE_API_URL), so browser requests
  // arrive as /api/v1/devices, /api/{org}/graphql, etc. The Go API's tenant
  // routes are registered at /v1/ (not /api/v1/), so the prefix must be
  // removed. Paths without /api (e.g. /v1/devices, /health) are forwarded
  // unchanged.
  var apiStrippedPath = pathname;
  if (pathname.startsWith('/api/') || pathname === '/api') {
    apiStrippedPath = pathname.slice(4); // remove "/api" → keeps leading "/"
  }
  var fullPath = search ? apiStrippedPath + search : apiStrippedPath;

  // Determine if this request needs signing.
  // Signing is required when:
  //   - request signing is enabled, AND
  //   - the endpoint is not an auth endpoint, AND
  //   - either there is a session signing key (session auth) OR an API key is
  //     present (API key auth — the signing secret is derived from the key).
  var hasApiKey = typeof reqHeaders['x-api-key'] === 'string' && reqHeaders['x-api-key'].startsWith('vxyz_');
  var needsSigning = config.signRequests && !isAuthEndpoint(apiStrippedPath) && (sess.signingKey !== null || hasApiKey);
  // CSRF is required for ALL state-changing requests, including auth endpoints.
  var needsCsrf = isStateChanging(method);

  // Build outgoing headers.
  var outHeaders = {};

  FORWARD_HEADERS.forEach(function (h) {
    var val = reqHeaders[h];
    if (typeof val === 'string') {
      outHeaders[h] = val;
    } else if (Array.isArray(val)) {
      outHeaders[h] = val.join(', ');
    }
  });

  // Fetch CSRF token + cookie BEFORE building the Cookie header, so the
  // double-submit cookie is available for injection.
  if (needsCsrf) {
    var csrf = await ensureCsrfToken(sess, config.apiTarget);
    if (csrf) {
      outHeaders['x-csrf-token'] = csrf;
    }
  }

  // Inject Go API session cookie (captured during login).
  var cookies = [];
  if (sess.apiSessionCookie) {
    cookies.push('vyz_session=' + sess.apiSessionCookie);
  }
  // Inject CSRF cookie (double-submit: header + cookie must match).
  // Go API uses _csrf cookie name (gorilla/csrf default).
  var csrfCookieVal = sess.csrfCookie || extractCookie(cookieHeader, '_csrf') || extractCookie(cookieHeader, 'vyz_csrf');
  if (csrfCookieVal) {
    cookies.push('_csrf=' + csrfCookieVal);
  }
  if (cookies.length > 0) {
    outHeaders['cookie'] = cookies.join('; ');
  }

  // Inject auth token if available (for token-based auth).
  if (sess.accessToken) {
    outHeaders['authorization'] = 'Bearer ' + sess.accessToken;
  }

  // Inject organization context.
  if (sess.organizationId && !outHeaders['x-organization-id']) {
    outHeaders['x-organization-id'] = sess.organizationId;
  }

  // Sign the request if needed.
  var signed = false;
  if (needsSigning) {
    var signingKey = sess.signingKey;
    if (hasApiKey) {
      // API key auth: derive the signing secret from the full key.
      // This MUST match the Go server's deriveAPIKeySigningSecret.
      signingKey = deriveAPIKeySigningSecret(reqHeaders['x-api-key']);
      debug('Using API-key-derived signing secret');
    }

    if (signingKey) {
      var signingHeaders = signRequest(method, fullPath, reqBody, signingKey);
      outHeaders['x-vyzorix-timestamp'] = signingHeaders['X-Vyzorix-Timestamp'];
      outHeaders['x-vyzorix-nonce'] = signingHeaders['X-Vyzorix-Nonce'];
      outHeaders['x-vyzorix-signature'] = signingHeaders['X-Vyzorix-Signature'];
      signed = true;
    }
  }

  // Forward to Go API.
  var apiTargetUrl = new URL(config.apiTarget);
  var forwardOptions = {
    hostname: apiTargetUrl.hostname,
    port: apiTargetUrl.port,
    path: fullPath,
    method: method,
    headers: outHeaders,
  };

  var upstreamResult;
  try {
    upstreamResult = await forwardRequest(forwardOptions, reqBody);
  } catch (err) {
    error('Upstream connection error', { error: err.message, path: fullPath });
    return {
      status: 502,
      headers: { 'content-type': 'application/json' },
      body: Buffer.from(JSON.stringify({
        error: 'proxy_upstream_error',
        message: 'Cannot reach API server: ' + err.message,
      })),
      setCookie: sessionCookieHeader(sess),
    };
  }

  var status = upstreamResult.status;
  var headers = upstreamResult.headers;
  var body = upstreamResult.body;

  // Build header map for interceptor.
  var headerMap = new Map();
  for (var key in headers) {
    if (headers[key] !== undefined) {
      headerMap.set(key.toLowerCase(), headers[key]);
    }
  }

  // Intercept auth responses (includes org select, logout, etc.).
  var responseBody = body;
  var intercepted = false;
  if (isAuthInterceptEndpoint(apiStrippedPath)) {
    var result = interceptAuthResponse(method, apiStrippedPath, status, headerMap, body, sess);
    responseBody = result.body;
    intercepted = true;
  }

  // Build response headers (filter hop-by-hop and cookies).
  var resHeaders = {};
  headerMap.forEach(function (value, key) {
    if (HOP_BY_HOP.has(key)) return;
    if (key === 'set-cookie') {
      // Don't forward Go API session cookies — proxy manages cookies.
      // Forward non-session cookies like csrf.
      var cookieArr = Array.isArray(value) ? value : [value];
      var safeCookies = cookieArr.filter(function (c) {
        var name = c.split('=')[0].trim();
        return name !== 'vyz_session' && name !== 'session';
      });
      if (safeCookies.length > 0) {
        resHeaders['set-cookie'] = safeCookies;
      }
      return;
    }
    if (key === 'content-length') return; // recalculated below
    resHeaders[key] = value;
  });

  resHeaders['content-length'] = String(responseBody.length);

  var durationMs = Date.now() - startTime;
  logRequest(method, fullPath, status, durationMs, { signed: signed, intercepted: intercepted });

  // Handle logout: destroy session.
  if (apiStrippedPath === '/v1/auth/logout' && method === 'POST' && status >= 200 && status < 300) {
    return {
      status: status,
      headers: resHeaders,
      body: responseBody,
      setCookie: clearSessionCookieHeader(),
    };
  }

  return {
    status: status,
    headers: resHeaders,
    body: responseBody,
    setCookie: sessionCookieHeader(sess),
  };
}

/**
 * Forward a request to the Go API using node:http.
 * @returns {Promise<{status, headers, body}>}
 */
function forwardRequest(options, body) {
  return new Promise(function (resolve, reject) {
    var req = http.request(options, function (res) {
      var chunks = [];
      res.on('data', function (chunk) { chunks.push(chunk); });
      res.on('end', function () {
        resolve({
          status: res.statusCode || 500,
          headers: res.headers,
          body: Buffer.concat(chunks),
        });
      });
      res.on('error', reject);
    });

    req.on('error', function (err) {
      reject(err);
    });

    if (body && body.length > 0) {
      req.write(body);
    }
    req.end();
  });
}

/**
 * Buffer the entire request body into a single Buffer.
 * @returns {Promise<Buffer>}
 */
function bufferBody(req) {
  return new Promise(function (resolve, reject) {
    var chunks = [];
    req.on('data', function (chunk) { chunks.push(chunk); });
    req.on('end', function () { resolve(Buffer.concat(chunks)); });
    req.on('error', reject);
  });
}

module.exports = { handleApiRequest, bufferBody, forwardRequest };
