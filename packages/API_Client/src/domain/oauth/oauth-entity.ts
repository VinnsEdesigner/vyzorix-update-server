export type OAuthProvider = "google" | "github";

export interface OAuthCallbackResult {
  success: boolean;
  isNewUser: boolean;
  provider?: OAuthProvider;
  error?: string;
}

export interface OAuthRedirectUrlOptions {
  state?: string;
  redirectUrl?: string;
}

export interface OAuthEndpoints {
  login: string;
  callback: string;
}

export const OAUTH_ENDPOINTS: Record<OAuthProvider, OAuthEndpoints> = {
  google: {
    login: "/v1/auth/google",
    callback: "/v1/auth/google/callback",
  },
  github: {
    login: "/v1/auth/github",
    callback: "/v1/auth/github/callback",
  },
};
