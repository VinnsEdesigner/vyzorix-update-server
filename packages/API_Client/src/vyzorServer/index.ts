


// Universal (browser + Node) surface. Node-only subsystems that depend on the
// `node:crypto` built-in (or the browser `WebSocket` global with a Node polyfill)
// are intentionally NOT re-exported here to keep this entry bundle-safe for both
// runtimes. Import them from `@vyzorix/api-client/node` instead.
export * from './config';
export * from './rest';
export * from './graphql';
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
