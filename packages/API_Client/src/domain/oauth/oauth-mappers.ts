import type { OAuthProvider, OAuthCallbackResult } from "./oauth-entity";

export function parseOAuthCallbackUrl(url: string): OAuthCallbackResult {
  try {
    const urlObj = new URL(url, "http://localhost");
    const oauth = urlObj.searchParams.get("oauth");
    const isNew = urlObj.searchParams.get("new");
    const provider = urlObj.searchParams.get("provider");

    if (oauth === "success") {
      return {
        success: true,
        isNewUser: isNew === "true",
        provider: provider as OAuthProvider || undefined,
      };
    }

    return {
      success: false,
      isNewUser: false,
      error: oauth || "unknown_error",
    };
  } catch {
    return {
      success: false,
      isNewUser: false,
      error: "invalid_url",
    };
  }
}

export function buildOAuthState(): string {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  return Array.from(array, (b) => b.toString(16).padStart(2, "0")).join("");
}
