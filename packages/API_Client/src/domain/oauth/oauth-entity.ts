export type OAuthProvider = "google" | "github";


export const OAUTH_ERRORS = {
    EMAIL_REQUIRED: "email_required",
    EMAIL_NOT_VERIFIED: "email_not_verified",
    LOGIN_FAILED: "login_failed",
    STATE_INVALID: "state_invalid",
    EXCHANGE_FAILED: "token_exchange_failed",
    CONFIG_MISSING: "oauth_not_configured",
} as const;

export type OAuthErrorCode = (typeof OAUTH_ERRORS)[keyof typeof OAUTH_ERRORS];

export interface OAuthCallbackResult {
    success: boolean;
    isNewUser: boolean;
    provider?: OAuthProvider;
    error?: OAuthErrorCode;
    errorMessage?: string;
    helpUrl?: string;
    retryable: boolean;
}

export interface OAuthError {
    code: OAuthErrorCode;
    message: string;
    provider?: OAuthProvider;
    helpUrl?: string;
    retryable: boolean;
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


export const OAUTH_EMAIL_HELP: Record<OAuthProvider, string> = {
    google: "https:
    github: "https:
};
