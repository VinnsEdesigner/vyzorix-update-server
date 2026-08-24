import { getMainDefinition } from '@apollo/client/utilities';
import type { DocumentNode } from 'graphql';
import { graphqlClient } from '../../vyzorServer/graphql/_shared/graphql-client';

/**
 * GraphQL executor — the bridge between the orval-style generated typed
 * documents (src/generated/graphql/) and the existing Apollo transport
 * (graphqlClient). All HMAC signing, org-scoped /:org/graphql routing, and
 * request batching live in graphqlClient; this just dispatches to it.
 *
 * Queries go through `client.query`, mutations through `client.mutate`,
 * determined from the document's operation type.
 */

type AnyVariables = Record<string, unknown>;

function operationKind(document: DocumentNode): 'query' | 'mutation' | 'subscription' {
  const def = getMainDefinition(document);
  return def.kind === 'OperationDefinition' ? def.operation : 'query';
}

/** Execute a typed query document against the org-scoped GraphQL endpoint. */
export async function query<TResult = unknown, TVariables extends AnyVariables = AnyVariables>(
  document: DocumentNode,
  variables?: TVariables,
): Promise<TResult> {
  const result = await graphqlClient.getClient().query({
    query: document,
    variables,
    fetchPolicy: 'network-only',
  });
  return result.data as TResult;
}

/** Execute a typed mutation document against the org-scoped GraphQL endpoint. */
export async function mutate<TResult = unknown, TVariables extends AnyVariables = AnyVariables>(
  document: DocumentNode,
  variables?: TVariables,
): Promise<TResult> {
  const result = await graphqlClient.getClient().mutate({
    mutation: document,
    variables,
  });
  return result.data as TResult;
}

/** Execute a document, dispatching to query/mutate by its operation type. */
export async function execute<TResult = unknown, TVariables extends AnyVariables = AnyVariables>(
  document: DocumentNode,
  variables?: TVariables,
): Promise<TResult> {
  return operationKind(document) === 'mutation'
    ? mutate<TResult, TVariables>(document, variables)
    : query<TResult, TVariables>(document, variables);
}
