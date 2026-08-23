/**
 * Integration tests for the auth presentation hooks.
 *
 * These render the REAL hooks via React Testing Library. The hooks call the
 * REAL `@vyzorix/api-client` (real restClient/axios + domain mappers) and the
 * REAL `authContext` / `useAuthStore`. MSW intercepts the HTTP requests and
 * returns mock server responses mirroring the Go server auth contract
 * (snake_case field names).
 *
 * No vi.mock / vi.hoisted — the real code path runs end-to-end.
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { waitFor, act } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { setupIntegrationTest } from '../helpers/integration-test-setup';
import { useAuthStore } from '@/stores/auth-store';
import {
  useLogin,
  useRegister,
  useLogout,
  useMfaVerify,
  useForgotPassword,
  useResetPassword,
  useResendPasswordReset,
  useAuthCallback,
} from '@/hooks/auth';

const { server } = setupIntegrationTest();

const API_BASE = '/v1/auth';

// resetApiState() clears authContext but not the store's mfaChallenge (which is
// store-local, not mirrored from authContext). Reset the store fully per test.
beforeEach(() => {
  useAuthStore.getState().clear();
});

function rawLoginTokensMfaRequired() {
  return HttpResponse.json({
    requires_mfa: true,
    operator_id: 'operator-1',
    email: 'test@vyzorix.com',
    name: 'Test Operator',
    mfa_enabled: true,
  });
}

describe('useLogin', () => {
  it('ingests tokens into authContext on success and clears mfaChallenge', async () => {
    const { result } = renderHookWithQueryClient(() => useLogin());

    await act(async () => {
      result.current.mutate({ email: 'test@vyzorix.com', password: 'password123' });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.accessToken).toBe('mock-access-token');
    expect(state.operator?.id).toBe('operator-test-1');
    expect(state.mfaChallenge).toBeNull();
    expect(state.status).toBe('authenticated');
  });

  it('stages an MfaChallenge (status → mfa_required) when the server requires MFA', async () => {
    server.use(http.post(`${API_BASE}/login/tokens`, () => rawLoginTokensMfaRequired()));

    const { result } = renderHookWithQueryClient(() => useLogin());

    await act(async () => {
      result.current.mutate({ email: 'test@vyzorix.com', password: 'password123' });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.mfaChallenge?.operatorId).toBe('operator-1');
    expect(state.mfaChallenge?.mfaEnabled).toBe(true);
    expect(state.status).toBe('mfa_required');
  });

  it('surfaces invalid-credentials errors without authenticating', async () => {
    server.use(
      http.post(`${API_BASE}/login/tokens`, () =>
        HttpResponse.json({ error: 'invalid email or password' }, { status: 401 }),
      ),
    );

    const { result } = renderHookWithQueryClient(() => useLogin());

    await act(async () => {
      result.current.mutate({ email: 'bad@vyzorix.com', password: 'wrong' });
    });
    await waitFor(() => expect(result.current.isError).toBe(true));

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.mfaChallenge).toBeNull();
  });
});

describe('useMfaVerify', () => {
  it('verifies MFA, ingests tokens, and fetches /me to reconcile operator', async () => {
    // Login first to stage the challenge (operatorId).
    server.use(http.post(`${API_BASE}/login/tokens`, () => rawLoginTokensMfaRequired()));
    const login = renderHookWithQueryClient(() => useLogin());
    await act(async () => {
      login.result.current.mutate({ email: 'test@vyzorix.com', password: 'password123' });
    });
    await waitFor(() => expect(login.result.current.isSuccess).toBe(true));
    expect(useAuthStore.getState().mfaChallenge?.operatorId).toBe('operator-1');

    const { result } = renderHookWithQueryClient(() => useMfaVerify());

    await act(async () => {
      result.current.mutate({ operatorId: 'operator-1', code: '123456' });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.accessToken).toBe('mock-access-token-mfa');
    // /me reconciliation populated organizations + selected org.
    expect(state.operator?.organizations).toHaveLength(1);
    expect(state.operator?.selected_organization?.id).toBe('org-test-1');
    expect(state.mfaChallenge).toBeNull();
    expect(state.status).toBe('authenticated');
  });

  it('surfaces an invalid MFA code without authenticating', async () => {
    server.use(
      http.post(`${API_BASE}/mfa/verify`, () =>
        HttpResponse.json({ error: 'invalid code' }, { status: 400 }),
      ),
    );

    const { result } = renderHookWithQueryClient(() => useMfaVerify());

    await act(async () => {
      result.current.mutate({ operatorId: 'operator-1', code: '000000' });
    });
    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});

describe('useRegister', () => {
  it('registers and returns the operator id (no auto-login)', async () => {
    const { result } = renderHookWithQueryClient(() => useRegister());

    await act(async () => {
      result.current.mutate({ email: 'new@vyzorix.com', password: 'password123', name: 'New User' });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.operator_id).toBe('operator-test-2');
    expect(result.current.data?.email).toBe('new@vyzorix.com');
    // Registration must not establish a session.
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it('surfaces a duplicate-email conflict', async () => {
    server.use(
      http.post(`${API_BASE}/register`, () =>
        HttpResponse.json({ error: 'an account with this email already exists' }, { status: 409 }),
      ),
    );

    const { result } = renderHookWithQueryClient(() => useRegister());

    await act(async () => {
      result.current.mutate({ email: 'dup@vyzorix.com', password: 'password123', name: 'Dup' });
    });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useLogout', () => {
  it('clears local auth state even when the server logout succeeds', async () => {
    // Seed an authenticated session.
    const login = renderHookWithQueryClient(() => useLogin());
    await act(async () => {
      login.result.current.mutate({ email: 'test@vyzorix.com', password: 'password123' });
    });
    await waitFor(() => expect(login.result.current.isSuccess).toBe(true));
    expect(useAuthStore.getState().isAuthenticated).toBe(true);

    const { result } = renderHookWithQueryClient(() => useLogout());
    await act(async () => {
      result.current.mutate();
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.operator).toBeNull();
    expect(state.accessToken).toBeNull();
  });

  it('clears local auth state even when the server logout request fails', async () => {
    const login = renderHookWithQueryClient(() => useLogin());
    await act(async () => {
      login.result.current.mutate({ email: 'test@vyzorix.com', password: 'password123' });
    });
    await waitFor(() => expect(login.result.current.isSuccess).toBe(true));

    server.use(
      http.post(`${API_BASE}/logout`, () =>
        HttpResponse.json({ error: 'server down' }, { status: 500 }),
      ),
    );

    const { result } = renderHookWithQueryClient(() => useLogout());
    await act(async () => {
      result.current.mutate();
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    // Must still be logged out despite the failed API call.
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});

describe('useForgotPassword / useResetPassword / useResendPasswordReset', () => {
  it('forgot-password returns success', async () => {
    const { result } = renderHookWithQueryClient(() => useForgotPassword());
    await act(async () => {
      result.current.mutate({ email: 'test@vyzorix.com' });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.message).toBeTruthy();
  });

  it('reset-password returns success with a valid token', async () => {
    const { result } = renderHookWithQueryClient(() => useResetPassword());
    await act(async () => {
      result.current.mutate({ token: 'valid-reset-token', newPassword: 'newpass123' });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.message).toBeTruthy();
  });

  it('reset-password surfaces an invalid token', async () => {
    server.use(
      http.post(`${API_BASE}/reset-password`, () =>
        HttpResponse.json({ error: 'invalid token' }, { status: 400 }),
      ),
    );
    const { result } = renderHookWithQueryClient(() => useResetPassword());
    await act(async () => {
      result.current.mutate({ token: 'bad', newPassword: 'newpass123' });
    });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('resend-password-reset returns success', async () => {
    const { result } = renderHookWithQueryClient(() => useResendPasswordReset());
    await act(async () => {
      result.current.mutate({ email: 'test@vyzorix.com' });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.message).toBeTruthy();
  });
});

describe('useAuthCallback', () => {
  it('parses a successful callback URL and hydrates auth via /me', async () => {
    const successUrl = 'https://app.vyzorix.dev/auth/callback?oauth=success&new=false&provider=google';

    const { result } = renderHookWithQueryClient(() =>
      useAuthCallback({ url: successUrl, fetchMe: true }),
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.result?.success).toBe(true);
    expect(result.current.error).toBeNull();

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.operator?.id).toBe('operator-test-1');
  });

  it('surfaces an error for a failed callback URL', async () => {
    const errorUrl = 'https://app.vyzorix.dev/auth/callback?oauth=error&code=login_failed&message=OAuth%20authentication%20failed&provider=google';

    const { result } = renderHookWithQueryClient(() =>
      useAuthCallback({ url: errorUrl, fetchMe: false }),
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.result?.success).toBe(false);
    expect(result.current.error).toBe('OAuth authentication failed');
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it('reports an error when no callback URL is available', async () => {
    const { result } = renderHookWithQueryClient(() =>
      useAuthCallback({ url: '', fetchMe: false }),
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.error).toBe('No callback URL');
  });
});
