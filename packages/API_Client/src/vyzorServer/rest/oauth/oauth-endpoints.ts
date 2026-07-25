import { buildOAuthState, parseOAuthCallbackUrl } from "@/domain/oauth";

const PATHS = {
  google: {
    login: "/v1/auth/google",
    callback: "/v1/auth/google/callback",
  },
  github: {
    login: "/v1/auth/github",
    callback: "/v1/auth/github/callback",
  },
} as const;

export interface OAuthLoginOptions {
  state?: string;
  redirectAfterCallback?: string;
}

export const oauth = {
  getGoogleLoginUrl(options?: OAuthLoginOptions): string {
    const state = options?.state || buildOAuthState();
    const params = new URLSearchParams();
    params.set("state", state);
    if (options?.redirectAfterCallback) {
      params.set("redirect_uri", options.redirectAfterCallback);
    }
    return `${PATHS.google.login}?${params.toString()}`;
  },

  getGitHubLoginUrl(options?: OAuthLoginOptions): string {
    const state = options?.state || buildOAuthState();
    const params = new URLSearchParams();
    params.set("state", state);
    if (options?.redirectAfterCallback) {
      params.set("redirect_uri", options.redirectAfterCallback);
    }
    return `${PATHS.github.login}?${params.toString()}`;
  },

  parseCallback(url: string) {
    return parseOAuthCallbackUrl(url);
  },

  initiateGoogleLogin(options?: OAuthLoginOptions): void {
    window.location.href = this.getGoogleLoginUrl(options);
  },

  initiateGitHubLogin(options?: OAuthLoginOptions): void {
    window.location.href = this.getGitHubLoginUrl(options);
  },
};
