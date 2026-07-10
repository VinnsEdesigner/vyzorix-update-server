/**
 * GraphQL Client Configuration
 * 
 * Shared GraphQL client setup for all features.
 * Uses session-based authentication (cookies).
 */

import type { FetchOptions } from "../../rest/_shared/rest-client";

// ============================================================================
// Configuration
// ============================================================================

/**
 * GraphQL endpoint configuration
 */
export interface GraphQLConfig {
  endpoint: string;
  credentials?: RequestCredentials;
}

/**
 * Default GraphQL config from environment
 */
export function getGraphQLConfig(): GraphQLConfig {
  const endpoint = import.meta.env.VITE_API_URL 
    ? `${import.meta.env.VITE_API_URL}/graphql`
    : "/api/graphql";
  
  return {
    endpoint,
    credentials: "include", // Send cookies for session auth
  };
}

// ============================================================================
// GraphQL Client
// ============================================================================

/**
 * GraphQL request options
 */
export interface GraphQLRequestOptions extends FetchOptions {
  query: string;
  variables?: Record<string, unknown>;
  operationName?: string;
}

/**
 * GraphQL response wrapper
 */
export interface GraphQLResponse<T> {
  data?: T;
  errors?: Array<{
    message: string;
    locations?: Array<{ line: number; column: number }>;
    path?: string[];
  }>;
}

/**
 * Execute a GraphQL query or mutation
 */
export async function graphqlRequest<T>(
  options: GraphQLRequestOptions
): Promise<GraphQLResponse<T>> {
  const config = getGraphQLConfig();
  const { query, variables, operationName, ...fetchOptions } = options;
  
  const response = await fetch(config.endpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...fetchOptions.headers,
    },
    body: JSON.stringify({
      query,
      variables,
      operationName,
    }),
    credentials: config.credentials,
    ...fetchOptions,
  });
  
  if (!response.ok) {
    throw new Error(`GraphQL request failed: ${response.status} ${response.statusText}`);
  }
  
  return response.json();
}

// ============================================================================
// Query Helper
// ============================================================================

/**
 * Execute a GraphQL query (GET method for caching)
 */
export async function graphqlQuery<T>(
  query: string,
  variables?: Record<string, unknown>,
  options?: Omit<GraphQLRequestOptions, "query" | "variables">
): Promise<GraphQLResponse<T>> {
  return graphqlRequest<T>({
    query,
    variables,
    method: "POST",
    ...options,
  });
}

// ============================================================================
// Error Handling
// ============================================================================

/**
 * Extract error message from GraphQL response
 */
export function getGraphQLErrorMessage(response: GraphQLResponse<unknown>): string | null {
  if (response.errors && response.errors.length > 0) {
    return response.errors.map((e) => e.message).join(", ");
  }
  return null;
}

/**
 * Check if GraphQL response has errors
 */
export function hasGraphQLErrors(response: GraphQLResponse<unknown>): boolean {
  return Boolean(response.errors && response.errors.length > 0);
}

// ============================================================================
// Type Utilities
// ============================================================================

/**
 * Unwrap GraphQL data or throw
 */
export function unwrapGraphQLData<T>(
  response: GraphQLResponse<T>,
  operationName?: string
): T {
  if (hasGraphQLErrors(response)) {
    const message = getGraphQLErrorMessage(response) ?? "GraphQL error";
    throw new Error(operationName ? `${operationName}: ${message}` : message);
  }
  
  if (!response.data) {
    throw new Error("No data returned from GraphQL");
  }
  
  return response.data;
}