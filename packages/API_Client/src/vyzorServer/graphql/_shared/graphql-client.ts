import { ApolloClient, InMemoryCache, createHttpLink, type NormalizedCacheObject } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

export interface GraphQLConfig {
  organizationId: string;
  credentials?: RequestCredentials;
}

function getGraphQLConfig(): GraphQLConfig {
  const baseUrl = import.meta.env.VITE_API_URL || '/api';
  return {
    organizationId: '',
    credentials: 'include',
    getUri: () => `${baseUrl}/v1/orgs/${getGraphQLConfig().organizationId}/graphql`,
  };
}

function createApolloClient(config: GraphQLConfig): ApolloClient<NormalizedCacheObject> {
  const httpLink = createHttpLink({
    uri: config.getUri(),
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
let currentOrgId: string = '';

export function getApolloClient(): ApolloClient<NormalizedCacheObject> {
  return apolloClient || createApolloClient({
    organizationId: '',
    credentials: 'include',
    getUri: () => {
      const baseUrl = import.meta.env.VITE_API_URL || '/api';
      return `${baseUrl}/v1/orgs/${currentOrgId}/graphql`;
    },
  });
}

export function setOrganizationContext(organizationId: string): void {
  currentOrgId = organizationId;
  if (apolloClient) {
    apolloClient.stop();
    apolloClient = null;
  }
  apolloClient = createApolloClient({
    organizationId,
    credentials: 'include',
    getUri: () => {
      const baseUrl = import.meta.env.VITE_API_URL || '/api';
      return `${baseUrl}/v1/orgs/${organizationId}/graphql`;
    },
  });
}

export const graphqlClient = {
  getClient: getApolloClient,
  setOrganization: setOrganizationContext,
};
