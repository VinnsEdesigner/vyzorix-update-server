import { ApolloClient, InMemoryCache, createHttpLink, type NormalizedCacheObject, ApolloLink, Observable, type Operation, type FetchResult } from '@apollo/client';
import { getMainDefinition } from '@apollo/client/utilities';
import { print } from 'graphql';
import { setContext } from '@apollo/client/link/context';
import { getGraphQLBatcher } from '../../rest/_batching';

type CredentialsType = 'omit' | 'same-origin' | 'include';

export interface GraphQLConfig {
  organizationId: string;
  credentials?: CredentialsType | undefined;
  getUri: () => string;
}

// Single encapsulated holder for all graphql-client mutable state (previously
// three module-level `let` bindings — the singleton-globals smell).
const graphqlState = {
  currentOrgId: '' as string,
  currentAuthToken: '' as string,
  apolloClient: null as ApolloClient<NormalizedCacheObject> | null,
};

function getGraphQLUri(orgId: string): string {
  const baseUrl = (import.meta.env as Record<string, string | undefined>).VITE_API_URL || '/api';
  return `${baseUrl}/v1/orgs/${orgId}/graphql`;
}

function getGraphQLBaseUrl(): string {
  const baseUrl = (import.meta.env as Record<string, string | undefined>).VITE_API_URL || '/api';
  return baseUrl.replace(/\/$/, '');
}

function createBatchingLink(httpLink: ApolloLink): ApolloLink {
  return new ApolloLink((operation: Operation) => {
    const batcher = getGraphQLBatcher();
    
    if (!batcher) {
      return forward(operation, httpLink);
    }

    const definition = getMainDefinition(operation.query);
    const queryString = definition.kind === 'OperationDefinition' 
      ? print(operation.query)
      : '';
    const { variables, operationName } = operation;

    return new Observable<FetchResult>((observer) => {
      batcher
        .execute(
          queryString,
          variables,
          operationName,
          async () => {
            return new Promise<FetchResult>((resolve, reject) => {
              forward(operation, httpLink).subscribe({
                next: resolve,
                error: reject,
                complete: () => {},
              });
            });
          }
        )
        .then(result => {
          observer.next(result as FetchResult);
          observer.complete();
        })
        .catch(error => {
          observer.error(error);
        });
    });
  });
}

function forward(operation: Operation, link: ApolloLink): Observable<FetchResult> {
  return new Observable<FetchResult>(observer => {
    const subscription = link.request(operation)?.subscribe(observer);
    return () => subscription?.unsubscribe();
  });
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

  const batchingLink = createBatchingLink(httpLink);

  return new ApolloClient({
    link: authLink.concat(batchingLink),
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

export function getApolloClient(): ApolloClient<NormalizedCacheObject> {
  return graphqlState.apolloClient || createApolloClient({
    organizationId: '',
    credentials: 'include',
    getUri: () => getGraphQLUri(graphqlState.currentOrgId),
  });
}

export function setOrganizationContext(organizationId: string, authToken?: string): void {
  graphqlState.currentOrgId = organizationId;
  if (authToken) {
    graphqlState.currentAuthToken = authToken;
  }
  
  const batcher = getGraphQLBatcher();
  if (batcher) {
    batcher.configure(getGraphQLBaseUrl(), organizationId, graphqlState.currentAuthToken);
  }

  if (graphqlState.apolloClient) {
    graphqlState.apolloClient.stop();
    graphqlState.apolloClient = null;
  }
  graphqlState.apolloClient = createApolloClient({
    organizationId,
    credentials: 'include',
    getUri: () => getGraphQLUri(organizationId),
  });
}

export function setAuthToken(token: string): void {
  graphqlState.currentAuthToken = token;
  const batcher = getGraphQLBatcher();
  if (batcher && graphqlState.currentOrgId) {
    batcher.configure(getGraphQLBaseUrl(), graphqlState.currentOrgId, token);
  }
}

export const graphqlClient = {
  getClient: getApolloClient,
  setOrganization: setOrganizationContext,
  setAuthToken,
};
