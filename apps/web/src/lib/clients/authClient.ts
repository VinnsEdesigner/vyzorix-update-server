/**
 * authClient.ts - Credentials authentication client for vyzorix-update-server.
 *
 * Handles register, login, logout, and session management.
 * Uses HttpOnly cookies for session management.
 */

import { logger } from "@/lib/logger";

// API base URL - defaults to same origin for SSR
const API_BASE = "";

// Types
export interface SignUpPayload {
  fullName: string;
  email: string;
  password: string;
}

export interface LoginPayload {
  identity: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  expiresAt: number;
  operator: OperatorResponse;
}

export interface OperatorResponse {
  id: string;
  email: string;
  name: string;
  role: "viewer" | "operator" | "super_admin";
  createdAt: number;
  emailVerified?: boolean;
}

export interface MessageResponse {
  message: string;
}

interface ErrorResponse {
  error: string;
  message: string;
}

// JSON response helper
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
 * Register a new operator account.
 */
export async function registerOperator(payload: SignUpPayload): Promise<MessageResponse> {
  logger.info("auth", "-> POST /v1/auth/register", {
    email: payload.email,
  });
  const res = await fetch(`${API_BASE}/v1/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(payload),
  });
  const out = await jsonOrThrow<MessageResponse>(res);
  logger.info("auth", "<- register OK");
  return out;
}

/**
 * Login with email/password credentials.
 */
export async function loginOperator(identity: string, password: string): Promise<AuthResponse> {
  logger.info("auth", "-> POST /v1/auth/login", { identity });
  const res = await fetch(`${API_BASE}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ identity, password }),
  });
  const out = await jsonOrThrow<AuthResponse>(res);
  logger.info("auth", "<- login OK", { role: out.operator.role });
  return out;
}

/**
 * Logout the current operator.
 */
export async function logoutOperator(): Promise<void> {
  logger.info("auth", "-> POST /v1/auth/logout");
  try {
    await fetch(`${API_BASE}/v1/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
  } catch (e) {
    logger.warn("auth", `logout failed: ${e instanceof Error ? e.message : String(e)}`);
  }
  logger.info("auth", "<- logout OK");
}

/**
 * Get current operator profile from session cookie.
 */
export async function getCurrentSession(): Promise<OperatorResponse | null> {
  logger.info("auth", "-> GET /v1/auth/me");
  try {
    const res = await fetch(`${API_BASE}/v1/auth/me`, {
      method: "GET",
      credentials: "include",
    });
    if (!res.ok) {
      return null;
    }
    const out = await jsonOrThrow<OperatorResponse>(res);
    logger.info("auth", "<- session OK", { email: out.email });
    return out;
  } catch {
    return null;
  }
}
