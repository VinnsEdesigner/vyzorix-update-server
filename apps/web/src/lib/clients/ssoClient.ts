/**
 * ssoClient.ts - Single Sign-On client for Google and GitHub OAuth.
 *
 * Handles OAuth initiation and callback processing.
 */

import { logger } from "@/lib/logger";

const API_BASE = "";

// Types
export interface OperatorResponse {
  id: string;
  email: string;
  name: string;
  role: "viewer" | "operator" | "super_admin";
  createdAt: number;
  emailVerified?: boolean;
}

export type SSOProvider = "Google" | "GitHub";

interface ErrorResponse {
  error: string;
  message: string;
}

const jsonOrThrow = async <T>(res: Response): Promise<T> => {
  const contentType = res.headers.get("content-type") ?? "";
  if (!res.ok) {
    let msg = `HTTP ${res.status}`;
    if (contentType.includes("application/json")) {
      const body = (await res.json()) as ErrorResponse;
      msg = body.message ?? body.error ?? msg;
    }
    throw new Error(msg);
  }
  if (contentType.includes("application/json")) {
    return (await res.json()) as T;
  }
  throw new Error(`Expected JSON, got ${contentType}`);
};

/**
 * Initiate SSO OAuth flow - redirects to provider's auth page.
 */
export function initiateSSO(provider: SSOProvider): void {
  const target = `/v1/auth/${provider.toLowerCase()}`;
  logger.info("sso", `-> Redirecting to ${provider} OAuth`, { target });
  window.location.href = target;
}

/**
 * Handle OAuth callback - called after redirect from provider.
 * This is typically called on the /auth/callback route.
 */
export async function handleSSOCallback(
  provider: SSOProvider,
  code: string,
  state: string,
): Promise<OperatorResponse> {
  logger.info("sso", "-> POST OAuth callback", { provider });
  const res = await fetch(
    `${API_BASE}/v1/auth/${provider.toLowerCase()}/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`,
    { credentials: "include" },
  );
  const out = await jsonOrThrow<OperatorResponse>(res);
  logger.info("sso", "<- OAuth callback OK", { provider, email: out.email });
  return out;
}
