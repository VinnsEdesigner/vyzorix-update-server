// Re-export shared utilities
export * from "./_shared";

// Auth exports - but NOT MFA functions (those come from ./mfa)
export {
  fetchCSRFToken,
  login,
  loginWithTokens,
  register,
  forgotPassword,
  resetPassword,
  resendPasswordReset,
  verifyEmail,
  resendVerification,
  logout,
  getMe,
  updateName,
} from "./auth/rest-auth-endpoints";

export type {
  LoginResult,
  LoginWithTokensResult,
} from "./auth/rest-auth-endpoints";

export * from "./registration";

// Settings module
export { settings } from "./settings";

// Diagnostics exports
export * from "./diagnostics";
export * from "./apikey";
export * from "./commands";
export * from "./updates";
export * from "./device";
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
export * from "../websocket";
