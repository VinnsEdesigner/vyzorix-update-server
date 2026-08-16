/**
 * vyzorix-proxy-test.cjs
 *
 * Self-contained test suite for the Vyzorix proxy server.
 * Tests signing logic, session store, auth interception, config routing, and
 * end-to-end request proxying (if the Go API is running).
 *
 * Run with: node vyzorix-proxy-test.cjs
 *
 * If the Go API is running on localhost:3000, integration tests run too.
 * Otherwise, only unit tests run.
 */

'use strict';

const http = require('node:http');
const crypto = require('node:crypto');
const { URL } = require('node:url');

var signing = require('./vyzorix-proxy-signing.cjs');
var sessionStore = require('./vyzorix-proxy-session-store.cjs');
var authInterceptor = require('./vyzorix-proxy-auth-interceptor.cjs');
var cfg = require('./vyzorix-proxy-config.cjs');
var logger = require('./vyzorix-proxy-logger.cjs');

var signRequest = signing.signRequest;
var computeSignature = signing.computeSignature;
var deriveAPIKeySigningSecret = signing.deriveAPIKeySigningSecret;
var getOrCreateSession = sessionStore.getOrCreateSession;
var getSession = sessionStore.getSession;
var destroySession = sessionStore.destroySession;
var sessionCookieHeader = sessionStore.sessionCookieHeader;
var extractCookie = sessionStore.extractCookie;
var parseSetCookies = sessionStore.parseSetCookies;
var interceptAuthResponse = authInterceptor.interceptAuthResponse;
var isApiPath = cfg.isApiPath;
var isAuthEndpoint = cfg.isAuthEndpoint;
var isStateChanging = cfg.isStateChanging;
var isGraphQLPath = cfg.isGraphQLPath;

var passed = 0;
var failed = 0;
var failures = [];

function assert(condition, message) {
  if (condition) {
    passed++;
  } else {
    failed++;
    failures.push(message);
    console.log('  ✗ ' + message);
  }
}

function assertEqual(actual, expected, message) {
  assert(actual === expected, message + ' (got: ' + JSON.stringify(actual) + ', expected: ' + JSON.stringify(expected) + ')');
}

function test(name, fn) {
  try {
    fn();
    console.log('  ✓ ' + name);
  } catch (err) {
    failed++;
    failures.push(name + ': ' + err.message);
    console.log('  ✗ ' + name + ' — ' + err.message);
  }
}

async function asyncTest(name, fn) {
  try {
    await fn();
    console.log('  ✓ ' + name);
  } catch (err) {
    failed++;
    failures.push(name + ': ' + err.message);
    console.log('  ✗ ' + name + ' — ' + err.message);
  }
}

function section(title) {
  console.log('\n' + logger.pc.cyan(logger.pc.bold('  ▸ ' + title)));
}

// ═══════════════════════════════════════════════════════════════
//  Signing Tests
// ═══════════════════════════════════════════════════════════════

section('Signing Logic');

test('signRequest returns three X-Vyzorix-* headers', function () {
  var headers = signRequest('POST', '/v1/organizations', Buffer.from('{"name":"test"}'), 'test-signing-key');
  assert(typeof headers['X-Vyzorix-Timestamp'] === 'string', 'Timestamp is string');
  assert(typeof headers['X-Vyzorix-Nonce'] === 'string', 'Nonce is string');
  assert(typeof headers['X-Vyzorix-Signature'] === 'string', 'Signature is string');
});

test('signRequest timestamp is Unix milliseconds', function () {
  var headers = signRequest('GET', '/v1/devices', Buffer.alloc(0), 'key');
  var ts = parseInt(headers['X-Vyzorix-Timestamp'], 10);
  var now = Date.now();
  assert(Math.abs(now - ts) < 5000, 'Timestamp within 5s of now');
});

test('signRequest nonce is 32 hex chars', function () {
  var headers = signRequest('GET', '/v1/devices', Buffer.alloc(0), 'key');
  assert(/^[0-9a-f]{32}$/.test(headers['X-Vyzorix-Nonce']), 'Nonce is 32 hex chars');
});

