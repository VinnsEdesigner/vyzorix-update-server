// GraphQL Client using Apollo Client
// Handles all GraphQL API communication

import { ApolloClient, InMemoryCache, createHttpLink, type NormalizedCacheObject } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

// ============================================================================
// Configuration
// ============================================================================

export interface GraphQLConfig {
  uri: string;
  credentials?: RequestCredentials;
}

function getGraphQLConfig(): GraphQLConfig {
  return {
    uri: import.meta.env.VITE_API_URL 
      ? `${import.meta.env.VITE_API_URL}/graphql` 
      : '/api/graphql',
    credentials: 'include',
  };
}

// ============================================================================
// Apollo Client
// ============================================================================

function createApolloClient(config: GraphQLConfig): ApolloClient<NormalizedCacheObject> {
  const httpLink = createHttpLink({
    uri: config.uri,
    credentials: config.credentials,
  });

  const authLink = setContext((_, { headers }) => {
    return {
      headers: {
        ...headers,
      },
    };
  });

  return new ApolloClient({
    link: authLink.concat(httpLink),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: {
        fetchPolicy: 'cache-and-network',
      },
      query: {
        fetchPolicy: 'network-only',
        errorPolicy: 'all',
      },
      mutate: {
        errorPolicy: 'all',
      },
    },
  });
}

let apolloClient: ApolloClient<NormalizedCacheObject> | null = null;

export function getApolloClient(): ApolloClient<NormalizedCacheObject> {
  if (!apolloClient) {
    apolloClient = createApolloClient(getGraphQLConfig());
  }
  return apolloClient;
}

// ============================================================================
// Client Instance (default export)
// ============================================================================

export const graphqlClient = getApolloClient();
