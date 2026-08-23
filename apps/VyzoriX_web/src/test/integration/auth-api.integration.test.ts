/**
 * Integration tests for the generated auth SDK surface (orval).
 *
 * These tests call the REAL generated endpoint functions (getAuth(), getMfa(),
 * getSessions()) which go through the REAL restClient (axios instance with
 * interceptors, CSRF preflight, org headers). MSW intercepts the HTTP requests
 * and returns mock server responses mirroring the Go server contract
 * (snake_case wire fields).
 */
import { describe, it, expect, beforeAll, afterAll, afterEach, beforeEach, vi } from 'vitest';
import { createVyzorMswServer } from '@/test/msw/vyzor-msw-server';
import {
  getAuth,
  getMfa,
  getSettings,
  getSessions,
  getInvitations,
  fetchAndSetCSRFToken,
  getCSRFToken,
  resetClientState,
  authContext,
} from '@vyzorix/api-client';

const server = createVyzorMswServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

beforeEach(() => {
  vi.stubEnv('VITE_API_URL', '');
  vi.stubEnv('VITE_REST_WITH_CREDENTIALS', 'false');
  resetClientState();
  authContext.clear();
});

afterEach(() => {
  resetClientState();
  authContext.clear();
});

describe('auth core (generated SDK)', () => {
  describe('CSRF preflight', () => {
    it('fetches and caches the CSRF token', async () => {
      const token = await fetchAndSetCSRFToken();
      expect(token).toBe('mock-csrf-token');
      expect(getCSRFToken()).toBe('mock-csrf-token');
    });
  });

  describe('login (browser — cookie-based)', () => {
    it('returns the operator payload on success', async () => {
      const result = await getAuth().postAuthLogin({
        email: 'test@vyzorix.com',
        password: 'password123',
      });
      expect(result.operator_id).toBe('operator-test-1');
      expect(result.email).toBe('test@vyzorix.com');
      expect(result.name).toBe('Test Operator');
      expect(result.organizations?.[0]?.id).toBe('org-test-1');
    });

    it('rejects with 401 for missing credentials', async () => {
      await expect(
        getAuth().postAuthLogin({ email: '', password: '' }),
      ).rejects.toThrow();
    });
  });

  describe('loginWithTokens (JWT + refresh token)', () => {
    it('returns tokens and operator fields', async () => {
      const result = await getAuth().postAuthLoginTokens({
        email: 'test@vyzorix.com',
        password: 'password123',
      });
      expect(result.access_token).toBe('mock-access-token');
      expect(result.refresh_token).toBe('mock-refresh-token');
      expect(result.operator_id).toBe('operator-test-1');
      expect(result.selected_organization?.id).toBe('org-test-1');
    });

    it('rejects with 401 for missing credentials', async () => {
      await expect(
        getAuth().postAuthLoginTokens({ email: '', password: '' }),
      ).rejects.toThrow();
    });
  });

  describe('register', () => {
    it('returns the created operator (no auto-login)', async () => {
      const result = await getAuth().postAuthRegister({
        email: 'new@vyzorix.com',
        password: 'password123',
        name: 'New User',
      });
      expect(result.operator_id).toBe('operator-test-2');
      expect(result.email).toBe('new@vyzorix.com');
    });

    it('rejects for missing fields', async () => {
      await expect(
        getAuth().postAuthRegister({ email: '', password: '', name: '' }),
      ).rejects.toThrow();
    });
  });

  describe('logout', () => {
    it('resolves successfully', async () => {
      await expect(getAuth().postAuthLogout()).resolves.toBeDefined();
    });
  });

  describe('getMe', () => {
    it('returns the operator profile', async () => {
      const result = await getAuth().getAuthMe();
      expect(result.id).toBe('operator-test-1');
      expect(result.email).toBe('test@vyzorix.com');
      expect(result.email_verified).toBe(true);
      expect(result.organizations?.[0]?.id).toBe('org-test-1');
    });
  });

  describe('updateName', () => {
    it('patches the operator name', async () => {
      const result = await getSettings().patchAuthMe({ name: 'Renamed Operator' });
      expect(result).toBeDefined();
    });
  });

  describe('refreshToken', () => {
    it('returns refreshed tokens', async () => {
      const result = await getAuth().postAuthRefresh({
        refresh_token: 'mock-refresh-token',
      });
      expect(result.access_token).toBe('mock-access-token-refreshed');
      expect(result.refresh_token).toBe('mock-refresh-token-refreshed');
    });

    it('rejects without a refresh token', async () => {
      await expect(getAuth().postAuthRefresh({})).rejects.toThrow();
    });
  });
});

describe('organization context (generated SDK)', () => {
  it('lists my organizations', async () => {
    const result = await getAuth().getAuthOrganizations();
    expect(result.organizations?.[0]?.id).toBe('org-test-1');
  });

  it('selects an organization', async () => {
    const result = await getAuth().postAuthOrganizationsSelect({
      organization_id: 'org-test-1',
    });
    expect(result).toBeDefined();
  });

  it('lists invitations pending for the current operator', async () => {
    const result = await getInvitations().getMeInvitations();
    expect(result).toBeDefined();
  });
});