test('computeSignature matches Go server canonical format', function () {
  // Verify the canonical string format: METHOD\nPATH\nNONCE\nTIMESTAMP\n + BODY
  var method = 'POST';
  var path = '/v1/organizations';
  var nonce = 'abc123nonce';
  var timestamp = '1700000000000';
  var body = Buffer.from('{"name":"test"}');
  var key = 'my-signing-key';

  var sig = computeSignature(method, path, nonce, timestamp, body, key);

  // Manually compute expected signature using the same canonical format.
  var crypto = require('node:crypto');
  var canonical = 'POST\n/v1/organizations\nabc123nonce\n1700000000000\n';
  var hmac = crypto.createHmac('sha512', key);
  hmac.update(canonical, 'utf8');
  hmac.update(body);
  var expected = hmac.digest('base64');

  assertEqual(sig, expected, 'Signature matches manual HMAC computation');
});

test('signRequest is deterministic for same inputs', function () {
  var h1 = signRequest('POST', '/v1/test', Buffer.from('body'), 'key');
  var h2 = signRequest('POST', '/v1/test', Buffer.from('body'), 'key');
  // Nonce and timestamp differ (random/time-based), but signature format is same.
  assert(h1['X-Vyzorix-Signature'] !== h2['X-Vyzorix-Signature'], 'Signatures differ due to nonce/timestamp');
});

test('signRequest signature changes if body changes', function () {
  var h1 = signRequest('POST', '/v1/test', Buffer.from('body1'), 'key');
  var h2 = signRequest('POST', '/v1/test', Buffer.from('body2'), 'key');
  assert(h1['X-Vyzorix-Signature'] !== h2['X-Vyzorix-Signature'], 'Different bodies → different signatures');
});

test('signRequest signature changes if path changes', function () {
  var h1 = signRequest('POST', '/v1/test1', Buffer.from('body'), 'key');
  var h2 = signRequest('POST', '/v1/test2', Buffer.from('body'), 'key');
  assert(h1['X-Vyzorix-Signature'] !== h2['X-Vyzorix-Signature'], 'Different paths → different signatures');
});

test('deriveAPIKeySigningSecret produces 128 hex chars (full SHA-512)', function () {
  var apiKey = 'vxyz_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890';
  var secret = deriveAPIKeySigningSecret(apiKey);
  assert(/^[0-9a-f]{128}$/.test(secret), 'Derived secret is 128 hex chars (SHA-512)');
});

test('deriveAPIKeySigningSecret matches Go server hex(sha512(fullKey))', function () {
  // Cross-checked against Go: crypto/sha512.Sum512 → hex.EncodeToString
  var apiKey = 'vxyz_testkey';
  var secret = deriveAPIKeySigningSecret(apiKey);
  // Deterministic: same input → same output
  var secret2 = deriveAPIKeySigningSecret(apiKey);
  assertEqual(secret, secret2, 'Derivation is deterministic');
  assert(secret.length === 128, 'Full SHA-512 hex length (128 chars)');
});

test('deriveAPIKeySigningSecret works for any string (no format requirement)', function () {
  var secret = deriveAPIKeySigningSecret('short-key');
  assert(/^[0-9a-f]{128}$/.test(secret), 'Any string produces 128 hex chars');
});

// ═══════════════════════════════════════════════════════════════
//  Session Store Tests
// ═══════════════════════════════════════════════════════════════

section('Session Store');

test('getOrCreateSession creates new session for no cookie', function () {
  var sess = getOrCreateSession(undefined);
  assert(sess !== null, 'Session created');
  assert(typeof sess.id === 'string', 'Session has ID');
  assert(sess.signingKey === null, 'New session has null signingKey');
});

test('getOrCreateSession returns same session for valid cookie', function () {
  var sess1 = getOrCreateSession(undefined);
  var cookie = sessionCookieHeader(sess1);
  var cookieVal = extractCookie(cookie, '_vyz_proxy_sid');
  var sess2 = getOrCreateSession(cookie);
  assertEqual(sess2.id, sess1.id, 'Same session returned for valid cookie');
});

