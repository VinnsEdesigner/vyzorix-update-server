// GraphQL client for Vyzorix API
// Uses session cookie authentication (cookies sent automatically)

import { GraphQLClient } from "graphql-request";

import { logger } from "@/lib/logger";

// Create a singleton GraphQL client
// The API server uses session cookies, so we need credentials: 'include'
const createGraphQLClient = (baseUrl: string) => {
  return new GraphQLClient(`${baseUrl}/graphql`, {
    credentials: "include", // Send cookies for session auth
    headers: {
      "Content-Type": "application/json",
    },
  });
};

// Get the API base URL from environment or window
export const getApiBaseUrl = (): string => {
  // Server-side rendering
  if (typeof window === "undefined") {
    return process.env.VYZORIX_API_URL ?? "http://localhost:3000";
  }
  // Client-side - use same origin
  return window.location.origin;
};

// Singleton client instance
let graphqlClient: GraphQLClient | null = null;

export const getGraphQLClient = (): GraphQLClient => {
  if (!graphqlClient) {
    graphqlClient = createGraphQLClient(getApiBaseUrl());
    logger.info("graphql", "GraphQL client initialized");
  }
  return graphqlClient;
};

// Reset client (useful for testing or re-initialization)
export const resetGraphQLClient = () => {
  graphqlClient = null;
};

export { GraphQLClient };
