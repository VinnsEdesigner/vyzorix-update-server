

import { restClient, getCSRFToken, fetchAndSetCSRFToken, clearAuthContext, setAuthToken, setRefreshToken } from "../_shared/rest-client";
import {
  loginRequestToRaw,
  registerRequestToRaw,
  updateNameRequestToRaw,
  loginResponseFromRaw,
  isMFAResponse,
  registerResponseFromRaw,
  meResponseFromRaw,
  authTokensFromRaw,
} from "../../../domain/auth";
import type {
  LoginResponse,
  LoginWithTokensResponse,
  RegisterResponse,
  MeResponse,
  OperatorRole,
  AuthTokens,
} from "../../../domain/auth";
import {
  type RawLoginResponse,
  type RawLoginMFARequiredResponse,
  type RawLoginWithTokensResponse,
  type RawLoginWithTokensMFARequiredResponse,
  type RawMeResponse,
  loginWithTokensResponseFromRaw,
  isLoginWithTokensMFARequired,
} from "../../../domain/auth/auth-mappers";

const AUTH_PATHS = {
  login: "/v1/auth/login",
  loginTokens: "/v1/auth/login/tokens",
  register: "/v1/auth/register",
  logout: "/v1/auth/logout",
  refresh: "/v1/auth/refresh",
  me: "/v1/auth/me",
  updateName: "/v1/auth/me",
  sessions: "/v1/auth/sessions",
  sessionsConcurrent: "/v1/auth/sessions/concurrent",
  sessionsRevokeAll: "/v1/auth/sessions/revoke-all",
} as const;

async function ensureCSRFToken(): Promise<void> {
  if (!getCSRFToken()) {
    await fetchAndSetCSRFToken();
  }
}

interface LoginSuccess {
  success: true;
  data: LoginResponse;
}

interface LoginMFARequired {
  mfaRequired: true;
  operatorId: string;
}

export type LoginResult = LoginSuccess | LoginMFARequired;

interface LoginWithTokensSuccess {
  success: true;
  data: LoginWithTokensResponse;
}

interface LoginWithTokensMFARequired {
  mfaRequired: true;
  operatorId: string;
  email: string;
  name: string;
  mfaEnabled: boolean;
}

export type LoginWithTokensResult = LoginWithTokensSuccess | LoginWithTokensMFARequired;

export async function fetchCSRFToken(): Promise<string> {
  return fetchAndSetCSRFToken();
}

export async function login(credentials: {
  email: string;
  password: string;
}): Promise<LoginResult> {
  await ensureCSRFToken();
  const raw = await restClient.post<RawLoginResponse | RawLoginMFARequiredResponse>(
    AUTH_PATHS.login,
    loginRequestToRaw(credentials)
  );

  if (isMFAResponse(raw)) {
    return { mfaRequired: true, operatorId: raw.operator_id };
  }

  return { success: true, data: loginResponseFromRaw(raw) };
}


export async function loginWithTokens(credentials: {
  email: string;
  password: string;
}): Promise<LoginWithTokensResult> {
  await ensureCSRFToken();
  const raw = await restClient.post<RawLoginWithTokensResponse | RawLoginWithTokensMFARequiredResponse>(
    AUTH_PATHS.loginTokens,
    loginRequestToRaw(credentials)
  );

  if (isLoginWithTokensMFARequired(raw)) {
    return {
      mfaRequired: true,
      operatorId: raw.operator_id,
      email: raw.email,
      name: raw.name,
      mfaEnabled: raw.mfa_enabled,
    };
  }

  const result = loginWithTokensResponseFromRaw(raw);
  setAuthToken(result.access_token);
  setRefreshToken(result.refresh_token);
  return { success: true, data: result };
}


export async function register(credentials: {
  email: string;
  password: string;
  name: string;
}): Promise<RegisterResponse> {
  await ensureCSRFToken();
  const raw = await restClient.post<{
    operator_id: string;
    email: string;
    name: string;
  }>(AUTH_PATHS.register, registerRequestToRaw(credentials));
  return registerResponseFromRaw(raw);
}


export async function logout(): Promise<{ success: boolean }> {
  const result = await restClient.post<{ success: boolean }>(AUTH_PATHS.logout);
  clearAuthContext();
  return result;
}


export async function getMe(): Promise<MeResponse | null> {
  try {
    const raw = await restClient.get<RawMeResponse>(AUTH_PATHS.me);
    return meResponseFromRaw(raw);
  } catch {
    return null;
  }
}

export interface UpdateNameResponse {
  id: string;
  email: string;
  name: string;
  role?: string;
  mfa_enabled: boolean;
  email_verified: boolean;
}

export async function updateName(name: string): Promise<UpdateNameResponse> {
  const raw = await restClient.patch<{
    id: string;
    email: string;
    name: string;
    role: OperatorRole;
    mfa_enabled: boolean;
    email_verified: boolean;
  }>(AUTH_PATHS.updateName, updateNameRequestToRaw(name));
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    mfa_enabled: raw.mfa_enabled,
    email_verified: raw.email_verified,
  };
}


export async function refreshToken(refreshToken: string): Promise<AuthTokens> {
  const raw = await restClient.post<{
    access_token: string;
    refresh_token: string;
    expires_at: number;
    session_id: string;
  }>(AUTH_PATHS.refresh, { refresh_token: refreshToken });
  return authTokensFromRaw(raw);
}
