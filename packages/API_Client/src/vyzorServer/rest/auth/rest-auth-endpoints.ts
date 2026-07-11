import { restClient } from "../_shared/rest-client";
import {
  loginRequestToRaw,
  registerRequestToRaw,
  refreshRequestToRaw,
  updateNameRequestToRaw,
  loginResponseFromRaw,
  isMFAResponse,
  registerResponseFromRaw,
  meResponseFromRaw,
  authTokensFromRaw,
} from "@/domain/auth";
import type {
  LoginResponse,
  RegisterResponse,
  MeResponse,
  AuthTokens,
  Operator,
} from "@/domain/auth";
import type { RawLoginResponse, RawLoginMFARequiredResponse } from "@/domain/auth/auth-mappers";

const AUTH_PATHS = {
  login: "/v1/auth/login",
  register: "/v1/auth/register",
  logout: "/v1/auth/logout",
  refresh: "/v1/auth/refresh",
  me: "/v1/auth/me",
  updateName: "/v1/auth/me",
  settings: "/v1/auth/me/settings",
  thresholds: "/v1/auth/me/thresholds",
  notifications: "/v1/auth/me/notifications",
} as const;

interface LoginResult {
  success: true;
  data: LoginResponse;
}

interface MFAResult {
  mfaRequired: true;
  operatorId: string;
}

export type LoginSuccessResult = LoginResult | MFAResult;

export async function login(
  credentials: { email: string; password: string }
): Promise<LoginSuccessResult> {
  const raw = await restClient.post<RawLoginResponse | RawLoginMFARequiredResponse>(
    AUTH_PATHS.login,
    loginRequestToRaw(credentials)
  );

  if (isMFAResponse(raw)) {
    return { mfaRequired: true, operatorId: raw.operator_id };
  }

  return { success: true, data: loginResponseFromRaw(raw) };
}

export async function register(
  credentials: { email: string; password: string; name: string }
): Promise<RegisterResponse> {
  const raw = await restClient.post<{ operator_id: string; email: string; name: string }>(
    AUTH_PATHS.register,
    registerRequestToRaw(credentials)
  );
  return registerResponseFromRaw(raw);
}

export async function logout(): Promise<{ success: boolean }> {
  const raw = await restClient.post<{ success: boolean }>(AUTH_PATHS.logout);
  clearStoredAuth();
  return raw;
}

export async function refreshToken(refreshToken: string): Promise<AuthTokens> {
  const raw = await restClient.post<{
    access_token: string;
    refresh_token: string;
    expires_at: number;
    session_id: string;
  }>(AUTH_PATHS.refresh, refreshRequestToRaw(refreshToken));
  return authTokensFromRaw(raw);
}

export async function getMe(): Promise<MeResponse | null> {
  try {
    const raw = await restClient.get<{
      id: string;
      email: string;
      name: string;
      role: string;
      mfa_enabled: boolean;
      email_verified: boolean;
      thresholds?: unknown;
      client?: unknown;
    }>(AUTH_PATHS.me);
    return meResponseFromRaw(raw as any);
  } catch {
    return null;
  }
}

export async function updateName(name: string): Promise<Operator> {
  const raw = await restClient.patch<{
    id: string;
    email: string;
    name: string;
    role: string;
    mfa_enabled: boolean;
    email_verified: boolean;
  }>(AUTH_PATHS.updateName, updateNameRequestToRaw(name));
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    role: raw.role as any,
    mfaEnabled: raw.mfa_enabled,
    emailVerified: raw.email_verified,
  };
}

const AUTH_TOKENS_KEY = "vyz_auth_tokens";

export interface StoredAuth {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  sessionId: string;
  operatorId: string;
  email: string;
}

export function storeAuth(auth: StoredAuth): void {
  try {
    localStorage.setItem(AUTH_TOKENS_KEY, JSON.stringify(auth));
  } catch {
    // localStorage unavailable
  }
}

export function getStoredAuth(): StoredAuth | null {
  try {
    const stored = localStorage.getItem(AUTH_TOKENS_KEY);
    return stored ? JSON.parse(stored) : null;
  } catch {
    return null;
  }
}

export function clearStoredAuth(): void {
  try {
    localStorage.removeItem(AUTH_TOKENS_KEY);
  } catch {
    // localStorage unavailable
  }
}
