/**
 * Integration tests for the auth store.
 *
 * These tests use the REAL authContext and REAL useAuthStore — no mocks.
 * They verify that:
 *   1. Store actions correctly delegate to authContext
 *   2. The store snapshot reflects authContext state (tokens, operator, org)
 *   3. Status derivation (unauthenticated → mfa_required → authenticated)
 *   4. Lockout state overrides everything to 'locked'
 *   5. clear() resets everything to unauthenticated
 *
 * No MSW is needed — the store doesn't make HTTP calls; it wraps authContext.
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { authContext, resetClientState, type LoginWithTokensResult, type MeResult, type OrganizationInfo } from '@vyzorix/api-client';
import { useAuthStore, type MfaChallenge } from '@/stores/auth-store';

const MOCK_ORG: OrganizationInfo = {
  id: 'org-test-1',
  name: 'Test Organization',
  role: 'admin',
};

function makeLoginResponse(overrides: Partial<LoginWithTokensResult> = {}): LoginWithTokensResult {
  return {
    operator_id: 'operator-1',
    email: 'test@vyzorix.com',
    name: 'Test Operator',
    mfa_enabled: false,
    access_token: 'access-token-123',
    refresh_token: 'refresh-token-123',
    expires_at: Math.floor(Date.now() / 1000) + 3600,
    needs_organization: false,
    organizations: [MOCK_ORG],
    last_organization_id: 'org-test-1',
    selected_organization: MOCK_ORG,
    ...overrides,
  };
}

function makeMeResponse(overrides: Partial<MeResult> = {}): MeResult {
  return {
    id: 'operator-1',
    email: 'test@vyzorix.com',
    name: 'Test Operator',
    mfa_enabled: false,
    email_verified: true,
    needs_organization: false,
    organizations: [MOCK_ORG],
    last_organization_id: 'org-test-1',
    selected_organization: MOCK_ORG,
    ...overrides,
  };
}

beforeEach(() => {
  vi.useFakeTimers();
  resetClientState();
  authContext.clear();
  useAuthStore.getState().clear();
  useAuthStore.getState().setMfaChallenge(null);
});

afterEach(() => {
  vi.clearAllTimers();
  vi.useRealTimers();
  resetClientState();
  authContext.clear();
});

describe('auth-store — setFromLoginWithTokens', () => {
  it('sets isAuthenticated, operator, tokens, and organizationId from login response', () => {
    const response = makeLoginResponse();
    useAuthStore.getState().setFromLoginWithTokens(response);

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.operator?.id).toBe('operator-1');
    expect(state.operator?.email).toBe('test@vyzorix.com');
    expect(state.operator?.name).toBe('Test Operator');
    expect(state.accessToken).toBe('access-token-123');
    expect(state.refreshToken).toBe('refresh-token-123');
    expect(state.organizationId).toBe('org-test-1');
    expect(state.status).toBe('authenticated');
  });

  it('derives needs_organization status when needs_organization=true', () => {
    const response = makeLoginResponse({
      needs_organization: true,
      organizations: [],
      selected_organization: undefined,
      last_organization_id: undefined,
    });
    useAuthStore.getState().setFromLoginWithTokens(response);

    const state = useAuthStore.getState();
    expect(state.status).toBe('needs_organization');
  });

  it('clears mfaChallenge on successful login', () => {
    useAuthStore.getState().setMfaChallenge({
      operatorId: 'operator-1',
      email: 'test@vyzorix.com',
      name: 'Test Operator',
      mfaEnabled: true,
    });
    expect(useAuthStore.getState().mfaChallenge).not.toBeNull();

    useAuthStore.getState().setFromLoginWithTokens(makeLoginResponse());
    expect(useAuthStore.getState().mfaChallenge).toBeNull();
    expect(useAuthStore.getState().status).toBe('authenticated');
  });
});

describe('auth-store — setFromMeResponse', () => {
  it('sets operator and organizationId from /me response', () => {
    const me = makeMeResponse();
    useAuthStore.getState().setFromMeResponse(me);

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.operator?.id).toBe('operator-1');
    expect(state.operator?.email).toBe('test@vyzorix.com');
    expect(state.organizationId).toBe('org-test-1');
    expect(state.status).toBe('authenticated');
  });

  it('derives needs_organization when /me says so', () => {
    const me = makeMeResponse({
      needs_organization: true,
      organizations: [],
      selected_organization: undefined,
      last_organization_id: undefined,
    });
    useAuthStore.getState().setFromMeResponse(me);
    expect(useAuthStore.getState().status).toBe('needs_organization');
  });
});

describe('auth-store — setMfaChallenge', () => {
  it('sets status to mfa_required', () => {
    const challenge: MfaChallenge = {
      operatorId: 'operator-1',
      email: 'test@vyzorix.com',
      name: 'Test Operator',
      mfaEnabled: true,
    };
    useAuthStore.getState().setMfaChallenge(challenge);

    const state = useAuthStore.getState();
    expect(state.mfaChallenge).toEqual(challenge);
    expect(state.status).toBe('mfa_required');
  });

  it('clearing mfaChallenge reverts to unauthenticated (if no auth)', () => {
    useAuthStore.getState().setMfaChallenge({
      operatorId: 'operator-1',
      email: 'test@vyzorix.com',
      name: 'Test Operator',
      mfaEnabled: true,
    });
    expect(useAuthStore.getState().status).toBe('mfa_required');

    useAuthStore.getState().setMfaChallenge(null);
    expect(useAuthStore.getState().status).toBe('unauthenticated');
  });
});

describe('auth-store — setOrganization', () => {
  it('updates organizationId', () => {
    useAuthStore.getState().setFromLoginWithTokens(makeLoginResponse());
    useAuthStore.getState().setOrganization('org-other');

    expect(useAuthStore.getState().organizationId).toBe('org-other');
  });

  it('can set organizationId to null', () => {
    useAuthStore.getState().setFromLoginWithTokens(makeLoginResponse());
    useAuthStore.getState().setOrganization(null);
    expect(useAuthStore.getState().organizationId).toBeNull();
  });
});

describe('auth-store — setAccessToken', () => {
  it('updates accessToken in store snapshot', () => {
    useAuthStore.getState().setFromLoginWithTokens(makeLoginResponse());
    useAuthStore.getState().setAccessToken('new-token-456');
    expect(useAuthStore.getState().accessToken).toBe('new-token-456');
  });
});

describe('auth-store — setLockout', () => {
  it('sets status to locked when lockout is active', () => {
    useAuthStore.getState().setFromLoginWithTokens(makeLoginResponse());
    expect(useAuthStore.getState().status).toBe('authenticated');

    useAuthStore.getState().setLockout({
      isLocked: true,
      retryAfter: 60,
      lockedUntil: Date.now() + 60000,
    });

    expect(useAuthStore.getState().isLocked).toBe(true);
    expect(useAuthStore.getState().status).toBe('locked');
  });

  it('unlocking reverts to authenticated (if auth was valid)', () => {
    useAuthStore.getState().setFromLoginWithTokens(makeLoginResponse());
    useAuthStore.getState().setLockout({
      isLocked: true,
      retryAfter: 60,
      lockedUntil: Date.now() + 60000,
    });
    expect(useAuthStore.getState().status).toBe('locked');

    useAuthStore.getState().setLockout({ isLocked: false, retryAfter: 0, lockedUntil: 0 });
    expect(useAuthStore.getState().status).toBe('authenticated');
  });

  it('lockout takes precedence over mfa_required', () => {
    useAuthStore.getState().setMfaChallenge({
      operatorId: 'operator-1',
      email: 'test@vyzorix.com',
      name: 'Test Operator',
      mfaEnabled: true,
    });
    expect(useAuthStore.getState().status).toBe('mfa_required');

    useAuthStore.getState().setLockout({
      isLocked: true,
      retryAfter: 30,
      lockedUntil: Date.now() + 30000,
    });
    expect(useAuthStore.getState().status).toBe('locked');
  });
});

describe('auth-store — clear', () => {
  it('resets to unauthenticated state', () => {
    useAuthStore.getState().setFromLoginWithTokens(makeLoginResponse());
    expect(useAuthStore.getState().isAuthenticated).toBe(true);

    useAuthStore.getState().clear();

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.operator).toBeNull();
    expect(state.accessToken).toBeNull();
    expect(state.refreshToken).toBeNull();
    expect(state.organizationId).toBeNull();
    expect(state.mfaChallenge).toBeNull();
    expect(state.status).toBe('unauthenticated');
  });

  it('clears lockout state', () => {
    useAuthStore.getState().setLockout({
      isLocked: true,
      retryAfter: 60,
      lockedUntil: Date.now() + 60000,
    });
    useAuthStore.getState().clear();
    expect(useAuthStore.getState().isLocked).toBe(false);
    expect(useAuthStore.getState().status).toBe('unauthenticated');
  });
});

describe('auth-store — status derivation priority', () => {
  it('locked > mfa_required > needs_organization > authenticated > unauthenticated', () => {
    // Start: locked + mfa challenge + needs_organization + authenticated
    useAuthStore.getState().setFromLoginWithTokens(makeLoginResponse({
      needs_organization: true,
      organizations: [],
      selected_organization: undefined,
      last_organization_id: undefined,
    }));
    useAuthStore.getState().setMfaChallenge({
      operatorId: 'operator-1',
      email: 'test@vyzorix.com',
      name: 'Test Operator',
      mfaEnabled: true,
    });
    useAuthStore.getState().setLockout({
      isLocked: true,
      retryAfter: 60,
      lockedUntil: Date.now() + 60000,
    });

    // Locked wins
    expect(useAuthStore.getState().status).toBe('locked');

    // Unlock → mfa_required wins
    useAuthStore.getState().setLockout({ isLocked: false, retryAfter: 0, lockedUntil: 0 });
    expect(useAuthStore.getState().status).toBe('mfa_required');

    // Clear MFA challenge → needs_organization wins
    useAuthStore.getState().setMfaChallenge(null);
    expect(useAuthStore.getState().status).toBe('needs_organization');

    // Set org → authenticated
    useAuthStore.getState().setFromMeResponse(makeMeResponse({ needs_organization: false }));
    expect(useAuthStore.getState().status).toBe('authenticated');

    // Clear → unauthenticated
    useAuthStore.getState().clear();
    expect(useAuthStore.getState().status).toBe('unauthenticated');
  });
});