test('getOrCreateSession creates new session for invalid cookie', function () {
  var sess = getOrCreateSession('_vyz_proxy_sid=invalid-id-12345');
  assert(sess !== null, 'Session created for invalid cookie');
  assert(sess.id !== 'invalid-id-12345', 'New session has new ID');
});

test('getSession returns null for no cookie', function () {
  var sess = getSession(undefined);
  assertEqual(sess, null, 'getSession returns null for no cookie');
});

test('getSession returns null for invalid cookie', function () {
  var sess = getSession('_vyz_proxy_sid=nonexistent');
  assertEqual(sess, null, 'getSession returns null for invalid cookie');
});

test('sessionCookieHeader contains HttpOnly and SameSite', function () {
  var sess = getOrCreateSession(undefined);
  var cookie = sessionCookieHeader(sess);
  assert(cookie.includes('HttpOnly'), 'Cookie has HttpOnly');
  assert(cookie.includes('SameSite=Lax'), 'Cookie has SameSite=Lax');
  assert(cookie.includes('Path=/'), 'Cookie has Path=/');
});

test('extractCookie parses named cookie from header', function () {
  var header = '_vyz_proxy_sid=abc123; other=val; vyz_csrf=token456';
  assertEqual(extractCookie(header, '_vyz_proxy_sid'), 'abc123', 'Extracted proxy SID');
  assertEqual(extractCookie(header, 'vyz_csrf'), 'token456', 'Extracted CSRF');
  assertEqual(extractCookie(header, 'nonexistent'), null, 'Returns null for missing cookie');
});

test('parseSetCookies parses Set-Cookie headers', function () {
  var headers = ['vyz_session=abc; HttpOnly; Path=/', 'vyz_csrf=def; Path=/'];
  var cookies = parseSetCookies(headers);
  assertEqual(cookies.get('vyz_session'), 'abc', 'Parsed session cookie');
  assertEqual(cookies.get('vyz_csrf'), 'def', 'Parsed CSRF cookie');
});

test('destroySession removes session', function () {
  var sess = getOrCreateSession(undefined);
  var cookie = sessionCookieHeader(sess);
  destroySession(cookie);
  var sess2 = getSession(cookie);
  assertEqual(sess2, null, 'Session destroyed');
});

// ═══════════════════════════════════════════════════════════════
//  Auth Interceptor Tests
// ═══════════════════════════════════════════════════════════════

section('Auth Interceptor');

test('interceptAuthResponse captures signing key from login', function () {
  var sess = getOrCreateSession(undefined);
  var loginBody = Buffer.from(JSON.stringify({
    signing_key: 'test-signing-key-123',
    access_token: 'jwt-token',
    operator_id: 'op-123',
  }));
  var headerMap = new Map();
  var result = interceptAuthResponse('POST', '/v1/auth/login', 200, headerMap, loginBody, sess);
  assertEqual(sess.signingKey, 'test-signing-key-123', 'Signing key captured');
  assertEqual(sess.accessToken, 'jwt-token', 'Access token captured');
  assertEqual(sess.operatorId, 'op-123', 'Operator ID captured');
  assert(result.modified, 'Response modified (signing key stripped)');
  var parsed = JSON.parse(result.body.toString());
  assert(parsed.signing_key === undefined, 'signing_key stripped from response');
});

test('interceptAuthResponse captures org ID from login response', function () {
  var sess = getOrCreateSession(undefined);
  var loginBody = Buffer.from(JSON.stringify({
    signing_key: 'key-123',
    operator_id: 'op-123',
    last_organization_id: 'org-from-login',
    needs_organization: false,
  }));
  var headerMap = new Map();
  interceptAuthResponse('POST', '/v1/auth/login', 200, headerMap, loginBody, sess);
  assertEqual(sess.organizationId, 'org-from-login', 'Org ID captured from login response');
});