describe('MFA endpoints (generated SDK)', () => {
  it('getAuthMfaStatus returns the MFA status', async () => {
    const result = await getMfa().getAuthMfaStatus();
    expect(result.mfa_enabled).toBe(false);
  });

  it('enroll returns secret + uri', async () => {
    const result = await getMfa().postAuthMfaEnroll();
    expect(result).toBeDefined();
  });

  it('verify-setup verifies the code', async () => {
    const result = await getMfa().postAuthMfaVerifySetup({ code: '123456' });
    expect(result).toBeDefined();
  });

  it('enable returns backup codes for a correct code', async () => {
    const result = await getMfa().postAuthMfaEnable({ code: '123456' });
    expect(result.success).toBe(true);
    expect(result.backup_codes).toHaveLength(5);
  });

  it('enable rejects a wrong code', async () => {
    await expect(getMfa().postAuthMfaEnable({ code: '000000' })).rejects.toThrow();
  });

  it('disable resolves for a correct code', async () => {
    const result = await getMfa().postAuthMfaDisable({ code: '123456' });
    expect(result).toBeDefined();
  });

  it('verify completes login and returns tokens', async () => {
    const result = await getMfa().postAuthMfaVerify({
      operator_id: 'operator-test-1',
      code: '123456',
    });
    expect(result.success).toBe(true);
    expect(result.access_token).toBe('mock-access-token-mfa');
    expect(result.operator?.id).toBe('operator-test-1');
  });

  it('verify rejects a wrong code', async () => {
    await expect(
      getMfa().postAuthMfaVerify({ operator_id: 'operator-test-1', code: '000000' }),
    ).rejects.toThrow();
  });

  it('verify-backup validates a backup code', async () => {
    const result = await getMfa().postAuthMfaVerifyBackup({ code: 'bc-001' });
    expect(result).toBeDefined();
  });

  it('regenerate-backup-codes returns new codes', async () => {
    const result = await getMfa().postAuthMfaRegenerateBackupCodes();
    expect(result).toBeDefined();
  });
});

describe('password endpoints (generated SDK)', () => {
  it('forgot-password resolves', async () => {
    await expect(
      getAuth().postAuthForgotPassword({ email: 'test@vyzorix.com' }),
    ).resolves.toBeDefined();
  });

  it('forgot-password rejects for missing email', async () => {
    await expect(getAuth().postAuthForgotPassword({})).rejects.toThrow();
  });

  it('reset-password resolves for a valid token', async () => {
    await expect(
      getAuth().postAuthResetPassword({ token: 'valid-reset-token', newPassword: 'newpass123' }),
    ).resolves.toBeDefined();
  });

  it('resend-password-reset resolves', async () => {
    await expect(
      getAuth().postAuthResendPasswordReset({ email: 'test@vyzorix.com' }),
    ).resolves.toBeDefined();
  });
});

describe('email verification endpoints (generated SDK)', () => {
  it('verify-email (POST) verifies a valid token', async () => {
    const result = await getAuth().postAuthVerifyEmail({ token: 'valid-token' });
    expect(result.verified).toBe(true);
  });

  it('verify-email (POST) rejects an invalid token', async () => {
    const result = await getAuth().postAuthVerifyEmail({ token: 'bogus' });
    expect(result.verified).toBe(false);
  });

  it('verify-email (GET) verifies a valid token', async () => {
    const result = await getAuth().getAuthVerifyEmail({ token: 'valid-token' });
    expect(result.verified).toBe(true);
  });

  it('resend-verification resolves', async () => {
    await expect(
      getAuth().postAuthResendVerification({ email: 'test@vyzorix.com' }),
    ).resolves.toBeDefined();
  });

  it('cancel-verification resolves', async () => {
    await expect(
      getAuth().postAuthCancelVerification({ email: 'test@vyzorix.com' }),
    ).resolves.toBeDefined();
  });

  it('poll-verification reports verification state', async () => {
    const result = await getAuth().getAuthPollVerification({ token: 'valid-token' });
    expect(result).toBeDefined();
  });
});

describe('session endpoints (generated SDK)', () => {
  it('lists sessions', async () => {
    const result = await getSessions().getAuthSessions();
    expect(result.sessions).toHaveLength(1);
    expect(result.total).toBe(1);
  });

  it('checks concurrent sessions', async () => {
    const result = await getSessions().getAuthSessionsConcurrent();
    expect(result.has_concurrent).toBe(false);
    expect(result.count).toBe(1);
  });

  it('revokes a session', async () => {
    await expect(getSessions().deleteAuthSessionsId('session-test-1')).resolves.toBeDefined();
  });

  it('revokes all other sessions', async () => {
    await expect(getSessions().deleteAuthSessions()).resolves.toBeDefined();
  });

  it('revokes all device sessions', async () => {
    await expect(getSessions().postAuthSessionsRevokeAll()).resolves.toBeDefined();
  });
});
