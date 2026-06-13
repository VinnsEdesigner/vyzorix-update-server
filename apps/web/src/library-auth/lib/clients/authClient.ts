import { SuccessReport } from "../../types";
import { ARCHITECTURE_CONFIG } from "../config";

const IS_SIMULATED = ARCHITECTURE_CONFIG.IS_SIMULATED;
const API_BASE = ARCHITECTURE_CONFIG.API_BASE_URL;

export interface SignUpPayload {
  fullName: string;
  email: string;
  username: string;
}

/**
 * Client 1: Credentials & Standard Security Authentication Client
 */
export async function registerOperator(
  payload: SignUpPayload,
): Promise<{ success: boolean; message: string }> {
  if (IS_SIMULATED) {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          success: true,
          message: `Oauth operator registration completed. Secure verification token sent to ${payload.email}`,
        });
      }, 1000);
    });
  }

  // Go REST Endpoint: Register Operator -> inserts UUIDv7 operator registry
  const response = await fetch(`${API_BASE}/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || "Operator account registration failed.");
  }

  return response.json();
}

export async function loginOperator(
  usernameOrEmail: string,
  passwordSecret: string,
): Promise<SuccessReport> {
  if (IS_SIMULATED) {
    return new Promise((resolve) => {
      setTimeout(() => {
        const atIndex = usernameOrEmail.indexOf("@");
        const username = atIndex > 0 ? usernameOrEmail.substring(0, atIndex) : usernameOrEmail;
        resolve({
          fullName: "Alexis Thorne",
          email: usernameOrEmail.includes("@") ? usernameOrEmail : `${usernameOrEmail}@vyzorix.com`,
          username,
          memberId: "VXZ-64981",
          operatorRole: "Operator",
          region: "Paris, France",
          createdAt: new Date().toISOString().replace("T", " ").substring(0, 19) + " UTC",
          method: "Standard Email",
        });
      }, 1200);
    });
  }

  // Go REST Endpoint: Auth Credentials login -> Returns success token session
  const response = await fetch(`${API_BASE}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ identity: usernameOrEmail, password: passwordSecret }),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || "Authentication rejected by security systems.");
  }

  return response.json();
}

export async function requestPasswordReset(
  emailAddress: string,
): Promise<{ success: boolean; message: string }> {
  if (IS_SIMULATED) {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          success: true,
          message: "Password restoration protocol initiated. Link has been sent.",
        });
      }, 1000);
    });
  }

  // Go REST Endpoint: Password reset dispatch
  const response = await fetch(`${API_BASE}/forgot-password`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: emailAddress }),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || "Restoration dispatch rejected.");
  }

  return response.json();
}

/**
 * Client 1.B: Session Check Endpoint
 * Checks if the current operator has an active session cookie or token on the Go server.
 */
export async function getCurrentSession(): Promise<SuccessReport | null> {
  if (IS_SIMULATED) {
    // Under simulation, check localStorage (implemented in the hydration layer)
    return null;
  }

  try {
    const response = await fetch(`${API_BASE}/me`);
    if (!response.ok) {
      return null;
    }
    return await response.json();
  } catch {
    return null;
  }
}

/**
 * Client 1.C: Logout Session Endpoint
 * Instructs the server to invalidate session and pending cookies.
 */
export async function logoutOperator(): Promise<{ success: boolean }> {
  if (IS_SIMULATED) {
    return { success: true };
  }

  try {
    const response = await fetch(`${API_BASE}/logout`, {
      method: "POST",
    });
    if (!response.ok) {
      return { success: false };
    }
    return await response.json();
  } catch {
    return { success: false };
  }
}
