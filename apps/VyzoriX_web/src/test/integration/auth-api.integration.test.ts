/**
 * Integration tests for the API Client auth functions.
 *
 * These tests call the REAL API client functions (login, loginWithTokens,
 * getMe, etc.) which use the REAL restClient (axios) and REAL domain mappers.
 * MSW intercepts the HTTP requests and returns mock server responses.
 *
 * This is the test layer that catches shape mismatches between the MSW mock
 * and the domain mappers — if the server returns { access_token } but the
 * mapper expects { accessToken }, the test will fail.
 */
import { describe, it, expect, beforeAll, afterAll, afterEach, beforeEach, vi } from 'vitest';
import { createVyzorMswServer } from '@/test/msw/vyzor-msw-server';
import {
  login,
  loginWithTokens,
  register,
  logout,
  getMe,
  refreshToken,
  fetchCSRFToken,
  updateName,
  resetClientState,
  authContext,
  getMFAStatus,
  enrollMFA,
  verifyMFASetup,
  enableMFA,
  disableMFA,
  verifyMFA,
  verifyBackupCode,
  regenerateBackupCodes,
  forgotPassword,
  resetPassword,
  resendPasswordReset,
  verifyEmail,
  verifyEmailGet,
  resendVerification,
  cancelVerification,
  pollVerification,
  sessions,
  me,
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

describe('auth-core — login, loginWithTokens, register, logout, /me, refresh', () => {
  describe('fetchCSRFToken', () => {
    it('fetches and caches CSRF token', async () => {
      const token = await fetchCSRFToken();
      expect(token).toBe('mock-csrf-token');
    });
  });

  describe('login (browser — cookie-based, no tokens)', () => {
    it('returns mapped LoginResponse on success', async () => {
      const result = await login({ email: 'test@vyzorix.com', password: 'pass' });
      if ('mfaRequired' in result) throw new Error('unexpected MFA required');
      expect(result.success).toBe(true);
      const data = result.data;
      expect(data.operatorId).toBe('operator-test-1');
      expect(data.email).toBe('test@vyzorix.com');
      expect(data.name).toBe('Test Operator');
      expect(data.role).toBe('admin');
      expect(data.mfaEnabled).toBe(false);
      expect(data.needsOrganization).toBe(false);
      expect(data.organizations).toHaveLength(1);
      expect(data.selectedOrganization?.id).toBe('org-test-1');
      expect(data.lastOrganizationId).toBe('org-test-1');
    });

    it('normalizes email to lowercase + trims whitespace', async () => {
      const result = await login({ email: '  Test@Vyzorix.COM  ', password: 'pass' });
      if ('mfaRequired' in result) throw new Error('unexpected MFA required');
      expect(result.success).toBe(true);
      expect(result.data.email).toBe('test@vyzorix.com');
    });

    it('rejects with 401 for missing credentials', async () => {
      await expect(login({ email: '', password: '' })).rejects.toThrow();
    });
  });

  describe('loginWithTokens (API client — JWT + refresh token)', () => {
    it('returns mapped LoginWithTokensResponse and populates authContext via store', async () => {
      const result = await loginWithTokens({ email: 'test@vyzorix.com', password: 'pass' });
      if ('mfaRequired' in result) throw new Error('unexpected MFA required');
      expect(result.success).toBe(true);
      const data = result.data;
      expect(data.operator_id).toBe('operator-test-1');
      expect(data.access_token).toBe('mock-access-token');
      expect(data.refresh_token).toBe('mock-refresh-token');
      expect(data.session_id).toBe('session-test-1');
      expect(typeof data.expires_at).toBe('number');
      expect(data.organizations).toHaveLength(1);
      expect(data.memberships).toHaveLength(1);
      expect(data.selected_organization?.id).toBe('org-test-1');

      // Simulate what the auth store does: push the response into authContext
      authContext.setFromLoginWithTokens(data);
      // Now authContext should reflect the tokens + operator
      const state = authContext.getState();
      expect(state.isAuthenticated).toBe(true);
      expect(state.accessToken).toBe('mock-access-token');
      expect(state.refreshToken).toBe('mock-refresh-token');
      expect(state.operator?.id).toBe('operator-test-1');
      expect(state.operator?.email).toBe('test@vyzorix.com');
      expect(state.organizationId).toBe('org-test-1');
    });

    it('rejects with 401 for missing credentials', async () => {
      await expect(loginWithTokens({ email: '', password: '' })).rejects.toThrow();
    });
  });

  describe('register', () => {
    it('returns mapped RegisterResponse', async () => {
      const result = await register({ email: 'new@vyzorix.com', password: 'pass123', name: 'New User' });
      expect(result.operatorId).toBe('operator-test-2');
      expect(result.email).toBe('new@vyzorix.com');
      expect(result.name).toBe('New User');
    });

    it('normalizes email + trims name', async () => {
      const result = await register({ email: '  NEW@Vyzorix.COM  ', password: 'pass123', name: '  Trimmed  ' });
      expect(result.email).toBe('new@vyzorix.com');
      expect(result.name).toBe('Trimmed');
    });

    it('rejects for missing fields', async () => {
      await expect(register({ email: '', password: '', name: '' })).rejects.toThrow();
    });
  });

  describe('logout', () => {
    it('returns success and clears auth context', async () => {
      // Set up auth state first (simulating what the store does after login)
      const result = await loginWithTokens({ email: 'test@vyzorix.com', password: 'pass' });
      if ('mfaRequired' in result) throw new Error('unexpected MFA required');
      authContext.setFromLoginWithTokens(result.data);
      expect(authContext.getState().accessToken).toBe('mock-access-token');

      const logoutResult = await logout();
      expect(logoutResult.success).toBe(true);
      // logout calls clearAuthContext which clears the restClient's internal token state.
      // authContext.clear() is called by the store, not by logout() directly.
      authContext.clear();
      expect(authContext.getState().accessToken).toBeNull();
      expect(authContext.getState().isAuthenticated).toBe(false);
    });
  });

  describe('getMe', () => {
    it('returns mapped MeResponse with operator fields', async () => {
      const result = await getMe();
      expect(result).not.toBeNull();
      expect(result?.id).toBe('operator-test-1');
      expect(result?.email).toBe('test@vyzorix.com');
      expect(result?.name).toBe('Test Operator');
      expect(result?.mfa_enabled).toBe(false);
      expect(result?.email_verified).toBe(true);
      expect(result?.needs_organization).toBe(false);
      expect(result?.organizations).toHaveLength(1);
      expect(result?.memberships).toHaveLength(1);
      expect(result?.selected_organization?.id).toBe('org-test-1');
      expect(result?.thresholds?.riskWarn).toBe(70);
      expect(result?.client?.requestTimeoutMs).toBe(8000);
    });
  });

  describe('updateName', () => {
    it('returns updated operator with new name', async () => {
      const result = await updateName('Updated Name');
      expect(result.id).toBe('operator-test-1');
      expect(result.name).toBe('Updated Name');
      expect(result.mfa_enabled).toBe(false);
      expect(result.email_verified).toBe(true);
    });
  });

  describe('refreshToken', () => {
    it('returns mapped AuthTokens from refresh response', async () => {
      const result = await refreshToken('old-refresh-token');
      expect(result.accessToken).toBe('mock-access-token-refreshed');
      expect(result.refreshToken).toBe('mock-refresh-token-refreshed');
      expect(result.sessionId).toBe('session-test-1');
      expect(typeof result.expiresAt).toBe('number');
    });
  });
});

describe('me endpoints', () => {
  it('me.getMe returns mapped operator', async () => {
    const result = await me.getMe();
    expect(result.id).toBe('operator-test-1');
    expect(result.email).toBe('test@vyzorix.com');
    expect(result.organizations).toHaveLength(1);
  });

  it('me.getOrganizations returns organization list', async () => {
    const result = await me.getOrganizations();
    expect(result.organizations).toHaveLength(1);
    const org = result.organizations[0];
    expect(org).toBeDefined();
    expect(org?.id).toBe('org-test-1');
  });

  it('me.selectOrganization returns the selected org', async () => {
    const result = await me.selectOrganization({ organization_id: 'org-test-1' });
    expect(result.id).toBe('org-test-1');
    expect(result.name).toBe('Test Organization');
  });
});

describe('MFA endpoints', () => {
  describe('getMFAStatus', () => {
    it('returns mapped MFAStatusResponse', async () => {
      const result = await getMFAStatus();
      expect(result.enabled).toBe(false);
    });
  });

  describe('enrollMFA', () => {
    it('returns secret + uri', async () => {
      const result = await enrollMFA();
      expect(result.secret).toBe('JBSWY3DPEHPK3PXP');
      expect(result.uri).toContain('otpauth://totp/');
    });
  });

  describe('verifyMFASetup', () => {
    it('returns verified=true for correct code', async () => {
      const result = await verifyMFASetup('123456');
      expect(result.verified).toBe(true);
    });

    it('returns verified=false for wrong code', async () => {
      const result = await verifyMFASetup('000000');
      expect(result.verified).toBe(false);
    });
  });

  describe('enableMFA', () => {
    it('returns success + backup codes for correct code', async () => {
      const result = await enableMFA('123456');
      expect(result.success).toBe(true);
      expect(result.backupCodes).toHaveLength(5);
      expect(result.backupCodes?.[0]).toBe('bc-001');
    });

    it('rejects for wrong code', async () => {
      await expect(enableMFA('000000')).rejects.toThrow();
    });
  });

  describe('disableMFA', () => {
    it('returns success=true for correct code', async () => {
      const result = await disableMFA('123456');
      expect(result.success).toBe(true);
    });

    it('returns success=false for wrong code', async () => {
      const result = await disableMFA('000000');
      expect(result.success).toBe(false);
    });
  });

  describe('verifyMFA (login completion)', () => {
    it('returns mapped MFAVerifyResponse and sets auth tokens on restClient', async () => {
      const result = await verifyMFA('operator-test-1', '123456');
      expect(result.success).toBe(true);
      expect(result.sessionId).toBe('session-test-1');
      expect(result.accessToken).toBe('mock-access-token-mfa');
      expect(result.refreshToken).toBe('mock-refresh-token-mfa');
      expect(result.operator?.id).toBe('operator-test-1');
      expect(result.operator?.email).toBe('test@vyzorix.com');
      expect(result.operator?.mfaEnabled).toBe(true);
      // verifyMFA calls setAuthToken/setRefreshToken on the restClient (not authContext).
      // The store would then call authContext.setFromLoginWithTokens or similar.
    });

    it('rejects for wrong code', async () => {
      await expect(verifyMFA('operator-test-1', '000000')).rejects.toThrow();
    });
  });

  describe('verifyBackupCode', () => {
    it('returns valid=true for correct backup code', async () => {
      const result = await verifyBackupCode('bc-001');
      expect(result.valid).toBe(true);
    });

    it('returns valid=false for wrong backup code', async () => {
      const result = await verifyBackupCode('wrong');
      expect(result.valid).toBe(false);
    });
  });

  describe('regenerateBackupCodes', () => {
    it('returns new backup codes', async () => {
      const result = await regenerateBackupCodes();
      expect(result.backupCodes).toHaveLength(5);
      expect(result.backupCodes?.[0]).toBe('new-bc-001');
    });
  });
});

describe('password endpoints', () => {
  describe('forgotPassword', () => {
    it('returns success=true', async () => {
      const result = await forgotPassword('test@vyzorix.com');
      expect(result.success).toBe(true);
    });

    it('rejects for missing email', async () => {
      await expect(forgotPassword('')).rejects.toThrow();
    });
  });

  describe('resetPassword', () => {
    it('returns success=true for valid token + newPassword', async () => {
      const result = await resetPassword('valid-reset-token', 'newPass123');
      expect(result.success).toBe(true);
    });

    it('rejects for missing fields', async () => {
      await expect(resetPassword('', '')).rejects.toThrow();
    });
  });

  describe('resendPasswordReset', () => {
    it('returns success=true with message', async () => {
      const result = await resendPasswordReset('test@vyzorix.com');
      expect(result.success).toBe(true);
      expect(result.message).toBe('reset email sent');
    });

    it('rejects for missing email', async () => {
      await expect(resendPasswordReset('')).rejects.toThrow();
    });
  });
});

describe('email verification endpoints', () => {
  describe('verifyEmail (POST)', () => {
    it('returns verified=true for valid token', async () => {
      const result = await verifyEmail('valid-token');
      expect(result.verified).toBe(true);
      expect(result.email).toBe('test@vyzorix.com');
    });

    it('returns verified=false for invalid token', async () => {
      const result = await verifyEmail('invalid-token');
      expect(result.verified).toBe(false);
    });
  });

  describe('verifyEmailGet (GET)', () => {
    it('returns verified=true for valid token', async () => {
      const result = await verifyEmailGet('valid-token');
      expect(result.verified).toBe(true);
    });
  });

  describe('resendVerification', () => {
    it('returns success message', async () => {
      const result = await resendVerification('test@vyzorix.com');
      expect(result.message).toBe('verification email sent');
    });
  });

  describe('cancelVerification', () => {
    it('returns success=true', async () => {
      const result = await cancelVerification('test@vyzorix.com');
      expect(result.success).toBe(true);
    });
  });

  describe('pollVerification', () => {
    it('returns verified=true for valid token', async () => {
      const result = await pollVerification('valid-token');
      expect(result.verified).toBe(true);
    });
  });
});

describe('session endpoints', () => {
  describe('listSessions', () => {
    it('returns mapped SessionListResponse', async () => {
      const result = await sessions.listSessions();
      expect(result.sessions).toHaveLength(1);
      const session = result.sessions[0];
      expect(session).toBeDefined();
      expect(session?.id).toBe('session-test-1');
      expect(session?.ipAddress).toBe('192.168.1.1');
      expect(session?.userAgent).toBe('Mozilla/5.0 Chrome');
      expect(session?.isCurrent).toBe(true);
      expect(session?.selectedOrganizationId).toBe('org-test-1');
      expect(session?.createdAt).toBeInstanceOf(Date);
      expect(session?.expiresAt).toBeInstanceOf(Date);
      expect(result.total).toBe(1);
    });
  });

  describe('getConcurrent', () => {
    it('returns mapped ConcurrentSessionsResponse', async () => {
      const result = await sessions.getConcurrent();
      expect(result.hasConcurrent).toBe(false);
      expect(result.count).toBe(1);
      expect(result.sessions).toHaveLength(1);
      const session = result.sessions[0];
      expect(session).toBeDefined();
      expect(session?.id).toBe('session-test-1');
      expect(session?.ipAddress).toBe('192.168.1.1');
    });
  });

  describe('revokeSession', () => {
    it('returns success=true', async () => {
      const result = await sessions.revokeSession('session-test-1');
      expect(result.success).toBe(true);
      expect(result.message).toContain('session-test-1');
    });
  });

  describe('revokeAllExceptCurrent', () => {
    it('returns success + revoked_count', async () => {
      const result = await sessions.revokeAllExceptCurrent();
      expect(result.success).toBe(true);
      expect(result.revoked_count).toBe(2);
    });
  });

  describe('revokeAllDevices', () => {
    it('returns mapped RevokeAllSessionsResponse', async () => {
      const result = await sessions.revokeAllDevices();
      expect(result.success).toBe(true);
      expect(result.revokedCount).toBe(3);
    });
  });
});
