import { createVyzorStore } from '@/lib/state';
import {
  authContext,
  getMe,
  type AuthState,
  type LockoutState,
  type LoginWithTokensResponse,
  type MeResponse,
  type MFAVerifyResponse,
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
  setFromMfaVerify: (response: MFAVerifyResponse) => Promise<MeResponse | null>;
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

export const useAuthStore = createVyzorStore<AuthStoreState>('AuthStore', (set) => {
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
    setFromMfaVerify: async (response) => {
      authContext.setFromMfaVerify(response);
      set(buildSnapshot({ mfaChallenge: null }));
      // The MFA verify response carries only a partial operator. Fetch /me
      // to populate organizations/memberships/needs_organization so the
      // derived status (needs_organization vs authenticated) is correct.
      if (response.success && response.accessToken) {
        try {
          const me = await getMe();
          if (me) {
            authContext.setFromMeResponse(me);
            set(buildSnapshot({ mfaChallenge: null }));
            return me;
          }
        } catch {
          // Tokens are already set; the operator may be incomplete but the
          // session is valid. A later /me query will reconcile.
        }
      }
      return null;
    },
    setMfaChallenge: (challenge) => {
      set(buildSnapshot({ mfaChallenge: challenge }));
    },
    setOrganization: (orgId) => {
      authContext.setOrganization(orgId);
      set((state) => buildSnapshot({ mfaChallenge: state.mfaChallenge }));
    },
    setAccessToken: (token) => {
      authContext.setAccessToken(token);
      set((state) => buildSnapshot({ mfaChallenge: state.mfaChallenge }));
    },
    refreshTokens: async () => {
      await authContext.refreshTokens();
      set((state) => buildSnapshot({ mfaChallenge: state.mfaChallenge }));
    },
    setLockout: (lockout) => {
      authContext.setLockout(lockout);
      set((state) => buildSnapshot({ mfaChallenge: state.mfaChallenge }));
    },
    clear: () => {
      authContext.clear();
      set(buildSnapshot({ mfaChallenge: null }));
    },
  };
});
