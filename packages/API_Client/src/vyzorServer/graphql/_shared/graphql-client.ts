import { ApolloClient, InMemoryCache, createHttpLink, type NormalizedCacheObject } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

type CredentialsType = 'omit' | 'same-origin' | 'include';

export interface GraphQLConfig {
  organizationId: string;
  credentials?: CredentialsType | undefined;
  getUri: () => string;
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

function getGraphQLUri(orgId: string): string {
  const baseUrl = (import.meta.env as Record<string, string | undefined>).VITE_API_URL || '/api';
  return `${baseUrl}/v1/orgs/${orgId}/graphql`;
}

export function getApolloClient(): ApolloClient<NormalizedCacheObject> {
  return apolloClient || createApolloClient({
    organizationId: '',
    credentials: 'include',
    getUri: () => getGraphQLUri(currentOrgId),
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
    getUri: () => getGraphQLUri(organizationId),
  });
}

export const graphqlClient = {
  getClient: getApolloClient,
  setOrganization: setOrganizationContext,
};
