/**
 * Request Collapsing & Batching Logic
 * 
 * Reduces network round trips by combining multiple queries into one.
 * 
 * How it works:
 * - Query Consolidation (REST): Track all API requests fired within a short time window (50ms)
 *   If multiple requests to the same endpoint/pattern are detected, collapse them into a single
 *   batched request and return same data to all original callers.
 * - GraphQL Batching: Combine multiple GraphQL operations into one HTTP request
 * - Cache Integration: Combine with response caching for maximum efficiency
 */

function simpleHashPayload(data: unknown): string {
  const str = typeof data === 'string' ? data : JSON.stringify(data);
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash;
  }
  return Math.abs(hash).toString(36);
}

// ==========================================
// TYPES & INTERFACES
// ==========================================

export interface BatchedRequest<T = unknown> {
  key: string;
  promise: Promise<T>;
  url: string;
  method: string;
  data?: unknown;
  config?: Record<string, unknown>;
  callers: BatchedCaller<T>[];
  timer?: ReturnType<typeof setTimeout>;
  executed: boolean;
}

export interface BatchedCaller<T> {
  resolve: (value: T) => void;
  reject: (error: Error) => void;
}

export interface GraphQLBatchedOperation {
  query: string;
  variables?: Record<string, unknown>;
  operationName?: string;
}

export interface BatchConfig {
  /** Time window in ms to collect requests before flushing (default: 50ms) */
  windowMs: number;
  /** Maximum requests to batch together (default: 10) */
  maxBatchSize: number;
  /** Enable REST request collapsing (default: true) */
  enableREST: boolean;
  /** Enable GraphQL batching (default: true) */
  enableGraphQL: boolean;
}

const DEFAULT_CONFIG: BatchConfig = {
  windowMs: 50,
  maxBatchSize: 10,
  enableREST: true,
  enableGraphQL: true,
};

// ==========================================
// REST REQUEST COLLAPSING
// ==========================================

/**
 * REST Request Batcher
 * 
 * Tracks in-flight requests and collapses identical ones within a time window.
 */
export class RESTRequestBatcher {
  private pendingRequests: Map<string, BatchedRequest> = new Map();
  private config: BatchConfig;

  constructor(config: Partial<BatchConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  /**
   * Generate a batch key for REST requests
   */
  private getBatchKey(method: string, url: string, data?: unknown): string {
    return `${method}:${url}:${simpleHashPayload(data)}`;
  }

  /**
   * Check if a request can be batched (same method + url + data)
   */
  private canBatch(existing: BatchedRequest, method: string, url: string, data?: unknown): boolean {
    return existing.method === method && existing.url === url && simpleHashPayload(existing.data) === simpleHashPayload(data);
  }

  /**
   * Execute or queue a REST request
   */
  async execute<T>(
    method: string,
    url: string,
    data: unknown,
    config: Record<string, unknown> | undefined,
    executor: () => Promise<T>
  ): Promise<T> {
    if (!this.config.enableREST) {
      return executor();
    }

    const batchKey = this.getBatchKey(method, url, data);
    const existing = this.pendingRequests.get(batchKey);

    // If same request is already pending, attach to its callers
    if (existing && this.canBatch(existing, method, url, data)) {
      return new Promise<T>((resolve, reject) => {
        existing.callers.push({ resolve: resolve as (value: unknown) => void, reject });
        console.debug(`[Batcher] Request collapsed: ${method} ${url}`);
      });
    }

    // Create new batched request
    return new Promise<T>((resolve, reject) => {
      const batchedRequest: BatchedRequest<T> = {
        key: batchKey,
        promise: null as unknown as Promise<T>,
        url,
        method,
        data,
        config,
        callers: [{ resolve: resolve as (value: unknown) => void, reject }],
        executed: false,
      };

      // Set up timer to flush batch after window
      batchedRequest.timer = setTimeout(() => {
        this.flushBatch(batchedRequest, executor);
      }, this.config.windowMs);

      this.pendingRequests.set(batchKey, batchedRequest as BatchedRequest);

      console.debug(`[Batcher] Request queued: ${method} ${url}`);
    });
  }

  /**
   * Flush a batch - execute request and distribute results
   */
  private async flushBatch<T>(
    request: BatchedRequest<T>,
    executor: () => Promise<T>
  ): Promise<void> {
    if (request.executed) return;
    request.executed = true;

    // Remove from pending
    this.pendingRequests.delete(request.key);

    if (request.timer) {
      clearTimeout(request.timer);
    }

    try {
      // Execute the actual request
      const result = await executor();

      // Resolve all callers with the same result
      for (const caller of request.callers) {
        caller.resolve(result);
      }
    } catch (error) {
      // Reject all callers with the same error
      const err = error instanceof Error ? error : new Error(String(error));
      for (const caller of request.callers) {
        caller.reject(err);
      }
    }
  }

  /**
   * Cancel a pending batched request
   */
  cancel(method: string, url: string, data?: unknown): void {
    const batchKey = this.getBatchKey(method, url, data);
    const request = this.pendingRequests.get(batchKey);

    if (request) {
      if (request.timer) {
        clearTimeout(request.timer);
      }

      const error = new Error('Request cancelled');
      for (const caller of request.callers) {
        caller.reject(error);
      }

      this.pendingRequests.delete(batchKey);
    }
  }

  /**
   * Clear all pending requests
   */
  clear(): void {
    for (const request of this.pendingRequests.values()) {
      if (request.timer) {
        clearTimeout(request.timer);
      }
      const error = new Error('Batcher cleared');
      for (const caller of request.callers) {
        caller.reject(error);
      }
    }
    this.pendingRequests.clear();
  }

  /**
   * Get count of pending requests
   */
  getPendingCount(): number {
    return this.pendingRequests.size;
  }

  /**
   * Get batch statistics
   */
  getStats(): { pending: number; collapsed: number } {
    let collapsed = 0;
    for (const request of this.pendingRequests.values()) {
      collapsed += request.callers.length - 1;
    }
    return { pending: this.pendingRequests.size, collapsed };
  }
}

// ==========================================
// GRAPHQL REQUEST BATCHING
// ==========================================

interface BatchedGraphQLOperation {
  query: string;
  variables?: Record<string, unknown>;
  operationName?: string;
  callers: {
    resolve: (value: unknown) => void;
    reject: (error: Error) => void;
    operationName?: string;
  }[];
  timer?: ReturnType<typeof setTimeout>;
  executed: boolean;
}

/**
 * GraphQL Request Batcher
 * 
 * Combines multiple GraphQL operations into a single HTTP request using the /batch endpoint.
 */
export class GraphQLBatcher {
  private pendingOperations: Map<string, BatchedGraphQLOperation> = new Map();
  private config: BatchConfig;
  private batchUrl: string = '';
  private authToken: string = '';
  private orgId: string = '';

