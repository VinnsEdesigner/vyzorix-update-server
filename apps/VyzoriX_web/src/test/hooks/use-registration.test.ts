import { describe, it, expect, beforeEach, vi } from 'vitest';
import { waitFor } from '@testing-library/react';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { useAuthStore } from '@/stores/auth-store';

const {
  registrationMock,
  queryInboxEntriesMock,
  queryInboxEntryMock,
  mutateAckInboxMock,
  mutateDeregisterDeviceMock,
  authContextStub,
} = vi.hoisted(() => ({
  registrationMock: {
    getInbox: vi.fn(),
    getInboxEntry: vi.fn(),
    createInboxRequest: vi.fn(),
    confirmDevice: vi.fn(),
    acknowledgeInbox: vi.fn(),
    dismissInbox: vi.fn(),
    resendInboxApproval: vi.fn(),
    getDevices: vi.fn(),
    getDevice: vi.fn(),
    deregisterDevice: vi.fn(),
  },
  queryInboxEntriesMock: vi.fn(),
  queryInboxEntryMock: vi.fn(),
  mutateAckInboxMock: vi.fn(),
  mutateDeregisterDeviceMock: vi.fn(),
  authContextStub: {
    getState: vi.fn(() => ({
      isAuthenticated: false,
      operator: null,
      organizationId: null,
      accessToken: null,
      refreshToken: null,
      tokenExpiresAt: null,
    })),
    getLockoutState: vi.fn(() => ({ isLocked: false, retryAfter: 0, lockedUntil: 0 })),
    onChange: vi.fn(() => () => {}),
    setOrganization: vi.fn(),
    setAccessToken: vi.fn(),
    setFromLoginWithTokens: vi.fn(),
    setFromMeResponse: vi.fn(),
    refreshTokens: vi.fn(async () => {}),
    setLockout: vi.fn(),
    clear: vi.fn(),
  },
}));

vi.mock('@vyzorix/api-client', () => ({
  registration: registrationMock,
  queryInboxEntries: (...args: unknown[]) => queryInboxEntriesMock(...args),
  queryInboxEntry: (...args: unknown[]) => queryInboxEntryMock(...args),
  mutateAckInbox: (...args: unknown[]) => mutateAckInboxMock(...args),
  mutateDeregisterDevice: (...args: unknown[]) => mutateDeregisterDeviceMock(...args),
  authContext: authContextStub,
  getCurrentOrganizationId: vi.fn(() => null),
  initConnectivityMonitor: vi.fn(() => ({
    subscribe: vi.fn(() => () => {}),
    getState: vi.fn(() => ({
      isOnline: true,
      wasOnline: true,
      lastChecked: 0,
      effectiveType: '4g',
      downlink: 10,
      rtt: 50,
    })),
    getQueueSize: vi.fn(() => 0),
    getQueuedRequests: vi.fn(() => []),
    checkConnectivity: vi.fn(async () => true),
    flushQueue: vi.fn(async () => {}),
    clearQueue: vi.fn(),
  })),
  getConnectivityMonitor: vi.fn(() => ({
    getQueueSize: vi.fn(() => 0),
    getQueuedRequests: vi.fn(() => []),
  })),
}));

const {
  useInbox,
  useInboxEntry,
  useAcknowledgeInbox,
  useCreateInboxRequest,
  useConfirmDevice,
  useDismissInbox,
  useResendInboxApproval,
  useRegisteredDevices,
  useRegisteredDevice,
  useDeregisterRegisteredDevice,
  useRegistrationStatus,
  useRegisterDevice,
} = await import('@/hooks/registration');

const INBOX_RESULT = {
  requests: [{ imei: '123', status: 'pending', createdAt: new Date('2024-01-01') }],
  pagination: { page: 1, limit: 20, total: 1, totalPages: 1 },
};

const INBOX_ENTRY = {
  id: 'e1',
  imei: '123',
  deviceName: 'Dev',
  deviceClass: 'phone',
  model: 'X',
  manufacturer: 'Acme',
  osVersion: '14',
  appVersion: '1.0',
  fcmToken: 'tok',
  firebaseInstallId: 'fid',
  status: 'pending',
  acknowledgedAt: null,
  approvingAt: null,
  approvedAt: null,
  rejectedAt: null,
  notes: null,
  operatorId: null,
  createdAt: new Date('2024-01-01'),
};

