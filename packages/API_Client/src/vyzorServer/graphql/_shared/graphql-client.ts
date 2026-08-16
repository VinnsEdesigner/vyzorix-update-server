import { ApolloClient, InMemoryCache, createHttpLink, type NormalizedCacheObject, ApolloLink, Observable, type Operation, type FetchResult } from '@apollo/client';
import { getMainDefinition } from '@apollo/client/utilities';
import { print } from 'graphql';
import { setContext } from '@apollo/client/link/context';
import { getGraphQLBatcher } from '../../rest/_batching';
import { getSigningKey } from '../../rest/_shared/rest-client';
import { signHttpRequestBrowser } from '../../crypto/browser-sign';

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

// The server registers GraphQL at /:org/graphql (server_graphql.go). The REST
// client's baseURL (VITE_API_URL, default "/api") is prepended so the vite
// proxy forwards the request to the API server.
function getGraphQLUri(orgId: string): string {
  const baseUrl = (import.meta.env as Record<string, string | undefined>).VITE_API_URL ?? '/api';
  return `${baseUrl}/${orgId}/graphql`;
}

function getGraphQLBaseUrl(): string {
  const baseUrl = (import.meta.env as Record<string, string | undefined>).VITE_API_URL ?? '/api';
  return baseUrl.replace(/\/$/, '');
}

/**
 * Extract the request path (pathname + search) from a URL. The server verifies
 * the HMAC against r.URL.RequestURI() (path + query string), so we sign
 * exactly that — mirroring the REST client's extractRequestPath.
 */
function extractRequestPath(uri: string): string {
  if (uri.startsWith('http://') || uri.startsWith('https://')) {
    try {
      const parsed = new URL(uri);
      return parsed.pathname + parsed.search;
    } catch {
      return uri;
    }
  }
  return uri;
}

/**
 * Create a fetch wrapper that signs each outgoing GraphQL request with the
 * per-session HMAC signing key (same X-Vyzorix-* scheme as the REST client).
 * If no signing key is available yet (pre-auth), the request is sent unsigned.
 */
function createSignedFetch(): typeof fetch {
  const signedFetch = async (input: Parameters<typeof fetch>[0], init?: RequestInit): Promise<Response> => {
    const signingKey = getSigningKey();

    if (signingKey && init?.body) {
      const method = (init.method || 'POST').toUpperCase();
      const bodyStr = typeof init.body === 'string' ? init.body : '';
      const uri = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
      const path = extractRequestPath(uri);

      const signHeaders = await signHttpRequestBrowser(method, path, bodyStr, signingKey);

      const headers = new Headers(init.headers);
      headers.set('X-Vyzorix-Timestamp', signHeaders['X-Vyzorix-Timestamp']);
      headers.set('X-Vyzorix-Nonce', signHeaders['X-Vyzorix-Nonce']);
      headers.set('X-Vyzorix-Signature', signHeaders['X-Vyzorix-Signature']);
      init.headers = headers;
    }

    return fetch(input, init);
  };
  return signedFetch as unknown as typeof fetch;
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
    fetch: createSignedFetch(),
  });

  const authLink = setContext((_, { headers }) => {
    const derivedHeaders: Record<string, string> = { ...headers };
    if (graphqlState.currentAuthToken) {
      derivedHeaders['Authorization'] = `Bearer ${graphqlState.currentAuthToken}`;
    }
    return { headers: derivedHeaders };
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
  if (!graphqlState.apolloClient) {
    graphqlState.apolloClient = createApolloClient({
      organizationId: '',
      credentials: 'include',
      getUri: () => getGraphQLUri(graphqlState.currentOrgId),
    });
  }
  return graphqlState.apolloClient;
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
