/**
 * vyzorix-proxy-session-store.cjs
 *
 * Server-side session store.
 *
 * Each session holds the per-session signing key, auth tokens, CSRF token,
 * and organization context — none of which are exposed to the browser. The
 * browser only holds an opaque httpOnly cookie (_vyz_proxy_sid) that maps
 * to a session ID in this store.
 *
 * Sessions expire after sessionTtlMs (default 24h) and are cleaned up lazily.
 */

'use strict';

const crypto = require('node:crypto');
var cfg = require('./vyzorix-proxy-config.cjs');
var logger = require('./vyzorix-proxy-logger.cjs');
var config = cfg.config;
var debug = logger.debug;
var warn = logger.warn;

function newSessionId() {
  return crypto.randomBytes(32).toString('hex');
}

function createSession() {
  var now = Date.now();
  return {
    id: newSessionId(),
    createdAt: now,
    lastAccess: now,
    signingKey: null,
    accessToken: null,
    refreshToken: null,
    tokenExpiresAt: null,
    csrfToken: null,
    csrfCookie: null,
    organizationId: null,
    operatorId: null,
    apiSessionCookie: null,
  };
}

var sessions = new Map();

/**
 * Get or create a session for the given cookie header.
 * @param {string|undefined} cookieHeader
 * @returns {ProxySession}
 */
function getOrCreateSession(cookieHeader) {
  var existingId = extractCookie(cookieHeader, config.sessionCookieName);
  if (existingId) {
    var sess = sessions.get(existingId);
    if (sess && !isExpired(sess)) {
      sess.lastAccess = Date.now();
      return sess;
    }
    if (sess) {
      debug('Session expired, removing', { id: existingId.slice(0, 12) });
      sessions.delete(existingId);
    }
  }
  var newSess = createSession();
  sessions.set(newSess.id, newSess);
  debug('Created new proxy session', { id: newSess.id.slice(0, 12) });
  return newSess;
}

/**
 * Get an existing session without creating one.
 * @param {string|undefined} cookieHeader
 * @returns {ProxySession|null}
 */
function getSession(cookieHeader) {
  var id = extractCookie(cookieHeader, config.sessionCookieName);
  if (!id) return null;
  var sess = sessions.get(id);
  if (!sess || isExpired(sess)) {
    if (sess) sessions.delete(id);
    return null;
  }
  sess.lastAccess = Date.now();
  return sess;
}

/**
 * Delete a session (called on logout).
 */
function destroySession(cookieHeader) {
  var id = extractCookie(cookieHeader, config.sessionCookieName);
  if (id) {
    sessions.delete(id);
    debug('Destroyed proxy session', { id: id.slice(0, 12) });
  }
}

/**
 * Build the Set-Cookie header value for a session.
 */
function sessionCookieHeader(sess) {
  var flags = [
    config.sessionCookieName + '=' + sess.id,
    'Path=/',
    'HttpOnly',
    'SameSite=Lax',
    'Max-Age=' + Math.floor(config.sessionTtlMs / 1000),
  ];
  if (config.mode === 'production') flags.push('Secure');
  return flags.join('; ');
}

/**
 * Clear the session cookie header value.
 */
function clearSessionCookieHeader() {
  return config.sessionCookieName + '=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0';
}

function isExpired(sess) {
  return Date.now() - sess.lastAccess > config.sessionTtlMs;
}

/**
 * Lazy cleanup of expired sessions.
 */
function cleanupExpiredSessions() {
  var now = Date.now();
  var removed = 0;
  sessions.forEach(function (sess, id) {
    if (now - sess.lastAccess > config.sessionTtlMs) {
      sessions.delete(id);
      removed++;
    }
  });
  if (removed > 0) {
    warn('Cleaned up expired sessions', { count: removed });
  }
  return removed;
}

/**
 * Parse a cookie header and extract a named cookie value.
 */
function extractCookie(cookieHeader, name) {
  if (!cookieHeader) return null;
  var match = cookieHeader.match(new RegExp('(?:^|;\\s*)' + name + '=([^;]+)'));
  return match ? match[1] : null;
}

/**
 * Parse all Set-Cookie headers from an upstream response.
 * @returns {Map<string, string>}
 */
function parseSetCookies(setCookieHeaders) {
  var cookies = new Map();
  if (!setCookieHeaders) return cookies;
  if (!Array.isArray(setCookieHeaders)) setCookieHeaders = [setCookieHeaders];
  setCookieHeaders.forEach(function (header) {
    var match = header.match(/^([^=]+)=([^;]+)/);
    if (match) {
      cookies.set(match[1].trim(), match[2].trim());
    }
  });
  return cookies;
}

function activeSessionCount() {
  return sessions.size;
}

module.exports = {
  getOrCreateSession,
  getSession,
  destroySession,
  sessionCookieHeader,
  clearSessionCookieHeader,
  cleanupExpiredSessions,
  extractCookie,
  parseSetCookies,
  activeSessionCount,
};