describe('useInbox', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: null });
  });

  it('is disabled when no organization is selected', () => {
    const { result } = renderHookWithQueryClient(() => useInbox());
    expect(result.current.fetchStatus).toBe('idle');
    expect(registrationMock.getInbox).not.toHaveBeenCalled();
  });

  it('fetches inbox via REST when organization is set', async () => {
    useAuthStore.setState({ organizationId: 'org-1' });
    registrationMock.getInbox.mockResolvedValue(INBOX_RESULT);
    const { result } = renderHookWithQueryClient(() => useInbox({ status: 'pending' }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.getInbox).toHaveBeenCalledWith({ status: 'pending', organizationId: 'org-1' });
    expect(queryInboxEntriesMock).not.toHaveBeenCalled();
  });

  it('falls back to GraphQL when REST rejects', async () => {
    useAuthStore.setState({ organizationId: 'org-1' });
    registrationMock.getInbox.mockRejectedValue(new Error('REST down'));
    queryInboxEntriesMock.mockResolvedValue({
      inbox: {
        requests: [{ imei: '123', status: 'pending', createdAt: 1704067200 }],
        pagination: { page: 1, limit: 20, total: 1, totalPages: 1 },
      },
    });
    const { result } = renderHookWithQueryClient(() => useInbox());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryInboxEntriesMock).toHaveBeenCalledWith({
      organizationId: 'org-1',
      status: undefined,
      page: undefined,
      limit: undefined,
    });
    expect(result.current.data!.requests).toHaveLength(1);
    expect(result.current.data!.requests[0]!.imei).toBe('123');
  });
});

describe('useInboxEntry', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useInboxEntry(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches entry via REST', async () => {
    registrationMock.getInboxEntry.mockResolvedValue(INBOX_ENTRY);
    const { result } = renderHookWithQueryClient(() => useInboxEntry('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.getInboxEntry).toHaveBeenCalledWith('123', 'org-1');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    registrationMock.getInboxEntry.mockRejectedValue(new Error('REST down'));
    queryInboxEntryMock.mockResolvedValue({ inboxEntry: { imei: '123', status: 'pending', createdAt: 1704067200 } });
    const { result } = renderHookWithQueryClient(() => useInboxEntry('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryInboxEntryMock).toHaveBeenCalledWith({ organizationId: 'org-1', imei: '123' });
    expect(result.current.data?.imei).toBe('123');
  });

  it('is disabled (idle) when no org is selected, even if REST would reject', () => {
    useAuthStore.setState({ organizationId: null });
    registrationMock.getInboxEntry.mockRejectedValue(new Error('no org'));
    const { result } = renderHookWithQueryClient(() => useInboxEntry('123'));
    expect(result.current.fetchStatus).toBe('idle');
    expect(registrationMock.getInboxEntry).not.toHaveBeenCalled();
    expect(queryInboxEntryMock).not.toHaveBeenCalled();
  });
});

describe('useAcknowledgeInbox', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls acknowledgeInbox via REST', async () => {
    registrationMock.acknowledgeInbox.mockResolvedValue({ id: '1', imei: '123', status: 'approved' });
    const { result } = renderHookWithQueryClient(() => useAcknowledgeInbox());
    result.current.mutate({ imei: '123', action: 'approve' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.acknowledgeInbox).toHaveBeenCalledWith('123', 'approve', undefined, 'org-1');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    registrationMock.acknowledgeInbox.mockRejectedValue(new Error('REST down'));
    mutateAckInboxMock.mockResolvedValue({ ackInbox: { id: '1', imei: '123', status: 'approved' } });
    const { result } = renderHookWithQueryClient(() => useAcknowledgeInbox());
    result.current.mutate({ imei: '123', action: 'reject', notes: 'no' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mutateAckInboxMock).toHaveBeenCalledWith({ imei: '123', action: 'reject', notes: 'no' });
  });
});

describe('useCreateInboxRequest', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls createInboxRequest with org', async () => {
    registrationMock.createInboxRequest.mockResolvedValue({ id: '1', imei: '123', status: 'pending', createdAt: new Date() });
    const { result } = renderHookWithQueryClient(() => useCreateInboxRequest());
    const request = { imei: '123', fcmToken: 't', firebaseInstallId: 'f' };
    result.current.mutate(request);
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.createInboxRequest).toHaveBeenCalledWith(request, 'org-1');
  });
});

