// Hand-rolled GraphQL domain modules were superseded by the generated typed
// documents in src/generated/graphql/ (graphql-codegen from
// apps/api/swag/graphql/schema.graphql). What remains here is the transport
// (Apollo graphqlClient: HMAC signing, org path, batching) and the wire→domain
// normalizers the web GraphQL fallbacks use.
export * from "./_shared";
