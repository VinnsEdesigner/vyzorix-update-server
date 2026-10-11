/**
 * Unit tests for VyzorixRequestQueue — the transport-level request
 * deduplicator (§3 of VYZORIX_FRONTEND_ARCHITECTURE).
 *
 * Real class, no mocks. The executor is an inline counter so we assert exactly
 * how many times the underlying transport would fire.
 */
import { describe, it, expect, vi } from 'vitest';
import { VyzorixRequestQueue } from '@vyzorix/api-client';

describe('VyzorixRequestQueue', () => {
  it('shares one in-flight promise across concurrent identical requests', async () => {
    const q = new VyzorixRequestQueue();
    let calls = 0;
    const executor = () => {
      calls += 1;
      return new Promise<{ n: number }>((resolve) => setTimeout(() => resolve({ n: 1 }), 10));
    };

    const [a, b, c] = await Promise.all([
      q.execute('GET', '/v1/devices', undefined, executor),
      q.execute('GET', '/v1/devices', undefined, executor),
      q.execute('GET', '/v1/devices', undefined, executor),
    ]);

    expect(calls).toBe(1);
    expect(a).toEqual({ n: 1 });
    expect(b).toBe(a);
    expect(c).toBe(a);
  });

  it('dedups non-GET methods by method+url+body (the duplicate-POST case)', async () => {
    const q = new VyzorixRequestQueue();
    let calls = 0;
    const executor = () => {
      calls += 1;
      return Promise.resolve({ ok: true });
    };

    await Promise.all([
      q.execute('POST', '/v1/devices/x/command', { command: 'restart' }, executor),
      q.execute('POST', '/v1/devices/x/command', { command: 'restart' }, executor),
    ]);
    expect(calls).toBe(1);

    // A different body is a different operation.
    await q.execute('POST', '/v1/devices/x/command', { command: 'reboot' }, executor);
    expect(calls).toBe(2);
  });

  it('serves a repeated GET from cache within staleMs without re-executing', async () => {
    const q = new VyzorixRequestQueue({ staleMs: 5000 });
    let calls = 0;
    const executor = () => {
      calls += 1;
      return Promise.resolve('payload');
    };

    await q.execute('GET', '/v1/me', undefined, executor);
    await q.execute('GET', '/v1/me', undefined, executor);
    expect(calls).toBe(1);
  });

  it('does NOT cache non-GET responses', async () => {
    const q = new VyzorixRequestQueue({ staleMs: 5000 });
    let calls = 0;
    const executor = () => {
      calls += 1;
      return Promise.resolve('ok');
    };

    await q.execute('PATCH', '/v1/me', { name: 'a' }, executor);
    await q.execute('PATCH', '/v1/me', { name: 'a' }, executor);
    expect(calls).toBe(2);
  });

  it('re-executes a GET once the stale window elapses', async () => {
    vi.useFakeTimers();
    try {
      const q = new VyzorixRequestQueue({ staleMs: 100 });
      let calls = 0;
      const executor = () => {
        calls += 1;
        return Promise.resolve(calls);
      };

      await q.execute('GET', '/v1/x', undefined, executor);
      vi.advanceTimersByTime(150);
      await q.execute('GET', '/v1/x', undefined, executor);
      expect(calls).toBe(2);
    } finally {
      vi.useRealTimers();
    }
  });

  it('key ordering is stable (params hashed independent of key order)', async () => {
    const q = new VyzorixRequestQueue();
    let calls = 0;
    const executor = () => {
      calls += 1;
      return new Promise<number>((resolve) => setTimeout(() => resolve(1), 10));
    };
    // Concurrent, so in-flight dedup (always on) applies regardless of staleMs.
    await Promise.all([
      q.execute('GET', '/v1/x', { a: 1, b: 2 }, executor),
      q.execute('GET', '/v1/x', { b: 2, a: 1 }, executor),
    ]);
    expect(calls).toBe(1);
  });

  it('propagates executor rejection to all sharers and clears in-flight', async () => {
    const q = new VyzorixRequestQueue();
    const boom = new Error('nope');
    const executor = () => Promise.reject(boom);

    const p1 = q.execute('POST', '/v1/x', { a: 1 }, executor);
    const p2 = q.execute('POST', '/v1/x', { a: 1 }, executor);
    await expect(p1).rejects.toBe(boom);
    await expect(p2).rejects.toBe(boom);
    expect(q.getInFlightCount()).toBe(0);
  });

  it('does NOT dedupe across org scopes (header-scoped requests)', async () => {
    const q = new VyzorixRequestQueue();
    let calls = 0;
    const executor = () => {
      calls += 1;
      return new Promise<number>((resolve) => setTimeout(() => resolve(calls), 10));
    };
    await Promise.all([
      q.execute('GET', '/v1/devices', undefined, executor, 'org-a'),
      q.execute('GET', '/v1/devices', undefined, executor, 'org-b'),
    ]);
    // Same URL + params but different org -> must not share a response.
    expect(calls).toBe(2);
  });

  it('bypasses dedupe when a per-caller AbortSignal is supplied', async () => {
    const q = new VyzorixRequestQueue();
    let calls = 0;
    const executor = () => {
      calls += 1;
      return Promise.resolve(1);
    };
    const ctrl = new AbortController();
    await Promise.all([
      q.execute('GET', '/v1/x', undefined, executor, undefined, ctrl.signal),
      q.execute('GET', '/v1/x', undefined, executor, undefined, ctrl.signal),
    ]);
    // Each abortable caller runs its own request.
    expect(calls).toBe(2);
  });
});