describe('useConfirmDevice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls confirmDevice with imei, commandSecret and org', async () => {
    registrationMock.confirmDevice.mockResolvedValue({ deviceId: 'd1', confirmed: true });
    const { result } = renderHookWithQueryClient(() => useConfirmDevice());
    result.current.mutate({ imei: '123', commandSecret: 'secret' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.confirmDevice).toHaveBeenCalledWith('123', 'secret', 'org-1');
  });
});

describe('useDismissInbox', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls dismissInbox with org', async () => {
    registrationMock.dismissInbox.mockResolvedValue({ status: 'pending' });
    const { result } = renderHookWithQueryClient(() => useDismissInbox());
    result.current.mutate('123');
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.dismissInbox).toHaveBeenCalledWith('123', 'org-1');
  });
});

describe('useResendInboxApproval', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls resendInboxApproval with org', async () => {
    registrationMock.resendInboxApproval.mockResolvedValue({ success: true, message: 'ok' });
    const { result } = renderHookWithQueryClient(() => useResendInboxApproval());
    result.current.mutate('123');
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.resendInboxApproval).toHaveBeenCalledWith('123', 'org-1');
  });
});

describe('useRegisteredDevices', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('fetches devices via REST', async () => {
    registrationMock.getDevices.mockResolvedValue({
      devices: [{ imei: '123', online: true }],
      pagination: { page: 1, limit: 20, total: 1, totalPages: 1 },
    });
    const { result } = renderHookWithQueryClient(() => useRegisteredDevices({ status: 'online' }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.getDevices).toHaveBeenCalledWith({ status: 'online', organizationId: 'org-1' });
  });

  it('errors when REST rejects (no GraphQL fallback for devices)', async () => {
    registrationMock.getDevices.mockRejectedValue(new Error('REST down'));
    const { result } = renderHookWithQueryClient(() => useRegisteredDevices());
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(queryInboxEntriesMock).not.toHaveBeenCalled();
  });
});

describe('useRegisteredDevice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useRegisteredDevice(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches device via REST', async () => {
    registrationMock.getDevice.mockResolvedValue({ imei: '123', online: true });
    const { result } = renderHookWithQueryClient(() => useRegisteredDevice('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.getDevice).toHaveBeenCalledWith('123', 'org-1');
  });
});

describe('useDeregisterRegisteredDevice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls deregisterDevice via REST', async () => {
    registrationMock.deregisterDevice.mockResolvedValue({ imei: '123', status: 'deregistered' });
    const { result } = renderHookWithQueryClient(() => useDeregisterRegisteredDevice());
    result.current.mutate({ imei: '123' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.deregisterDevice).toHaveBeenCalledWith('123', 'org-1');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    registrationMock.deregisterDevice.mockRejectedValue(new Error('REST down'));
    mutateDeregisterDeviceMock.mockResolvedValue({ deregisterDevice: { imei: '123', status: 'deregistered' } });
    const { result } = renderHookWithQueryClient(() => useDeregisterRegisteredDevice());
    result.current.mutate({ imei: '123', hard: true });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mutateDeregisterDeviceMock).toHaveBeenCalledWith({ imei: '123', hard: true });
  });
});

describe('useRegistrationStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useRegistrationStatus(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('returns status from the REST inbox entry', async () => {
    registrationMock.getInboxEntry.mockResolvedValue(INBOX_ENTRY);
    const { result } = renderHookWithQueryClient(() => useRegistrationStatus('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ imei: '123', status: 'pending', entry: INBOX_ENTRY });
  });

  it('returns null when no entry exists', async () => {
    registrationMock.getInboxEntry.mockResolvedValue(null);
    const { result } = renderHookWithQueryClient(() => useRegistrationStatus('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeNull();
  });
});

describe('useRegisterDevice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls createInboxRequest with org', async () => {
    registrationMock.createInboxRequest.mockResolvedValue({ id: '1', imei: '123', status: 'pending', createdAt: new Date() });
    const { result } = renderHookWithQueryClient(() => useRegisterDevice());
    const request = { imei: '123', fcmToken: 't', firebaseInstallId: 'f' };
    result.current.mutate(request);
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(registrationMock.createInboxRequest).toHaveBeenCalledWith(request, 'org-1');
  });
});
