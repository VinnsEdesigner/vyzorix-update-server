

import { buildOAuthState, parseOAuthCallbackUrl, OAUTH_ENDPOINTS, type OAuthProvider } from "@/domain/oauth";

export const OAuthErrorCode = {
  EMAIL_REQUIRED: "email_required",
  EMAIL_NOT_VERIFIED: "email_not_verified",
  LOGIN_FAILED: "login_failed",
  STATE_INVALID: "state_invalid",
  TOKEN_EXCHANGE_FAILED: "token_exchange_failed",
  CONFIG_MISSING: "oauth_not_configured",
} as const;

export type OAuthErrorCode = typeof OAuthErrorCode[keyof typeof OAuthErrorCode];

export interface OAuthErrorDetails {
  code: OAuthErrorCode;
  message: string;
  provider?: string;
  helpUrl?: string;
  retryable: boolean;
}

export interface OAuthCallbackResult {
  success: boolean;
  isNew: boolean;
  provider: OAuthProvider;
  error?: OAuthErrorDetails;
}


export function parseOAuthCallback(url: string): OAuthCallbackResult {
  const parsed = parseOAuthCallbackUrl(url);

  if (!parsed.success && parsed.error) {
    return {
      success: false,
      isNew: false,
      provider: parsed.provider ?? "google",
      error: {
        code: (parsed.error as OAuthErrorCode) || OAuthErrorCode.LOGIN_FAILED,
        message: parsed.errorMessage || "OAuth authentication failed",
        provider: parsed.provider,
        retryable: parsed.retryable,
      },
    };
  }

  return {
    success: true,
    isNew: parsed.isNewUser || false,
    provider: parsed.provider ?? "google",
  };
}


export function getGoogleLoginUrl(options?: {
  state?: string;
  redirectAfterCallback?: string;
}): string {
  const state = options?.state || buildOAuthState();
  const params = new URLSearchParams();
  params.set("state", state);
  if (options?.redirectAfterCallback) {
    params.set("redirect_uri", options.redirectAfterCallback);
  }
  return `${OAUTH_ENDPOINTS.google.login}?${params.toString()}`;
}


export function getGitHubLoginUrl(options?: {
  state?: string;
  redirectAfterCallback?: string;
}): string {
  const state = options?.state || buildOAuthState();
  const params = new URLSearchParams();
  params.set("state", state);
  if (options?.redirectAfterCallback) {
    params.set("redirect_uri", options.redirectAfterCallback);
  }
  return `${OAUTH_ENDPOINTS.github.login}?${params.toString()}`;
}


export function initiateGoogleLogin(options?: {
  state?: string;
  redirectAfterCallback?: string;
}): void {
  window.location.href = getGoogleLoginUrl(options);
}


export function initiateGitHubLogin(options?: {
  state?: string;
  redirectAfterCallback?: string;
}): void {
  window.location.href = getGitHubLoginUrl(options);
}
