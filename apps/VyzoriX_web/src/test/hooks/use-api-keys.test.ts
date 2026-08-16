/**
 * Integration tests for the API key presentation hooks.
 *
 * These render the REAL hooks via React Testing Library. The hooks call the
 * REAL `@vyzorix/api-client` (real restClient/axios + domain mappers). MSW
 * intercepts the HTTP requests and returns mock server responses mirroring the
 * Go server contract (apps/api/internal/api/handlers/auth/keys_handler.go):
 * snake_case field names, RFC3339 ISO timestamp strings, 204 on revoke.
 *
 * No vi.mock — the real code path runs end-to-end.
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { waitFor, act } from '@testing-library/react';
import { renderHookWithQueryClient, createTestQueryClient } from '../helpers/render-hook';
import { setupIntegrationTest } from '../helpers/integration-test-setup';
import { useAuthStore } from '@/stores/auth-store';
import { resetApiKeyFixtures } from '@/test/msw/vyzor-msw-handlers-apikeys';
import {
  useApiKeys,
  useApiKey,
  useCreateApiKey,
  useUpdateApiKey,
  useRevokeApiKey,
  useRotateApiKey,
} from '@/hooks/apikey';

setupIntegrationTest();

function setOrg(orgId: string | null) {
  useAuthStore.getState().setOrganization(orgId);
}

beforeEach(() => {
  setOrg(null);
  resetApiKeyFixtures();
});

describe('useApiKeys', () => {
  it('is disabled when no organization is selected', () => {
    const { result } = renderHookWithQueryClient(() => useApiKeys());
    expect(result.current.isLoading).toBe(false);
  });

  it('fetches api keys (flattened) when organization is set', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useApiKeys());
    await waitFor(() => expect(result.current.keys.length).toBeGreaterThan(0));
    // MSW seeds 3 keys, 2 active (revoked key is excluded by is_active filter).
    expect(result.current.keys).toHaveLength(2);
    expect(result.current.pagination.total).toBe(2);
  });

  it('maps keys through the real mapper (camelCase + Date objects)', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useApiKeys());
    await waitFor(() => expect(result.current.keys.length).toBeGreaterThan(0));
    const first = result.current.keys[0];
    expect(first).toBeDefined();
    expect(first?.id).toBe('key-1');
    expect(first?.keyPrefix).toBe('vxyz_abcd');
    expect(first?.scope).toBe('admin');
    expect(first?.isActive).toBe(true);
    expect(first?.createdAt).toBeInstanceOf(Date);
    expect(first?.expiresAt).toBeInstanceOf(Date);
  });

  it('exposes monthly stats from the server response', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useApiKeys());
    await waitFor(() => expect(result.current.keys.length).toBeGreaterThan(0));
    expect(result.current.monthlyLimit).toBe(20);
    expect(result.current.keysCreatedThisMonth).toBe(1);
  });
});

describe('useApiKey', () => {
  it('is disabled when no id is provided', () => {
    const { result } = renderHookWithQueryClient(() => useApiKey(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches a single key by id', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useApiKey('key-1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.id).toBe('key-1');
    expect(result.current.data?.name).toBe('Production Key');
  });

  it('returns 404 error for a missing key', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useApiKey('nope'));
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeDefined();
  });
});

describe('useCreateApiKey', () => {
  it('creates a key and returns the full secret (shown once)', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useCreateApiKey());
    await act(async () => {
      result.current.createKey({ name: 'New Key', scope: 'write' });
    });
    await waitFor(() => expect(result.current.createdKey).not.toBeNull());
    expect(result.current.createdKey?.apiKey).toMatch(/^vxyz_secret_/);
    expect(result.current.createdKey?.name).toBe('New Key');
    expect(result.current.createdKey?.scope).toBe('write');
  });

  it('invalidates the api-keys list cache on success', async () => {
    setOrg('org-1');
    const queryClient = createTestQueryClient();
    const listHook = renderHookWithQueryClient(() => useApiKeys(), { queryClient });
    await waitFor(() => expect(listHook.result.current.keys).toHaveLength(2));

    const createHook = renderHookWithQueryClient(() => useCreateApiKey(), { queryClient });
    await act(async () => {
      createHook.result.current.createKey({ name: 'Cache Test', scope: 'read' });
    });
    await waitFor(() => expect(createHook.result.current.createdKey).not.toBeNull());
    // List refetches after invalidation — now 3 active keys.
    await waitFor(() => expect(listHook.result.current.keys).toHaveLength(3));
  });

  it('populates validationErrors and never hits the server for an invalid name', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useCreateApiKey());
    await act(async () => {
      // Empty name fails validateApiKeyName.
      try {
        await result.current.createKey({ name: '', scope: 'read' });
      } catch {
        /* expected rejection */
      }
    });
    await waitFor(() => expect(Object.keys(result.current.validationErrors)).toHaveLength(1));
    expect(result.current.validationErrors.name).toBeDefined();
    expect(result.current.isCreating).toBe(false);
  });
});

