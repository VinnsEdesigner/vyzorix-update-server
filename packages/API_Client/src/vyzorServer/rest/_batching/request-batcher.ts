/**
 * GraphQL request batching.
 *
 * Combines multiple GraphQL operations into one HTTP request against the
 * `/graphql/batch` endpoint. REST request de-duplication now lives in
 * `../_shared/request-queue.ts` (VyzorixRequestQueue).
 */

import { getSigningKey } from '../_shared/rest-client';
import { signHttpRequestBrowser } from '../../crypto/browser-sign';

function simpleHashPayload(data: unknown): string {
  const str = typeof data === 'string' ? data : (data === undefined ? '' : JSON.stringify(data)) ?? '';
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

// ==========================================
// GRAPHQL REQUEST BATCHING
// ==========================================

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

  constructor(config: Partial<BatchConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  configure(baseUrl: string, orgId: string, authToken: string): void {
    this.batchUrl = `${baseUrl}/${orgId}/graphql/batch`;
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
      const body = JSON.stringify(operations);
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      if (this.authToken) {
        headers['Authorization'] = `Bearer ${this.authToken}`;
      }

      const signingKey = getSigningKey();
      if (signingKey) {
        const signHeaders = await signHttpRequestBrowser('POST', this.batchUrl, body, signingKey);
        headers['X-Vyzorix-Timestamp'] = signHeaders['X-Vyzorix-Timestamp'];
        headers['X-Vyzorix-Nonce'] = signHeaders['X-Vyzorix-Nonce'];
        headers['X-Vyzorix-Signature'] = signHeaders['X-Vyzorix-Signature'];
      }

      const response = await fetch(this.batchUrl, {
        method: 'POST',
        headers,
        body,
      });

      if (!response.ok) {
        throw new Error(`Batch request failed: ${response.status}`);
      }

      const results = await response.json() as { data?: unknown; errors?: { message: string }[] }[];

      for (let i = 0; i < keyedOperations.length; i++) {
        const entry = keyedOperations[i];
        if (!entry) continue;
        const { op } = entry;
        const result = results[i];

        if (!result) {
          for (const caller of op.callers) {
            caller.reject(new Error('Batch response missing result'));
          }
          continue;
        }

        if (result.errors && result.errors.length > 0) {
          const firstError = result.errors[0];
          const err = new Error(firstError?.message ?? 'GraphQL batch error');
          for (const caller of op.callers) {
            caller.reject(err);
          }
        } else {
          for (const caller of op.callers) {
            caller.resolve(result);
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

// Single encapsulated holder for batcher singleton instances (previously two
// module-level `let` bindings — the singleton-globals smell).
const batcherState = {
  graphql: null as GraphQLBatcher | null,
};

/**
 * Get GraphQL request batcher instance
 */
export function getGraphQLBatcher(config?: Partial<BatchConfig>): GraphQLBatcher {
  if (!batcherState.graphql) {
    batcherState.graphql = new GraphQLBatcher(config);
  }
  return batcherState.graphql;
}

/**
 * Reset all batcher instances
 */
export function resetBatchers(): void {
  if (batcherState.graphql) {
    batcherState.graphql.clear();
  }
  batcherState.graphql = null;
}
