import { SuccessReport } from "../../types";
import { ARCHITECTURE_CONFIG } from "../config";

const IS_SIMULATED = ARCHITECTURE_CONFIG.IS_SIMULATED;
const API_BASE = ARCHITECTURE_CONFIG.API_BASE_URL;

/**
 * Client 2: Single Sign-On (SSO) Connection Client (Google & GitHub)
 */
export async function initiateSSO(provider: "Google" | "GitHub"): Promise<SuccessReport> {
  if (IS_SIMULATED) {
    return new Promise((resolve) => {
      setTimeout(() => {
        const email = provider === "Google" ? "sso-google@vyzorix.com" : "sso-github@vyzorix.com";
        const username = provider === "Google" ? "google_member" : "github_member";
        const fullName =
          provider === "Google" ? "Google Authenticated User" : "GitHub Operator Tech";

        resolve({
          fullName,
          email,
          username,
          memberId: "VXZ-64981",
          operatorRole: "Operator",
          region: "Paris, France",
          createdAt: new Date().toISOString().replace("T", " ").substring(0, 19) + " UTC",
          method: "SSO",
        });
      }, 1000);
    });
  }

  // Production Go Integration:
  // Dispatches a state token handshake, then redirects the operator to authorize with the provider.
  if (typeof window !== "undefined") {
    window.location.href = `${API_BASE}/sso/${provider.toLowerCase()}`;
  }

  throw new Error("Redirecting to Single Sign-On window...");
}

/**
 * Validates code query token received from Google or GitHub redirection exchange.
 * Used internally, or during Go URL handshake validation on page load.
 */
export async function handleSSOCallback(
  provider: "Google" | "GitHub",
  code: string,
  state: string,
): Promise<SuccessReport> {
  if (IS_SIMULATED) {
    return initiateSSO(provider);
  }

  const response = await fetch(
    `${API_BASE}/sso/${provider.toLowerCase()}/callback?code=${code}&state=${state}`,
  );
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || `${provider} authorization validation handshakes failed.`);
  }

  return response.json();
}