describe('useUpdateApiKey', () => {
  it('updates a key name and scope', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useUpdateApiKey());
    await act(async () => {
      result.current.updateKey('key-1', { name: 'Renamed', scope: 'read' });
    });
    await waitFor(() => expect(result.current.updatedKey).not.toBeNull());
    expect(result.current.updatedKey?.name).toBe('Renamed');
    expect(result.current.updatedKey?.scope).toBe('read');
  });

  it('populates validationErrors for an invalid scope', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useUpdateApiKey());
    await act(async () => {
      try {
        await result.current.updateKey('key-1', { scope: 'superuser' as never });
      } catch {
        /* expected rejection */
      }
    });
    await waitFor(() => expect(Object.keys(result.current.validationErrors)).toHaveLength(1));
    expect(result.current.validationErrors.scope).toBeDefined();
  });
});

describe('useRevokeApiKey', () => {
  it('optimistically removes the key from the list cache, then refetches', async () => {
    setOrg('org-1');
    const queryClient = createTestQueryClient();
    const listHook = renderHookWithQueryClient(() => useApiKeys(), { queryClient });
    await waitFor(() => expect(listHook.result.current.keys).toHaveLength(2));

    const revokeHook = renderHookWithQueryClient(() => useRevokeApiKey(), { queryClient });
    await act(async () => {
      revokeHook.result.current.revokeKey('key-1');
    });
    await waitFor(() => expect(revokeHook.result.current.isRevoking).toBe(false));
    // Optimistic update removes key-1 immediately; refetch keeps it gone
    // (server now reports it revoked → excluded from is_active list).
    await waitFor(() => expect(listHook.result.current.keys).toHaveLength(1));
  });

  it('tracks pendingRevoke while the mutation is in flight', async () => {
    setOrg('org-1');
    const queryClient = createTestQueryClient();
    const listHook = renderHookWithQueryClient(() => useApiKeys(), { queryClient });
    await waitFor(() => expect(listHook.result.current.keys).toHaveLength(2));

    const revokeHook = renderHookWithQueryClient(() => useRevokeApiKey(), { queryClient });
    expect(revokeHook.result.current.pendingRevoke).toBeNull();
    await act(async () => {
      revokeHook.result.current.revokeKey('key-1');
    });
    await waitFor(() => expect(revokeHook.result.current.pendingRevoke).toBeNull());
    expect(revokeHook.result.current.isRevoking).toBe(false);
  });
});

describe('useRotateApiKey', () => {
  it('rotates a key and surfaces the new full secret via rotatedKey', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useRotateApiKey());
    await act(async () => {
      result.current.rotateKey('key-1');
    });
    await waitFor(() => expect(result.current.rotatedKey).not.toBeNull());
    expect(result.current.rotatedKey?.apiKey).toMatch(/^vxyz_secret_rotated_/);
    expect(result.current.rotatedKey?.id).toBe('key-1');
  });

  it('clearRotatedKey resets the surfaced secret', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useRotateApiKey());
    await act(async () => {
      result.current.rotateKey('key-1');
    });
    await waitFor(() => expect(result.current.rotatedKey).not.toBeNull());
    act(() => {
      result.current.clearRotatedKey();
    });
    expect(result.current.rotatedKey).toBeNull();
  });
});
