/**
 * vyzorix-proxy-signing.cjs
 *
 * HTTP request signing using the same canonical format as the Go API server's
 * Verifier.Verify (infrastructure/crypto/verifier.go:210-213):
 *
 *   canonical = METHOD + "\n" + PATH + "\n" + NONCE + "\n" + TIMESTAMP + "\n" + BODY
 *   signature = base64(HMAC-SHA512(signingKey, canonical))
 *
 * MUST stay in sync with:
 *   - Go:        internal/infrastructure/crypto/verifier.go (Verify method)
 *   - API Client: packages/API_Client/src/vyzorServer/crypto/index.ts (computeHttpSignature)
 */

'use strict';

const crypto = require('node:crypto');

/**
 * Compute the HMAC-SHA512 signature for an HTTP request.
 * @param {string} method   - HTTP method (GET, POST, etc.)
 * @param {string} path     - Request path + query string (what r.URL.RequestURI() returns)
 * @param {string} nonce    - Random nonce
 * @param {string} timestamp - Unix milliseconds string
 * @param {Buffer} body     - Request body bytes
 * @param {string} signingKey - Per-session HMAC secret
 * @returns {string} Base64-encoded signature
 */
function computeSignature(method, path, nonce, timestamp, body, signingKey) {
  var canonical = method.toUpperCase() + '\n' + path + '\n' + nonce + '\n' + timestamp + '\n';
  var hmac = crypto.createHmac('sha512', signingKey);
  hmac.update(canonical, 'utf8');
  hmac.update(body);
  return hmac.digest('base64');
}

/**
 * Generate the three X-Vyzorix-* signing headers for an outgoing API request.
 * @returns {{'X-Vyzorix-Timestamp': string, 'X-Vyzorix-Nonce': string, 'X-Vyzorix-Signature': string}}
 */
function signRequest(method, path, body, signingKey) {
  var timestamp = Date.now().toString();
  var nonce = crypto.randomBytes(16).toString('hex');
  var signature = computeSignature(method, path, nonce, timestamp, body, signingKey);
  return {
    'X-Vyzorix-Timestamp': timestamp,
    'X-Vyzorix-Nonce': nonce,
    'X-Vyzorix-Signature': signature,
  };
}

/**
 * Derive the API key signing secret from a full API key string.
 * Mirrors the Go server's deriveAPIKeySigningSecret (api_key_service.go:412).
 *
 * API key format: vxyz_{secret}  (prefix + hex-encoded random bytes)
 * Signing secret = hex(SHA-512(fullKey))
 *
 * The server stores this derived secret at key-creation time and uses it to
 * verify HMAC request signatures. The client/proxy derives the same value from
 * the full key it holds. MUST stay in sync with the Go implementation.
 */
function deriveAPIKeySigningSecret(fullKey) {
  var hash = crypto.createHash('sha512');
  hash.update(fullKey);
  return hash.digest('hex');
}

module.exports = { computeSignature, signRequest, deriveAPIKeySigningSecret };
