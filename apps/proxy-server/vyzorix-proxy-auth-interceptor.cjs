/**
 * vyzorix-proxy-auth-interceptor.cjs
 *
 * Inspects responses from Go API auth endpoints and extracts auth artifacts
 * (signing key, tokens, cookies) into the server-side session. This keeps
 * sensitive credentials out of the browser — the browser only holds the
 * opaque _vyz_proxy_sid cookie.
 *
 * Supported interception points:
 *   - POST /v1/auth/login         → capture signing_key, tokens, session cookie
 *   - POST /v1/auth/login/tokens  → same (token-based variant)
 *   - POST /v1/auth/mfa/verify    → capture signing_key from MFA completion
 *   - POST /v1/auth/refresh       → capture new access_token
 *   - POST /v1/auth/organizations/select → capture organization_id
 *   - POST /v1/auth/logout        → destroy session credentials
 */

'use strict';

var cfg = require('./vyzorix-proxy-config.cjs');
var logger = require('./vyzorix-proxy-logger.cjs');
var sessionStore = require('./vyzorix-proxy-session-store.cjs');
var config = cfg.config;
var debug = logger.debug;
var info = logger.info;
var warn = logger.warn;
var parseSetCookies = sessionStore.parseSetCookies;

/**
 * Intercept auth responses and update session.
 * @returns {{modified: boolean, body: Buffer}} Possibly-modified body.
 */
function interceptAuthResponse(method, pathname, status, responseHeaders, body, sess) {
  if (status < 200 || status >= 300) {
    return { modified: false, body: body };
  }

  // Extract Set-Cookie headers from upstream response.
  var setCookieRaw = responseHeaders.get('set-cookie');
  var setCookies = parseSetCookies(
    Array.isArray(setCookieRaw) ? setCookieRaw : setCookieRaw ? [setCookieRaw] : undefined
  );

  // ── Login ──
  if (method === 'POST' && (pathname === '/v1/auth/login' || pathname === '/v1/auth/login/tokens')) {
    return interceptLogin(body, setCookies, sess);
  }

  // ── MFA verify ──
  if (method === 'POST' && pathname.includes('/v1/auth/mfa/verify')) {
    return interceptMfaVerify(body, setCookies, sess);
  }

  // ── Token refresh ──
  if (method === 'POST' && pathname === '/v1/auth/refresh') {
    return interceptRefresh(body, sess);
  }

  // ── Organization select ──
  if (method === 'POST' && pathname === '/v1/auth/organizations/select') {
    return interceptOrgSelect(body, sess);
  }

  // ── Logout ──
  if (method === 'POST' && pathname === '/v1/auth/logout') {
    info('Logout intercepted, clearing session credentials', { operatorId: sess.operatorId });
    sess.signingKey = null;
    sess.accessToken = null;
    sess.refreshToken = null;
    sess.apiSessionCookie = null;
    return { modified: false, body: body };
  }

  return { modified: false, body: body };
}

function interceptLogin(body, setCookies, sess) {
  var parsed;
  try {
    parsed = JSON.parse(body.toString('utf8'));
  } catch {
    warn('Login response is not valid JSON, skipping interception');
    return { modified: false, body: body };
  }

  // Capture signing key.
  if (typeof parsed.signing_key === 'string' && parsed.signing_key.length > 0) {
    sess.signingKey = parsed.signing_key;
    info('Captured signing key from login response', { keyLen: parsed.signing_key.length });
  }

  // Capture tokens (token-based login).
  if (typeof parsed.access_token === 'string') sess.accessToken = parsed.access_token;
  if (typeof parsed.refresh_token === 'string') sess.refreshToken = parsed.refresh_token;
  if (typeof parsed.expires_at === 'number') sess.tokenExpiresAt = parsed.expires_at * 1000;

  // Capture operator info.
  if (typeof parsed.operator_id === 'string') sess.operatorId = parsed.operator_id;
  if (parsed.operator && typeof parsed.operator.id === 'string') sess.operatorId = parsed.operator.id;

  // Capture organization ID (auto-selected during login if user has one).
  var orgId = parsed.last_organization_id
    || (parsed.selected_organization && parsed.selected_organization.id)
    || (parsed.organization_id);
  if (typeof orgId === 'string' && orgId.length > 0) {
    sess.organizationId = orgId;
    debug('Captured organization ID from login response', { orgId: orgId });
  }

  // Capture Go API session cookie (for cookie-based auth forwarding).
  var goSessionCookie = setCookies.get('vyz_session') || setCookies.get('session');
  if (goSessionCookie) {
    sess.apiSessionCookie = goSessionCookie;
    debug('Captured Go API session cookie');
  }

  // Optionally strip signing_key from response before forwarding to browser.
  if (config.stripSigningKey && typeof parsed.signing_key === 'string') {
    delete parsed.signing_key;
    var modifiedBody = Buffer.from(JSON.stringify(parsed), 'utf8');
    info('Stripped signing_key from login response (BFF mode)');
    return { modified: true, body: modifiedBody };
  }

  return { modified: false, body: body };
}

