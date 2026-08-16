/**
 * Node entry point.
 *
 * Re-exports the universal surface (`./index`) plus the Node-only subsystems
 * that depend on the `node:crypto` built-in or assume a Node runtime:
 *   - `vyzorServer/crypto`       (HMAC, AES-256-GCM, HTTP signing via node:crypto)
 *   - `vyzorServer/security`     (SSL pinning via node:crypto / tls)
 *   - `vyzorServer/websocket`    (WebSocket client; pairs with node:crypto signing)
 *   - `vyzorServer/device`       (device client; signs requests via node:crypto)
 *
 * Browser bundles should import from the package root (`@vyzorix/api-client`)
 * instead, which omits these modules so `node:crypto`/`Buffer` never enters the
 * browser graph.
 */
export * from "./index";
export {
  deriveKey,
  encryptAES256GCM,
  decryptAES256GCM,
  sha512,
  computeHttpSignature,
  signHttpRequest,
  verifyHttpSignature,
  constantTimeCompare,
  type CommandFrame,
  signCommand,
  buildCanonicalString,
  validateCommandHMAC,
  validateTimestamp,
  signWebSocketConnect,
  generateNonce,
  generateTimestamp,
  generateTimestampMs,
} from "./vyzorServer/crypto";
export * from "./vyzorServer/security";
export * from "./vyzorServer/websocket";
export * from "./vyzorServer/device";
