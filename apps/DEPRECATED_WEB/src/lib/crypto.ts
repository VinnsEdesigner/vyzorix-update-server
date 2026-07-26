// Cryptographic utilities for request signing and response encryption.
// Implements AES-256-GCM encryption and HMAC-SHA512 signing per PRD spec.

/**
 * Derives a 32-byte key from a secret using SHA-512.
 * Used to derive AES-256 key from client_secret.
 */
export const deriveKey = async (secret: string): Promise<Uint8Array> => {
  const encoder = new TextEncoder();
  const data = encoder.encode(secret);
  const hash = await sha512(data);
  return hash.slice(0, 32);
};

/**
 * SHA-512 hash function using Web Crypto API.
 */
export const sha512 = async (data: BufferSource): Promise<Uint8Array> => {
  const hashBuffer = await crypto.subtle.digest("SHA-512", data);
  return new Uint8Array(hashBuffer);
};

/**
 * SHA-512 hash that returns hex string.
 */
export const sha512Hex = async (data: string): Promise<string> => {
  const encoder = new TextEncoder();
  const hash = await sha512(encoder.encode(data));
  return Array.from(hash)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
};

/**
 * HMAC-SHA512 using Web Crypto API.
 */
export const hmacSha512 = async (key: string, message: string): Promise<Uint8Array> => {
  const encoder = new TextEncoder();
  const keyData = encoder.encode(key);

  const cryptoKey = await crypto.subtle.importKey(
    "raw",
    keyData,
    { name: "HMAC", hash: "SHA-512" },
    false,
    ["sign"],
  );

  const signature = await crypto.subtle.sign("HMAC", cryptoKey, encoder.encode(message));
  return new Uint8Array(signature);
};

/**
 * HMAC-SHA512 that returns hex string.
 */
export const hmacSha512Hex = async (key: string, message: string): Promise<string> => {
  const sig = await hmacSha512(key, message);
  return Array.from(sig)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
};

/**
 * Generate cryptographically secure random bytes.
 */
export const randomBytes = (length: number): Uint8Array => {
  return crypto.getRandomValues(new Uint8Array(length));
};

/**
 * Generate a random nonce for AES-GCM (12 bytes).
 */
export const generateNonce = (): Uint8Array => {
  return randomBytes(12);
};

/**
 * AES-256-GCM encryption.
 * Returns nonce || ciphertext (both base64 encoded).
 */
export const aes256GcmEncrypt = async (
  secret: string,
  plaintext: string | object,
): Promise<{ nonce: string; ciphertext: string }> => {
  const key = await deriveKey(secret);
  const nonce = generateNonce();

  // Handle object or string input
  const data = typeof plaintext === "string" ? plaintext : JSON.stringify(plaintext);

  const encoder = new TextEncoder();
  const plaintextBytes = encoder.encode(data);

  const cryptoKey = await crypto.subtle.importKey("raw", key as BufferSource, "AES-GCM", false, [
    "encrypt",
  ]);

  const ciphertext = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: nonce as BufferSource },
    cryptoKey,
    plaintextBytes,
  );

  return {
    nonce: b64Encode(nonce),
    ciphertext: b64Encode(new Uint8Array(ciphertext)),
  };
};

/**
 * AES-256-GCM decryption.
 * Expects nonce and ciphertext as base64 strings.
 */
export const aes256GcmDecrypt = async (
  secret: string,
  nonceB64: string,
  ciphertextB64: string,
): Promise<string> => {
  const key = await deriveKey(secret);
  const nonce = b64Decode(nonceB64);
  const ciphertext = b64Decode(ciphertextB64);

  const cryptoKey = await crypto.subtle.importKey("raw", key as BufferSource, "AES-GCM", false, [
    "decrypt",
  ]);

  const plaintext = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: nonce as BufferSource },
    cryptoKey,
    ciphertext as BufferSource,
  );

  const decoder = new TextDecoder();
  return decoder.decode(plaintext);
};

/**
 * AES-256-GCM decryption from combined ciphertext (nonce prepended).
 */
export const aes256GcmDecryptCombined = async (
  secret: string,
  combinedB64: string,
): Promise<string> => {
  const combined = b64Decode(combinedB64);
  const nonce = combined.slice(0, 12);
  const ciphertext = combined.slice(12);

  const key = await deriveKey(secret);

  const cryptoKey = await crypto.subtle.importKey("raw", key as BufferSource, "AES-GCM", false, [
    "decrypt",
  ]);

  const plaintext = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: nonce as BufferSource },
    cryptoKey,
    ciphertext as BufferSource,
  );

  const decoder = new TextDecoder();
  return decoder.decode(plaintext);
};

// Base64 utilities using native btoa/atob
const b64Encode = (data: Uint8Array): string => {
  const str = String.fromCharCode(...data);
  return btoa(str);
};

const b64Decode = (b64: string): Uint8Array => {
  const str = atob(b64);
  const bytes = new Uint8Array(str.length);
  for (let i = 0; i < str.length; i++) {
    bytes[i] = str.charCodeAt(i);
  }
  return bytes;
};

/**
 * Decode base64 string to text.
 */
export const b64ToString = (b64: string): string => {
  return atob(b64);
};

/**
 * Encode string to base64.
 */
export const stringToB64 = (str: string): string => {
  return btoa(str);
};