function interceptMfaVerify(body, setCookies, sess) {
  var parsed;
  try {
    parsed = JSON.parse(body.toString('utf8'));
  } catch {
    return { modified: false, body: body };
  }

  if (parsed.signing_key && typeof parsed.signing_key === 'string') {
    sess.signingKey = parsed.signing_key;
    info('Captured signing key from MFA verify response');
  }
  if (parsed.access_token && typeof parsed.access_token === 'string') {
    sess.accessToken = parsed.access_token;
  }
  if (parsed.refresh_token && typeof parsed.refresh_token === 'string') {
    sess.refreshToken = parsed.refresh_token;
  }

  var goSessionCookie = setCookies.get('vyz_session') || setCookies.get('session');
  if (goSessionCookie) {
    sess.apiSessionCookie = goSessionCookie;
  }

  if (config.stripSigningKey && typeof parsed.signing_key === 'string') {
    delete parsed.signing_key;
    return { modified: true, body: Buffer.from(JSON.stringify(parsed), 'utf8') };
  }

  return { modified: false, body: body };
}

function interceptRefresh(body, sess) {
  var parsed;
  try {
    parsed = JSON.parse(body.toString('utf8'));
  } catch {
    return { modified: false, body: body };
  }

  if (typeof parsed.access_token === 'string') {
    sess.accessToken = parsed.access_token;
    info('Captured refreshed access token');
  }
  if (typeof parsed.refresh_token === 'string') {
    sess.refreshToken = parsed.refresh_token;
  }
  if (typeof parsed.expires_at === 'number') {
    sess.tokenExpiresAt = parsed.expires_at * 1000;
  }

  return { modified: false, body: body };
}

function interceptOrgSelect(body, sess) {
  var parsed;
  try {
    parsed = JSON.parse(body.toString('utf8'));
  } catch {
    return { modified: false, body: body };
  }

  var orgId = parsed.organization_id || parsed.id;
  if (typeof orgId === 'string') {
    sess.organizationId = orgId;
    info('Captured organization selection', { orgId: orgId });
  }

  return { modified: false, body: body };
}

/**
 * Auto-fetch a CSRF token from the Go API if the session doesn't have one.
 * Called lazily before state-changing requests.
 *
 * Captures both the token (for X-CSRF-Token header) and the _csrf cookie
 * (for Cookie header), since the Go API uses double-submit CSRF validation.
 * @returns {Promise<string|null>}
 */
async function ensureCsrfToken(sess, apiTarget) {
  if (sess.csrfToken) return sess.csrfToken;

  try {
    var headers = {};
    if (sess.apiSessionCookie) {
      headers['Cookie'] = 'vyz_session=' + sess.apiSessionCookie;
    }

    var res = await fetch(apiTarget + '/v1/auth/csrf-token', {
      method: 'GET',
      headers: headers,
    });

    if (!res.ok) {
      warn('CSRF token fetch failed', { status: res.status });
      return null;
    }

    // Capture _csrf cookie from Set-Cookie header (for double-submit).
    // Node's fetch (undici) exposes set-cookie via getSetCookie() (array) or
    // get('set-cookie') (joined string).
    var setCookieArr = [];
    if (typeof res.headers.getSetCookie === 'function') {
      setCookieArr = res.headers.getSetCookie();
    } else {
      var sc = res.headers.get('set-cookie');
      if (sc) setCookieArr = [sc];
    }
    for (var i = 0; i < setCookieArr.length; i++) {
      var cookieMatch = setCookieArr[i].match(/_csrf=([^;]+)/);
      if (cookieMatch) {
        sess.csrfCookie = cookieMatch[1];
        debug('Captured _csrf cookie from CSRF token fetch');
        break;
      }
    }

    var data = await res.json();
    if (data.csrf_token) {
      sess.csrfToken = data.csrf_token;
      debug('Auto-fetched CSRF token');
      return data.csrf_token;
    }
  } catch (err) {
    warn('CSRF token fetch error', { error: String(err) });
  }

  return null;
}

module.exports = { interceptAuthResponse, ensureCsrfToken };
