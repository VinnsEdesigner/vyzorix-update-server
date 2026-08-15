import { setOrganizationContext, setAuthToken, setCSRFToken, restClient } from "../rest/_shared/rest-client";
import type { MeResponse, LoginWithTokensResponse } from "../../domain/auth";

export interface AuthState {
  isAuthenticated: boolean;
  operator: MeResponse | null;
  organizationId: string | null;
  accessToken: string | null;
  refreshToken: string | null;
  tokenExpiresAt: number | null;
}

type TokenRefreshCallback = (tokens: { accessToken: string; refreshToken: string; expiresAt: number }) => void;

export interface LockoutState {
  isLocked: boolean;
  retryAfter: number;
  lockedUntil: number;
}

type AuthChangeCallback = (state: AuthState) => void;

interface AuthContextState {
  auth: AuthState;
  refreshTimer: ReturnType<typeof setTimeout> | null;
  lockout: LockoutState;
}

// Single encapsulated holder for all auth mutable state (previously three
// separate module-level `let` bindings — the singleton-globals smell).
const authContextState: AuthContextState = {
  auth: {
    isAuthenticated: false,
    operator: null,
    organizationId: null,
    accessToken: null,
    refreshToken: null,
    tokenExpiresAt: null,
  },
  refreshTimer: null,
  lockout: {
    isLocked: false,
    retryAfter: 0,
    lockedUntil: 0,
  },
};

const listeners = new Set<AuthChangeCallback>();
const refreshListeners = new Set<TokenRefreshCallback>();

function notifyListeners(): void {
  listeners.forEach((cb) => cb({ ...authContextState.auth }));
}

function notifyRefreshListeners(tokens: { accessToken: string; refreshToken: string; expiresAt: number }): void {
  refreshListeners.forEach((cb) => cb(tokens));
}

function scheduleTokenRefresh(expiresAt: number): void {
    if (authContextState.refreshTimer) {
    clearTimeout(authContextState.refreshTimer);
  }
  
    const now = Date.now();
  const refreshTime = (expiresAt * 1000) - 60000;   
  if (refreshTime > now) {
    authContextState.refreshTimer = setTimeout(() => {
      refreshTokens().catch(console.error);
    }, refreshTime - now);
  }
}

async function refreshTokens(): Promise<void> {
  const refreshToken = authContextState.auth.refreshToken;
  if (!refreshToken) {
    console.warn('[Auth] No refresh token available');
    return;
  }

  try {
    const response = await restClient.post<{
      access_token: string;
      refresh_token: string;
      expires_at: number;
      session_id: string;
    }>('/v1/auth/refresh', { refresh_token: refreshToken });

        authContextState.auth.accessToken = response.access_token;
    authContextState.auth.refreshToken = response.refresh_token;
    authContextState.auth.tokenExpiresAt = response.expires_at;
    
    setAuthToken(response.access_token);
    
        scheduleTokenRefresh(response.expires_at);
    
        notifyRefreshListeners({
      accessToken: response.access_token,
      refreshToken: response.refresh_token,
      expiresAt: response.expires_at,
    });
    
    console.debug('[Auth] Tokens refreshed successfully');
  } catch (error) {
    console.error('[Auth] Token refresh failed:', error);
    authContext.clear();
    throw error;
  }
}

export function isTokenExpired(): boolean {
  if (!authContextState.auth.tokenExpiresAt) return false;
  return Date.now() >= (authContextState.auth.tokenExpiresAt * 1000);
}

export function getTimeUntilExpiry(): number {
  if (!authContextState.auth.tokenExpiresAt) return 0;
  return (authContextState.auth.tokenExpiresAt * 1000) - Date.now();
}

export const authContext = {
    getState(): AuthState {
    return { ...authContextState.auth };
  },

    onChange(callback: AuthChangeCallback): () => void {
    listeners.add(callback);
    return () => listeners.delete(callback);
  },

    onTokenRefresh(callback: TokenRefreshCallback): () => void {
    refreshListeners.add(callback);
    return () => refreshListeners.delete(callback);
  },

    setFromLoginWithTokens(response: LoginWithTokensResponse): void {
    authContextState.auth = {
      isAuthenticated: true,
      operator: {
        id: "",
        email: response.email,
        name: response.name,
        mfa_enabled: response.mfa_enabled,
        email_verified: false,
        needs_organization: response.needs_organization,
        organizations: response.organizations || [],
        memberships: [],
        last_organization_id: response.last_organization_id,
        selected_organization: response.selected_organization,
      },
      organizationId: response.selected_organization?.id || response.last_organization_id || null,
      accessToken: response.access_token,
      refreshToken: response.refresh_token || null,
      tokenExpiresAt: response.expires_at || null,
    };

    if (authContextState.auth.accessToken) {
      setAuthToken(authContextState.auth.accessToken);
    }
    if (authContextState.auth.organizationId) {
      setOrganizationContext(authContextState.auth.organizationId);
    }
        if (response.expires_at) {
      scheduleTokenRefresh(response.expires_at);
    }
    
    notifyListeners();
  },

    setFromMeResponse(me: MeResponse): void {
    authContextState.auth.operator = me;
    authContextState.auth.organizationId = me.selected_organization?.id || me.last_organization_id || null;
    authContextState.auth.isAuthenticated = true;

    if (authContextState.auth.organizationId) {
      setOrganizationContext(authContextState.auth.organizationId);
    }

    notifyListeners();
  },

    setAccessToken(token: string | null): void {
    authContextState.auth.accessToken = token;
    setAuthToken(token);
    notifyListeners();
  },

    setOrganization(orgId: string | null): void {
    authContextState.auth.organizationId = orgId;
    setOrganizationContext(orgId);
    notifyListeners();
  },

    async refreshTokens(): Promise<void> {
    await refreshTokens();
  },

    setLockout(state: LockoutState): void {
    authContextState.lockout = state;
  },

    getLockoutState(): LockoutState {
    return { ...authContextState.lockout };
  },

    clear(): void {
    authContextState.auth = {
      isAuthenticated: false,
      operator: null,
      organizationId: null,
      accessToken: null,
      refreshToken: null,
      tokenExpiresAt: null,
    };
    authContextState.lockout = { isLocked: false, retryAfter: 0, lockedUntil: 0 };
    
    if (authContextState.refreshTimer) {
      clearTimeout(authContextState.refreshTimer);
      authContextState.refreshTimer = null;
    }
    
    setAuthToken(null);
    setCSRFToken(null);
    setOrganizationContext(null);
    notifyListeners();
  },
};

export function getCurrentOrganizationId(): string | null {
  return authContextState.auth.organizationId;
}

export function isAuthenticated(): boolean {
  return authContextState.auth.isAuthenticated;
}

export function isAccountLocked(): boolean {
  if (!authContextState.lockout.isLocked) return false;
    if (Date.now() >= authContextState.lockout.lockedUntil) {
    authContextState.lockout.isLocked = false;
    return false;
  }
  return true;
}
