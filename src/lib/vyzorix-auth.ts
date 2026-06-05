/**
 * vyzorix-auth.ts — Frontend authentication client for the Go/SQLite auth backend.
 *
 * All operator auth (login, register, Google OAuth, logout, me) goes through this module.
 * Tokens are stored in localStorage and sent as Bearer Authorization headers.
 *
 * Flow:
 *   - Login/Register → POST /v1/auth/login|register → JWT → stored in localStorage
 *   - Google OAuth  → GET /v1/auth/google → browser redirected to Google
 *                    → Google redirects to /v1/auth/google/callback → JWT in URL param
 *                    → callback page stores JWT and redirects to /dashboard
 *   - Protected calls → Authorization: Bearer <token> on every request
 *   - Logout → POST /v1/auth/logout (validates JWT, deletes session from DB)
 *   - Session check → GET /v1/auth/me (validates JWT, returns operator profile)
 */

import { logger } from "@/lib/logger";
import { useCallback } from "react";

// ─── Types ─────────────────────────────────────────────────────────────────────

export interface Operator {
  id: string;
  email: string;
  name: string;
  role: "viewer" | "operator" | "super_admin";
  createdAt: number;
}

export interface AuthResponse {
  token: string;
  expiresAt: number;
  operator: Operator;
}

export interface ErrorResponse {
  error: string;
  message: string;
}

// ─── Token storage ───────────────────────────────────────────────────────────

const TOKEN_KEY = "vyz.auth.token";
const OPERATOR_KEY = "vyz.auth.operator";

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

function setToken(token: string): void {
  try {
    localStorage.setItem(TOKEN_KEY, token);
  } catch {}
}

function clearToken(): void {
  try {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(OPERATOR_KEY);
  } catch {}
}

export function getStoredOperator(): Operator | null {
  try {
    const raw = localStorage.getItem(OPERATOR_KEY);
    return raw ? (JSON.parse(raw) as Operator) : null;
  } catch {
    return null;
  }
}

function setStoredOperator(op: Operator): void {
  try {
    localStorage.setItem(OPERATOR_KEY, JSON.stringify(op));
  } catch {}
}

// ─── Core API ─────────────────────────────────────────────────────────────────

async function jsonOrThrow<T>(res: Response): Promise<T> {
  const contentType = res.headers.get("content-type") ?? "";
  if (!res.ok) {
    let msg = `HTTP ${res.status}`;
    if (contentType.includes("application/json")) {
      const body = (await res.json()) as ErrorResponse;
      msg = body.message || body.error || msg;
    }
    throw new Error(msg);
  }
  if (contentType.includes("application/json")) {
    return (await res.json()) as T;
  }
  throw new Error(`Expected JSON, got ${contentType}`);
}

export async function login(serverUrl: string, email: string, password: string): Promise<AuthResponse> {
  logger.info("auth", `→ POST /v1/auth/login`, { email });
  const res = await fetch(`${serverUrl}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  const out = await jsonOrThrow<AuthResponse>(res);
  setToken(out.token);
  setStoredOperator(out.operator);
  logger.info("auth", `← login OK · ${out.operator.role} · ${out.operator.email}`);
  return out;
}

export async function register(serverUrl: string, email: string, password: string, name: string): Promise<AuthResponse> {
  logger.info("auth", `→ POST /v1/auth/register`, { email, name });
  const res = await fetch(`${serverUrl}/v1/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, name }),
  });
  const out = await jsonOrThrow<AuthResponse>(res);
  setToken(out.token);
  setStoredOperator(out.operator);
  logger.info("auth", `← register OK · ${out.operator.role} · ${out.operator.email}`);
  return out;
}

export async function logout(serverUrl: string): Promise<void> {
  const token = getToken();
  if (!token) {
    clearToken();
    return;
  }
  logger.info("auth", `→ POST /v1/auth/logout`);
  try {
    await fetch(`${serverUrl}/v1/auth/logout`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });
  } catch (e) {
    logger.warn("auth", `logout API failed: ${e instanceof Error ? e.message : String(e)}`);
  } finally {
    clearToken();
  }
}

export async function updateName(serverUrl: string, name: string): Promise<Operator> {
  const token = getToken();
  if (!token) throw new Error("not authenticated");
  logger.info("auth", `→ PATCH /v1/auth/me`, { name });
  const res = await fetch(`${serverUrl}/v1/auth/me`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name }),
  });
  const out = await jsonOrThrow<Operator>(res);
  setStoredOperator(out);
  logger.info("auth", `← name updated → ${out.name}`);
  return out;
}

export async function me(serverUrl: string): Promise<Operator> {
  const token = getToken();
  if (!token) throw new Error("not authenticated");
  logger.info("auth", `→ GET /v1/auth/me`);
  const res = await fetch(`${serverUrl}/v1/auth/me`, {
    method: "GET",
    headers: { Authorization: `Bearer ${token}` },
  });
  const out = await jsonOrThrow<Operator>(res);
  setStoredOperator(out);
  return out;
}

// ─── Google OAuth redirect ─────────────────────────────────────────────────────

/**
 * Initiates the Google OAuth flow by redirecting the browser to the Go server's
 * /v1/auth/google endpoint. The Go server redirects to Google's consent screen.
 * After the user approves, Google redirects back to /v1/auth/google/callback on
 * the Go server, which then redirects the browser to the frontend callback page
 * with the JWT token embedded in the URL.
 */
export function redirectToGoogleOAuth(serverUrl: string, frontendCallbackPath = "/"): void {
  const target = `${serverUrl}/v1/auth/google?state=${encodeURIComponent(frontendCallbackPath)}`;
  logger.info("auth", `→ GET /v1/auth/google (OAuth redirect)`, { target });
  window.location.href = target;
}

/**
 * Processes the OAuth callback on the frontend.
 * Call this on the /auth/callback route when the URL contains ?token=...&isNew=...
 * It stores the token and returns the auth data.
 */
export function handleOAuthCallback(token: string, isNew: string): AuthResponse | null {
  try {
    // Decode the operator from the token's payload (we don't verify signature client-side)
    const parts = token.split(".");
    if (parts.length !== 3) return null;
    const payload = JSON.parse(atob(parts[1].replace(/-/g, "+").replace(/_/g, "/")));
    const operator: Operator = {
      id: payload.oid ?? "",
      email: payload.email ?? "",
      name: payload.name ?? payload.email?.split("@")[0] ?? "Operator",
      role: (payload.role as Operator["role"]) ?? "operator",
      createdAt: payload.iat ? payload.iat * 1000 : Date.now(),
    };
    if (!operator.id || !operator.email) return null;
    setToken(token);
    setStoredOperator(operator);
    logger.info("auth", `OAuth callback OK · ${operator.role} · ${operator.email}`);
    return { token, expiresAt: (payload.exp ?? 0) * 1000, operator };
  } catch {
    return null;
  }
}

// ─── Hooks ─────────────────────────────────────────────────────────────────────

export function useAuth() {
  const token = getToken();
  const operator = getStoredOperator();

  const isAuthenticated = useCallback((): boolean => {
    return !!token && !!operator;
  }, [token, operator]);

  return {
    token,
    operator,
    isAuthenticated: isAuthenticated(),
  };
}