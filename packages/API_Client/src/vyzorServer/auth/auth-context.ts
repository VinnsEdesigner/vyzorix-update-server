import { setOrganizationContext, setAuthToken, setCSRFToken, restClient } from "../rest/_shared/rest-client";
import type { MeResponse, LoginWithTokensResponse } from "@/domain/auth";

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

let authState: AuthState = {
  isAuthenticated: false,
  operator: null,
  organizationId: null,
  accessToken: null,
  refreshToken: null,
  tokenExpiresAt: null,
};

const listeners = new Set<AuthChangeCallback>();
const refreshListeners = new Set<TokenRefreshCallback>();
let tokenRefreshTimer: ReturnType<typeof setTimeout> | null = null;

let lockoutState: LockoutState = {
  isLocked: false,
  retryAfter: 0,
  lockedUntil: 0,
};

function notifyListeners(): void {
  listeners.forEach((cb) => cb({ ...authState }));
}

function notifyRefreshListeners(tokens: { accessToken: string; refreshToken: string; expiresAt: number }): void {
  refreshListeners.forEach((cb) => cb(tokens));
}

function scheduleTokenRefresh(expiresAt: number): void {
    if (tokenRefreshTimer) {
    clearTimeout(tokenRefreshTimer);
  }
  
    const now = Date.now();
  const refreshTime = (expiresAt * 1000) - 60000;   
  if (refreshTime > now) {
    tokenRefreshTimer = setTimeout(() => {
      refreshTokens().catch(console.error);
    }, refreshTime - now);
  }
}

async function refreshTokens(): Promise<void> {
  const refreshToken = authState.refreshToken;
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

        authState.accessToken = response.access_token;
    authState.refreshToken = response.refresh_token;
    authState.tokenExpiresAt = response.expires_at;
    
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
  if (!authState.tokenExpiresAt) return false;
  return Date.now() >= (authState.tokenExpiresAt * 1000);
}

export function getTimeUntilExpiry(): number {
  if (!authState.tokenExpiresAt) return 0;
  return (authState.tokenExpiresAt * 1000) - Date.now();
}

export const authContext = {
    getState(): AuthState {
    return { ...authState };
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
    authState = {
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

    if (authState.accessToken) {
      setAuthToken(authState.accessToken);
    }
    if (authState.organizationId) {
      setOrganizationContext(authState.organizationId);
    }
        if (response.expires_at) {
      scheduleTokenRefresh(response.expires_at);
    }
    
    notifyListeners();
  },

    setFromMeResponse(me: MeResponse): void {
    authState.operator = me;
    authState.organizationId = me.selected_organization?.id || me.last_organization_id || null;
    authState.isAuthenticated = true;

    if (authState.organizationId) {
      setOrganizationContext(authState.organizationId);
    }

    notifyListeners();
  },

    setAccessToken(token: string | null): void {
    authState.accessToken = token;
    setAuthToken(token);
    notifyListeners();
  },

    setOrganization(orgId: string | null): void {
    authState.organizationId = orgId;
    setOrganizationContext(orgId);
    notifyListeners();
  },

    async refreshTokens(): Promise<void> {
    await refreshTokens();
  },

    setLockout(state: LockoutState): void {
    lockoutState = state;
  },

    getLockoutState(): LockoutState {
    return { ...lockoutState };
  },

    clear(): void {
    authState = {
      isAuthenticated: false,
      operator: null,
      organizationId: null,
      accessToken: null,
      refreshToken: null,
      tokenExpiresAt: null,
    };
    lockoutState = { isLocked: false, retryAfter: 0, lockedUntil: 0 };
    
    if (tokenRefreshTimer) {
      clearTimeout(tokenRefreshTimer);
      tokenRefreshTimer = null;
    }
    
    setAuthToken(null);
    setCSRFToken(null);
    setOrganizationContext(null);
    notifyListeners();
  },
};

export function getCurrentOrganizationId(): string | null {
  return authState.organizationId;
}

export function isAuthenticated(): boolean {
  return authState.isAuthenticated;
}

export function isAccountLocked(): boolean {
  if (!lockoutState.isLocked) return false;
    if (Date.now() >= lockoutState.lockedUntil) {
    lockoutState.isLocked = false;
    return false;
  }
  return true;
}
