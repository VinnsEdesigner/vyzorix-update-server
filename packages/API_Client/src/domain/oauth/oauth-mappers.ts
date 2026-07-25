import { OAUTH_ERRORS, type OAuthProvider, type OAuthCallbackResult, type OAuthErrorCode } from "./oauth-entity";

export function parseOAuthCallbackUrl(url: string): OAuthCallbackResult {
    try {
        const urlObj = new URL(url, "http://localhost");
        const oauth = urlObj.searchParams.get("oauth");
        const isNew = urlObj.searchParams.get("new");
        const provider = urlObj.searchParams.get("provider") as OAuthProvider | null;
        const code = urlObj.searchParams.get("code") as OAuthErrorCode | null;
        const message = urlObj.searchParams.get("message");
        const retryable = urlObj.searchParams.get("retryable") !== "false";

        if (oauth === "success") {
            return {
                success: true,
                isNewUser: isNew === "true",
                provider: provider || undefined,
                retryable: true,
            };
        }

        if (oauth === "error" && code) {
            return {
                success: false,
                isNewUser: false,
                error: code,
                errorMessage: message || getDefaultErrorMessage(code),
                provider: provider || undefined,
                helpUrl: getHelpUrl(code, provider || undefined),
                retryable,
            };
        }

        return {
            success: false,
            isNewUser: false,
            error: OAUTH_ERRORS.LOGIN_FAILED,
            errorMessage: "An unknown error occurred during sign in.",
            retryable: true,
        };
    } catch {
        return {
            success: false,
            isNewUser: false,
            error: OAUTH_ERRORS.LOGIN_FAILED,
            errorMessage: "Failed to parse authentication response.",
            retryable: true,
        };
    }
}

function getDefaultErrorMessage(code: OAuthErrorCode): string {
    switch (code) {
        case OAUTH_ERRORS.EMAIL_REQUIRED:
            return "Your account must have an email address to sign up. Please add a verified email to your account and try again.";
        case OAUTH_ERRORS.EMAIL_NOT_VERIFIED:
            return "Your email must be verified to sign up. Please verify your email and try again.";
        case OAUTH_ERRORS.STATE_INVALID:
            return "The sign-in session expired. Please try again.";
        case OAUTH_ERRORS.EXCHANGE_FAILED:
            return "Failed to complete the sign-in. Please try again.";
        case OAUTH_ERRORS.CONFIG_MISSING:
            return "This sign-in method is not configured. Contact support.";
        default:
            return "Sign in failed. Please try again.";
    }
}

function getHelpUrl(code: OAuthErrorCode, provider?: OAuthProvider): string | undefined {
    if (!provider) return undefined;

    switch (code) {
        case OAUTH_ERRORS.EMAIL_REQUIRED:
        case OAUTH_ERRORS.EMAIL_NOT_VERIFIED:
            if (provider === "google") {
                return "https://support.google.com/accounts/answer/162744";
            }
            if (provider === "github") {
                return "https://docs.github.com/get-started/signing-up-for-github/signing-up-with-a-new-email-address";
            }
            break;
        default:
            return undefined;
    }
    return undefined;
}

export function buildOAuthState(): string {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return Array.from(array, (b) => b.toString(16).padStart(2, "0")).join("");
}


export function formatOAuthErrorForDisplay(result: OAuthCallbackResult): {
    title: string;
    description: string;
    actionLabel?: string;
    actionUrl?: string;
    showRetry?: boolean;
} {
    if (result.success) {
        return {
            title: "Success!",
            description: "You have successfully signed in.",
            showRetry: false,
        };
    }

    const code = result.error;
    const provider = result.provider;

    switch (code) {
        case OAUTH_ERRORS.EMAIL_REQUIRED:
            return {
                title: "Email Required",
                description: result.errorMessage || "Your account must have a verified email address to sign up.",
                actionLabel: provider === "github" ? "Update GitHub Email Settings" : "Update Google Account",
                actionUrl: result.helpUrl || (provider === "github" ? "https://github.com/settings/emails" : "https://myaccount.google.com/email"),
                showRetry: false,
            };

        case OAUTH_ERRORS.EMAIL_NOT_VERIFIED:
            return {
                title: "Email Not Verified",
                description: result.errorMessage || "Your email must be verified before signing up.",
                actionLabel: provider === "github" ? "Resend Verification Email" : "Verify Google Email",
                actionUrl: result.helpUrl || "https://support.google.com/accounts/answer/162744",
                showRetry: false,
            };

        case OAUTH_ERRORS.STATE_INVALID:
            return {
                title: "Session Expired",
                description: result.errorMessage || "The sign-in session expired. Please try again.",
                showRetry: true,
            };

        case OAUTH_ERRORS.EXCHANGE_FAILED:
            return {
                title: "Connection Error",
                description: result.errorMessage || "Failed to connect to the identity provider.",
                showRetry: true,
            };

        default:
            return {
                title: "Sign In Failed",
                description: result.errorMessage || "An unexpected error occurred.",
                showRetry: result.retryable,
            };
    }
}
