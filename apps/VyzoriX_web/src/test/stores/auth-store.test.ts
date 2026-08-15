import { describe, it, expect, beforeEach, vi } from 'vitest';

const { mockAuthState, mockLockoutState, listeners, authContextMock } = vi.hoisted(() => {
  const mockAuthState = {
    isAuthenticated: false,
    operator: null as Record<string, unknown> | null,
    organizationId: null as string | null,
    accessToken: null as string | null,
    refreshToken: null as string | null,
    tokenExpiresAt: null as number | null,
  };

  const mockLockoutState = { isLocked: false, retryAfter: 0, lockedUntil: 0 };

  const listeners = new Set<() => void>();

  const authContextMock = {
    getState: vi.fn(() => ({ ...mockAuthState })),
    getLockoutState: vi.fn(() => ({ ...mockLockoutState })),
    onChange: vi.fn((cb: () => void) => {
      listeners.add(cb);
      return () => listeners.delete(cb);
    }),
    setFromLoginWithTokens: vi.fn((response: Record<string, unknown>) => {
      mockAuthState.isAuthenticated = true;
      mockAuthState.accessToken = response.access_token as string;
      mockAuthState.refreshToken = (response.refresh_token as string) ?? null;
      mockAuthState.tokenExpiresAt = (response.expires_at as number) ?? null;
      mockAuthState.organizationId =
        (response.selected_organization as { id?: string } | undefined)?.id ??
        (response.last_organization_id as string) ??
        null;
      mockAuthState.operator = {
        id: '',
        email: response.email ?? '',
        name: response.name ?? '',
        mfa_enabled: response.mfa_enabled ?? false,
        email_verified: false,
        needs_organization: response.needs_organization ?? false,
        organizations: response.organizations ?? [],
        memberships: [],
        last_organization_id: response.last_organization_id ?? null,
        selected_organization: response.selected_organization ?? null,
      };
    }),
    setFromMeResponse: vi.fn((me: Record<string, unknown>) => {
      mockAuthState.operator = me;
      mockAuthState.isAuthenticated = true;
      mockAuthState.organizationId =
        (me.selected_organization as { id?: string } | undefined)?.id ??
        (me.last_organization_id as string) ??
        null;
    }),
    setAccessToken: vi.fn((token: string | null) => {
      mockAuthState.accessToken = token;
    }),
    setOrganization: vi.fn((orgId: string | null) => {
      mockAuthState.organizationId = orgId;
    }),
    refreshTokens: vi.fn(async () => {}),
    setLockout: vi.fn((state: typeof mockLockoutState) => {
      Object.assign(mockLockoutState, state);
    }),
    clear: vi.fn(() => {
      Object.assign(mockAuthState, {
        isAuthenticated: false,
        operator: null,
        organizationId: null,
        accessToken: null,
        refreshToken: null,
        tokenExpiresAt: null,
      });
      Object.assign(mockLockoutState, { isLocked: false, retryAfter: 0, lockedUntil: 0 });
    }),
  };

  return { mockAuthState, mockLockoutState, listeners, authContextMock };
});

vi.mock('@vyzorix/api-client', () => ({
  authContext: authContextMock,
  getCurrentOrganizationId: vi.fn(() => mockAuthState.organizationId),
  initConnectivityMonitor: vi.fn(),
  getConnectivityMonitor: vi.fn(),
}));

const { useAuthStore } = await import('@/stores/auth-store');

function notifyListeners() {
  for (const cb of listeners) cb();
}

describe('useAuthStore', () => {
  beforeEach(() => {
    Object.assign(mockAuthState, {
      isAuthenticated: false,
      operator: null,
      organizationId: null,
      accessToken: null,
      refreshToken: null,
      tokenExpiresAt: null,
    });
    Object.assign(mockLockoutState, { isLocked: false, retryAfter: 0, lockedUntil: 0 });
    listeners.clear();
    vi.clearAllMocks();
    useAuthStore.getState().clear();
  });

  it('starts unauthenticated', () => {
    const state = useAuthStore.getState();
    expect(state.status).toBe('unauthenticated');
    expect(state.isAuthenticated).toBe(false);
    expect(state.operator).toBeNull();
  });

  it('setFromLoginWithTokens delegates to authContext and updates status', () => {
    useAuthStore.getState().setFromLoginWithTokens({
      access_token: 'tok-1',
      refresh_token: 'ref-1',
      expires_at: 9999999999,
      email: 'user@test.com',
      name: 'User',
      mfa_enabled: false,
      needs_organization: false,
      selected_organization: { id: 'org-1' },
    } as never);
    expect(authContextMock.setFromLoginWithTokens).toHaveBeenCalledOnce();
    const state = useAuthStore.getState();
    expect(state.accessToken).toBe('tok-1');
    expect(state.organizationId).toBe('org-1');
    expect(state.status).toBe('authenticated');
  });

  it('derives needs_organization status when operator lacks org', () => {
    useAuthStore.getState().setFromLoginWithTokens({
      access_token: 'tok-1',
      expires_at: 9999999999,
      email: 'user@test.com',
      name: 'User',
      mfa_enabled: false,
      needs_organization: true,
    } as never);
    notifyListeners();
    const state = useAuthStore.getState();
    expect(state.status).toBe('needs_organization');
  });

  it('setMfaChallenge sets mfa_required status', () => {
    useAuthStore.getState().setMfaChallenge({
      operatorId: 'op-1',
      email: 'user@test.com',
      name: 'User',
      mfaEnabled: true,
    });
    expect(useAuthStore.getState().status).toBe('mfa_required');
    expect(useAuthStore.getState().mfaChallenge).not.toBeNull();
  });

  it('setOrganization updates organizationId', () => {
    useAuthStore.getState().setOrganization('org-new');
    expect(authContextMock.setOrganization).toHaveBeenCalledWith('org-new');
    expect(useAuthStore.getState().organizationId).toBe('org-new');
  });

  it('setLockout derives locked status', () => {
    useAuthStore.getState().setLockout({ isLocked: true, retryAfter: 60, lockedUntil: 999 });
    expect(useAuthStore.getState().status).toBe('locked');
    expect(useAuthStore.getState().isLocked).toBe(true);
  });

  it('clear resets to unauthenticated', () => {
    useAuthStore.getState().setFromLoginWithTokens({
      access_token: 'tok-1',
      expires_at: 9999999999,
      email: 'u@t.com',
      name: 'U',
      mfa_enabled: false,
      needs_organization: false,
      selected_organization: { id: 'org-1' },
    } as never);
    useAuthStore.getState().clear();
    expect(useAuthStore.getState().status).toBe('unauthenticated');
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().mfaChallenge).toBeNull();
  });

  it('refreshTokens delegates to authContext', async () => {
    await useAuthStore.getState().refreshTokens();
    expect(authContextMock.refreshTokens).toHaveBeenCalledOnce();
  });
});
