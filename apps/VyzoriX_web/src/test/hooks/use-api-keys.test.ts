/**
 * Integration tests for the API key hooks (generated-rq).
 *
 * The hooks are thin wrappers over the orval react-query hooks
 * (src/generated-rq/api-keys). These render the REAL hooks via RTL; the
 * generated hooks call the REAL customAxios bridge → restClient. MSW
 * intercepts the HTTP requests mirroring the Go server contract
 * (snake_case, ISO timestamps).
 *
 * Contract change vs the old wrapper hooks: the generated hooks return raw
 * TanStack Query results — data is the orval response envelope
 * ({ status, data }) and mutations take { data: ... } / { keyId } args.
 * There is no org gating, flattening, validation, or optimistic layer in the
 * generated surface.
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { waitFor, act } from '@testing-library/react';
import { renderHookWithQueryClient, createTestQueryClient } from '../helpers/render-hook';
import { setupIntegrationTest } from '../helpers/integration-test-setup';
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

beforeEach(() => {
  resetApiKeyFixtures();
});

// The orval response envelope: { status: 200, data: APIKeyListResult }.
function listData(res: { data?: unknown }) {
  return res.data as { keys?: unknown[]; pagination?: { total?: number }; monthly_limit?: number; keys_created_this_month?: number } | undefined;
}

describe('useApiKeys', () => {
  it('fetches api keys', async () => {
    const { result } = renderHookWithQueryClient(() => useApiKeys());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const data = listData(result.current);
    expect(data?.keys).toHaveLength(2);
    expect(data?.pagination?.total).toBe(2);
  });

  it('returns the wire key shape (snake_case, ISO timestamps)', async () => {
    const { result } = renderHookWithQueryClient(() => useApiKeys());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const data = listData(result.current);
    const first = (data?.keys ?? [])[0] as Record<string, unknown> | undefined;
    expect(first).toBeDefined();
    expect(first?.id).toBe('key-1');
    expect(first?.key_prefix).toBe('vxyz_abcd');
    expect(first?.scope).toBe('admin');
    expect(first?.is_active).toBe(true);
  });

  it('exposes monthly stats from the server response', async () => {
    const { result } = renderHookWithQueryClient(() => useApiKeys());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const data = listData(result.current);
    expect(data?.monthly_limit).toBe(20);
    expect(data?.keys_created_this_month).toBe(1);
  });
});

describe('useApiKey', () => {
  it('is disabled when no id is provided', () => {
    const { result } = renderHookWithQueryClient(() => useApiKey(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches a single key by id', async () => {
    const { result } = renderHookWithQueryClient(() => useApiKey('key-1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const body = result.current.data as { id?: string; name?: string } | undefined;
    expect(body?.id).toBe('key-1');
    expect(body?.name).toBe('Production Key');
  });

  it('returns an error for a missing key', async () => {
    const { result } = renderHookWithQueryClient(() => useApiKey('nope'));
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeDefined();
  });
});

describe('useCreateApiKey', () => {
  it('creates a key and returns the full secret', async () => {
    const { result } = renderHookWithQueryClient(() => useCreateApiKey());
    await act(async () => {
      result.current.mutate({ data: { name: 'New Key', scope: 'write' } });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const body = result.current.data as { api_key?: string; name?: string; scope?: string } | undefined;
    expect(body?.api_key).toMatch(/^vxyz_secret_/);
    expect(body?.name).toBe('New Key');
    expect(body?.scope).toBe('write');
  });

  it('invalidates the api-keys list cache on success', async () => {
    const queryClient = createTestQueryClient();
    const listHook = renderHookWithQueryClient(() => useApiKeys(), { queryClient });
    await waitFor(() => expect((listData(listHook.result.current)?.keys ?? [])).toHaveLength(2));

    const createHook = renderHookWithQueryClient(() => useCreateApiKey(), { queryClient });
    await act(async () => {
      createHook.result.current.mutate({ data: { name: 'Cache Test', scope: 'read' } });
    });
    await waitFor(() => expect(createHook.result.current.isSuccess).toBe(true));
    await waitFor(() => expect((listData(listHook.result.current)?.keys ?? [])).toHaveLength(3));
  });
});

describe('useUpdateApiKey', () => {
  it('updates a key name and scope', async () => {
    const { result } = renderHookWithQueryClient(() => useUpdateApiKey());
    await act(async () => {
      result.current.mutate({ keyId: 'key-1', data: { name: 'Renamed', scope: 'read' } });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const body = result.current.data as { name?: string; scope?: string } | undefined;
    expect(body?.name).toBe('Renamed');
    expect(body?.scope).toBe('read');
  });
});

describe('useRevokeApiKey', () => {
  it('revokes a key and refetches the list', async () => {
    const queryClient = createTestQueryClient();
    const listHook = renderHookWithQueryClient(() => useApiKeys(), { queryClient });
    await waitFor(() => expect((listData(listHook.result.current)?.keys ?? [])).toHaveLength(2));

    const revokeHook = renderHookWithQueryClient(() => useRevokeApiKey(), { queryClient });
    await act(async () => {
      revokeHook.result.current.mutate({ keyId: 'key-1' });
    });
    await waitFor(() => expect(revokeHook.result.current.isSuccess).toBe(true));
    await waitFor(() => expect((listData(listHook.result.current)?.keys ?? [])).toHaveLength(1));
  });
});

describe('useRotateApiKey', () => {
  it('rotates a key and returns the new full secret', async () => {
    const { result } = renderHookWithQueryClient(() => useRotateApiKey());
    await act(async () => {
      result.current.mutate({ keyId: 'key-1' });
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const body = result.current.data as { api_key?: string; id?: string } | undefined;
    expect(body?.api_key).toMatch(/^vxyz_secret_rotated_/);
    expect(body?.id).toBe('key-1');
  });
});