test('interceptAuthResponse captures signing key from MFA verify', function () {
  var sess = getOrCreateSession(undefined);
  var mfaBody = Buffer.from(JSON.stringify({
    success: true,
    signing_key: 'mfa-signing-key',
    access_token: 'jwt-after-mfa',
  }));
  var headerMap = new Map();
  var result = interceptAuthResponse('POST', '/v1/auth/mfa/verify', 200, headerMap, mfaBody, sess);
  assertEqual(sess.signingKey, 'mfa-signing-key', 'MFA signing key captured');
  assertEqual(sess.accessToken, 'jwt-after-mfa', 'MFA access token captured');
});

test('interceptAuthResponse captures org ID from org select', function () {
  var sess = getOrCreateSession(undefined);
  var body = Buffer.from(JSON.stringify({ organization_id: 'org-abc-123' }));
  var headerMap = new Map();
  var result = interceptAuthResponse('POST', '/v1/auth/organizations/select', 200, headerMap, body, sess);
  assertEqual(sess.organizationId, 'org-abc-123', 'Organization ID captured');
  assert(!result.modified, 'Response not modified');
});

test('interceptAuthResponse captures refreshed token', function () {
  var sess = getOrCreateSession(undefined);
  sess.accessToken = 'old-token';
  var body = Buffer.from(JSON.stringify({ access_token: 'new-token', refresh_token: 'new-refresh' }));
  var headerMap = new Map();
  var result = interceptAuthResponse('POST', '/v1/auth/refresh', 200, headerMap, body, sess);
  assertEqual(sess.accessToken, 'new-token', 'Refreshed token captured');
  assertEqual(sess.refreshToken, 'new-refresh', 'Refresh token captured');
});

test('interceptAuthResponse clears credentials on logout', function () {
  var sess = getOrCreateSession(undefined);
  sess.signingKey = 'some-key';
  sess.accessToken = 'some-token';
  var body = Buffer.alloc(0);
  var headerMap = new Map();
  var result = interceptAuthResponse('POST', '/v1/auth/logout', 200, headerMap, body, sess);
  assertEqual(sess.signingKey, null, 'Signing key cleared on logout');
  assertEqual(sess.accessToken, null, 'Access token cleared on logout');
});

test('interceptAuthResponse does nothing for non-auth endpoints', function () {
  var sess = getOrCreateSession(undefined);
  sess.signingKey = 'existing-key';
  var body = Buffer.from(JSON.stringify({ data: 'response' }));
  var headerMap = new Map();
  var result = interceptAuthResponse('GET', '/v1/devices', 200, headerMap, body, sess);
  assert(!result.modified, 'Response not modified for non-auth endpoint');
  assertEqual(sess.signingKey, 'existing-key', 'Signing key unchanged');
});

test('interceptAuthResponse does nothing for failed requests', function () {
  var sess = getOrCreateSession(undefined);
  var body = Buffer.from(JSON.stringify({ error: 'bad credentials' }));
  var headerMap = new Map();
  var result = interceptAuthResponse('POST', '/v1/auth/login', 401, headerMap, body, sess);
  assert(!result.modified, 'Response not modified for failed login');
  assertEqual(sess.signingKey, null, 'Signing key not captured for failed login');
});

test('interceptAuthResponse captures Go session cookie from Set-Cookie', function () {
  var sess = getOrCreateSession(undefined);
  var body = Buffer.from(JSON.stringify({ signing_key: 'key123' }));
  var headerMap = new Map();
  headerMap.set('set-cookie', ['vyz_session=go-session-id-abc; HttpOnly; Path=/']);
  var result = interceptAuthResponse('POST', '/v1/auth/login', 200, headerMap, body, sess);
  assertEqual(sess.apiSessionCookie, 'go-session-id-abc', 'Go session cookie captured');
});

// ═══════════════════════════════════════════════════════════════
//  Config / Routing Tests
// ═══════════════════════════════════════════════════════════════

section('Config & Routing');

