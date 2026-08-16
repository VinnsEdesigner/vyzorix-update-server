/**
 * MSW handlers for GraphQL-over-HTTP requests.
 *
 * The API client's GraphQL functions (queryUpdates, mutatePushUpdate, etc.)
 * use Apollo Client which sends POST requests to /{orgId}/graphql.
 * These handlers intercept those requests and return mock GraphQL responses.
 *
 * Each handler checks the operationName in the request body to route to the
 * correct mock response, mirroring how a real GraphQL server works.
 */
import { http, HttpResponse, delay } from 'msw';

const GraphQLEndpoint = '/:orgId/graphql';
const GraphQLBatchEndpoint = '/:orgId/graphql/batch';

type GraphQLHandler = (variables: Record<string, unknown>) => unknown;

const handlers: Record<string, GraphQLHandler> = {};

/** Register a GraphQL response handler for a given operation name. */
export function registerGraphQLResponse(operationName: string, handler: GraphQLHandler): void {
  handlers[operationName] = handler;
}

/** Clear all registered GraphQL response handlers. */
export function clearGraphQLResponses(): void {
  for (const key of Object.keys(handlers)) {
    delete handlers[key];
  }
}

function processOperation(body: {
  operationName?: string;
  query?: string;
  variables?: Record<string, unknown>;
}): unknown {
  const operationName = body.operationName ?? extractOperationName(body.query);
  const variables = body.variables ?? {};

  const handler = handlers[operationName ?? ''];
  if (!handler) {
    throw new Error(`No mock handler for GraphQL operation: ${operationName}`);
  }
  return handler(variables);
}

export function createGraphQLHandlers() {
  return [
    // Single-operation GraphQL requests (standard Apollo httpLink)
    http.post(GraphQLEndpoint, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as {
        operationName?: string;
        query?: string;
        variables?: Record<string, unknown>;
      };

      try {
        const data = processOperation(body);
        return HttpResponse.json({ data });
      } catch (err) {
        return HttpResponse.json(
          { errors: [{ message: err instanceof Error ? err.message : 'Unknown error' }] },
          { status: 200 },
        );
      }
    }),

    // Batched GraphQL requests (Apollo createBatchingLink)
    http.post(GraphQLBatchEndpoint, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as Array<{
        operationName?: string;
        query?: string;
        variables?: Record<string, unknown>;
      }>;

      const results = body.map((op) => {
        try {
          const data = processOperation(op);
          return { data };
        } catch (err) {
          return { errors: [{ message: err instanceof Error ? err.message : 'Unknown error' }] };
        }
      });

      return HttpResponse.json(results);
    }),
  ];
}

function extractOperationName(query?: string): string | undefined {
  if (!query) return undefined;
  const match = query.match(/(?:query|mutation)\s+(\w+)/);
  return match?.[1];
}
