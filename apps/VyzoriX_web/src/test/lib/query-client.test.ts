import { describe, it, expect } from 'vitest';
import { createQueryClient } from '@/lib/query-client';

describe('createQueryClient', () => {
  it('returns a QueryClient instance', () => {
    const client = createQueryClient();
    expect(client).toBeDefined();
    expect(client.getDefaultOptions).toBeDefined();
  });

  it('sets staleTime to 30s', () => {
    const client = createQueryClient();
    expect(client.getDefaultOptions().queries?.staleTime).toBe(30_000);
  });

  it('sets gcTime to 5min', () => {
    const client = createQueryClient();
    expect(client.getDefaultOptions().queries?.gcTime).toBe(5 * 60_000);
  });

  it('disables refetchOnWindowFocus', () => {
    const client = createQueryClient();
    expect(client.getDefaultOptions().queries?.refetchOnWindowFocus).toBe(false);
  });

  it('disables mutation retries', () => {
    const client = createQueryClient();
    expect(client.getDefaultOptions().mutations?.retry).toBe(false);
  });

  it('returns a new instance each call', () => {
    const a = createQueryClient();
    const b = createQueryClient();
    expect(a).not.toBe(b);
  });

  it('retry function returns false for 4xx errors (non-408/429)', () => {
    const client = createQueryClient();
    const retry = client.getDefaultOptions().queries?.retry;
    expect(typeof retry).toBe('function');
    const result = (retry as (count: number, error: unknown) => boolean)(0, { status: 404 });
    expect(result).toBe(false);
  });

  it('retry function returns true for 5xx errors on first attempt', () => {
    const client = createQueryClient();
    const retry = client.getDefaultOptions().queries?.retry as (count: number, error: unknown) => boolean;
    expect(retry(0, { status: 500 })).toBe(true);
  });

  it('retry function stops after 2 failures', () => {
    const client = createQueryClient();
    const retry = client.getDefaultOptions().queries?.retry as (count: number, error: unknown) => boolean;
    expect(retry(2, { status: 500 })).toBe(false);
  });
});