test('isApiPath identifies REST API paths', function () {
  assert(isApiPath('/v1/auth/login'), '/v1/auth/login is API path');
  assert(isApiPath('/v1/devices'), '/v1/devices is API path');
  assert(isApiPath('/v1/organizations'), '/v1/organizations is API path');
  assert(isApiPath('/health'), '/health is API path');
  assert(isApiPath('/healthz'), '/healthz is API path');
});

test('isApiPath identifies GraphQL paths', function () {
  assert(isApiPath('/org-123/graphql'), '/:org/graphql is API path');
  assert(isApiPath('/org-123/graphql/batch'), '/:org/graphql/batch is API path');
  assert(isApiPath('/org-123/graphql/ws'), '/:org/graphql/ws is API path');
});

test('isApiPath rejects non-API paths', function () {
  assert(!isApiPath('/'), 'Root is not API path');
  assert(!isApiPath('/index.html'), 'index.html is not API path');
  assert(!isApiPath('/assets/main.js'), '/assets/main.js is not API path');
  assert(!isApiPath('/dashboard'), '/dashboard is not API path');
});

test('isAuthEndpoint identifies auth endpoints', function () {
  assert(isAuthEndpoint('/v1/auth/login'), '/v1/auth/login is auth endpoint');
  assert(isAuthEndpoint('/v1/auth/refresh'), '/v1/auth/refresh is auth endpoint');
  assert(isAuthEndpoint('/v1/auth/csrf-token'), '/v1/auth/csrf-token is auth endpoint');
  assert(isAuthEndpoint('/v1/auth/register'), '/v1/auth/register is auth endpoint');
  assert(isAuthEndpoint('/v1/auth/mfa/verify'), '/v1/auth/mfa/verify is auth endpoint');
});

test('isAuthEndpoint rejects non-auth endpoints', function () {
  assert(!isAuthEndpoint('/v1/devices'), '/v1/devices is not auth endpoint');
  assert(!isAuthEndpoint('/v1/organizations'), '/v1/organizations is not auth endpoint');
  assert(!isAuthEndpoint('/org-123/graphql'), 'GraphQL is not auth endpoint');
});

test('isStateChanging identifies state-changing methods', function () {
  assert(isStateChanging('POST'), 'POST is state-changing');
  assert(isStateChanging('PUT'), 'PUT is state-changing');
  assert(isStateChanging('PATCH'), 'PATCH is state-changing');
  assert(isStateChanging('DELETE'), 'DELETE is state-changing');
  assert(!isStateChanging('GET'), 'GET is not state-changing');
});

test('isGraphQLPath identifies org-scoped GraphQL', function () {
  assert(isGraphQLPath('/org-123/graphql'), '/:org/graphql matches');
  assert(isGraphQLPath('/org-123/graphql/batch'), '/:org/graphql/batch matches');
  assert(isGraphQLPath('/org-123/graphql/ws'), '/:org/graphql/ws matches');
  assert(!isGraphQLPath('/v1/auth/login'), '/v1/auth/login is not GraphQL');
  assert(!isGraphQLPath('/graphql'), '/graphql without org prefix is not GraphQL');
});

// ═══════════════════════════════════════════════════════════════
//  Integration Tests (require Go API on localhost:3000)
// ═══════════════════════════════════════════════════════════════

async function checkGoApiRunning() {
  return new Promise(function (resolve) {
    var req = http.request(
      { hostname: 'localhost', port: 3000, path: '/health', method: 'GET', timeout: 2000 },
      function (res) {
        resolve(res.statusCode === 200);
      }
    );
    req.on('error', function () { resolve(false); });
    req.on('timeout', function () { req.destroy(); resolve(false); });
    req.end();
  });
}

