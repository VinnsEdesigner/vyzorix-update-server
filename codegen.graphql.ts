import type { CodegenConfig } from '@graphql-codegen/cli';

// Generates a typed GraphQL operation surface for the API Client from the
// code-first server schema (apps/api/swag/graphql/schema.graphql, produced by
// `go run ./cmd/graphql-schema`). Output is a tags-split-style surface under
// src/generated/graphql/ — the GraphQL analog of the orval REST SDK.
//
// Transport: typed documents are executed through the existing Apollo
// graphqlClient (src/vyzorServer/graphql/_shared/graphql-client.ts) via the
// handwritten executor in src/generated/graphql/executor.ts — HMAC signing,
// org-scoped /:org/graphql path, and request batching all come from there.
const config: CodegenConfig = {
  schema: 'apps/api/swag/graphql/schema.graphql',
  documents: [
    'packages/API_Client/src/vyzorServer/graphql/device/**/*.ts',
    'packages/API_Client/src/vyzorServer/graphql/settings/**/*.ts',
    'packages/API_Client/src/vyzorServer/graphql/updates/**/*.ts',
    'packages/API_Client/src/vyzorServer/graphql/diagnostics/**/*.ts',
    'packages/API_Client/src/vyzorServer/graphql/logs/**/*.ts',
    'packages/API_Client/src/vyzorServer/graphql/commands/**/*.ts',
    'packages/API_Client/src/vyzorServer/graphql/registration/**/*.ts',
  ],
  generates: {
    'packages/API_Client/src/generated/graphql/': {
      preset: 'client',
      presetConfig: {
        fragmentMasking: false,
      },
      config: {
        // Server schema uses camelCase field names; the Go resolvers return
        // epoch-millis numbers for timestamps. Keep scalars strict.
        scalars: {
          DateTime: 'string',
          Date: 'string',
          Time: 'string',
        },
      },
    },
  },
};

export default config;
