// GraphQL QueryClient integration
// This provides a way to use GraphQL with TanStack Query

import { QueryClient } from "@tanstack/react-query";

// Create a GraphQL-specific query client configuration
// This can be used alongside the main query client for GraphQL operations
export const createGraphQLQueryClient = (): QueryClient => {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // GraphQL queries typically need longer cache times
        // since the data is fetched via a single endpoint
        staleTime: 1000 * 60 * 5, // 5 minutes
        gcTime: 1000 * 60 * 30, // 30 minutes (formerly cacheTime)
        refetchOnWindowFocus: false,
        retry: 1,
      },
      mutations: {
        // Mutations should retry less since they modify data
        retry: 0,
      },
    },
  });
};

// Export a singleton for use across the app
export const graphqlQueryClient = createGraphQLQueryClient();