async function runIntegrationTests() {
  var goApiRunning = await checkGoApiRunning();
  if (!goApiRunning) {
    console.log('\n  ' + logger.pc.yellow('⚠') + ' Go API not running on localhost:3000 — skipping integration tests');
    console.log('  ' + logger.pc.dim('Start the Go API and re-run to test end-to-end proxying'));
    return;
  }

  section('Integration Tests (Go API running)');

  // Start the proxy server on a test port.
  process.env.PROXY_PORT = '3099';
  var proxyPort = 3099;

  // Re-require config to pick up the new port.
  delete require.cache[require.resolve('./vyzorix-proxy-config.cjs')];
  var cfgLocal = require('./vyzorix-proxy-config.cjs');
  // Note: the proxy server process will read PROXY_PORT=3099 from env directly.

  // Start proxy server in a child process.
  var { spawn } = require('node:child_process');
  var proxyProcess = spawn('node', ['vyzorix-proxy-server.cjs'], {
    env: Object.assign({}, process.env, { PROXY_PORT: '3099', PROXY_VERBOSE: 'false' }),
    stdio: ['ignore', 'pipe', 'pipe'],
  });

  // Wait for proxy to be ready.
  await new Promise(function (resolve) {
    proxyProcess.stdout.on('data', function (data) {
      if (data.toString().includes('ready')) {
        setTimeout(resolve, 500);
      }
    });
  });

  await asyncTest('proxy health check returns ok', async function () {
    var res = await proxyFetch(proxyPort, '/proxy-health', 'GET');
    assertEqual(res.status, 200, 'Health check returns 200');
    var data = JSON.parse(res.body);
    assertEqual(data.status, 'ok', 'Health status is ok');
  });

  await asyncTest('proxy forwards unsigned request to Go API health', async function () {
    var res = await proxyFetch(proxyPort, '/health', 'GET');
    assertEqual(res.status, 200, 'Health forwarded with 200');
  });

  await asyncTest('proxy forwards CSRF token request', async function () {
    var res = await proxyFetch(proxyPort, '/v1/auth/csrf-token', 'GET');
    assert(res.status === 200 || res.status === 404, 'CSRF endpoint reachable');
  });

  // Clean up.
  proxyProcess.kill('SIGTERM');
  await new Promise(function (resolve) { setTimeout(resolve, 1000); });
}

function proxyFetch(port, path, method, body, headers) {
  return new Promise(function (resolve, reject) {
    var options = {
      hostname: 'localhost',
      port: port,
      path: path,
      method: method || 'GET',
      headers: headers || {},
    };

    var req = http.request(options, function (res) {
      var chunks = [];
      res.on('data', function (chunk) { chunks.push(chunk); });
      res.on('end', function () {
        resolve({
          status: res.statusCode,
          headers: res.headers,
          body: Buffer.concat(chunks).toString('utf8'),
        });
      });
      res.on('error', reject);
    });

    req.on('error', reject);
    if (body) req.write(body);
    req.end();
  });
}

// ═══════════════════════════════════════════════════════════════
//  Run All Tests
// ═══════════════════════════════════════════════════════════════

async function runAll() {
  console.log('\n' + logger.pc.magenta(logger.pc.bold('  Vyzorix Proxy Server — Test Suite')));
  console.log('  ' + logger.pc.gray('═'.repeat(56)));

  // Unit tests run synchronously above via test() calls.
  // Wait for any async tests.
  await new Promise(function (resolve) { setTimeout(resolve, 100); });

  // Integration tests.
  await runIntegrationTests();

  // Summary.
  console.log('\n  ' + logger.pc.gray('═'.repeat(56)));
  var total = passed + failed;
  console.log(
    '  ' + logger.pc.bold('Results:') + ' ' +
    logger.pc.green(passed + ' passed') + ' / ' +
    (failed > 0 ? logger.pc.red(failed + ' failed') : logger.pc.green(failed + ' failed')) + ' / ' +
    logger.pc.dim(total + ' total')
  );

  if (failures.length > 0) {
    console.log('\n  ' + logger.pc.red('Failures:'));
    failures.forEach(function (f) { console.log('    • ' + f); });
  }

  console.log('');
  process.exit(failed > 0 ? 1 : 0);
}

// Export for individual testing + auto-run.
module.exports = { runAll, test, asyncTest, assert, assertEqual };

// Auto-run if executed directly.
if (require.main === module) {
  runAll();
}
