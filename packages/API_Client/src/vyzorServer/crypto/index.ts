

import { createHmac, randomBytes, createCipheriv, createDecipheriv, createHash } from 'crypto';

const AES256_KEY_SIZE = 32;
const NONCE_SIZE = 12;



export function deriveKey(secret: string): Buffer {
  const hash = createHash('sha512');
  hash.update(secret);
  return hash.digest().slice(0, AES256_KEY_SIZE);
}



export function encryptAES256GCM(secret: string, plaintext: Buffer | string): string {
  const key = deriveKey(secret);
  const nonce = randomBytes(NONCE_SIZE);
  const plaintextBytes = typeof plaintext === 'string' ? Buffer.from(plaintext) : plaintext;

  const cipher = createCipheriv('aes-256-gcm', key, nonce);
  const encrypted = Buffer.concat([cipher.update(plaintextBytes), cipher.final()]);
  const authTag = cipher.getAuthTag();

    const result = Buffer.concat([nonce, encrypted, authTag]);
  return result.toString('base64');
}


export function decryptAES256GCM(secret: string, encryptedBase64: string): Buffer {
  const key = deriveKey(secret);
  const encrypted = Buffer.from(encryptedBase64, 'base64');

  if (encrypted.length < NONCE_SIZE + 16) {     throw new Error('Ciphertext too short');
  }

  const nonce = encrypted.slice(0, NONCE_SIZE);
  const authTag = encrypted.slice(encrypted.length - 16);
  const ciphertext = encrypted.slice(NONCE_SIZE, encrypted.length - 16);

  const decipher = createDecipheriv('aes-256-gcm', key, nonce);
  decipher.setAuthTag(authTag);

  return Buffer.concat([decipher.update(ciphertext), decipher.final()]);
}



export function sha512(data: Buffer | string): Buffer {
  const hash = createHash('sha512');
  hash.update(data);
  return hash.digest();
}


export function computeHttpSignature(
  method: string,
  path: string,
  timestampMs: string,
  nonce: string,
  body: Buffer | string,
  secret: string
): string {
  const bodyBytes = typeof body === 'string' ? Buffer.from(body) : body;

  // Canonical format per server: {method}\n{path}\n{nonce}\n{timestamp}\n{body}
  // Note: No empty line between headers and body
  const canonical = [
    method.toUpperCase(),
    path,
    nonce,
    timestampMs,
  ].join('\n') + '\n';

  const hmac = createHmac('sha512', secret);
  hmac.update(canonical, 'utf8');
  hmac.update(bodyBytes);
  
  return hmac.digest('base64');
}


export function signHttpRequest(
  method: string,
  path: string,
  body: Buffer | string,
  deviceId: string,
  secret: string
): {
  'X-Vyzorix-Timestamp': string;
  'X-Vyzorix-Nonce': string;
  'X-Vyzorix-Signature': string;
} {
  const timestampMs = Date.now().toString();
  const nonce = generateNonce();

  const signature = computeHttpSignature(method, path, timestampMs, nonce, body, secret);

  return {
    'X-Vyzorix-Timestamp': timestampMs,
    'X-Vyzorix-Nonce': nonce,
    'X-Vyzorix-Signature': signature,
  };
}


export function verifyHttpSignature(
  method: string,
  path: string,
  timestampMs: string,
  nonce: string,
  body: Buffer | string,
  signature: string,
  secret: string
): boolean {
  const expected = computeHttpSignature(method, path, timestampMs, nonce, body, secret);
  return constantTimeCompare(expected, signature);
}



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



export interface CommandFrame {
  dispatchId: string;
  command: string;
  args?: string;
  timestamp: number;
  nonce?: string;
  signature?: string;
}


export function signCommand(
  frame: CommandFrame,
  deviceId: string,
  secret: string
): { nonce: string; signature: string } {
    const nonce = randomBytes(16).toString('hex');

    const argsStr = frame.args || '{}';
  const canonical = [
    frame.dispatchId,
    deviceId,
    frame.command,
    frame.timestamp.toString(),
    nonce,
    argsStr,
  ].join('|');

    const hmac = createHmac('sha512', secret);
  hmac.update(canonical);
  const signature = hmac.digest('hex');

  return { nonce, signature };
}


export function buildCanonicalString(
  dispatchId: string,
  deviceId: string,
  command: string,
  timestamp: number,
  nonce: string,
  args: string
): string {
  return [dispatchId, deviceId, command, timestamp.toString(), nonce, args || '{}'].join('|');
}


export function validateCommandHMAC(
  frame: CommandFrame,
  deviceId: string,
  secret: string
): boolean {
  if (!frame.nonce || !frame.signature) {
    return false;
  }

  const argsStr = frame.args || '{}';
  const canonical = [
    frame.dispatchId,
    deviceId,
    frame.command,
    frame.timestamp.toString(),
    frame.nonce,
    argsStr,
  ].join('|');

  const hmac = createHmac('sha512', secret);
  hmac.update(canonical);
  const expected = hmac.digest('hex');

  return constantTimeCompare(expected, frame.signature);
}


export function validateTimestamp(frame: CommandFrame, maxDriftMs: number = 30000): boolean {
  const nowMs = Date.now();
  const drift = Math.abs(nowMs - frame.timestamp);
  return drift <= maxDriftMs;
}


export function signWebSocketConnect(
  deviceId: string,
  timestamp: string,
  nonce: string,
  secret: string
): string {
  const canonical = `CONNECT:${deviceId}:${timestamp}:${nonce}`;
  const hmac = createHmac('sha512', secret);
  hmac.update(canonical);
  return hmac.digest('hex');
}


export function generateNonce(): string {
  return randomBytes(16).toString('hex');
}


export function generateTimestamp(): string {
  return Math.floor(Date.now() / 1000).toString();
}


export function generateTimestampMs(): number {
  return Date.now();
}
