import { create } from 'zustand';
import {
  authContext,
  type AuthState,
  type LockoutState,
  type LoginWithTokensResponse,
  type MeResponse,
} from '@vyzorix/api-client';

export interface MfaChallenge {
  operatorId: string;
  email: string;
  name: string;
  mfaEnabled: boolean;
}

export type AuthStatus =
  | 'unauthenticated'
  | 'mfa_required'
  | 'needs_organization'
  | 'authenticated'
  | 'locked';

export interface AuthStoreState extends AuthState, LockoutState {
  mfaChallenge: MfaChallenge | null;
  status: AuthStatus;
  setFromLoginWithTokens: (response: LoginWithTokensResponse) => void;
  setFromMeResponse: (me: MeResponse) => void;
  setMfaChallenge: (challenge: MfaChallenge | null) => void;
  setOrganization: (orgId: string | null) => void;
  setAccessToken: (token: string | null) => void;
  refreshTokens: () => Promise<void>;
  setLockout: (state: LockoutState) => void;
  clear: () => void;
}

function deriveStatus(
  auth: AuthState,
  lockout: LockoutState,
  mfaChallenge: MfaChallenge | null,
): AuthStatus {
  if (lockout.isLocked) return 'locked';
  if (mfaChallenge) return 'mfa_required';
  if (!auth.isAuthenticated || !auth.operator) return 'unauthenticated';
  if (auth.operator.needs_organization) return 'needs_organization';
  return 'authenticated';
}

function buildSnapshot(overrides?: Partial<AuthStoreState>): Partial<AuthStoreState> {
  const auth = authContext.getState();
  const lockout = authContext.getLockoutState();
  const base: Partial<AuthStoreState> = {
    isAuthenticated: auth.isAuthenticated,
    operator: auth.operator,
    organizationId: auth.organizationId,
    accessToken: auth.accessToken,
    refreshToken: auth.refreshToken,
    tokenExpiresAt: auth.tokenExpiresAt,
    isLocked: lockout.isLocked,
    retryAfter: lockout.retryAfter,
    lockedUntil: lockout.lockedUntil,
  };
  const merged = { ...base, ...overrides } as AuthStoreState;
  return { ...merged, status: deriveStatus(auth, lockout, merged.mfaChallenge ?? null) };
}

export const useAuthStore = create<AuthStoreState>((set) => {
  const initial = buildSnapshot() as AuthStoreState;

  authContext.onChange(() => {
    set(buildSnapshot());
  });

  return {
    ...initial,
    mfaChallenge: null,
    setFromLoginWithTokens: (response) => {
      authContext.setFromLoginWithTokens(response);
      set(buildSnapshot({ mfaChallenge: null }));
    },
    setFromMeResponse: (me) => {
      authContext.setFromMeResponse(me);
      set(buildSnapshot({ mfaChallenge: null }));
    },
    setMfaChallenge: (challenge) => {
      set(buildSnapshot({ mfaChallenge: challenge }));
    },
    setOrganization: (orgId) => {
      authContext.setOrganization(orgId);
      set(buildSnapshot());
    },
    setAccessToken: (token) => {
      authContext.setAccessToken(token);
      set(buildSnapshot());
    },
    refreshTokens: async () => {
      await authContext.refreshTokens();
      set(buildSnapshot());
    },
    setLockout: (lockout) => {
      authContext.setLockout(lockout);
      set(buildSnapshot());
    },
    clear: () => {
      authContext.clear();
      set(buildSnapshot({ mfaChallenge: null }));
    },
  };
});