  constructor(config: Partial<BatchConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  configure(baseUrl: string, orgId: string, authToken: string): void {
    this.batchUrl = `${baseUrl}/v1/orgs/${orgId}/graphql/batch`;
    this.orgId = orgId;
    this.authToken = authToken;
  }

  private getBatchKey(query: string, variables?: Record<string, unknown>): string {
    return simpleHashPayload({ query, variables });
  }

  async execute<T>(
    query: string,
    variables: Record<string, unknown> | undefined,
    operationName: string | undefined,
    _executor: () => Promise<T>
  ): Promise<T> {
    if (!this.config.enableGraphQL || !this.batchUrl) {
      return _executor();
    }

    const batchKey = this.getBatchKey(query, variables);

    const existing = this.pendingOperations.get(batchKey);
    if (existing && existing.query === query && JSON.stringify(existing.variables) === JSON.stringify(variables)) {
      return new Promise<T>((resolve, reject) => {
        existing.callers.push({
          resolve: resolve as (value: unknown) => void,
          reject,
          operationName,
        });
      });
    }

    return new Promise<T>((resolve, reject) => {
      const batchedOp: BatchedGraphQLOperation = {
        query,
        variables,
        operationName,
        callers: [{ resolve: resolve as (value: unknown) => void, reject, operationName }],
        timer: setTimeout(() => {
          this.flushBatch();
        }, this.config.windowMs),
        executed: false,
      };

      this.pendingOperations.set(batchKey, batchedOp);
    });
  }

  /**
   * Flush a GraphQL batch - sends all pending operations to the /batch endpoint
   */
  private async flushBatch(): Promise<void> {
    const operations: GraphQLBatchedOperation[] = [];
    const keyedOperations: { key: string; op: BatchedGraphQLOperation }[] = [];

    for (const [key, op] of this.pendingOperations.entries()) {
      if (op.executed) continue;
      op.executed = true;
      if (op.timer) clearTimeout(op.timer);

      operations.push({
        query: op.query,
        variables: op.variables,
        operationName: op.operationName,
      });
      keyedOperations.push({ key, op });

      if (operations.length >= this.config.maxBatchSize) break;
    }

    if (operations.length === 0) return;

    this.pendingOperations.clear();

    try {
      const response = await fetch(this.batchUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.authToken}`,
        },
        body: JSON.stringify(operations),
      });

      if (!response.ok) {
        throw new Error(`Batch request failed: ${response.status}`);
      }

      const results = await response.json() as { data?: unknown; errors?: { message: string }[] }[];

      for (let i = 0; i < keyedOperations.length; i++) {
        const { op } = keyedOperations[i];
        const result = results[i];

        if (result.errors && result.errors.length > 0) {
          const err = new Error(result.errors[0].message);
          for (const caller of op.callers) {
            caller.reject(err);
          }
        } else {
          for (const caller of op.callers) {
            caller.resolve(result.data);
          }
        }
      }
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      for (const { op } of keyedOperations) {
        for (const caller of op.callers) {
          caller.reject(err);
        }
      }
    }
  }

  /**
   * Clear all pending operations
   */
  clear(): void {
    for (const op of this.pendingOperations.values()) {
      if (op.timer) clearTimeout(op.timer);
      const error = new Error('GraphQL batcher cleared');
      for (const caller of op.callers) {
        caller.reject(error);
      }
    }
    this.pendingOperations.clear();
  }

  /**
   * Get pending count
   */
  getPendingCount(): number {
    return this.pendingOperations.size;
  }
}

// ==========================================
// COMBINED BATCHER INSTANCE
// ==========================================

// Singleton instances
let restBatcherInstance: RESTRequestBatcher | null = null;
let graphqlBatcherInstance: GraphQLBatcher | null = null;

/**
 * Get REST request batcher instance
 */
export function getRESTBatcher(config?: Partial<BatchConfig>): RESTRequestBatcher {
  if (!restBatcherInstance) {
    restBatcherInstance = new RESTRequestBatcher(config);
  }
  return restBatcherInstance;
}

/**
 * Get GraphQL request batcher instance
 */
export function getGraphQLBatcher(config?: Partial<BatchConfig>): GraphQLBatcher {
  if (!graphqlBatcherInstance) {
    graphqlBatcherInstance = new GraphQLBatcher(config);
  }
  return graphqlBatcherInstance;
}

/**
 * Reset all batcher instances
 */
export function resetBatchers(): void {
  if (restBatcherInstance) {
    restBatcherInstance.clear();
  }
  if (graphqlBatcherInstance) {
    graphqlBatcherInstance.clear();
  }
  restBatcherInstance = null;
  graphqlBatcherInstance = null;
}
