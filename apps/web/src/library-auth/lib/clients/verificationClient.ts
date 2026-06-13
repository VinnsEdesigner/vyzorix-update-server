import { SuccessReport } from "../../types";
import { ARCHITECTURE_CONFIG } from "../config";

const IS_SIMULATED = ARCHITECTURE_CONFIG.IS_SIMULATED;
const API_BASE = ARCHITECTURE_CONFIG.API_BASE_URL;

/**
 * Client 3: Operator Verification Polling & Lifecycle Client
 */
export async function pollVerificationStatus(
  token: string,
): Promise<{ status: "waiting" | "success"; report?: SuccessReport }> {
  if (IS_SIMULATED) {
    return new Promise((resolve) => {
      // Automatically mocks a pending state
      setTimeout(() => {
        resolve({ status: "waiting" });
      }, 500);
    });
  }

  // Go REST Endpoint: Polls SQLite DB for token verification status using standard UUIDv7
  const response = await fetch(`${API_BASE}/poll-verification?token=${encodeURIComponent(token)}`);
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || "Verification link check failed.");
  }

  return response.json();
}

/**
 * Dispatches request for standard email link transmission (resend action)
 */
export async function triggerTokenResend(
  email: string,
): Promise<{ success: boolean; message: string }> {
  if (IS_SIMULATED) {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          success: true,
          message: `Verification link successfully re-routed to ${email}`,
        });
      }, 800);
    });
  }

  // Go REST Endpoint: Resend validation link trigger
  const response = await fetch(`${API_BASE}/resend-token`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || "Verification link re-routing failed.");
  }

  return response.json();
}

/**
 * Cancels pending token validation pipelines
 */
export async function cancelVerificationSession(email: string): Promise<{ success: boolean }> {
  if (IS_SIMULATED) {
    return { success: true };
  }

  // Go REST Endpoint: Terminate validation link mapping
  const response = await fetch(`${API_BASE}/cancel-verification`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });

  if (!response.ok) {
    throw new Error("Lock recycle termination failed.");
  }

  return response.json();
}
