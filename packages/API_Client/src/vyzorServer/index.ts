


// Universal (browser + Node) surface. Node-only subsystems that depend on the
// `node:crypto` built-in (or the browser `WebSocket` global with a Node polyfill)
// are intentionally NOT re-exported here to keep this entry bundle-safe for both
// runtimes. Import them from `@vyzorix/api-client/node` instead.
//
// Exception: `crypto/browser-sign` uses the Web Crypto API (`crypto.subtle`),
// not `node:crypto`, so it is browser-safe and is re-exported here for the
// browser-side request-signing interceptor.
export * from './config';
export * from './rest';
export * from './graphql';
export {
  signHttpRequestBrowser,
  computeHttpSignature,
  generateNonce as generateSigningNonce,
  constantTimeCompare as constantTimeCompareStrings,
  type SignHeaders,
} from './crypto/browser-sign';
export {
  authContext,
  getCurrentOrganizationId,
  isAuthenticated,
  isAccountLocked,
  isTokenExpired,
  getTimeUntilExpiry,
  type AuthState,
  type LockoutState,
} from './auth';
export * from './client';
