/**
 * VyzorixRequestQueue — transport-level request deduplication.
 *
 * Unifies the previous split between the GET-only request *batcher* (a 50ms
 * collapse window) and the per-method in-flight `Map` in `restClient`. Every
 * HTTP method now goes through one queue that:
 *
 *   - hashes `method + url + params/body` into a dedup key;
 *   - shares the in-flight promise across all callers with the same key (so a
 *     duplicate POST/PATCH isn't fired twice — the multi-conversation case);
 *   - serves a completed GET from a short stale cache (default 5s) so a burst
 *     of identical reads within the window costs one round trip;
 *   - keeps the dedup map consistent without a lock (JS is single-threaded),
 *     while still letting unrelated requests run concurrently.
 *
 * The primitive is deliberately small. The connectivity monitor's offline
 * queue, the circuit breaker, and HMAC signing are transport concerns handled
 * by `restClient` *inside* the executor this queue is given.
 */

export interface RequestQueueConfig {
  /**
   * How long a completed GET stays served from the queue's cache, in ms.
   *
   * Defaults to 0 (disabled). A non-zero window is an *opaque* HTTP-level
   * cache that TanStack Query's `invalidateQueries` cannot see, so a
   * "mutate then refetch" would observe stale data. In an app whose query
   * cache is the freshness authority, cross-request GET caching must be
   * opt-in, not the default. In-flight dedup below is always on.
   */
  staleMs: number;
  /** Cap on tracked in-flight entries; the oldest are evicted past this. */
  maxQueueSize: number;
}

export const DEFAULT_REQUEST_QUEUE_CONFIG: RequestQueueConfig = {
  staleMs: 0,
  maxQueueSize: 100,
};

interface InFlightEntry {
  promise: Promise<unknown>;
}

interface CacheEntry {
  value: unknown;
  expiresAt: number;
}

function stableStringify(value: unknown): string {
  if (value === undefined || value === null) return '';
  if (typeof value === 'string') return value;
  if (typeof value !== 'object') return String(value);
  if (Array.isArray(value)) {
    return '[' + value.map(stableStringify).join(',') + ']';
  }
  const obj = value as Record<string, unknown>;
  const keys = Object.keys(obj).sort();
  return '{' + keys.map((k) => JSON.stringify(k) + ':' + stableStringify(obj[k])).join(',') + '}';
}

function dedupKey(method: string, url: string, payload: unknown, scope?: string): string {
  return `${scope ? scope + '|' : ''}${method.toUpperCase()}:${url}:${stableStringify(payload)}`;
}

export class VyzorixRequestQueue {
  private config: RequestQueueConfig;
  private inFlight: Map<string, InFlightEntry> = new Map();
  private cache: Map<string, CacheEntry> = new Map();

  private stats = { deduped: 0, cached: 0, executed: 0 };

  constructor(config: Partial<RequestQueueConfig> = {}) {
    this.config = { ...DEFAULT_REQUEST_QUEUE_CONFIG, ...config };
  }

  /**
   * Execute (or dedupe) a request.
   *
   * @param method  HTTP method — only GET responses are stale-cached.
   * @param url     request path (must already include org scope)
   * @param payload params (GET/DELETE) or body (POST/PUT/PATCH) used for the key
   * @param executor performs the real transport call
   * @param scope   optional extra key component (e.g. the active org id) so
   *                requests scoped by a header can't collide across scopes
   * @param signal  when present the call is NOT deduped: the caller can abort
   *                it at any time, and aborting a shared promise would cancel
   *                the request out from under the other sharers.
   */
  execute<T>(
    method: string,
    url: string,
    payload: unknown,
    executor: () => Promise<T>,
    scope?: string,
    signal?: { aborted: boolean },
  ): Promise<T> {
    const upper = method.toUpperCase();

    // Per-caller cancellation must never be shared.
    if (signal) {
      this.stats.executed += 1;
      return executor();
    }

    const key = dedupKey(upper, url, payload, scope);
    const now = Date.now();

    // 1. Fresh cache (GET only).
    if (upper === 'GET') {
      const hit = this.cache.get(key);
      if (hit && hit.expiresAt > now) {
        this.stats.cached += 1;
        return Promise.resolve(hit.value as T);
      }
      if (hit) this.cache.delete(key);
    }

    // 2. In-flight sharing (all methods).
    const pending = this.inFlight.get(key);
    if (pending) {
      this.stats.deduped += 1;
      return pending.promise as Promise<T>;
    }

    // 3. New request. The in-flight entry is registered synchronously (below)
    //    before the executor runs, so a caller arriving later on the same tick
    //    sees it. JS is single-threaded, so the dedup map needs no lock; the
    //    design doc's "single worker" is intentionally NOT used here because
    //    serialising unrelated requests would defeat HTTP/2 multiplexing.
    const promise = this.run<T>(upper, key, executor).finally(() => {
      if (this.inFlight.get(key)?.promise === promise) {
        this.inFlight.delete(key);
      }
    });

    this.inFlight.set(key, { promise });
    this.evictIfNeeded();
    return promise;
  }

  private async run<T>(method: string, key: string, executor: () => Promise<T>): Promise<T> {
    this.stats.executed += 1;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const value: any = await executor();
    if (method === 'GET') {
      this.cache.set(key, { value, expiresAt: Date.now() + this.config.staleMs });
    }
    return value as T;
  }

  private evictIfNeeded(): void {
    while (this.inFlight.size > this.config.maxQueueSize) {
      const oldest = this.inFlight.keys().next().value;
      if (oldest === undefined) break;
      this.inFlight.delete(oldest);
    }
  }

  /** Drop a cached GET (e.g. after a mutation to the same resource). */
  invalidate(method: string, url: string, payload?: unknown, scope?: string): void {
    this.cache.delete(dedupKey(method, url, payload, scope));
  }

  clear(): void {
    this.cache.clear();
    // Reject nothing: in-flight callers keep their promises; we only stop
    // tracking them so a fresh test starts clean.
    this.inFlight.clear();
  }

  getStats(): { inFlight: number; cachedEntries: number; deduped: number; executed: number } {
    return { inFlight: this.inFlight.size, cachedEntries: this.cache.size, deduped: this.stats.deduped, executed: this.stats.executed };
  }

  /** Test helper — number of in-flight entries. */
  getInFlightCount(): number {
    return this.inFlight.size;
  }
}

// Single encapsulated holder for the queue singleton instance.
const queueState: { queue: VyzorixRequestQueue | null } = { queue: null };

export function getRequestQueue(config?: Partial<RequestQueueConfig>): VyzorixRequestQueue {
  if (!queueState.queue) {
    queueState.queue = new VyzorixRequestQueue(config);
  }
  return queueState.queue;
}

export function resetRequestQueue(): void {
  queueState.queue?.clear();
  queueState.queue = null;
}
