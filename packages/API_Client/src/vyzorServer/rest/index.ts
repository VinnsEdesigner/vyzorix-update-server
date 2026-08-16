// Re-export shared utilities
export * from "./_shared";

// Auth core exports — login, register, logout, /me, token refresh.
// MFA endpoints live in ./mfa, password-reset in ./password, email verification in ./email.
export {
  fetchCSRFToken,
  login,
  loginWithTokens,
  register,
  logout,
  getMe,
  updateName,
  refreshToken,
} from "./auth/rest-auth-endpoints";

export type {
  LoginResult,
  LoginWithTokensResult,
} from "./auth/rest-auth-endpoints";

export * from "./registration";

export { devices } from "./device";
export type { DeviceParams, DeviceSettings, DeviceSettingsUpdateRequest, DeviceThresholdUpdateRequest, ConnectionStatus } from "./device";

export { settings } from "./settings";

// Diagnostics exports
export * from "./diagnostics";
export * from "./apikey";
export * from "./commands";
export * from "./updates";
export * from "./logs";
export * from "./session";
export * from "./admin";
export * from "./admin-clients";
export * from "./oauth";

// Organization - export all except settings (which conflicts with ./settings)
export { organizations } from "./organization";
export { members } from "./organization";
export { invitations } from "./organization";
export { settings as orgSettings, type OrganizationSettings, type ThresholdUpdateRequest, type SettingsUpdateRequest } from "./organization";

export * from "./invitation";
export * from "./me";
export * from "./clientcredentials";
export * from "./connections";
export * from "./telemetry";
export * from "./mfa";
export * from "./password";
export * from "./email";
export * from "./events";
export * from "./health";
export * from "./metrics";

// WebSocket, device-signing, and crypto subsystems depend on Node-only
// (`node:crypto`) or environment-specific globals. They are exported from the
// `@vyzorix/api-client/node` entry to keep the universal REST surface
// bundle-safe for browsers.
