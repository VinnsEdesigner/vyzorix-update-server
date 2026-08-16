/**
 * Browser-compatible HTTP request signing using the Web Crypto API.
 *
 * This module mirrors the canonical format and header names of the Node-only
 * `vyzorServer/crypto` `signHttpRequest` function but uses `crypto.subtle`
 * instead of `node:crypto` so it is safe to import in a browser bundle.
 *
 * Canonical string (matches server `verifier.go`):
 *   {METHOD}\n{PATH}\n{NONCE}\n{TIMESTAMP_MS}\n{BODY}
 *
 * The HMAC is SHA-512, output as base64 — identical to what the Go
 * `infrastructure/crypto.Verifier.Verify` expects.
 */

const SIGNING_ALGO = 'HMAC';
const HASH_ALGO = 'SHA-512';

export interface SignHeaders {
  'X-Vyzorix-Timestamp': string;
  'X-Vyzorix-Nonce': string;
  'X-Vyzorix-Signature': string;
}

/**
 * Generate a cryptographically random nonce (32 hex chars = 16 bytes).
 * Uses `crypto.getRandomValues` (available in both browser and Node 18+).
 */
export function generateNonce(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
}

/**
 * Convert an ArrayBuffer to a base64 string.
 */
function arrayBufferToBase64(buf: ArrayBuffer): string {
  const bytes = new Uint8Array(buf);
  let binary = '';
  for (let i = 0; i < bytes.length; i++) {
    const byte = bytes[i];
    binary += String.fromCharCode(byte ?? 0);
  }
  return btoa(binary);
}

/**
 * Encode a string as UTF-8 bytes.
 */
function encodeUtf8(text: string): Uint8Array {
  const encoded = new TextEncoder().encode(text);
  // Copy into a fresh ArrayBuffer-backed view so crypto.subtle APIs (which
  // require BufferSource<ArrayBuffer>, not SharedArrayBuffer-backed views)
  // accept it under TS lib DOM's strict ArrayBufferLike checks.
  const out = new Uint8Array(encoded.byteLength);
  out.set(encoded);
  return out;
}

/**
 * Import a raw secret string as a Web Crypto HMAC key.
 * The key material is the UTF-8 bytes of the secret — matching the Go side
 * which uses `[]byte(secret)` as the HMAC key directly.
 */
async function importHmacKey(secret: string): Promise<CryptoKey> {
  const keyBytes = encodeUtf8(secret);
  return crypto.subtle.importKey(
    'raw',
    keyBytes.buffer.slice(keyBytes.byteOffset, keyBytes.byteOffset + keyBytes.byteLength) as ArrayBuffer,
    { name: SIGNING_ALGO, hash: HASH_ALGO },
    false,
    ['sign'],
  );
}

/**
 * Compute the HMAC-SHA512 signature for an HTTP request.
 *
 * @param method    HTTP method (e.g. "GET", "POST")
 * @param path      Request path including query string (e.g. "/v1/devices?limit=10")
 * @param timestampMs  Unix timestamp in milliseconds (as a string)
 * @param nonce     Random nonce
 * @param body      Request body as a string (empty string for GET/DELETE without body)
 * @param secret    The signing secret (UTF-8 bytes used as HMAC key)
 * @returns Base64-encoded HMAC-SHA512 signature
 */
export async function computeHttpSignature(
  method: string,
  path: string,
  timestampMs: string,
  nonce: string,
  body: string,
  secret: string,
): Promise<string> {
  const canonical = [method.toUpperCase(), path, nonce, timestampMs].join('\n') + '\n';

  const key = await importHmacKey(secret);
  const data = new Uint8Array(encodeUtf8(canonical).length + encodeUtf8(body).length);
  data.set(encodeUtf8(canonical), 0);
  data.set(encodeUtf8(body), encodeUtf8(canonical).length);

  const signature = await crypto.subtle.sign(
    SIGNING_ALGO,
    key,
    data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength) as ArrayBuffer,
  );
  return arrayBufferToBase64(signature);
}

/**
 * Sign an HTTP request and produce the three `X-Vyzorix-*` headers.
 *
 * @param method  HTTP method
 * @param path    Request path (URL pathname + search, without origin/baseURL)
 * @param body    Serialized request body (empty string for bodyless requests)
 * @param secret  The signing secret
 * @returns The three signing headers to set on the request
 */
export async function signHttpRequestBrowser(
  method: string,
  path: string,
  body: string,
  secret: string,
): Promise<SignHeaders> {
  const timestampMs = Date.now().toString();
  const nonce = generateNonce();

  const signature = await computeHttpSignature(
    method,
    path,
    timestampMs,
    nonce,
    body,
    secret,
  );

  return {
    'X-Vyzorix-Timestamp': timestampMs,
    'X-Vyzorix-Nonce': nonce,
    'X-Vyzorix-Signature': signature,
  };
}

/**
 * Constant-time string comparison to prevent timing attacks.
 */
export function constantTimeCompare(a: string, b: string): boolean {
  if (a.length !== b.length) {
    return false;
  }
  let result = 0;
  for (let i = 0; i < a.length; i++) {
    result |= a.charCodeAt(i) ^ b.charCodeAt(i);
  }
  return result === 0;
}
